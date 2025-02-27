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

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	cloudapi "github.com/streamnative/cloud-api-server/pkg/apis/cloud/v1alpha1"
)

// SecretReconciler reconciles a StreamNative Cloud Secret object
type SecretReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	ConnectionManager *ConnectionManager
	// watcherMap stores active watchers for Secrets
	watcherMap map[types.NamespacedName]watch.Interface
	// watcherMutex protects watcherMap
	watcherMutex sync.RWMutex
}

//+kubebuilder:rbac:groups=resource.streamnative.io,resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=secrets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=secrets/finalizers,verbs=update
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=streamnativecloudconnections,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

// handleWatchEvents processes events from the watch interface
func (r *SecretReconciler) handleWatchEvents(ctx context.Context, namespacedName types.NamespacedName, watcher watch.Interface) {
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
				// Check if the object is a Secret
				_, ok := event.Object.(*cloudapi.Secret)
				if !ok {
					logger.Error(fmt.Errorf("unexpected object type"), "Failed to convert object to Secret")
					continue
				}

				// Get the local secret
				localSecret := &resourcev1alpha1.Secret{}
				if err := r.Get(ctx, namespacedName, localSecret); err != nil {
					logger.Error(err, "Failed to get local Secret")
					continue
				}

				// Update status
				r.updateSecretStatus(ctx, localSecret, nil, "Ready", "Secret synced successfully")
			}
		}
	}
}

