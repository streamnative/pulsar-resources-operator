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

package controllers

import (
	"context"
	"fmt"
	"sync"
	"time"

	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	controllers2 "github.com/streamnative/pulsar-resources-operator/pkg/streamnativecloud"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	cloudapi "github.com/streamnative/pulsar-resources-operator/pkg/streamnativecloud/apis/cloud/v1alpha1"
)

// APIKeyReconciler reconciles a StreamNative Cloud APIKey object
type APIKeyReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	ConnectionManager *ConnectionManager
	// watcherMap stores active watchers for APIKeys
	watcherMap map[types.NamespacedName]watch.Interface
	// watcherMutex protects watcherMap
	watcherMutex sync.RWMutex
}

const APIKeyFinalizer = "apikey.resource.streamnative.io/finalizer"

//+kubebuilder:rbac:groups=resource.streamnative.io,resources=apikeys,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=apikeys/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=apikeys/finalizers,verbs=update
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=streamnativecloudconnections,verbs=get;list;watch

// handleWatchEvents processes events from the watch interface
func (r *APIKeyReconciler) handleWatchEvents(ctx context.Context, namespacedName types.NamespacedName, watcher watch.Interface) {
	logger := log.FromContext(ctx)
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-watcher.ResultChan():
			if !ok {
				logger.Info("Watch channel closed", "namespace", namespacedName.Namespace, "name", namespacedName.Name)
				// Remove the watcher from the map
				r.watcherMutex.Lock()
				delete(r.watcherMap, namespacedName)
				r.watcherMutex.Unlock()
				return
			}

			if event.Type == watch.Modified {
				// Check if the object is an APIKey
				_, ok := event.Object.(*cloudapi.APIKey)
				if !ok {
					logger.Error(fmt.Errorf("unexpected object type"), "Failed to convert object to APIKey")
					continue
				}

				// Get the local APIKey
				localAPIKey := &resourcev1alpha1.APIKey{}
				if err := r.Get(ctx, namespacedName, localAPIKey); err != nil {
					logger.Error(err, "Failed to get local APIKey")
					continue
				}

				// Update status
				r.updateAPIKeyStatus(ctx, localAPIKey, nil, "Ready", "APIKey synced successfully")
			}
		}
	}
}

// setupWatch creates a new watcher for an APIKey
func (r *APIKeyReconciler) setupWatch(ctx context.Context, apiKey *resourcev1alpha1.APIKey, apiKeyClient *controllers2.APIKeyClient) error {
	namespacedName := types.NamespacedName{
		Namespace: apiKey.Namespace,
		Name:      apiKey.Name,
	}

	// Check if we already have a watcher
	r.watcherMutex.RLock()
	_, exists := r.watcherMap[namespacedName]
	r.watcherMutex.RUnlock()
	if exists {
		return nil
	}

	// Create new watcher
	watcher, err := apiKeyClient.WatchAPIKey(ctx, apiKey.Name)
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}

	// Store watcher in map
	r.watcherMutex.Lock()
	r.watcherMap[namespacedName] = watcher
	r.watcherMutex.Unlock()

	// Start watching in a new goroutine
	go r.handleWatchEvents(ctx, namespacedName, watcher)
	return nil
}

