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
	"encoding/base64"
	"fmt"
	"time"

	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	controllers2 "github.com/streamnative/pulsar-resources-operator/pkg/streamnativecloud"
	"github.com/streamnative/pulsar-resources-operator/pkg/utils"

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

// ServiceAccountReconciler reconciles a StreamNative Cloud ServiceAccount object
type ServiceAccountReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	ConnectionManager *ConnectionManager
}

const ServiceAccountFinalizer = "serviceaccount.resource.streamnative.io/finalizer"

//+kubebuilder:rbac:groups=resource.streamnative.io,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=serviceaccounts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=serviceaccounts/finalizers,verbs=update
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=streamnativecloudconnections,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

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
			logger.Info("ServiceAccount not found. Reconciliation will stop.", "namespace", req.Namespace, "name", req.Name)
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
		if controllerutil.ContainsFinalizer(serviceAccount, ServiceAccountFinalizer) {
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
			controllerutil.RemoveFinalizer(serviceAccount, ServiceAccountFinalizer)
			if err := r.Update(ctx, serviceAccount); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer if it doesn't exist
	if !controllerutil.ContainsFinalizer(serviceAccount, ServiceAccountFinalizer) {
		controllerutil.AddFinalizer(serviceAccount, ServiceAccountFinalizer)
		if err := r.Update(ctx, serviceAccount); err != nil {
			return ctrl.Result{}, err
		}
		// Requeue after adding finalizer to ensure the update is processed before proceeding
		return ctrl.Result{Requeue: true}, nil
	}

	// Check if ServiceAccount exists
	existingSA, err := saClient.GetServiceAccount(ctx, serviceAccount.Name)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			logger.Info("Failed to get ServiceAccount", "error", err, "existingSA", existingSA)
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

			// If the credentials type is ServiceAccountCredentialsType, create a Secret
			if resultSA.Status.PrivateKeyType == utils.ServiceAccountCredentialsType {
				credentialsData, err := base64.StdEncoding.DecodeString(resultSA.Status.PrivateKeyData)
				if err != nil {
					logger.Error(err, "Failed to decode private key data")
				} else {
					if err := utils.CreateOrUpdateServiceAccountCredentialsSecret(ctx, r.Client, serviceAccount, serviceAccount.Namespace, serviceAccount.Name, string(credentialsData)); err != nil {
						logger.Error(err, "Failed to create or update service account credentials secret")
					}
				}
			}
		}

		// Update status
		r.updateServiceAccountStatus(ctx, serviceAccount, nil, "Ready", "ServiceAccount created successfully")
	} else {
		// Update status with private key information from existing ServiceAccount
		if existingSA.Status.PrivateKeyType != "" {
			serviceAccount.Status.PrivateKeyType = existingSA.Status.PrivateKeyType
		}
		if existingSA.Status.PrivateKeyData != "" {
			serviceAccount.Status.PrivateKeyData = existingSA.Status.PrivateKeyData

			// If the credentials type is ServiceAccountCredentialsType, ensure Secret exists
			if existingSA.Status.PrivateKeyType == utils.ServiceAccountCredentialsType {
				credentialsData, err := base64.StdEncoding.DecodeString(existingSA.Status.PrivateKeyData)
				if err != nil {
					logger.Error(err, "Failed to decode private key data")
				} else {
					if err := utils.CreateOrUpdateServiceAccountCredentialsSecret(ctx, r.Client, serviceAccount, serviceAccount.Namespace, serviceAccount.Name, string(credentialsData)); err != nil {
						logger.Error(err, "Failed to create or update service account credentials secret")
					}
				}
			}
		}

		// Update status
		r.updateServiceAccountStatus(ctx, serviceAccount, nil, "Ready", "ServiceAccount synced successfully")
		logger.Info("ServiceAccount reconciled", "namespace", serviceAccount.Namespace, "name", serviceAccount.Name)
	}

	return ctrl.Result{RequeueAfter: requeueInterval}, nil
}

// updateServiceAccountStatus updates the status of the ServiceAccount resource
func (r *ServiceAccountReconciler) updateServiceAccountStatus(
	ctx context.Context,
	serviceAccount *resourcev1alpha1.ServiceAccount,
	err error,
	reason string,
	message string,
) {
	logger := log.FromContext(ctx)
	serviceAccount.Status.ObservedGeneration = serviceAccount.Generation

	// Set ready condition based on error
	meta := metav1.Now()
	readyCondition := metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: meta,
	}

	if err != nil {
		readyCondition.Status = metav1.ConditionFalse
		logger.Error(err, "ServiceAccount reconciliation failed",
			"reason", reason, "message", message)
	}

	// Update ready condition
	if len(serviceAccount.Status.Conditions) == 0 {
		serviceAccount.Status.Conditions = []metav1.Condition{readyCondition}
	} else {
		for i, condition := range serviceAccount.Status.Conditions {
			if condition.Type == "Ready" {
				// Only update time if status changes
				if condition.Status != readyCondition.Status {
					readyCondition.LastTransitionTime = meta
				} else {
					readyCondition.LastTransitionTime = condition.LastTransitionTime
				}
				serviceAccount.Status.Conditions[i] = readyCondition
				break
			}
		}
	}

	// Update ServiceAccount status
	if err := r.Status().Update(ctx, serviceAccount); err != nil {
		logger.Error(err, "Failed to update ServiceAccount status")
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceAccountReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&resourcev1alpha1.ServiceAccount{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}
