// Copyright (c) 2019 Sylabs, Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package slurm

import (
	"archive/zip"
	"bytes"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/dptech-corp/wlm-operator/pkg/tail"
)

const (
	sbatchBinaryName   = "sbatch"
	scancelBinaryName  = "scancel"
	scontrolBinaryName = "scontrol"
	sacctBinaryName    = "sacct"
	sinfoBinaryName    = "sinfo"

	submitTime = "SubmitTime"
	startTime  = "StartTime"
	runTime    = "RunTime"
	timeLimit  = "TimeLimit"
)

var (
	// ErrDurationIsUnlimited means that duration field has value UNLIMITED
	ErrDurationIsUnlimited = errors.New("duration is unlimited")

	// ErrInvalidSacctResponse is returned when trying to parse sacct
	// response that is invalid.
	ErrInvalidSacctResponse = errors.New("unable to parse sacct response")

	// ErrFileNotFound is returned when Open fails to find a file.
	ErrFileNotFound = errors.New("file is not found")
)

type (
	// Client implements Slurm interface for communicating with
	// a local Slurm cluster by calling Slurm binaries directly.
	Client struct{}

	// JobInfo contains information about a Slurm job.
	JobInfo struct {
		ID         string         `json:"id" slurm:"JobId"`
		UserID     string         `json:"user_id" slurm:"UserId"`
		ArrayJobID string         `json:"array_job_id" slurm:"ArrayJobId"`
		Name       string         `json:"name" slurm:"JobName"`
		ExitCode   string         `json:"exit_code" slurm:"ExitCode"`
		State      string         `json:"state" slurm:"JobState"`
		SubmitTime *time.Time     `json:"submit_time" slurm:"SubmitTime"`
		StartTime  *time.Time     `json:"start_time" slurm:"StartTime"`
		RunTime    *time.Duration `json:"run_time" slurm:"RunTime"`
		TimeLimit  *time.Duration `json:"time_limit" slurm:"TimeLimit"`
		WorkDir    string         `json:"work_dir" slurm:"WorkDir"`
		StdOut     string         `json:"std_out" slurm:"StdOut"`
		StdErr     string         `json:"std_err" slurm:"StdErr"`
		Partition  string         `json:"partition" slurm:"Partition"`
		NodeList   string         `json:"node_list" slurm:"NodeList"`
		BatchHost  string         `json:"batch_host" slurm:"BatchHost"`
		NumNodes   string         `json:"num_nodes" slurm:"NumNodes"`
	}

	// JobStepInfo contains information about a single Slurm job step.
	JobStepInfo struct {
		ID         string     `json:"id"`
		Name       string     `json:"name"`
		StartedAt  *time.Time `json:"started_at"`
		FinishedAt *time.Time `json:"finished_at"`
		ExitCode   int        `json:"exit_code"`
		State      string     `json:"state"`
	}

	// Feature represents a single feature enabled on a Slurm partition.
	// TODO use it.
	Feature struct {
		Name     string
		Version  string
		Quantity int64
	}

	// Resources contain a list of available resources on a Slurm partition.
	Resources struct {
		Nodes      int64
		MemPerNode int64
		CPUPerNode int64
		WallTime   time.Duration
		Features   []Feature
	}
)

// NewClient returns new local client.
func NewClient() (*Client, error) {
	var missing []string
	for _, bin := range []string{
		sacctBinaryName,
		sbatchBinaryName,
		scancelBinaryName,
		scontrolBinaryName,
		sinfoBinaryName,
	} {
		_, err := exec.LookPath(bin)
		if err != nil {
			missing = append(missing, bin)
		}
	}
	if len(missing) != 0 {
		return nil, errors.Errorf("no slurm binaries found: %s", strings.Join(missing, ", "))
	}
	return &Client{}, nil
}

// SBatch submits batch job and returns job id if succeeded.
func (*Client) SBatch(script, partition string) (int64, error) {
	var partitionOpt string
	if partition != "" {
		partitionOpt = "--partition=" + partition
	}
	cmd := exec.Command(sbatchBinaryName, "--parsable", partitionOpt)
	cmd.Stdin = bytes.NewBufferString(script)

	out, err := cmd.CombinedOutput()
	if err != nil {
		if out != nil {
			log.Println(string(out))
		}
		return 0, errors.Wrap(err, "failed to execute sbatch")
	}

	id, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return 0, errors.Wrap(err, "could not parse job id")
	}

	return int64(id), nil
}

