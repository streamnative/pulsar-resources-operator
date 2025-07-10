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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RoleBindingSpec defines the desired state of RoleBinding
type RoleBindingSpec struct {
	// APIServerRef is the reference to the StreamNativeCloudConnection
	// +required
	APIServerRef corev1.LocalObjectReference `json:"apiServerRef"`

	// Users is a list of Users that will be granted the role
	// +optional
	Users []string `json:"users"`

	// IdentityPools is a list of IdentityPools that will be granted the role
	// +optional
	IdentityPools []string `json:"identityPools,omitempty"`

	// ServiceAccounts is a list of ServiceAccounts that will be granted the role
	// +optional
	ServiceAccounts []string `json:"serviceAccounts,omitempty"`

	// ClusterRole is the reference to the role that will be granted
	// +required
	ClusterRole string `json:"clusterRole"`

	// CEL is an optional CEL expression for the role binding
	// +optional
	CEL *string `json:"cel,omitempty"`

	// SRNOrganization is the organization of the SRN
	// +optional
	SRNOrganization *string `json:"srnOrganization,omitempty"`

	// SRNInstance is the pulsar instance of the SRN
	// +optional
	SRNInstance *string `json:"srnInstance,omitempty"`

	// SRNCluster is the cluster of the SRN
	// +optional
	SRNCluster *string `json:"srnCluster,omitempty"`

	// SRNTenant is the tenant of the SRN
	// +optional
	SRNTenant *string `json:"srnTenant,omitempty"`

	// SRNNamespace is the namespace of the SRN
	// +optional
	SRNNamespace *string `json:"srnNamespace,omitempty"`

	// SRNTopicDomain is the topic domain of the SRN
	// +optional
	SRNTopicDomain *string `json:"srnTopicDomain,omitempty"`

	// SRNTopicName is the topic of the SRN
	// +optional
	SRNTopicName *string `json:"srnTopicName,omitempty"`

	// SRNSubscription is the subscription of the SRN
	// +optional
	SRNSubscription *string `json:"srnSubscription,omitempty"`

	// SRNServiceAccount is the service account of the SRN
	// +optional
	SRNServiceAccount *string `json:"srnServiceAccount,omitempty"`

	// SRNSecret is the secret of the SRN
	// +optional
	SRNSecret *string `json:"srnSecret,omitempty"`
}

// RoleBindingStatus defines the observed state of RoleBinding
type RoleBindingStatus struct {
	// Conditions represent the latest available observations of an object's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the last observed generation
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// FailedClusters is a list of clusters where the role binding failed
	// +optional
	FailedClusters []string `json:"failedClusters,omitempty"`

	// SyncedClusters is a map of clusters where the role binding is synced
	// +optional
	SyncedClusters map[string]string `json:"syncedClusters,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Namespaced,categories={streamnative,all}
//+kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status"
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RoleBinding is the Schema for the RoleBindings API
type RoleBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RoleBindingSpec   `json:"spec,omitempty"`
	Status RoleBindingStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RoleBindingList contains a list of RoleBinding
type RoleBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RoleBinding `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RoleBinding{}, &RoleBindingList{})
}
