// Copyright 2024 StreamNative
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
	"reflect"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/streamnative/pulsar-resources-operator/pkg/reconciler"
)

// SecretKeyRef indicates a secret name and key
type SecretKeyRef struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

// ValueOrSecretRef is a string or a secret reference of the authentication
type ValueOrSecretRef struct {
	// +optional
	Value *string `json:"value,omitempty"`

	// +optional
	SecretRef *SecretKeyRef `json:"secretRef,omitempty"`
}

// PulsarAuthentication defines the authentication configuration for Pulsar resources.
// It supports two authentication methods: Token-based and OAuth2-based.
// Only one authentication method should be specified at a time.
type PulsarAuthentication struct {
	// Token specifies the configuration for token-based authentication.
	// This can be either a direct token value or a reference to a secret containing the token.
	// If using a secret, the token should be stored under the specified key in the secret.
	// +optional
	Token *ValueOrSecretRef `json:"token,omitempty"`

	// OAuth2 specifies the configuration for OAuth2-based authentication.
	// This includes all necessary parameters for setting up OAuth2 authentication with Pulsar.
	// For detailed information on the OAuth2 fields, refer to the PulsarAuthenticationOAuth2 struct.
	// +optional
	OAuth2 *PulsarAuthenticationOAuth2 `json:"oauth2,omitempty"`
}

// PulsarResourceLifeCyclePolicy defines the behavior for managing Pulsar resources
// when the corresponding custom resource (CR) is deleted from the Kubernetes cluster.
// This policy allows users to control whether Pulsar resources should be retained or
// removed from the Pulsar cluster after the CR is deleted.
type PulsarResourceLifeCyclePolicy string

const (
	// KeepAfterDeletion instructs the operator to retain the Pulsar resource
	// in the Pulsar cluster even after the corresponding CR is deleted from Kubernetes.
	// Use this option when:
	// - You want to preserve data or configurations in Pulsar for backup or future use.
	// - You're performing temporary maintenance on Kubernetes resources without affecting Pulsar.
	// - You need to migrate or recreate Kubernetes resources without losing Pulsar data.
	KeepAfterDeletion PulsarResourceLifeCyclePolicy = "KeepAfterDeletion"

	// CleanUpAfterDeletion instructs the operator to remove the Pulsar resource
	// from the Pulsar cluster when the corresponding CR is deleted from Kubernetes.
	// Use this option when:
	// - You want to ensure complete removal of resources to free up Pulsar cluster capacity.
	// - You're decommissioning services or environments and need to clean up all associated resources.
	// - You want to maintain strict synchronization between Kubernetes and Pulsar resources.
	CleanUpAfterDeletion PulsarResourceLifeCyclePolicy = "CleanUpAfterDeletion"
)

// PulsarAuthenticationOAuth2 represents the configuration for Pulsar OAuth2 authentication.
// This struct aligns with Pulsar's OAuth2 authentication mechanism as described in
// https://pulsar.apache.org/docs/3.3.x/security-oauth2/ and
// https://docs.streamnative.io/docs/access-cloud-clusters-oauth
type PulsarAuthenticationOAuth2 struct {
	// IssuerEndpoint is the URL of the OAuth2 authorization server.
	// This is typically the base URL of your identity provider's OAuth2 service.
	IssuerEndpoint string `json:"issuerEndpoint"`

	// ClientID is the OAuth2 client identifier issued to the client during the registration process.
	ClientID string `json:"clientID"`

	// Audience is the intended recipient of the token. In Pulsar's context, this is usually
	// the URL of your Pulsar cluster or a specific identifier for your Pulsar service.
	Audience string `json:"audience"`

	// Key is either the client secret or the path to a JSON credentials file.
	// For confidential clients, this would be the client secret.
	// For public clients using JWT authentication, this would be the path to the JSON credentials file.
	Key *ValueOrSecretRef `json:"key"`

	// Scope is an optional field to request specific permissions from the OAuth2 server.
	// If not specified, the default scope defined by the OAuth2 server will be used.
	Scope string `json:"scope,omitempty"`
}