// SCancel cancels batch job.
func (*Client) SCancel(jobID int64) error {
	cmd := exec.Command(scancelBinaryName, strconv.FormatInt(jobID, 10))

	out, err := cmd.CombinedOutput()
	if err != nil && out != nil {
		log.Println(string(out))
	}
	return errors.Wrap(err, "failed to execute scancel")
}

// Open opens arbitrary file at path in a read-only mode.
func (*Client) Open(path string) (io.ReadCloser, error) {
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, ErrFileNotFound
	}
	return file, errors.Wrapf(err, "could not open %s", path)
}

// Create opens a file at path in write mode.
func (*Client) Create(path string) (io.WriteCloser, error) {
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return nil, err
	}
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	return file, errors.Wrapf(err, "could not create %s", path)
}

// Tail opens arbitrary file at path in a read-only mode.
// Unlike Open, Tail will watch file changes in a real-time.
func (*Client) Tail(path string) (io.ReadCloser, error) {
	tr, err := tail.NewReader(path)
	if err != nil {
		return nil, errors.Wrap(err, "could not create tail reader")
	}

	return tr, nil
}

// Zip file or directory
func (*Client) Zip(path string, target string) error {
	err := zipFile(path, target)
	if err != nil {
		return errors.Wrap(err, "could not zip file or directory")
	}
	return nil
}

// Unzip file or directory
func (*Client) Unzip(source string, path string) error {
	err := unzip(source, path)
	if err != nil {
		return errors.Wrap(err, "could not unzip file")
	}
	return nil
}

// SJobInfo returns information about a particular slurm job by ID.
func (*Client) SJobInfo(jobID int64) ([]*JobInfo, error) {
	cmd := exec.Command(scontrolBinaryName, "show", "jobid", strconv.FormatInt(jobID, 10))

	out, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get info for jobid: %d", jobID)
	}

	ji, err := jobInfoFromScontrolResponse(string(out))
	if err != nil {
		return nil, errors.Wrap(err, "could not parse scontrol response")
	}

	return ji, nil
}

// SJobSteps returns information about a submitted batch job.
func (*Client) SJobSteps(jobID int64) ([]*JobStepInfo, error) {
	cmd := exec.Command(sacctBinaryName,
		"-p",
		"-n",
		"-j",
		strconv.FormatInt(jobID, 10),
		"-o start,end,exitcode,state,jobid,jobname",
	)

	out, err := cmd.Output()
	if err != nil {
		ee, ok := err.(*exec.ExitError)
		if ok {
			return nil, errors.Wrapf(err, "failed to execute sacct: %s", ee.Stderr)
		}
		return nil, errors.Wrap(err, "failed to execute sacct")
	}

	jInfo, err := parseSacctResponse(string(out))
	if err != nil {
		return nil, errors.Wrap(err, ErrInvalidSacctResponse.Error())
	}

	return jInfo, nil
}

// Resources returns available resources for a partition.
func (*Client) Resources(partition string) (*Resources, error) {
	cmd := exec.Command(scontrolBinaryName, "show", "partition", partition)
	out, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrap(err, "could not get partition info")
	}

	r, err := parseResources(string(out))
	if err != nil {
		return nil, errors.Wrap(err, "could not parse partition resources")
	}

	return r, nil
}

// Partitions returns a list of partition names.
func (*Client) Partitions() ([]string, error) {
	cmd := exec.Command(scontrolBinaryName, "show", "partition")
	out, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrap(err, "could not get partition info")
	}
	return parsePartitionsNames(string(out)), nil
}

// Version returns slurm version
func (*Client) Version() (string, error) {
	cmd := exec.Command(sinfoBinaryName, "-V")
	out, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, "could not get slurm info")
	}

	s := strings.Split(string(out), " ")
	if len(s) != 2 {
		return "", errors.Wrapf(err, "could not parse sinfo response %s", string(out))
	}

	return s[1], nil
}

