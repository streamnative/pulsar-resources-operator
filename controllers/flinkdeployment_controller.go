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
	"encoding/json"
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

	computeapi "github.com/streamnative/cloud-api-server/pkg/apis/compute/v1alpha1"
)

// FlinkDeploymentReconciler reconciles a FlinkDeployment object
type FlinkDeploymentReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	ConnectionManager *ConnectionManager
	// watcherMap stores active watchers for FlinkDeployments
	watcherMap map[types.NamespacedName]watch.Interface
	// watcherMutex protects watcherMap
	watcherMutex sync.RWMutex
}

//+kubebuilder:rbac:groups=resource.streamnative.io,resources=computeflinkdeployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=computeflinkdeployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=computeflinkdeployments/finalizers,verbs=update
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=streamnativecloudconnections,verbs=get;list;watch
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=computeworkspaces,verbs=get;list;watch

// handleWatchEvents processes events from the watch interface
func (r *FlinkDeploymentReconciler) handleWatchEvents(ctx context.Context, namespacedName types.NamespacedName, watcher watch.Interface) {
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
				remoteDeployment, ok := event.Object.(*computeapi.FlinkDeployment)
				if !ok {
					logger.Error(fmt.Errorf("unexpected object type"), "Failed to convert object to FlinkDeployment")
					continue
				}

				// Get the local deployment
				localDeployment := &resourcev1alpha1.ComputeFlinkDeployment{}
				if err := r.Get(ctx, namespacedName, localDeployment); err != nil {
					logger.Error(err, "Failed to get local FlinkDeployment")
					continue
				}

				// Convert status to RawExtension
				statusBytes, err := json.Marshal(remoteDeployment.Status)
				if err != nil {
					logger.Error(err, "Failed to marshal deployment status")
					continue
				}

				// Create new status
				newStatus := localDeployment.Status.DeepCopy()
				newStatus.DeploymentStatus = &runtime.RawExtension{Raw: statusBytes}

				// Check if status has changed
				if !controllers2.FlinkDeploymentStatusHasChanged(&localDeployment.Status, newStatus) {
					continue
				}

				// Update status
				localDeployment.Status = *newStatus
				if err := r.Status().Update(ctx, localDeployment); err != nil {
					logger.Error(err, "Failed to update FlinkDeployment status")
				}
			}
		}
	}
}

