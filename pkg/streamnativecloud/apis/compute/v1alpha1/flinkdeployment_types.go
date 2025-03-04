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
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FlinkDeployment is the Schema for the flinkdeployments API
// +k8s:openapi-gen=true
// +resource:path=flinkdeployments,strategy=FlinkDeploymentStrategy
// +kubebuilder:categories=all,compute
type FlinkDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   FlinkDeploymentSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status FlinkDeploymentStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

//+kubebuilder:object:root=true

// FlinkDeploymentList contains a list of FlinkDeployments
type FlinkDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FlinkDeployment `json:"items"`
}

type StreamNativeCloudProtocolIntegration struct {
	Pulsar *bool `json:"pulsar,omitempty" protobuf:"bytes,1,opt,name=pulsar"`
	Kafka  *bool `json:"kafka,omitempty" protobuf:"bytes,2,opt,name=kafka"`
}

// FlinkDeploymentSpec defines the desired state of FlinkDeployment
type FlinkDeploymentSpec struct {
	// WorkspaceName is the reference to the workspace, and is required
	// +kubebuilder:validation:Required
	WorkspaceName string `json:"workspaceName" protobuf:"bytes,1,opt,name=workspaceName"`

	// PoolMemberRef is the pool member to deploy the cluster.
	// admission controller will infer this information automatically
	// if not provided, the controller will use PoolMember from the workspace
	// +optional
	PoolMemberRef *PoolMemberReference `json:"poolMemberRef,omitempty" protobuf:"bytes,2,opt,name=poolMemberRef"`

	// +optional
	Labels map[string]string `json:"labels,omitempty" protobuf:"bytes,3,opt,name=labels"`

	// +optional
	Annotations map[string]string `json:"annotations,omitempty" protobuf:"bytes,4,opt,name=annotations"`

	// ververica-platform: https://docs.ververica.com/vvp/platform-operations/k8s-operator/op-configs
	// opensource: https://nightlies.apache.org/flink/flink-docs-stable/docs/deployment/config/
	// +kubebuilder:validation:Required
	FlinkDeploymentTemplate `json:",inline" protobuf:"bytes,5,req"`

	// DefaultPulsarCluster is the default pulsar cluster to use. If not provided, the controller will use the first pulsar cluster from the workspace.
	// +optional
	DefaultPulsarCluster *string `json:"defaultPulsarCluster,omitempty" protobuf:"bytes,6,opt,name=defaultPulsarCluster"`

	// Configuration is the list of configurations to deploy with the Flink deployment.
	// +optional
	Configuration *Configuration `json:"configuration,omitempty" protobuf:"bytes,7,opt,name=configuration"`
}

type FlinkDeploymentTemplate struct {
	// +optional
	Template *VvpDeploymentTemplate `json:"template,omitempty" protobuf:"bytes,1,opt,name=template"`

	// +optional
	CommunityTemplate *CommunityDeploymentTemplate `json:"communityTemplate,omitempty" protobuf:"bytes,2,opt,name=communityTemplate"`
}

// VvpDeploymentTemplate defines the desired state of VvpDeployment
type VvpDeploymentTemplate struct {
	// +optional
	// +kubebuilder:validation:Enum=PUT;PATCH
	SyncingMode string `json:"syncingMode,omitempty" protobuf:"bytes,1,opt,name=syncingMode"`

	// +kubebuilder:validation:Required
	Deployment VvpDeploymentTemplateSpec `json:"deployment" protobuf:"bytes,2,req,name=deployment"`
}

// VvpDeploymentTemplateSpec defines the desired state of VvpDeployment
type VvpDeploymentTemplateSpec struct {
	// +kubebuilder:validation:Required
	UserMetadata *UserMetadata `json:"userMetadata" protobuf:"bytes,1,req,name=userMetadata"`

	// +kubebuilder:validation:Required
	Spec VvpDeploymentDetails `json:"spec" protobuf:"bytes,2,req,name=spec"`
}

// CommunityDeploymentTemplate defines the desired state of CommunityDeployment
type CommunityDeploymentTemplate struct {
	// TODO: dummy
}

