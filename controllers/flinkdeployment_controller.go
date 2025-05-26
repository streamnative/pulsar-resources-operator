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

// FlinkDeploymentReconciler reconciles a FlinkDeployment object
type FlinkDeploymentReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	ConnectionManager *ConnectionManager
}

//+kubebuilder:rbac:groups=resource.streamnative.io,resources=computeflinkdeployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=computeflinkdeployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=computeflinkdeployments/finalizers,verbs=update
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=streamnativecloudconnections,verbs=get;list;watch
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=computeworkspaces,verbs=get;list;watch

// Reconcile handles the reconciliation of FlinkDeployment objects
func (r *FlinkDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling FlinkDeployment", "namespace", req.Namespace, "name", req.Name)

	requeueInterval := time.Minute

	deployment := &resourcev1alpha1.ComputeFlinkDeployment{}
	if err := r.Get(ctx, req.NamespacedName, deployment); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("FlinkDeployment resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get FlinkDeployment resource")
		return ctrl.Result{}, err
	}

	apiServerRef := deployment.Spec.APIServerRef
	if apiServerRef.Name == "" {
		workspace := &resourcev1alpha1.ComputeWorkspace{}
		if err := r.Get(ctx, types.NamespacedName{
			Namespace: req.Namespace,
			Name:      deployment.Spec.WorkspaceName,
		}, workspace); err != nil {
			r.updateDeploymentStatus(ctx, deployment, err, "GetWorkspaceFailed",
				fmt.Sprintf("Failed to get ComputeWorkspace %s: %v", deployment.Spec.WorkspaceName, err))
			return ctrl.Result{}, err
		}
		if workspace.Spec.APIServerRef.Name == "" {
			err := fmt.Errorf("APIServerRef is empty in both FlinkDeployment spec and referenced ComputeWorkspace %s spec", workspace.Name)
			r.updateDeploymentStatus(ctx, deployment, err, "ValidationFailed", err.Error())
			return ctrl.Result{}, err
		}
		apiServerRef = workspace.Spec.APIServerRef
	}

	apiConnResource := &resourcev1alpha1.StreamNativeCloudConnection{}
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: req.Namespace,
		Name:      apiServerRef.Name,
	}, apiConnResource); err != nil {
		r.updateDeploymentStatus(ctx, deployment, err, "GetAPIServerConnectionFailed",
			fmt.Sprintf("Failed to get StreamNativeCloudConnection %s: %v", apiServerRef.Name, err))
		return ctrl.Result{}, err
	}

	conn, err := r.ConnectionManager.GetOrCreateConnection(apiConnResource, nil)
	if err != nil {
		if _, ok := err.(*NotInitializedError); ok {
			logger.Info("Connection not initialized, requeueing", "connectionName", apiConnResource.Name, "error", err.Error())
			return ctrl.Result{Requeue: true}, nil
		}
		r.updateDeploymentStatus(ctx, deployment, err, "GetConnectionFailed",
			fmt.Sprintf("Failed to get active connection for %s: %v", apiConnResource.Name, err))
		return ctrl.Result{}, err
	}

	if apiConnResource.Spec.Organization == "" {
		err := fmt.Errorf("organization is required in StreamNativeCloudConnection %s but not specified", apiConnResource.Name)
		r.updateDeploymentStatus(ctx, deployment, err, "ValidationFailed", err.Error())
		return ctrl.Result{}, err
	}
	deploymentClient, err := controllers2.NewFlinkDeploymentClient(conn, apiConnResource.Spec.Organization)
	if err != nil {
		r.updateDeploymentStatus(ctx, deployment, err, "CreateDeploymentClientFailed",
			fmt.Sprintf("Failed to create Flink deployment client: %v", err))
		return ctrl.Result{}, err
	}

	finalizerName := controllers2.FlinkDeploymentFinalizer
	if !deployment.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(deployment, finalizerName) {
			if err := deploymentClient.DeleteFlinkDeployment(ctx, deployment); err != nil {
				if !apierrors.IsNotFound(err) {
					r.updateDeploymentStatus(ctx, deployment, err, "DeleteFailed",
						fmt.Sprintf("Failed to delete remote FlinkDeployment: %v", err))
					return ctrl.Result{}, err
				}
				logger.Info("Remote FlinkDeployment already deleted or not found", "deploymentName", deployment.Name)
			}

			controllerutil.RemoveFinalizer(deployment, finalizerName)
			if err := r.Update(ctx, deployment); err != nil {
				logger.Error(err, "Failed to remove finalizer from FlinkDeployment")
				return ctrl.Result{}, err
			}
			logger.Info("Successfully removed finalizer from FlinkDeployment", "deploymentName", deployment.Name)
		}
		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(deployment, finalizerName) {
		controllerutil.AddFinalizer(deployment, finalizerName)
		if err := r.Update(ctx, deployment); err != nil {
			r.updateDeploymentStatus(ctx, deployment, err, "UpdateFinalizerFailed", fmt.Sprintf("Failed to add finalizer: %v", err))
			return ctrl.Result{}, err
		}
		logger.Info("Successfully added finalizer to FlinkDeployment", "deploymentName", deployment.Name)
		return ctrl.Result{Requeue: true}, nil
	}

	existingRemoteDeployment, err := deploymentClient.GetFlinkDeployment(ctx, deployment.Name)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			r.updateDeploymentStatus(ctx, deployment, err, "GetRemoteDeploymentFailed", fmt.Sprintf("Failed to get remote FlinkDeployment: %v", err))
			return ctrl.Result{}, err
		}
		existingRemoteDeployment = nil
	}

	if existingRemoteDeployment == nil {
		logger.Info("Remote FlinkDeployment not found, creating new one.", "deploymentName", deployment.Name)
		createdRemoteDeployment, err := deploymentClient.CreateFlinkDeployment(ctx, deployment)
		if err != nil {
			r.updateDeploymentStatus(ctx, deployment, err, "CreateRemoteDeploymentFailed", fmt.Sprintf("Failed to create remote FlinkDeployment: %v", err))
			return ctrl.Result{}, err
		}
		logger.Info("Successfully created remote FlinkDeployment.", "deploymentName", deployment.Name)

		statusBytes, err := json.Marshal(createdRemoteDeployment.Status)
		if err != nil {
			r.updateDeploymentStatus(ctx, deployment, err, "StatusMarshalFailed", fmt.Sprintf("Failed to marshal created FlinkDeployment status: %v", err))
			return ctrl.Result{}, err
		}
		deployment.Status.DeploymentStatus = &runtime.RawExtension{Raw: statusBytes}
		r.updateDeploymentStatus(ctx, deployment, nil, "Ready", "FlinkDeployment created and synced successfully")
	} else {
		logger.Info("Remote FlinkDeployment found, attempting to update.", "deploymentName", deployment.Name)
		updatedRemoteDeployment, err := deploymentClient.UpdateFlinkDeployment(ctx, deployment)
		if err != nil {
			r.updateDeploymentStatus(ctx, deployment, err, "UpdateRemoteDeploymentFailed", fmt.Sprintf("Failed to update remote FlinkDeployment: %v", err))
			return ctrl.Result{}, err
		}
		logger.Info("Successfully updated remote FlinkDeployment.", "deploymentName", deployment.Name)

		statusBytes, err := json.Marshal(updatedRemoteDeployment.Status)
		if err != nil {
			r.updateDeploymentStatus(ctx, deployment, err, "StatusMarshalFailed", fmt.Sprintf("Failed to marshal updated FlinkDeployment status: %v", err))
			return ctrl.Result{}, err
		}
		deployment.Status.DeploymentStatus = &runtime.RawExtension{Raw: statusBytes}
		r.updateDeploymentStatus(ctx, deployment, nil, "Ready", "FlinkDeployment updated and synced successfully")
	}

	logger.Info("Successfully reconciled FlinkDeployment", "deploymentName", deployment.Name)
	return ctrl.Result{RequeueAfter: requeueInterval}, nil
}

