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
)

type ObjectMeta struct {
	// Name of the resource within a namespace. It must be unique.
	// +optional
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`

	// Namespace of the resource.
	// +optional
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,2,opt,name=namespace"`

	// Labels of the resource.
	// +optional
	// +nullable
	Labels map[string]string `json:"labels,omitempty" protobuf:"bytes,3,rep,name=labels"`

	// Annotations of the resource.
	// +optional
	// +nullable
	Annotations map[string]string `json:"annotations,omitempty" protobuf:"bytes,4,rep,name=annotations"`
}

// PodTemplate defines the common pod configuration for Pods, including when used in StatefulSets.
type PodTemplate struct {
	// Metadata of the pod.
	// +optional
	Metadata ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Spec of the pod.
	// +optional
	Spec PodTemplateSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

type PodTemplateSpec struct {
	// +optional
	Volumes []Volume `json:"volumes,omitempty" protobuf:"bytes,1,rep,name=volumes"`

	// +optional
	InitContainers []Container `json:"initContainers,omitempty" protobuf:"bytes,2,rep,name=initContainers"`

	Containers []Container `json:"containers,omitempty" protobuf:"bytes,3,rep,name=containers"`

	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty" protobuf:"bytes,4,opt,name=serviceAccountName"`

	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty" protobuf:"bytes,5,rep,name=nodeSelector"`

	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty" protobuf:"bytes,6,opt,name=affinity"`

	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty" protobuf:"bytes,7,rep,name=tolerations"`

	// +optional
	SecurityContext *SecurityContext `json:"securityContext,omitempty" protobuf:"bytes,8,opt,name=securityContext"`

	// +optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty" protobuf:"bytes,9,rep,name=imagePullSecrets"`

	// +optional
	ShareProcessNamespace *bool `json:"shareProcessNamespace,omitempty" protobuf:"bytes,10,opt,name=shareProcessNamespace"`
}
