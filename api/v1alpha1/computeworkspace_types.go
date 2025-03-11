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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ComputeWorkspaceSpec defines the desired state of Workspace
type ComputeWorkspaceSpec struct {
	// APIServerRef is the reference to the StreamNativeCloudConnection
	// +required
	APIServerRef StreamNativeCloudConnectionRef `json:"apiServerRef"`

	// PulsarClusterNames is the list of Pulsar clusters that the workspace will have access to.
	// +optional
	PulsarClusterNames []string `json:"pulsarClusterNames,omitempty"`

	// PoolRef is the reference to the pool that the workspace will be access to.
	// +optional
	PoolRef *PoolRef `json:"poolRef"`

	// UseExternalAccess is the flag to indicate whether the workspace will use external access.
	// +optional
	UseExternalAccess *bool `json:"useExternalAccess,omitempty"`

	// FlinkBlobStorage is the configuration for the Flink blob storage.
	// +optional
	FlinkBlobStorage *FlinkBlobStorage `json:"flinkBlobStorage,omitempty"`
}

// PoolRef is a reference to a pool with a given name.
type PoolRef struct {
	// Namespace is the namespace of the pool
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Name is the name of the pool
	// +required
	Name string `json:"name"`
}

// FlinkBlobStorage defines the configuration for the Flink blob storage.
type FlinkBlobStorage struct {
	// Bucket is required if you want to use cloud storage.
	// +required
	Bucket string `json:"bucket"`

	// Path is the sub path in the bucket.
	// Leave it empty if you want to use the whole bucket.
	// +optional
	Path string `json:"path,omitempty"`
}

// ComputeWorkspaceStatus defines the observed state of Workspace
type ComputeWorkspaceStatus struct {
	// Conditions represent the latest available observations of an object's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the last observed generation.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// WorkspaceID is the ID of the workspace in the API server
	// +optional
	WorkspaceID string `json:"workspaceId,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Namespaced,categories={streamnative,all}
//+kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status"

// ComputeWorkspace is the Schema for the workspaces API
type ComputeWorkspace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ComputeWorkspaceSpec   `json:"spec,omitempty"`
	Status ComputeWorkspaceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ComputeWorkspaceList contains a list of ComputeWorkspace
type ComputeWorkspaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ComputeWorkspace `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ComputeWorkspace{}, &ComputeWorkspaceList{})
}
