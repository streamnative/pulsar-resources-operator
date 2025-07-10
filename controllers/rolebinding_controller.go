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

	cloudapi "github.com/streamnative/pulsar-resources-operator/pkg/streamnativecloud/apis/cloud/v1alpha1"
)

const (
	roleBindingFinalizer = "rolebinding.resource.streamnative.io/finalizer"
)

// RoleBindingReconciler reconciles a RoleBinding object
type RoleBindingReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	ConnectionManager *ConnectionManager
}

//+kubebuilder:rbac:groups=resource.streamnative.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=rolebindings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=rolebindings/finalizers,verbs=update
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=streamnativecloudconnections,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *RoleBindingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling RoleBinding", "name", req.Name, "namespace", req.Namespace)

	// Fetch the RoleBinding instance
	roleBinding := &resourcev1alpha1.RoleBinding{}
	err := r.Get(ctx, req.NamespacedName, roleBinding)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected.
			logger.Info("RoleBinding resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		logger.Error(err, "Failed to get RoleBinding")
		return ctrl.Result{}, err
	}

	// Get the APIServerConnection
	connection := &resourcev1alpha1.StreamNativeCloudConnection{}
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: req.Namespace,
		Name:      roleBinding.Spec.APIServerRef.Name,
	}, connection); err != nil {
		r.updateRoleBindingStatus(ctx, roleBinding, err, "ConnectionNotFound",
			fmt.Sprintf("Failed to get APIServerConnection: %v", err))
		return ctrl.Result{}, err
	}

	// Get API connection
	apiConn, err := r.ConnectionManager.GetOrCreateConnection(connection, nil)
	if err != nil {
		if _, ok := err.(*NotInitializedError); ok {
			logger.Info("Connection not initialized, requeueing", "error", err.Error())
			return ctrl.Result{Requeue: true}, nil
		}
		r.updateRoleBindingStatus(ctx, roleBinding, err, "GetConnectionFailed",
			fmt.Sprintf("Failed to get connection: %v", err))
		return ctrl.Result{}, err
	}

	// Get organization from connection
	organization := connection.Spec.Organization
	if organization == "" {
		err := fmt.Errorf("organization is required but not specified")
		r.updateRoleBindingStatus(ctx, roleBinding, err, "ValidationFailed", err.Error())
		return ctrl.Result{}, err
	}

	// Create RoleBinding client
	rbClient, err := controllers2.NewRoleBindingClient(apiConn, organization)
	if err != nil {
		logger.Error(err, "Failed to create RoleBinding client")
		r.updateRoleBindingStatus(ctx, roleBinding, err, "ClientError", "Failed to create RoleBinding client")
		return ctrl.Result{}, err
	}

	// Check if RoleBinding is being deleted
	if roleBinding.DeletionTimestamp != nil {
		return r.handleRoleBindingDeletion(ctx, rbClient, roleBinding)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(roleBinding, roleBindingFinalizer) {
		controllerutil.AddFinalizer(roleBinding, roleBindingFinalizer)
		err := r.Update(ctx, roleBinding)
		if err != nil {
			logger.Error(err, "Failed to add finalizer")
			return ctrl.Result{}, err
		}
		logger.Info("Added finalizer to RoleBinding")
		return ctrl.Result{}, nil
	}

	// Reconcile the RoleBinding
	return r.reconcileRoleBinding(ctx, rbClient, roleBinding)
}

func (r *RoleBindingReconciler) handleRoleBindingDeletion(ctx context.Context, rbClient *controllers2.RoleBindingClient, roleBinding *resourcev1alpha1.RoleBinding) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Handling RoleBinding deletion", "name", roleBinding.Name)

	if controllerutil.ContainsFinalizer(roleBinding, roleBindingFinalizer) {
		// Delete the RoleBinding from the API server
		err := rbClient.DeleteRoleBinding(ctx, roleBinding)
		if err != nil {
			logger.Error(err, "Failed to delete RoleBinding from API server")
			r.updateRoleBindingStatus(ctx, roleBinding, err, "DeletionError", "Failed to delete RoleBinding from API server")
			return ctrl.Result{}, err
		}

		// Remove the finalizer
		controllerutil.RemoveFinalizer(roleBinding, roleBindingFinalizer)
		err = r.Update(ctx, roleBinding)
		if err != nil {
			logger.Error(err, "Failed to remove finalizer")
			return ctrl.Result{}, err
		}
		logger.Info("Successfully deleted RoleBinding and removed finalizer")
	}

	return ctrl.Result{}, nil
}

