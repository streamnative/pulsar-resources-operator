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

package reconciler

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// StatefulResource represents a Kubernetes resource that needs stateful reconciliation
type StatefulResource interface {
	runtime.Object
	metav1.Object
}

// StatefulReconciler provides a framework for managing resources that need to track
// their previous state to perform proper cleanup operations during reconciliation.
// It uses annotations to store the previous state and compares it with the current
// desired state to determine what changes need to be made.
type StatefulReconciler[T StatefulResource] interface {
	// GetStateAnnotationKey returns the annotation key used to store the previous state
	GetStateAnnotationKey() string

	// ExtractCurrentState extracts the current desired state from the resource spec
	// that needs to be compared with the previous state for change detection
	ExtractCurrentState(resource T) (interface{}, error)

	// CompareStates compares the previous state with the current state and returns
	// the operations that need to be performed to reconcile the differences
	CompareStates(previous, current interface{}) (StateChangeOperations, error)

	// ApplyOperations applies the state change operations to the external system
	ApplyOperations(ctx context.Context, resource T, ops StateChangeOperations) error

	// UpdateStateAnnotation updates the annotation with the current state after
	// successful reconciliation
	UpdateStateAnnotation(resource T, currentState interface{}) error
}

// StateChangeOperations represents the operations needed to reconcile state changes
type StateChangeOperations struct {
	// ItemsToAdd represents items that need to be added to the external system
	ItemsToAdd []interface{} `json:"itemsToAdd,omitempty"`

	// ItemsToRemove represents items that need to be removed from the external system
	ItemsToRemove []interface{} `json:"itemsToRemove,omitempty"`

	// ItemsToUpdate represents items that need to be updated in the external system
	ItemsToUpdate []interface{} `json:"itemsToUpdate,omitempty"`

	// ContextChanged indicates that the context (like resource type or name) has changed
	// and requires special handling like cleanup of old context
	ContextChanged bool `json:"contextChanged,omitempty"`

	// PreviousContext contains the previous context information for cleanup
	PreviousContext interface{} `json:"previousContext,omitempty"`
}

// BaseStatefulReconciler provides a base implementation of StatefulReconciler
// that handles the common annotation-based state tracking logic
type BaseStatefulReconciler[T StatefulResource] struct {
	Logger logr.Logger
}

// NewBaseStatefulReconciler creates a new BaseStatefulReconciler instance
func NewBaseStatefulReconciler[T StatefulResource](logger logr.Logger) *BaseStatefulReconciler[T] {
	return &BaseStatefulReconciler[T]{
		Logger: logger,
	}
}

// Reconcile performs the complete stateful reconciliation workflow:
// 1. Extract current desired state from resource
// 2. Retrieve previous state from annotation
// 3. Compare states to determine required operations
// 4. Apply operations to external system
// 5. Update state annotation on success
func (r *BaseStatefulReconciler[T]) Reconcile(ctx context.Context, resource T, impl StatefulReconciler[T]) error {
	// Extract current desired state
	currentState, err := impl.ExtractCurrentState(resource)
	if err != nil {
		r.Logger.Error(err, "Failed to extract current state from resource")
		return err
	}

	// Get previous state from annotation
	previousState, err := r.getPreviousState(resource, impl.GetStateAnnotationKey())
	if err != nil {
		r.Logger.Error(err, "Failed to get previous state from annotation")
		return err
	}

	// Compare states to determine operations
	ops, err := impl.CompareStates(previousState, currentState)
	if err != nil {
		r.Logger.Error(err, "Failed to compare states")
		return err
	}

	// Log the operations that will be performed
	if len(ops.ItemsToAdd) > 0 || len(ops.ItemsToRemove) > 0 || len(ops.ItemsToUpdate) > 0 || ops.ContextChanged {
		r.Logger.Info("State changes detected, applying operations",
			"itemsToAdd", len(ops.ItemsToAdd),
			"itemsToRemove", len(ops.ItemsToRemove),
			"itemsToUpdate", len(ops.ItemsToUpdate),
			"contextChanged", ops.ContextChanged)
	} else {
		r.Logger.V(1).Info("No state changes detected, skipping operations")
		return nil
	}

	// Apply operations to external system
	if err := impl.ApplyOperations(ctx, resource, ops); err != nil {
		r.Logger.Error(err, "Failed to apply state change operations")
		return err
	}

	// Update state annotation after successful operations
	if err := impl.UpdateStateAnnotation(resource, currentState); err != nil {
		r.Logger.Error(err, "Failed to update state annotation")
		return err
	}

	r.Logger.Info("Successfully completed stateful reconciliation")
	return nil
}

// getPreviousState retrieves the previous state from the resource annotation
func (r *BaseStatefulReconciler[T]) getPreviousState(resource T, annotationKey string) (interface{}, error) {
	annotations := resource.GetAnnotations()
	if annotations == nil {
		r.Logger.V(1).Info("No annotations found, treating as first reconciliation")
		return nil, nil
	}

	stateJSON, exists := annotations[annotationKey]
	if !exists {
		r.Logger.V(1).Info("No previous state annotation found, treating as first reconciliation")
		return nil, nil
	}

	// Try to unmarshal as generic interface{}
	var previousState interface{}
	if err := json.Unmarshal([]byte(stateJSON), &previousState); err != nil {
		r.Logger.Error(err, "Failed to unmarshal previous state annotation, treating as first reconciliation",
			"annotation", stateJSON)
		return nil, nil
	}

	return previousState, nil
}

// UpdateAnnotation is a helper method to update the state annotation
func (r *BaseStatefulReconciler[T]) UpdateAnnotation(resource T, annotationKey string, state interface{}) error {
	stateJSON, err := json.Marshal(state)
	if err != nil {
		return err
	}

	annotations := resource.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
		resource.SetAnnotations(annotations)
	}

	// Only update if the value has changed
	currentValue := annotations[annotationKey]
	newValue := string(stateJSON)

	if currentValue != newValue {
		annotations[annotationKey] = newValue
		r.Logger.V(1).Info("Updated state annotation", "key", annotationKey)
		return nil
	}

	r.Logger.V(2).Info("State annotation unchanged, skipping update", "key", annotationKey)
	return nil
}

// HasStateChanged is a utility method to check if two states are different
func (r *BaseStatefulReconciler[T]) HasStateChanged(previous, current interface{}) bool {
	return !reflect.DeepEqual(previous, current)
}