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

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceAccountBinding
// +k8s:openapi-gen=true
// +resource:path=serviceaccountbindings,strategy=ServiceAccountBindingStrategy
// +kubebuilder:categories=all
type ServiceAccountBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   ServiceAccountBindingSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status ServiceAccountBindingStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// ServiceAccountBindingSpec defines the desired state of ServiceAccountBinding
type ServiceAccountBindingSpec struct {
	// refers to the ServiceAccount under the same namespace as this binding object
	ServiceAccountName string `json:"serviceAccountName,omitempty" protobuf:"bytes,4,name=serviceAccountName"`

	// refers to a PoolMember in the current namespace or other namespaces
	PoolMemberRef PoolMemberReference `json:"poolMemberRef,omitempty" protobuf:"bytes,5,name=poolMemberRef"`
}

// ServiceAccountBindingStatus defines the observed state of ServiceAccountBinding
type ServiceAccountBindingStatus struct {
	// Conditions is an array of current observed service account binding conditions.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

//+kubebuilder:object:root=true

// ServiceAccountBindingList contains a list of ServiceAccountBinding
type ServiceAccountBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceAccountBinding `json:"items"`
}