// VvpDeploymentDetails defines the desired state of VvpDeploymentDetails
type VvpDeploymentDetails struct {
	DeploymentTargetName     *string `json:"deploymentTargetName,omitempty" protobuf:"bytes,1,opt,name=deploymentTargetName"`
	JobFailureExpirationTime *string `json:"jobFailureExpirationTime,omitempty" protobuf:"bytes,2,opt,name=jobFailureExpirationTime"`

	// +kubebuilder:validation:Minimum=1
	MaxJobCreationAttempts *int32 `json:"maxJobCreationAttempts,omitempty" protobuf:"varint,3,opt,name=maxJobCreationAttempts"`

	// +kubebuilder:validation:Minimum=1
	MaxSavepointCreationAttempts *int32 `json:"maxSavepointCreationAttempts,omitempty" protobuf:"varint,4,opt,name=maxSavepointCreationAttempts"`

	RestoreStrategy    *VvpRestoreStrategy `json:"restoreStrategy,omitempty" protobuf:"bytes,5,opt,name=restoreStrategy"`
	SessionClusterName *string             `json:"sessionClusterName,omitempty" protobuf:"bytes,6,opt,name=sessionClusterName"`

	// +kubebuilder:validation:Enum=RUNNING;SUSPENDED;CANCELLED
	State string `json:"state,omitempty" protobuf:"bytes,7,opt,name=state"`

	// +kubebuilder:validation:Required
	Template VvpDeploymentDetailsTemplate `json:"template" protobuf:"bytes,8,req,name=template"`
}

// VvpRestoreStrategy defines the restore strategy of the deployment
type VvpRestoreStrategy struct {
	AllowNonRestoredState bool `json:"allowNonRestoredState,omitempty" protobuf:"varint,1,opt,name=allowNonRestoredState"`

	Kind string `json:"kind,omitempty" protobuf:"bytes,2,opt,name=kind"`
}

// VvpDeploymentDetailsTemplate defines the desired state of VvpDeploymentDetails
type VvpDeploymentDetailsTemplate struct {
	Metadata *VvpDeploymentDetailsTemplateMetadata `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// +kubebuilder:validation:Required
	Spec VvpDeploymentDetailsTemplateSpec `json:"spec" protobuf:"bytes,2,req,name=spec"`
}

// VvpDeploymentDetailsTemplateMetadata defines the desired state of VvpDeploymentDetailsTemplateMetadata
type VvpDeploymentDetailsTemplateMetadata struct {
	Annotations map[string]string `json:"annotations,omitempty" protobuf:"bytes,1,rep,name=annotations"`
}

// VvpDeploymentDetailsTemplateSpec defines the desired state of VvpDeploymentDetails
type VvpDeploymentDetailsTemplateSpec struct {
	// +kubebuilder:validation:Required
	Artifact *Artifact `json:"artifact" protobuf:"bytes,1,req,name=artifact"`

	// +optional
	FlinkConfiguration map[string]string `json:"flinkConfiguration,omitempty" protobuf:"bytes,2,rep,name=flinkConfiguration"`

	// +optional
	Kubernetes *VvpDeploymentDetailsTemplateSpecKubernetesSpec `json:"kubernetes,omitempty" protobuf:"bytes,3,opt,name=kubernetes"`

	// +optional
	LatestCheckpointFetchInterval *int32 `json:"latestCheckpointFetchInterval,omitempty" protobuf:"varint,4,opt,name=latestCheckpointFetchInterval"`

	// +optional
	Logging *Logging `json:"logging,omitempty" protobuf:"bytes,5,opt,name=logging"`

	// +optional
	NumberOfTaskManagers *int32 `json:"numberOfTaskManagers,omitempty" protobuf:"varint,6,opt,name=numberOfTaskManagers"`

	// +optional
	Parallelism *int32 `json:"parallelism,omitempty" protobuf:"varint,7,opt,name=parallelism"`

	// +optional
	Resources *VvpDeploymentKubernetesResources `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
}

// VvpDeploymentDetailsTemplateSpecKubernetesSpec defines the desired state of VvpDeploymentDetails
type VvpDeploymentDetailsTemplateSpecKubernetesSpec struct {
	// +optional
	JobManagerPodTemplate *PodTemplate `json:"jobManagerPodTemplate,omitempty" protobuf:"bytes,1,opt,name=jobManagerPodTemplate"`

	// +optional
	TaskManagerPodTemplate *PodTemplate `json:"taskManagerPodTemplate,omitempty" protobuf:"bytes,2,opt,name=taskManagerPodTemplate"`

	// +optional
	Labels map[string]string `json:"labels,omitempty" protobuf:"bytes,3,opt,name=labels"`
}

