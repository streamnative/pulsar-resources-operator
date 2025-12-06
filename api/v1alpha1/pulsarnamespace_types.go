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

	adminutils "github.com/apache/pulsar-client-go/pulsaradmin/pkg/utils"
	"github.com/streamnative/pulsar-resources-operator/pkg/utils"
)

// TopicAutoCreationConfig defines the configuration for automatic topic creation
type TopicAutoCreationConfig struct {
	// Allow specifies whether to allow automatic topic creation
	Allow bool `json:"allow,omitempty"`

	// Type specifies the type of automatically created topics
	// +kubebuilder:validation:Enum=partitioned;non-partitioned
	Type string `json:"type,omitempty"`

	// Partitions specifies the default number of partitions for automatically created topics
	// +optional
	Partitions *int32 `json:"partitions,omitempty"`
}

// DispatchRate defines the message dispatch rate limiting configuration
type DispatchRate struct {
	// DispatchThrottlingRateInMsg specifies the maximum number of messages per second allowed
	// -1 means unlimited
	// +optional
	DispatchThrottlingRateInMsg *int32 `json:"dispatchThrottlingRateInMsg,omitempty"`

	// DispatchThrottlingRateInByte specifies the maximum number of bytes per second allowed
	// -1 means unlimited
	// +optional
	DispatchThrottlingRateInByte *int64 `json:"dispatchThrottlingRateInByte,omitempty"`

	// RatePeriodInSecond specifies the time window in seconds for rate calculation
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	// +optional
	RatePeriodInSecond *int32 `json:"ratePeriodInSecond,omitempty"`
}

// PublishRate defines the publish rate limiting configuration
type PublishRate struct {
	// PublishThrottlingRateInMsg specifies the maximum number of messages per second that producers can publish
	// -1 means unlimited
	// +optional
	PublishThrottlingRateInMsg *int32 `json:"publishThrottlingRateInMsg,omitempty"`

	// PublishThrottlingRateInByte specifies the maximum number of bytes per second that producers can publish
	// -1 means unlimited
	// +optional
	PublishThrottlingRateInByte *int64 `json:"publishThrottlingRateInByte,omitempty"`
}

// SubscribeRate defines the subscription rate limiting configuration
type SubscribeRate struct {
	// SubscribeThrottlingRatePerConsumer specifies the maximum subscribe rate per consumer
	// -1 means unlimited
	// +optional
	SubscribeThrottlingRatePerConsumer *int32 `json:"subscribeThrottlingRatePerConsumer,omitempty"`

	// RatePeriodInSecond specifies the time window in seconds for rate calculation
	// +kubebuilder:default=30
	// +kubebuilder:validation:Minimum=1
	// +optional
	RatePeriodInSecond *int32 `json:"ratePeriodInSecond,omitempty"`
}

// PersistencePolicies defines the persistence configuration for BookKeeper
type PersistencePolicies struct {
	// BookkeeperEnsemble specifies the number of bookies to use for a ledger
	// This determines the replication factor for storing data
	// +kubebuilder:validation:Minimum=1
	// +optional
	BookkeeperEnsemble *int32 `json:"bookkeeperEnsemble,omitempty"`

	// BookkeeperWriteQuorum specifies the number of replicas to write for each entry
	// Must be <= BookkeeperEnsemble
	// +kubebuilder:validation:Minimum=1
	// +optional
	BookkeeperWriteQuorum *int32 `json:"bookkeeperWriteQuorum,omitempty"`

	// BookkeeperAckQuorum specifies the number of replicas to wait for acknowledgment
	// Must be <= BookkeeperWriteQuorum
	// +kubebuilder:validation:Minimum=1
	// +optional
	BookkeeperAckQuorum *int32 `json:"bookkeeperAckQuorum,omitempty"`

	// ManagedLedgerMaxMarkDeleteRate specifies the maximum rate at which mark-delete operations can be performed
	// This helps control the rate of acknowledgment processing
	// Value should be a decimal number as string (e.g., "1.5", "2.0")
	// +optional
	ManagedLedgerMaxMarkDeleteRate *string `json:"managedLedgerMaxMarkDeleteRate,omitempty"`
}

