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

// ServiceAccountBindingSpec defines the desired state of ServiceAccountBinding
type ServiceAccountBindingSpec struct {
	// ServiceAccountName refers to the ServiceAccount under the same namespace as this binding object
	// +required
	ServiceAccountName string `json:"serviceAccountName"`

	// APIServerRef is the reference to the StreamNativeCloudConnection
	// If not provided, it will be retrieved from the referenced ServiceAccount
	// +optional
	APIServerRef *corev1.LocalObjectReference `json:"apiServerRef,omitempty"`

	// PoolMemberRefs refers to a list of PoolMembers in the current namespace or other namespaces
	// +required
	PoolMemberRefs []PoolMemberReference `json:"poolMemberRefs"`
}

// ServiceAccountBindingStatus defines the observed state of ServiceAccountBinding
type ServiceAccountBindingStatus struct {
	// Conditions represent the latest available observations of an object's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the last observed generation
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Namespaced,categories={streamnative,all}
//+kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status"
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceAccountBinding is the Schema for the ServiceAccountBindings API
type ServiceAccountBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceAccountBindingSpec   `json:"spec,omitempty"`
	Status ServiceAccountBindingStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceAccountBindingList contains a list of ServiceAccountBinding
type ServiceAccountBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceAccountBinding `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ServiceAccountBinding{}, &ServiceAccountBindingList{})
}
