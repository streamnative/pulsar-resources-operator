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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PulsarTenantSpec defines the desired state of PulsarTenant.
// It corresponds to the configuration options available in Pulsar's tenant admin API.
type PulsarTenantSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// TODO make these fields immutable

	// Name is the tenant name.
	// This field is required and must be unique within the Pulsar cluster.
	Name string `json:"name"`

	// ConnectionRef is the reference to the PulsarConnection resource
	// used to connect to the Pulsar cluster for this tenant.
	ConnectionRef corev1.LocalObjectReference `json:"connectionRef"`

	// AdminRoles is a list of roles that have administrative privileges for this tenant.
	// These roles can perform actions like creating namespaces, topics, and managing permissions.
	// +optional
	AdminRoles []string `json:"adminRoles,omitempty"`

	// AllowedClusters is a list of clusters that this tenant is allowed to access.
	// This field is optional and can be used to restrict the clusters a tenant can connect to.
	// Please use `GeoReplicationRefs` instead if you are setting up geo-replication
	// between multiple Pulsar instances.
	// +optional
	AllowedClusters []string `json:"allowedClusters,omitempty"`

	// LifecyclePolicy determines whether to keep or delete the Pulsar tenant
	// when the Kubernetes resource is deleted.
	// +kubebuilder:validation:Enum=CleanUpAfterDeletion;KeepAfterDeletion
	// +optional
	LifecyclePolicy PulsarResourceLifeCyclePolicy `json:"lifecyclePolicy,omitempty"`

	// GeoReplicationRefs is a list of references to PulsarGeoReplication resources,
	// used to configure geo-replication for this tenant across multiple Pulsar instances.
	// This is **ONLY** used when you are using PulsarGeoReplication for setting up geo-replication
	// between multiple Pulsar instances.
	// Please use `AllowedClusters` instead if you are allowing a tenant to be available within
	// specific clusters in a same Pulsar instance.
	// +optional
	GeoReplicationRefs []*corev1.LocalObjectReference `json:"geoReplicationRefs,omitempty"`
}

// PulsarTenantStatus defines the observed state of PulsarTenant.
// It contains information about the current state of the Pulsar tenant.
type PulsarTenantStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ObservedGeneration is the most recent generation observed for this resource.
	// It corresponds to the metadata generation, which is updated on mutation by the API Server.
	// This field is used to track whether the controller has processed the latest changes.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions represent the latest available observations of the PulsarTenant's current state.
	// It follows the Kubernetes conventions for condition types and status.
	// The "Ready" condition type is typically used to indicate the overall status of the tenant.
	// Other condition types may be used to provide more detailed status information.
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:categories=pulsar;pulsarres,shortName=ptenant
//+kubebuilder:printcolumn:name="RESOURCE_NAME",type=string,JSONPath=`.spec.name`
//+kubebuilder:printcolumn:name="GENERATION",type=string,JSONPath=`.metadata.generation`
//+kubebuilder:printcolumn:name="OBSERVED_GENERATION",type=string,JSONPath=`.status.observedGeneration`
//+kubebuilder:printcolumn:name="READY",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`

// PulsarTenant is the Schema for the pulsartenants API
type PulsarTenant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PulsarTenantSpec   `json:"spec,omitempty"`
	Status PulsarTenantStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PulsarTenantList contains a list of PulsarTenant
type PulsarTenantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PulsarTenant `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PulsarTenant{}, &PulsarTenantList{})
}