// UserMetadata Specify the metadata for the resource we are deploying.
type UserMetadata struct {
	// +optional
	Annotations map[string]string `json:"annotations,omitempty" protobuf:"bytes,1,opt,name=annotations"`

	// +optional
	DisplayName string `json:"displayName,omitempty" protobuf:"bytes,2,opt,name=displayName"`

	// +optional
	Labels map[string]string `json:"labels,omitempty" protobuf:"bytes,3,opt,name=labels"`

	// +optional
	Name string `json:"name,omitempty" protobuf:"bytes,4,opt,name=name"`

	// +optional
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,5,opt,name=namespace"`
}

// Artifact is the artifact configs to deploy.
type Artifact struct {
	// +optional
	//nolint:stylecheck
	JarUri string `json:"jarUri,omitempty" protobuf:"bytes,1,opt,name=jarUri"`

	// +kubebuilder:validation:Enum=JAR;PYTHON;sqlscript
	// +kubebuilder:validation:Required
	Kind string `json:"kind,omitempty" protobuf:"bytes,2,req,name=kind"`

	// +optional
	//nolint:stylecheck
	PythonArtifactUri string `json:"pythonArtifactUri,omitempty" protobuf:"bytes,3,opt,name=pythonArtifactUri"`

	// +optional
	//nolint:stylecheck
	SqlScript string `json:"sqlScript,omitempty" protobuf:"bytes,4,opt,name=sqlScript"`

	// +optional
	AdditionalDependencies []string `json:"additionalDependencies,omitempty" protobuf:"bytes,5,opt,name=additionalDependencies"`

	// +optional
	AdditionalPythonArchives []string `json:"additionalPythonArchives,omitempty" protobuf:"bytes,6,opt,name=additionalPythonArchives"`

	// +optional
	AdditionalPythonLibraries []string `json:"additionalPythonLibraries,omitempty" protobuf:"bytes,7,opt,name=additionalPythonLibraries"`

	// +optional
	// +kubebuilder:validation:Enum=PYTHON;SQLSCRIPT;JAR;UNKNOWN
	ArtifactKind string `json:"artifactKind,omitempty" protobuf:"bytes,8,opt,name=artifactKind"`

	// +optional
	EntryClass string `json:"entryClass,omitempty" protobuf:"bytes,9,opt,name=entryClass"`

	// +optional
	EntryModule string `json:"entryModule,omitempty" protobuf:"bytes,10,opt,name=entryModule"`

	// +optional
	FlinkImageRegistry string `json:"flinkImageRegistry,omitempty" protobuf:"bytes,11,opt,name=flinkImageRegistry"`

	// +optional
	FlinkImageRepository string `json:"flinkImageRepository,omitempty" protobuf:"bytes,12,opt,name=flinkImageRepository"`

	// +optional
	FlinkImageTag string `json:"flinkImageTag,omitempty" protobuf:"bytes,13,opt,name=flinkImageTag"`

	// +optional
	FlinkVersion string `json:"flinkVersion,omitempty" protobuf:"bytes,14,opt,name=flinkVersion"`

	// +optional
	MainArgs string `json:"mainArgs,omitempty" protobuf:"bytes,15,opt,name=mainArgs"`

	// +optional
	//nolint:stylecheck
	Uri string `json:"uri,omitempty" protobuf:"bytes,16,opt,name=uri"`

	// +optional
	ArtifactImage string `json:"artifactImage,omitempty" protobuf:"bytes,17,opt,name=artifactImage"`
}

// Logging defines the logging configuration for the Flink deployment.
type Logging struct {
	// +optional
	Log4j2ConfigurationTemplate string `json:"log4j2ConfigurationTemplate,omitempty" protobuf:"bytes,1,opt,name=log4j2ConfigurationTemplate"`

	// +optional
	Log4jLoggers map[string]string `json:"log4jLoggers,omitempty" protobuf:"bytes,2,opt,name=log4jLoggers"`

	// +optional
	LoggingProfile string `json:"loggingProfile,omitempty" protobuf:"bytes,3,opt,name=loggingProfile"`
}

