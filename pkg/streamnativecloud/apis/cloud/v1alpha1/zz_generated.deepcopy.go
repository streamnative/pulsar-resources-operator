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
	"k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSCloudConnection) DeepCopyInto(out *AWSCloudConnection) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSCloudConnection.
func (in *AWSCloudConnection) DeepCopy() *AWSCloudConnection {
	if in == nil {
		return nil
	}
	out := new(AWSCloudConnection)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AmqpConfig) DeepCopyInto(out *AmqpConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AmqpConfig.
func (in *AmqpConfig) DeepCopy() *AmqpConfig {
	if in == nil {
		return nil
	}
	out := new(AmqpConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AuditLog) DeepCopyInto(out *AuditLog) {
	*out = *in
	if in.Categories != nil {
		in, out := &in.Categories, &out.Categories
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AuditLog.
func (in *AuditLog) DeepCopy() *AuditLog {
	if in == nil {
		return nil
	}
	out := new(AuditLog)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AzureConnection) DeepCopyInto(out *AzureConnection) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AzureConnection.
func (in *AzureConnection) DeepCopy() *AzureConnection {
	if in == nil {
		return nil
	}
	out := new(AzureConnection)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BookKeeperSetReference) DeepCopyInto(out *BookKeeperSetReference) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BookKeeperSetReference.
func (in *BookKeeperSetReference) DeepCopy() *BookKeeperSetReference {
	if in == nil {
		return nil
	}
	out := new(BookKeeperSetReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Config) DeepCopyInto(out *Config) {
	*out = *in
	if in.WebsocketEnabled != nil {
		in, out := &in.WebsocketEnabled, &out.WebsocketEnabled
		*out = new(bool)
		**out = **in
	}
	if in.FunctionEnabled != nil {
		in, out := &in.FunctionEnabled, &out.FunctionEnabled
		*out = new(bool)
		**out = **in
	}
	if in.TransactionEnabled != nil {
		in, out := &in.TransactionEnabled, &out.TransactionEnabled
		*out = new(bool)
		**out = **in
	}
	if in.Protocols != nil {
		in, out := &in.Protocols, &out.Protocols
		*out = new(ProtocolsConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.AuditLog != nil {
		in, out := &in.AuditLog, &out.AuditLog
		*out = new(AuditLog)
		(*in).DeepCopyInto(*out)
	}
	if in.LakehouseStorage != nil {
		in, out := &in.LakehouseStorage, &out.LakehouseStorage
		*out = new(LakehouseStorageConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.Custom != nil {
		in, out := &in.Custom, &out.Custom
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Config.
func (in *Config) DeepCopy() *Config {
	if in == nil {
		return nil
	}
	out := new(Config)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Domain) DeepCopyInto(out *Domain) {
	*out = *in
	if in.TLS != nil {
		in, out := &in.TLS, &out.TLS
		*out = new(DomainTLS)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Domain.
func (in *Domain) DeepCopy() *Domain {
	if in == nil {
		return nil
	}
	out := new(Domain)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DomainTLS) DeepCopyInto(out *DomainTLS) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DomainTLS.
func (in *DomainTLS) DeepCopy() *DomainTLS {
	if in == nil {
		return nil
	}
	out := new(DomainTLS)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GCPCloudConnection) DeepCopyInto(out *GCPCloudConnection) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GCPCloudConnection.
func (in *GCPCloudConnection) DeepCopy() *GCPCloudConnection {
	if in == nil {
		return nil
	}
	out := new(GCPCloudConnection)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KafkaConfig) DeepCopyInto(out *KafkaConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KafkaConfig.
func (in *KafkaConfig) DeepCopy() *KafkaConfig {
	if in == nil {
		return nil
	}
	out := new(KafkaConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LakehouseStorageConfig) DeepCopyInto(out *LakehouseStorageConfig) {
	*out = *in
	if in.LakehouseType != nil {
		in, out := &in.LakehouseType, &out.LakehouseType
		*out = new(string)
		**out = **in
	}
	if in.CatalogType != nil {
		in, out := &in.CatalogType, &out.CatalogType
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LakehouseStorageConfig.
func (in *LakehouseStorageConfig) DeepCopy() *LakehouseStorageConfig {
	if in == nil {
		return nil
	}
	out := new(LakehouseStorageConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MqttConfig) DeepCopyInto(out *MqttConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MqttConfig.
func (in *MqttConfig) DeepCopy() *MqttConfig {
	if in == nil {
		return nil
	}
	out := new(MqttConfig)
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
func (in *ProtocolsConfig) DeepCopyInto(out *ProtocolsConfig) {
	*out = *in
	if in.Kafka != nil {
		in, out := &in.Kafka, &out.Kafka
		*out = new(KafkaConfig)
		**out = **in
	}
	if in.Amqp != nil {
		in, out := &in.Amqp, &out.Amqp
		*out = new(AmqpConfig)
		**out = **in
	}
	if in.Mqtt != nil {
		in, out := &in.Mqtt, &out.Mqtt
		*out = new(MqttConfig)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ProtocolsConfig.
func (in *ProtocolsConfig) DeepCopy() *ProtocolsConfig {
	if in == nil {
		return nil
	}
	out := new(ProtocolsConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RoleRef) DeepCopyInto(out *RoleRef) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RoleRef.
func (in *RoleRef) DeepCopy() *RoleRef {
	if in == nil {
		return nil
	}
	out := new(RoleRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Secret) DeepCopyInto(out *Secret) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
	if in.Data != nil {
		in, out := &in.Data, &out.Data
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.PoolMemberRef != nil {
		in, out := &in.PoolMemberRef, &out.PoolMemberRef
		*out = new(PoolMemberReference)
		**out = **in
	}
	if in.Tolerations != nil {
		in, out := &in.Tolerations, &out.Tolerations
		*out = make([]Toleration, len(*in))
		copy(*out, *in)
	}
	if in.Type != nil {
		in, out := &in.Type, &out.Type
		*out = new(v1.SecretType)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Secret.
func (in *Secret) DeepCopy() *Secret {
	if in == nil {
		return nil
	}
	out := new(Secret)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Secret) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecretList) DeepCopyInto(out *SecretList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Secret, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecretList.
func (in *SecretList) DeepCopy() *SecretList {
	if in == nil {
		return nil
	}
	out := new(SecretList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SecretList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecretRef) DeepCopyInto(out *SecretRef) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecretRef.
func (in *SecretRef) DeepCopy() *SecretRef {
	if in == nil {
		return nil
	}
	out := new(SecretRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecretSpec) DeepCopyInto(out *SecretSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecretSpec.
func (in *SecretSpec) DeepCopy() *SecretSpec {
	if in == nil {
		return nil
	}
	out := new(SecretSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecretStatus) DeepCopyInto(out *SecretStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecretStatus.
func (in *SecretStatus) DeepCopy() *SecretStatus {
	if in == nil {
		return nil
	}
	out := new(SecretStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SharingConfig) DeepCopyInto(out *SharingConfig) {
	*out = *in
	if in.Namespaces != nil {
		in, out := &in.Namespaces, &out.Namespaces
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SharingConfig.
func (in *SharingConfig) DeepCopy() *SharingConfig {
	if in == nil {
		return nil
	}
	out := new(SharingConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Subject) DeepCopyInto(out *Subject) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Subject.
func (in *Subject) DeepCopy() *Subject {
	if in == nil {
		return nil
	}
	out := new(Subject)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Taint) DeepCopyInto(out *Taint) {
	*out = *in
	if in.TimeAdded != nil {
		in, out := &in.TimeAdded, &out.TimeAdded
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Taint.
func (in *Taint) DeepCopy() *Taint {
	if in == nil {
		return nil
	}
	out := new(Taint)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Toleration) DeepCopyInto(out *Toleration) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Toleration.
func (in *Toleration) DeepCopy() *Toleration {
	if in == nil {
		return nil
	}
	out := new(Toleration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ZooKeeperSetReference) DeepCopyInto(out *ZooKeeperSetReference) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ZooKeeperSetReference.
func (in *ZooKeeperSetReference) DeepCopy() *ZooKeeperSetReference {
	if in == nil {
		return nil
	}
	out := new(ZooKeeperSetReference)
	in.DeepCopyInto(out)
	return out
}
