// Copyright 2022 StreamNative
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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PulsarTopicSpec defines the desired state of PulsarTopic
type PulsarTopicSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// TODO make these fields immutable

	// Name is the topic name
	Name string `json:"name"`

	// +kubebuilder:default=true
	// +optional
	Persistent *bool `json:"persistent,omitempty"`

	// +kubebuilder:default=0
	// +optional
	Partitions *int32 `json:"partitions,omitempty"`

	// ConnectionRef is the reference to the PulsarConnection resource
	ConnectionRef corev1.LocalObjectReference `json:"connectionRef"`

	// +kubebuilder:validation:Enum=CleanUpAfterDeletion;KeepAfterDeletion
	// +optional
	LifecyclePolicy PulsarResourceLifeCyclePolicy `json:"lifecyclePolicy,omitempty"`

	// Topic Policy Setting
	// +optional
	MaxProducers *int32 `json:"maxProducers,omitempty"`

	// +optional
	MaxConsumers *int32 `json:"maxConsumers,omitempty"`

	// +optional
	MessageTTL *metav1.Duration `json:"messageTTL,omitempty"`

	// Max unacked messages
	// +optional
	MaxUnAckedMessagesPerConsumer *int32 `json:"maxUnAckedMessagesPerConsumer,omitempty"`

	// +optional
	MaxUnAckedMessagesPerSubscription *int32 `json:"maxUnAckedMessagesPerSubscription,omitempty"`

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

	// +optional
	SchemaInfo *SchemaInfo `json:"schemaInfo,omitempty"`

	// GeoReplicationRef is the reference to the PulsarConnection resource
	// +optional
	GeoReplicationRef *corev1.LocalObjectReference `json:"geoReplicationRef,omitempty"`
}

// SchemaInfo defines the Pulsar Schema.
// It is stored and enforced on a per-topic basis and cannot be stored at the namespace or tenant level.
type SchemaInfo struct {
	// Type determines how to interpret the schema data
	Type string `json:"type,omitempty"`
	// Schema is schema data
	Schema string `json:"schema,omitempty"`
	// Properties is a user defined properties as a string/string map
	Properties map[string]string `json:"properties,omitempty"`
}

// PulsarTopicStatus defines the observed state of PulsarTopic
type PulsarTopicStatus struct {
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
//+kubebuilder:resource:categories=pulsar;pulsarres,shortName=ptopic
//+kubebuilder:printcolumn:name="RESOURCE_NAME",type=string,JSONPath=`.spec.name`
//+kubebuilder:printcolumn:name="GENERATION",type=string,JSONPath=`.metadata.generation`
//+kubebuilder:printcolumn:name="OBSERVED_GENERATION",type=string,JSONPath=`.status.observedGeneration`
//+kubebuilder:printcolumn:name="READY",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`

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