// Reconcile handles the reconciliation of APIKey objects
func (r *APIKeyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling APIKey", "namespace", req.Namespace, "name", req.Name)

	// Add requeue interval for status sync
	requeueInterval := time.Minute

	// Get the APIKey resource
	apiKey := &resourcev1alpha1.APIKey{}
	if err := r.Get(ctx, req.NamespacedName, apiKey); err != nil {
		if apierrors.IsNotFound(err) {
			// Stop and remove watcher if it exists
			r.watcherMutex.Lock()
			if watcher, exists := r.watcherMap[req.NamespacedName]; exists {
				watcher.Stop()
				delete(r.watcherMap, req.NamespacedName)
			}
			r.watcherMutex.Unlock()
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Get the APIServerConnection
	connection := &resourcev1alpha1.StreamNativeCloudConnection{}
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: req.Namespace,
		Name:      apiKey.Spec.APIServerRef.Name,
	}, connection); err != nil {
		r.updateAPIKeyStatus(ctx, apiKey, err, "ConnectionNotFound",
			fmt.Sprintf("Failed to get APIServerConnection: %v", err))
		return ctrl.Result{}, err
	}

	// Get API connection
	apiConn, err := r.ConnectionManager.GetOrCreateConnection(connection, nil)
	if err != nil {
		// If connection is not initialized, requeue the request
		if _, ok := err.(*NotInitializedError); ok {
			logger.Info("Connection not initialized, requeueing", "error", err.Error())
			return ctrl.Result{Requeue: true}, nil
		}
		r.updateAPIKeyStatus(ctx, apiKey, err, "GetConnectionFailed",
			fmt.Sprintf("Failed to get connection: %v", err))
		return ctrl.Result{}, err
	}

	// Get organization from connection
	organization := connection.Spec.Organization
	if organization == "" {
		err := fmt.Errorf("organization is required but not specified")
		r.updateAPIKeyStatus(ctx, apiKey, err, "ValidationFailed", err.Error())
		return ctrl.Result{}, err
	}

	// Create APIKey client
	apiKeyClient, err := controllers2.NewAPIKeyClient(apiConn, organization)
	if err != nil {
		r.updateAPIKeyStatus(ctx, apiKey, err, "ClientCreationFailed",
			fmt.Sprintf("Failed to create APIKey client: %v", err))
		return ctrl.Result{}, err
	}

	// Handle deletion
	if !apiKey.DeletionTimestamp.IsZero() {
		if controllers2.ContainsString(apiKey.Finalizers, APIKeyFinalizer) {
			// Try to delete remote APIKey
			if err := apiKeyClient.DeleteAPIKey(ctx, apiKey); err != nil {
				if !apierrors.IsNotFound(err) {
					r.updateAPIKeyStatus(ctx, apiKey, err, "DeleteFailed",
						fmt.Sprintf("Failed to delete external resources: %v", err))
					return ctrl.Result{}, err
				}
				// If the resource is already gone, that's fine
				logger.Info("Remote APIKey already deleted or not found",
					"apiKey", apiKey.Name)
			}

			// Remove finalizer after successful deletion
			apiKey.Finalizers = controllers2.RemoveString(apiKey.Finalizers, APIKeyFinalizer)
			if err := r.Update(ctx, apiKey); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer if it doesn't exist
	if !controllers2.ContainsString(apiKey.Finalizers, APIKeyFinalizer) {
		apiKey.Finalizers = append(apiKey.Finalizers, APIKeyFinalizer)
		if err := r.Update(ctx, apiKey); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Check if ApiKey exists
	existingAPIKey, err := apiKeyClient.GetAPIKey(ctx, apiKey.Name)
	if err != nil {
		logger.Info("Failed to get APIKey", "error", err, "existingAPIKey", existingAPIKey)
		if !apierrors.IsNotFound(err) {
			r.updateAPIKeyStatus(ctx, apiKey, err, "GetAPIKeyFailed",
				fmt.Sprintf("Failed to get APIKey: %v", err))
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		existingAPIKey = nil
	}

	if existingAPIKey == nil {
		// Create APIKey
		resultAPIKey, err := apiKeyClient.CreateAPIKey(ctx, apiKey)
		if err != nil {
			r.updateAPIKeyStatus(ctx, apiKey, err, "CreateAPIKeyFailed",
				fmt.Sprintf("Failed to create APIKey: %v", err))
			return ctrl.Result{}, err
		}

		// Update status with token information
		if resultAPIKey.Status.Token != nil {
			apiKey.Status.Token = resultAPIKey.Status.Token
		}
		if resultAPIKey.Status.KeyID != nil {
			apiKey.Status.KeyID = resultAPIKey.Status.KeyID
		}
		if resultAPIKey.Status.IssuedAt != nil {
			apiKey.Status.IssuedAt = resultAPIKey.Status.IssuedAt
		}
		if resultAPIKey.Status.ExpiresAt != nil {
			apiKey.Status.ExpiresAt = resultAPIKey.Status.ExpiresAt
		}
		if resultAPIKey.Status.EncryptedToken != nil {
			apiKey.Status.EncryptedToken = &resourcev1alpha1.EncryptedToken{
				JWE: resultAPIKey.Status.EncryptedToken.JWE,
			}
		}

		// Update status
		r.updateAPIKeyStatus(ctx, apiKey, nil, "Ready", "APIKey created successfully")
	} else {
		// Update APIKey
		resultAPIKey, err := apiKeyClient.UpdateAPIKey(ctx, apiKey)
		if err != nil {
			r.updateAPIKeyStatus(ctx, apiKey, err, "UpdateAPIKeyFailed",
				fmt.Sprintf("Failed to update APIKey: %v", err))
			return ctrl.Result{}, err
		}

		// Update status with token information
		if resultAPIKey.Status.KeyID != nil {
			apiKey.Status.KeyID = resultAPIKey.Status.KeyID
		}
		if resultAPIKey.Status.IssuedAt != nil {
			apiKey.Status.IssuedAt = resultAPIKey.Status.IssuedAt
		}
		if resultAPIKey.Status.ExpiresAt != nil {
			apiKey.Status.ExpiresAt = resultAPIKey.Status.ExpiresAt
		}
		if resultAPIKey.Status.RevokedAt != nil {
			apiKey.Status.RevokedAt = resultAPIKey.Status.RevokedAt
		}

		// Update status
		r.updateAPIKeyStatus(ctx, apiKey, nil, "Ready", "APIKey updated successfully")
	}

	// Setup watch after APIKey is created/updated
	if err := r.setupWatch(ctx, apiKey, apiKeyClient); err != nil {
		logger.Error(err, "Failed to setup watch")
		// Don't return error, just log it
	}

	return ctrl.Result{RequeueAfter: requeueInterval}, nil
}

func (r *APIKeyReconciler) updateAPIKeyStatus(
	ctx context.Context,
	apiKey *resourcev1alpha1.APIKey,
	err error,
	reason string,
	message string,
) {
	// Create new condition
	condition := metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: apiKey.Generation,
		LastTransitionTime: metav1.Now(),
	}

	if err != nil {
		condition.Status = metav1.ConditionFalse
	}

	// Create new conditions slice for comparison
	newConditions := make([]metav1.Condition, 0)
	for _, c := range apiKey.Status.Conditions {
		if c.Type != condition.Type {
			newConditions = append(newConditions, c)
		}
	}
	newConditions = append(newConditions, condition)

	// Check if status has actually changed
	if !controllers2.StatusHasChanged(apiKey.Status.Conditions, newConditions) {
		return
	}

	// Update conditions
	apiKey.Status.Conditions = newConditions

	// Update observed generation
	apiKey.Status.ObservedGeneration = apiKey.Generation

	// Update status
	if err := r.Status().Update(ctx, apiKey); err != nil {
		log.FromContext(ctx).Error(err, "Failed to update APIKey status")
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *APIKeyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Initialize the watcher map
	r.watcherMap = make(map[types.NamespacedName]watch.Interface)

	return ctrl.NewControllerManagedBy(mgr).
		For(&resourcev1alpha1.APIKey{}).
		WithEventFilter(predicate.Or(
			// Trigger on spec changes
			predicate.GenerationChangedPredicate{},
			// Trigger periodically to sync status
			predicate.NewPredicateFuncs(func(object client.Object) bool {
				// Trigger every minute to sync status
				return true
			}),
		)).
		Complete(r)
}
