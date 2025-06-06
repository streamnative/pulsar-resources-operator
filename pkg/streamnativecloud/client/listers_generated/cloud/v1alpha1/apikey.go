// Copyright 2025 StreamNative
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/streamnative/pulsar-resources-operator/pkg/streamnativecloud/apis/cloud/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// APIKeyLister helps list APIKeys.
// All objects returned here must be treated as read-only.
type APIKeyLister interface {
	// List lists all APIKeys in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.APIKey, err error)
	// APIKeys returns an object that can list and get APIKeys.
	APIKeys(namespace string) APIKeyNamespaceLister
	APIKeyListerExpansion
}

// aPIKeyLister implements the APIKeyLister interface.
type aPIKeyLister struct {
	indexer cache.Indexer
}

// NewAPIKeyLister returns a new APIKeyLister.
func NewAPIKeyLister(indexer cache.Indexer) APIKeyLister {
	return &aPIKeyLister{indexer: indexer}
}

// List lists all APIKeys in the indexer.
func (s *aPIKeyLister) List(selector labels.Selector) (ret []*v1alpha1.APIKey, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.APIKey))
	})
	return ret, err
}

// APIKeys returns an object that can list and get APIKeys.
func (s *aPIKeyLister) APIKeys(namespace string) APIKeyNamespaceLister {
	return aPIKeyNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// APIKeyNamespaceLister helps list and get APIKeys.
// All objects returned here must be treated as read-only.
type APIKeyNamespaceLister interface {
	// List lists all APIKeys in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.APIKey, err error)
	// Get retrieves the APIKey from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.APIKey, error)
	APIKeyNamespaceListerExpansion
}

// aPIKeyNamespaceLister implements the APIKeyNamespaceLister
// interface.
type aPIKeyNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all APIKeys in the indexer for a given namespace.
func (s aPIKeyNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.APIKey, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.APIKey))
	})
	return ret, err
}

// Get retrieves the APIKey from the indexer for a given namespace and name.
func (s aPIKeyNamespaceLister) Get(name string) (*v1alpha1.APIKey, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("apikey"), name)
	}
	return obj.(*v1alpha1.APIKey), nil
}
