// Copyright 2025 StreamNative
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
// +kubebuilder:validation:XValidation:rule="!(has(self.backlogQuotaLimitSize) || has(self.backlogQuotaLimitTime) || has(self.backlogQuotaType)) || has(self.backlogQuotaRetentionPolicy)",message="backlogQuotaRetentionPolicy is required when configuring backlog quota"
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
	// Retention Quota must exceed configured backlog quota for topic.
	// Use "-1" for infinite retention time.
	// Valid formats: "1h", "30m", "5s", "-1"
	// +optional
	RetentionTime *utils.Duration `json:"retentionTime,omitempty"`

	// RetentionSize specifies the maximum size of backlog retained on the topic.
	// Should be set in conjunction with RetentionTime for effective retention policy.
	// Retention Quota must exceed configured backlog quota for topic.
	// Use "-1" for infinite retention size.
	// Valid formats: "1Gi", "500Mi", "100M", "-1"
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
	// +kubebuilder:validation:Enum=producer_request_hold;producer_exception;consumer_backlog_eviction
	// +optional
	BacklogQuotaRetentionPolicy *string `json:"backlogQuotaRetentionPolicy,omitempty"`

	// BacklogQuotaType controls how the backlog quota is enforced.
	// "destination_storage" limits backlog by size (in bytes), while "message_age" limits by time.
	// +kubebuilder:validation:Enum=destination_storage;message_age
	// +optional
	BacklogQuotaType *string `json:"backlogQuotaType,omitempty"`

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

	// Deduplication controls whether to enable message deduplication for the topic.
	// +optional
	Deduplication *bool `json:"deduplication,omitempty"`

	// CompactionThreshold specifies the size threshold in bytes for automatic topic compaction.
	// When the topic reaches this size, compaction will be triggered automatically.
	// +optional
	CompactionThreshold *int64 `json:"compactionThreshold,omitempty"`

	// PersistencePolicies defines the persistence configuration for the topic.
	// This controls how data is stored and replicated in BookKeeper.
	// +optional
	PersistencePolicies *PersistencePolicies `json:"persistencePolicies,omitempty"`

	// DelayedDelivery defines the delayed delivery policy for the topic.
	// This allows messages to be delivered with a delay.
	// +optional
	DelayedDelivery *DelayedDeliveryData `json:"delayedDelivery,omitempty"`

	// DispatchRate defines the message dispatch rate limiting policy for the topic.
	// This controls the rate at which messages are delivered to consumers.
	// +optional
	DispatchRate *DispatchRate `json:"dispatchRate,omitempty"`

	// PublishRate defines the message publish rate limiting policy for the topic.
	// This controls the rate at which producers can publish messages.
	// +optional
	PublishRate *PublishRate `json:"publishRate,omitempty"`

	// InactiveTopicPolicies defines the inactive topic cleanup policy for the topic.
	// This controls how inactive topics are automatically cleaned up.
	// +optional
	InactiveTopicPolicies *InactiveTopicPolicies `json:"inactiveTopicPolicies,omitempty"`

	// SubscribeRate defines the subscription rate limiting policy for the topic.
	// This controls the rate at which new subscriptions can be created.
	// +optional
	SubscribeRate *SubscribeRate `json:"subscribeRate,omitempty"`

	// MaxMessageSize specifies the maximum size of messages that can be published to the topic.
	// Messages larger than this size will be rejected.
	// +optional
	MaxMessageSize *int32 `json:"maxMessageSize,omitempty"`

	// MaxConsumersPerSubscription sets the maximum number of consumers allowed per subscription.
	// +optional
	MaxConsumersPerSubscription *int32 `json:"maxConsumersPerSubscription,omitempty"`

	// MaxSubscriptionsPerTopic sets the maximum number of subscriptions allowed on the topic.
	// +optional
	MaxSubscriptionsPerTopic *int32 `json:"maxSubscriptionsPerTopic,omitempty"`

	// SchemaValidationEnforced determines whether schema validation is enforced for the topic.
	// When enabled, only messages that conform to the topic's schema will be accepted.
	// +optional
	SchemaValidationEnforced *bool `json:"schemaValidationEnforced,omitempty"`

	// SubscriptionDispatchRate defines the message dispatch rate limiting policy for subscriptions.
	// This controls the rate at which messages are delivered to consumers per subscription.
	// +optional
	SubscriptionDispatchRate *DispatchRate `json:"subscriptionDispatchRate,omitempty"`

	// ReplicatorDispatchRate defines the message dispatch rate limiting policy for replicators.
	// This controls the rate at which messages are replicated to other clusters.
	// +optional
	ReplicatorDispatchRate *DispatchRate `json:"replicatorDispatchRate,omitempty"`

	// DeduplicationSnapshotInterval specifies the interval for taking deduplication snapshots.
	// This affects the deduplication performance and storage overhead.
	// +optional
	DeduplicationSnapshotInterval *int32 `json:"deduplicationSnapshotInterval,omitempty"`

	// OffloadPolicies defines the offload policies for the topic.
	// This controls how data is offloaded to external storage systems.
	// +optional
	OffloadPolicies *OffloadPolicies `json:"offloadPolicies,omitempty"`

	// AutoSubscriptionCreation defines the auto subscription creation override for the topic.
	// This controls whether subscriptions can be created automatically.
	// +optional
	AutoSubscriptionCreation *AutoSubscriptionCreationOverride `json:"autoSubscriptionCreation,omitempty"`

	// SchemaCompatibilityStrategy defines the schema compatibility strategy for the topic.
	// This controls how schema evolution is handled.
	// +optional
	// +kubebuilder:validation:Enum=UNDEFINED;ALWAYS_INCOMPATIBLE;ALWAYS_COMPATIBLE;BACKWARD;FORWARD;FULL;BACKWARD_TRANSITIVE;FORWARD_TRANSITIVE;FULL_TRANSITIVE
	SchemaCompatibilityStrategy *SchemaCompatibilityStrategy `json:"schemaCompatibilityStrategy,omitempty"`

	// Properties is a map of user-defined properties associated with the topic.
	// These can be used to store additional metadata about the topic.
	// +optional
	Properties map[string]string `json:"properties,omitempty"`
}