func (r *FlinkDeploymentReconciler) updateDeploymentStatus(
	ctx context.Context,
	deployment *resourcev1alpha1.ComputeFlinkDeployment,
	errEncountered error,
	reason string,
	message string,
) {
	logger := log.FromContext(ctx)

	currentStatus := deployment.Status.DeepCopy()

	condition := metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: deployment.Generation,
		LastTransitionTime: metav1.Now(),
	}

	if errEncountered != nil {
		logger.Error(errEncountered, message, "flinkDeploymentName", deployment.Name)
		condition.Status = metav1.ConditionFalse
	}

	newConditions := []metav1.Condition{}
	foundReady := false
	for _, c := range currentStatus.Conditions {
		if c.Type == condition.Type {
			if c.ObservedGeneration <= condition.ObservedGeneration {
				newConditions = append(newConditions, condition)
				foundReady = true
			} else {
				newConditions = append(newConditions, c)
				foundReady = true
			}
		} else {
			newConditions = append(newConditions, c)
		}
	}
	if !foundReady {
		newConditions = append(newConditions, condition)
	}
	currentStatus.Conditions = newConditions

	deployment.Status = *currentStatus

	if err := r.Status().Update(ctx, deployment); err != nil {
		logger.Error(err, "Failed to update FlinkDeployment status", "flinkDeploymentName", deployment.Name)
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *FlinkDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&resourcev1alpha1.ComputeFlinkDeployment{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}
