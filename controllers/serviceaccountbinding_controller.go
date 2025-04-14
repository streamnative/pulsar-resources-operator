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

// ServiceAccountBindingReconciler reconciles a StreamNative Cloud ServiceAccountBinding object
type ServiceAccountBindingReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	ConnectionManager *ConnectionManager
	// watcherMap stores active watchers for ServiceAccountBindings
	watcherMap map[types.NamespacedName]watch.Interface
	// watcherMutex protects watcherMap
	watcherMutex sync.RWMutex
}

const ServiceAccountBindingFinalizer = "serviceaccountbinding.resource.streamnative.io/finalizer"

//+kubebuilder:rbac:groups=resource.streamnative.io,resources=serviceaccountbindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=serviceaccountbindings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=serviceaccountbindings/finalizers,verbs=update
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=serviceaccounts,verbs=get;list;watch
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=streamnativecloudconnections,verbs=get;list;watch

// handleWatchEvents processes events from the watch interface
func (r *ServiceAccountBindingReconciler) handleWatchEvents(ctx context.Context, namespacedName types.NamespacedName, watcher watch.Interface) {
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
				// Check if the object is a ServiceAccountBinding
				_, ok := event.Object.(*cloudapi.ServiceAccountBinding)
				if !ok {
					logger.Error(fmt.Errorf("unexpected object type"), "Failed to convert object to ServiceAccountBinding")
					continue
				}

				// Get the local ServiceAccountBinding
				localBinding := &resourcev1alpha1.ServiceAccountBinding{}
				if err := r.Get(ctx, namespacedName, localBinding); err != nil {
					logger.Error(err, "Failed to get local ServiceAccountBinding")
					continue
				}

				// Update status
				r.updateServiceAccountBindingStatus(ctx, localBinding, nil, "Ready", "ServiceAccountBinding synced successfully")
			}
		}
	}
}

