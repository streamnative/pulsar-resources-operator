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

	// AdminServiceURL is the admin service url of the pulsar cluster
	// +kubebuilder:validation:Pattern="^https?://.+$"
	AdminServiceURL string `json:"adminServiceURL"`

	// Authentication defines authentication configurations
	// +optional
	Authentication *PulsarAuthentication `json:"authentication,omitempty"`
}

// PulsarConnectionStatus defines the observed state of PulsarConnection
type PulsarConnectionStatus struct {
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
//+kubebuilder:resource:categories=pulsar;pulsarres,shortName=pconn
//+kubebuilder:printcolumn:name="ADMIN_SERVICE_URL",type=string,JSONPath=`.spec.adminServiceUrl`
//+kubebuilder:printcolumn:name="GENERATION",type=string,JSONPath=`.metadata.generation`
//+kubebuilder:printcolumn:name="OBSERVED_GENERATION",type=string,JSONPath=`.status.observedGeneration`
//+kubebuilder:printcolumn:name="READY",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`

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
