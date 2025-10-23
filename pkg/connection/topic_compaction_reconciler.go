package connection

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/go-logr/logr"

	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	"github.com/streamnative/pulsar-resources-operator/pkg/admin"
	"github.com/streamnative/pulsar-resources-operator/pkg/reconciler"
)

const (
	// PulsarTopicCompactionStateAnnotation stores the last applied compaction threshold state
	PulsarTopicCompactionStateAnnotation = "pulsartopics.resource.streamnative.io/compaction-state"
)

// topicCompactionState represents the compaction policy state we track on PulsarTopic resources.
type topicCompactionState struct {
	CompactionThreshold *int64 `json:"compactionThreshold,omitempty"`
}

// topicCompactionStateReconciler reconciles compaction threshold changes using the stateful reconciler helper.
type topicCompactionStateReconciler struct {
	base    *reconciler.BaseStatefulReconciler[*resourcev1alpha1.PulsarTopic]
	admin   admin.PulsarAdmin
	log     logr.Logger
	changed bool
}

func newTopicCompactionStateReconciler(logger logr.Logger, adminClient admin.PulsarAdmin) *topicCompactionStateReconciler {
	compactionLogger := logger.WithName("CompactionState")
	return &topicCompactionStateReconciler{
		base:  reconciler.NewBaseStatefulReconciler[*resourcev1alpha1.PulsarTopic](compactionLogger),
		admin: adminClient,
		log:   compactionLogger,
	}
}

func (r *topicCompactionStateReconciler) reconcile(ctx context.Context, topic *resourcev1alpha1.PulsarTopic) (bool, error) {
	r.changed = false
	if err := r.base.Reconcile(ctx, topic, r); err != nil {
		return false, err
	}
	return r.changed, nil
}

// GetStateAnnotationKey implements reconciler.StatefulReconciler.
func (*topicCompactionStateReconciler) GetStateAnnotationKey() string {
	return PulsarTopicCompactionStateAnnotation
}

// ExtractCurrentState implements reconciler.StatefulReconciler.
func (*topicCompactionStateReconciler) ExtractCurrentState(topic *resourcev1alpha1.PulsarTopic) (interface{}, error) {
	state := topicCompactionState{}
	if topic.Spec.CompactionThreshold != nil {
		value := *topic.Spec.CompactionThreshold
		state.CompactionThreshold = &value
	}
	return state, nil
}

// CompareStates implements reconciler.StatefulReconciler.
func (r *topicCompactionStateReconciler) CompareStates(previous, current interface{}) (reconciler.StateChangeOperations, error) {
	prevState, err := decodeTopicCompactionState(previous)
	if err != nil {
		return reconciler.StateChangeOperations{}, err
	}
	currState, err := decodeTopicCompactionState(current)
	if err != nil {
		return reconciler.StateChangeOperations{}, err
	}

	prevThreshold := prevState.CompactionThreshold
	currThreshold := currState.CompactionThreshold

	ops := reconciler.StateChangeOperations{}

	switch {
	case prevThreshold == nil && currThreshold != nil:
		ops.ItemsToAdd = append(ops.ItemsToAdd, *currThreshold)
	case prevThreshold != nil && currThreshold == nil:
		ops.ItemsToRemove = append(ops.ItemsToRemove, *prevThreshold)
	case prevThreshold != nil && currThreshold != nil && *prevThreshold != *currThreshold:
		ops.ItemsToUpdate = append(ops.ItemsToUpdate, *currThreshold)
	default:
		return ops, nil
	}

	r.changed = true
	return ops, nil
}

// ApplyOperations implements reconciler.StatefulReconciler.
func (r *topicCompactionStateReconciler) ApplyOperations(ctx context.Context, topic *resourcev1alpha1.PulsarTopic,
	ops reconciler.StateChangeOperations) error {
	for _, item := range ops.ItemsToAdd {
		value, ok := item.(int64)
		if !ok {
			return fmt.Errorf("unexpected compaction threshold type %T in add operation", item)
		}
		r.log.V(1).Info("Setting topic compaction threshold",
			"topicSpecName", topic.Spec.Name,
			"threshold", value)
		if err := r.admin.SetTopicCompactionThreshold(topic.Spec.Name, topic.Spec.Persistent, value); err != nil {
			return err
		}
	}

	for _, item := range ops.ItemsToUpdate {
		value, ok := item.(int64)
		if !ok {
			return fmt.Errorf("unexpected compaction threshold type %T in update operation", item)
		}
		r.log.V(1).Info("Updating topic compaction threshold",
			"topicSpecName", topic.Spec.Name,
			"threshold", value)
		if err := r.admin.SetTopicCompactionThreshold(topic.Spec.Name, topic.Spec.Persistent, value); err != nil {
			return err
		}
	}

	if len(ops.ItemsToRemove) > 0 {
		var previous interface{}
		if len(ops.ItemsToRemove) == 1 {
			previous = ops.ItemsToRemove[0]
		} else {
			previous = ops.ItemsToRemove
		}
		r.log.V(1).Info("Removing topic compaction threshold",
			"topicSpecName", topic.Spec.Name,
			"previous", previous)
		if err := r.admin.RemoveTopicCompactionThreshold(topic.Spec.Name, topic.Spec.Persistent); err != nil {
			return err
		}
	}

	return nil
}

// UpdateStateAnnotation implements reconciler.StatefulReconciler.
func (r *topicCompactionStateReconciler) UpdateStateAnnotation(topic *resourcev1alpha1.PulsarTopic, currentState interface{}) error {
	return r.base.UpdateAnnotation(topic, PulsarTopicCompactionStateAnnotation, currentState)
}

func decodeTopicCompactionState(state interface{}) (topicCompactionState, error) {
	if state == nil {
		return topicCompactionState{}, nil
	}

	switch value := state.(type) {
	case topicCompactionState:
		return value, nil
	case map[string]interface{}:
		return decodeTopicCompactionStateFromMap(value)
	default:
		bytes, err := json.Marshal(value)
		if err != nil {
			return topicCompactionState{}, err
		}
		var result topicCompactionState
		if err := json.Unmarshal(bytes, &result); err != nil {
			return topicCompactionState{}, err
		}
		return result, nil
	}
}

func decodeTopicCompactionStateFromMap(raw map[string]interface{}) (topicCompactionState, error) {
	result := topicCompactionState{}
	if raw == nil {
		return result, nil
	}

	value, exists := raw["compactionThreshold"]
	if !exists || value == nil {
		return result, nil
	}

	switch typed := value.(type) {
	case float64:
		threshold := int64(typed)
		result.CompactionThreshold = &threshold
	case json.Number:
		parsed, err := strconv.ParseInt(string(typed), 10, 64)
		if err != nil {
			return topicCompactionState{}, err
		}
		result.CompactionThreshold = &parsed
	case int64:
		threshold := typed
		result.CompactionThreshold = &threshold
	case int:
		threshold := int64(typed)
		result.CompactionThreshold = &threshold
	default:
		return topicCompactionState{}, fmt.Errorf("unsupported compaction threshold value type %T", value)
	}

	return result, nil
}
