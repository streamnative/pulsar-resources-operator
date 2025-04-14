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

// ServiceAccountReconciler reconciles a StreamNative Cloud ServiceAccount object
type ServiceAccountReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	ConnectionManager *ConnectionManager
	// watcherMap stores active watchers for ServiceAccounts
	watcherMap map[types.NamespacedName]watch.Interface
	// watcherMutex protects watcherMap
	watcherMutex sync.RWMutex
}

const ServiceAccountFinalizer = "serviceaccount.resource.streamnative.io/finalizer"

//+kubebuilder:rbac:groups=resource.streamnative.io,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=serviceaccounts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=serviceaccounts/finalizers,verbs=update
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=streamnativecloudconnections,verbs=get;list;watch

// handleWatchEvents processes events from the watch interface
func (r *ServiceAccountReconciler) handleWatchEvents(ctx context.Context, namespacedName types.NamespacedName, watcher watch.Interface) {
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
				// Check if the object is a ServiceAccount
				_, ok := event.Object.(*cloudapi.ServiceAccount)
				if !ok {
					logger.Error(fmt.Errorf("unexpected object type"), "Failed to convert object to ServiceAccount")
					continue
				}

				// Get the local ServiceAccount
				localServiceAccount := &resourcev1alpha1.ServiceAccount{}
				if err := r.Get(ctx, namespacedName, localServiceAccount); err != nil {
					logger.Error(err, "Failed to get local ServiceAccount")
					continue
				}

				// Update status
				r.updateServiceAccountStatus(ctx, localServiceAccount, nil, "Ready", "ServiceAccount synced successfully")
			}
		}
	}
}

