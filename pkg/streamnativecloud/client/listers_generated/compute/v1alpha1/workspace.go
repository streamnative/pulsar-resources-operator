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
	v1alpha1 "github.com/streamnative/pulsar-resources-operator/pkg/streamnativecloud/apis/compute/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// WorkspaceLister helps list Workspaces.
// All objects returned here must be treated as read-only.
type WorkspaceLister interface {
	// List lists all Workspaces in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.Workspace, err error)
	// Workspaces returns an object that can list and get Workspaces.
	Workspaces(namespace string) WorkspaceNamespaceLister
	WorkspaceListerExpansion
}

// workspaceLister implements the WorkspaceLister interface.
type workspaceLister struct {
	indexer cache.Indexer
}

// NewWorkspaceLister returns a new WorkspaceLister.
func NewWorkspaceLister(indexer cache.Indexer) WorkspaceLister {
	return &workspaceLister{indexer: indexer}
}

// List lists all Workspaces in the indexer.
func (s *workspaceLister) List(selector labels.Selector) (ret []*v1alpha1.Workspace, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Workspace))
	})
	return ret, err
}

// Workspaces returns an object that can list and get Workspaces.
func (s *workspaceLister) Workspaces(namespace string) WorkspaceNamespaceLister {
	return workspaceNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// WorkspaceNamespaceLister helps list and get Workspaces.
// All objects returned here must be treated as read-only.
type WorkspaceNamespaceLister interface {
	// List lists all Workspaces in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.Workspace, err error)
	// Get retrieves the Workspace from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.Workspace, error)
	WorkspaceNamespaceListerExpansion
}

// workspaceNamespaceLister implements the WorkspaceNamespaceLister
// interface.
type workspaceNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Workspaces in the indexer for a given namespace.
func (s workspaceNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.Workspace, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Workspace))
	})
	return ret, err
}

// Get retrieves the Workspace from the indexer for a given namespace and name.
func (s workspaceNamespaceLister) Get(name string) (*v1alpha1.Workspace, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("workspace"), name)
	}
	return obj.(*v1alpha1.Workspace), nil
}