// InactiveTopicPolicies defines the policies for inactive topic management
type InactiveTopicPolicies struct {
	// InactiveTopicDeleteMode specifies how inactive topics should be handled
	// Valid values: "delete_when_no_subscriptions", "delete_when_subscriptions_caught_up"
	// +kubebuilder:validation:Enum=delete_when_no_subscriptions;delete_when_subscriptions_caught_up
	// +optional
	InactiveTopicDeleteMode *string `json:"inactiveTopicDeleteMode,omitempty"`

	// MaxInactiveDurationInSeconds specifies how long a topic can remain inactive before being deleted
	// +kubebuilder:validation:Minimum=1
	// +optional
	MaxInactiveDurationInSeconds *int32 `json:"maxInactiveDurationInSeconds,omitempty"`

	// DeleteWhileInactive specifies whether to delete topics while they are inactive
	// +optional
	DeleteWhileInactive *bool `json:"deleteWhileInactive,omitempty"`
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PulsarNamespaceSpec defines the desired state of a Pulsar namespace.
// It corresponds to the configuration options available in Pulsar's namespace admin API.
type PulsarNamespaceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// TODO make these fields immutable

	// Name is the fully qualified namespace name in the format "tenant/namespace".
	Name string `json:"name"`

	// Bundles specifies the number of bundles to split the namespace into.
	// This affects how the namespace is distributed across the cluster.
	Bundles *int32 `json:"bundles,omitempty"`

	// ConnectionRef is the reference to the PulsarConnection resource
	// used to connect to the Pulsar cluster for this namespace.
	ConnectionRef corev1.LocalObjectReference `json:"connectionRef"`

	// LifecyclePolicy determines whether to keep or delete the Pulsar namespace
	// when the Kubernetes resource is deleted.
	// +kubebuilder:validation:Enum=CleanUpAfterDeletion;KeepAfterDeletion
	// +optional
	LifecyclePolicy PulsarResourceLifeCyclePolicy `json:"lifecyclePolicy,omitempty"`

	// SchemaCompatibilityStrategy defines the schema compatibility strategy for this namespace.
	// If not specified, the cluster's default schema compatibility strategy will be used.
	// This setting controls how schema evolution is handled for topics within this namespace.
	// +optional
	// +kubebuilder:validation:Enum=UNDEFINED;ALWAYS_INCOMPATIBLE;ALWAYS_COMPATIBLE;BACKWARD;FORWARD;FULL;BACKWARD_TRANSITIVE;FORWARD_TRANSITIVE;FULL_TRANSITIVE
	SchemaCompatibilityStrategy *SchemaCompatibilityStrategy `json:"schemaCompatibilityStrategy,omitempty"`

	// SchemaValidationEnforced controls whether schema validation is enforced for this namespace.
	// When enabled, producers must provide a schema when publishing messages.
	// If not specified, the cluster's default schema validation enforcement setting will be used.
	// +optional
	SchemaValidationEnforced *bool `json:"schemaValidationEnforced,omitempty"`

	// MaxProducersPerTopic sets the maximum number of producers allowed on a single topic in the namespace.
	// +optional
	MaxProducersPerTopic *int32 `json:"maxProducersPerTopic,omitempty"`

	// MaxConsumersPerTopic sets the maximum number of consumers allowed on a single topic in the namespace.
	// +optional
	MaxConsumersPerTopic *int32 `json:"maxConsumersPerTopic,omitempty"`

	// MaxConsumersPerSubscription sets the maximum number of consumers allowed on a single subscription in the namespace.
	// +optional
	MaxConsumersPerSubscription *int32 `json:"maxConsumersPerSubscription,omitempty"`

	// MessageTTL specifies the Time to Live (TTL) for messages in the namespace.
	// Messages older than this TTL will be automatically marked as consumed.
	// +optional
	MessageTTL *utils.Duration `json:"messageTTL,omitempty"`

	// RetentionTime specifies the minimum time to retain messages in the namespace.
	// Should be set in conjunction with RetentionSize for effective retention policy.
	// Retention Quota must exceed configured backlog quota for namespace
	// +optional
	RetentionTime *utils.Duration `json:"retentionTime,omitempty"`

	// RetentionSize specifies the maximum size of backlog retained in the namespace.
	// Should be set in conjunction with RetentionTime for effective retention policy.
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

	// OffloadThresholdTime specifies the time limit for message offloading.
	// Messages older than this limit will be offloaded to the tiered storage.
	// +optional
	OffloadThresholdTime *utils.Duration `json:"offloadThresholdTime,omitempty"`

	// OffloadThresholdSize specifies the size limit for message offloading.
	// When the limit is reached, older messages will be offloaded to the tiered storage.
	// +optional
	OffloadThresholdSize *resource.Quantity `json:"offloadThresholdSize,omitempty"`

	// GeoReplicationRefs is a list of references to PulsarGeoReplication resources,
	// used to configure geo-replication for this namespace.
	// This is **ONLY** used when you are using PulsarGeoReplication for setting up geo-replication
	// between two Pulsar instances.
	// Please use `ReplicationClusters` instead if you are replicating clusters within the same Pulsar instance.
	// +optional
	GeoReplicationRefs []*corev1.LocalObjectReference `json:"geoReplicationRefs,omitempty"`

	// ReplicationClusters is the list of clusters to which the namespace is replicated
	// This is **ONLY** used if you are replicating clusters within the same Pulsar instance.
	// Please use `GeoReplicationRefs` instead if you are setting up geo-replication
	// between two Pulsar instances.
	// +optional
	ReplicationClusters []string `json:"replicationClusters,omitempty"`

	// Deduplication controls whether to enable message deduplication for the namespace.
	// +optional
	Deduplication *bool `json:"deduplication,omitempty"`

	// BookieAffinityGroup is the name of the namespace isolation policy to apply to the namespace.
	BookieAffinityGroup *BookieAffinityGroupData `json:"bookieAffinityGroup,omitempty"`

	// TopicAutoCreationConfig controls whether automatic topic creation is allowed in this namespace
	// and configures properties of automatically created topics
	// +optional
	TopicAutoCreationConfig *TopicAutoCreationConfig `json:"topicAutoCreationConfig,omitempty"`

	// DispatchRate sets the message dispatch rate for topics in this namespace.
	// This controls how fast messages are delivered to consumers.
	// +optional
	DispatchRate *DispatchRate `json:"dispatchRate,omitempty"`

	// SubscriptionDispatchRate sets the subscription dispatch rate for topics in this namespace.
	// This controls the rate at which subscriptions can dispatch messages.
	// +optional
	SubscriptionDispatchRate *DispatchRate `json:"subscriptionDispatchRate,omitempty"`

	// ReplicatorDispatchRate sets the replicator dispatch rate for topics in this namespace.
	// This controls the rate at which messages are replicated across clusters.
	// +optional
	ReplicatorDispatchRate *DispatchRate `json:"replicatorDispatchRate,omitempty"`

	// PublishRate sets the publish rate limit for producers in this namespace.
	// This controls how fast producers can publish messages.
	// +optional
	PublishRate *PublishRate `json:"publishRate,omitempty"`

	// SubscribeRate sets the subscribe rate limit for consumers in this namespace.
	// This controls how fast consumers can subscribe to topics.
	// +optional
	SubscribeRate *SubscribeRate `json:"subscribeRate,omitempty"`

	// PersistencePolicies sets the persistence policies for this namespace.
	// This controls how data is stored and replicated in BookKeeper.
	// +optional
	PersistencePolicies *PersistencePolicies `json:"persistencePolicies,omitempty"`

	// CompactionThreshold sets the compaction threshold for topics in this namespace.
	// Topics will be compacted when they reach this size threshold (in bytes).
	// +optional
	CompactionThreshold *int64 `json:"compactionThreshold,omitempty"`

	// InactiveTopicPolicies sets the policies for handling inactive topics in this namespace.
	// This controls automatic cleanup of unused topics.
	// +optional
	InactiveTopicPolicies *InactiveTopicPolicies `json:"inactiveTopicPolicies,omitempty"`

	// SubscriptionExpirationTime specifies the time after which inactive subscriptions will expire.
	// Subscriptions that haven't consumed messages for this duration will be automatically deleted.
	// +optional
	SubscriptionExpirationTime *utils.Duration `json:"subscriptionExpirationTime,omitempty"`

	// Properties is a map of custom properties for this namespace.
	// These are arbitrary key-value pairs that can be used for namespace metadata.
	// +optional
	Properties map[string]string `json:"properties,omitempty"`

	// IsAllowAutoUpdateSchema specifies whether to allow automatic schema updates.
	// When enabled, producers can automatically update schemas without manual approval.
	// +optional
	IsAllowAutoUpdateSchema *bool `json:"isAllowAutoUpdateSchema,omitempty"`

	// ValidateProducerName specifies whether to validate producer names.
	// When enabled, producer names must follow specific naming conventions.
	// +optional
	ValidateProducerName *bool `json:"validateProducerName,omitempty"`

	// EncryptionRequired specifies whether message encryption is required for this namespace.
	// When enabled, all messages published to topics in this namespace must be encrypted.
	// +optional
	EncryptionRequired *bool `json:"encryptionRequired,omitempty"`

	// SubscriptionAuthMode specifies the subscription authentication mode for this namespace.
	// Valid values are "None" and "Prefix".
	// +kubebuilder:validation:Enum=None;Prefix
	// +optional
	SubscriptionAuthMode *adminutils.SubscriptionAuthMode `json:"subscriptionAuthMode,omitempty"`

	// AntiAffinityGroup specifies the anti-affinity group for this namespace.
	// Namespaces in the same anti-affinity group will be placed on different brokers.
	// +optional
	AntiAffinityGroup *string `json:"antiAffinityGroup,omitempty"`

	// SchemaAutoUpdateCompatibilityStrategy specifies the compatibility strategy for automatic schema updates.
	// This controls how schema evolution is handled when schemas are automatically updated.
	// +optional
	// +kubebuilder:validation:Enum=AutoUpdateDisabled;Backward;Forward;Full;AlwaysCompatible;BackwardTransitive;ForwardTransitive;FullTransitive
	SchemaAutoUpdateCompatibilityStrategy *adminutils.SchemaAutoUpdateCompatibilityStrategy `json:"schemaAutoUpdateCompatibilityStrategy,omitempty"`
}

