// Copyright 2023 StreamNative
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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PulsarGeoReplicationSpec defines the desired state of PulsarGeoReplication
type PulsarGeoReplicationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ConnectionRef is the reference to the source PulsarConnection
	ConnectionRef PulsarConnectionRef `json:"connectionRef"`

	// DestinationConnectionRef is the connection reference to the remote cluster
	DestinationConnectionRef corev1.LocalObjectReference `json:"destinationConnectionRef"`

	// +kubebuilder:validation:Enum=CleanUpAfterDeletion;KeepAfterDeletion
	// +optional
	LifecyclePolicy PulsarResourceLifeCyclePolicy `json:"lifecyclePolicy,omitempty"`
}

// PulsarGeoReplicationStatus defines the observed state of PulsarGeoReplication
type PulsarGeoReplicationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ObservedGeneration is the most recent generation observed for this resource.
	// It corresponds to the metadata generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions Represents the observations of a connection's current state.
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PulsarGeoReplication is the Schema for the pulsargeoreplications API
type PulsarGeoReplication struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PulsarGeoReplicationSpec   `json:"spec,omitempty"`
	Status PulsarGeoReplicationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PulsarGeoReplicationList contains a list of PulsarGeoReplication
type PulsarGeoReplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PulsarGeoReplication `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PulsarGeoReplication{}, &PulsarGeoReplicationList{})
}

// ClusterInfo indicates the cluster info that will be used in the setup of GEO replication.
type ClusterInfo struct {
	// Name is the pulsar cluster name
	Name string `json:"name,omitempty"`
	// ConnectionRef is the connection reference that can connect to the pulsar cluster
	ConnectionRef PulsarConnectionRef `json:"connectionRef"`
}
