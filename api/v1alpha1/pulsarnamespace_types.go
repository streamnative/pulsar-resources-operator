// Copyright (c) 2022 StreamNative, Inc.. All Rights Reserved.

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PulsarNamespaceSpec defines the desired state of PulsarNamespace
type PulsarNamespaceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// TODO make these fields immutable

	// Name is the namespace name
	Name string `json:"name"`

	Bundles *int32 `json:"bundles,omitempty"`

	// ConnectionRef is the reference to the PulsarConnection resource
	ConnectionRef corev1.LocalObjectReference `json:"connectionRef"`

	// +kubebuilder:validation:Enum=CleanUpAfterDeletion;KeepAfterDeletion
	// +optional
	LifecyclePolicy PulsarResourceLifeCyclePolicy `json:"lifecyclePolicy,omitempty"`

	// Tenant Policy Setting
	// +optional
	MaxProducersPerTopic *int32 `json:"maxProducersPerTopic,omitempty"`

	// +optional
	MaxConsumersPerTopic *int32 `json:"maxConsumersPerTopic,omitempty"`

	// +optional
	MaxConsumersPerSubscription *int32 `json:"maxConsumersPerSubscription,omitempty"`

	// +optional
	MessageTTL *metav1.Duration `json:"messageTTL,omitempty"`

	// Retention
	// Should set at least one of them if setting retention
	// +optional
	RetentionTime *metav1.Duration `json:"retentionTime,omitempty"`

	// +optional
	RetentionSize *resource.Quantity `json:"retentionSize,omitempty"`

	// Backlog
	// Should set at least one of them if setting backlog
	// +optional
	BacklogQuotaLimitTime *metav1.Duration `json:"backlogQuotaLimitTime,omitempty"`

	// +optional
	BacklogQuotaLimitSize *resource.Quantity `json:"backlogQuotaLimitSize,omitempty"`

	// +optional
	BacklogQuotaRetentionPolicy *string `json:"backlogQuotaRetentionPolicy,omitempty"`
}

// PulsarNamespaceStatus defines the observed state of PulsarNamespace
type PulsarNamespaceStatus struct {
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
//+kubebuilder:resource:categories=pulsar;pulsarres,shortName=pns
//+kubebuilder:printcolumn:name="RESOURCE_NAME",type=string,JSONPath=`.spec.name`
//+kubebuilder:printcolumn:name="GENERATION",type=string,JSONPath=`.metadata.generation`
//+kubebuilder:printcolumn:name="OBSERVED_GENERATION",type=string,JSONPath=`.status.observedGeneration`
//+kubebuilder:printcolumn:name="READY",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`

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
