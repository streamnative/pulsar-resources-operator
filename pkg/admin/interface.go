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

package admin

import (
	"fmt"
	"os"
	"strings"

	"github.com/apache/pulsar-client-go/oauth2"
	"github.com/apache/pulsar-client-go/pulsaradmin/pkg/admin"
	"github.com/apache/pulsar-client-go/pulsaradmin/pkg/admin/auth"
	"github.com/apache/pulsar-client-go/pulsaradmin/pkg/admin/config"
	utils2 "github.com/apache/pulsar-client-go/pulsaradmin/pkg/utils"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	"github.com/streamnative/pulsar-resources-operator/pkg/utils"
)

// TenantParams indicates the parameters for creating a tenant
type TenantParams struct {
	AdminRoles      []string
	AllowedClusters []string
	Changed         bool
}

// NamespaceParams indicates the parameters for creating a namespace
type NamespaceParams struct {
	Bundles                               *int32
	MaxProducersPerTopic                  *int32
	MaxConsumersPerTopic                  *int32
	MaxConsumersPerSubscription           *int32
	MessageTTL                            *utils.Duration
	RetentionTime                         *utils.Duration
	RetentionSize                         *resource.Quantity
	BacklogQuotaLimitTime                 *utils.Duration
	BacklogQuotaLimitSize                 *resource.Quantity
	BacklogQuotaRetentionPolicy           *string
	BacklogQuotaType                      *string
	OffloadThresholdTime                  *utils.Duration
	OffloadThresholdSize                  *resource.Quantity
	ReplicationClusters                   []string
	Deduplication                         *bool
	BookieAffinityGroup                   *v1alpha1.BookieAffinityGroupData
	TopicAutoCreationConfig               *v1alpha1.TopicAutoCreationConfig
	SchemaCompatibilityStrategy           *v1alpha1.SchemaCompatibilityStrategy
	SchemaValidationEnforced              *bool
	DispatchRate                          *v1alpha1.DispatchRate
	SubscriptionDispatchRate              *v1alpha1.DispatchRate
	ReplicatorDispatchRate                *v1alpha1.DispatchRate
	PublishRate                           *v1alpha1.PublishRate
	SubscribeRate                         *v1alpha1.SubscribeRate
	PersistencePolicies                   *v1alpha1.PersistencePolicies
	CompactionThreshold                   *int64
	InactiveTopicPolicies                 *v1alpha1.InactiveTopicPolicies
	SubscriptionExpirationTime            *utils.Duration
	Properties                            map[string]string
	IsAllowAutoUpdateSchema               *bool
	ValidateProducerName                  *bool
	EncryptionRequired                    *bool
	SubscriptionAuthMode                  *utils2.SubscriptionAuthMode
	AntiAffinityGroup                     *string
	SchemaAutoUpdateCompatibilityStrategy *utils2.SchemaAutoUpdateCompatibilityStrategy
}

// TopicParams indicates the parameters for creating a topic
type TopicParams struct {
	Persistent                        *bool
	Partitions                        *int32
	MaxProducers                      *int32
	MaxConsumers                      *int32
	MessageTTL                        *utils.Duration
	MaxUnAckedMessagesPerConsumer     *int32
	MaxUnAckedMessagesPerSubscription *int32
	RetentionTime                     *utils.Duration
	RetentionSize                     *resource.Quantity
	BacklogQuotaLimitTime             *utils.Duration
	BacklogQuotaLimitSize             *resource.Quantity
	BacklogQuotaRetentionPolicy       *string
	ReplicationClusters               []string
	Deduplication                     *bool
	CompactionThreshold               *int64
	PersistencePolicies               *v1alpha1.PersistencePolicies
	DelayedDelivery                   *v1alpha1.DelayedDeliveryData
	DispatchRate                      *v1alpha1.DispatchRate
	PublishRate                       *v1alpha1.PublishRate
	InactiveTopicPolicies             *v1alpha1.InactiveTopicPolicies
	SubscribeRate                     *v1alpha1.SubscribeRate
	MaxMessageSize                    *int32
	MaxConsumersPerSubscription       *int32
	MaxSubscriptionsPerTopic          *int32
	SchemaValidationEnforced          *bool
	SubscriptionDispatchRate          *v1alpha1.DispatchRate
	ReplicatorDispatchRate            *v1alpha1.DispatchRate
	DeduplicationSnapshotInterval     *int32
	OffloadPolicies                   *v1alpha1.OffloadPolicies
	AutoSubscriptionCreation          *v1alpha1.AutoSubscriptionCreationOverride
	SchemaCompatibilityStrategy       *v1alpha1.SchemaCompatibilityStrategy
	Properties                        map[string]string
}

// ClusterParams indicate the parameters for creating a cluster
type ClusterParams struct {
	ServiceURL                     string
	ServiceSecureURL               string
	BrokerServiceURL               string
	BrokerServiceSecureURL         string
	AuthPlugin                     string
	AuthParameters                 string
	BrokerClientTrustCertsFilePath string
}

