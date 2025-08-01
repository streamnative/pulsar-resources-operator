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
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/apache/pulsar-client-go/pulsaradmin/pkg/admin"
	"github.com/apache/pulsar-client-go/pulsaradmin/pkg/utils"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"

	"github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	rutils "github.com/streamnative/pulsar-resources-operator/pkg/utils"
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

// Type conversion functions for external library types

// convertOffloadPolicies converts our local OffloadPolicies to the external library type
func convertOffloadPolicies(local *v1alpha1.OffloadPolicies) *utils.OffloadPolicies {
	if local == nil {
		return nil
	}
	return &utils.OffloadPolicies{
		ManagedLedgerOffloadDriver:                        local.ManagedLedgerOffloadDriver,
		ManagedLedgerOffloadMaxThreads:                    local.ManagedLedgerOffloadMaxThreads,
		ManagedLedgerOffloadThresholdInBytes:              local.ManagedLedgerOffloadThresholdInBytes,
		ManagedLedgerOffloadDeletionLagInMillis:           local.ManagedLedgerOffloadDeletionLagInMillis,
		ManagedLedgerOffloadAutoTriggerSizeThresholdBytes: local.ManagedLedgerOffloadAutoTriggerSizeThresholdBytes,
		S3ManagedLedgerOffloadBucket:                      local.S3ManagedLedgerOffloadBucket,
		S3ManagedLedgerOffloadRegion:                      local.S3ManagedLedgerOffloadRegion,
		S3ManagedLedgerOffloadServiceEndpoint:             local.S3ManagedLedgerOffloadServiceEndpoint,
		S3ManagedLedgerOffloadCredentialID:                local.S3ManagedLedgerOffloadCredentialID,
		S3ManagedLedgerOffloadCredentialSecret:            local.S3ManagedLedgerOffloadCredentialSecret,
		S3ManagedLedgerOffloadRole:                        local.S3ManagedLedgerOffloadRole,
		S3ManagedLedgerOffloadRoleSessionName:             local.S3ManagedLedgerOffloadRoleSessionName,
		OffloadersDirectory:                               local.OffloadersDirectory,
		ManagedLedgerOffloadDriverMetadata:                local.ManagedLedgerOffloadDriverMetadata,
	}
}

// convertAutoSubscriptionCreation converts our local AutoSubscriptionCreationOverride to the external library type
func convertAutoSubscriptionCreation(local *v1alpha1.AutoSubscriptionCreationOverride) *utils.AutoSubscriptionCreationOverride {
	if local == nil {
		return nil
	}
	return &utils.AutoSubscriptionCreationOverride{
		AllowAutoSubscriptionCreation: local.AllowAutoSubscriptionCreation,
	}
}

