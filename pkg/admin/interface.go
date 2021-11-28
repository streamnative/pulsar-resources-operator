// Copyright (c) 2020 StreamNative, Inc.. All Rights Reserved.

package admin

import (
	"io/ioutil"
	"os"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/apache/pulsar-client-go/oauth2"
	"github.com/streamnative/pulsarctl/pkg/auth"
	"github.com/streamnative/pulsarctl/pkg/pulsar"
	pulsarctlcommon "github.com/streamnative/pulsarctl/pkg/pulsar/common"
)

type TenantParams struct {
	AdminRoles      []string
	AllowedClusters []string
}

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
}

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

type PulsarAdmin interface {
	ApplyTenant(name string, params *TenantParams) error

	DeleteTenant(name string) error

	ApplyNamespace(name string, params *NamespaceParams) error

	DeleteNamespace(name string) error

	ApplyTopic(name string, params *TopicParams) error

	DeleteTopic(name string) error

	// GrantPermissions grants permissions to multiple role with multiple actions
	// on a namespace or topic, each role will be granted the same actions
	GrantPermissions(p Permissioner) error

	// RevokePermissions revoke permissions from roles on a namespace or topic.
	// it will revoke all actions which granted to a role on a namespace or topic
	RevokePermissions(p Permissioner) error

	Close() error
}

type PulsarAdminCreator func(config PulsarAdminConfig) (PulsarAdmin, error)

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

type PermissionParams struct {
	ResourceName string
	ResourceType string
	Roles        []string
	Actions      []string
}

type NamespacePermission struct {
	ResourceName string
	Roles        []string
	Actions      []string
}
type TopicPermission struct {
	ResourceName string
	Roles        []string
	Actions      []string
}

type Permissioner interface {
	// Grant grants permission to role on a resource
	Grant(client pulsar.Client) error
	// Revoke revokes permission from role on a resource
	Revoke(client pulsar.Client) error
}
