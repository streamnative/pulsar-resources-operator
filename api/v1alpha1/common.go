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
	"reflect"

	corev1 "k8s.io/api/core/v1"

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

	// +optional
	TLS *PulsarAuthenticationTLS `json:"tls,omitempty"`
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

// PulsarAuthenticationTLS indicates the parameters which are need by pulsar TLS Authentication
type PulsarAuthenticationTLS struct {
	ClientCertificatePath    string `json:"clientCertificatePath"`
	ClientCertificateKeyPath string `json:"clientCertificateKeyPath"`
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

// PodTemplate defines the pod template configuration
type PodTemplate struct {
	// Standard object's metadata.
	// +optional
	ObjectMeta ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the pod.
	// +optional
	Spec PodTemplateSpec `json:"spec,omitempty"`
}

// ObjectMeta is metadata that all persisted resources must have
type ObjectMeta struct {
	// Name of the resource
	// +optional
	Name string `json:"name,omitempty"`

	// Namespace of the resource
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Labels of the resource
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations of the resource
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

// PodTemplateSpec describes the data a pod should have when created from a template
type PodTemplateSpec struct {
	// List of volumes that can be mounted by containers belonging to the pod.
	// +optional
	Volumes []Volume `json:"volumes,omitempty"`

	// List of initialization containers belonging to the pod.
	// +optional
	InitContainers []Container `json:"initContainers,omitempty"`

	// List of containers belonging to the pod.
	// +optional
	Containers []Container `json:"containers,omitempty"`

	// ServiceAccountName is the name of the ServiceAccount to use to run this pod.
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// NodeSelector is a selector which must be true for the pod to fit on a node.
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// SecurityContext holds pod-level security attributes and common container settings.
	// +optional
	SecurityContext *corev1.PodSecurityContext `json:"securityContext,omitempty"`

	// ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.
	// +optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// If specified, the pod's scheduling constraints
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// If specified, the pod's tolerations.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

// Container defines a single application container
type Container struct {
	// Name of the container specified as a DNS_LABEL.
	// +optional
	Name string `json:"name,omitempty"`

	// Docker image name.
	// +optional
	Image string `json:"image,omitempty"`

	// Entrypoint array. Not executed within a shell.
	// +optional
	Command []string `json:"command,omitempty"`

	// Arguments to the entrypoint.
	// +optional
	Args []string `json:"args,omitempty"`

	// Container's working directory.
	// +optional
	WorkingDir string `json:"workingDir,omitempty"`

	// List of environment variables to set in the container.
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`

	// List of sources to populate environment variables in the container.
	// +optional
	EnvFrom []corev1.EnvFromSource `json:"envFrom,omitempty"`

	// Compute Resources required by this container.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Pod volumes to mount into the container's filesystem.
	// +optional
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`

	// Image pull policy.
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// Security context at container level
	// +optional
	SecurityContext *corev1.SecurityContext `json:"securityContext,omitempty"`
}

// Volume represents a named volume in a pod
type Volume struct {
	// Volume's name.
	// +required
	Name string `json:"name"`

	// VolumeSource represents the location and type of the mounted volume.
	// +required
	VolumeSource `json:",inline"`
}

// VolumeSource represents the source location of a volume to mount.
type VolumeSource struct {
	// ConfigMap represents a configMap that should populate this volume
	// +optional
	ConfigMap *corev1.ConfigMapVolumeSource `json:"configMap,omitempty"`

	// Secret represents a secret that should populate this volume.
	// +optional
	Secret *corev1.SecretVolumeSource `json:"secret,omitempty"`
}
