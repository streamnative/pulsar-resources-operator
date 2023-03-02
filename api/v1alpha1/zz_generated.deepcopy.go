// Copyright 2023 StreamNative
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

//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterInfo) DeepCopyInto(out *ClusterInfo) {
	*out = *in
	out.ConnectionRef = in.ConnectionRef
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterInfo.
func (in *ClusterInfo) DeepCopy() *ClusterInfo {
	if in == nil {
		return nil
	}
	out := new(ClusterInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PulsarAuthentication) DeepCopyInto(out *PulsarAuthentication) {
	*out = *in
	if in.Token != nil {
		in, out := &in.Token, &out.Token
		*out = new(ValueOrSecretRef)
		(*in).DeepCopyInto(*out)
	}
	if in.OAuth2 != nil {
		in, out := &in.OAuth2, &out.OAuth2
		*out = new(PulsarAuthenticationOAuth2)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PulsarAuthentication.
func (in *PulsarAuthentication) DeepCopy() *PulsarAuthentication {
	if in == nil {
		return nil
	}
	out := new(PulsarAuthentication)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PulsarAuthenticationOAuth2) DeepCopyInto(out *PulsarAuthenticationOAuth2) {
	*out = *in
	in.Key.DeepCopyInto(&out.Key)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PulsarAuthenticationOAuth2.
func (in *PulsarAuthenticationOAuth2) DeepCopy() *PulsarAuthenticationOAuth2 {
	if in == nil {
		return nil
	}
	out := new(PulsarAuthenticationOAuth2)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PulsarConnection) DeepCopyInto(out *PulsarConnection) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PulsarConnection.
func (in *PulsarConnection) DeepCopy() *PulsarConnection {
	if in == nil {
		return nil
	}
	out := new(PulsarConnection)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PulsarConnection) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PulsarConnectionList) DeepCopyInto(out *PulsarConnectionList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]PulsarConnection, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PulsarConnectionList.
func (in *PulsarConnectionList) DeepCopy() *PulsarConnectionList {
	if in == nil {
		return nil
	}
	out := new(PulsarConnectionList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PulsarConnectionList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PulsarConnectionSpec) DeepCopyInto(out *PulsarConnectionSpec) {
	*out = *in
	if in.Authentication != nil {
		in, out := &in.Authentication, &out.Authentication
		*out = new(PulsarAuthentication)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PulsarConnectionSpec.
func (in *PulsarConnectionSpec) DeepCopy() *PulsarConnectionSpec {
	if in == nil {
		return nil
	}
	out := new(PulsarConnectionSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PulsarConnectionStatus) DeepCopyInto(out *PulsarConnectionStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PulsarConnectionStatus.
func (in *PulsarConnectionStatus) DeepCopy() *PulsarConnectionStatus {
	if in == nil {
		return nil
	}
	out := new(PulsarConnectionStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PulsarGeoReplication) DeepCopyInto(out *PulsarGeoReplication) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PulsarGeoReplication.
func (in *PulsarGeoReplication) DeepCopy() *PulsarGeoReplication {
	if in == nil {
		return nil
	}
	out := new(PulsarGeoReplication)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PulsarGeoReplication) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PulsarGeoReplicationList) DeepCopyInto(out *PulsarGeoReplicationList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]PulsarGeoReplication, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PulsarGeoReplicationList.
func (in *PulsarGeoReplicationList) DeepCopy() *PulsarGeoReplicationList {
	if in == nil {
		return nil
	}
	out := new(PulsarGeoReplicationList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PulsarGeoReplicationList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PulsarGeoReplicationSpec) DeepCopyInto(out *PulsarGeoReplicationSpec) {
	*out = *in
	out.SourceCluster = in.SourceCluster
	out.DestinationCluster = in.DestinationCluster
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PulsarGeoReplicationSpec.
func (in *PulsarGeoReplicationSpec) DeepCopy() *PulsarGeoReplicationSpec {
	if in == nil {
		return nil
	}
	out := new(PulsarGeoReplicationSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PulsarGeoReplicationStatus) DeepCopyInto(out *PulsarGeoReplicationStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PulsarGeoReplicationStatus.
func (in *PulsarGeoReplicationStatus) DeepCopy() *PulsarGeoReplicationStatus {
	if in == nil {
		return nil
	}
	out := new(PulsarGeoReplicationStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PulsarNamespace) DeepCopyInto(out *PulsarNamespace) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PulsarNamespace.
func (in *PulsarNamespace) DeepCopy() *PulsarNamespace {
	if in == nil {
		return nil
	}
	out := new(PulsarNamespace)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PulsarNamespace) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PulsarNamespaceList) DeepCopyInto(out *PulsarNamespaceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]PulsarNamespace, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PulsarNamespaceList.
func (in *PulsarNamespaceList) DeepCopy() *PulsarNamespaceList {
	if in == nil {
		return nil
	}
	out := new(PulsarNamespaceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PulsarNamespaceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PulsarNamespaceSpec) DeepCopyInto(out *PulsarNamespaceSpec) {
	*out = *in
	if in.Bundles != nil {
		in, out := &in.Bundles, &out.Bundles
		*out = new(int32)
		**out = **in
	}
	out.ConnectionRef = in.ConnectionRef
	if in.MaxProducersPerTopic != nil {
		in, out := &in.MaxProducersPerTopic, &out.MaxProducersPerTopic
		*out = new(int32)
		**out = **in
	}
	if in.MaxConsumersPerTopic != nil {
		in, out := &in.MaxConsumersPerTopic, &out.MaxConsumersPerTopic
		*out = new(int32)
		**out = **in
	}
	if in.MaxConsumersPerSubscription != nil {
		in, out := &in.MaxConsumersPerSubscription, &out.MaxConsumersPerSubscription
		*out = new(int32)
		**out = **in
	}
	if in.MessageTTL != nil {
		in, out := &in.MessageTTL, &out.MessageTTL
		*out = new(v1.Duration)
		**out = **in
	}
	if in.RetentionTime != nil {
		in, out := &in.RetentionTime, &out.RetentionTime
		*out = new(v1.Duration)
		**out = **in
	}
	if in.RetentionSize != nil {
		in, out := &in.RetentionSize, &out.RetentionSize
		x := (*in).DeepCopy()
		*out = &x
	}
	if in.BacklogQuotaLimitTime != nil {
		in, out := &in.BacklogQuotaLimitTime, &out.BacklogQuotaLimitTime
		*out = new(v1.Duration)
		**out = **in
	}
	if in.BacklogQuotaLimitSize != nil {
		in, out := &in.BacklogQuotaLimitSize, &out.BacklogQuotaLimitSize
		x := (*in).DeepCopy()
		*out = &x
	}
	if in.BacklogQuotaRetentionPolicy != nil {
		in, out := &in.BacklogQuotaRetentionPolicy, &out.BacklogQuotaRetentionPolicy
		*out = new(string)
		**out = **in
	}
	if in.BacklogQuotaType != nil {
		in, out := &in.BacklogQuotaType, &out.BacklogQuotaType
		*out = new(string)
		**out = **in
	}
	if in.GeoReplicationRefs != nil {
		in, out := &in.GeoReplicationRefs, &out.GeoReplicationRefs
		*out = make([]*corev1.LocalObjectReference, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(corev1.LocalObjectReference)
				**out = **in
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PulsarNamespaceSpec.
func (in *PulsarNamespaceSpec) DeepCopy() *PulsarNamespaceSpec {
	if in == nil {
		return nil
	}
	out := new(PulsarNamespaceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PulsarNamespaceStatus) DeepCopyInto(out *PulsarNamespaceStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PulsarNamespaceStatus.
func (in *PulsarNamespaceStatus) DeepCopy() *PulsarNamespaceStatus {
	if in == nil {
		return nil
	}
	out := new(PulsarNamespaceStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PulsarPermission) DeepCopyInto(out *PulsarPermission) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PulsarPermission.
func (in *PulsarPermission) DeepCopy() *PulsarPermission {
	if in == nil {
		return nil
	}
	out := new(PulsarPermission)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PulsarPermission) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PulsarPermissionList) DeepCopyInto(out *PulsarPermissionList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]PulsarPermission, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PulsarPermissionList.
func (in *PulsarPermissionList) DeepCopy() *PulsarPermissionList {
	if in == nil {
		return nil
	}
	out := new(PulsarPermissionList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PulsarPermissionList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PulsarPermissionSpec) DeepCopyInto(out *PulsarPermissionSpec) {
	*out = *in
	out.ConnectionRef = in.ConnectionRef
	if in.Roles != nil {
		in, out := &in.Roles, &out.Roles
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Actions != nil {
		in, out := &in.Actions, &out.Actions
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PulsarPermissionSpec.
func (in *PulsarPermissionSpec) DeepCopy() *PulsarPermissionSpec {
	if in == nil {
		return nil
	}
	out := new(PulsarPermissionSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PulsarPermissionStatus) DeepCopyInto(out *PulsarPermissionStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PulsarPermissionStatus.
func (in *PulsarPermissionStatus) DeepCopy() *PulsarPermissionStatus {
	if in == nil {
		return nil
	}
	out := new(PulsarPermissionStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PulsarTenant) DeepCopyInto(out *PulsarTenant) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PulsarTenant.
func (in *PulsarTenant) DeepCopy() *PulsarTenant {
	if in == nil {
		return nil
	}
	out := new(PulsarTenant)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PulsarTenant) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PulsarTenantList) DeepCopyInto(out *PulsarTenantList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]PulsarTenant, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PulsarTenantList.
func (in *PulsarTenantList) DeepCopy() *PulsarTenantList {
	if in == nil {
		return nil
	}
	out := new(PulsarTenantList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PulsarTenantList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PulsarTenantSpec) DeepCopyInto(out *PulsarTenantSpec) {
	*out = *in
	out.ConnectionRef = in.ConnectionRef
	if in.AdminRoles != nil {
		in, out := &in.AdminRoles, &out.AdminRoles
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.AllowedClusters != nil {
		in, out := &in.AllowedClusters, &out.AllowedClusters
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.GeoReplicationRefs != nil {
		in, out := &in.GeoReplicationRefs, &out.GeoReplicationRefs
		*out = make([]*corev1.LocalObjectReference, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(corev1.LocalObjectReference)
				**out = **in
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PulsarTenantSpec.
func (in *PulsarTenantSpec) DeepCopy() *PulsarTenantSpec {
	if in == nil {
		return nil
	}
	out := new(PulsarTenantSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PulsarTenantStatus) DeepCopyInto(out *PulsarTenantStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PulsarTenantStatus.
func (in *PulsarTenantStatus) DeepCopy() *PulsarTenantStatus {
	if in == nil {
		return nil
	}
	out := new(PulsarTenantStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PulsarTopic) DeepCopyInto(out *PulsarTopic) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PulsarTopic.
func (in *PulsarTopic) DeepCopy() *PulsarTopic {
	if in == nil {
		return nil
	}
	out := new(PulsarTopic)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PulsarTopic) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PulsarTopicList) DeepCopyInto(out *PulsarTopicList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]PulsarTopic, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PulsarTopicList.
func (in *PulsarTopicList) DeepCopy() *PulsarTopicList {
	if in == nil {
		return nil
	}
	out := new(PulsarTopicList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PulsarTopicList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PulsarTopicSpec) DeepCopyInto(out *PulsarTopicSpec) {
	*out = *in
	if in.Persistent != nil {
		in, out := &in.Persistent, &out.Persistent
		*out = new(bool)
		**out = **in
	}
	if in.Partitions != nil {
		in, out := &in.Partitions, &out.Partitions
		*out = new(int32)
		**out = **in
	}
	out.ConnectionRef = in.ConnectionRef
	if in.MaxProducers != nil {
		in, out := &in.MaxProducers, &out.MaxProducers
		*out = new(int32)
		**out = **in
	}
	if in.MaxConsumers != nil {
		in, out := &in.MaxConsumers, &out.MaxConsumers
		*out = new(int32)
		**out = **in
	}
	if in.MessageTTL != nil {
		in, out := &in.MessageTTL, &out.MessageTTL
		*out = new(v1.Duration)
		**out = **in
	}
	if in.MaxUnAckedMessagesPerConsumer != nil {
		in, out := &in.MaxUnAckedMessagesPerConsumer, &out.MaxUnAckedMessagesPerConsumer
		*out = new(int32)
		**out = **in
	}
	if in.MaxUnAckedMessagesPerSubscription != nil {
		in, out := &in.MaxUnAckedMessagesPerSubscription, &out.MaxUnAckedMessagesPerSubscription
		*out = new(int32)
		**out = **in
	}
	if in.RetentionTime != nil {
		in, out := &in.RetentionTime, &out.RetentionTime
		*out = new(v1.Duration)
		**out = **in
	}
	if in.RetentionSize != nil {
		in, out := &in.RetentionSize, &out.RetentionSize
		x := (*in).DeepCopy()
		*out = &x
	}
	if in.BacklogQuotaLimitTime != nil {
		in, out := &in.BacklogQuotaLimitTime, &out.BacklogQuotaLimitTime
		*out = new(v1.Duration)
		**out = **in
	}
	if in.BacklogQuotaLimitSize != nil {
		in, out := &in.BacklogQuotaLimitSize, &out.BacklogQuotaLimitSize
		x := (*in).DeepCopy()
		*out = &x
	}
	if in.BacklogQuotaRetentionPolicy != nil {
		in, out := &in.BacklogQuotaRetentionPolicy, &out.BacklogQuotaRetentionPolicy
		*out = new(string)
		**out = **in
	}
	if in.SchemaInfo != nil {
		in, out := &in.SchemaInfo, &out.SchemaInfo
		*out = new(SchemaInfo)
		(*in).DeepCopyInto(*out)
	}
	if in.GeoReplicationRefs != nil {
		in, out := &in.GeoReplicationRefs, &out.GeoReplicationRefs
		*out = make([]*corev1.LocalObjectReference, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(corev1.LocalObjectReference)
				**out = **in
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PulsarTopicSpec.
func (in *PulsarTopicSpec) DeepCopy() *PulsarTopicSpec {
	if in == nil {
		return nil
	}
	out := new(PulsarTopicSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PulsarTopicStatus) DeepCopyInto(out *PulsarTopicStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PulsarTopicStatus.
func (in *PulsarTopicStatus) DeepCopy() *PulsarTopicStatus {
	if in == nil {
		return nil
	}
	out := new(PulsarTopicStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SchemaInfo) DeepCopyInto(out *SchemaInfo) {
	*out = *in
	if in.Properties != nil {
		in, out := &in.Properties, &out.Properties
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SchemaInfo.
func (in *SchemaInfo) DeepCopy() *SchemaInfo {
	if in == nil {
		return nil
	}
	out := new(SchemaInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecretKeyRef) DeepCopyInto(out *SecretKeyRef) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecretKeyRef.
func (in *SecretKeyRef) DeepCopy() *SecretKeyRef {
	if in == nil {
		return nil
	}
	out := new(SecretKeyRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ValueOrSecretRef) DeepCopyInto(out *ValueOrSecretRef) {
	*out = *in
	if in.Value != nil {
		in, out := &in.Value, &out.Value
		*out = new(string)
		**out = **in
	}
	if in.SecretRef != nil {
		in, out := &in.SecretRef, &out.SecretRef
		*out = new(SecretKeyRef)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ValueOrSecretRef.
func (in *ValueOrSecretRef) DeepCopy() *ValueOrSecretRef {
	if in == nil {
		return nil
	}
	out := new(ValueOrSecretRef)
	in.DeepCopyInto(out)
	return out
}