// setupWatch creates a new watcher for a Secret
func (r *SecretReconciler) setupWatch(ctx context.Context, secret *resourcev1alpha1.Secret, secretClient *controllers2.SecretClient) error {
	namespacedName := types.NamespacedName{
		Namespace: secret.Namespace,
		Name:      secret.Name,
	}

	// Check if we already have a watcher
	r.watcherMutex.RLock()
	_, exists := r.watcherMap[namespacedName]
	r.watcherMutex.RUnlock()
	if exists {
		return nil
	}

	// Create new watcher
	watcher, err := secretClient.WatchSecret(ctx, secret.Name)
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

// getSecretData obtains the Secret data either from direct Data field or from SecretRef
func (r *SecretReconciler) getSecretData(ctx context.Context, secret *resourcev1alpha1.Secret) (map[string]string, *corev1.SecretType, error) {
	// If direct data is provided, use it
	if len(secret.Spec.Data) > 0 {
		return secret.Spec.Data, secret.Spec.Type, nil
	}

	// If SecretRef is provided, fetch from the referenced Kubernetes Secret
	if secret.Spec.SecretRef != nil {
		// Get the referenced Kubernetes Secret
		nsName := secret.Spec.SecretRef.ToNamespacedName()
		k8sSecret := &corev1.Secret{}
		if err := r.Get(ctx, nsName, k8sSecret); err != nil {
			return nil, nil, fmt.Errorf("failed to get referenced Secret %s/%s: %w", nsName.Namespace, nsName.Name, err)
		}

		// Convert the binary data to string data
		stringData := make(map[string]string)
		for k, v := range k8sSecret.Data {
			stringData[k] = string(v)
		}
		return stringData, &k8sSecret.Type, nil
	}

	// Neither Data nor SecretRef is provided
	return nil, nil, fmt.Errorf("neither Data nor SecretRef is specified in the Secret spec")
}

// Reconcile handles the reconciliation of Secret objects
func (r *SecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling Secret", "namespace", req.Namespace, "name", req.Name)

	// Add requeue interval for status sync
	requeueInterval := time.Minute

	// Get the Secret resource
	secret := &resourcev1alpha1.Secret{}
	if err := r.Get(ctx, req.NamespacedName, secret); err != nil {
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
		Name:      secret.Spec.APIServerRef.Name,
	}, connection); err != nil {
		r.updateSecretStatus(ctx, secret, err, "ConnectionNotFound",
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
		r.updateSecretStatus(ctx, secret, err, "GetConnectionFailed",
			fmt.Sprintf("Failed to get connection: %v", err))
		return ctrl.Result{}, err
	}

	// Get organization from connection
	organization := connection.Spec.Organization
	if organization == "" {
		err := fmt.Errorf("organization is required but not specified")
		r.updateSecretStatus(ctx, secret, err, "ValidationFailed", err.Error())
		return ctrl.Result{}, err
	}

	// Create secret client
	secretClient, err := controllers2.NewSecretClient(apiConn, organization)
	if err != nil {
		r.updateSecretStatus(ctx, secret, err, "ClientCreationFailed",
			fmt.Sprintf("Failed to create secret client: %v", err))
		return ctrl.Result{}, err
	}

	// Handle deletion
	if !secret.DeletionTimestamp.IsZero() {
		if controllers2.ContainsString(secret.Finalizers, controllers2.SecretFinalizer) {
			// Try to delete remote secret
			if err := secretClient.DeleteSecret(ctx, secret); err != nil {
				if !apierrors.IsNotFound(err) {
					r.updateSecretStatus(ctx, secret, err, "DeleteFailed",
						fmt.Sprintf("Failed to delete external resources: %v", err))
					return ctrl.Result{}, err
				}
				// If the resource is already gone, that's fine
				logger.Info("Remote Secret already deleted or not found",
					"secret", secret.Name)
			}

			// Remove finalizer after successful deletion
			secret.Finalizers = controllers2.RemoveString(secret.Finalizers, controllers2.SecretFinalizer)
			if err := r.Update(ctx, secret); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer if it doesn't exist
	if !controllers2.ContainsString(secret.Finalizers, controllers2.SecretFinalizer) {
		secret.Finalizers = append(secret.Finalizers, controllers2.SecretFinalizer)
		if err := r.Update(ctx, secret); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Get secret data (either from direct Data field or from SecretRef)
	secretData, secretType, err := r.getSecretData(ctx, secret)
	if err != nil {
		r.updateSecretStatus(ctx, secret, err, "GetSecretDataFailed",
			fmt.Sprintf("Failed to get secret data: %v", err))
		return ctrl.Result{}, err
	}

	// Update the secret's Data field with the fetched data if SecretRef was used
	if secret.Spec.SecretRef != nil && len(secret.Spec.Data) == 0 {
		secret.Spec.Data = secretData
		if secret.Spec.Type == nil || *secret.Spec.Type == "" {
			secret.Spec.Type = secretType
		}
		if err := r.Update(ctx, secret); err != nil {
			r.updateSecretStatus(ctx, secret, err, "UpdateSecretFailed",
				fmt.Sprintf("Failed to update secret with data from referenced Secret: %v", err))
			return ctrl.Result{}, err
		}
	}

	// Check if secret exists
	existingSecret, err := secretClient.GetSecret(ctx, secret.Name)
	if err != nil {
		logger.Info("Failed to get secret", "error", err, "existingSecret", existingSecret)
		if !apierrors.IsNotFound(err) {
			r.updateSecretStatus(ctx, secret, err, "GetSecretFailed",
				fmt.Sprintf("Failed to get secret: %v", err))
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		existingSecret = nil
	}

	if existingSecret == nil {
		// Create secret
		_, err := secretClient.CreateSecret(ctx, secret)
		if err != nil {
			r.updateSecretStatus(ctx, secret, err, "CreateSecretFailed",
				fmt.Sprintf("Failed to create secret: %v", err))
			return ctrl.Result{}, err
		}

		// Update status
		r.updateSecretStatus(ctx, secret, nil, "Ready", "Secret created successfully")
	} else {
		// Update secret
		_, err := secretClient.UpdateSecret(ctx, secret)
		if err != nil {
			r.updateSecretStatus(ctx, secret, err, "UpdateSecretFailed",
				fmt.Sprintf("Failed to update secret: %v", err))
			return ctrl.Result{}, err
		}
		// Update status
		r.updateSecretStatus(ctx, secret, nil, "Ready", "Secret updated successfully")
	}

	// Setup watch after secret is created/updated
	if err := r.setupWatch(ctx, secret, secretClient); err != nil {
		logger.Error(err, "Failed to setup watch")
		// Don't return error, just log it
	}

	return ctrl.Result{RequeueAfter: requeueInterval}, nil
}

func (r *SecretReconciler) updateSecretStatus(
	ctx context.Context,
	secret *resourcev1alpha1.Secret,
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
		ObservedGeneration: secret.Generation,
		LastTransitionTime: metav1.Now(),
	}

	if err != nil {
		condition.Status = metav1.ConditionFalse
	}

	// Create new conditions slice for comparison
	newConditions := make([]metav1.Condition, 0)
	for _, c := range secret.Status.Conditions {
		if c.Type != condition.Type {
			newConditions = append(newConditions, c)
		}
	}
	newConditions = append(newConditions, condition)

	// Check if status has actually changed
	if !controllers2.StatusHasChanged(secret.Status.Conditions, newConditions) {
		return
	}

	// Update conditions
	secret.Status.Conditions = newConditions

	// Update observed generation
	secret.Status.ObservedGeneration = secret.Generation

	// Update status
	if err := r.Status().Update(ctx, secret); err != nil {
		log.FromContext(ctx).Error(err, "Failed to update Secret status")
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *SecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Initialize the watcher map
	r.watcherMap = make(map[types.NamespacedName]watch.Interface)

	return ctrl.NewControllerManagedBy(mgr).
		For(&resourcev1alpha1.Secret{}).
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
