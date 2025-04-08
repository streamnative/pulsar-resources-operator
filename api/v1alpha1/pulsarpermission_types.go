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

// PulsarPermissionSpec defines the desired state of PulsarPermission.
// It specifies the configuration for granting permissions to Pulsar resources.
type PulsarPermissionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ConnectionRef is the reference to the PulsarConnection resource
	// used to connect to the Pulsar cluster for this permission.
	ConnectionRef corev1.LocalObjectReference `json:"connectionRef"`

	// ResourceName is the name of the target resource (namespace or topic)
	// to which the permissions will be granted.
	ResourceName string `json:"resourceName"`

	// ResourceType indicates whether the permission is for a namespace or a topic.
	// +kubebuilder:validation:Enum=namespace;topic
	ResoureType PulsarResourceType `json:"resourceType"`

	// Roles is a list of role names that will be granted the specified permissions
	// for the target resource.
	Roles []string `json:"roles"`

	// Actions is a list of permissions to grant.
	// Valid options include "produce", "consume", and "functions".
	// +optional
	Actions []string `json:"actions,omitempty"`

	// LifecyclePolicy determines how to handle the Pulsar permissions
	// when the PulsarPermission resource is deleted.
	// +optional
	LifecyclePolicy PulsarResourceLifeCyclePolicy `json:"lifecyclePolicy,omitempty"`
}

// PulsarResourceType indicates the type of Pulsar resource for which permissions can be granted.
// Currently, it supports namespace and topic level permissions.
type PulsarResourceType string

const (
	// PulsarResourceTypeNamespace represents a Pulsar namespace resource.
	// Use this when granting permissions at the namespace level.
	PulsarResourceTypeNamespace PulsarResourceType = "namespace"

	// PulsarResourceTypeTopic represents a Pulsar topic resource.
	// Use this when granting permissions at the individual topic level.
	PulsarResourceTypeTopic PulsarResourceType = "topic"
)

// PulsarPermissionStatus defines the observed state of PulsarPermission.
// It provides information about the current status of the Pulsar permission configuration.
type PulsarPermissionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ObservedGeneration is the most recent generation observed for this resource.
	// It corresponds to the metadata generation, which is updated on mutation by the API Server.
	// This field is used to track whether the controller has processed the latest changes.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions represent the latest available observations of the PulsarPermission's current state.
	// It follows the Kubernetes conventions for condition types and status.
	// The "Ready" condition type is typically used to indicate the overall status of the permission configuration.
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:categories=pulsar;pulsarres,shortName=ppermission
//+kubebuilder:printcolumn:name="RESOURCE NAME",type=string,JSONPath=`.spec.resourceName`
//+kubebuilder:printcolumn:name="RESOURCE TYPE",type=string,JSONPath=`.spec.resourceType`
//+kubebuilder:printcolumn:name="ROLES",type=string,JSONPath=`.spec.roles`
//+kubebuilder:printcolumn:name="ACTIONS",type=string,JSONPath=`.spec.actions`
//+kubebuilder:printcolumn:name="GENERATION",type=string,JSONPath=`.metadata.generation`
//+kubebuilder:printcolumn:name="OBSERVED GENERATION",type=string,JSONPath=`.status.observedGeneration`
//+kubebuilder:printcolumn:name="READY",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`

// PulsarPermission is the Schema for the pulsarpermissions API.
// It represents a set of permissions granted to specific roles for a Pulsar resource (namespace or topic).
type PulsarPermission struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PulsarPermissionSpec   `json:"spec,omitempty"`
	Status PulsarPermissionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PulsarPermissionList contains a list of PulsarPermission resources.
// It is used by the Kubernetes API to return multiple PulsarPermission objects.
type PulsarPermissionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PulsarPermission `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PulsarPermission{}, &PulsarPermissionList{})
}
