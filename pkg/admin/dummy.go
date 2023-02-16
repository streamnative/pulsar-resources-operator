// Copyright 2022 StreamNative
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

package admin

import (
	"github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
)

// NewDummyPulsarAdmin is a dummy initialization function
func NewDummyPulsarAdmin(config PulsarAdminConfig) (PulsarAdmin, error) {
	return &DummyPulsarAdmin{}, nil
}

// DummyPulsarAdmin is a dummy struct of PulsarAdmin
type DummyPulsarAdmin struct {
}

var _ PulsarAdmin = &DummyPulsarAdmin{}

// ApplyTenant is a fake implements of ApplyTenant
func (d *DummyPulsarAdmin) ApplyTenant(name string, params *TenantParams) error {
	return nil
}

// DeleteTenant is a fake implements of DeleteTenant
func (d *DummyPulsarAdmin) DeleteTenant(name string) error {
	return nil
}

// ApplyNamespace is a fake implements of ApplyNamespace
func (d *DummyPulsarAdmin) ApplyNamespace(name string, params *NamespaceParams) error {
	return nil
}

// DeleteNamespace is a fake implements of DeleteNamespace
func (d *DummyPulsarAdmin) DeleteNamespace(name string) error {
	return nil
}

// ResetNamespaceCluster is a fake implements of ResetNamespaceCluster
func (d *DummyPulsarAdmin) ResetNamespaceCluster(name string) error {
	return nil
}

// ApplyTopic is a fake implements of ApplyTopic
func (d *DummyPulsarAdmin) ApplyTopic(name string, params *TopicParams) error {
	return nil
}

// DeleteTopic is a fake implements of DeleteTopic
func (d *DummyPulsarAdmin) DeleteTopic(name string) error {
	return nil
}

// Close is a fake implements of Close
func (d *DummyPulsarAdmin) Close() error {
	return nil
}

// GrantPermissions is a fake implements of GrantPermissions
func (d *DummyPulsarAdmin) GrantPermissions(p Permissioner) error {
	return nil
}

// RevokePermissions is a fake implements of RevokePermissions
func (d *DummyPulsarAdmin) RevokePermissions(p Permissioner) error {
	return nil
}

// GetSchema is a fake implements of GetSchema
func (d *DummyPulsarAdmin) GetSchema(topic string) (*v1alpha1.SchemaInfo, error) {
	return nil, nil
}

// UploadSchema is a fake implements of UploadSchema
func (d *DummyPulsarAdmin) UploadSchema(topic string, params *SchemaParams) error {
	return nil
}

// DeleteSchema is a fake implements of DeleteSchema
func (d *DummyPulsarAdmin) DeleteSchema(topic string) error {
	return nil
}

// CreateCluster is a fake implements of CreateCluster
func (d *DummyPulsarAdmin) CreateCluster(name string, param *ClusterParams) error {
	return nil
}

// UpdateCluster is a fake implements of UpdateCluster
func (d *DummyPulsarAdmin) UpdateCluster(name string, param *ClusterParams) error {
	return nil
}

// DeleteCluster is a fake implements of DeleteCluster
func (d *DummyPulsarAdmin) DeleteCluster(name string) error {
	return nil
}
func (d *DummyPulsarAdmin) CheckClusterExist(name string) bool {
	return true
}
