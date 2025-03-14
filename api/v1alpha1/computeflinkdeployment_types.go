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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ComputeFlinkDeploymentSpec defines the desired state of ComputeFlinkDeployment
type ComputeFlinkDeploymentSpec struct {
	// APIServerRef is the reference to the StreamNativeCloudConnection.
	// If not specified, the APIServerRef from the referenced ComputeWorkspace will be used.
	// +optional
	APIServerRef StreamNativeCloudConnectionRef `json:"apiServerRef,omitempty"`

	// WorkspaceName is the reference to the workspace, and is required
	// +kubebuilder:validation:Required
	// +required
	WorkspaceName string `json:"workspaceName"`

	// Labels to add to the deployment
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations to add to the deployment
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// Template is the VVP deployment template
	// +optional
	Template *VvpDeploymentTemplate `json:"template,omitempty"`

	// CommunityTemplate is the community deployment template
	// +optional
	CommunityTemplate *CommunityDeploymentTemplate `json:"communityTemplate,omitempty"`

	// DefaultPulsarCluster is the default pulsar cluster to use
	// +optional
	DefaultPulsarCluster *string `json:"defaultPulsarCluster,omitempty"`

	// Configuration is the list of configurations to deploy with the Flink deployment.
	// +optional
	Configuration *Configuration `json:"configuration,omitempty"`

	// ImagePullSecrets is the list of image pull secrets to use for the deployment.
	// +optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
}

// ComputeFlinkDeploymentStatus defines the observed state of ComputeFlinkDeployment
type ComputeFlinkDeploymentStatus struct {
	// Conditions represent the latest available observations of an object's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the last observed generation.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// DeploymentStatus represents the status from the API server
	// +optional
	DeploymentStatus *runtime.RawExtension `json:"deploymentStatus,omitempty"`
}

// VvpDeploymentTemplate defines the VVP deployment template
type VvpDeploymentTemplate struct {
	// SyncingMode defines how the deployment should be synced
	// +optional
	SyncingMode string `json:"syncingMode,omitempty"`

	// Deployment defines the deployment configuration
	// +required
	Deployment VvpDeploymentTemplateSpec `json:"deployment"`
}

// VvpDeploymentTemplateSpec defines the deployment configuration
type VvpDeploymentTemplateSpec struct {
	// UserMetadata defines the metadata for the deployment
	// +required
	UserMetadata UserMetadata `json:"userMetadata"`

	// Spec defines the deployment specification
	// +required
	Spec VvpDeploymentDetails `json:"spec"`
}

// UserMetadata defines the metadata for the deployment
type UserMetadata struct {
	// Name of the deployment
	// +optional
	Name string `json:"name,omitempty"`

	// Namespace of the deployment
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Labels of the deployment
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations of the deployment
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// DisplayName of the deployment
	// +optional
	DisplayName string `json:"displayName,omitempty"`
}

// VvpDeploymentDetails defines the deployment details
type VvpDeploymentDetails struct {
	// DeploymentTargetName defines the target name for the deployment
	// +optional
	DeploymentTargetName *string `json:"deploymentTargetName,omitempty"`

	// JobFailureExpirationTime defines the expiration time for job failures
	// +optional
	JobFailureExpirationTime *string `json:"jobFailureExpirationTime,omitempty"`

	// MaxJobCreationAttempts defines the maximum number of job creation attempts
	// +kubebuilder:validation:Minimum=1
	// +optional
	MaxJobCreationAttempts *int32 `json:"maxJobCreationAttempts,omitempty"`

	// MaxSavepointCreationAttempts defines the maximum number of savepoint creation attempts
	// +kubebuilder:validation:Minimum=1
	// +optional
	MaxSavepointCreationAttempts *int32 `json:"maxSavepointCreationAttempts,omitempty"`

	// RestoreStrategy defines the restore strategy for the deployment
	// +optional
	RestoreStrategy *VvpRestoreStrategy `json:"restoreStrategy,omitempty"`

	// SessionClusterName defines the name of the session cluster
	// +optional
	SessionClusterName *string `json:"sessionClusterName,omitempty"`

	// State of the deployment
	// +kubebuilder:validation:Enum=RUNNING;SUSPENDED;CANCELLED
	// +optional
	State string `json:"state,omitempty"`

	// Template defines the deployment template
	// +required
	Template VvpDeploymentDetailsTemplate `json:"template"`
}