// setupWatch creates a new watcher for a ServiceAccountBinding
func (r *ServiceAccountBindingReconciler) setupWatch(ctx context.Context, binding *resourcev1alpha1.ServiceAccountBinding, bindingClient *controllers2.ServiceAccountBindingClient) error {
	namespacedName := types.NamespacedName{
		Namespace: binding.Namespace,
		Name:      binding.Name,
	}

	// Check if we already have a watcher
	r.watcherMutex.RLock()
	_, exists := r.watcherMap[namespacedName]
	r.watcherMutex.RUnlock()
	if exists {
		return nil
	}

	// Create new watcher
	watcher, err := bindingClient.WatchServiceAccountBinding(ctx, binding.Name)
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

// Reconcile handles the reconciliation of ServiceAccountBinding objects
func (r *ServiceAccountBindingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling ServiceAccountBinding", "namespace", req.Namespace, "name", req.Name)

	// Add requeue interval for status sync
	requeueInterval := time.Minute

	// Get the ServiceAccountBinding resource
	binding := &resourcev1alpha1.ServiceAccountBinding{}
	if err := r.Get(ctx, req.NamespacedName, binding); err != nil {
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

	// Get the ServiceAccount
	serviceAccount := &resourcev1alpha1.ServiceAccount{}
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: req.Namespace,
		Name:      binding.Spec.ServiceAccountName,
	}, serviceAccount); err != nil {
		r.updateServiceAccountBindingStatus(ctx, binding, err, "ServiceAccountNotFound",
			fmt.Sprintf("Failed to get ServiceAccount: %v", err))
		return ctrl.Result{}, err
	}

	// Get the APIServerConnection from the ServiceAccount
	connection := &resourcev1alpha1.StreamNativeCloudConnection{}
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: req.Namespace,
		Name:      serviceAccount.Spec.APIServerRef.Name,
	}, connection); err != nil {
		r.updateServiceAccountBindingStatus(ctx, binding, err, "ConnectionNotFound",
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
		r.updateServiceAccountBindingStatus(ctx, binding, err, "GetConnectionFailed",
			fmt.Sprintf("Failed to get connection: %v", err))
		return ctrl.Result{}, err
	}

	// Get organization from connection
	organization := connection.Spec.Organization
	if organization == "" {
		err := fmt.Errorf("organization is required but not specified")
		r.updateServiceAccountBindingStatus(ctx, binding, err, "ValidationFailed", err.Error())
		return ctrl.Result{}, err
	}

	// Create ServiceAccountBinding client
	bindingClient, err := controllers2.NewServiceAccountBindingClient(apiConn, organization)
	if err != nil {
		r.updateServiceAccountBindingStatus(ctx, binding, err, "ClientCreationFailed",
			fmt.Sprintf("Failed to create ServiceAccountBinding client: %v", err))
		return ctrl.Result{}, err
	}

	// Handle deletion
	if !binding.DeletionTimestamp.IsZero() {
		if controllers2.ContainsString(binding.Finalizers, ServiceAccountBindingFinalizer) {
			// Try to delete remote ServiceAccountBinding
			if err := bindingClient.DeleteServiceAccountBinding(ctx, binding); err != nil {
				if !apierrors.IsNotFound(err) {
					r.updateServiceAccountBindingStatus(ctx, binding, err, "DeleteFailed",
						fmt.Sprintf("Failed to delete external resources: %v", err))
					return ctrl.Result{}, err
				}
				// If the resource is already gone, that's fine
				logger.Info("Remote ServiceAccountBinding already deleted or not found",
					"binding", binding.Name)
			}

			// Remove finalizer after successful deletion
			binding.Finalizers = controllers2.RemoveString(binding.Finalizers, ServiceAccountBindingFinalizer)
			if err := r.Update(ctx, binding); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer if it doesn't exist
	if !controllers2.ContainsString(binding.Finalizers, ServiceAccountBindingFinalizer) {
		binding.Finalizers = append(binding.Finalizers, ServiceAccountBindingFinalizer)
		if err := r.Update(ctx, binding); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Check if ServiceAccountBinding exists
	existingBinding, err := bindingClient.GetServiceAccountBinding(ctx, binding.Name)
	if err != nil {
		logger.Info("Failed to get ServiceAccountBinding", "error", err, "existingBinding", existingBinding)
		if !apierrors.IsNotFound(err) {
			r.updateServiceAccountBindingStatus(ctx, binding, err, "GetServiceAccountBindingFailed",
				fmt.Sprintf("Failed to get ServiceAccountBinding: %v", err))
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		existingBinding = nil
	}

	if existingBinding == nil {
		// Create ServiceAccountBinding
		_, err := bindingClient.CreateServiceAccountBinding(ctx, binding)
		if err != nil {
			r.updateServiceAccountBindingStatus(ctx, binding, err, "CreateServiceAccountBindingFailed",
				fmt.Sprintf("Failed to create ServiceAccountBinding: %v", err))
			return ctrl.Result{}, err
		}

		// Update status
		r.updateServiceAccountBindingStatus(ctx, binding, nil, "Ready", "ServiceAccountBinding created successfully")
	} else {
		// Update ServiceAccountBinding
		_, err := bindingClient.UpdateServiceAccountBinding(ctx, binding)
		if err != nil {
			r.updateServiceAccountBindingStatus(ctx, binding, err, "UpdateServiceAccountBindingFailed",
				fmt.Sprintf("Failed to update ServiceAccountBinding: %v", err))
			return ctrl.Result{}, err
		}

		// Update status
		r.updateServiceAccountBindingStatus(ctx, binding, nil, "Ready", "ServiceAccountBinding updated successfully")
	}

	// Setup watch after ServiceAccountBinding is created/updated
	if err := r.setupWatch(ctx, binding, bindingClient); err != nil {
		logger.Error(err, "Failed to setup watch")
		// Don't return error, just log it
	}

	return ctrl.Result{RequeueAfter: requeueInterval}, nil
}

func (r *ServiceAccountBindingReconciler) updateServiceAccountBindingStatus(
	ctx context.Context,
	binding *resourcev1alpha1.ServiceAccountBinding,
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
		ObservedGeneration: binding.Generation,
		LastTransitionTime: metav1.Now(),
	}

	if err != nil {
		condition.Status = metav1.ConditionFalse
	}

	// Create new conditions slice for comparison
	newConditions := make([]metav1.Condition, 0)
	for _, c := range binding.Status.Conditions {
		if c.Type != condition.Type {
			newConditions = append(newConditions, c)
		}
	}
	newConditions = append(newConditions, condition)

	// Check if status has actually changed
	if !controllers2.StatusHasChanged(binding.Status.Conditions, newConditions) {
		return
	}

	// Update conditions
	binding.Status.Conditions = newConditions

	// Update observed generation
	binding.Status.ObservedGeneration = binding.Generation

	// Update status
	if err := r.Status().Update(ctx, binding); err != nil {
		log.FromContext(ctx).Error(err, "Failed to update ServiceAccountBinding status")
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceAccountBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Initialize the watcher map
	r.watcherMap = make(map[types.NamespacedName]watch.Interface)

	return ctrl.NewControllerManagedBy(mgr).
		For(&resourcev1alpha1.ServiceAccountBinding{}).
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