type BookieAffinityGroupData struct {
	BookkeeperAffinityGroupPrimary string `json:"bookkeeperAffinityGroupPrimary"`

	// +optional
	BookkeeperAffinityGroupSecondary string `json:"bookkeeperAffinityGroupSecondary,omitempty"`
}

// PulsarNamespaceStatus defines the observed state of PulsarNamespace
type PulsarNamespaceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ObservedGeneration is the most recent generation observed for this resource.
	// It corresponds to the metadata generation, which is updated on mutation by the API Server.
	// This field is used to track whether the controller has processed the latest changes.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions represent the latest available observations of the namespace's current state.
	// It follows the Kubernetes conventions for condition types and status.
	// The "Ready" condition type is typically used to indicate the overall status of the namespace.
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// GeoReplicationEnabled indicates whether geo-replication between two Pulsar instances (via PulsarGeoReplication)
	// is enabled for the namespace
	// +optional
	GeoReplicationEnabled bool `json:"geoReplicationEnabled,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:categories=pulsar;pulsarres,shortName=pns
//+kubebuilder:printcolumn:name="RESOURCE_NAME",type=string,JSONPath=`.spec.name`
//+kubebuilder:printcolumn:name="GENERATION",type=string,JSONPath=`.metadata.generation`
//+kubebuilder:printcolumn:name="OBSERVED_GENERATION",type=string,JSONPath=`.status.observedGeneration`
//+kubebuilder:printcolumn:name="READY",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`

// PulsarNamespace is the Schema for the pulsarnamespaces API
// It represents a Pulsar namespace in the Kubernetes cluster and includes both
// the desired state (Spec) and the observed state (Status) of the namespace.
type PulsarNamespace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PulsarNamespaceSpec   `json:"spec,omitempty"`
	Status PulsarNamespaceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PulsarNamespaceList contains a list of PulsarNamespace resources.
// It is used by the Kubernetes API to return multiple PulsarNamespace objects.
type PulsarNamespaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PulsarNamespace `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PulsarNamespace{}, &PulsarNamespaceList{})
}
