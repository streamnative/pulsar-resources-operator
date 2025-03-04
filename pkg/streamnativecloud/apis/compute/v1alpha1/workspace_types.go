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

// Copyright (c) 2020 StreamNative, Inc.. All Rights Reserved.

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Workspace
// +k8s:openapi-gen=true
// +resource:path=workspaces,strategy=WorkspaceStrategy
// +kubebuilder:categories=all,compute
type Workspace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   WorkspaceSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status WorkspaceStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

//+kubebuilder:object:root=true

// WorkspaceList contains a list of Workspaces
type WorkspaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workspace `json:"items"`
}

// WorkspaceSpec defines the desired state of Workspace
type WorkspaceSpec struct {
	// PulsarClusterNames is the list of Pulsar clusters that the workspace will have access to.
	PulsarClusterNames []string `json:"pulsarClusterNames" protobuf:"bytes,1,rep,name=pulsarClusterNames"`

	// PoolRef is the reference to the pool that the workspace will be access to.
	PoolRef *PoolRef `json:"poolRef" protobuf:"bytes,2,opt,name=poolRef"`

	// UseExternalAccess is the flag to indicate whether the workspace will use external access.
	// +optional
	UseExternalAccess *bool `json:"useExternalAccess,omitempty" protobuf:"bytes,3,opt,name=useExternalAccess"`

	// FlinkBlobStorage is the configuration for the Flink blob storage.
	// +optional
	FlinkBlobStorage *FlinkBlobStorage `json:"flinkBlobStorage" protobuf:"bytes,4,opt,name=flinkBlobStorage"`

	// FlinkBlobStorageCredentials is the credentials for the Flink blob storage.
	// +optional
	//FlinkBlobStorageCredentials *FlinkBlobStorageCredentials `json:"flinkBlobStorageCredentials,omitempty" protobuf:"bytes,5,opt,name=flinkBlobStorageCredentials"`
}

// FlinkBlobStorage defines the configuration for the Flink blob storage.
type FlinkBlobStorage struct {
	// Bucket is required if you want to use cloud storage.
	Bucket string `json:"bucket,omitempty" protobuf:"bytes,1,opt,name=bucket"`

	// Path is the sub path in the bucket. Leave it empty if you want to use the whole bucket.
	// +optional
	Path string `json:"path,omitempty" protobuf:"bytes,2,opt,name=path"`

	// S3 is the configuration for the S3 blob storage.
	// +optional
	//S3 *FlinkBlobStorageS3Config `json:"s3" protobuf:"bytes,3,opt,name=s3"`

	// OSS is the configuration for the OSS blob storage.
	// +optional
	//OSS *FlinkBlobStorageOSSConfig `json:"oss" protobuf:"bytes,4,opt,name=oss"`
}

// FlinkBlobStorageS3Config defines the configuration for the S3 blob storage.
type FlinkBlobStorageS3Config struct {
	// Endpoint is the endpoint for the S3 blob storage.
	// +optional
	Endpoint string `json:"endpoint,omitempty" protobuf:"bytes,1,opt,name=endpoint"`

	// Region is the region for the S3 blob storage.
	// +optional
	Region string `json:"region,omitempty" protobuf:"bytes,2,opt,name=region"`
}

// FlinkBlobStorageOSSConfig defines the configuration for the OSS blob storage.
type FlinkBlobStorageOSSConfig struct {
	// Endpoint is the endpoint for the OSS blob storage.
	// +optional
	Endpoint string `json:"endpoint,omitempty" protobuf:"bytes,1,opt,name=endpoint"`
}

// FlinkBlobStorageCredentials defines the credentials for the Flink blob storage.
type FlinkBlobStorageCredentials struct {
	// ExistingSecret is the reference to the existing secret that contains the credentials for the Flink blob storage.
	// Use an existing Kubernetes secret instead of providing credentials in this file. The keys
	// within the secret must follow the format: `<provider>.<credential>`
	// For example: `s3.accessKeyId` or `azure.connectionString`
	// +optional
	ExistingSecret *corev1.SecretReference `json:"existingSecret,omitempty" protobuf:"bytes,1,opt,name=existingSecret"`
}

// WorkspaceStatus defines the observed state of Workspace
type WorkspaceStatus struct {
	// Conditions is an array of current observed pool conditions.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}
