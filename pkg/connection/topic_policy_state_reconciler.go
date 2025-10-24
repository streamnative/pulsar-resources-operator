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

package connection

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/apache/pulsar-client-go/pulsaradmin/pkg/utils"
	"github.com/go-logr/logr"

	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	"github.com/streamnative/pulsar-resources-operator/pkg/admin"
	"github.com/streamnative/pulsar-resources-operator/pkg/reconciler"
)

const (
	// PulsarTopicPolicyStateAnnotation stores the last applied topic policy state.
	// We keep the historic key to remain backward compatible with existing annotations.
	PulsarTopicPolicyStateAnnotation = "pulsartopics.resource.streamnative.io/compaction-state"
)

type topicPolicyState struct {
	CompactionThreshold               *int64   `json:"compactionThreshold,omitempty"`
	MessageTTL                        bool     `json:"messageTTL,omitempty"`
	MaxProducers                      bool     `json:"maxProducers,omitempty"`
	MaxConsumers                      bool     `json:"maxConsumers,omitempty"`
	MaxUnAckedMessagesPerConsumer     bool     `json:"maxUnAckedMessagesPerConsumer,omitempty"`
	MaxUnAckedMessagesPerSubscription bool     `json:"maxUnAckedMessagesPerSubscription,omitempty"`
	Retention                         bool     `json:"retention,omitempty"`
	BacklogQuotaType                  *string  `json:"backlogQuotaType,omitempty"`
	Deduplication                     bool     `json:"deduplication,omitempty"`
	PersistencePolicies               bool     `json:"persistencePolicies,omitempty"`
	DelayedDelivery                   bool     `json:"delayedDelivery,omitempty"`
	DispatchRate                      bool     `json:"dispatchRate,omitempty"`
	PublishRate                       bool     `json:"publishRate,omitempty"`
	InactiveTopicPolicies             bool     `json:"inactiveTopicPolicies,omitempty"`
	SubscribeRate                     bool     `json:"subscribeRate,omitempty"`
	MaxMessageSize                    bool     `json:"maxMessageSize,omitempty"`
	MaxConsumersPerSubscription       bool     `json:"maxConsumersPerSubscription,omitempty"`
	MaxSubscriptionsPerTopic          bool     `json:"maxSubscriptionsPerTopic,omitempty"`
	SchemaValidationEnforced          bool     `json:"schemaValidationEnforced,omitempty"`
	SubscriptionDispatchRate          bool     `json:"subscriptionDispatchRate,omitempty"`
	ReplicatorDispatchRate            bool     `json:"replicatorDispatchRate,omitempty"`
	DeduplicationSnapshotInterval     bool     `json:"deduplicationSnapshotInterval,omitempty"`
	OffloadPolicies                   bool     `json:"offloadPolicies,omitempty"`
	AutoSubscriptionCreation          bool     `json:"autoSubscriptionCreation,omitempty"`
	SchemaCompatibilityStrategy       bool     `json:"schemaCompatibilityStrategy,omitempty"`
	PropertiesKeys                    []string `json:"propertiesKeys,omitempty"`
}

type topicPolicyOperation struct {
	Kind  string      `json:"kind"`
	Value interface{} `json:"value,omitempty"`
}