// SchemaParams indicates the parameters for uploading a schema
type SchemaParams struct {
	// Type determines how to interpret the schema data
	Type string `json:"type,omitempty"`
	// Schema is schema data
	Schema string `json:"schema,omitempty"`
	// Properties is a user defined properties as a string/string map
	Properties map[string]string `json:"properties,omitempty"`
}

// PulsarAdmin is the interface that defines the functions to call pulsar admin
type PulsarAdmin interface {
	// ApplyTenant creates or updates a tenant with parameters
	ApplyTenant(name string, params *TenantParams) error

	// DeleteTenant delete a specific tenant
	DeleteTenant(name string) error

	// ApplyNamespace creates a namespace with parameters
	ApplyNamespace(name string, params *NamespaceParams) error

	// DeleteNamespace delete a specific namespace
	DeleteNamespace(name string) error

	// GetNamespaceClusters get the assigned clusters of the namespace to the local default cluster
	GetNamespaceClusters(completeNSName string) ([]string, error)
	// SetNamespaceClusters resets the assigned clusters of the namespace to the local default cluster
	SetNamespaceClusters(name string, clusters []string) error

	// ApplyTopic creates a topic with parameters
	ApplyTopic(name string, params *TopicParams) (error, error)

	// DeleteTopic delete a specific topic
	DeleteTopic(name string) error

	// GetTopicClusters get the assigned clusters of the topic to the local default cluster
	GetTopicClusters(name string, persistent *bool) ([]string, error)
	// SetTopicClusters resets the assigned clusters of the topic to the local default cluster
	SetTopicClusters(name string, persistent *bool, clusters []string) error

	// GrantPermissions grants permissions to multiple role with multiple actions
	// on a namespace or topic, each role will be granted the same actions
	GrantPermissions(p Permissioner) error

	// RevokePermissions revoke permissions from roles on a namespace or topic.
	// it will revoke all actions which granted to a role on a namespace or topic
	RevokePermissions(p Permissioner) error

	// GetNamespacePermissions get permissions by namespace
	GetNamespacePermissions(namespace string) (map[string][]utils2.AuthAction, error)

	// GetTopicPermissions get permissions by topic
	GetTopicPermissions(topic string) (map[string][]utils2.AuthAction, error)

	// Close releases the connection with pulsar admin
	Close() error

	// GetSchema retrieves the latest schema of a topic
	GetSchema(topic string) (*v1alpha1.SchemaInfo, error)

	// UploadSchema creates or updates a schema for a given topic
	UploadSchema(topic string, params *SchemaParams) error

	// DeleteSchema deletes the schema associated with a given topic
	DeleteSchema(topic string) error

	// CreateCluster creates cluster info
	CreateCluster(name string, param *ClusterParams) error

	// UpdateCluster updates cluster info
	UpdateCluster(name string, param *ClusterParams) error

	// DeleteCluster delete cluster info
	DeleteCluster(name string) error

	// CheckClusterExist check whether the cluster is created or not
	CheckClusterExist(name string) (bool, error)

	// DeletePulsarPackage delete pulsar package
	DeletePulsarPackage(packageURL string) error

	// CheckPulsarPackageExist check whether the package is created or not
	CheckPulsarPackageExist(packageURL string) (bool, error)

	// ApplyPulsarPackage apply pulsar package
	ApplyPulsarPackage(packageURL, filePath, description, contact string, properties map[string]string, changed bool) error

	// DeletePulsarFunction delete pulsar function
	DeletePulsarFunction(tenant, namespace, name string) error

	// CheckPulsarFunctionExist check whether the function is created or not
	CheckPulsarFunctionExist(tenant, namespace, name string) (bool, error)

	// ApplyPulsarFunction apply pulsar function
	ApplyPulsarFunction(tenant, namespace, name, packageURL string, param *v1alpha1.PulsarFunctionSpec, changed bool) error

	// DeletePulsarSink delete pulsar sink
	DeletePulsarSink(tenant, namespace, name string) error

	// CheckPulsarSinkExist check whether the sink is created or not
	CheckPulsarSinkExist(tenant, namespace, name string) (bool, error)

	// ApplyPulsarSink apply pulsar sink
	ApplyPulsarSink(tenant, namespace, name, packageURL string, param *v1alpha1.PulsarSinkSpec, changed bool) error

	// DeletePulsarSource delete pulsar source
	DeletePulsarSource(tenant, namespace, name string) error

	// CheckPulsarSourceExist check whether the source is created or not
	CheckPulsarSourceExist(tenant, namespace, name string) (bool, error)

	// ApplyPulsarSource apply pulsar source
	ApplyPulsarSource(tenant, namespace, name, packageURL string, param *v1alpha1.PulsarSourceSpec, changed bool) error

	// GetTenantAllowedClusters get the allowed clusters of the tenant
	GetTenantAllowedClusters(name string) ([]string, error)

	// GetNSIsolationPolicy get the ns-isolation-policy
	GetNSIsolationPolicy(policyName, clusterName string) (*utils2.NamespaceIsolationData, error)

	// CreateNSIsolationPolicy create a ns-isolation-policy
	CreateNSIsolationPolicy(policyName, clusterName string, policyData utils2.NamespaceIsolationData) error

	// DeleteNSIsolationPolicy delete the ns-isolation-policy
	DeleteNSIsolationPolicy(policyName, clusterName string) error

	// GetPulsarPackageMetadata retrieves package information
	GetPulsarPackageMetadata(packageURL string) (*utils2.PackageMetadata, error)
}

