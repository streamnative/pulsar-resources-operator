//go:build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Artifact) DeepCopyInto(out *Artifact) {
	*out = *in
	if in.AdditionalDependencies != nil {
		in, out := &in.AdditionalDependencies, &out.AdditionalDependencies
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.AdditionalPythonArchives != nil {
		in, out := &in.AdditionalPythonArchives, &out.AdditionalPythonArchives
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.AdditionalPythonLibraries != nil {
		in, out := &in.AdditionalPythonLibraries, &out.AdditionalPythonLibraries
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Artifact.
func (in *Artifact) DeepCopy() *Artifact {
	if in == nil {
		return nil
	}
	out := new(Artifact)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CommunityDeploymentTemplate) DeepCopyInto(out *CommunityDeploymentTemplate) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CommunityDeploymentTemplate.
func (in *CommunityDeploymentTemplate) DeepCopy() *CommunityDeploymentTemplate {
	if in == nil {
		return nil
	}
	out := new(CommunityDeploymentTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Condition) DeepCopyInto(out *Condition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Condition.
func (in *Condition) DeepCopy() *Condition {
	if in == nil {
		return nil
	}
	out := new(Condition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Configuration) DeepCopyInto(out *Configuration) {
	*out = *in
	if in.Envs != nil {
		in, out := &in.Envs, &out.Envs
		*out = make([]EnvVar, len(*in))
		copy(*out, *in)
	}
	if in.Secrets != nil {
		in, out := &in.Secrets, &out.Secrets
		*out = make([]SecretReference, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Configuration.
func (in *Configuration) DeepCopy() *Configuration {
	if in == nil {
		return nil
	}
	out := new(Configuration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Container) DeepCopyInto(out *Container) {
	*out = *in
	if in.Command != nil {
		in, out := &in.Command, &out.Command
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Args != nil {
		in, out := &in.Args, &out.Args
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.EnvFrom != nil {
		in, out := &in.EnvFrom, &out.EnvFrom
		*out = make([]v1.EnvFromSource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make([]v1.EnvVar, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.Resources.DeepCopyInto(&out.Resources)
	if in.VolumeMounts != nil {
		in, out := &in.VolumeMounts, &out.VolumeMounts
		*out = make([]v1.VolumeMount, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.LivenessProbe != nil {
		in, out := &in.LivenessProbe, &out.LivenessProbe
		*out = new(v1.Probe)
		(*in).DeepCopyInto(*out)
	}
	if in.ReadinessProbe != nil {
		in, out := &in.ReadinessProbe, &out.ReadinessProbe
		*out = new(v1.Probe)
		(*in).DeepCopyInto(*out)
	}
	if in.StartupProbe != nil {
		in, out := &in.StartupProbe, &out.StartupProbe
		*out = new(v1.Probe)
		(*in).DeepCopyInto(*out)
	}
	if in.SecurityContext != nil {
		in, out := &in.SecurityContext, &out.SecurityContext
		*out = new(SecurityContext)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Container.
func (in *Container) DeepCopy() *Container {
	if in == nil {
		return nil
	}
	out := new(Container)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EnvVar) DeepCopyInto(out *EnvVar) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EnvVar.
func (in *EnvVar) DeepCopy() *EnvVar {
	if in == nil {
		return nil
	}
	out := new(EnvVar)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlinkBlobStorage) DeepCopyInto(out *FlinkBlobStorage) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlinkBlobStorage.
func (in *FlinkBlobStorage) DeepCopy() *FlinkBlobStorage {
	if in == nil {
		return nil
	}
	out := new(FlinkBlobStorage)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlinkBlobStorageCredentials) DeepCopyInto(out *FlinkBlobStorageCredentials) {
	*out = *in
	if in.ExistingSecret != nil {
		in, out := &in.ExistingSecret, &out.ExistingSecret
		*out = new(v1.SecretReference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlinkBlobStorageCredentials.
func (in *FlinkBlobStorageCredentials) DeepCopy() *FlinkBlobStorageCredentials {
	if in == nil {
		return nil
	}
	out := new(FlinkBlobStorageCredentials)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlinkBlobStorageOSSConfig) DeepCopyInto(out *FlinkBlobStorageOSSConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlinkBlobStorageOSSConfig.
func (in *FlinkBlobStorageOSSConfig) DeepCopy() *FlinkBlobStorageOSSConfig {
	if in == nil {
		return nil
	}
	out := new(FlinkBlobStorageOSSConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlinkBlobStorageS3Config) DeepCopyInto(out *FlinkBlobStorageS3Config) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlinkBlobStorageS3Config.
func (in *FlinkBlobStorageS3Config) DeepCopy() *FlinkBlobStorageS3Config {
	if in == nil {
		return nil
	}
	out := new(FlinkBlobStorageS3Config)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlinkDeployment) DeepCopyInto(out *FlinkDeployment) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlinkDeployment.
func (in *FlinkDeployment) DeepCopy() *FlinkDeployment {
	if in == nil {
		return nil
	}
	out := new(FlinkDeployment)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *FlinkDeployment) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlinkDeploymentList) DeepCopyInto(out *FlinkDeploymentList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]FlinkDeployment, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlinkDeploymentList.
func (in *FlinkDeploymentList) DeepCopy() *FlinkDeploymentList {
	if in == nil {
		return nil
	}
	out := new(FlinkDeploymentList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *FlinkDeploymentList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlinkDeploymentSpec) DeepCopyInto(out *FlinkDeploymentSpec) {
	*out = *in
	if in.PoolMemberRef != nil {
		in, out := &in.PoolMemberRef, &out.PoolMemberRef
		*out = new(PoolMemberReference)
		**out = **in
	}
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	in.FlinkDeploymentTemplate.DeepCopyInto(&out.FlinkDeploymentTemplate)
	if in.DefaultPulsarCluster != nil {
		in, out := &in.DefaultPulsarCluster, &out.DefaultPulsarCluster
		*out = new(string)
		**out = **in
	}
	if in.Configuration != nil {
		in, out := &in.Configuration, &out.Configuration
		*out = new(Configuration)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlinkDeploymentSpec.
func (in *FlinkDeploymentSpec) DeepCopy() *FlinkDeploymentSpec {
	if in == nil {
		return nil
	}
	out := new(FlinkDeploymentSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlinkDeploymentStatus) DeepCopyInto(out *FlinkDeploymentStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.DeploymentStatus != nil {
		in, out := &in.DeploymentStatus, &out.DeploymentStatus
		*out = new(VvpDeploymentStatus)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlinkDeploymentStatus.
func (in *FlinkDeploymentStatus) DeepCopy() *FlinkDeploymentStatus {
	if in == nil {
		return nil
	}
	out := new(FlinkDeploymentStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlinkDeploymentTemplate) DeepCopyInto(out *FlinkDeploymentTemplate) {
	*out = *in
	if in.Template != nil {
		in, out := &in.Template, &out.Template
		*out = new(VvpDeploymentTemplate)
		(*in).DeepCopyInto(*out)
	}
	if in.CommunityTemplate != nil {
		in, out := &in.CommunityTemplate, &out.CommunityTemplate
		*out = new(CommunityDeploymentTemplate)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlinkDeploymentTemplate.
func (in *FlinkDeploymentTemplate) DeepCopy() *FlinkDeploymentTemplate {
	if in == nil {
		return nil
	}
	out := new(FlinkDeploymentTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Logging) DeepCopyInto(out *Logging) {
	*out = *in
	if in.Log4jLoggers != nil {
		in, out := &in.Log4jLoggers, &out.Log4jLoggers
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Logging.
func (in *Logging) DeepCopy() *Logging {
	if in == nil {
		return nil
	}
	out := new(Logging)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ObjectMeta) DeepCopyInto(out *ObjectMeta) {
	*out = *in
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ObjectMeta.
func (in *ObjectMeta) DeepCopy() *ObjectMeta {
	if in == nil {
		return nil
	}
	out := new(ObjectMeta)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodTemplate) DeepCopyInto(out *PodTemplate) {
	*out = *in
	in.Metadata.DeepCopyInto(&out.Metadata)
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodTemplate.
func (in *PodTemplate) DeepCopy() *PodTemplate {
	if in == nil {
		return nil
	}
	out := new(PodTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodTemplateSpec) DeepCopyInto(out *PodTemplateSpec) {
	*out = *in
	if in.Volumes != nil {
		in, out := &in.Volumes, &out.Volumes
		*out = make([]Volume, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.InitContainers != nil {
		in, out := &in.InitContainers, &out.InitContainers
		*out = make([]Container, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Containers != nil {
		in, out := &in.Containers, &out.Containers
		*out = make([]Container, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.NodeSelector != nil {
		in, out := &in.NodeSelector, &out.NodeSelector
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Affinity != nil {
		in, out := &in.Affinity, &out.Affinity
		*out = new(v1.Affinity)
		(*in).DeepCopyInto(*out)
	}
	if in.Tolerations != nil {
		in, out := &in.Tolerations, &out.Tolerations
		*out = make([]v1.Toleration, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.SecurityContext != nil {
		in, out := &in.SecurityContext, &out.SecurityContext
		*out = new(SecurityContext)
		(*in).DeepCopyInto(*out)
	}
	if in.ImagePullSecrets != nil {
		in, out := &in.ImagePullSecrets, &out.ImagePullSecrets
		*out = make([]v1.LocalObjectReference, len(*in))
		copy(*out, *in)
	}
	if in.ShareProcessNamespace != nil {
		in, out := &in.ShareProcessNamespace, &out.ShareProcessNamespace
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodTemplateSpec.
func (in *PodTemplateSpec) DeepCopy() *PodTemplateSpec {
	if in == nil {
		return nil
	}
	out := new(PodTemplateSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PoolMemberReference) DeepCopyInto(out *PoolMemberReference) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PoolMemberReference.
func (in *PoolMemberReference) DeepCopy() *PoolMemberReference {
	if in == nil {
		return nil
	}
	out := new(PoolMemberReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PoolRef) DeepCopyInto(out *PoolRef) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PoolRef.
func (in *PoolRef) DeepCopy() *PoolRef {
	if in == nil {
		return nil
	}
	out := new(PoolRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceSpec) DeepCopyInto(out *ResourceSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceSpec.
func (in *ResourceSpec) DeepCopy() *ResourceSpec {
	if in == nil {
		return nil
	}
	out := new(ResourceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecretReference) DeepCopyInto(out *SecretReference) {
	*out = *in
	if in.ValueFrom != nil {
		in, out := &in.ValueFrom, &out.ValueFrom
		*out = new(v1.SecretKeySelector)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecretReference.
func (in *SecretReference) DeepCopy() *SecretReference {
	if in == nil {
		return nil
	}
	out := new(SecretReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecurityContext) DeepCopyInto(out *SecurityContext) {
	*out = *in
	if in.RunAsUser != nil {
		in, out := &in.RunAsUser, &out.RunAsUser
		*out = new(int64)
		**out = **in
	}
	if in.RunAsGroup != nil {
		in, out := &in.RunAsGroup, &out.RunAsGroup
		*out = new(int64)
		**out = **in
	}
	if in.RunAsNonRoot != nil {
		in, out := &in.RunAsNonRoot, &out.RunAsNonRoot
		*out = new(bool)
		**out = **in
	}
	if in.FSGroup != nil {
		in, out := &in.FSGroup, &out.FSGroup
		*out = new(int64)
		**out = **in
	}
	if in.ReadOnlyRootFilesystem != nil {
		in, out := &in.ReadOnlyRootFilesystem, &out.ReadOnlyRootFilesystem
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecurityContext.
func (in *SecurityContext) DeepCopy() *SecurityContext {
	if in == nil {
		return nil
	}
	out := new(SecurityContext)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StreamNativeCloudProtocolIntegration) DeepCopyInto(out *StreamNativeCloudProtocolIntegration) {
	*out = *in
	if in.Pulsar != nil {
		in, out := &in.Pulsar, &out.Pulsar
		*out = new(bool)
		**out = **in
	}
	if in.Kafka != nil {
		in, out := &in.Kafka, &out.Kafka
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StreamNativeCloudProtocolIntegration.
func (in *StreamNativeCloudProtocolIntegration) DeepCopy() *StreamNativeCloudProtocolIntegration {
	if in == nil {
		return nil
	}
	out := new(StreamNativeCloudProtocolIntegration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *UserMetadata) DeepCopyInto(out *UserMetadata) {
	*out = *in
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new UserMetadata.
func (in *UserMetadata) DeepCopy() *UserMetadata {
	if in == nil {
		return nil
	}
	out := new(UserMetadata)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Volume) DeepCopyInto(out *Volume) {
	*out = *in
	in.VolumeSource.DeepCopyInto(&out.VolumeSource)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Volume.
func (in *Volume) DeepCopy() *Volume {
	if in == nil {
		return nil
	}
	out := new(Volume)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeSource) DeepCopyInto(out *VolumeSource) {
	*out = *in
	if in.ConfigMap != nil {
		in, out := &in.ConfigMap, &out.ConfigMap
		*out = new(v1.ConfigMapVolumeSource)
		(*in).DeepCopyInto(*out)
	}
	if in.Secret != nil {
		in, out := &in.Secret, &out.Secret
		*out = new(v1.SecretVolumeSource)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeSource.
func (in *VolumeSource) DeepCopy() *VolumeSource {
	if in == nil {
		return nil
	}
	out := new(VolumeSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VvpCustomResourceStatus) DeepCopyInto(out *VvpCustomResourceStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VvpCustomResourceStatus.
func (in *VvpCustomResourceStatus) DeepCopy() *VvpCustomResourceStatus {
	if in == nil {
		return nil
	}
	out := new(VvpCustomResourceStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VvpDeploymentDetails) DeepCopyInto(out *VvpDeploymentDetails) {
	*out = *in
	if in.DeploymentTargetName != nil {
		in, out := &in.DeploymentTargetName, &out.DeploymentTargetName
		*out = new(string)
		**out = **in
	}
	if in.JobFailureExpirationTime != nil {
		in, out := &in.JobFailureExpirationTime, &out.JobFailureExpirationTime
		*out = new(string)
		**out = **in
	}
	if in.MaxJobCreationAttempts != nil {
		in, out := &in.MaxJobCreationAttempts, &out.MaxJobCreationAttempts
		*out = new(int32)
		**out = **in
	}
	if in.MaxSavepointCreationAttempts != nil {
		in, out := &in.MaxSavepointCreationAttempts, &out.MaxSavepointCreationAttempts
		*out = new(int32)
		**out = **in
	}
	if in.RestoreStrategy != nil {
		in, out := &in.RestoreStrategy, &out.RestoreStrategy
		*out = new(VvpRestoreStrategy)
		**out = **in
	}
	if in.SessionClusterName != nil {
		in, out := &in.SessionClusterName, &out.SessionClusterName
		*out = new(string)
		**out = **in
	}
	in.Template.DeepCopyInto(&out.Template)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VvpDeploymentDetails.
func (in *VvpDeploymentDetails) DeepCopy() *VvpDeploymentDetails {
	if in == nil {
		return nil
	}
	out := new(VvpDeploymentDetails)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VvpDeploymentDetailsTemplate) DeepCopyInto(out *VvpDeploymentDetailsTemplate) {
	*out = *in
	if in.Metadata != nil {
		in, out := &in.Metadata, &out.Metadata
		*out = new(VvpDeploymentDetailsTemplateMetadata)
		(*in).DeepCopyInto(*out)
	}
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VvpDeploymentDetailsTemplate.
func (in *VvpDeploymentDetailsTemplate) DeepCopy() *VvpDeploymentDetailsTemplate {
	if in == nil {
		return nil
	}
	out := new(VvpDeploymentDetailsTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VvpDeploymentDetailsTemplateMetadata) DeepCopyInto(out *VvpDeploymentDetailsTemplateMetadata) {
	*out = *in
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VvpDeploymentDetailsTemplateMetadata.
func (in *VvpDeploymentDetailsTemplateMetadata) DeepCopy() *VvpDeploymentDetailsTemplateMetadata {
	if in == nil {
		return nil
	}
	out := new(VvpDeploymentDetailsTemplateMetadata)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VvpDeploymentDetailsTemplateSpec) DeepCopyInto(out *VvpDeploymentDetailsTemplateSpec) {
	*out = *in
	if in.Artifact != nil {
		in, out := &in.Artifact, &out.Artifact
		*out = new(Artifact)
		(*in).DeepCopyInto(*out)
	}
	if in.FlinkConfiguration != nil {
		in, out := &in.FlinkConfiguration, &out.FlinkConfiguration
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Kubernetes != nil {
		in, out := &in.Kubernetes, &out.Kubernetes
		*out = new(VvpDeploymentDetailsTemplateSpecKubernetesSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.LatestCheckpointFetchInterval != nil {
		in, out := &in.LatestCheckpointFetchInterval, &out.LatestCheckpointFetchInterval
		*out = new(int32)
		**out = **in
	}
	if in.Logging != nil {
		in, out := &in.Logging, &out.Logging
		*out = new(Logging)
		(*in).DeepCopyInto(*out)
	}
	if in.NumberOfTaskManagers != nil {
		in, out := &in.NumberOfTaskManagers, &out.NumberOfTaskManagers
		*out = new(int32)
		**out = **in
	}
	if in.Parallelism != nil {
		in, out := &in.Parallelism, &out.Parallelism
		*out = new(int32)
		**out = **in
	}
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = new(VvpDeploymentKubernetesResources)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VvpDeploymentDetailsTemplateSpec.
func (in *VvpDeploymentDetailsTemplateSpec) DeepCopy() *VvpDeploymentDetailsTemplateSpec {
	if in == nil {
		return nil
	}
	out := new(VvpDeploymentDetailsTemplateSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VvpDeploymentDetailsTemplateSpecKubernetesSpec) DeepCopyInto(out *VvpDeploymentDetailsTemplateSpecKubernetesSpec) {
	*out = *in
	if in.JobManagerPodTemplate != nil {
		in, out := &in.JobManagerPodTemplate, &out.JobManagerPodTemplate
		*out = new(PodTemplate)
		(*in).DeepCopyInto(*out)
	}
	if in.TaskManagerPodTemplate != nil {
		in, out := &in.TaskManagerPodTemplate, &out.TaskManagerPodTemplate
		*out = new(PodTemplate)
		(*in).DeepCopyInto(*out)
	}
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VvpDeploymentDetailsTemplateSpecKubernetesSpec.
func (in *VvpDeploymentDetailsTemplateSpecKubernetesSpec) DeepCopy() *VvpDeploymentDetailsTemplateSpecKubernetesSpec {
	if in == nil {
		return nil
	}
	out := new(VvpDeploymentDetailsTemplateSpecKubernetesSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VvpDeploymentKubernetesResources) DeepCopyInto(out *VvpDeploymentKubernetesResources) {
	*out = *in
	if in.Jobmanager != nil {
		in, out := &in.Jobmanager, &out.Jobmanager
		*out = new(ResourceSpec)
		**out = **in
	}
	if in.Taskmanager != nil {
		in, out := &in.Taskmanager, &out.Taskmanager
		*out = new(ResourceSpec)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VvpDeploymentKubernetesResources.
func (in *VvpDeploymentKubernetesResources) DeepCopy() *VvpDeploymentKubernetesResources {
	if in == nil {
		return nil
	}
	out := new(VvpDeploymentKubernetesResources)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VvpDeploymentRunningStatus) DeepCopyInto(out *VvpDeploymentRunningStatus) {
	*out = *in
	in.TransitionTime.DeepCopyInto(&out.TransitionTime)
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]VvpDeploymentStatusCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VvpDeploymentRunningStatus.
func (in *VvpDeploymentRunningStatus) DeepCopy() *VvpDeploymentRunningStatus {
	if in == nil {
		return nil
	}
	out := new(VvpDeploymentRunningStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VvpDeploymentStatus) DeepCopyInto(out *VvpDeploymentStatus) {
	*out = *in
	out.CustomResourceStatus = in.CustomResourceStatus
	in.DeploymentStatus.DeepCopyInto(&out.DeploymentStatus)
	in.DeploymentSystemMetadata.DeepCopyInto(&out.DeploymentSystemMetadata)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VvpDeploymentStatus.
func (in *VvpDeploymentStatus) DeepCopy() *VvpDeploymentStatus {
	if in == nil {
		return nil
	}
	out := new(VvpDeploymentStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VvpDeploymentStatusCondition) DeepCopyInto(out *VvpDeploymentStatusCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VvpDeploymentStatusCondition.
func (in *VvpDeploymentStatusCondition) DeepCopy() *VvpDeploymentStatusCondition {
	if in == nil {
		return nil
	}
	out := new(VvpDeploymentStatusCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VvpDeploymentStatusDeploymentStatus) DeepCopyInto(out *VvpDeploymentStatusDeploymentStatus) {
	*out = *in
	in.Running.DeepCopyInto(&out.Running)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VvpDeploymentStatusDeploymentStatus.
func (in *VvpDeploymentStatusDeploymentStatus) DeepCopy() *VvpDeploymentStatusDeploymentStatus {
	if in == nil {
		return nil
	}
	out := new(VvpDeploymentStatusDeploymentStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VvpDeploymentSystemMetadata) DeepCopyInto(out *VvpDeploymentSystemMetadata) {
	*out = *in
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	in.CreatedAt.DeepCopyInto(&out.CreatedAt)
	in.ModifiedAt.DeepCopyInto(&out.ModifiedAt)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VvpDeploymentSystemMetadata.
func (in *VvpDeploymentSystemMetadata) DeepCopy() *VvpDeploymentSystemMetadata {
	if in == nil {
		return nil
	}
	out := new(VvpDeploymentSystemMetadata)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VvpDeploymentTemplate) DeepCopyInto(out *VvpDeploymentTemplate) {
	*out = *in
	in.Deployment.DeepCopyInto(&out.Deployment)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VvpDeploymentTemplate.
func (in *VvpDeploymentTemplate) DeepCopy() *VvpDeploymentTemplate {
	if in == nil {
		return nil
	}
	out := new(VvpDeploymentTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VvpDeploymentTemplateSpec) DeepCopyInto(out *VvpDeploymentTemplateSpec) {
	*out = *in
	if in.UserMetadata != nil {
		in, out := &in.UserMetadata, &out.UserMetadata
		*out = new(UserMetadata)
		(*in).DeepCopyInto(*out)
	}
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VvpDeploymentTemplateSpec.
func (in *VvpDeploymentTemplateSpec) DeepCopy() *VvpDeploymentTemplateSpec {
	if in == nil {
		return nil
	}
	out := new(VvpDeploymentTemplateSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VvpRestoreStrategy) DeepCopyInto(out *VvpRestoreStrategy) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VvpRestoreStrategy.
func (in *VvpRestoreStrategy) DeepCopy() *VvpRestoreStrategy {
	if in == nil {
		return nil
	}
	out := new(VvpRestoreStrategy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Workspace) DeepCopyInto(out *Workspace) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Workspace.
func (in *Workspace) DeepCopy() *Workspace {
	if in == nil {
		return nil
	}
	out := new(Workspace)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Workspace) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WorkspaceList) DeepCopyInto(out *WorkspaceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Workspace, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WorkspaceList.
func (in *WorkspaceList) DeepCopy() *WorkspaceList {
	if in == nil {
		return nil
	}
	out := new(WorkspaceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *WorkspaceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WorkspaceReference) DeepCopyInto(out *WorkspaceReference) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WorkspaceReference.
func (in *WorkspaceReference) DeepCopy() *WorkspaceReference {
	if in == nil {
		return nil
	}
	out := new(WorkspaceReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WorkspaceSpec) DeepCopyInto(out *WorkspaceSpec) {
	*out = *in
	if in.PulsarClusterNames != nil {
		in, out := &in.PulsarClusterNames, &out.PulsarClusterNames
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.PoolRef != nil {
		in, out := &in.PoolRef, &out.PoolRef
		*out = new(PoolRef)
		**out = **in
	}
	if in.UseExternalAccess != nil {
		in, out := &in.UseExternalAccess, &out.UseExternalAccess
		*out = new(bool)
		**out = **in
	}
	if in.FlinkBlobStorage != nil {
		in, out := &in.FlinkBlobStorage, &out.FlinkBlobStorage
		*out = new(FlinkBlobStorage)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WorkspaceSpec.
func (in *WorkspaceSpec) DeepCopy() *WorkspaceSpec {
	if in == nil {
		return nil
	}
	out := new(WorkspaceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WorkspaceStatus) DeepCopyInto(out *WorkspaceStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WorkspaceStatus.
func (in *WorkspaceStatus) DeepCopy() *WorkspaceStatus {
	if in == nil {
		return nil
	}
	out := new(WorkspaceStatus)
	in.DeepCopyInto(out)
	return out
}
