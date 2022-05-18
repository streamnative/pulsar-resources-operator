// Copyright 2022 Stream Native
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

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/pointer"

	"github.com/streamnative/pulsarctl/pkg/pulsar"
	"github.com/streamnative/pulsarctl/pkg/pulsar/common"
	pulsarutils "github.com/streamnative/pulsarctl/pkg/pulsar/utils"
)

// PulsarAdminClient define the client to call pulsar
type PulsarAdminClient struct {
	adminClient pulsar.Client
	keyFile     *os.File
}

const (
	// TopicDomainSeparator is the separator to separate the topic name
	TopicDomainSeparator = "://"
	// TopicDomainPersistent is the prefix for persistent topic
	TopicDomainPersistent = "persistent"
	// TopicDomainNonPersistent is the prefix for non persistent topic
	TopicDomainNonPersistent = "non-persistent"
)

// ApplyTenant creates or updates a tenant, if AllowdClusters is not provided, it will list all clusters in pular
// When updates a tenant,  If AdminRoles is empty, the current set of roles won't be modified
func (p *PulsarAdminClient) ApplyTenant(name string, params *TenantParams) error {
	param := pulsarutils.TenantData{
		Name:       name,
		AdminRoles: []string{},
	}
	if params != nil {
		if len(params.AdminRoles) != 0 {
			param.AdminRoles = params.AdminRoles
		}
		if len(params.AllowedClusters) != 0 {
			param.AllowedClusters = params.AllowedClusters
		} else {
			clusters, err := p.adminClient.Clusters().List()
			if err != nil {
				return err
			}
			param.AllowedClusters = clusters
		}
	}
	if params.Changed {
		err := p.adminClient.Tenants().Update(param)
		if err != nil {
			return err
		}
	} else {
		err := p.adminClient.Tenants().Create(param)
		if err != nil && !IsAlreadyExist(err) {
			return err
		}
	}
	return nil
}

// ApplyNamespace creates a namespace with policies
func (p *PulsarAdminClient) ApplyNamespace(name string, params *NamespaceParams) error {
	if params.Bundles == nil {
		params.Bundles = pointer.Int32Ptr(4)
	}
	err := p.adminClient.Namespaces().CreateNsWithPolices(name, pulsarutils.Policies{
		Bundles: &pulsarutils.BundlesData{
			NumBundles: int(*params.Bundles),
		},
		SchemaCompatibilityStrategy: pulsarutils.AlwaysCompatible,
		SubscriptionAuthMode:        pulsarutils.None,
		ReplicationClusters:         []string{},
	})
	if err != nil && !IsAlreadyExist(err) {
		return err
	}

	err = p.applyTenantPolicies(name, params)
	if err != nil {
		return err
	}

	return nil
}

// ApplyTopic creates a topic with policies
func (p *PulsarAdminClient) ApplyTopic(name string, params *TopicParams) error {
	completeTopicName := makeCompleteTopicName(name, params.Persistent)
	topicName, err := pulsarutils.GetTopicName(completeTopicName)
	if err != nil {
		return err
	}
	err = p.adminClient.Topics().Create(*topicName, int(*params.Partitions))
	if err != nil && !IsAlreadyExist(err) {
		return err
	}

	err = p.applyTopicPolicies(topicName, params)
	if err != nil {
		return err
	}

	return nil
}

// DeleteTenant deletes a specific tenant
func (p *PulsarAdminClient) DeleteTenant(name string) error {
	if err := p.adminClient.Tenants().Delete(name); err != nil {
		return err
	}
	return nil
}

// DeleteNamespace deletes a specific namespace
func (p *PulsarAdminClient) DeleteNamespace(name string) error {
	if err := p.adminClient.Namespaces().DeleteNamespace(name); err != nil {
		return err
	}
	return nil
}

// DeleteTopic deletes a specific topic
func (p *PulsarAdminClient) DeleteTopic(name string) error {
	topic, err := pulsarutils.GetTopicName(name)
	if err != nil {
		return err
	}
	topicMeta, err := p.adminClient.Topics().GetMetadata(*topic)
	if err != nil {
		return err
	}
	nonPartitioned := true
	if topicMeta.Partitions > 0 {
		nonPartitioned = false
	}
	if err := p.adminClient.Topics().Delete(*topic, true, nonPartitioned); err != nil {
		return err
	}
	return nil
}

// Close do nothing for now
func (p *PulsarAdminClient) Close() error {
	return nil
}

