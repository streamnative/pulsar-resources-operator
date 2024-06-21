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
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PulsarSinkSpec defines the desired state of PulsarSink
type PulsarSinkSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// TopicsPattern is the pattern of topics to consume from Pulsar
	// +optional
	TopicsPattern *string `json:"topicsPattern,omitempty"`

	// Resources is the resource requirements for the PulsarSink
	// +optional
	Resources *Resources `json:"resources,omitempty"`

	// TimeoutMs is the timeout in milliseconds for the PulsarSink
	// +optional
	TimeoutMs *int64 `json:"timeoutMs,omitempty"`

	// CleanupSubscription is the flag to enable or disable the cleanup of subscription
	// +optional
	CleanupSubscription bool `json:"cleanupSubscription,omitempty"`

	// RetainOrdering is the flag to enable or disable the retain ordering
	// +optional
	RetainOrdering bool `json:"retainOrdering,omitempty"`

	// RetainKeyOrdering is the flag to enable or disable the retain key ordering
	// +optional
	RetainKeyOrdering bool `json:"retainKeyOrdering,omitempty"`

	// AutoAck is the flag to enable or disable the auto ack
	// +optional
	AutoAck bool `json:"autoAck,omitempty"`

	// Parallelism is the parallelism of the PulsarSink
	// +optional
	Parallelism int `json:"parallelism,omitempty"`

	// Tenant is the tenant of the PulsarSink
	// +optional
	Tenant string `json:"tenant,omitempty"`

	// Namespace is the namespace of the PulsarSink
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Name is the name of the PulsarSink
	// +optional
	Name string `json:"name,omitempty"`

	// ClassName is the class name of the PulsarSink
	// +optional
	ClassName string `json:"className,omitempty"`

	// SinkType is the type of the PulsarSink
	// +optional
	SinkType string `json:"sinkType,omitempty"`

	// Archive is the archive of the PulsarSink
	// +optional
	Archive *PackageContentRef `json:"archive,omitempty"`

	// ProcessingGuarantees is the processing guarantees of the PulsarSink
	// +optional
	ProcessingGuarantees string `json:"processingGuarantees,omitempty"`

	// SourceSubscriptionName is the source subscription name of the PulsarSink
	// +optional
	SourceSubscriptionName string `json:"sourceSubscriptionName,omitempty"`

	// SourceSubscriptionPosition is the source subscription position of the PulsarSink
	// +optional
	SourceSubscriptionPosition string `json:"sourceSubscriptionPosition,omitempty"`

	// RuntimeFlags is the runtime flags of the PulsarSink
	// +optional
	RuntimeFlags string `json:"runtimeFlags,omitempty"`

	// Inputs is the list of inputs of the PulsarSink
	// +optional
	Inputs []string `json:"inputs,omitempty"`

	// TopicToSerdeClassName is the map of topic to serde class name of the PulsarSink
	// +optional
	TopicToSerdeClassName map[string]string `json:"topicToSerdeClassName,omitempty"`

	// TopicToSchemaType is the map of topic to schema type of the PulsarSink
	// +optional
	TopicToSchemaType map[string]string `json:"topicToSchemaType,omitempty"`

	// InputSpecs is the map of input specs of the PulsarSink
	// +optional
	InputSpecs map[string]ConsumerConfig `json:"inputSpecs,omitempty"`

	// Configs is the map of configs of the PulsarSink
	// +optional
	Configs *apiextensionsv1.JSON `json:"configs,omitempty"`

	// TopicToSchemaProperties is the map of topic to schema properties of the PulsarSink
	// +optional
	TopicToSchemaProperties map[string]string `json:"topicToSchemaProperties,omitempty"`

	// CustomRuntimeOptions is the custom runtime options of the PulsarSink
	// +optional
	CustomRuntimeOptions *apiextensionsv1.JSON `json:"customRuntimeOptions,omitempty"`

	// Secrets is the map of secrets of the PulsarSink
	// +optional
	Secrets map[string]SecretKeyRef `json:"secrets,omitempty"`

	// MaxMessageRetries is the max message retries of the PulsarSink
	// +optional
	MaxMessageRetries int `json:"maxMessageRetries,omitempty"`

	// DeadLetterTopic is the dead letter topic of the PulsarSink
	// +optional
	DeadLetterTopic string `json:"deadLetterTopic,omitempty"`

	// NegativeAckRedeliveryDelayMs is the negative ack redelivery delay in milliseconds of the PulsarSink
	// +optional
	NegativeAckRedeliveryDelayMs int64 `json:"negativeAckRedeliveryDelayMs,omitempty"`

	// TransformFunction is the transform function of the PulsarSink
	// +optional
	TransformFunction string `json:"transformFunction,omitempty"`

	// TransformFunctionClassName is the transform function class name of the PulsarSink
	// +optional
	TransformFunctionClassName string `json:"transformFunctionClassName,omitempty"`

	// TransformFunctionConfig is the transform function config of the PulsarSink
	// +optional
	TransformFunctionConfig string `json:"transformFunctionConfig,omitempty"`

	// ConnectionRef is the reference to the PulsarConnection resource
	ConnectionRef corev1.LocalObjectReference `json:"connectionRef"`

	// +kubebuilder:validation:Enum=CleanUpAfterDeletion;KeepAfterDeletion
	// +optional
	LifecyclePolicy PulsarResourceLifeCyclePolicy `json:"lifecyclePolicy,omitempty"`
}

// PulsarSinkStatus defines the observed state of PulsarSink
type PulsarSinkStatus struct {
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
//+kubebuilder:resource:categories=pulsar;pulsarres,shortName=psink
//+kubebuilder:printcolumn:name="RESOURCE_NAME",type=string,JSONPath=`.spec.name`
//+kubebuilder:printcolumn:name="GENERATION",type=string,JSONPath=`.metadata.generation`
//+kubebuilder:printcolumn:name="OBSERVED_GENERATION",type=string,JSONPath=`.status.observedGeneration`
//+kubebuilder:printcolumn:name="READY",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`

// PulsarSink is the Schema for the pulsar functions API
type PulsarSink struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PulsarSinkSpec   `json:"spec,omitempty"`
	Status PulsarSinkStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PulsarSinkList contains a list of PulsarSink
type PulsarSinkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PulsarSink `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PulsarSink{}, &PulsarSinkList{})
}
