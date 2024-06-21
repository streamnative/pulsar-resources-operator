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

// PulsarSourceSpec defines the desired state of PulsarSource
type PulsarSourceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Tenant is the tenant of the PulsarSource
	// +optional
	Tenant string `json:"tenant,omitempty"`

	// Namespace is the namespace of the PulsarSource
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Name is the name of the PulsarSource
	// +optional
	Name string `json:"name,omitempty"`

	// ClassName is the class name of the
	// +optional
	ClassName string `json:"className,omitempty"`

	// ProducerConfig is the producer config of the PulsarSource
	// +optional
	ProducerConfig *ProducerConfig `json:"producerConfig,omitempty"`

	// TopicName is the topic name of the PulsarSource
	// +optional
	TopicName string `json:"topicName,omitempty"`

	// SerdeClassName is the serde class name of the PulsarSource
	// +optional
	SerdeClassName string `json:"serdeClassName,omitempty"`

	// SchemaType is the schema type of the PulsarSource
	// +optional
	SchemaType string `json:"schemaType,omitempty"`

	// Configs is the map of configs of the PulsarSource
	// +optional
	Configs *apiextensionsv1.JSON `json:"configs,omitempty"`

	// Secrets is the map of secrets of the PulsarSource
	// +optional
	Secrets map[string]SecretKeyRef `json:"secrets,omitempty"`

	// Parallelism is the parallelism of the PulsarSource
	// +optional
	Parallelism int `json:"parallelism,omitempty"`

	// ProcessingGuarantees is the processing guarantees of the PulsarSource
	// +optional
	ProcessingGuarantees string `json:"processingGuarantees,omitempty"`

	// Resources is the resources of the PulsarSource
	// +optional
	Resources *Resources `json:"resources,omitempty"`

	// Archive is the archive of the PulsarSource
	// +optional
	Archive *PackageContentRef `json:"archive,omitempty"`

	// RuntimeFlags is the runtime flags of the PulsarSource
	// +optional
	RuntimeFlags string `json:"runtimeFlags,omitempty"`

	// CustomRuntimeOptions is the custom runtime options of the PulsarSource
	// +optional
	CustomRuntimeOptions *apiextensionsv1.JSON `json:"customRuntimeOptions,omitempty"`

	// BatchSourceConfig is the batch source config of the PulsarSource
	// +optional
	BatchSourceConfig *BatchSourceConfig `json:"batchSourceConfig,omitempty"`

	// BatchBuilder is the batch builder of the PulsarSource
	// +optional
	BatchBuilder string `json:"batchBuilder,omitempty"`

	// ConnectionRef is the reference to the PulsarConnection resource
	ConnectionRef corev1.LocalObjectReference `json:"connectionRef"`

	// +kubebuilder:validation:Enum=CleanUpAfterDeletion;KeepAfterDeletion
	// +optional
	LifecyclePolicy PulsarResourceLifeCyclePolicy `json:"lifecyclePolicy,omitempty"`
}

// BatchSourceConfig represents the batch source config of the PulsarSource
type BatchSourceConfig struct {
	DiscoveryTriggererClassName string `json:"discoveryTriggererClassName" yaml:"discoveryTriggererClassName"`

	DiscoveryTriggererConfig *apiextensionsv1.JSON `json:"discoveryTriggererConfig" yaml:"discoveryTriggererConfig"`
}

// PulsarSourceStatus defines the observed state of PulsarSource
type PulsarSourceStatus struct {
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
//+kubebuilder:resource:categories=pulsar;pulsarres,shortName=psource
//+kubebuilder:printcolumn:name="RESOURCE_NAME",type=string,JSONPath=`.spec.name`
//+kubebuilder:printcolumn:name="GENERATION",type=string,JSONPath=`.metadata.generation`
//+kubebuilder:printcolumn:name="OBSERVED_GENERATION",type=string,JSONPath=`.status.observedGeneration`
//+kubebuilder:printcolumn:name="READY",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`

// PulsarSource is the Schema for the pulsar functions API
type PulsarSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PulsarSourceSpec   `json:"spec,omitempty"`
	Status PulsarSourceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PulsarSourceList contains a list of PulsarSource
type PulsarSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PulsarSource `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PulsarSource{}, &PulsarSourceList{})
}