// VvpDeploymentDetailsTemplate defines the deployment template
type VvpDeploymentDetailsTemplate struct {
	// Metadata of the deployment
	// +optional
	Metadata VvpDeploymentDetailsTemplateMetadata `json:"metadata,omitempty"`

	// Spec defines the deployment specification
	// +required
	Spec VvpDeploymentDetailsTemplateSpec `json:"spec"`
}

// VvpDeploymentDetailsTemplateMetadata defines the metadata for the deployment template
type VvpDeploymentDetailsTemplateMetadata struct {
	// Annotations to add to the deployment
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

// VvpDeploymentDetailsTemplateSpec defines the deployment template specification
type VvpDeploymentDetailsTemplateSpec struct {
	// Artifact defines the deployment artifact
	// +required
	Artifact *Artifact `json:"artifact"`

	// FlinkConfiguration defines the Flink configuration
	// +optional
	FlinkConfiguration map[string]string `json:"flinkConfiguration,omitempty"`

	// +optional
	Kubernetes *VvpDeploymentDetailsTemplateSpecKubernetesSpec `json:"kubernetes,omitempty"`

	// +optional
	LatestCheckpointFetchInterval *int32 `json:"latestCheckpointFetchInterval,omitempty"`

	// +optional
	Logging *Logging `json:"logging,omitempty"`

	// +optional
	NumberOfTaskManagers *int32 `json:"numberOfTaskManagers,omitempty"`

	// +optional
	Parallelism *int32 `json:"parallelism,omitempty"`

	// +optional
	Resources *VvpDeploymentKubernetesResources `json:"resources,omitempty"`
}

// KubernetesSpec defines the Kubernetes specific configuration
type KubernetesSpec struct {
	// JobManagerPodTemplate defines the job manager pod template
	// +optional
	JobManagerPodTemplate *PodTemplate `json:"jobManagerPodTemplate,omitempty"`

	// TaskManagerPodTemplate defines the task manager pod template
	// +optional
	TaskManagerPodTemplate *PodTemplate `json:"taskManagerPodTemplate,omitempty"`

	// Labels defines the labels to add to all resources
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
}

// Artifact is the artifact configs to deploy.
type Artifact struct {
	// +optional
	JarURI string `json:"jarUri,omitempty"`

	// +kubebuilder:validation:Enum=JAR;PYTHON;sqlscript
	// +kubebuilder:validation:Required
	Kind string `json:"kind,omitempty"`

	// +optional
	PythonArtifactURI string `json:"pythonArtifactUri,omitempty"`

	// +optional
	SQLScript string `json:"sqlScript,omitempty"`

	// +optional
	AdditionalDependencies []string `json:"additionalDependencies,omitempty"`

	// +optional
	AdditionalPythonArchives []string `json:"additionalPythonArchives,omitempty"`

	// +optional
	AdditionalPythonLibraries []string `json:"additionalPythonLibraries,omitempty"`

	// +optional
	// +kubebuilder:validation:Enum=PYTHON;SQLSCRIPT;JAR;UNKNOWN
	ArtifactKind string `json:"artifactKind,omitempty"`

	// +optional
	EntryClass string `json:"entryClass,omitempty"`

	// +optional
	EntryModule string `json:"entryModule,omitempty"`

	// +optional
	FlinkImageRegistry string `json:"flinkImageRegistry,omitempty"`

	// +optional
	FlinkImageRepository string `json:"flinkImageRepository,omitempty"`

	// +optional
	FlinkImageTag string `json:"flinkImageTag,omitempty"`

	// +optional
	FlinkVersion string `json:"flinkVersion,omitempty"`

	// +optional
	MainArgs string `json:"mainArgs,omitempty"`

	// +optional
	URI string `json:"uri,omitempty"`

	// +optional
	ArtifactImage string `json:"artifactImage,omitempty"`
}

// VvpDeploymentStatus defines the deployment status
type VvpDeploymentStatus struct {
	// CustomResourceState defines the state of the custom resource
	// +optional
	CustomResourceState string `json:"customResourceState,omitempty"`

	// DeploymentID defines the ID of the deployment
	// +optional
	DeploymentID string `json:"deploymentId,omitempty"`

	// StatusState defines the state of the status
	// +optional
	StatusState string `json:"statusState,omitempty"`

	// DeploymentNamespace defines the namespace of the deployment
	// +optional
	DeploymentNamespace string `json:"deploymentNamespace,omitempty"`
}

// CommunityDeploymentTemplate defines the community deployment template
type CommunityDeploymentTemplate struct {
	// Metadata defines the metadata for the deployment
	// +optional
	Metadata CommunityDeploymentMetadata `json:"metadata,omitempty"`

	// Spec defines the deployment specification
	// +required
	Spec CommunityDeploymentSpec `json:"spec"`
}

// CommunityDeploymentMetadata defines the metadata for the community deployment
type CommunityDeploymentMetadata struct {
	// Annotations to add to the deployment
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// Labels to add to the deployment
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
}

// CommunityDeploymentSpec defines the community deployment specification
type CommunityDeploymentSpec struct {
	// Image defines the Flink image
	// +required
	Image string `json:"image"`

	// FlinkConfiguration defines the Flink configuration
	// +optional
	FlinkConfiguration map[string]string `json:"flinkConfiguration,omitempty"`

	// JobManagerPodTemplate defines the job manager pod template
	// +optional
	JobManagerPodTemplate *PodTemplate `json:"jobManagerPodTemplate,omitempty"`

	// TaskManagerPodTemplate defines the task manager pod template
	// +optional
	TaskManagerPodTemplate *PodTemplate `json:"taskManagerPodTemplate,omitempty"`

	// JarURI defines the URI of the JAR file
	// +required
	JarURI string `json:"jarUri"`

	// EntryClass defines the entry class of the JAR
	// +optional
	EntryClass string `json:"entryClass,omitempty"`

	// MainArgs defines the main arguments
	// +optional
	MainArgs string `json:"mainArgs,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Namespaced,categories={streamnative,all}
//+kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status"

// ComputeFlinkDeployment is the Schema for the flinkdeployments API
type ComputeFlinkDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ComputeFlinkDeploymentSpec   `json:"spec,omitempty"`
	Status ComputeFlinkDeploymentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ComputeFlinkDeploymentList contains a list of ComputeFlinkDeployment
type ComputeFlinkDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ComputeFlinkDeployment `json:"items"`
}

// VvpRestoreStrategy defines the restore strategy of the deployment
type VvpRestoreStrategy struct {
	AllowNonRestoredState bool   `json:"allowNonRestoredState,omitempty"`
	Kind                  string `json:"kind,omitempty"`
}

// Logging defines the logging configuration for the Flink deployment.
type Logging struct {
	// +optional
	Log4j2ConfigurationTemplate string `json:"log4j2ConfigurationTemplate,omitempty"`

	// +optional
	Log4jLoggers map[string]string `json:"log4jLoggers,omitempty"`

	// +optional
	LoggingProfile string `json:"loggingProfile,omitempty"`
}

// VvpDeploymentKubernetesResources defines the Kubernetes resources for the VvpDeployment.
type VvpDeploymentKubernetesResources struct {
	// +optional
	Jobmanager *ResourceSpec `json:"jobmanager,omitempty"`

	// +optional
	Taskmanager *ResourceSpec `json:"taskmanager,omitempty"`
}

// ResourceSpec defines the resource requirements for a component.
type ResourceSpec struct {
	// CPU represents the minimum amount of CPU required.
	// +kubebuilder:validation:Required
	CPU string `json:"cpu"`

	// Memory represents the minimum amount of memory required.
	// +kubebuilder:validation:Required
	Memory string `json:"memory"`
}

// VvpDeploymentDetailsTemplateSpecKubernetesSpec defines the Kubernetes spec for the deployment
type VvpDeploymentDetailsTemplateSpecKubernetesSpec struct {
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
}

// SecretReference references a StreamNative Cloud secret.
type SecretReference struct {
	// Name of the ENV variable.
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`

	// ValueFrom references a secret in the same namespace.
	ValueFrom *corev1.SecretKeySelector `json:"valueFrom,omitempty" protobuf:"bytes,2,opt,name=valueFrom"`
}

// EnvVar defines an environment variable.
type EnvVar struct {
	// Name of the environment variable.
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`

	// Value of the environment variable.
	Value string `json:"value" protobuf:"bytes,2,opt,name=value"`
}

// Configuration defines the additional configuration for the Flink deployment
type Configuration struct {
	// Envs is the list of environment variables to set in the Flink deployment.
	// +optional
	Envs []EnvVar `json:"envs,omitempty" protobuf:"bytes,1,opt,name=envs"`

	// Secrets is the list of secrets referenced to deploy with the Flink deployment.
	// +optional
	Secrets []SecretReference `json:"secrets,omitempty" protobuf:"bytes,2,opt,name=secrets"`
}

func init() {
	SchemeBuilder.Register(&ComputeFlinkDeployment{}, &ComputeFlinkDeploymentList{})
}
