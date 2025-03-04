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

import "k8s.io/apimachinery/pkg/types"

type DomainType string

const (
	DomainTypeService DomainType = "service"
	DomainTypeIngress DomainType = ""
)

const (
	AnnotationReconcilePaused = "cloud.streamnative.io/reconcile-paused"
	AnnotationReconcileMode   = "cloud.streamnative.io/reconcile-mode"
)

const (
	// ReconcileModeMinimal mode only reconcile minimal resources as necessary to keep the cluster running properly like certificates
	ReconcileModeMinimal = "minimal"
	// ReconcileModeBasic mode reconcile resources before zookeeper
	ReconcileModeBasic = "basic"
)

// SecretRef is a reference to a Cloud Secret with a given name
type SecretRef struct {
	// Namespace is the secret namespace
	// +optional
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,1,opt,name=namespace"`

	// Name is the secret name
	Name string `json:"name,omitempty" protobuf:"bytes,2,opt,name=name"`
}

func (s *SecretRef) ToNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: s.Namespace,
		Name:      s.Name,
	}
}

type Domain struct {
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// +optional
	Type DomainType `json:"type,omitempty" protobuf:"bytes,2,opt,name=type"`
	// +optional
	TLS *DomainTLS `json:"tls,omitempty" protobuf:"bytes,3,opt,name=tls"`
}

type DomainTLS struct {
	// CertificateName specifies the certificate object name to use.
	// +optional
	CertificateName string `json:"certificateName,omitempty" protobuf:"bytes,1,opt,name=name"`
}

// PoolRef is a reference to a pool with a given name.
type PoolRef struct {
	Namespace string `json:"namespace" protobuf:"bytes,1,opt,name=namespace"`
	Name      string `json:"name" protobuf:"bytes,2,opt,name=name"`
}

func (r PoolRef) ToNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: r.Namespace,
		Name:      r.Name,
	}
}

// PoolMemberReference is a reference to a pool member with a given name.
type PoolMemberReference struct {
	Namespace string `json:"namespace" protobuf:"bytes,1,opt,name=namespace"`
	Name      string `json:"name" protobuf:"bytes,2,opt,name=name"`
}

func (r PoolMemberReference) ToNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: r.Namespace,
		Name:      r.Name,
	}
}

// ZooKeeperSetReference is a fully-qualified reference to a ZooKeeperSet with a given name.
type ZooKeeperSetReference struct {
	// +optional
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,1,opt,name=namespace"`
	Name      string `json:"name" protobuf:"bytes,2,opt,name=name"`
}

func (r *ZooKeeperSetReference) ToNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: r.Namespace,
		Name:      r.Name,
	}
}

// BookKeeperSetReference is a fully-qualified reference to a BookKeeperSet with a given name.
type BookKeeperSetReference struct {
	// +optional
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,1,opt,name=namespace"`
	Name      string `json:"name" protobuf:"bytes,2,opt,name=name"`
}

func (r *BookKeeperSetReference) ToNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: r.Namespace,
		Name:      r.Name,
	}
}

// PullPolicy describes a policy for if/when to pull a container image
type PullPolicy string

const (
	// PullAlways means that kubelet always attempts to pull the latest image. Container will fail If the pull fails.
	PullAlways PullPolicy = "Always"
	// PullNever means that kubelet never pulls an image, but only uses a local image. Container will fail if the image isn't present
	PullNever PullPolicy = "Never"
	// PullIfNotPresent means that kubelet pulls if the image isn't present on disk. Container will fail if the image isn't present and the pull fails.
	PullIfNotPresent PullPolicy = "IfNotPresent"
)