func (r *RoleBindingReconciler) reconcileRoleBinding(ctx context.Context, rbClient *controllers2.RoleBindingClient, roleBinding *resourcev1alpha1.RoleBinding) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling RoleBinding", "name", roleBinding.Name)

	// Check if RoleBinding exists on the API server
	cloudRoleBinding, err := rbClient.GetRoleBinding(ctx, roleBinding.Name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// RoleBinding doesn't exist, create it
			logger.Info("RoleBinding not found on API server, creating", "name", roleBinding.Name)
			cloudRoleBinding, err = rbClient.CreateRoleBinding(ctx, roleBinding)
			if err != nil {
				logger.Error(err, "Failed to create RoleBinding on API server")
				r.updateRoleBindingStatus(ctx, roleBinding, err, "CreationError", "Failed to create RoleBinding on API server")
				return ctrl.Result{}, err
			}
			logger.Info("Successfully created RoleBinding on API server", "name", roleBinding.Name)
		} else {
			logger.Error(err, "Failed to get RoleBinding from API server")
			r.updateRoleBindingStatus(ctx, roleBinding, err, "GetError", "Failed to get RoleBinding from API server")
			return ctrl.Result{}, err
		}
	} else {
		// RoleBinding exists, check if it needs to be updated
		if r.needsUpdate(roleBinding, cloudRoleBinding) {
			logger.Info("RoleBinding needs update", "name", roleBinding.Name)
			cloudRoleBinding, err = rbClient.UpdateRoleBinding(ctx, roleBinding)
			if err != nil {
				logger.Error(err, "Failed to update RoleBinding on API server")
				r.updateRoleBindingStatus(ctx, roleBinding, err, "UpdateError", "Failed to update RoleBinding on API server")
				return ctrl.Result{}, err
			}
			logger.Info("Successfully updated RoleBinding on API server", "name", roleBinding.Name)
		}
	}

	// Update status
	r.updateRoleBindingStatus(ctx, roleBinding, nil, "Ready", "RoleBinding is ready")

	// Requeue for periodic sync
	requeueInterval := 5 * time.Minute
	logger.Info("Successfully reconciled RoleBinding", "roleBindingName", roleBinding.Name, "requeueAfter", requeueInterval)
	return ctrl.Result{RequeueAfter: requeueInterval}, nil
}

func (r *RoleBindingReconciler) needsUpdate(local *resourcev1alpha1.RoleBinding, cloud *cloudapi.RoleBinding) bool {
	// For now, we'll assume updates are needed if the generation has changed
	// In a more sophisticated implementation, we would compare the specs
	return local.Generation != local.Status.ObservedGeneration
}

func (r *RoleBindingReconciler) updateRoleBindingStatus(
	ctx context.Context,
	roleBinding *resourcev1alpha1.RoleBinding,
	err error,
	reason string,
	message string,
) {
	logger := log.FromContext(ctx)
	roleBinding.Status.ObservedGeneration = roleBinding.Generation

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
	for i, condition := range roleBinding.Status.Conditions {
		if condition.Type == "Ready" {
			// Only update LastTransitionTime if Status or Reason or Message changes
			if condition.Status != newCondition.Status || condition.Reason != newCondition.Reason || condition.Message != newCondition.Message {
				roleBinding.Status.Conditions[i] = newCondition
			} else {
				// If nothing changed, keep the old LastTransitionTime
				newCondition.LastTransitionTime = condition.LastTransitionTime
				roleBinding.Status.Conditions[i] = newCondition
			}
			found = true
			break
		}
	}
	if !found {
		roleBinding.Status.Conditions = append(roleBinding.Status.Conditions, newCondition)
	}

	// Persist status update
	if statusUpdateErr := r.Status().Update(ctx, roleBinding); statusUpdateErr != nil {
		logger.Error(statusUpdateErr, "Failed to update RoleBinding status")
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *RoleBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&resourcev1alpha1.RoleBinding{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}
