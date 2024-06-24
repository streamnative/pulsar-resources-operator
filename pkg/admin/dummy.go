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
func NewDummyPulsarAdmin(PulsarAdminConfig) (PulsarAdmin, error) {
	return &DummyPulsarAdmin{}, nil
}

// DummyPulsarAdmin is a dummy struct of PulsarAdmin
type DummyPulsarAdmin struct {
}

var _ PulsarAdmin = &DummyPulsarAdmin{}

// ApplyTenant is a fake implements of ApplyTenant
func (d *DummyPulsarAdmin) ApplyTenant(string, *TenantParams) error {
	return nil
}

// DeleteTenant is a fake implements of DeleteTenant
func (d *DummyPulsarAdmin) DeleteTenant(string) error {
	return nil
}

// ApplyNamespace is a fake implements of ApplyNamespace
func (d *DummyPulsarAdmin) ApplyNamespace(string, *NamespaceParams) error {
	return nil
}

// DeleteNamespace is a fake implements of DeleteNamespace
func (d *DummyPulsarAdmin) DeleteNamespace(string) error {
	return nil
}

// GetNamespaceClusters is a fake implements of GetNamespaceClusters
func (d *DummyPulsarAdmin) GetNamespaceClusters(string) ([]string, error) {
	return []string{}, nil
}

// SetNamespaceClusters is a fake implements of SetNamespaceClusters
func (d *DummyPulsarAdmin) SetNamespaceClusters(string, []string) error {
	return nil
}

// ApplyTopic is a fake implements of ApplyTopic
func (d *DummyPulsarAdmin) ApplyTopic(string, *TopicParams) error {
	return nil
}

// DeleteTopic is a fake implements of DeleteTopic
func (d *DummyPulsarAdmin) DeleteTopic(string) error {
	return nil
}

// GetTopicClusters is a fake implements of GetTopicClusters
func (d *DummyPulsarAdmin) GetTopicClusters(string, *bool) ([]string, error) {
	return []string{}, nil
}

// SetTopicClusters is a fake implements of SetTopicClusters
func (d *DummyPulsarAdmin) SetTopicClusters(string, *bool, []string) error {
	return nil
}

// Close is a fake implements of Close
func (d *DummyPulsarAdmin) Close() error {
	return nil
}

// GrantPermissions is a fake implements of GrantPermissions
func (d *DummyPulsarAdmin) GrantPermissions(Permissioner) error {
	return nil
}

// RevokePermissions is a fake implements of RevokePermissions
func (d *DummyPulsarAdmin) RevokePermissions(Permissioner) error {
	return nil
}

// GetSchema is a fake implements of GetSchema
func (d *DummyPulsarAdmin) GetSchema(string) (*v1alpha1.SchemaInfo, error) {
	return nil, nil
}

// UploadSchema is a fake implements of UploadSchema
func (d *DummyPulsarAdmin) UploadSchema(string, *SchemaParams) error {
	return nil
}

// DeleteSchema is a fake implements of DeleteSchema
func (d *DummyPulsarAdmin) DeleteSchema(string) error {
	return nil
}

// CreateCluster is a fake implements of CreateCluster
func (d *DummyPulsarAdmin) CreateCluster(string, *ClusterParams) error {
	return nil
}

// UpdateCluster is a fake implements of UpdateCluster
func (d *DummyPulsarAdmin) UpdateCluster(string, *ClusterParams) error {
	return nil
}

// DeleteCluster is a fake implements of DeleteCluster
func (d *DummyPulsarAdmin) DeleteCluster(string) error {
	return nil
}

// CheckClusterExist checks whether the cluster exists
func (d *DummyPulsarAdmin) CheckClusterExist(string) (bool, error) {
	return true, nil
}

// DeletePulsarPackage is a fake implements of DeletePulsarPackage
func (d *DummyPulsarAdmin) DeletePulsarPackage(_ string) error {
	return nil
}

// ApplyPulsarPackage is a fake implements of ApplyPulsarPackage
func (d *DummyPulsarAdmin) ApplyPulsarPackage(_, _, _, _ string, _ map[string]string, _ bool) error {
	return nil
}

// DeletePulsarFunction is a fake implements of DeletePulsarFunction
func (d *DummyPulsarAdmin) DeletePulsarFunction(_, _, _ string) error {
	return nil
}

// ApplyPulsarFunction is a fake implements of ApplyPulsarFunction
func (d *DummyPulsarAdmin) ApplyPulsarFunction(_, _, _, _ string, _ *v1alpha1.PulsarFunctionSpec, _ bool) error {
	return nil
}

// DeletePulsarSink is a fake implements of DeletePulsarSink
func (d *DummyPulsarAdmin) DeletePulsarSink(_, _, _ string) error {
	return nil
}

// ApplyPulsarSink is a fake implements of ApplyPulsarSink
func (d *DummyPulsarAdmin) ApplyPulsarSink(_, _, _, _ string, _ *v1alpha1.PulsarSinkSpec, _ bool) error {
	return nil
}

// DeletePulsarSource is a fake implements of DeletePulsarSource
func (d *DummyPulsarAdmin) DeletePulsarSource(_, _, _ string) error {
	return nil
}

// ApplyPulsarSource is a fake implements of ApplyPulsarSource
func (d *DummyPulsarAdmin) ApplyPulsarSource(_, _, _, _ string, _ *v1alpha1.PulsarSourceSpec, _ bool) error {
	return nil
}