func (p *PulsarAdminClient) applyTopicPolicies(topicName *pulsarutils.TopicName, params *TopicParams) error {
	var err error
	if params.MessageTTL != nil {
		err = p.adminClient.Topics().SetMessageTTL(*topicName, int(params.MessageTTL.Seconds()))
		if err != nil {
			return err
		}
	}

	if params.MaxProducers != nil {
		err = p.adminClient.Topics().SetMaxProducers(*topicName, int(*params.MaxProducers))
		if err != nil {
			return err
		}
	}

	if params.MaxConsumers != nil {
		err = p.adminClient.Topics().SetMaxConsumers(*topicName, int(*params.MaxConsumers))
		if err != nil {
			return err
		}
	}

	if params.MaxUnAckedMessagesPerConsumer != nil {
		err = p.adminClient.Topics().SetMaxUnackMessagesPerConsumer(*topicName, int(*params.MaxUnAckedMessagesPerConsumer))
		if err != nil {
			return err
		}
	}

	if params.MaxUnAckedMessagesPerSubscription != nil {
		err = p.adminClient.Topics().SetMaxUnackMessagesPerSubscription(*topicName,
			int(*params.MaxUnAckedMessagesPerSubscription))
		if err != nil {
			return err
		}
	}

	if params.RetentionTime != nil || params.RetentionSize != nil {
		retentionTime := -1
		retentionSize := -1
		if params.RetentionTime != nil {
			retentionTime = int(params.RetentionTime.Minutes())
		}
		if params.RetentionSize != nil {
			retentionSize = int(params.RetentionSize.ScaledValue(resource.Mega))
		}
		retentionPolicy := pulsarutils.NewRetentionPolicies(retentionTime, retentionSize)
		err = p.adminClient.Topics().SetRetention(*topicName, retentionPolicy)
		if err != nil {
			return err
		}
	}

	if (params.BacklogQuotaLimitTime != nil || params.BacklogQuotaLimitSize != nil) &&
		params.BacklogQuotaRetentionPolicy != nil {
		backlogTime := int64(-1)
		backlogSize := int64(-1)
		var backlogQuotaType pulsarutils.BacklogQuotaType
		if params.BacklogQuotaLimitTime != nil {
			backlogTime = int64(params.BacklogQuotaLimitTime.Seconds())
			backlogQuotaType = pulsarutils.MessageAge
		}
		if params.BacklogQuotaLimitSize != nil {
			backlogSize = params.BacklogQuotaLimitSize.Value()
			backlogQuotaType = pulsarutils.DestinationStorage
		}
		backlogQuotaPolicy := pulsarutils.BacklogQuota{
			LimitTime: backlogTime,
			LimitSize: backlogSize,
			Policy:    pulsarutils.RetentionPolicy(*params.BacklogQuotaRetentionPolicy),
		}
		err = p.adminClient.Topics().SetBacklogQuota(*topicName, backlogQuotaPolicy, backlogQuotaType)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *PulsarAdminClient) applyTenantPolicies(completeNSName string, params *NamespaceParams) error {
	naName, err := pulsarutils.GetNamespaceName(completeNSName)
	if err != nil {
		return err
	}

	if params.MessageTTL != nil {
		err = p.adminClient.Namespaces().SetNamespaceMessageTTL(completeNSName, int(params.MessageTTL.Seconds()))
		if err != nil {
			return err
		}
	}

	if params.MaxProducersPerTopic != nil {
		err = p.adminClient.Namespaces().SetMaxProducersPerTopic(*naName, int(*params.MaxProducersPerTopic))
		if err != nil {
			return err
		}
	}

	if params.MaxConsumersPerTopic != nil {
		err = p.adminClient.Namespaces().SetMaxConsumersPerTopic(*naName, int(*params.MaxConsumersPerTopic))
		if err != nil {
			return err
		}
	}

	if params.MaxConsumersPerSubscription != nil {
		err = p.adminClient.Namespaces().SetMaxConsumersPerSubscription(*naName, int(*params.MaxConsumersPerSubscription))
		if err != nil {
			return err
		}
	}

	if params.RetentionTime != nil || params.RetentionSize != nil {
		retentionTime := -1
		retentionSize := -1
		if params.RetentionTime != nil {
			retentionTime = int(params.RetentionTime.Minutes())
		}
		if params.RetentionSize != nil {
			retentionSize = int(params.RetentionSize.ScaledValue(resource.Mega))
		}
		retentionPolicy := pulsarutils.NewRetentionPolicies(retentionTime, retentionSize)
		err = p.adminClient.Namespaces().SetRetention(completeNSName, retentionPolicy)
		if err != nil {
			return err
		}
	}

	if (params.BacklogQuotaLimitTime != nil || params.BacklogQuotaLimitSize != nil) &&
		params.BacklogQuotaRetentionPolicy != nil {
		backlogTime := int64(-1)
		backlogSize := int64(-1)
		if params.BacklogQuotaLimitTime != nil {
			backlogTime = int64(params.BacklogQuotaLimitTime.Seconds())
		}
		if params.BacklogQuotaLimitSize != nil {
			backlogSize = params.BacklogQuotaLimitSize.Value()
		}
		backlogQuotaPolicy := pulsarutils.BacklogQuota{
			LimitTime: backlogTime,
			LimitSize: backlogSize,
			Policy:    pulsarutils.RetentionPolicy(*params.BacklogQuotaRetentionPolicy),
		}

		var backlogQuotaType pulsarutils.BacklogQuotaType
		if params.BacklogQuotaType != nil {
			backlogQuotaType, err = pulsarutils.ParseBacklogQuotaType(*params.BacklogQuotaType)
			if err != nil {
				return err
			}
		}
		err = p.adminClient.Namespaces().SetBacklogQuota(completeNSName, backlogQuotaPolicy, backlogQuotaType)
		if err != nil {
			return err
		}
	}

	return nil
}

func makeCompleteTopicName(topicName string, persistent *bool) string {
	if strings.Contains(topicName, TopicDomainSeparator) {
		return topicName
	}
	if *persistent {
		return fmt.Sprintf("%s://%s", TopicDomainPersistent, topicName)
	}
	return fmt.Sprintf("%s://%s", TopicDomainNonPersistent, topicName)
}

// GrantPermissions grants permissions to multiple role with multiple actions
// on a namespace or topic, each role will be granted the same actions
func (p *PulsarAdminClient) GrantPermissions(permission Permissioner) error {
	if err := permission.Grant(p.adminClient); err != nil {
		return err
	}

	return nil
}

// RevokePermissions revoke permissions from roles on a namespace or topic.
// it will revoke all actions which granted to a role on a namespace or topic
func (p *PulsarAdminClient) RevokePermissions(permission Permissioner) error {
	if err := permission.Revoke(p.adminClient); err != nil {
		return err
	}

	return nil
}

// convertActions converts actions type from string to common.AuthAction
func convertActions(actions []string) ([]common.AuthAction, error) {
	r := make([]common.AuthAction, 0)
	for _, v := range actions {
		a, err := common.ParseAuthAction(v)
		if err != nil {
			return nil, err
		}
		r = append(r, a)
	}

	return r, nil
}

// Grant implements the grant method for namespace permission
func (n *NamespacePermission) Grant(client pulsar.Client) error {
	actions, err := convertActions(n.Actions)
	if err != nil {
		return err
	}

	nsName, err := pulsarutils.GetNamespaceName(n.ResourceName)
	if err != nil {
		return err
	}

	adminNs := client.Namespaces()
	for _, role := range n.Roles {
		err := adminNs.GrantNamespacePermission(*nsName, role, actions)
		if err != nil {
			return err
		}
	}

	return nil
}

// Revoke implements the revoke method for namespace permission
func (n *NamespacePermission) Revoke(client pulsar.Client) error {
	nsName, err := pulsarutils.GetNamespaceName(n.ResourceName)
	if err != nil {
		return err
	}

	adminNs := client.Namespaces()
	for _, role := range n.Roles {
		err := adminNs.RevokeNamespacePermission(*nsName, role)
		if err != nil {
			return err
		}
	}

	return nil
}

// Grant implements the grant method for topic permission
func (t *TopicPermission) Grant(client pulsar.Client) error {
	actions, err := convertActions(t.Actions)
	if err != nil {
		return err
	}

	topicName, err := pulsarutils.GetTopicName(t.ResourceName)
	if err != nil {
		return err
	}
	adminTopic := client.Topics()

	for _, role := range t.Roles {
		err := adminTopic.GrantPermission(*topicName, role, actions)
		if err != nil {
			return err
		}
	}
	return nil
}

// Revoke implements the revoke method for topic permission
func (t *TopicPermission) Revoke(client pulsar.Client) error {
	topicName, err := pulsarutils.GetTopicName(t.ResourceName)
	if err != nil {
		return err
	}
	adminTopic := client.Topics()

	for _, role := range t.Roles {
		err := adminTopic.RevokePermission(*topicName, role)
		if err != nil && !IsPermissionNotFound(err) {
			return err
		}
	}
	return nil
}

// IsPermissionNotFound returns true if the permission is not set
func IsPermissionNotFound(err error) bool {
	return strings.Contains(err.Error(), "Permissions are not set at the topic level")
}
