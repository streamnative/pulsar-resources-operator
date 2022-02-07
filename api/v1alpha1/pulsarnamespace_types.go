// Copyright (c) 2022 StreamNative, Inc.. All Rights Reserved.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PulsarNamespaceSpec defines the desired state of PulsarNamespace
type PulsarNamespaceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of PulsarNamespace. Edit pulsarnamespace_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// PulsarNamespaceStatus defines the observed state of PulsarNamespace
type PulsarNamespaceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PulsarNamespace is the Schema for the pulsarnamespaces API
type PulsarNamespace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PulsarNamespaceSpec   `json:"spec,omitempty"`
	Status PulsarNamespaceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PulsarNamespaceList contains a list of PulsarNamespace
type PulsarNamespaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PulsarNamespace `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PulsarNamespace{}, &PulsarNamespaceList{})
}