const (
	policyOpSetCompactionThreshold         = "setCompactionThreshold"
	policyOpRemoveCompactionThreshold      = "removeCompactionThreshold"
	policyOpRemoveMessageTTL               = "removeMessageTTL"
	policyOpRemoveMaxProducers             = "removeMaxProducers"
	policyOpRemoveMaxConsumers             = "removeMaxConsumers"
	policyOpRemoveMaxUnackConsumer         = "removeMaxUnackMessagesPerConsumer"
	policyOpRemoveMaxUnackSubscription     = "removeMaxUnackMessagesPerSubscription"
	policyOpRemoveRetention                = "removeRetention"
	policyOpRemoveBacklogQuota             = "removeBacklogQuota"
	policyOpRemoveDeduplication            = "removeDeduplication"
	policyOpRemovePersistence              = "removePersistence"
	policyOpRemoveDelayedDelivery          = "removeDelayedDelivery"
	policyOpRemoveDispatchRate             = "removeDispatchRate"
	policyOpRemovePublishRate              = "removePublishRate"
	policyOpRemoveInactiveTopicPolicies    = "removeInactiveTopicPolicies"
	policyOpRemoveSubscribeRate            = "removeSubscribeRate"
	policyOpRemoveMaxMessageSize           = "removeMaxMessageSize"
	policyOpRemoveMaxConsumersPerSub       = "removeMaxConsumersPerSubscription"
	policyOpRemoveMaxSubscriptionsPerTopic = "removeMaxSubscriptionsPerTopic"
	policyOpRemoveSchemaValidation         = "removeSchemaValidationEnforced"
	policyOpRemoveSubscriptionDispatch     = "removeSubscriptionDispatchRate"
	policyOpRemoveReplicatorDispatch       = "removeReplicatorDispatchRate"
	policyOpRemoveDedupSnapshotInterval    = "removeDeduplicationSnapshotInterval"
	policyOpRemoveOffloadPolicies          = "removeOffloadPolicies"
	policyOpRemoveAutoSubscription         = "removeAutoSubscriptionCreation"
	policyOpRemoveSchemaCompatibility      = "removeSchemaCompatibilityStrategy"
	policyOpRemoveProperty                 = "removeProperty"
)

type topicPolicyStateReconciler struct {
	base    *reconciler.BaseStatefulReconciler[*resourcev1alpha1.PulsarTopic]
	admin   admin.PulsarAdmin
	log     logr.Logger
	changed bool
}

func newTopicPolicyStateReconciler(logger logr.Logger, adminClient admin.PulsarAdmin) *topicPolicyStateReconciler {
	stateLogger := logger.WithName("PolicyState")
	return &topicPolicyStateReconciler{
		base:  reconciler.NewBaseStatefulReconciler[*resourcev1alpha1.PulsarTopic](stateLogger),
		admin: adminClient,
		log:   stateLogger,
	}
}

func (r *topicPolicyStateReconciler) reconcile(ctx context.Context, topic *resourcev1alpha1.PulsarTopic) (bool, error) {
	r.changed = false
	if err := r.base.Reconcile(ctx, topic, r); err != nil {
		return false, err
	}
	return r.changed, nil
}

// GetStateAnnotationKey implements reconciler.StatefulReconciler.
func (*topicPolicyStateReconciler) GetStateAnnotationKey() string {
	return PulsarTopicPolicyStateAnnotation
}

