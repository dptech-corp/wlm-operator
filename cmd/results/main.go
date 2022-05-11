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

package main

import (
	"archive/zip"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/sylabs/wlm-operator/pkg/workload/api"
	"google.golang.org/grpc"
)

var (
	version = "unknown"

	from   = flag.String("from", "", "specify path to transfer")
	to     = flag.String("to", "", "specify directory where to put file")
	upload = flag.Bool("upload", false, "whether to upload file to remote, download for default")

	redBoxSock = flag.String("sock", "", "path to red-box socket")
)

func main() {
	fmt.Printf("version: %s\n", version)

	flag.Parse()

	if *from == "" {
		panic("from can't be empty")
	}

	if *to == "" {
		panic("to can't be empty")
	}

	if *redBoxSock == "" {
		panic("path to red-box socket can't be empty")
	}

	conn, err := grpc.Dial("unix://"+*redBoxSock, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can't connect to %s %s", *redBoxSock, err)
	}
	client := api.NewWorkloadManagerClient(conn)

	if *upload {

		err = zipFile(*from, *from+".zip")
		if err != nil {
			log.Fatalf("can't zip local file err: %s", err)
		}

		createReq, err := client.CreateFile(context.Background())
		if err != nil {
			log.Fatalf("can't create file err: %s", err)
		}

		fd, err := os.Open(*from + ".zip")
		if err != nil {
			log.Fatalf("can't open local file err: %s", err)
		}
		defer fd.Close()

		if err := createReq.Send(&api.CreateFileRequest{Path: path.Join(*to, filepath.Base(*from)+".zip")}); err != nil {
			log.Fatalf("can't send file err: %s", err)
		}

		buff := make([]byte, 128)
		for {
			n, err := fd.Read(buff)

			if n > 0 {
				if err := createReq.Send(&api.CreateFileRequest{Content: buff[:n]}); err != nil {
					log.Fatalf("can't send file err: %s", err)
				}
			}

			if err != nil {
				if err == io.EOF {
					break
				}
				log.Fatalf("err while reading file %s", err)
			}
		}

		if _, err = createReq.CloseAndRecv(); err != nil {
			log.Fatalf("can't send file err: %s", err)
		}

		_, err = client.Unzip(context.Background(), &api.UnzipRequest{Source: path.Join(*to, filepath.Base(*from)+".zip"), Path: *to})
		if err != nil {
			log.Fatalf("can't unzip remote file err: %s", err)
		}

		log.Println("Uploading data ended")
		log.Printf("File is located at %s", *to)

	} else {

		_, err = client.Zip(context.Background(), &api.ZipRequest{Path: *from, Target: *from + ".zip"})
		if err != nil {
			log.Fatalf("can't zip remote file err: %s", err)
		}

		openReq, err := client.OpenFile(context.Background(), &api.OpenFileRequest{Path: *from + ".zip"})
		if err != nil {
			log.Fatalf("can't open file err: %s", err)
		}

		if err := os.MkdirAll(*to, 0755); err != nil {
			log.Fatalf("can't create dir on mounted volume err: %s", err)
		}

		filePath := path.Join(*to, filepath.Base(*from+".zip"))
		toFile, err := os.Create(filePath)
		if err != nil {
			log.Fatalf("can't not create file with results on mounted volume err: %s", err)
		}
		defer toFile.Close()

		for {
			chunk, err := openReq.Recv()

			if chunk != nil {
				if _, err := toFile.Write(chunk.Content); err != nil {
					log.Fatalf("can't write to file err: %s", err)
				}
			}

			if err != nil {
				if err == io.EOF {
					break
				}

				log.Fatalf("err while receiving file %s", err)
			}
		}

		err = unzip(filePath, ".")
		if err != nil {
			log.Fatalf("can't unzip local file err: %s", err)
		}

		log.Println("Collecting results ended")
		log.Printf("File is located at %s", *to)
	}
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