// DelayedDeliveryData defines the delayed delivery policy for a topic
type DelayedDeliveryData struct {
	// Active determines whether delayed delivery is enabled for the topic
	// +optional
	Active *bool `json:"active,omitempty"`

	// TickTimeMillis specifies the tick time for delayed message delivery in milliseconds
	// +optional
	TickTimeMillis *int64 `json:"tickTimeMillis,omitempty"`
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

// OffloadPolicies defines the offload policies for a topic.
// This is a local type definition that mirrors the external library's OffloadPolicies
// to ensure proper Kubernetes deep copy generation.
type OffloadPolicies struct {
	ManagedLedgerOffloadDriver                        string            `json:"managedLedgerOffloadDriver,omitempty"`
	ManagedLedgerOffloadMaxThreads                    int               `json:"managedLedgerOffloadMaxThreads,omitempty"`
	ManagedLedgerOffloadThresholdInBytes              int64             `json:"managedLedgerOffloadThresholdInBytes,omitempty"`
	ManagedLedgerOffloadDeletionLagInMillis           int64             `json:"managedLedgerOffloadDeletionLagInMillis,omitempty"`
	ManagedLedgerOffloadAutoTriggerSizeThresholdBytes int64             `json:"managedLedgerOffloadAutoTriggerSizeThresholdBytes,omitempty"`
	S3ManagedLedgerOffloadBucket                      string            `json:"s3ManagedLedgerOffloadBucket,omitempty"`
	S3ManagedLedgerOffloadRegion                      string            `json:"s3ManagedLedgerOffloadRegion,omitempty"`
	S3ManagedLedgerOffloadServiceEndpoint             string            `json:"s3ManagedLedgerOffloadServiceEndpoint,omitempty"`
	S3ManagedLedgerOffloadCredentialID                string            `json:"s3ManagedLedgerOffloadCredentialId,omitempty"`
	S3ManagedLedgerOffloadCredentialSecret            string            `json:"s3ManagedLedgerOffloadCredentialSecret,omitempty"`
	S3ManagedLedgerOffloadRole                        string            `json:"s3ManagedLedgerOffloadRole,omitempty"`
	S3ManagedLedgerOffloadRoleSessionName             string            `json:"s3ManagedLedgerOffloadRoleSessionName,omitempty"`
	OffloadersDirectory                               string            `json:"offloadersDirectory,omitempty"`
	ManagedLedgerOffloadDriverMetadata                map[string]string `json:"managedLedgerOffloadDriverMetadata,omitempty"`
}

// AutoSubscriptionCreationOverride defines the auto subscription creation override for a topic.
// This is a local type definition that mirrors the external library's AutoSubscriptionCreationOverride
// to ensure proper Kubernetes deep copy generation.
type AutoSubscriptionCreationOverride struct {
	AllowAutoSubscriptionCreation bool `json:"allowAutoSubscriptionCreation,omitempty"`
}

// SchemaCompatibilityStrategy defines the schema compatibility strategy for a topic.
// This is a local type definition that mirrors the external library's SchemaCompatibilityStrategy
// to ensure proper Kubernetes deep copy generation.
type SchemaCompatibilityStrategy string

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
