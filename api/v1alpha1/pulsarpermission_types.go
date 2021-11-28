// Copyright (c) 2021 StreamNative, Inc.. All Rights Reserved.

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PulsarPermissionSpec defines the desired state of PulsarPermission
type PulsarPermissionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ConnectionRef is the reference to the PulsarConnection resource
	ConnectionRef corev1.LocalObjectReference `json:"connectionRef"`

	// ResourceName name of the target resource which will be granted the permssions
	ResourceName string `json:"resourceName"`

	// +kubebuilder:validation:Enum=namespace;topic
	// ResourceType indicates the resource type, the options include namespace and topic
	ResoureType PulsarResoureType `json:"resourceType"`
	// Roles contains a list of role which will be granted the same permissions
	// for the same target
	Roles []string `json:"roles"`
	// Actions contains a list of action to grant.
	// the options include produce,consume,functions
	Actions []string `json:"actions,omitempty"`

	// LifecyclePolicy is the policy that how to deal with pulsar resource when
	// PulsarPermission is deleted
	// +optional
	LifecyclePolicy PulsarResourceLifeCyclePolicy `json:"lifecyclePolicy,omitempty"`
}

// PulsarResourceType indicates the resource type, the options include namespace and topic
type PulsarResoureType string

const (
	PulsarResourceTypeNamespace PulsarResoureType = "namespace"
	PulsarResourceTypeTopic     PulsarResoureType = "topic"
)

// PulsarPermissionStatus defines the observed state of PulsarPermission
type PulsarPermissionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ObservedGeneration is the most recent generation observed for this resource.
	// It corresponds to the metadata generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Represents the observations of a connection's current state.
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

// PulsarPermission is the Schema for the pulsarpermissions API
type PulsarPermission struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PulsarPermissionSpec   `json:"spec,omitempty"`
	Status PulsarPermissionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PulsarPermissionList contains a list of PulsarPermission
type PulsarPermissionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PulsarPermission `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PulsarPermission{}, &PulsarPermissionList{})
}
