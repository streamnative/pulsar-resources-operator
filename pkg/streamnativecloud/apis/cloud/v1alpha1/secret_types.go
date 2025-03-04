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

// Copyright (c) 2020 StreamNative, Inc.. All Rights Reserved.

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Secret
// +k8s:openapi-gen=true
// +resource:path=secrets,strategy=SecretStrategy
// +kubebuilder:categories=all
type Secret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   SecretSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status SecretStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`

	// InstanceName is the name of the instance this secret is for (e.g. pulsar-instance)
	// +optional
	InstanceName string `json:"instanceName" protobuf:"bytes,4,opt,name=instanceName"`

	// Location is the location of the secret.
	// +optional
	Location string `json:"location" protobuf:"bytes,5,opt,name=location"`

	// the value should be base64 encoded
	Data map[string]string `json:"data,omitempty" protobuf:"bytes,6,opt,name=data"`

	// PoolMemberRef is the pool member to deploy the secret.
	// admission controller will infer this information automatically
	// +optional
	PoolMemberRef *PoolMemberReference `json:"poolMemberRef,omitempty" protobuf:"bytes,7,opt,name=poolMemberRef"`

	// +optional
	// +listType=atomic
	Tolerations []Toleration `json:"tolerations,omitempty" protobuf:"bytes,8,opt,name=tolerations"`

	// Type Used to facilitate programmatic handling of secret data.
	// +optional
	Type *corev1.SecretType `json:"type,omitempty" protobuf:"bytes,9,opt,name=type"`
}

//+kubebuilder:object:root=true

// SecretList contains a list of Secret
type SecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Secret `json:"items"`
}

type SecretSpec struct {
}

type SecretStatus struct {
}
