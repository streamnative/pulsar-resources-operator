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

// PulsarFunctionSpec defines the desired state of PulsarFunction
type PulsarFunctionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// TimeoutMs is the function timeout in milliseconds
	// +optional
	TimeoutMs *int64 `json:"timeoutMs,omitempty"`

	// TopicsPattern is the topics pattern that the function subscribes to
	// +optional
	TopicsPattern *string `json:"topicsPattern,omitempty"`

	// CleanupSubscription is the flag to indicate whether the subscription should be cleaned up when the function is deleted
	// +optional
	CleanupSubscription bool `json:"cleanupSubscription"`

	// RetainOrdering is the flag to indicate whether the function should retain ordering
	// +optional
	RetainOrdering bool `json:"retainOrdering"`

	// RetainKeyOrdering is the flag to indicate whether the function should retain key ordering
	// +optional
	RetainKeyOrdering bool `json:"retainKeyOrdering"`

	// BatchBuilder is the batch builder that the function uses
	// +optional
	BatchBuilder *string `json:"batchBuilder,omitempty"`

	// ForwardSourceMessageProperty is the flag to indicate whether the function should forward source message properties
	// +optional
	ForwardSourceMessageProperty bool `json:"forwardSourceMessageProperty"`

	// AutoAck is the flag to indicate whether the function should auto ack
	// +optional
	AutoAck bool `json:"autoAck"`

	// Parallelism is the parallelism of the function
	// +optional
	Parallelism int `json:"parallelism,omitempty"`

	// MaxMessageRetries is the max message retries of the function
	// +optional
	MaxMessageRetries *int `json:"maxMessageRetries,omitempty"`

	// Output is the output of the function
	// +optional
	Output string `json:"output,omitempty"`

	// ProducerConfig is the producer config of the function
	// +optional
	ProducerConfig *ProducerConfig `json:"producerConfig,omitempty"`

	// CustomSchemaOutputs is the custom schema outputs of the function
	// +optional
	CustomSchemaOutputs map[string]string `json:"customSchemaOutputs,omitempty"`

	// OutputSerdeClassName is the output serde class name of the function
	// +optional
	OutputSerdeClassName string `json:"outputSerdeClassName,omitempty"`

	// LogTopic is the log topic of the function
	// +optional
	LogTopic string `json:"logTopic,omitempty"`

	// ProcessingGuarantees is the processing guarantees of the function
	// +optional
	ProcessingGuarantees string `json:"processingGuarantees,omitempty"`

	// OutputSchemaType is the output schema type of the function
	// +optional
	OutputSchemaType string `json:"outputSchemaType,omitempty"`

	// OutputTypeClassName is the output type class name of the function
	// +optional
	OutputTypeClassName string `json:"outputTypeClassName,omitempty"`

	// DeadLetterTopic is the dead letter topic of the function
	// +optional
	DeadLetterTopic string `json:"deadLetterTopic,omitempty"`

	// SubName is the sub name of the function
	// +optional
	SubName string `json:"subName,omitempty"`

	// Jar is the jar of the function
	// +optional
	Jar *PackageContentRef `json:"jar,omitempty"`

	// Py is the py of the function
	// +optional
	Py *PackageContentRef `json:"py,omitempty"`

	// Go is the go of the function
	// +optional
	Go *PackageContentRef `json:"go,omitempty"`

	// RuntimeFlags is the runtime flags of the function
	// +optional
	RuntimeFlags string `json:"runtimeFlags,omitempty"`

	// Tenant is the tenant of the function
	// +optional
	Tenant string `json:"tenant,omitempty"`

	// Namespace is the namespace of the function
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Name is the name of the function
	// +optional
	Name string `json:"name,omitempty"`

	// ClassName is the class name of the function
	// +optional
	ClassName string `json:"className,omitempty"`

	// Resources is the resources of the function
	// +optional
	Resources *Resources `json:"resources,omitempty"`

	// WindowConfig is the window config of the function
	// +optional
	WindowConfig *WindowConfig `json:"windowConfig,omitempty"`

	// Inputs is the inputs of the function
	// +optional
	Inputs []string `json:"inputs,omitempty"`

	// UserConfig is the user config of the function
	// +optional
	UserConfig *apiextensionsv1.JSON `json:"userConfig,omitempty"`

	// CustomSerdeInputs is the custom serde inputs of the function
	// +optional
	CustomSerdeInputs map[string]string `json:"customSerdeInputs,omitempty"`

	// CustomSchemaInputs is the custom schema inputs of the function
	// +optional
	CustomSchemaInputs map[string]string `json:"customSchemaInputs,omitempty"`

	// InputSpecs is the input specs of the function
	// +optional
	InputSpecs map[string]ConsumerConfig `json:"inputSpecs,omitempty"`

	// InputTypeClassName is the input type class name of the function
	// +optional
	InputTypeClassName string `json:"inputTypeClassName,omitempty"`

	// CustomRuntimeOptions is the custom runtime options of the function
	// +optional
	CustomRuntimeOptions *apiextensionsv1.JSON `json:"customRuntimeOptions,omitempty"`

	// Secrets is the secrets of the function
	// +optional
	Secrets map[string]SecretKeyRef `json:"secrets,omitempty"`

	// MaxPendingAsyncRequests is the max pending async requests of the function
	// +optional
	MaxPendingAsyncRequests int `json:"maxPendingAsyncRequests,omitempty"`

	// ExposePulsarAdminClientEnabled is the flag to indicate whether the function should expose pulsar admin client
	// +optional
	ExposePulsarAdminClientEnabled bool `json:"exposePulsarAdminClientEnabled"`

	// SkipToLatest is the flag to indicate whether the function should skip to latest
	// +optional
	SkipToLatest bool `json:"skipToLatest"`

	// SubscriptionPosition is the subscription position of the function
	// +optional
	SubscriptionPosition string `json:"subscriptionPosition,omitempty"`

	// ConnectionRef is the reference to the PulsarConnection resource
	ConnectionRef corev1.LocalObjectReference `json:"connectionRef"`

	// +kubebuilder:validation:Enum=CleanUpAfterDeletion;KeepAfterDeletion
	// +optional
	LifecyclePolicy PulsarResourceLifeCyclePolicy `json:"lifecyclePolicy,omitempty"`
}

