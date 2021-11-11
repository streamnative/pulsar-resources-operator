// Copyright (c) 2021 StreamNative, Inc.. All Rights Reserved.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PulsarConnectionSpec defines the desired state of PulsarConnection
type PulsarConnectionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of PulsarConnection. Edit pulsarconnection_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// PulsarConnectionStatus defines the observed state of PulsarConnection
type PulsarConnectionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PulsarConnection is the Schema for the pulsarconnections API
type PulsarConnection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PulsarConnectionSpec   `json:"spec,omitempty"`
	Status PulsarConnectionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PulsarConnectionList contains a list of PulsarConnection
type PulsarConnectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PulsarConnection `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PulsarConnection{}, &PulsarConnectionList{})
}