func jobInfoFromScontrolResponse(jobInfo string) ([]*JobInfo, error) {
	jobInfo = strings.TrimSpace(jobInfo)
	rawInfos := strings.Split(jobInfo, "\n\n")

	infos := make([]*JobInfo, len(rawInfos))
	for i, raw := range rawInfos {
		rFields := strings.Fields(raw)
		slurmFields := make(map[string]string)
		for _, f := range rFields {
			s := strings.Split(f, "=")
			if len(s) != 2 {
				// just skipping empty fields
				continue
			}
			slurmFields[s[0]] = s[1]
		}

		var ji JobInfo
		if err := ji.fillFromSlurmFields(slurmFields); err != nil {
			return nil, err
		}
		infos[i] = &ji
	}
	return infos, nil
}

func (ji *JobInfo) fillFromSlurmFields(fields map[string]string) error {
	t := reflect.TypeOf(*ji)
	for i := 0; i < t.NumField(); i++ {
		tagV, ok := t.Field(i).Tag.Lookup("slurm")
		if !ok {
			continue
		}

		sField, ok := fields[tagV]
		if !ok {
			continue
		}

		var val reflect.Value
		switch tagV {
		case submitTime, startTime:
			t, err := parseTime(sField)
			if err != nil {
				return errors.Wrapf(err, "could not parse time: %s", sField)
			}
			val = reflect.ValueOf(t)
		case runTime, timeLimit:
			d, err := ParseDuration(sField)
			if err != nil {
				if err == ErrDurationIsUnlimited {
					continue
				}

				return errors.Wrapf(err, "could not parse duration: %s", sField)
			}
			val = reflect.ValueOf(d)
		default:
			val = reflect.ValueOf(sField)
		}

		reflect.ValueOf(ji).Elem().Field(i).Set(val)
	}

	return nil
}

func zipFile(source, target string) error {
    // 1. Create a ZIP file and zip.Writer
    f, err := os.Create(target)
    if err != nil {
        return err
    }
    defer f.Close()

    writer := zip.NewWriter(f)
    defer writer.Close()

    // 2. Go through all the files of the source
    return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        // 3. Create a local file header
        header, err := zip.FileInfoHeader(info)
        if err != nil {
            return err
        }

        // set compression
        header.Method = zip.Deflate

        // 4. Set relative path of a file as the header name
        header.Name, err = filepath.Rel(filepath.Dir(source), path)
        if err != nil {
            return err
        }
        if info.IsDir() {
            header.Name += "/"
        }

        // 5. Create writer for the file header and save content of the file
        headerWriter, err := writer.CreateHeader(header)
        if err != nil {
            return err
        }

        if info.IsDir() {
            return nil
        }

        f, err := os.Open(path)
        if err != nil {
            return err
        }
        defer f.Close()

        _, err = io.Copy(headerWriter, f)
        return err
    })
}

func unzip(source, destination string) error {
    // 1. Open the zip file
    reader, err := zip.OpenReader(source)
    if err != nil {
        return err
    }
    defer reader.Close()

    // 2. Get the absolute destination path
    destination, err = filepath.Abs(destination)
    if err != nil {
        return err
    }

    // 3. Iterate over zip files inside the archive and unzip each of them
    for _, f := range reader.File {
        err := unzipFile(f, destination)
        if err != nil {
            return err
        }
    }

    return nil
}

func unzipFile(f *zip.File, destination string) error {
    // 4. Check if file paths are not vulnerable to Zip Slip
    filePath := filepath.Join(destination, f.Name)
    if !strings.HasPrefix(filePath, filepath.Clean(destination)+string(os.PathSeparator)) {
        return errors.Errorf("invalid file path: %s", filePath)
    }

    // 5. Create directory tree
    if f.FileInfo().IsDir() {
        if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
            return err
        }
        return nil
    }

    if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
        return err
    }

    // 6. Create a destination file for unzipped content
    destinationFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
    if err != nil {
        return err
    }
    defer destinationFile.Close()

    // 7. Unzip the content of a file and copy it to the destination file
    zippedFile, err := f.Open()
    if err != nil {
        return err
    }
    defer zippedFile.Close()

    if _, err := io.Copy(destinationFile, zippedFile); err != nil {
        return err
    }
    return nil
}