type Config struct {
	// WebsocketEnabled controls whether websocket is enabled.
	// +optional
	WebsocketEnabled *bool `json:"websocketEnabled,omitempty" protobuf:"varint,1,opt,name=websocketEnabled"`

	// FunctionEnabled controls whether function is enabled.
	// +optional
	FunctionEnabled *bool `json:"functionEnabled,omitempty" protobuf:"varint,2,opt,name=functionEnabled"`

	// TransactionEnabled controls whether transaction is enabled.
	// +optional
	TransactionEnabled *bool `json:"transactionEnabled,omitempty" protobuf:"varint,3,opt,name=transactionEnabled"`

	// Protocols controls the protocols enabled in brokers.
	// +optional
	Protocols *ProtocolsConfig `json:"protocols,omitempty" protobuf:"bytes,5,opt,name=protocols"`

	// AuditLog controls the custom config of audit log.
	// +optional
	AuditLog *AuditLog `json:"auditLog,omitempty" protobuf:"bytes,6,opt,name=auditLog"`

	// LakehouseStorage controls the lakehouse storage config.
	// +optional
	LakehouseStorage *LakehouseStorageConfig `json:"lakehouseStorage,omitempty" protobuf:"bytes,7,opt,name=lakehouseStorage"`

	// Custom accepts custom configurations.
	// +optional
	Custom map[string]string `json:"custom,omitempty" protobuf:"bytes,4,rep,name=custom"`
}

type ProtocolsConfig struct {
	// Kafka controls whether to enable Kafka protocol in brokers
	// +optional
	Kafka *KafkaConfig `json:"kafka,omitempty" protobuf:"bytes,1,opt,name=kafka"`

	// Amqp controls whether to enable Amqp protocol in brokers
	// +optional
	Amqp *AmqpConfig `json:"amqp,omitempty" protobuf:"bytes,2,opt,name=amqp"`

	// Mqtt controls whether to enable mqtt protocol in brokers
	// +optional
	Mqtt *MqttConfig `json:"mqtt,omitempty" protobuf:"bytes,3,opt,name=mqtt"`
}

type KafkaConfig struct {
}

type AmqpConfig struct {
}

// MqttConfig defines the mqtt protocol config
type MqttConfig struct {
}

type Subject struct {
	Kind string `json:"kind" protobuf:"bytes,1,opt,name=kind"`
	Name string `json:"name" protobuf:"bytes,2,opt,name=name"`
	// +optional
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,3,opt,name=namespace"`
	APIGroup  string `json:"apiGroup" protobuf:"bytes,4,opt,name=apiGroup"`
}

type RoleRef struct {
	Kind     string `json:"kind" protobuf:"bytes,1,opt,name=kind"`
	Name     string `json:"name" protobuf:"bytes,2,opt,name=name"`
	APIGroup string `json:"apiGroup" protobuf:"bytes,3,opt,name=apiGroup"`
}

type SharingConfig struct {
	// +listType=atomic
	Namespaces []string `json:"namespaces" protobuf:"bytes,1,rep,name=namespaces"`
}

type AuditLog struct {
	// +listType=set
	Categories []string `json:"categories" protobuf:"bytes,1,rep,name=categories"`
}

type LakehouseStorageConfig struct {

	// +optional
	// +kubebuilder:validation:Enum:=iceberg
	// +kubebuilder:default=iceberg
	LakehouseType *string `json:"lakehouseType,omitempty" protobuf:"bytes,1,opt,name=lakehouseType"`

	// +optional
	// +kubebuilder:validation:Enum:=rest
	// +kubebuilder:default=rest
	CatalogType *string `json:"catalogType,omitempty" protobuf:"bytes,2,opt,name=catalogType"`

	// todo: maybe we need to support mount secrets as the catalog credentials?
	// +kubebuilder:validation:Required
	CatalogCredentials string `json:"catalogCredentials" protobuf:"bytes,3,opt,name=catalogCredentials"`

	// +kubebuilder:validation:Required
	CatalogConnectionUrl string `json:"catalogConnectionUrl" protobuf:"bytes,4,opt,name=catalogConnectionUrl"`

	// +kubebuilder:validation:Required
	CatalogWarehouse string `json:"catalogWarehouse" protobuf:"bytes,5,opt,name=catalogWarehouse"`
}

type AWSCloudConnection struct {
	AccountId string `json:"accountId" protobuf:"bytes,1,name=accountId"`
}

type GCPCloudConnection struct {
	ProjectId string `json:"projectId" protobuf:"bytes,1,name=projectId"`
}

type AzureConnection struct {
	SubscriptionId  string `json:"subscriptionId" protobuf:"bytes,1,name=subscriptionId"`
	TenantId        string `json:"tenantId" protobuf:"bytes,2,name=tenantId"`
	ClientId        string `json:"clientId" protobuf:"bytes,3,name=clientId"`
	SupportClientId string `json:"supportClientId" protobuf:"bytes,4,name=supportClientId"`
}
