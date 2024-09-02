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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/streamnative/pulsar-resources-operator/pkg/utils"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PulsarTopicSpec defines the desired state of PulsarTopic.
// It corresponds to the configuration options available in Pulsar's topic admin API.
type PulsarTopicSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// TODO make these fields immutable

	// Name is the topic name
	Name string `json:"name"`

	// Persistent determines if the topic is persistent (true) or non-persistent (false).
	// Defaults to true if not specified.
	// +kubebuilder:default=true
	// +optional
	Persistent *bool `json:"persistent,omitempty"`

	// Partitions specifies the number of partitions for a partitioned topic.
	// Set to 0 for a non-partitioned topic.
	// +kubebuilder:default=0
	// +optional
	Partitions *int32 `json:"partitions,omitempty"`

	// ConnectionRef is the reference to the PulsarConnection resource
	// used to connect to the Pulsar cluster for this topic.
	ConnectionRef corev1.LocalObjectReference `json:"connectionRef"`

	// LifecyclePolicy determines whether to keep or delete the Pulsar topic
	// when the Kubernetes resource is deleted.
	// +kubebuilder:validation:Enum=CleanUpAfterDeletion;KeepAfterDeletion
	// +optional
	LifecyclePolicy PulsarResourceLifeCyclePolicy `json:"lifecyclePolicy,omitempty"`

	// Topic Policy Setting

	// MaxProducers sets the maximum number of producers allowed on the topic.
	// +optional
	MaxProducers *int32 `json:"maxProducers,omitempty"`

	// MaxConsumers sets the maximum number of consumers allowed on the topic.
	// +optional
	MaxConsumers *int32 `json:"maxConsumers,omitempty"`

	// MessageTTL specifies the Time to Live (TTL) for messages on the topic.
	// Messages older than this TTL will be automatically marked as deleted.
	// +optional
	MessageTTL *utils.Duration `json:"messageTTL,omitempty"`

	// MaxUnAckedMessagesPerConsumer sets the maximum number of unacknowledged
	// messages allowed for a consumer before it's blocked from receiving more messages.
	// +optional
	MaxUnAckedMessagesPerConsumer *int32 `json:"maxUnAckedMessagesPerConsumer,omitempty"`

	// MaxUnAckedMessagesPerSubscription sets the maximum number of unacknowledged
	// messages allowed for a subscription before it's blocked from receiving more messages.
	// +optional
	MaxUnAckedMessagesPerSubscription *int32 `json:"maxUnAckedMessagesPerSubscription,omitempty"`

	// RetentionTime specifies the minimum time to retain messages on the topic.
	// Should be set in conjunction with RetentionSize for effective retention policy.
	// Retention Quota must exceed configured backlog quota for topic
	// +optional
	RetentionTime *utils.Duration `json:"retentionTime,omitempty"`

	// RetentionSize specifies the maximum size of backlog retained on the topic.
	// Should be set in conjunction with RetentionTime for effective retention policy.
	// Retention Quota must exceed configured backlog quota for topic
	// +optional
	RetentionSize *resource.Quantity `json:"retentionSize,omitempty"`

	// BacklogQuotaLimitTime specifies the time limit for message backlog.
	// Messages older than this limit will be removed or handled according to the retention policy.
	// +optional
	BacklogQuotaLimitTime *utils.Duration `json:"backlogQuotaLimitTime,omitempty"`

	// BacklogQuotaLimitSize specifies the size limit for message backlog.
	// When the limit is reached, older messages will be removed or handled according to the retention policy.
	// +optional
	BacklogQuotaLimitSize *resource.Quantity `json:"backlogQuotaLimitSize,omitempty"`

	// BacklogQuotaRetentionPolicy specifies the retention policy for messages when backlog quota is exceeded.
	// Valid values are "producer_request_hold", "producer_exception", or "consumer_backlog_eviction".
	// +optional
	BacklogQuotaRetentionPolicy *string `json:"backlogQuotaRetentionPolicy,omitempty"`

	// SchemaInfo defines the schema for the topic, if any.
	// +optional
	SchemaInfo *SchemaInfo `json:"schemaInfo,omitempty"`

	// GeoReplicationRefs is a list of references to PulsarGeoReplication resources,
	// used to configure geo-replication for this topic across multiple Pulsar instances.
	// This is **ONLY** used when you are using PulsarGeoReplication for setting up geo-replication
	// between two Pulsar instances.
	// +optional
	GeoReplicationRefs []*corev1.LocalObjectReference `json:"geoReplicationRefs,omitempty"`

	// ReplicationClusters is the list of clusters to which the topic is replicated
	// This is **ONLY** used if you are replicating clusters within the same Pulsar instance.
	// Please use `GeoReplicationRefs` instead if you are setting up geo-replication
	// between two Pulsar instances.
	// +optional
	ReplicationClusters []string `json:"replicationClusters,omitempty"`
}

