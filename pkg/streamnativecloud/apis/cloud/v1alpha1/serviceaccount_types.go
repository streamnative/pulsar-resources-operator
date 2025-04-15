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

// ServiceAccount is a resource that represents a service account for a cluster
// +k8s:openapi-gen=true
// +resource:path=serviceaccounts,strategy=ServiceAccountStrategy
// +kubebuilder:categories=all
type ServiceAccount struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   ServiceAccountSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status ServiceAccountStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// ServiceAccountSpec defines the desired state of ServiceAccount
type ServiceAccountSpec struct {
}

// ServiceAccountStatus defines the observed state of ServiceAccount
type ServiceAccountStatus struct {

	// PrivateKeyType indicates the type of private key information
	// +optional
	PrivateKeyType string `json:"privateKeyType,omitempty" protobuf:"bytes,1,opt,name=privateKeyType"`

	// PrivateKeyData provides the private key data (in base-64 format) for authentication purposes
	// +optional
	PrivateKeyData string `json:"privateKeyData,omitempty" protobuf:"bytes,2,opt,name=privateKeyData"`

	// Conditions is an array of current observed service account conditions.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,3,rep,name=conditions"`
}

//+kubebuilder:object:root=true

// ServiceAccountList contains a list of ServiceAccount
type ServiceAccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceAccount `json:"items"`
}
