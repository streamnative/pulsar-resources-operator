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
	"time"

	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	controllers2 "github.com/streamnative/pulsar-resources-operator/pkg/streamnativecloud"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// WorkspaceReconciler reconciles a Workspace object
type WorkspaceReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	ConnectionManager *ConnectionManager
}

//+kubebuilder:rbac:groups=resource.streamnative.io,resources=computeworkspaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=computeworkspaces/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=computeworkspaces/finalizers,verbs=update
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=streamnativecloudconnections,verbs=get;list;watch

// Reconcile handles the reconciliation of Workspace objects
func (r *WorkspaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling Workspace", "namespace", req.Namespace, "name", req.Name)

	// Add requeue interval for status sync
	requeueInterval := time.Minute

	// Get the Workspace resource
	workspace := &resourcev1alpha1.ComputeWorkspace{}
	if err := r.Get(ctx, req.NamespacedName, workspace); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Workspace not found. Reconciliation will stop.", "namespace", req.Namespace, "name", req.Name)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Get the APIServerConnection
	connection := &resourcev1alpha1.StreamNativeCloudConnection{}
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: req.Namespace,
		Name:      workspace.Spec.APIServerRef.Name,
	}, connection); err != nil {
		r.updateWorkspaceStatus(ctx, workspace, err, "ConnectionNotFound",
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
		r.updateWorkspaceStatus(ctx, workspace, err, "GetConnectionFailed",
			fmt.Sprintf("Failed to get connection: %v", err))
		return ctrl.Result{}, err
	}

	// Get organization from connection or workspace
	organization := connection.Spec.Organization
	if organization == "" {
		err := fmt.Errorf("organization is required but not specified")
		r.updateWorkspaceStatus(ctx, workspace, err, "ValidationFailed", err.Error())
		return ctrl.Result{}, err
	}

	// Create workspace client
	workspaceClient, err := controllers2.NewWorkspaceClient(apiConn, organization)
	if err != nil {
		r.updateWorkspaceStatus(ctx, workspace, err, "ClientCreationFailed",
			fmt.Sprintf("Failed to create workspace client: %v", err))
		return ctrl.Result{}, err
	}

	// Handle deletion
	if !workspace.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(workspace, controllers2.WorkspaceFinalizer) {
			// Try to delete remote workspace
			if err := workspaceClient.DeleteWorkspace(ctx, workspace); err != nil {
				if !apierrors.IsNotFound(err) {
					r.updateWorkspaceStatus(ctx, workspace, err, "DeleteFailed",
						fmt.Sprintf("Failed to delete external resources: %v", err))
					return ctrl.Result{}, err
				}
				// If the resource is already gone, that's fine
				logger.Info("Remote Workspace already deleted or not found",
					"workspace", workspace.Name)
			}

			// Remove finalizer after successful deletion
			controllerutil.RemoveFinalizer(workspace, controllers2.WorkspaceFinalizer)
			if err := r.Update(ctx, workspace); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer if it doesn't exist
	if !controllerutil.ContainsFinalizer(workspace, controllers2.WorkspaceFinalizer) {
		controllerutil.AddFinalizer(workspace, controllers2.WorkspaceFinalizer)
		if err := r.Update(ctx, workspace); err != nil {
			return ctrl.Result{}, err
		}
		// Requeue after adding finalizer to ensure the update is processed before proceeding
		return ctrl.Result{Requeue: true}, nil
	}

	// Check if workspace exists
	existingWorkspace, err := workspaceClient.GetWorkspace(ctx, workspace.Name)
	if err != nil {
		logger.Info("Failed to get workspace", "error", err, "existingWorkspace", existingWorkspace)
		if !apierrors.IsNotFound(err) {
			r.updateWorkspaceStatus(ctx, workspace, err, "GetWorkspaceFailed",
				fmt.Sprintf("Failed to get workspace: %v", err))
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		existingWorkspace = nil
	}

	if existingWorkspace == nil {
		// Create workspace
		resp, err := workspaceClient.CreateWorkspace(ctx, workspace)
		if err != nil {
			r.updateWorkspaceStatus(ctx, workspace, err, "CreateWorkspaceFailed",
				fmt.Sprintf("Failed to create workspace: %v", err))
			return ctrl.Result{}, err
		}

		// Update status
		workspace.Status.WorkspaceID = resp.Name
		r.updateWorkspaceStatus(ctx, workspace, nil, "Ready", "Workspace created successfully")
	} else {
		// Update workspace
		updatedWorkspace, err := workspaceClient.UpdateWorkspace(ctx, workspace)
		if err != nil {
			r.updateWorkspaceStatus(ctx, workspace, err, "UpdateWorkspaceFailed",
				fmt.Sprintf("Failed to update workspace: %v", err))
			return ctrl.Result{}, err
		}
		workspace.Status.WorkspaceID = updatedWorkspace.Name
		r.updateWorkspaceStatus(ctx, workspace, nil, "Ready", "Workspace updated successfully")
	}

	return ctrl.Result{RequeueAfter: requeueInterval}, nil
}

func (r *WorkspaceReconciler) updateWorkspaceStatus(
	ctx context.Context,
	workspace *resourcev1alpha1.ComputeWorkspace,
	err error,
	reason string,
	message string,
) {
	logger := log.FromContext(ctx)
	workspace.Status.ObservedGeneration = workspace.Generation

	newCondition := metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.Now(),
	}

	if err != nil {
		newCondition.Status = metav1.ConditionFalse
	}

	// Update ready condition
	found := false
	for i, condition := range workspace.Status.Conditions {
		if condition.Type == "Ready" {
			if condition.Status != newCondition.Status || condition.Reason != newCondition.Reason || condition.Message != newCondition.Message {
				workspace.Status.Conditions[i] = newCondition
			} else {
				newCondition.LastTransitionTime = condition.LastTransitionTime
				workspace.Status.Conditions[i] = newCondition
			}
			found = true
			break
		}
	}

	if !found {
		workspace.Status.Conditions = append(workspace.Status.Conditions, newCondition)
	}

	// Persist status update
	if statusUpdateErr := r.Status().Update(ctx, workspace); statusUpdateErr != nil {
		logger.Error(statusUpdateErr, "Failed to update Workspace status")
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *WorkspaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&resourcev1alpha1.ComputeWorkspace{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}
