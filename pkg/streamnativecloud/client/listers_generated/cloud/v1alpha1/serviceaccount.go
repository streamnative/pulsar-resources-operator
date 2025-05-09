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

// ServiceAccountLister helps list ServiceAccounts.
// All objects returned here must be treated as read-only.
type ServiceAccountLister interface {
	// List lists all ServiceAccounts in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.ServiceAccount, err error)
	// ServiceAccounts returns an object that can list and get ServiceAccounts.
	ServiceAccounts(namespace string) ServiceAccountNamespaceLister
	ServiceAccountListerExpansion
}

// serviceAccountLister implements the ServiceAccountLister interface.
type serviceAccountLister struct {
	indexer cache.Indexer
}

// NewServiceAccountLister returns a new ServiceAccountLister.
func NewServiceAccountLister(indexer cache.Indexer) ServiceAccountLister {
	return &serviceAccountLister{indexer: indexer}
}

// List lists all ServiceAccounts in the indexer.
func (s *serviceAccountLister) List(selector labels.Selector) (ret []*v1alpha1.ServiceAccount, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.ServiceAccount))
	})
	return ret, err
}

// ServiceAccounts returns an object that can list and get ServiceAccounts.
func (s *serviceAccountLister) ServiceAccounts(namespace string) ServiceAccountNamespaceLister {
	return serviceAccountNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// ServiceAccountNamespaceLister helps list and get ServiceAccounts.
// All objects returned here must be treated as read-only.
type ServiceAccountNamespaceLister interface {
	// List lists all ServiceAccounts in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.ServiceAccount, err error)
	// Get retrieves the ServiceAccount from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.ServiceAccount, error)
	ServiceAccountNamespaceListerExpansion
}

// serviceAccountNamespaceLister implements the ServiceAccountNamespaceLister
// interface.
type serviceAccountNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all ServiceAccounts in the indexer for a given namespace.
func (s serviceAccountNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.ServiceAccount, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.ServiceAccount))
	})
	return ret, err
}

// Get retrieves the ServiceAccount from the indexer for a given namespace and name.
func (s serviceAccountNamespaceLister) Get(name string) (*v1alpha1.ServiceAccount, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("serviceaccount"), name)
	}
	return obj.(*v1alpha1.ServiceAccount), nil
}