// VvpDeploymentKubernetesResources defines the Kubernetes resources for the VvpDeployment.
type VvpDeploymentKubernetesResources struct {
	// +optional
	Jobmanager *ResourceSpec `json:"jobmanager,omitempty" protobuf:"bytes,1,opt,name=jobmanager"`

	// +optional
	Taskmanager *ResourceSpec `json:"taskmanager,omitempty" protobuf:"bytes,2,opt,name=taskmanager"`
}

// ResourceSpec defines the resource requirements for a component.
type ResourceSpec struct {
	// CPU represents the minimum amount of CPU required.
	// +kubebuilder:validation:Required
	//nolint:stylecheck
	Cpu string `json:"cpu" protobuf:"bytes,1,req,name=cpu"`

	// Memory represents the minimum amount of memory required.
	// +kubebuilder:validation:Required
	Memory string `json:"memory" protobuf:"bytes,2,req,name=memory"`
}

// WorkspaceReference is a reference to a workspace with a given name.
type WorkspaceReference struct {
	// +kubebuilder:validation:Required
	Namespace string `json:"namespace" protobuf:"bytes,1,req,name=namespace"`
	// +kubebuilder:validation:Required
	Name string `json:"name" protobuf:"bytes,2,req,name=name"`
}

// FlinkDeploymentStatus defines the observed state of FlinkDeployment
type FlinkDeploymentStatus struct {
	// Conditions is an array of current observed conditions.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,name=conditions"`

	// DeploymentStatus is the status of the vvp deployment.
	// +optional
	DeploymentStatus *VvpDeploymentStatus `json:"deploymentStatus,omitempty" protobuf:"bytes,2,opt,name=deploymentStatus"`

	// observedGeneration represents the .metadata.generation that the condition was set based upon.
	// For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
	// with respect to the current state of the instance.
	// +optional
	// +kubebuilder:validation:Minimum=0
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,3,opt,name=observedGeneration"`
}

// VvpDeploymentStatus defines the observed state of VvpDeployment
type VvpDeploymentStatus struct {
	CustomResourceStatus     VvpCustomResourceStatus             `json:"customResourceStatus,omitempty" protobuf:"bytes,1,opt,name=customResourceStatus"`
	DeploymentStatus         VvpDeploymentStatusDeploymentStatus `json:"deploymentStatus,omitempty" protobuf:"bytes,2,opt,name=deploymentStatus"`
	DeploymentSystemMetadata VvpDeploymentSystemMetadata         `json:"deploymentSystemMetadata,omitempty" protobuf:"bytes,3,opt,name=deploymentSystemMetadata"`
}

// VvpCustomResourceStatus defines the observed state of VvpCustomResource
type VvpCustomResourceStatus struct {
	// +kubebuilder:validation:Enum=SYNCING;IDLING;DELETING;REF_ENTITY_MANAGED_BY_ANOTHER_CR;SYNCING_MODE_INCONSISTENT;REF_NAMESPACE_UNAVAILABLE;CONCURRENT_RECONCILIATION;ERROR
	CustomResourceState VvpDeploymentCustomResourceState `json:"customResourceState,omitempty" protobuf:"bytes,1,opt,name=customResourceState"`
	//nolint:stylecheck
	DeploymentId      string `json:"deploymentId,omitempty" protobuf:"bytes,2,opt,name=deploymentId"`
	ObservedSpecState string `json:"observedSpecState,omitempty" protobuf:"bytes,3,opt,name=observedSpecState"`
	// +kubebuilder:validation:Enum=RUNNING;SUSPENDED;CANCELLED;TRANSITIONING;FAILED;FINISHED
	StatusState         VvpDeploymentStatusState `json:"statusState,omitempty" protobuf:"bytes,4,opt,name=statusState"`
	DeploymentNamespace string                   `json:"deploymentNamespace,omitempty" protobuf:"bytes,5,opt,name=deploymentNamespace"`
}

type VvpDeploymentStatusState string
type VvpDeploymentCustomResourceState string