// IsPulsarResourceReady returns true if resource satisfies with these condition
// 1. The instance is not deleted
// 2. Status ObservedGeneration is equal with meta.ObservedGeneration
// 3. StatusCondition Ready is true
func IsPulsarResourceReady(instance reconciler.Object) bool {
	objVal := reflect.ValueOf(instance).Elem()
	stVal := objVal.FieldByName("Status")

	ogVal := stVal.FieldByName("ObservedGeneration")
	observedGeneration := ogVal.Int()

	conditionsVal := stVal.FieldByName("Conditions")
	conditions := conditionsVal.Interface().([]metav1.Condition)

	return instance.GetDeletionTimestamp().IsZero() &&
		instance.GetGeneration() == observedGeneration &&
		meta.IsStatusConditionTrue(conditions, ConditionReady)
}

// PackageContentRef indicates the package content reference
type PackageContentRef struct {
	// +optional
	// persistentVolumeTemplate *corev1.PersistentVolumeClaim `json:"persistentVolumeTemplate,omitempty"`

	// +optional
	URL string `json:"url,omitempty"`
}

// Resources indicates the resources for the pulsar functions and connectors
type Resources struct {
	// +optional
	CPU string `json:"cpu,omitempty"`

	// +optional
	Disk int64 `json:"disk,omitempty"`

	// +optional
	RAM int64 `json:"ram,omitempty"`
}

// ProducerConfig represents the configuration for the producer of the pulsar functions and connectors
type ProducerConfig struct {
	// +optional
	MaxPendingMessages int `json:"maxPendingMessages,omitempty" yaml:"maxPendingMessages"`

	// +optional
	MaxPendingMessagesAcrossPartitions int `json:"maxPendingMessagesAcrossPartitions,omitempty" yaml:"maxPendingMessagesAcrossPartitions"`

	// +optional
	UseThreadLocalProducers bool `json:"useThreadLocalProducers,omitempty" yaml:"useThreadLocalProducers"`

	// +optional
	CryptoConfig *CryptoConfig `json:"cryptoConfig,omitempty" yaml:"cryptoConfig"`

	// +optional
	BatchBuilder string `json:"batchBuilder,omitempty" yaml:"batchBuilder"`

	// +optional
	CompressionType string `json:"compressionType,omitempty" yaml:"compressionType"`
}

// ConsumerConfig represents the configuration for the consumer of the pulsar functions and connectors
type ConsumerConfig struct {
	// +optional
	SchemaType string `json:"schemaType,omitempty" yaml:"schemaType"`

	// +optional
	SerdeClassName string `json:"serdeClassName,omitempty" yaml:"serdeClassName"`

	// +optional
	RegexPattern bool `json:"regexPattern,omitempty" yaml:"regexPattern"`

	// +optional
	ReceiverQueueSize int `json:"receiverQueueSize,omitempty" yaml:"receiverQueueSize"`

	// +optional
	SchemaProperties map[string]string `json:"schemaProperties,omitempty" yaml:"schemaProperties"`

	// +optional
	ConsumerProperties map[string]string `json:"consumerProperties,omitempty" yaml:"consumerProperties"`

	// +optional
	CryptoConfig *CryptoConfig `json:"cryptoConfig,omitempty" yaml:"cryptoConfig"`

	// +optional
	PoolMessages bool `json:"poolMessages,omitempty" yaml:"poolMessages"`
}

// CryptoConfig represents the configuration for the crypto of the pulsar functions and connectors
type CryptoConfig struct {

	// +optional
	CryptoKeyReaderClassName string `json:"cryptoKeyReaderClassName,omitempty" yaml:"cryptoKeyReaderClassName"`

	// +optional
	CryptoKeyReaderConfig map[string]string `json:"cryptoKeyReaderConfig,omitempty" yaml:"cryptoKeyReaderConfig"`

	// +optional
	EncryptionKeys []string `json:"encryptionKeys,omitempty" yaml:"encryptionKeys"`

	// +optional
	ProducerCryptoFailureAction string `json:"producerCryptoFailureAction,omitempty" yaml:"producerCryptoFailureAction"`

	// +optional
	ConsumerCryptoFailureAction string `json:"consumerCryptoFailureAction,omitempty" yaml:"consumerCryptoFailureAction"`
}

// FunctionSecretKeyRef indicates a secret name and key
type FunctionSecretKeyRef struct {
	Path string `json:"path"`
	Key  string `json:"key"`
}