// setupWatch creates a new watcher for a ServiceAccount
func (r *ServiceAccountReconciler) setupWatch(ctx context.Context, serviceAccount *resourcev1alpha1.ServiceAccount, saClient *controllers2.ServiceAccountClient) error {
	namespacedName := types.NamespacedName{
		Namespace: serviceAccount.Namespace,
		Name:      serviceAccount.Name,
	}

	// Check if we already have a watcher
	r.watcherMutex.RLock()
	_, exists := r.watcherMap[namespacedName]
	r.watcherMutex.RUnlock()
	if exists {
		return nil
	}

	// Create new watcher
	watcher, err := saClient.WatchServiceAccount(ctx, serviceAccount.Name)
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

// Reconcile handles the reconciliation of ServiceAccount objects
func (r *ServiceAccountReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling ServiceAccount", "namespace", req.Namespace, "name", req.Name)

	// Add requeue interval for status sync
	requeueInterval := time.Minute

	// Get the ServiceAccount resource
	serviceAccount := &resourcev1alpha1.ServiceAccount{}
	if err := r.Get(ctx, req.NamespacedName, serviceAccount); err != nil {
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
		Name:      serviceAccount.Spec.APIServerRef.Name,
	}, connection); err != nil {
		r.updateServiceAccountStatus(ctx, serviceAccount, err, "ConnectionNotFound",
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
		r.updateServiceAccountStatus(ctx, serviceAccount, err, "GetConnectionFailed",
			fmt.Sprintf("Failed to get connection: %v", err))
		return ctrl.Result{}, err
	}

	// Get organization from connection
	organization := connection.Spec.Organization
	if organization == "" {
		err := fmt.Errorf("organization is required but not specified")
		r.updateServiceAccountStatus(ctx, serviceAccount, err, "ValidationFailed", err.Error())
		return ctrl.Result{}, err
	}

	// Create ServiceAccount client
	saClient, err := controllers2.NewServiceAccountClient(apiConn, organization)
	if err != nil {
		r.updateServiceAccountStatus(ctx, serviceAccount, err, "ClientCreationFailed",
			fmt.Sprintf("Failed to create ServiceAccount client: %v", err))
		return ctrl.Result{}, err
	}

	// Handle deletion
	if !serviceAccount.DeletionTimestamp.IsZero() {
		if controllers2.ContainsString(serviceAccount.Finalizers, ServiceAccountFinalizer) {
			// Try to delete remote ServiceAccount
			if err := saClient.DeleteServiceAccount(ctx, serviceAccount); err != nil {
				if !apierrors.IsNotFound(err) {
					r.updateServiceAccountStatus(ctx, serviceAccount, err, "DeleteFailed",
						fmt.Sprintf("Failed to delete external resources: %v", err))
					return ctrl.Result{}, err
				}
				// If the resource is already gone, that's fine
				logger.Info("Remote ServiceAccount already deleted or not found",
					"serviceAccount", serviceAccount.Name)
			}

			// Remove finalizer after successful deletion
			serviceAccount.Finalizers = controllers2.RemoveString(serviceAccount.Finalizers, ServiceAccountFinalizer)
			if err := r.Update(ctx, serviceAccount); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer if it doesn't exist
	if !controllers2.ContainsString(serviceAccount.Finalizers, ServiceAccountFinalizer) {
		serviceAccount.Finalizers = append(serviceAccount.Finalizers, ServiceAccountFinalizer)
		if err := r.Update(ctx, serviceAccount); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Check if ServiceAccount exists
	existingSA, err := saClient.GetServiceAccount(ctx, serviceAccount.Name)
	if err != nil {
		logger.Info("Failed to get ServiceAccount", "error", err, "existingSA", existingSA)
		if !apierrors.IsNotFound(err) {
			r.updateServiceAccountStatus(ctx, serviceAccount, err, "GetServiceAccountFailed",
				fmt.Sprintf("Failed to get ServiceAccount: %v", err))
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		existingSA = nil
	}

	if existingSA == nil {
		// Create ServiceAccount
		resultSA, err := saClient.CreateServiceAccount(ctx, serviceAccount)
		if err != nil {
			r.updateServiceAccountStatus(ctx, serviceAccount, err, "CreateServiceAccountFailed",
				fmt.Sprintf("Failed to create ServiceAccount: %v", err))
			return ctrl.Result{}, err
		}

		// Update status with private key information
		if resultSA.Status.PrivateKeyType != "" {
			serviceAccount.Status.PrivateKeyType = resultSA.Status.PrivateKeyType
		}
		if resultSA.Status.PrivateKeyData != "" {
			serviceAccount.Status.PrivateKeyData = resultSA.Status.PrivateKeyData
		}

		// Update status
		r.updateServiceAccountStatus(ctx, serviceAccount, nil, "Ready", "ServiceAccount created successfully")
	} else {
		// Update ServiceAccount
		resultSA, err := saClient.UpdateServiceAccount(ctx, serviceAccount)
		if err != nil {
			r.updateServiceAccountStatus(ctx, serviceAccount, err, "UpdateServiceAccountFailed",
				fmt.Sprintf("Failed to update ServiceAccount: %v", err))
			return ctrl.Result{}, err
		}

		// Update status with private key information if available
		if resultSA.Status.PrivateKeyType != "" {
			serviceAccount.Status.PrivateKeyType = resultSA.Status.PrivateKeyType
		}
		if resultSA.Status.PrivateKeyData != "" {
			serviceAccount.Status.PrivateKeyData = resultSA.Status.PrivateKeyData
		}

		// Update status
		r.updateServiceAccountStatus(ctx, serviceAccount, nil, "Ready", "ServiceAccount updated successfully")
	}

	// Setup watch after ServiceAccount is created/updated
	if err := r.setupWatch(ctx, serviceAccount, saClient); err != nil {
		logger.Error(err, "Failed to setup watch")
		// Don't return error, just log it
	}

	return ctrl.Result{RequeueAfter: requeueInterval}, nil
}

func (r *ServiceAccountReconciler) updateServiceAccountStatus(
	ctx context.Context,
	serviceAccount *resourcev1alpha1.ServiceAccount,
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
		ObservedGeneration: serviceAccount.Generation,
		LastTransitionTime: metav1.Now(),
	}

	if err != nil {
		condition.Status = metav1.ConditionFalse
	}

	// Create new conditions slice for comparison
	newConditions := make([]metav1.Condition, 0)
	for _, c := range serviceAccount.Status.Conditions {
		if c.Type != condition.Type {
			newConditions = append(newConditions, c)
		}
	}
	newConditions = append(newConditions, condition)

	// Check if status has actually changed
	if !controllers2.StatusHasChanged(serviceAccount.Status.Conditions, newConditions) {
		return
	}

	// Update conditions
	serviceAccount.Status.Conditions = newConditions

	// Update observed generation
	serviceAccount.Status.ObservedGeneration = serviceAccount.Generation

	// Update status
	if err := r.Status().Update(ctx, serviceAccount); err != nil {
		log.FromContext(ctx).Error(err, "Failed to update ServiceAccount status")
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceAccountReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Initialize the watcher map
	r.watcherMap = make(map[types.NamespacedName]watch.Interface)

	return ctrl.NewControllerManagedBy(mgr).
		For(&resourcev1alpha1.ServiceAccount{}).
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
