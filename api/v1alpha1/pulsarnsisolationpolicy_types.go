// Copyright 2024 StreamNative
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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PulsarNSIsolationPolicySpec defines the desired state of a Pulsar namespace isolation policy.
// It corresponds to the configuration options available in Pulsar's namespaceIsolationPolicies admin API.
type PulsarNSIsolationPolicySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Name is the policy name
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Cluster is the name of the Pulsar Cluster
	// +kubebuilder:validation:Required
	Cluster string `json:"cluster"`

	// ConnectionRef is the reference to the PulsarConnection resource
	// used to connect to the Pulsar cluster for this namespace.
	ConnectionRef corev1.LocalObjectReference `json:"connectionRef"`

	// Namespaces namespaces-regex list
	// +kubebuilder:validation:Required
	Namespaces []string `json:"namespaces"`

	// Primary primary-broker-regex list
	// +kubebuilder:validation:Required
	Primary []string `json:"primary"`

	// Secondary secondary-broker-regex list, optional
	// +optional
	Secondary []string `json:"secondary,omitempty"`

	// AutoFailoverPolicyType auto failover policy type name, only support min_available now
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=min_available
	AutoFailoverPolicyType AutoFailoverPolicyType `json:"autoFailoverPolicyType"`

	// AutoFailoverPolicyParams auto failover policy parameters
	// +kubebuilder:validation:Required
	AutoFailoverPolicyParams map[string]string `json:"autoFailoverPolicyParams"`
}

type AutoFailoverPolicyType string

const (
	MinAvailable AutoFailoverPolicyType = "min_available"
)

// PulsarNSIsolationPolicyStatus defines the observed state of PulsarNSIsolationPolicy
type PulsarNSIsolationPolicyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ObservedGeneration is the most recent generation observed for this resource.
	// It corresponds to the metadata generation, which is updated on mutation by the API Server.
	// This field is used to track whether the controller has processed the latest changes.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions represent the latest available observations of the ns-isolation-policy's current state.
	// It follows the Kubernetes conventions for condition types and status.
	// The "Ready" condition type is typically used to indicate the overall status of the ns-isolation-policy.
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:categories=pulsar;pulsarres,shortName=pns
//+kubebuilder:printcolumn:name="RESOURCE_NAME",type=string,JSONPath=`.spec.name`
//+kubebuilder:printcolumn:name="GENERATION",type=string,JSONPath=`.metadata.generation`
//+kubebuilder:printcolumn:name="OBSERVED_GENERATION",type=string,JSONPath=`.status.observedGeneration`
//+kubebuilder:printcolumn:name="READY",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`

// PulsarNSIsolationPolicy is the Schema for the pulsar ns-isolation-policy API
// It represents a Pulsar NsIsolationPolicy in the Kubernetes cluster and includes both
// the desired state (Spec) and the observed state (Status) of the policy.
type PulsarNSIsolationPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PulsarNSIsolationPolicySpec   `json:"spec,omitempty"`
	Status PulsarNSIsolationPolicyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PulsarNSIsolationPolicyList contains a list of PulsarNSIsolationPolicy resources.
// It is used by the Kubernetes API to return multiple PulsarNSIsolationPolicy objects.
type PulsarNSIsolationPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PulsarNSIsolationPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PulsarNSIsolationPolicy{}, &PulsarNSIsolationPolicyList{})
}