// convertSchemaCompatibilityStrategy converts our local SchemaCompatibilityStrategy to the external library type
func convertSchemaCompatibilityStrategy(local *v1alpha1.SchemaCompatibilityStrategy) *utils.SchemaCompatibilityStrategy {
	if local == nil {
		return nil
	}
	strategy := utils.SchemaCompatibilityStrategy(string(*local))
	return &strategy
}

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
		params.Bundles = ptr.To(int32(4))
	}

	err := p.adminClient.Namespaces().CreateNsWithPolices(name, utils.Policies{
		Bundles: &utils.BundlesData{
			NumBundles: int(*params.Bundles),
		},
		SubscriptionAuthMode: utils.None,
		ReplicationClusters:  []string{},
	})
	if err != nil && !IsAlreadyExist(err) {
		return err
	}

	err = p.applyNamespacePolicies(name, params)
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
func (p *PulsarAdminClient) ApplyTopic(name string, params *TopicParams) (creationErr error, policyErr error) {
	completeTopicName := MakeCompleteTopicName(name, params.Persistent)
	topicName, err := utils.GetTopicName(completeTopicName)
	if err != nil {
		return err, nil
	}
	partitionNum := int(*params.Partitions)
	err = p.adminClient.Topics().Create(*topicName, partitionNum)
	if err != nil {
		if !IsAlreadyExist(err) {
			return err, nil
		}
		if partitionNum > 0 {
			// for partitioned topic, allow to change the partition number
			if err = p.adminClient.Topics().Update(*topicName, partitionNum); err != nil {
				return nil, err
			}
		}
	}

	err = p.applyTopicPolicies(topicName, params)
	if err != nil {
		return nil, err
	}

	return nil, nil
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
	if params.Deduplication != nil {
		err = p.adminClient.Topics().SetDeduplicationStatus(*topicName, *params.Deduplication)
		if err != nil {
			return err
		}
	}

	if params.CompactionThreshold != nil {
		err = p.adminClient.Topics().SetCompactionThreshold(*topicName, *params.CompactionThreshold)
		if err != nil {
			return err
		}
	}

	// Handle persistence policies
	if params.PersistencePolicies != nil {
		// Parse ManagedLedgerMaxMarkDeleteRate from string to float64
		var markDeleteRate float64
		if params.PersistencePolicies.ManagedLedgerMaxMarkDeleteRate != nil {
			var err error
			markDeleteRate, err = strconv.ParseFloat(*params.PersistencePolicies.ManagedLedgerMaxMarkDeleteRate, 64)
			if err != nil {
				return err
			}
		}

		persistenceData := utils.PersistenceData{
			ManagedLedgerMaxMarkDeleteRate: markDeleteRate,
		}
		if params.PersistencePolicies.BookkeeperEnsemble != nil {
			persistenceData.BookkeeperEnsemble = int64(*params.PersistencePolicies.BookkeeperEnsemble)
		}
		if params.PersistencePolicies.BookkeeperWriteQuorum != nil {
			persistenceData.BookkeeperWriteQuorum = int64(*params.PersistencePolicies.BookkeeperWriteQuorum)
		}
		if params.PersistencePolicies.BookkeeperAckQuorum != nil {
			persistenceData.BookkeeperAckQuorum = int64(*params.PersistencePolicies.BookkeeperAckQuorum)
		}
		err = p.adminClient.Topics().SetPersistence(*topicName, persistenceData)
		if err != nil {
			return err
		}
	}

	// Handle delayed delivery
	if params.DelayedDelivery != nil {
		delayedDeliveryData := utils.DelayedDeliveryData{
			Active: *params.DelayedDelivery.Active,
		}
		if params.DelayedDelivery.TickTimeMillis != nil {
			// Convert milliseconds to seconds (float64)
			delayedDeliveryData.TickTime = float64(*params.DelayedDelivery.TickTimeMillis) / 1000.0
		}
		err = p.adminClient.Topics().SetDelayedDelivery(*topicName, delayedDeliveryData)
		if err != nil {
			return err
		}
	}

	// Handle dispatch rate
	if params.DispatchRate != nil {
		dispatchRateData := utils.DispatchRateData{}
		if params.DispatchRate.DispatchThrottlingRateInMsg != nil {
			dispatchRateData.DispatchThrottlingRateInMsg = int64(*params.DispatchRate.DispatchThrottlingRateInMsg)
		}
		if params.DispatchRate.DispatchThrottlingRateInByte != nil {
			dispatchRateData.DispatchThrottlingRateInByte = *params.DispatchRate.DispatchThrottlingRateInByte
		}
		if params.DispatchRate.RatePeriodInSecond != nil {
			dispatchRateData.RatePeriodInSecond = int64(*params.DispatchRate.RatePeriodInSecond)
		}
		err = p.adminClient.Topics().SetDispatchRate(*topicName, dispatchRateData)
		if err != nil {
			return err
		}
	}

	// Handle publish rate
	if params.PublishRate != nil {
		publishRateData := utils.PublishRateData{}
		if params.PublishRate.PublishThrottlingRateInMsg != nil {
			publishRateData.PublishThrottlingRateInMsg = int64(*params.PublishRate.PublishThrottlingRateInMsg)
		}
		if params.PublishRate.PublishThrottlingRateInByte != nil {
			publishRateData.PublishThrottlingRateInByte = *params.PublishRate.PublishThrottlingRateInByte
		}
		err = p.adminClient.Topics().SetPublishRate(*topicName, publishRateData)
		if err != nil {
			return err
		}
	}

	// Handle inactive topic policies
	if params.InactiveTopicPolicies != nil {
		inactiveTopicPolicies := utils.InactiveTopicPolicies{}
		if params.InactiveTopicPolicies.InactiveTopicDeleteMode != nil {
			deleteMode := utils.InactiveTopicDeleteMode(*params.InactiveTopicPolicies.InactiveTopicDeleteMode)
			inactiveTopicPolicies.InactiveTopicDeleteMode = &deleteMode
		}
		if params.InactiveTopicPolicies.MaxInactiveDurationInSeconds != nil {
			inactiveTopicPolicies.MaxInactiveDurationSeconds = int(*params.InactiveTopicPolicies.MaxInactiveDurationInSeconds)
		}
		if params.InactiveTopicPolicies.DeleteWhileInactive != nil {
			inactiveTopicPolicies.DeleteWhileInactive = *params.InactiveTopicPolicies.DeleteWhileInactive
		}
		err = p.adminClient.Topics().SetInactiveTopicPolicies(*topicName, inactiveTopicPolicies)
		if err != nil {
			return err
		}
	}

	// Handle subscribe rate
	if params.SubscribeRate != nil {
		subscribeRateData := utils.SubscribeRate{
			SubscribeThrottlingRatePerConsumer: -1, // default to unlimited
			RatePeriodInSecond:                 30, // default period
		}
		if params.SubscribeRate.SubscribeThrottlingRatePerConsumer != nil {
			subscribeRateData.SubscribeThrottlingRatePerConsumer = int(*params.SubscribeRate.SubscribeThrottlingRatePerConsumer)
		}
		if params.SubscribeRate.RatePeriodInSecond != nil {
			subscribeRateData.RatePeriodInSecond = int(*params.SubscribeRate.RatePeriodInSecond)
		}
		err = p.adminClient.Topics().SetSubscribeRate(*topicName, subscribeRateData)
		if err != nil {
			return err
		}
	}

	// Handle max message size
	if params.MaxMessageSize != nil {
		err = p.adminClient.Topics().SetMaxMessageSize(*topicName, int(*params.MaxMessageSize))
		if err != nil {
			return err
		}
	}

	// Handle max consumers per subscription
	if params.MaxConsumersPerSubscription != nil {
		err = p.adminClient.Topics().SetMaxConsumersPerSubscription(*topicName, int(*params.MaxConsumersPerSubscription))
		if err != nil {
			return err
		}
	}

	// Handle max subscriptions per topic
	if params.MaxSubscriptionsPerTopic != nil {
		err = p.adminClient.Topics().SetMaxSubscriptionsPerTopic(*topicName, int(*params.MaxSubscriptionsPerTopic))
		if err != nil {
			return err
		}
	}

	// Handle schema validation enforced
	if params.SchemaValidationEnforced != nil {
		err = p.adminClient.Topics().SetSchemaValidationEnforced(*topicName, *params.SchemaValidationEnforced)
		if err != nil {
			return err
		}
	}

	// Handle subscription dispatch rate
	if params.SubscriptionDispatchRate != nil {
		dispatchRateData := utils.DispatchRateData{}
		if params.SubscriptionDispatchRate.DispatchThrottlingRateInMsg != nil {
			dispatchRateData.DispatchThrottlingRateInMsg = int64(*params.SubscriptionDispatchRate.DispatchThrottlingRateInMsg)
		}
		if params.SubscriptionDispatchRate.DispatchThrottlingRateInByte != nil {
			dispatchRateData.DispatchThrottlingRateInByte = *params.SubscriptionDispatchRate.DispatchThrottlingRateInByte
		}
		if params.SubscriptionDispatchRate.RatePeriodInSecond != nil {
			dispatchRateData.RatePeriodInSecond = int64(*params.SubscriptionDispatchRate.RatePeriodInSecond)
		}
		err = p.adminClient.Topics().SetSubscriptionDispatchRate(*topicName, dispatchRateData)
		if err != nil {
			return err
		}
	}

	// Handle replicator dispatch rate
	if params.ReplicatorDispatchRate != nil {
		dispatchRateData := utils.DispatchRateData{}
		if params.ReplicatorDispatchRate.DispatchThrottlingRateInMsg != nil {
			dispatchRateData.DispatchThrottlingRateInMsg = int64(*params.ReplicatorDispatchRate.DispatchThrottlingRateInMsg)
		}
		if params.ReplicatorDispatchRate.DispatchThrottlingRateInByte != nil {
			dispatchRateData.DispatchThrottlingRateInByte = *params.ReplicatorDispatchRate.DispatchThrottlingRateInByte
		}
		if params.ReplicatorDispatchRate.RatePeriodInSecond != nil {
			dispatchRateData.RatePeriodInSecond = int64(*params.ReplicatorDispatchRate.RatePeriodInSecond)
		}
		err = p.adminClient.Topics().SetReplicatorDispatchRate(*topicName, dispatchRateData)
		if err != nil {
			return err
		}
	}

	// Handle deduplication snapshot interval
	if params.DeduplicationSnapshotInterval != nil {
		err = p.adminClient.Topics().SetDeduplicationSnapshotInterval(*topicName, int(*params.DeduplicationSnapshotInterval))
		if err != nil {
			return err
		}
	}

	// Handle offload policies
	if params.OffloadPolicies != nil {
		externalOffloadPolicies := convertOffloadPolicies(params.OffloadPolicies)
		err = p.adminClient.Topics().SetOffloadPolicies(*topicName, *externalOffloadPolicies)
		if err != nil {
			return err
		}
	}

	// Handle auto subscription creation
	if params.AutoSubscriptionCreation != nil {
		externalAutoSubscription := convertAutoSubscriptionCreation(params.AutoSubscriptionCreation)
		err = p.adminClient.Topics().SetAutoSubscriptionCreation(*topicName, *externalAutoSubscription)
		if err != nil {
			return err
		}
	}

	// Handle schema compatibility strategy
	if params.SchemaCompatibilityStrategy != nil {
		externalSchemaStrategy := convertSchemaCompatibilityStrategy(params.SchemaCompatibilityStrategy)
		err = p.adminClient.Topics().SetSchemaCompatibilityStrategy(*topicName, *externalSchemaStrategy)
		if err != nil {
			return err
		}
	}

	// Handle topic properties
	if len(params.Properties) > 0 {
		err = p.adminClient.Topics().UpdateProperties(*topicName, params.Properties)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetTopicClusters get the assigned clusters of the topic to the local default cluster
func (p *PulsarAdminClient) GetTopicClusters(name string, persistent *bool) ([]string, error) {
	completeTopicName := MakeCompleteTopicName(name, persistent)
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
	completeTopicName := MakeCompleteTopicName(name, persistent)
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

func (p *PulsarAdminClient) applyNamespacePolicies(completeNSName string, params *NamespaceParams) error {
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

	if params.OffloadThresholdTime != nil {
		t, err := params.OffloadThresholdTime.Parse()
		if err != nil {
			return err
		}
		err = p.adminClient.Namespaces().SetOffloadThresholdInSeconds(*naName, int64(t.Seconds()))
		if err != nil {
			return err
		}
	}

	if params.OffloadThresholdSize != nil {
		s := params.OffloadThresholdSize.Value()
		err = p.adminClient.Namespaces().SetOffloadThreshold(*naName, s)
		if err != nil {
			return err
		}
	}

	if len(params.ReplicationClusters) > 0 {
		if c, err := p.adminClient.Namespaces().GetNamespaceReplicationClusters(completeNSName); err == nil {
			if !reflect.DeepEqual(c, params.ReplicationClusters) {
				err = p.adminClient.Namespaces().SetNamespaceReplicationClusters(completeNSName, params.ReplicationClusters)
				if err != nil {
					return err
				}
			}
		} else {
			return err
		}
	}

	if params.Deduplication != nil {
		err = p.adminClient.Namespaces().SetDeduplicationStatus(completeNSName, *params.Deduplication)
		if err != nil {
			return err
		}
	}
	if params.BookieAffinityGroup != nil {
		err = p.adminClient.Namespaces().SetBookieAffinityGroup(completeNSName, utils.BookieAffinityGroupData{
			BookkeeperAffinityGroupPrimary:   params.BookieAffinityGroup.BookkeeperAffinityGroupPrimary,
			BookkeeperAffinityGroupSecondary: params.BookieAffinityGroup.BookkeeperAffinityGroupSecondary,
		})
		if err != nil {
			return err
		}
	} else {
		err = p.adminClient.Namespaces().DeleteBookieAffinityGroup(completeNSName)
		if err != nil {
			return err
		}
	}

	// Handle topic auto-creation configuration
	if params.TopicAutoCreationConfig != nil {
		topicTypeStr, err := utils.ParseTopicType(params.TopicAutoCreationConfig.Type)
		if err != nil {
			return err
		}

		// Convert operator's TopicAutoCreationConfig to Pulsar client's TopicAutoCreationConfig
		config := &utils.TopicAutoCreationConfig{
			Allow: params.TopicAutoCreationConfig.Allow,
			Type:  topicTypeStr,
		}

		// Set default partitions
		if params.TopicAutoCreationConfig.Partitions != nil {
			partitions := int(*params.TopicAutoCreationConfig.Partitions)
			config.Partitions = &partitions
		}

		// Call Pulsar client API to set topic auto-creation configuration
		err = p.adminClient.Namespaces().SetTopicAutoCreation(*naName, *config)
		if err != nil {
			return err
		}
	} else {
		// If no configuration is specified, try to remove topic auto-creation configuration (ignore errors if it doesn't exist)
		err = p.adminClient.Namespaces().RemoveTopicAutoCreation(*naName)
		if err != nil && !IsNotFound(err) {
			return err
		}
	}

	// Handle schema validation enforcement
	if params.SchemaValidationEnforced != nil {
		err = p.adminClient.Namespaces().SetSchemaValidationEnforced(*naName, *params.SchemaValidationEnforced)
		if err != nil {
			return err
		}
	}

	// Handle dispatch rate limiting
	if params.DispatchRate != nil {
		rate := utils.DispatchRate{
			DispatchThrottlingRateInMsg:  -1, // default to unlimited
			DispatchThrottlingRateInByte: -1, // default to unlimited
			RatePeriodInSecond:           1,  // default period
		}

		if params.DispatchRate.DispatchThrottlingRateInMsg != nil {
			rate.DispatchThrottlingRateInMsg = int(*params.DispatchRate.DispatchThrottlingRateInMsg)
		}
		if params.DispatchRate.DispatchThrottlingRateInByte != nil {
			rate.DispatchThrottlingRateInByte = *params.DispatchRate.DispatchThrottlingRateInByte
		}
		if params.DispatchRate.RatePeriodInSecond != nil {
			rate.RatePeriodInSecond = int(*params.DispatchRate.RatePeriodInSecond)
		}

		err = p.adminClient.Namespaces().SetDispatchRate(*naName, rate)
		if err != nil {
			return err
		}
	}

	// Handle subscription dispatch rate limiting
	if params.SubscriptionDispatchRate != nil {
		rate := utils.DispatchRate{
			DispatchThrottlingRateInMsg:  -1, // default to unlimited
			DispatchThrottlingRateInByte: -1, // default to unlimited
			RatePeriodInSecond:           1,  // default period
		}

		if params.SubscriptionDispatchRate.DispatchThrottlingRateInMsg != nil {
			rate.DispatchThrottlingRateInMsg = int(*params.SubscriptionDispatchRate.DispatchThrottlingRateInMsg)
		}
		if params.SubscriptionDispatchRate.DispatchThrottlingRateInByte != nil {
			rate.DispatchThrottlingRateInByte = *params.SubscriptionDispatchRate.DispatchThrottlingRateInByte
		}
		if params.SubscriptionDispatchRate.RatePeriodInSecond != nil {
			rate.RatePeriodInSecond = int(*params.SubscriptionDispatchRate.RatePeriodInSecond)
		}

		err = p.adminClient.Namespaces().SetSubscriptionDispatchRate(*naName, rate)
		if err != nil {
			return err
		}
	}

	// Handle replicator dispatch rate limiting
	if params.ReplicatorDispatchRate != nil {
		rate := utils.DispatchRate{
			DispatchThrottlingRateInMsg:  -1, // default to unlimited
			DispatchThrottlingRateInByte: -1, // default to unlimited
			RatePeriodInSecond:           1,  // default period
		}

		if params.ReplicatorDispatchRate.DispatchThrottlingRateInMsg != nil {
			rate.DispatchThrottlingRateInMsg = int(*params.ReplicatorDispatchRate.DispatchThrottlingRateInMsg)
		}
		if params.ReplicatorDispatchRate.DispatchThrottlingRateInByte != nil {
			rate.DispatchThrottlingRateInByte = *params.ReplicatorDispatchRate.DispatchThrottlingRateInByte
		}
		if params.ReplicatorDispatchRate.RatePeriodInSecond != nil {
			rate.RatePeriodInSecond = int(*params.ReplicatorDispatchRate.RatePeriodInSecond)
		}

		err = p.adminClient.Namespaces().SetReplicatorDispatchRate(*naName, rate)
		if err != nil {
			return err
		}
	}

	// Handle publish rate limiting
	if params.PublishRate != nil {
		rate := utils.PublishRate{
			PublishThrottlingRateInMsg:  -1, // default to unlimited
			PublishThrottlingRateInByte: -1, // default to unlimited
		}

		if params.PublishRate.PublishThrottlingRateInMsg != nil {
			rate.PublishThrottlingRateInMsg = int(*params.PublishRate.PublishThrottlingRateInMsg)
		}
		if params.PublishRate.PublishThrottlingRateInByte != nil {
			rate.PublishThrottlingRateInByte = *params.PublishRate.PublishThrottlingRateInByte
		}

		err = p.adminClient.Namespaces().SetPublishRate(*naName, rate)
		if err != nil {
			return err
		}
	}

	// Handle subscribe rate limiting
	if params.SubscribeRate != nil {
		rate := utils.SubscribeRate{
			SubscribeThrottlingRatePerConsumer: -1, // default to unlimited
			RatePeriodInSecond:                 30, // default period
		}

		if params.SubscribeRate.SubscribeThrottlingRatePerConsumer != nil {
			rate.SubscribeThrottlingRatePerConsumer = int(*params.SubscribeRate.SubscribeThrottlingRatePerConsumer)
		}
		if params.SubscribeRate.RatePeriodInSecond != nil {
			rate.RatePeriodInSecond = int(*params.SubscribeRate.RatePeriodInSecond)
		}

		err = p.adminClient.Namespaces().SetSubscribeRate(*naName, rate)
		if err != nil {
			return err
		}
	}

	// Handle compaction threshold
	if params.CompactionThreshold != nil {
		err = p.adminClient.Namespaces().SetCompactionThreshold(*naName, *params.CompactionThreshold)
		if err != nil {
			return err
		}
	}

	// Handle schema auto-update policy
	if params.IsAllowAutoUpdateSchema != nil {
		err = p.adminClient.Namespaces().SetIsAllowAutoUpdateSchema(*naName, *params.IsAllowAutoUpdateSchema)
		if err != nil {
			return err
		}
	}

	if params.SchemaCompatibilityStrategy != nil {
		externalSchemaStrategy := convertSchemaCompatibilityStrategy(params.SchemaCompatibilityStrategy)
		err := p.adminClient.Namespaces().SetSchemaCompatibilityStrategy(*naName, *externalSchemaStrategy)
		if err != nil {
			return err
		}
	}

	// Handle encryption requirement
	if params.EncryptionRequired != nil {
		err = p.adminClient.Namespaces().SetEncryptionRequiredStatus(*naName, *params.EncryptionRequired)
		if err != nil {
			return err
		}
	}

	// Handle subscription authentication mode
	if params.SubscriptionAuthMode != nil {
		err = p.adminClient.Namespaces().SetSubscriptionAuthMode(*naName, *params.SubscriptionAuthMode)
		if err != nil {
			return err
		}
	}

	// Handle anti-affinity group
	if params.AntiAffinityGroup != nil {
		err = p.adminClient.Namespaces().SetNamespaceAntiAffinityGroup(completeNSName, *params.AntiAffinityGroup)
		if err != nil {
			return err
		}
	} else {
		// Remove anti-affinity group if not specified
		err = p.adminClient.Namespaces().DeleteNamespaceAntiAffinityGroup(completeNSName)
		if err != nil && !IsNotFound(err) {
			return err
		}
	}

	// Handle schema auto update compatibility strategy
	if params.SchemaAutoUpdateCompatibilityStrategy != nil {
		err = p.adminClient.Namespaces().SetSchemaAutoUpdateCompatibilityStrategy(*naName, *params.SchemaAutoUpdateCompatibilityStrategy)
		if err != nil {
			return err
		}
	}

	return nil
}

func MakeCompleteTopicName(topicName string, persistent *bool) string {
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

// GetTopicPermissions retrieve permission by name
func (p *PulsarAdminClient) GetTopicPermissions(topic string) (map[string][]utils.AuthAction, error) {
	topicName, err := utils.GetTopicName(topic)
	if err != nil {
		return nil, err
	}
	return p.adminClient.Topics().GetPermissions(*topicName)
}

// GetNamespacePermissions retrieve permission by name
func (p *PulsarAdminClient) GetNamespacePermissions(namespaceName string) (map[string][]utils.AuthAction, error) {
	namespace, err := utils.GetNamespaceName(namespaceName)
	if err != nil {
		return nil, err
	}
	return p.adminClient.Namespaces().GetNamespacePermissions(*namespace)
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
		clusterData.BrokerClientTrustCertsFilePath = param.BrokerClientTrustCertsFilePath
		clusterData.BrokerClientTLSEnabled = true
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

// DeletePulsarPackage deletes a pulsar package
func (p *PulsarAdminClient) DeletePulsarPackage(packageURL string) error {
	return p.adminClient.Packages().Delete(packageURL)
}

// ApplyPulsarPackage creates or updates a pulsar package
func (p *PulsarAdminClient) ApplyPulsarPackage(packageURL, filePath, description, contact string, properties map[string]string, changed bool) error {
	packageName, err := utils.GetPackageName(packageURL)
	if err != nil {
		return fmt.Errorf("failed to get package name: %w", err)
	}
	if changed {
		err = p.adminClient.Packages().UpdateMetadata(packageName.String(), description, contact, properties)
		if err != nil {
			if !IsAlreadyExist(err) {
				return fmt.Errorf("failed to update package metadata: %w", err)
			}
		}
	} else {
		err = p.adminClient.Packages().Upload(packageName.String(), filePath, description, contact, properties)
		if err != nil {
			if !IsAlreadyExist(err) {
				return fmt.Errorf("failed to upload package: %w", err)
			}
		}
	}

	return nil
}

// DeletePulsarFunction deletes a pulsar function
func (p *PulsarAdminClient) DeletePulsarFunction(tenant, namespace, name string) error {
	return p.adminClient.Functions().DeleteFunction(tenant, namespace, name)
}

// ApplyPulsarFunction creates or updates a pulsar function
func (p *PulsarAdminClient) ApplyPulsarFunction(tenant, namespace, name, packageURL string, param *v1alpha1.PulsarFunctionSpec, changed bool) error {
	functionConfig := utils.FunctionConfig{
		Tenant:                         tenant,
		Namespace:                      namespace,
		Name:                           name,
		ClassName:                      param.ClassName,
		Inputs:                         param.Inputs,
		Parallelism:                    param.Parallelism,
		TimeoutMs:                      param.TimeoutMs,
		TopicsPattern:                  param.TopicsPattern,
		CleanupSubscription:            param.CleanupSubscription,
		RetainOrdering:                 param.RetainOrdering,
		RetainKeyOrdering:              param.RetainKeyOrdering,
		ForwardSourceMessageProperty:   param.ForwardSourceMessageProperty,
		AutoAck:                        param.AutoAck,
		MaxMessageRetries:              param.MaxMessageRetries,
		CustomSerdeInputs:              param.CustomSerdeInputs,
		CustomSchemaInputs:             param.CustomSchemaInputs,
		InputTypeClassName:             param.InputTypeClassName,
		Output:                         param.Output,
		OutputSerdeClassName:           param.OutputSerdeClassName,
		OutputSchemaType:               param.OutputSchemaType,
		OutputTypeClassName:            param.OutputTypeClassName,
		CustomSchemaOutputs:            param.CustomSchemaOutputs,
		LogTopic:                       param.LogTopic,
		ProcessingGuarantees:           param.ProcessingGuarantees,
		DeadLetterTopic:                param.DeadLetterTopic,
		SubName:                        param.SubName,
		RuntimeFlags:                   param.RuntimeFlags,
		MaxPendingAsyncRequests:        param.MaxPendingAsyncRequests,
		ExposePulsarAdminClientEnabled: param.ExposePulsarAdminClientEnabled,
		SkipToLatest:                   param.SkipToLatest,
		SubscriptionPosition:           param.SubscriptionPosition,
	}

	if param.BatchBuilder != nil {
		functionConfig.BatchBuilder = *param.BatchBuilder
	}

	if param.ProducerConfig != nil {
		functionConfig.ProducerConfig = &utils.ProducerConfig{
			MaxPendingMessages:                 param.ProducerConfig.MaxPendingMessages,
			MaxPendingMessagesAcrossPartitions: param.ProducerConfig.MaxPendingMessagesAcrossPartitions,
			UseThreadLocalProducers:            param.ProducerConfig.UseThreadLocalProducers,
			BatchBuilder:                       param.ProducerConfig.BatchBuilder,
			CompressionType:                    param.ProducerConfig.CompressionType,
		}
		if param.ProducerConfig.CryptoConfig != nil {
			functionConfig.ProducerConfig.CryptoConfig = &utils.CryptoConfig{
				CryptoKeyReaderClassName:    param.ProducerConfig.CryptoConfig.CryptoKeyReaderClassName,
				CryptoKeyReaderConfig:       rutils.ConvertMap(param.ProducerConfig.CryptoConfig.CryptoKeyReaderConfig),
				EncryptionKeys:              param.ProducerConfig.CryptoConfig.EncryptionKeys,
				ProducerCryptoFailureAction: param.ProducerConfig.CryptoConfig.ProducerCryptoFailureAction,
				ConsumerCryptoFailureAction: param.ProducerConfig.CryptoConfig.ConsumerCryptoFailureAction,
			}
		}
	}

	if len(param.InputSpecs) > 0 {
		inputSpecs := make(map[string]utils.ConsumerConfig)
		for k, v := range param.InputSpecs {
			iSpec := utils.ConsumerConfig{
				SchemaType:         v.SchemaType,
				SerdeClassName:     v.SerdeClassName,
				RegexPattern:       v.RegexPattern,
				ReceiverQueueSize:  v.ReceiverQueueSize,
				SchemaProperties:   v.SchemaProperties,
				ConsumerProperties: v.ConsumerProperties,
				PoolMessages:       v.PoolMessages,
			}
			if v.CryptoConfig != nil {
				iSpec.CryptoConfig = &utils.CryptoConfig{
					CryptoKeyReaderClassName:    v.CryptoConfig.CryptoKeyReaderClassName,
					CryptoKeyReaderConfig:       rutils.ConvertMap(v.CryptoConfig.CryptoKeyReaderConfig),
					EncryptionKeys:              v.CryptoConfig.EncryptionKeys,
					ProducerCryptoFailureAction: v.CryptoConfig.ProducerCryptoFailureAction,
					ConsumerCryptoFailureAction: v.CryptoConfig.ConsumerCryptoFailureAction,
				}
			}
			inputSpecs[k] = iSpec
		}
		functionConfig.InputSpecs = inputSpecs
	}

	if param.Resources != nil {
		s, err := strconv.ParseFloat(param.Resources.CPU, 64)
		if err != nil {
			return fmt.Errorf("failed to parse cpu: %w", err)
		}
		functionConfig.Resources = &utils.Resources{
			CPU:  s,
			RAM:  param.Resources.RAM,
			Disk: param.Resources.Disk,
		}
	}

	if param.WindowConfig != nil {
		functionConfig.WindowConfig = &utils.WindowConfig{
			WindowLengthCount:             param.WindowConfig.WindowLengthCount,
			WindowLengthDurationMs:        param.WindowConfig.WindowLengthDurationMs,
			SlidingIntervalCount:          param.WindowConfig.SlidingIntervalCount,
			SlidingIntervalDurationMs:     param.WindowConfig.SlidingIntervalDurationMs,
			LateDataTopic:                 param.WindowConfig.LateDataTopic,
			MaxLagMs:                      param.WindowConfig.MaxLagMs,
			WatermarkEmitIntervalMs:       param.WindowConfig.WatermarkEmitIntervalMs,
			TimestampExtractorClassName:   param.WindowConfig.TimestampExtractorClassName,
			ActualWindowFunctionClassName: param.WindowConfig.ActualWindowFunctionClassName,
			ProcessingGuarantees:          param.WindowConfig.ProcessingGuarantees,
		}
	}

	if param.UserConfig != nil {
		var err error
		functionConfig.UserConfig, err = rutils.ConvertJSONToMapStringInterface(param.UserConfig)
		if err != nil {
			return fmt.Errorf("failed to convert user config: %w", err)
		}
	}

	if param.CustomRuntimeOptions != nil {
		jByte, err := param.CustomRuntimeOptions.MarshalJSON()
		if err != nil {
			return err
		}
		functionConfig.CustomRuntimeOptions = string(jByte)
	}

	if len(param.Secrets) > 0 {
		secrets := make(map[string]interface{})
		for k, v := range param.Secrets {
			secrets[k] = v
		}
		functionConfig.Secrets = secrets
	}

	if param.Jar != nil {
		functionConfig.Jar = &packageURL
	} else if param.Py != nil {
		functionConfig.Py = &packageURL
	} else if param.Go != nil {
		functionConfig.Go = &packageURL
	} else {
		return errors.New("FunctionConfig need to specify Jar, Py, or Go package URL")
	}

	var err error
	if changed {
		err = p.adminClient.Functions().UpdateFunctionWithURL(&functionConfig, packageURL, nil)
		if err != nil {
			if !IsAlreadyExist(err) {
				return fmt.Errorf("failed to update function: %w", err)
			}
		}
	} else {
		err = p.adminClient.Functions().CreateFuncWithURL(&functionConfig, packageURL)
		if err != nil {
			if !IsAlreadyExist(err) {
				return fmt.Errorf("failed to create function: %w", err)
			}
		}
	}

	return nil
}

// DeletePulsarSink deletes a pulsar sink
func (p *PulsarAdminClient) DeletePulsarSink(tenant, namespace, name string) error {
	return p.adminClient.Sinks().DeleteSink(tenant, namespace, name)
}

// ApplyPulsarSink creates or updates a pulsar sink
func (p *PulsarAdminClient) ApplyPulsarSink(tenant, namespace, name, packageURL string, param *v1alpha1.PulsarSinkSpec, changed bool) error {
	sinkConfig := utils.SinkConfig{
		Tenant:    tenant,
		Namespace: namespace,
		Name:      name,
		ClassName: param.ClassName,

		TopicsPattern: param.TopicsPattern,
		TimeoutMs:     param.TimeoutMs,

		CleanupSubscription: param.CleanupSubscription,
		RetainOrdering:      param.RetainOrdering,
		RetainKeyOrdering:   param.RetainKeyOrdering,
		AutoAck:             param.AutoAck,
		Parallelism:         param.Parallelism,

		SinkType: param.SinkType,
		Archive:  packageURL,

		ProcessingGuarantees:       param.ProcessingGuarantees,
		SourceSubscriptionName:     param.SourceSubscriptionName,
		SourceSubscriptionPosition: param.SourceSubscriptionPosition,
		RuntimeFlags:               param.RuntimeFlags,

		Inputs:                  param.Inputs,
		TopicToSerdeClassName:   param.TopicToSerdeClassName,
		TopicToSchemaType:       param.TopicToSchemaType,
		TopicToSchemaProperties: param.TopicToSchemaProperties,

		MaxMessageRetries:            param.MaxMessageRetries,
		DeadLetterTopic:              param.DeadLetterTopic,
		NegativeAckRedeliveryDelayMs: param.NegativeAckRedeliveryDelayMs,
		TransformFunction:            param.TransformFunction,
		TransformFunctionClassName:   param.TransformFunctionClassName,
		TransformFunctionConfig:      param.TransformFunctionConfig,
	}

	if param.Resources != nil {
		s, err := strconv.ParseFloat(param.Resources.CPU, 64)
		if err != nil {
			return fmt.Errorf("apply pulsar sink failed on parse resources: %s", err.Error())
		}
		sinkConfig.Resources = &utils.Resources{
			CPU:  s,
			RAM:  param.Resources.RAM,
			Disk: param.Resources.Disk,
		}
	}

	if len(param.InputSpecs) > 0 {
		inputSpecs := make(map[string]utils.ConsumerConfig)
		for k, v := range param.InputSpecs {
			iSpec := utils.ConsumerConfig{
				SchemaType:         v.SchemaType,
				SerdeClassName:     v.SerdeClassName,
				RegexPattern:       v.RegexPattern,
				ReceiverQueueSize:  v.ReceiverQueueSize,
				SchemaProperties:   v.SchemaProperties,
				ConsumerProperties: v.ConsumerProperties,
				PoolMessages:       v.PoolMessages,
			}
			if v.CryptoConfig != nil {
				iSpec.CryptoConfig = &utils.CryptoConfig{
					CryptoKeyReaderClassName:    v.CryptoConfig.CryptoKeyReaderClassName,
					CryptoKeyReaderConfig:       rutils.ConvertMap(v.CryptoConfig.CryptoKeyReaderConfig),
					EncryptionKeys:              v.CryptoConfig.EncryptionKeys,
					ProducerCryptoFailureAction: v.CryptoConfig.ProducerCryptoFailureAction,
					ConsumerCryptoFailureAction: v.CryptoConfig.ConsumerCryptoFailureAction,
				}
			}
			inputSpecs[k] = iSpec
		}
		sinkConfig.InputSpecs = inputSpecs
	}

	if param.Configs != nil {
		var err error
		sinkConfig.Configs, err = rutils.ConvertJSONToMapStringInterface(param.Configs)
		if err != nil {
			return fmt.Errorf("apply pulsar sink failed on convert configs: %s", err.Error())
		}
	}

	if param.CustomRuntimeOptions != nil {
		jByte, err := param.CustomRuntimeOptions.MarshalJSON()
		if err != nil {
			return err
		}
		sinkConfig.CustomRuntimeOptions = string(jByte)
	}

	if len(param.Secrets) > 0 {
		secrets := make(map[string]interface{})
		for k, v := range param.Secrets {
			secrets[k] = v
		}
		sinkConfig.Secrets = secrets
	}

	var err error
	if changed {
		if strings.HasPrefix(packageURL, "builtin://") {
			err = p.adminClient.Sinks().UpdateSink(&sinkConfig, packageURL, nil)
			if err != nil {
				if !IsAlreadyExist(err) {
					return fmt.Errorf("apply pulsar sink failed on update sink: %s", err.Error())
				}
			}
		} else {
			err = p.adminClient.Sinks().UpdateSinkWithURL(&sinkConfig, packageURL, nil)
			if err != nil {
				if !IsAlreadyExist(err) {
					return fmt.Errorf("apply pulsar sink failed on update sink with url: %s", err.Error())
				}
			}
		}
	} else {
		if strings.HasPrefix(packageURL, "builtin://") {
			err = p.adminClient.Sinks().CreateSink(&sinkConfig, packageURL)
			if err != nil {
				if !IsAlreadyExist(err) {
					return fmt.Errorf("apply pulsar sink failed on create sink: %s", err.Error())
				}
			}
		} else {
			err = p.adminClient.Sinks().CreateSinkWithURL(&sinkConfig, packageURL)
			if err != nil {
				if !IsAlreadyExist(err) {
					return fmt.Errorf("apply pulsar sink failed on create sink with url: %s", err.Error())
				}
			}
		}
	}

	return nil
}

// DeletePulsarSource deletes a pulsar source
func (p *PulsarAdminClient) DeletePulsarSource(tenant, namespace, name string) error {
	return p.adminClient.Sources().DeleteSource(tenant, namespace, name)
}

// ApplyPulsarSource creates or updates a pulsar source
func (p *PulsarAdminClient) ApplyPulsarSource(tenant, namespace, name, packageURL string, param *v1alpha1.PulsarSourceSpec, changed bool) error {
	sourceConfig := utils.SourceConfig{
		Tenant:    tenant,
		Namespace: namespace,
		Name:      name,
		ClassName: param.ClassName,

		TopicName:      param.TopicName,
		SerdeClassName: param.SerdeClassName,
		SchemaType:     param.SchemaType,

		Parallelism:          param.Parallelism,
		ProcessingGuarantees: param.ProcessingGuarantees,

		Archive: packageURL,

		RuntimeFlags: param.RuntimeFlags,

		BatchBuilder: param.BatchBuilder,
	}

	if param.Resources != nil {
		s, err := strconv.ParseFloat(param.Resources.CPU, 64)
		if err != nil {
			return fmt.Errorf("apply pulsar source failed on parse resources: %s", err.Error())
		}
		sourceConfig.Resources = &utils.Resources{
			CPU:  s,
			RAM:  param.Resources.RAM,
			Disk: param.Resources.Disk,
		}
	}

	if param.ProducerConfig != nil {
		sourceConfig.ProducerConfig = &utils.ProducerConfig{
			MaxPendingMessages:                 param.ProducerConfig.MaxPendingMessages,
			MaxPendingMessagesAcrossPartitions: param.ProducerConfig.MaxPendingMessagesAcrossPartitions,
			UseThreadLocalProducers:            param.ProducerConfig.UseThreadLocalProducers,
			BatchBuilder:                       param.ProducerConfig.BatchBuilder,
			CompressionType:                    param.ProducerConfig.CompressionType,
		}
		if param.ProducerConfig.CryptoConfig != nil {
			sourceConfig.ProducerConfig.CryptoConfig = &utils.CryptoConfig{
				CryptoKeyReaderClassName:    param.ProducerConfig.CryptoConfig.CryptoKeyReaderClassName,
				CryptoKeyReaderConfig:       rutils.ConvertMap(param.ProducerConfig.CryptoConfig.CryptoKeyReaderConfig),
				EncryptionKeys:              param.ProducerConfig.CryptoConfig.EncryptionKeys,
				ProducerCryptoFailureAction: param.ProducerConfig.CryptoConfig.ProducerCryptoFailureAction,
				ConsumerCryptoFailureAction: param.ProducerConfig.CryptoConfig.ConsumerCryptoFailureAction,
			}
		}
	}

	if param.BatchSourceConfig != nil {
		sourceConfig.BatchSourceConfig = &utils.BatchSourceConfig{
			DiscoveryTriggererClassName: param.BatchSourceConfig.DiscoveryTriggererClassName,
		}
		if param.BatchSourceConfig.DiscoveryTriggererConfig != nil {
			var err error
			sourceConfig.BatchSourceConfig.DiscoveryTriggererConfig, err = rutils.ConvertJSONToMapStringInterface(param.BatchSourceConfig.DiscoveryTriggererConfig)
			if err != nil {
				return fmt.Errorf("apply pulsar source failed on convert discovery triggerer config: %s", err.Error())
			}
		}
	}

	if param.Configs != nil {
		var err error
		sourceConfig.Configs, err = rutils.ConvertJSONToMapStringInterface(param.Configs)
		if err != nil {
			return fmt.Errorf("apply pulsar source failed on convert configs: %s", err.Error())
		}
	}

	if len(param.Secrets) > 0 {
		secrets := make(map[string]interface{})
		for k, v := range param.Secrets {
			secrets[k] = v
		}
		sourceConfig.Secrets = secrets
	}

	if param.CustomRuntimeOptions != nil {
		jByte, err := param.CustomRuntimeOptions.MarshalJSON()
		if err != nil {
			return err
		}
		sourceConfig.CustomRuntimeOptions = string(jByte)
	}

	var err error

	if changed {
		if strings.HasPrefix(packageURL, "builtin://") {
			err = p.adminClient.Sources().UpdateSource(&sourceConfig, packageURL, nil)
			if err != nil && !IsAlreadyExist(err) {
				return fmt.Errorf("apply pulsar source failed on update source: %s", err.Error())
			}
		} else {
			err = p.adminClient.Sources().UpdateSourceWithURL(&sourceConfig, packageURL, nil)
			if err != nil && !IsAlreadyExist(err) {
				return fmt.Errorf("apply pulsar source failed on update source with url: %s", err.Error())
			}
		}
	} else {
		if strings.HasPrefix(packageURL, "builtin://") {
			err = p.adminClient.Sources().CreateSource(&sourceConfig, packageURL)
			if err != nil && !IsAlreadyExist(err) {
				return fmt.Errorf("apply pulsar source failed on create source: %s", err.Error())
			}
		} else {
			err = p.adminClient.Sources().CreateSourceWithURL(&sourceConfig, packageURL)
			if err != nil && !IsAlreadyExist(err) {
				return fmt.Errorf("apply pulsar source failed on create source with url: %s", err.Error())
			}
		}
	}

	return nil
}

// CheckPulsarFunctionExist check whether the function is created or not
func (p *PulsarAdminClient) CheckPulsarFunctionExist(tenant, namespace, name string) (bool, error) {
	_, err := p.adminClient.Functions().GetFunction(tenant, namespace, name)

	if err != nil {
		if IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// CheckPulsarSinkExist check whether the sink is created or not
func (p *PulsarAdminClient) CheckPulsarSinkExist(tenant, namespace, name string) (bool, error) {
	_, err := p.adminClient.Sinks().GetSink(tenant, namespace, name)

	if err != nil {
		if IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// CheckPulsarSourceExist check whether the source is created or not
func (p *PulsarAdminClient) CheckPulsarSourceExist(tenant, namespace, name string) (bool, error) {
	_, err := p.adminClient.Sources().GetSource(tenant, namespace, name)

	if err != nil {
		if IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// CheckPulsarPackageExist check whether the package is created or not
func (p *PulsarAdminClient) CheckPulsarPackageExist(packageURL string) (bool, error) {
	_, err := p.adminClient.Packages().GetMetadata(packageURL)

	if err != nil {
		if IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// GetTenantAllowedClusters get the allowed clusters of the tenant
func (p *PulsarAdminClient) GetTenantAllowedClusters(tenantName string) ([]string, error) {
	tenant, err := p.adminClient.Tenants().Get(tenantName)
	if err != nil {
		return []string{}, err
	}

	return tenant.AllowedClusters, nil
}

// GetNSIsolationPolicy get the ns-isolation-policy
func (p *PulsarAdminClient) GetNSIsolationPolicy(policyName, clusterName string) (*utils.NamespaceIsolationData, error) {
	policyData, err := p.adminClient.NsIsolationPolicy().GetNamespaceIsolationPolicy(clusterName, policyName)
	if err != nil {
		return nil, err
	}

	return policyData, nil
}

// CreateNSIsolationPolicy create a ns-isolation-policy
func (p *PulsarAdminClient) CreateNSIsolationPolicy(policyName, clusterName string, policyData utils.NamespaceIsolationData) error {
	err := p.adminClient.NsIsolationPolicy().CreateNamespaceIsolationPolicy(clusterName, policyName, policyData)
	if err != nil {
		return err
	}

	return nil
}

// DeleteNSIsolationPolicy delete the ns-isolation-policy
func (p *PulsarAdminClient) DeleteNSIsolationPolicy(policyName, clusterName string) error {
	err := p.adminClient.NsIsolationPolicy().DeleteNamespaceIsolationPolicy(clusterName, policyName)
	if err != nil {
		return err
	}

	return nil
}

// GetPulsarPackageMetadata retrieves package information
func (p *PulsarAdminClient) GetPulsarPackageMetadata(packageURL string) (*utils.PackageMetadata, error) {
	pkg, err := p.adminClient.Packages().GetMetadata(packageURL)
	if err != nil {
		return nil, err
	}

	return &pkg, nil
}
