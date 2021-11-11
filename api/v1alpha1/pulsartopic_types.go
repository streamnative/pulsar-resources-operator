// Copyright (c) 2021 StreamNative, Inc.. All Rights Reserved.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PulsarTopicSpec defines the desired state of PulsarTopic
type PulsarTopicSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of PulsarTopic. Edit pulsartopic_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// PulsarTopicStatus defines the observed state of PulsarTopic
type PulsarTopicStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PulsarTopic is the Schema for the pulsartopics API
type PulsarTopic struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PulsarTopicSpec   `json:"spec,omitempty"`
	Status PulsarTopicStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PulsarTopicList contains a list of PulsarTopic
type PulsarTopicList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PulsarTopic `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PulsarTopic{}, &PulsarTopicList{})
}
