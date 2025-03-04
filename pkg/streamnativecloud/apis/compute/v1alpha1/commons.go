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

// Volume represents a named volume in a pod that may be accessed by any container in the pod.
// The Volume API from the core group is not used directly to avoid unneeded fields defined in `VolumeSource`
// and reduce the size of the CRD. New fields in VolumeSource could be added as needed.
type Volume struct {
	// Volume's name.
	// Must be a DNS_LABEL and unique within the pod.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`

	// VolumeSource represents the location and type of the mounted volume.
	// If not specified, the Volume is implied to be an EmptyDir.
	// This implied behavior is deprecated and will be removed in a future version.
	VolumeSource `json:",inline" protobuf:"bytes,2,opt,name=volumeSource"`
}

// Represents the source of a volume to mount.
// Only one of its members may be specified.
type VolumeSource struct {
	// ConfigMap represents a configMap that should populate this volume
	// +optional
	ConfigMap *corev1.ConfigMapVolumeSource `json:"configMap,omitempty" protobuf:"bytes,1,opt,name=configMap"`

	// Secret represents a secret that should populate this volume.
	// +optional
	Secret *corev1.SecretVolumeSource `json:"secret,omitempty" protobuf:"bytes,2,opt,name=secret"`
}

// A single application container that you want to run within a pod.
// The Container API from the core group is not used directly to avoid unneeded fields
// and reduce the size of the CRD. New fields could be added as needed.
type Container struct {
	// Name of the container specified as a DNS_LABEL.
	// Each container in a pod must have a unique name (DNS_LABEL).
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`

	// Docker image name.
	// +optional
	Image string `json:"image,omitempty" protobuf:"bytes,2,opt,name=image"`

	// Entrypoint array. Not executed within a shell.
	// +optional
	Command []string `json:"command,omitempty" protobuf:"bytes,3,rep,name=command"`

	// Arguments to the entrypoint.
	// +optional
	Args []string `json:"args,omitempty" protobuf:"bytes,4,rep,name=args"`

	// Container's working directory.
	// +optional
	WorkingDir string `json:"workingDir,omitempty" protobuf:"bytes,5,opt,name=workingDir"`

	// List of sources to populate environment variables in the container.
	// +optional
	EnvFrom []corev1.EnvFromSource `json:"envFrom,omitempty" protobuf:"bytes,6,rep,name=envFrom"`

	// List of environment variables to set in the container.
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	Env []corev1.EnvVar `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,7,rep,name=env"`

	// Compute Resources required by this container.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`

	// Pod volumes to mount into the container's filesystem.
	// +optional
	// +patchMergeKey=mountPath
	// +patchStrategy=merge
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty" patchStrategy:"merge" patchMergeKey:"mountPath" protobuf:"bytes,9,rep,name=volumeMounts"`

	// Periodic probe of container liveness.
	// +optional
	LivenessProbe *corev1.Probe `json:"livenessProbe,omitempty" protobuf:"bytes,10,opt,name=livenessProbe"`

	// Periodic probe of container service readiness.
	// More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	// +optional
	ReadinessProbe *corev1.Probe `json:"readinessProbe,omitempty" protobuf:"bytes,11,opt,name=readinessProbe"`

	// StartupProbe indicates that the Pod has successfully initialized.
	// +optional
	StartupProbe *corev1.Probe `json:"startupProbe,omitempty" protobuf:"bytes,12,opt,name=startupProbe"`

	// Image pull policy.
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty" protobuf:"bytes,13,opt,name=imagePullPolicy,casttype=k8s.io/api/core/v1.PullPolicy"`

	// SecurityContext holds pod-level security attributes and common container settings.
	// +optional
	SecurityContext *SecurityContext `json:"securityContext,omitempty" protobuf:"bytes,14,opt,name=securityContext"`
}

type SecurityContext struct {
	// +optional
	RunAsUser *int64 `json:"runAsUser,omitempty" protobuf:"varint,1,opt,name=runAsUser"`
	// +optional
	RunAsGroup *int64 `json:"runAsGroup,omitempty" protobuf:"varint,2,opt,name=runAsGroup"`
	// +optional
	RunAsNonRoot *bool `json:"runAsNonRoot,omitempty" protobuf:"varint,3,opt,name=runAsNonRoot"`
	// +optional
	FSGroup *int64 `json:"fsGroup,omitempty" protobuf:"varint,4,opt,name=fsGroup"`

	// ReadOnlyRootFilesystem specifies whether the container use a read-only filesystem.
	// +optional
	ReadOnlyRootFilesystem *bool `json:"readOnlyRootFilesystem,omitempty" protobuf:"varint,5,opt,name=readOnlyRootFilesystem"`
}

type PoolRef struct {
	Namespace string `json:"namespace" protobuf:"bytes,1,opt,name=namespace"`
	Name      string `json:"name" protobuf:"bytes,2,opt,name=name"`
}