// PulsarAdminCreator is the function type to create a PulsarAdmin with config
type PulsarAdminCreator func(config PulsarAdminConfig) (PulsarAdmin, error)

// PulsarAdminConfig indicates the configurations which are needed to initialize the pulsar admin
type PulsarAdminConfig struct {

	// WebServiceURL to connect to Pulsar.
	WebServiceURL string

	// Set the path to the trusted TLS certificate file
	TLSTrustCertsFilePath string
	// Configure whether the Pulsar client accept untrusted TLS certificate from broker (default: false)
	TLSAllowInsecureConnection bool

	TLSEnableHostnameVerification bool

	// Either Token or OAuth2 configuration must be provided
	// The Token used for authentication.
	Token string

	// OAuth2 related configuration used for authentication.
	IssuerEndpoint string
	ClientID       string
	Audience       string
	Key            string
	Scope          string

	// TLS Authentication related configuration
	ClientCertificatePath    string
	ClientCertificateKeyPath string

	PulsarAPIVersion *config.APIVersion
}

// NewPulsarAdmin initialize a pulsar admin client with configuration
func NewPulsarAdmin(conf PulsarAdminConfig) (PulsarAdmin, error) {
	var keyFile *os.File
	var keyFilePath string
	var err error
	var adminClient admin.Client

	config := &config.Config{
		WebServiceURL:                 conf.WebServiceURL,
		TLSAllowInsecureConnection:    conf.TLSAllowInsecureConnection,
		TLSEnableHostnameVerification: conf.TLSEnableHostnameVerification,
		TLSTrustCertsFilePath:         conf.TLSTrustCertsFilePath,
		// V2 admin endpoint contains operations for tenant, namespace and topic.
		PulsarAPIVersion: config.V2,
	}

	if conf.PulsarAPIVersion != nil {
		config.PulsarAPIVersion = *conf.PulsarAPIVersion
	}

	if conf.Key != "" {
		keyFile, err = os.CreateTemp("", "oauth2-key-")
		if err != nil {
			return nil, err
		}
		keyFilePath = keyFile.Name()
		_, err = keyFile.WriteString(conf.Key)
		if err != nil {
			return nil, err
		}

		config.IssuerEndpoint = conf.IssuerEndpoint
		config.ClientID = conf.ClientID
		config.Audience = conf.Audience
		config.KeyFile = keyFilePath
		config.Scope = conf.Scope

		oauthProvider, err := auth.NewAuthenticationOAuth2WithFlow(oauth2.Issuer{
			IssuerEndpoint: conf.IssuerEndpoint,
			ClientID:       conf.ClientID,
			Audience:       conf.Audience,
		}, oauth2.ClientCredentialsFlowOptions{
			KeyFile:          keyFilePath,
			AdditionalScopes: strings.Split(conf.Scope, " "),
		})

		if err != nil {
			return nil, err
		}
		adminClient, err = admin.NewPulsarClientWithAuthProvider(config, oauthProvider)
		if err != nil {
			return nil, err
		}
	} else if conf.Token != "" {
		config.Token = conf.Token

		adminClient, err = admin.New(config)
		if err != nil {
			return nil, err
		}
	} else if conf.ClientCertificatePath != "" {
		config.AuthPlugin = auth.TLSPluginName
		config.AuthParams = fmt.Sprintf("{\"tlsCertFile\": %q, \"tlsKeyFile\": %q}", conf.ClientCertificatePath, conf.ClientCertificateKeyPath)

		adminClient, err = admin.New(config)
		if err != nil {
			return nil, err
		}
	} else {
		adminClient, err = admin.New(config)
		if err != nil {
			return nil, err
		}
	}
	pulsarAdminClient := &PulsarAdminClient{
		adminClient,
		keyFile,
	}
	return pulsarAdminClient, nil
}

// NamespacePermission is the parameters to grant permission for a namespace
type NamespacePermission struct {
	ResourceName string
	Roles        []string
	Actions      []string
}

// TopicPermission is the parameters to grant permission for a topic
type TopicPermission struct {
	ResourceName string
	Roles        []string
	Actions      []string
}

// Permissioner implements the functions to grant and revoke permission for namespace and topic
type Permissioner interface {
	// Grant grants permission to role on a resource
	Grant(client admin.Client) error
	// Revoke revokes permission from role on a resource
	Revoke(client admin.Client) error
}
