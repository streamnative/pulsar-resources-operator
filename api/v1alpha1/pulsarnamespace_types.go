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

	"github.com/streamnative/pulsar-resources-operator/pkg/utils"
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

	// MessageTTL indicates the message ttl for the namespace
	// +optional
	MessageTTL *utils.Duration `json:"messageTTL,omitempty"`

	// Retention
	// Should set at least one of them if setting retention
	// Retention Quota must exceed configured backlog quota for namespace
	// +optional
	RetentionTime *utils.Duration `json:"retentionTime,omitempty"`

	// +optional
	RetentionSize *resource.Quantity `json:"retentionSize,omitempty"`

	// Backlog
	// Should set at least one of them if setting backlog
	// +optional
	BacklogQuotaLimitTime *utils.Duration `json:"backlogQuotaLimitTime,omitempty"`

	// +optional
	BacklogQuotaLimitSize *resource.Quantity `json:"backlogQuotaLimitSize,omitempty"`

	// +optional
	BacklogQuotaRetentionPolicy *string `json:"backlogQuotaRetentionPolicy,omitempty"`

	// BacklogQuotaType controls the backlog by setting the type to destination_storage or message_age
	// destination_storage limits backlog by size (in bytes). message_age limits backlog by time,
	// that is, message timestamp (broker or publish timestamp)
	// +kubebuilder:validation:Enum=destination_storage;message_age
	// +optional
	BacklogQuotaType *string `json:"backlogQuotaType,omitempty"`

	// GeoReplicationRefs is the reference list to the PulsarGeoReplication resource
	// +optional
	GeoReplicationRefs []*corev1.LocalObjectReference `json:"geoReplicationRefs,omitempty"`

	// ReplicationClusters
	// +optional
	ReplicationClusters []string `json:"replicationClusters,omitempty"`
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

	// GeoReplicationEnabled
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