// ExtractCurrentState implements reconciler.StatefulReconciler.
func (*topicPolicyStateReconciler) ExtractCurrentState(topic *resourcev1alpha1.PulsarTopic) (interface{}, error) {
	state := topicPolicyState{}

	if topic.Spec.CompactionThreshold != nil {
		value := *topic.Spec.CompactionThreshold
		state.CompactionThreshold = &value
	}
	if topic.Spec.MessageTTL != nil {
		state.MessageTTL = true
	}
	if topic.Spec.MaxProducers != nil {
		state.MaxProducers = true
	}
	if topic.Spec.MaxConsumers != nil {
		state.MaxConsumers = true
	}
	if topic.Spec.MaxUnAckedMessagesPerConsumer != nil {
		state.MaxUnAckedMessagesPerConsumer = true
	}
	if topic.Spec.MaxUnAckedMessagesPerSubscription != nil {
		state.MaxUnAckedMessagesPerSubscription = true
	}
	if topic.Spec.RetentionTime != nil || topic.Spec.RetentionSize != nil {
		state.Retention = true
	}
	if (topic.Spec.BacklogQuotaLimitTime != nil || topic.Spec.BacklogQuotaLimitSize != nil) &&
		topic.Spec.BacklogQuotaRetentionPolicy != nil {
		var quotaType string
		if topic.Spec.BacklogQuotaLimitSize != nil {
			quotaType = string(utils.DestinationStorage)
		} else {
			quotaType = string(utils.MessageAge)
		}
		state.BacklogQuotaType = &quotaType
	}
	if topic.Spec.Deduplication != nil {
		state.Deduplication = true
	}
	if topic.Spec.PersistencePolicies != nil {
		state.PersistencePolicies = true
	}
	if topic.Spec.DelayedDelivery != nil {
		state.DelayedDelivery = true
	}
	if topic.Spec.DispatchRate != nil {
		state.DispatchRate = true
	}
	if topic.Spec.PublishRate != nil {
		state.PublishRate = true
	}
	if topic.Spec.InactiveTopicPolicies != nil {
		state.InactiveTopicPolicies = true
	}
	if topic.Spec.SubscribeRate != nil {
		state.SubscribeRate = true
	}
	if topic.Spec.MaxMessageSize != nil {
		state.MaxMessageSize = true
	}
	if topic.Spec.MaxConsumersPerSubscription != nil {
		state.MaxConsumersPerSubscription = true
	}
	if topic.Spec.MaxSubscriptionsPerTopic != nil {
		state.MaxSubscriptionsPerTopic = true
	}
	if topic.Spec.SchemaValidationEnforced != nil {
		state.SchemaValidationEnforced = true
	}
	if topic.Spec.SubscriptionDispatchRate != nil {
		state.SubscriptionDispatchRate = true
	}
	if topic.Spec.ReplicatorDispatchRate != nil {
		state.ReplicatorDispatchRate = true
	}
	if topic.Spec.DeduplicationSnapshotInterval != nil {
		state.DeduplicationSnapshotInterval = true
	}
	if topic.Spec.OffloadPolicies != nil {
		state.OffloadPolicies = true
	}
	if topic.Spec.AutoSubscriptionCreation != nil {
		state.AutoSubscriptionCreation = true
	}
	if topic.Spec.SchemaCompatibilityStrategy != nil {
		state.SchemaCompatibilityStrategy = true
	}
	if len(topic.Spec.Properties) > 0 {
		keys := make([]string, 0, len(topic.Spec.Properties))
		for key := range topic.Spec.Properties {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		state.PropertiesKeys = keys
	}

	return state, nil
}

// CompareStates implements reconciler.StatefulReconciler.
func (r *topicPolicyStateReconciler) CompareStates(previous, current interface{}) (reconciler.StateChangeOperations, error) {
	prevState, err := decodeTopicPolicyState(previous)
	if err != nil {
		return reconciler.StateChangeOperations{}, err
	}
	currState, err := decodeTopicPolicyState(current)
	if err != nil {
		return reconciler.StateChangeOperations{}, err
	}

	ops := reconciler.StateChangeOperations{}

	// Compaction threshold transitions
	switch {
	case prevState.CompactionThreshold == nil && currState.CompactionThreshold != nil:
		r.changed = true
		ops.ItemsToAdd = append(ops.ItemsToAdd, topicPolicyOperation{
			Kind:  policyOpSetCompactionThreshold,
			Value: *currState.CompactionThreshold,
		})
	case prevState.CompactionThreshold != nil && currState.CompactionThreshold != nil &&
		*prevState.CompactionThreshold != *currState.CompactionThreshold:
		r.changed = true
		ops.ItemsToUpdate = append(ops.ItemsToUpdate, topicPolicyOperation{
			Kind:  policyOpSetCompactionThreshold,
			Value: *currState.CompactionThreshold,
		})
	case prevState.CompactionThreshold != nil && currState.CompactionThreshold == nil:
		r.changed = true
		ops.ItemsToRemove = append(ops.ItemsToRemove, topicPolicyOperation{Kind: policyOpRemoveCompactionThreshold})
	}

	// Boolean policy removals
	r.appendPolicyRemoval(prevState.MessageTTL, currState.MessageTTL, policyOpRemoveMessageTTL, &ops)
	r.appendPolicyRemoval(prevState.MaxProducers, currState.MaxProducers, policyOpRemoveMaxProducers, &ops)
	r.appendPolicyRemoval(prevState.MaxConsumers, currState.MaxConsumers, policyOpRemoveMaxConsumers, &ops)
	r.appendPolicyRemoval(prevState.MaxUnAckedMessagesPerConsumer, currState.MaxUnAckedMessagesPerConsumer,
		policyOpRemoveMaxUnackConsumer, &ops)
	r.appendPolicyRemoval(prevState.MaxUnAckedMessagesPerSubscription, currState.MaxUnAckedMessagesPerSubscription,
		policyOpRemoveMaxUnackSubscription, &ops)
	r.appendPolicyRemoval(prevState.Retention, currState.Retention, policyOpRemoveRetention, &ops)
	r.appendPolicyRemoval(prevState.Deduplication, currState.Deduplication, policyOpRemoveDeduplication, &ops)
	r.appendPolicyRemoval(prevState.PersistencePolicies, currState.PersistencePolicies, policyOpRemovePersistence, &ops)
	r.appendPolicyRemoval(prevState.DelayedDelivery, currState.DelayedDelivery, policyOpRemoveDelayedDelivery, &ops)
	r.appendPolicyRemoval(prevState.DispatchRate, currState.DispatchRate, policyOpRemoveDispatchRate, &ops)
	r.appendPolicyRemoval(prevState.PublishRate, currState.PublishRate, policyOpRemovePublishRate, &ops)
	r.appendPolicyRemoval(prevState.InactiveTopicPolicies, currState.InactiveTopicPolicies,
		policyOpRemoveInactiveTopicPolicies, &ops)
	r.appendPolicyRemoval(prevState.SubscribeRate, currState.SubscribeRate, policyOpRemoveSubscribeRate, &ops)
	r.appendPolicyRemoval(prevState.MaxMessageSize, currState.MaxMessageSize, policyOpRemoveMaxMessageSize, &ops)
	r.appendPolicyRemoval(prevState.MaxConsumersPerSubscription, currState.MaxConsumersPerSubscription,
		policyOpRemoveMaxConsumersPerSub, &ops)
	r.appendPolicyRemoval(prevState.MaxSubscriptionsPerTopic, currState.MaxSubscriptionsPerTopic,
		policyOpRemoveMaxSubscriptionsPerTopic, &ops)
	r.appendPolicyRemoval(prevState.SchemaValidationEnforced, currState.SchemaValidationEnforced,
		policyOpRemoveSchemaValidation, &ops)
	r.appendPolicyRemoval(prevState.SubscriptionDispatchRate, currState.SubscriptionDispatchRate,
		policyOpRemoveSubscriptionDispatch, &ops)
	r.appendPolicyRemoval(prevState.ReplicatorDispatchRate, currState.ReplicatorDispatchRate,
		policyOpRemoveReplicatorDispatch, &ops)
	r.appendPolicyRemoval(prevState.DeduplicationSnapshotInterval, currState.DeduplicationSnapshotInterval,
		policyOpRemoveDedupSnapshotInterval, &ops)
	r.appendPolicyRemoval(prevState.OffloadPolicies, currState.OffloadPolicies, policyOpRemoveOffloadPolicies, &ops)
	r.appendPolicyRemoval(prevState.AutoSubscriptionCreation, currState.AutoSubscriptionCreation,
		policyOpRemoveAutoSubscription, &ops)
	r.appendPolicyRemoval(prevState.SchemaCompatibilityStrategy, currState.SchemaCompatibilityStrategy,
		policyOpRemoveSchemaCompatibility, &ops)

	// Backlog quota
	if prevState.BacklogQuotaType != nil {
		if currState.BacklogQuotaType == nil || *prevState.BacklogQuotaType != *currState.BacklogQuotaType {
			r.changed = true
			ops.ItemsToRemove = append(ops.ItemsToRemove, topicPolicyOperation{
				Kind:  policyOpRemoveBacklogQuota,
				Value: *prevState.BacklogQuotaType,
			})
		}
	}

	// Topic properties
	if len(prevState.PropertiesKeys) > 0 {
		currKeys := make(map[string]struct{}, len(currState.PropertiesKeys))
		for _, key := range currState.PropertiesKeys {
			currKeys[key] = struct{}{}
		}
		for _, key := range prevState.PropertiesKeys {
			if _, exists := currKeys[key]; !exists {
				r.changed = true
				ops.ItemsToRemove = append(ops.ItemsToRemove, topicPolicyOperation{
					Kind:  policyOpRemoveProperty,
					Value: key,
				})
			}
		}
	}

	return ops, nil
}

// ApplyOperations implements reconciler.StatefulReconciler.
func (r *topicPolicyStateReconciler) ApplyOperations(ctx context.Context, topic *resourcev1alpha1.PulsarTopic,
	ops reconciler.StateChangeOperations) error {
	for _, item := range append(ops.ItemsToAdd, ops.ItemsToUpdate...) {
		op, err := normalizeTopicPolicyOperation(item)
		if err != nil {
			return err
		}
		switch op.Kind {
		case policyOpSetCompactionThreshold:
			value, err := toInt64(op.Value)
			if err != nil {
				return err
			}
			r.log.V(1).Info("Setting topic compaction threshold",
				"topicSpecName", topic.Spec.Name,
				"threshold", value)
			if err := r.admin.SetTopicCompactionThreshold(topic.Spec.Name, topic.Spec.Persistent, value); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported topic policy operation kind %q in add/update phase", op.Kind)
		}
	}

	for _, item := range ops.ItemsToRemove {
		op, err := normalizeTopicPolicyOperation(item)
		if err != nil {
			return err
		}
		switch op.Kind {
		case policyOpRemoveCompactionThreshold:
			r.log.V(1).Info("Removing topic compaction threshold", "topicSpecName", topic.Spec.Name)
			if err := r.admin.RemoveTopicCompactionThreshold(topic.Spec.Name, topic.Spec.Persistent); err != nil {
				return err
			}
		case policyOpRemoveMessageTTL:
			r.log.V(1).Info("Removing topic message TTL", "topicSpecName", topic.Spec.Name)
			if err := r.admin.RemoveTopicMessageTTL(topic.Spec.Name, topic.Spec.Persistent); err != nil {
				return err
			}
		case policyOpRemoveMaxProducers:
			r.log.V(1).Info("Removing topic max producers", "topicSpecName", topic.Spec.Name)
			if err := r.admin.RemoveTopicMaxProducers(topic.Spec.Name, topic.Spec.Persistent); err != nil {
				return err
			}
		case policyOpRemoveMaxConsumers:
			r.log.V(1).Info("Removing topic max consumers", "topicSpecName", topic.Spec.Name)
			if err := r.admin.RemoveTopicMaxConsumers(topic.Spec.Name, topic.Spec.Persistent); err != nil {
				return err
			}
		case policyOpRemoveMaxUnackConsumer:
			r.log.V(1).Info("Removing topic max unacked messages per consumer", "topicSpecName", topic.Spec.Name)
			if err := r.admin.RemoveTopicMaxUnackedMessagesPerConsumer(topic.Spec.Name, topic.Spec.Persistent); err != nil {
				return err
			}
		case policyOpRemoveMaxUnackSubscription:
			r.log.V(1).Info("Removing topic max unacked messages per subscription", "topicSpecName", topic.Spec.Name)
			if err := r.admin.RemoveTopicMaxUnackedMessagesPerSubscription(topic.Spec.Name, topic.Spec.Persistent); err != nil {
				return err
			}
		case policyOpRemoveRetention:
			r.log.V(1).Info("Removing topic retention policy", "topicSpecName", topic.Spec.Name)
			if err := r.admin.RemoveTopicRetention(topic.Spec.Name, topic.Spec.Persistent); err != nil {
				return err
			}
		case policyOpRemoveBacklogQuota:
			quotaType, err := toString(op.Value)
			if err != nil {
				return err
			}
			r.log.V(1).Info("Removing topic backlog quota", "topicSpecName", topic.Spec.Name, "type", quotaType)
			if err := r.admin.RemoveTopicBacklogQuota(topic.Spec.Name, topic.Spec.Persistent, quotaType); err != nil {
				return err
			}
		case policyOpRemoveDeduplication:
			r.log.V(1).Info("Removing topic deduplication status", "topicSpecName", topic.Spec.Name)
			if err := r.admin.RemoveTopicDeduplicationStatus(topic.Spec.Name, topic.Spec.Persistent); err != nil {
				return err
			}
		case policyOpRemovePersistence:
			r.log.V(1).Info("Removing topic persistence policies", "topicSpecName", topic.Spec.Name)
			if err := r.admin.RemoveTopicPersistence(topic.Spec.Name, topic.Spec.Persistent); err != nil {
				return err
			}
		case policyOpRemoveDelayedDelivery:
			r.log.V(1).Info("Removing topic delayed delivery policy", "topicSpecName", topic.Spec.Name)
			if err := r.admin.RemoveTopicDelayedDelivery(topic.Spec.Name, topic.Spec.Persistent); err != nil {
				return err
			}
		case policyOpRemoveDispatchRate:
			r.log.V(1).Info("Removing topic dispatch rate", "topicSpecName", topic.Spec.Name)
			if err := r.admin.RemoveTopicDispatchRate(topic.Spec.Name, topic.Spec.Persistent); err != nil {
				return err
			}
		case policyOpRemovePublishRate:
			r.log.V(1).Info("Removing topic publish rate", "topicSpecName", topic.Spec.Name)
			if err := r.admin.RemoveTopicPublishRate(topic.Spec.Name, topic.Spec.Persistent); err != nil {
				return err
			}
		case policyOpRemoveInactiveTopicPolicies:
			r.log.V(1).Info("Removing inactive topic policies", "topicSpecName", topic.Spec.Name)
			if err := r.admin.RemoveTopicInactiveTopicPolicies(topic.Spec.Name, topic.Spec.Persistent); err != nil {
				return err
			}
		case policyOpRemoveSubscribeRate:
			r.log.V(1).Info("Removing topic subscribe rate", "topicSpecName", topic.Spec.Name)
			if err := r.admin.RemoveTopicSubscribeRate(topic.Spec.Name, topic.Spec.Persistent); err != nil {
				return err
			}
		case policyOpRemoveMaxMessageSize:
			r.log.V(1).Info("Removing topic max message size", "topicSpecName", topic.Spec.Name)
			if err := r.admin.RemoveTopicMaxMessageSize(topic.Spec.Name, topic.Spec.Persistent); err != nil {
				return err
			}
		case policyOpRemoveMaxConsumersPerSub:
			r.log.V(1).Info("Removing topic max consumers per subscription", "topicSpecName", topic.Spec.Name)
			if err := r.admin.RemoveTopicMaxConsumersPerSubscription(topic.Spec.Name, topic.Spec.Persistent); err != nil {
				return err
			}
		case policyOpRemoveMaxSubscriptionsPerTopic:
			r.log.V(1).Info("Removing topic max subscriptions per topic", "topicSpecName", topic.Spec.Name)
			if err := r.admin.RemoveTopicMaxSubscriptionsPerTopic(topic.Spec.Name, topic.Spec.Persistent); err != nil {
				return err
			}
		case policyOpRemoveSchemaValidation:
			r.log.V(1).Info("Removing topic schema validation enforced flag", "topicSpecName", topic.Spec.Name)
			if err := r.admin.RemoveTopicSchemaValidationEnforced(topic.Spec.Name, topic.Spec.Persistent); err != nil {
				return err
			}
		case policyOpRemoveSubscriptionDispatch:
			r.log.V(1).Info("Removing topic subscription dispatch rate", "topicSpecName", topic.Spec.Name)
			if err := r.admin.RemoveTopicSubscriptionDispatchRate(topic.Spec.Name, topic.Spec.Persistent); err != nil {
				return err
			}
		case policyOpRemoveReplicatorDispatch:
			r.log.V(1).Info("Removing topic replicator dispatch rate", "topicSpecName", topic.Spec.Name)
			if err := r.admin.RemoveTopicReplicatorDispatchRate(topic.Spec.Name, topic.Spec.Persistent); err != nil {
				return err
			}
		case policyOpRemoveDedupSnapshotInterval:
			r.log.V(1).Info("Removing topic deduplication snapshot interval", "topicSpecName", topic.Spec.Name)
			if err := r.admin.RemoveTopicDeduplicationSnapshotInterval(topic.Spec.Name, topic.Spec.Persistent); err != nil {
				return err
			}
		case policyOpRemoveOffloadPolicies:
			r.log.V(1).Info("Removing topic offload policies", "topicSpecName", topic.Spec.Name)
			if err := r.admin.RemoveTopicOffloadPolicies(topic.Spec.Name, topic.Spec.Persistent); err != nil {
				return err
			}
		case policyOpRemoveAutoSubscription:
			r.log.V(1).Info("Removing topic auto subscription creation override", "topicSpecName", topic.Spec.Name)
			if err := r.admin.RemoveTopicAutoSubscriptionCreation(topic.Spec.Name, topic.Spec.Persistent); err != nil {
				return err
			}
		case policyOpRemoveSchemaCompatibility:
			r.log.V(1).Info("Removing topic schema compatibility strategy override", "topicSpecName", topic.Spec.Name)
			if err := r.admin.RemoveTopicSchemaCompatibilityStrategy(topic.Spec.Name, topic.Spec.Persistent); err != nil {
				return err
			}
		case policyOpRemoveProperty:
			key, err := toString(op.Value)
			if err != nil {
				return err
			}
			r.log.V(1).Info("Removing topic property", "topicSpecName", topic.Spec.Name, "key", key)
			if err := r.admin.RemoveTopicProperty(topic.Spec.Name, topic.Spec.Persistent, key); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported topic policy operation kind %q in removal phase", op.Kind)
		}
	}

	return nil
}

// UpdateStateAnnotation implements reconciler.StatefulReconciler.
func (r *topicPolicyStateReconciler) UpdateStateAnnotation(topic *resourcev1alpha1.PulsarTopic, currentState interface{}) error {
	return r.base.UpdateAnnotation(topic, PulsarTopicPolicyStateAnnotation, currentState)
}

func (r *topicPolicyStateReconciler) appendPolicyRemoval(previous, current bool, kind string,
	ops *reconciler.StateChangeOperations) {
	if previous && !current {
		r.changed = true
		ops.ItemsToRemove = append(ops.ItemsToRemove, topicPolicyOperation{Kind: kind})
	}
}

func decodeTopicPolicyState(state interface{}) (topicPolicyState, error) {
	if state == nil {
		return topicPolicyState{}, nil
	}

	switch value := state.(type) {
	case topicPolicyState:
		return value, nil
	case map[string]interface{}:
		bytes, err := json.Marshal(value)
		if err != nil {
			return topicPolicyState{}, err
		}
		var result topicPolicyState
		if err := json.Unmarshal(bytes, &result); err != nil {
			return topicPolicyState{}, err
		}
		return result, nil
	default:
		bytes, err := json.Marshal(value)
		if err != nil {
			return topicPolicyState{}, err
		}
		var result topicPolicyState
		if err := json.Unmarshal(bytes, &result); err != nil {
			return topicPolicyState{}, err
		}
		return result, nil
	}
}

func normalizeTopicPolicyOperation(item interface{}) (topicPolicyOperation, error) {
	switch value := item.(type) {
	case topicPolicyOperation:
		return value, nil
	case map[string]interface{}:
		bytes, err := json.Marshal(value)
		if err != nil {
			return topicPolicyOperation{}, err
		}
		var op topicPolicyOperation
		if err := json.Unmarshal(bytes, &op); err != nil {
			return topicPolicyOperation{}, err
		}
		return op, nil
	default:
		return topicPolicyOperation{}, fmt.Errorf("unexpected topic policy operation type %T", item)
	}
}

func toInt64(value interface{}) (int64, error) {
	switch typed := value.(type) {
	case int64:
		return typed, nil
	case int32:
		return int64(typed), nil
	case int:
		return int64(typed), nil
	case float64:
		return int64(typed), nil
	case json.Number:
		return typed.Int64()
	default:
		return 0, fmt.Errorf("unsupported int64 value type %T", value)
	}
}

func toString(value interface{}) (string, error) {
	switch typed := value.(type) {
	case string:
		return typed, nil
	case json.Number:
		return typed.String(), nil
	default:
		return "", fmt.Errorf("unsupported string value type %T", value)
	}
}
