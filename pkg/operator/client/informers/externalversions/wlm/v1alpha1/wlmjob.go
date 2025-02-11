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
	time "time"

	wlmv1alpha1 "github.com/dptech-corp/wlm-operator/pkg/operator/apis/wlm/v1alpha1"
	versioned "github.com/dptech-corp/wlm-operator/pkg/operator/client/clientset/versioned"
	internalinterfaces "github.com/dptech-corp/wlm-operator/pkg/operator/client/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/dptech-corp/wlm-operator/pkg/operator/client/listers/wlm/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// WlmJobInformer provides access to a shared informer and lister for
// WlmJobs.
type WlmJobInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.WlmJobLister
}

type wlmJobInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewWlmJobInformer constructs a new informer for WlmJob type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewWlmJobInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredWlmJobInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredWlmJobInformer constructs a new informer for WlmJob type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredWlmJobInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.WlmV1alpha1().WlmJobs(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.WlmV1alpha1().WlmJobs(namespace).Watch(options)
			},
		},
		&wlmv1alpha1.WlmJob{},
		resyncPeriod,
		indexers,
	)
}

func (f *wlmJobInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredWlmJobInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *wlmJobInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&wlmv1alpha1.WlmJob{}, f.defaultInformer)
}

func (f *wlmJobInformer) Lister() v1alpha1.WlmJobLister {
	return v1alpha1.NewWlmJobLister(f.Informer().GetIndexer())
}
