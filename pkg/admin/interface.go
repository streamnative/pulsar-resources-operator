// Copyright (c) 2022 StreamNative, Inc.. All Rights Reserved.

package admin

import (
	"io/ioutil"
	"os"

	"github.com/apache/pulsar-client-go/oauth2"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/streamnative/pulsarctl/pkg/auth"
	"github.com/streamnative/pulsarctl/pkg/pulsar"
	pulsarctlcommon "github.com/streamnative/pulsarctl/pkg/pulsar/common"
)

// TenantParams indicates the parameters for creating a tenant
type TenantParams struct {
	AdminRoles      []string
	AllowedClusters []string
}

// NamespaceParams indicates the parameters for creating a namespace
type NamespaceParams struct {
	Bundles                     *int32
	MaxProducersPerTopic        *int32
	MaxConsumersPerTopic        *int32
	MaxConsumersPerSubscription *int32
	MessageTTL                  *metav1.Duration
	RetentionTime               *metav1.Duration
	RetentionSize               *resource.Quantity
	BacklogQuotaLimitTime       *metav1.Duration
	BacklogQuotaLimitSize       *resource.Quantity
	BacklogQuotaRetentionPolicy *string
	BacklogQuotaType            *string
}

// TopicParams indicates the parameters for creating a topic
type TopicParams struct {
	Persistent                        *bool
	Partitions                        *int32
	MaxProducers                      *int32
	MaxConsumers                      *int32
	MessageTTL                        *metav1.Duration
	MaxUnAckedMessagesPerConsumer     *int32
	MaxUnAckedMessagesPerSubscription *int32
	RetentionTime                     *metav1.Duration
	RetentionSize                     *resource.Quantity
	BacklogQuotaLimitTime             *metav1.Duration
	BacklogQuotaLimitSize             *resource.Quantity
	BacklogQuotaRetentionPolicy       *string
}

// PulsarAdmin is the interface that defines the functions to call pulsar admin
type PulsarAdmin interface {
	// ApplyTenant creates a tenant with parameters
	ApplyTenant(name string, params *TenantParams) error

	// DeleteTenant delete a specific tenant
	DeleteTenant(name string) error

	// ApplyNamespace creates a namespace with parameters
	ApplyNamespace(name string, params *NamespaceParams) error

	// DeleteNamespace delete a specific namespace
	DeleteNamespace(name string) error

	// ApplyTopic creates a topic with parameters
	ApplyTopic(name string, params *TopicParams) error

	// DeleteTopic delete a specific topic
	DeleteTopic(name string) error

	// GrantPermissions grants permissions to multiple role with multiple actions
	// on a namespace or topic, each role will be granted the same actions
	GrantPermissions(p Permissioner) error

	// RevokePermissions revoke permissions from roles on a namespace or topic.
	// it will revoke all actions which granted to a role on a namespace or topic
	RevokePermissions(p Permissioner) error

	// Close releases the connection with pulsar admin
	Close() error
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
}

// NewPulsarAdmin initialize a pulsar admin client with configuration
func NewPulsarAdmin(conf PulsarAdminConfig) (PulsarAdmin, error) {
	var keyFile *os.File
	var keyFilePath string
	var err error
	var adminClient pulsar.Client

	config := &pulsarctlcommon.Config{
		WebServiceURL:              conf.WebServiceURL,
		TLSAllowInsecureConnection: true,
		// V2 admin endpoint contains operations for tenant, namespace and topic.
		PulsarAPIVersion: pulsarctlcommon.V2,
	}

	if conf.Key != "" {
		keyFile, err = ioutil.TempFile("", "oauth2-key-")
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

		oauthProvider, err := auth.NewAuthenticationOAuth2WithDefaultFlow(oauth2.Issuer{
			IssuerEndpoint: conf.IssuerEndpoint,
			ClientID:       conf.ClientID,
			Audience:       conf.Audience,
		}, keyFilePath)
		if err != nil {
			return nil, err
		}
		adminClient = pulsar.NewWithAuthProvider(config, oauthProvider)
	} else {
		config.Token = conf.Token

		adminClient, err = pulsar.New(config)
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
	Grant(client pulsar.Client) error
	// Revoke revokes permission from role on a resource
	Revoke(client pulsar.Client) error
}