// WindowConfig defines the window config of the function
type WindowConfig struct {
	WindowLengthCount             *int    `json:"windowLengthCount" yaml:"windowLengthCount"`
	WindowLengthDurationMs        *int64  `json:"windowLengthDurationMs" yaml:"windowLengthDurationMs"`
	SlidingIntervalCount          *int    `json:"slidingIntervalCount" yaml:"slidingIntervalCount"`
	SlidingIntervalDurationMs     *int64  `json:"slidingIntervalDurationMs" yaml:"slidingIntervalDurationMs"`
	LateDataTopic                 *string `json:"lateDataTopic" yaml:"lateDataTopic"`
	MaxLagMs                      *int64  `json:"maxLagMs" yaml:"maxLagMs"`
	WatermarkEmitIntervalMs       *int64  `json:"watermarkEmitIntervalMs" yaml:"watermarkEmitIntervalMs"`
	TimestampExtractorClassName   *string `json:"timestampExtractorClassName" yaml:"timestampExtractorClassName"`
	ActualWindowFunctionClassName *string `json:"actualWindowFunctionClassName" yaml:"actualWindowFunctionClassName"`
	ProcessingGuarantees          *string `json:"processingGuarantees" yaml:"processingGuarantees"`
}

// PulsarFunctionStatus defines the observed state of PulsarFunction
type PulsarFunctionStatus struct {
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
//+kubebuilder:resource:categories=pulsar;pulsarres,shortName=pfunction
//+kubebuilder:printcolumn:name="RESOURCE_NAME",type=string,JSONPath=`.spec.name`
//+kubebuilder:printcolumn:name="GENERATION",type=string,JSONPath=`.metadata.generation`
//+kubebuilder:printcolumn:name="OBSERVED_GENERATION",type=string,JSONPath=`.status.observedGeneration`
//+kubebuilder:printcolumn:name="READY",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`

// PulsarFunction is the Schema for the pulsar functions API
type PulsarFunction struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PulsarFunctionSpec   `json:"spec,omitempty"`
	Status PulsarFunctionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PulsarFunctionList contains a list of PulsarFunction
type PulsarFunctionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PulsarFunction `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PulsarFunction{}, &PulsarFunctionList{})
}
