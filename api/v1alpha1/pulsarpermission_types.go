// Copyright (c) 2022 StreamNative, Inc.. All Rights Reserved.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PulsarPermissionSpec defines the desired state of PulsarPermission
type PulsarPermissionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of PulsarPermission. Edit pulsarpermission_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// PulsarPermissionStatus defines the observed state of PulsarPermission
type PulsarPermissionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

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
