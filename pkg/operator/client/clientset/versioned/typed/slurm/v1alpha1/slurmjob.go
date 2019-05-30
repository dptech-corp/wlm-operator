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

// Code generated by main. DO NOT EDIT.

package v1alpha1

import (
	"time"

	v1alpha1 "github.com/sylabs/slurm-operator/pkg/operator/apis/slurm/v1alpha1"
	scheme "github.com/sylabs/slurm-operator/pkg/operator/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// SlurmJobsGetter has a method to return a SlurmJobInterface.
// A group's client should implement this interface.
type SlurmJobsGetter interface {
	SlurmJobs(namespace string) SlurmJobInterface
}

// SlurmJobInterface has methods to work with SlurmJob resources.
type SlurmJobInterface interface {
	Create(*v1alpha1.SlurmJob) (*v1alpha1.SlurmJob, error)
	Update(*v1alpha1.SlurmJob) (*v1alpha1.SlurmJob, error)
	UpdateStatus(*v1alpha1.SlurmJob) (*v1alpha1.SlurmJob, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.SlurmJob, error)
	List(opts v1.ListOptions) (*v1alpha1.SlurmJobList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.SlurmJob, err error)
	SlurmJobExpansion
}

// slurmJobs implements SlurmJobInterface
type slurmJobs struct {
	client rest.Interface
	ns     string
}

// newSlurmJobs returns a SlurmJobs
func newSlurmJobs(c *SlurmV1alpha1Client, namespace string) *slurmJobs {
	return &slurmJobs{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the slurmJob, and returns the corresponding slurmJob object, and an error if there is any.
func (c *slurmJobs) Get(name string, options v1.GetOptions) (result *v1alpha1.SlurmJob, err error) {
	result = &v1alpha1.SlurmJob{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("slurmjobs").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of SlurmJobs that match those selectors.
func (c *slurmJobs) List(opts v1.ListOptions) (result *v1alpha1.SlurmJobList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.SlurmJobList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("slurmjobs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested slurmJobs.
func (c *slurmJobs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("slurmjobs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a slurmJob and creates it.  Returns the server's representation of the slurmJob, and an error, if there is any.
func (c *slurmJobs) Create(slurmJob *v1alpha1.SlurmJob) (result *v1alpha1.SlurmJob, err error) {
	result = &v1alpha1.SlurmJob{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("slurmjobs").
		Body(slurmJob).
		Do().
		Into(result)
	return
}

// Update takes the representation of a slurmJob and updates it. Returns the server's representation of the slurmJob, and an error, if there is any.
func (c *slurmJobs) Update(slurmJob *v1alpha1.SlurmJob) (result *v1alpha1.SlurmJob, err error) {
	result = &v1alpha1.SlurmJob{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("slurmjobs").
		Name(slurmJob.Name).
		Body(slurmJob).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *slurmJobs) UpdateStatus(slurmJob *v1alpha1.SlurmJob) (result *v1alpha1.SlurmJob, err error) {
	result = &v1alpha1.SlurmJob{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("slurmjobs").
		Name(slurmJob.Name).
		SubResource("status").
		Body(slurmJob).
		Do().
		Into(result)
	return
}

// Delete takes name of the slurmJob and deletes it. Returns an error if one occurs.
func (c *slurmJobs) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("slurmjobs").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *slurmJobs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("slurmjobs").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched slurmJob.
func (c *slurmJobs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.SlurmJob, err error) {
	result = &v1alpha1.SlurmJob{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("slurmjobs").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