// setupWatch creates a new watcher for a FlinkDeployment
func (r *FlinkDeploymentReconciler) setupWatch(ctx context.Context, deployment *resourcev1alpha1.ComputeFlinkDeployment, deploymentClient *controllers2.FlinkDeploymentClient) error {
	namespacedName := types.NamespacedName{
		Namespace: deployment.Namespace,
		Name:      deployment.Name,
	}

	// Check if we already have a watcher
	r.watcherMutex.RLock()
	_, exists := r.watcherMap[namespacedName]
	r.watcherMutex.RUnlock()
	if exists {
		return nil
	}

	// Create new watcher
	watcher, err := deploymentClient.WatchFlinkDeployment(ctx, deployment.Name)
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

// Reconcile handles the reconciliation of FlinkDeployment objects
func (r *FlinkDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling FlinkDeployment", "namespace", req.Namespace, "name", req.Name)

	// Add requeue interval for status sync
	requeueInterval := time.Minute

	// Get the FlinkDeployment resource
	deployment := &resourcev1alpha1.ComputeFlinkDeployment{}
	if err := r.Get(ctx, req.NamespacedName, deployment); err != nil {
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

	// Get APIServerRef from ComputeWorkspace if not specified in FlinkDeployment
	apiServerRef := deployment.Spec.APIServerRef
	if apiServerRef.Name == "" {
		// Get the ComputeWorkspace
		workspace := &resourcev1alpha1.ComputeWorkspace{}
		if err := r.Get(ctx, types.NamespacedName{
			Namespace: req.Namespace,
			Name:      deployment.Spec.WorkspaceName,
		}, workspace); err != nil {
			r.updateDeploymentStatus(ctx, deployment, err, "GetWorkspaceFailed",
				fmt.Sprintf("Failed to get ComputeWorkspace: %v", err))
			return ctrl.Result{}, err
		}
		apiServerRef = workspace.Spec.APIServerRef
	}

	// Get the APIServerConnection
	apiConn := &resourcev1alpha1.StreamNativeCloudConnection{}
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: req.Namespace,
		Name:      apiServerRef.Name,
	}, apiConn); err != nil {
		r.updateDeploymentStatus(ctx, deployment, err, "GetAPIServerConnectionFailed",
			fmt.Sprintf("Failed to get APIServerConnection: %v", err))
		return ctrl.Result{}, err
	}

	// Get the connection
	conn, err := r.ConnectionManager.GetOrCreateConnection(apiConn, nil)
	if err != nil {
		// If connection is not initialized, requeue the request
		if _, ok := err.(*NotInitializedError); ok {
			logger.Info("Connection not initialized, requeueing", "error", err.Error())
			return ctrl.Result{Requeue: true}, nil
		}
		r.updateDeploymentStatus(ctx, deployment, err, "GetConnectionFailed",
			fmt.Sprintf("Failed to get connection: %v", err))
		return ctrl.Result{}, err
	}

	// Create deployment client
	deploymentClient, err := controllers2.NewFlinkDeploymentClient(conn, apiConn.Spec.Organization)
	if err != nil {
		r.updateDeploymentStatus(ctx, deployment, err, "CreateDeploymentClientFailed",
			fmt.Sprintf("Failed to create deployment client: %v", err))
		return ctrl.Result{}, err
	}

	// Handle deletion
	if !deployment.DeletionTimestamp.IsZero() {
		if controllers2.ContainsString(deployment.Finalizers, controllers2.FlinkDeploymentFinalizer) {
			// Try to delete remote deployment
			if err := deploymentClient.DeleteFlinkDeployment(ctx, deployment); err != nil {
				if !apierrors.IsNotFound(err) {
					r.updateDeploymentStatus(ctx, deployment, err, "DeleteFailed",
						fmt.Sprintf("Failed to delete external resources: %v", err))
					return ctrl.Result{}, err
				}
				// If the resource is already gone, that's fine
				logger.Info("Remote FlinkDeployment already deleted or not found",
					"deployment", deployment.Name)
			}

			// Remove finalizer after successful deletion
			deployment.Finalizers = controllers2.RemoveString(deployment.Finalizers, controllers2.FlinkDeploymentFinalizer)
			if err := r.Update(ctx, deployment); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer if it doesn't exist
	if !controllers2.ContainsString(deployment.Finalizers, controllers2.FlinkDeploymentFinalizer) {
		deployment.Finalizers = append(deployment.Finalizers, controllers2.FlinkDeploymentFinalizer)
		if err := r.Update(ctx, deployment); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Check if deployment exists
	existingDeployment, err := deploymentClient.GetFlinkDeployment(ctx, deployment.Name)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			r.updateDeploymentStatus(ctx, deployment, err, "GetDeploymentFailed",
				fmt.Sprintf("Failed to get deployment: %v", err))
			return ctrl.Result{}, err
		}
		existingDeployment = nil
	}

	if existingDeployment == nil {
		// Create deployment
		resp, err := deploymentClient.CreateFlinkDeployment(ctx, deployment)
		if err != nil {
			r.updateDeploymentStatus(ctx, deployment, err, "CreateDeploymentFailed",
				fmt.Sprintf("Failed to create deployment: %v", err))
			return ctrl.Result{}, err
		}

		// Convert status to RawExtension
		statusBytes, err := json.Marshal(resp.Status)
		if err != nil {
			r.updateDeploymentStatus(ctx, deployment, err, "StatusMarshalFailed",
				fmt.Sprintf("Failed to marshal deployment status: %v", err))
			return ctrl.Result{}, err
		}

		deployment.Status.DeploymentStatus = &runtime.RawExtension{
			Raw: statusBytes,
		}
		r.updateDeploymentStatus(ctx, deployment, nil, "Ready", "Deployment created successfully")
	} else {
		// Update deployment
		resp, err := deploymentClient.UpdateFlinkDeployment(ctx, deployment)
		if err != nil {
			r.updateDeploymentStatus(ctx, deployment, err, "UpdateDeploymentFailed",
				fmt.Sprintf("Failed to update deployment: %v", err))
			return ctrl.Result{}, err
		}

		// Convert status to RawExtension
		statusBytes, err := json.Marshal(resp.Status)
		if err != nil {
			r.updateDeploymentStatus(ctx, deployment, err, "StatusMarshalFailed",
				fmt.Sprintf("Failed to marshal deployment status: %v", err))
			return ctrl.Result{}, err
		}

		deployment.Status.DeploymentStatus = &runtime.RawExtension{
			Raw: statusBytes,
		}
		r.updateDeploymentStatus(ctx, deployment, nil, "Ready", "Deployment updated successfully")
	}

	// Setup watch after deployment is created/updated
	if err := r.setupWatch(ctx, deployment, deploymentClient); err != nil {
		logger.Error(err, "Failed to setup watch")
		// Don't return error, just log it
	}

	// Return with requeue interval for status sync
	return ctrl.Result{RequeueAfter: requeueInterval}, nil
}

func (r *FlinkDeploymentReconciler) updateDeploymentStatus(
	ctx context.Context,
	deployment *resourcev1alpha1.ComputeFlinkDeployment,
	err error,
	reason string,
	message string,
) {
	// Create new status for comparison
	newStatus := deployment.Status.DeepCopy()

	// Create new condition
	condition := metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: deployment.Generation,
		LastTransitionTime: metav1.Now(),
	}

	if err != nil {
		condition.Status = metav1.ConditionFalse
	}

	// Create new conditions slice
	newStatus.Conditions = make([]metav1.Condition, 0)
	for _, c := range deployment.Status.Conditions {
		if c.Type != condition.Type {
			newStatus.Conditions = append(newStatus.Conditions, c)
		}
	}
	newStatus.Conditions = append(newStatus.Conditions, condition)

	// Check if status has actually changed
	if !controllers2.FlinkDeploymentStatusHasChanged(&deployment.Status, newStatus) {
		return
	}

	// Update status
	deployment.Status = *newStatus
	if err := r.Status().Update(ctx, deployment); err != nil {
		log.FromContext(ctx).Error(err, "Failed to update FlinkDeployment status")
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *FlinkDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Initialize the watcher map
	r.watcherMap = make(map[types.NamespacedName]watch.Interface)

	return ctrl.NewControllerManagedBy(mgr).
		For(&resourcev1alpha1.ComputeFlinkDeployment{}).
		// Remove GenerationChangedPredicate to allow status updates
		// Add periodic reconciliation to sync status
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
