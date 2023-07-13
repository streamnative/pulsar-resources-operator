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

package admin

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/pointer"

	"github.com/streamnative/pulsar-admin-go/pkg/admin"
	"github.com/streamnative/pulsar-admin-go/pkg/utils"
	"github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
)

// PulsarAdminClient define the client to call pulsar
type PulsarAdminClient struct {
	adminClient admin.Client
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
	param := utils.TenantData{
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
	err := p.adminClient.Namespaces().CreateNsWithPolices(name, utils.Policies{
		Bundles: &utils.BundlesData{
			NumBundles: int(*params.Bundles),
		},
		SchemaCompatibilityStrategy: utils.AlwaysCompatible,
		SubscriptionAuthMode:        utils.None,
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

// GetNamespaceClusters get the assigned clusters of the namespace to the local default cluster
func (p *PulsarAdminClient) GetNamespaceClusters(completeNSName string) ([]string, error) {
	clusters, err := p.adminClient.Namespaces().GetNamespaceReplicationClusters(completeNSName)
	if err != nil {
		return []string{}, err
	}
	return clusters, nil
}

// SetNamespaceClusters resets the assigned clusters of the namespace to the local default cluster
func (p *PulsarAdminClient) SetNamespaceClusters(completeNSName string, clusters []string) error {
	err := p.adminClient.Namespaces().SetNamespaceReplicationClusters(completeNSName, clusters)
	if err != nil {
		return err
	}
	return nil
}

// ApplyTopic creates a topic with policies
func (p *PulsarAdminClient) ApplyTopic(name string, params *TopicParams) error {
	completeTopicName := makeCompleteTopicName(name, params.Persistent)
	topicName, err := utils.GetTopicName(completeTopicName)
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
	topic, err := utils.GetTopicName(name)
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

func (p *PulsarAdminClient) applyTopicPolicies(topicName *utils.TopicName, params *TopicParams) error {
	var err error
	if params.MessageTTL != nil {
		ttl, err := params.MessageTTL.Parse()
		if err != nil {
			return err
		}
		err = p.adminClient.Topics().SetMessageTTL(*topicName, int(ttl.Seconds()))
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
			t, err := params.RetentionTime.Parse()
			if err != nil {
				return err
			}
			retentionTime = int(t.Minutes())
		}
		if params.RetentionSize != nil {
			retentionSize = int(params.RetentionSize.ScaledValue(resource.Mega))
		}
		retentionPolicy := utils.NewRetentionPolicies(retentionTime, retentionSize)
		err = p.adminClient.Topics().SetRetention(*topicName, retentionPolicy)
		if err != nil {
			return err
		}
	}

	if (params.BacklogQuotaLimitTime != nil || params.BacklogQuotaLimitSize != nil) &&
		params.BacklogQuotaRetentionPolicy != nil {
		backlogTime := int64(-1)
		backlogSize := int64(-1)
		var backlogQuotaType utils.BacklogQuotaType
		if params.BacklogQuotaLimitTime != nil {
			t, err := params.BacklogQuotaLimitTime.Parse()
			if err != nil {
				return err
			}
			backlogTime = int64(t.Seconds())
			backlogQuotaType = utils.MessageAge
		}
		if params.BacklogQuotaLimitSize != nil {
			backlogSize = params.BacklogQuotaLimitSize.Value()
			backlogQuotaType = utils.DestinationStorage
		}
		backlogQuotaPolicy := utils.BacklogQuota{
			LimitTime: backlogTime,
			LimitSize: backlogSize,
			Policy:    utils.RetentionPolicy(*params.BacklogQuotaRetentionPolicy),
		}
		err = p.adminClient.Topics().SetBacklogQuota(*topicName, backlogQuotaPolicy, backlogQuotaType)
		if err != nil {
			return err
		}
	}
	if len(params.ReplicationClusters) != 0 {
		err = p.adminClient.Topics().SetReplicationClusters(*topicName, params.ReplicationClusters)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetTopicClusters get the assigned clusters of the topic to the local default cluster
func (p *PulsarAdminClient) GetTopicClusters(name string, persistent *bool) ([]string, error) {
	completeTopicName := makeCompleteTopicName(name, persistent)
	topicName, err := utils.GetTopicName(completeTopicName)
	if err != nil {
		return []string{}, err
	}
	clusters, err := p.adminClient.Topics().GetReplicationClusters(*topicName)
	if err != nil {
		return []string{}, err
	}
	return clusters, nil
}

// SetTopicClusters resets the assigned clusters of the topic to the local default cluster
func (p *PulsarAdminClient) SetTopicClusters(name string, persistent *bool, clusters []string) error {
	completeTopicName := makeCompleteTopicName(name, persistent)
	topicName, err := utils.GetTopicName(completeTopicName)
	if err != nil {
		return err
	}
	err = p.adminClient.Topics().SetReplicationClusters(*topicName, clusters)
	if err != nil {
		return err
	}
	return nil
}

func (p *PulsarAdminClient) applyTenantPolicies(completeNSName string, params *NamespaceParams) error {
	naName, err := utils.GetNamespaceName(completeNSName)
	if err != nil {
		return err
	}

	if params.MessageTTL != nil {
		ttl, err := params.MessageTTL.Parse()
		if err != nil {
			return err
		}
		err = p.adminClient.Namespaces().SetNamespaceMessageTTL(completeNSName, int(ttl.Seconds()))
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
			t, err := params.RetentionTime.Parse()
			if err != nil {
				return err
			}
			retentionTime = int(t.Minutes())
		}
		if params.RetentionSize != nil {
			retentionSize = int(params.RetentionSize.ScaledValue(resource.Mega))
		}
		retentionPolicy := utils.NewRetentionPolicies(retentionTime, retentionSize)
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
			t, err := params.BacklogQuotaLimitTime.Parse()
			if err != nil {
				return err
			}
			backlogTime = int64(t.Seconds())
		}
		if params.BacklogQuotaLimitSize != nil {
			backlogSize = params.BacklogQuotaLimitSize.Value()
		}
		backlogQuotaPolicy := utils.BacklogQuota{
			LimitTime: backlogTime,
			LimitSize: backlogSize,
			Policy:    utils.RetentionPolicy(*params.BacklogQuotaRetentionPolicy),
		}

		var backlogQuotaType utils.BacklogQuotaType
		if params.BacklogQuotaType != nil {
			backlogQuotaType, err = utils.ParseBacklogQuotaType(*params.BacklogQuotaType)
			if err != nil {
				return err
			}
		}
		err = p.adminClient.Namespaces().SetBacklogQuota(completeNSName, backlogQuotaPolicy, backlogQuotaType)
		if err != nil {
			return err
		}
	}

	if len(params.ReplicationClusters) != 0 {
		err = p.adminClient.Namespaces().SetNamespaceReplicationClusters(completeNSName, params.ReplicationClusters)
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
func convertActions(actions []string) ([]utils.AuthAction, error) {
	r := make([]utils.AuthAction, 0)
	for _, v := range actions {
		a, err := utils.ParseAuthAction(v)
		if err != nil {
			return nil, err
		}
		r = append(r, a)
	}

	return r, nil
}

// Grant implements the grant method for namespace permission
func (n *NamespacePermission) Grant(client admin.Client) error {
	actions, err := convertActions(n.Actions)
	if err != nil {
		return err
	}

	nsName, err := utils.GetNamespaceName(n.ResourceName)
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
func (n *NamespacePermission) Revoke(client admin.Client) error {
	nsName, err := utils.GetNamespaceName(n.ResourceName)
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
func (t *TopicPermission) Grant(client admin.Client) error {
	actions, err := convertActions(t.Actions)
	if err != nil {
		return err
	}

	topicName, err := utils.GetTopicName(t.ResourceName)
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
func (t *TopicPermission) Revoke(client admin.Client) error {
	topicName, err := utils.GetTopicName(t.ResourceName)
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

// GetSchema get schema info for a given topic
func (p *PulsarAdminClient) GetSchema(topic string) (*v1alpha1.SchemaInfo, error) {
	info, err := p.adminClient.Schemas().GetSchemaInfo(topic)
	if err != nil {
		return nil, err
	}
	rsp := &v1alpha1.SchemaInfo{
		Type:       info.Type,
		Schema:     string(info.Schema),
		Properties: info.Properties,
	}
	return rsp, nil
}

// UploadSchema creates or updates a schema for a given topic
func (p *PulsarAdminClient) UploadSchema(topic string, params *SchemaParams) error {
	var payload utils.PostSchemaPayload
	payload.SchemaType = params.Type
	payload.Schema = params.Schema
	payload.Properties = params.Properties

	err := p.adminClient.Schemas().CreateSchemaByPayload(topic, payload)
	if err != nil {
		return err
	}
	return nil
}

// DeleteSchema deletes the schema associated with a given topic
func (p *PulsarAdminClient) DeleteSchema(topic string) error {
	return p.adminClient.Schemas().DeleteSchema(topic)
}

// CreateCluster creates pulsar cluster
func (p *PulsarAdminClient) CreateCluster(name string, param *ClusterParams) error {
	clusterData := utils.ClusterData{
		Name:                     name,
		AuthenticationPlugin:     param.AuthPlugin,
		AuthenticationParameters: param.AuthParameters,
	}

	// Enable tls
	if param.ServiceSecureURL != "" && param.BrokerServiceSecureURL != "" {
		clusterData.ServiceURLTls = param.ServiceSecureURL
		clusterData.BrokerServiceURLTls = param.BrokerServiceSecureURL
		clusterData.BrokerClientTrustCertsFilePath = param.BrokerClientTrustCertsFilePath
		clusterData.BrokerClientTLSEnabled = true
	} else if param.ServiceURL != "" && param.BrokerServiceURL != "" {
		clusterData.ServiceURL = param.ServiceURL
		clusterData.BrokerServiceURL = param.BrokerServiceURL
	} else {
		return errors.New("BrokerServiceURL and ServiceURL shouldn't be empty")
	}

	err := p.adminClient.Clusters().Create(clusterData)
	if err != nil && !IsAlreadyExist(err) {
		return err
	}
	return nil
}

// UpdateCluster update pulsar cluster info
func (p *PulsarAdminClient) UpdateCluster(name string, param *ClusterParams) error {
	clusterData := utils.ClusterData{
		Name:                     name,
		AuthenticationPlugin:     param.AuthPlugin,
		AuthenticationParameters: param.AuthParameters,
	}

	if param.ServiceSecureURL != "" && param.BrokerServiceSecureURL != "" {
		clusterData.ServiceURLTls = param.ServiceSecureURL
		clusterData.BrokerServiceURLTls = param.BrokerServiceSecureURL
	} else if param.ServiceURL != "" && param.BrokerServiceURL != "" {
		clusterData.ServiceURL = param.ServiceURL
		clusterData.BrokerServiceURL = param.BrokerServiceURL
	} else {
		return errors.New("BrokerServiceURL and ServiceURL shouldn't be empty")
	}

	err := p.adminClient.Clusters().Update(clusterData)
	if err != nil {
		return err
	}
	return nil
}

// DeleteCluster deletes a pulsar cluster
func (p *PulsarAdminClient) DeleteCluster(name string) error {
	return p.adminClient.Clusters().Delete(name)
}

// CheckClusterExist checks whether the cluster exists
func (p *PulsarAdminClient) CheckClusterExist(name string) (bool, error) {
	_, err := p.adminClient.Clusters().Get(name)

	if err != nil {
		if IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
