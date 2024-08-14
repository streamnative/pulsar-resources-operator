// Copyright 2022 StreamNative
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

// ServiceAccountAuth references a service account to authenticate with pulsar
type ServiceAccountAuth struct {
	Name     string `json:"name,omitempty"`
	Audience string `json:"audience,omitempty"`
}

// PulsarAuthentication use the token or OAuth2 for pulsar authentication
type PulsarAuthentication struct {
	// +optional
	Token *ValueOrSecretRef `json:"token,omitempty"`

	// +optional
	ServiceAccount *ServiceAccountAuth `json:"serviceAccount,omitempty"`

	// +optional
	OAuth2 *PulsarAuthenticationOAuth2 `json:"oauth2,omitempty"`
}

// PulsarResourceLifeCyclePolicy indicates whether it will keep or delete the resource
// in pulsar cluster after resource is deleted by controller
// KeepAfterDeletion or CleanUpAfterDeletion
type PulsarResourceLifeCyclePolicy string

const (
	// KeepAfterDeletion keeps the resource in pulsar cluster when cr is deleted
	KeepAfterDeletion PulsarResourceLifeCyclePolicy = "KeepAfterDeletion"
	// CleanUpAfterDeletion deletes the resource in pulsar cluster when cr is deleted
	CleanUpAfterDeletion PulsarResourceLifeCyclePolicy = "CleanUpAfterDeletion"
)

// PulsarAuthenticationOAuth2 indicates the parameters which are need by pulsar OAuth2
type PulsarAuthenticationOAuth2 struct {
	IssuerEndpoint string           `json:"issuerEndpoint"`
	ClientID       string           `json:"clientID"`
	Audience       string           `json:"audience"`
	Key            ValueOrSecretRef `json:"key"`
	Scope          string           `json:"scope,omitempty"`
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