const (
	VvpDeploymentStatusStateRunning       VvpDeploymentStatusState = "RUNNING"
	VvpDeploymentStatusStateSuspended     VvpDeploymentStatusState = "SUSPENDED"
	VvpDeploymentStatusStateCancelled     VvpDeploymentStatusState = "CANCELLED"
	VvpDeploymentStatusStateTransitioning VvpDeploymentStatusState = "TRANSITIONING"
	VvpDeploymentStatusStateFailed        VvpDeploymentStatusState = "FAILED"
	VvpDeploymentStatusStateFinished      VvpDeploymentStatusState = "FINISHED"

	VvpDeploymentCustomResourceStateSyncing                     VvpDeploymentCustomResourceState = "SYNCING"
	VvpDeploymentCustomResourceStateIdling                      VvpDeploymentCustomResourceState = "IDLING"
	VvpDeploymentCustomResourceStateDeleting                    VvpDeploymentCustomResourceState = "DELETING"
	VvpDeploymentCustomResourceStateRefEntityManagedByAnotherCr VvpDeploymentCustomResourceState = "REF_ENTITY_MANAGED_BY_ANOTHER_CR"
	VvpDeploymentCustomResourceStateSyncingModeInconsistent     VvpDeploymentCustomResourceState = "SYNCING_MODE_INCONSISTENT"
	VvpDeploymentCustomResourceStateRefNamespaceUnavailable     VvpDeploymentCustomResourceState = "REF_NAMESPACE_UNAVAILABLE"
	VvpDeploymentCustomResourceStateConcurrentReconciliation    VvpDeploymentCustomResourceState = "CONCURRENT_RECONCILIATION"
	VvpDeploymentCustomResourceStateError                       VvpDeploymentCustomResourceState = "ERROR"
)

// VvpDeploymentStatusDeploymentStatus defines the observed state of VvpDeployment
type VvpDeploymentStatusDeploymentStatus struct {
	Running VvpDeploymentRunningStatus `json:"running,omitempty" protobuf:"bytes,1,opt,name=running"`

	// +kubebuilder:validation:Enum=RUNNING;SUSPENDED;CANCELLED
	State string `json:"state,omitempty" protobuf:"bytes,2,opt,name=state"`
}

// VvpDeploymentRunningStatus defines the observed state of VvpDeployment
type VvpDeploymentRunningStatus struct {
	//nolint:stylecheck
	JobId          string      `json:"jobId,omitempty" protobuf:"bytes,1,opt,name=jobId"`
	TransitionTime metav1.Time `json:"transitionTime,omitempty" protobuf:"bytes,2,opt,name=transitionTime"`
	// Conditions is an array of current observed conditions.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []VvpDeploymentStatusCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,3,name=conditions"`
}

type VvpDeploymentStatusCondition struct {
	Type               string                 `json:"type" protobuf:"bytes,1,opt,name=type"`
	Status             metav1.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/apimachinery/pkg/apis/meta/v1.ConditionStatus"`
	Reason             string                 `json:"reason,omitempty" protobuf:"bytes,3,opt,name=reason"`
	Message            string                 `json:"message,omitempty" protobuf:"bytes,4,opt,name=message"`
	LastTransitionTime metav1.Time            `json:"lastTransitionTime,omitempty" protobuf:"bytes,5,opt,name=lastTransitionTime"`
	// observedGeneration represents the .metadata.generation that the condition was set based upon.
	// For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
	// with respect to the current state of the instance.
	// +optional
	// +kubebuilder:validation:Minimum=0
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,6,opt,name=observedGeneration"`
}

// VvpDeploymentSystemMetadata defines the observed state of VvpDeployment
type VvpDeploymentSystemMetadata struct {
	// +optional
	Annotations map[string]string `json:"annotations,omitempty" protobuf:"bytes,1,opt,name=annotations"`

	// +optional
	Labels map[string]string `json:"labels,omitempty" protobuf:"bytes,2,opt,name=labels"`

	// +optional
	//nolint:stylecheck
	Id string `json:"name,omitempty" protobuf:"bytes,3,opt,name=name"`

	// +optional
	ResourceVersion int32 `json:"resourceVersion,omitempty" protobuf:"bytes,4,opt,name=resourceVersion"`

	// +optional
	CreatedAt metav1.Time `json:"createdAt,omitempty" protobuf:"bytes,5,opt,name=createdAt"`

	// +optional
	ModifiedAt metav1.Time `json:"modifiedAt,omitempty" protobuf:"bytes,6,opt,name=modifiedAt"`
}

// PoolMemberReference is a reference to a pool member with a given name.
type PoolMemberReference struct {
	Namespace string `json:"namespace" protobuf:"bytes,1,opt,name=namespace"`
	Name      string `json:"name" protobuf:"bytes,2,opt,name=name"`
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