// SchemaInfo defines the Pulsar Schema for a topic.
// It is stored and enforced on a per-topic basis and cannot be stored at the namespace or tenant level.
type SchemaInfo struct {
	// Type determines how to interpret the schema data.
	// Valid values include: "AVRO", "JSON", "PROTOBUF", "PROTOBUF_NATIVE", "KEY_VALUE", "BYTES", or "NONE".
	// For KEY_VALUE schemas, use the format "KEY_VALUE(KeyType,ValueType)" where KeyType and ValueType
	// are one of the other schema types.
	Type string `json:"type,omitempty"`

	// Schema contains the actual schema definition.
	// For AVRO and JSON schemas, this should be a JSON string of the schema definition.
	// For PROTOBUF schemas, this should be the protobuf definition string.
	// For BYTES or NONE schemas, this field can be empty.
	Schema string `json:"schema,omitempty"`

	// Properties is a map of user-defined properties associated with the schema.
	// These can be used to store additional metadata about the schema.
	Properties map[string]string `json:"properties,omitempty"`
}

// PulsarTopicStatus defines the observed state of PulsarTopic
type PulsarTopicStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ObservedGeneration is the most recent generation observed for this resource.
	// It corresponds to the metadata generation, which is updated on mutation by the API Server.
	// This field is used to track whether the controller has processed the latest changes.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions represent the latest available observations of the PulsarTopic's current state.
	// It follows the Kubernetes conventions for condition types and status.
	// The "Ready" condition type indicates the overall status of the topic.
	// The "PolicyReady" condition type indicates whether the topic policies have been successfully applied.
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// GeoReplicationEnabled indicates whether geo-replication is enabled for this topic.
	// This is set to true when GeoReplicationRefs are configured in the spec and successfully applied.
	// +optional
	GeoReplicationEnabled bool `json:"geoReplicationEnabled,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:categories=pulsar;pulsarres,shortName=ptopic
//+kubebuilder:printcolumn:name="RESOURCE_NAME",type=string,JSONPath=`.spec.name`
//+kubebuilder:printcolumn:name="GENERATION",type=string,JSONPath=`.metadata.generation`
//+kubebuilder:printcolumn:name="OBSERVED_GENERATION",type=string,JSONPath=`.status.observedGeneration`
//+kubebuilder:printcolumn:name="READY",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
//+kubebuilder:printcolumn:name="POLICY_READY",type=string,JSONPath=`.status.conditions[?(@.type=="PolicyReady")].status`

// PulsarTopic is the Schema for the pulsartopics API
// It represents a Pulsar topic in the Kubernetes cluster and includes both
// the desired state (Spec) and the observed state (Status) of the topic.
type PulsarTopic struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PulsarTopicSpec   `json:"spec,omitempty"`
	Status PulsarTopicStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PulsarTopicList contains a list of PulsarTopic resources.
// It is used by the Kubernetes API to return multiple PulsarTopic objects.
type PulsarTopicList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PulsarTopic `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PulsarTopic{}, &PulsarTopicList{})
}
