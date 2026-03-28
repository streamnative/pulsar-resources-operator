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

	"github.com/pkg/errors"
	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	controllers2 "github.com/streamnative/pulsar-resources-operator/pkg/streamnativecloud"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// ServiceAccountBindingReconciler reconciles a StreamNative Cloud ServiceAccountBinding object
type ServiceAccountBindingReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	ConnectionManager *ConnectionManager
}

const ServiceAccountBindingFinalizer = "serviceaccountbinding.resource.streamnative.io/finalizer"

//+kubebuilder:rbac:groups=resource.streamnative.io,resources=serviceaccountbindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=serviceaccountbindings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=serviceaccountbindings/finalizers,verbs=update
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=serviceaccounts,verbs=get;list;watch
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=streamnativecloudconnections,verbs=get;list;watch

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
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if !binding.DeletionTimestamp.IsZero() && shouldKeepRemoteResource(binding.Spec.LifecyclePolicy) {
		return finalizeKeptResource(ctx, r.Client, binding, ServiceAccountBindingFinalizer, "ServiceAccountBinding")
	}

	// Get the ServiceAccount
	serviceAccount := &resourcev1alpha1.ServiceAccount{}
	saErr := r.Get(ctx, types.NamespacedName{
		Namespace: req.Namespace,
		Name:      binding.Spec.ServiceAccountName,
	}, serviceAccount)

	// If binding is being deleted and SA not found, try to use binding's own APIServerRef
	// or skip remote cleanup entirely
	if !binding.DeletionTimestamp.IsZero() && apierrors.IsNotFound(saErr) {
		if binding.Spec.APIServerRef == nil || binding.Spec.APIServerRef.Name == "" {
			// Cannot resolve connection, remove finalizer directly
			logger.Info("ServiceAccount and own APIServerRef not found during deletion, removing finalizer without remote cleanup",
				"serviceAccount", binding.Spec.ServiceAccountName)
			if controllerutil.ContainsFinalizer(binding, ServiceAccountBindingFinalizer) {
				controllerutil.RemoveFinalizer(binding, ServiceAccountBindingFinalizer)
				if err := r.Update(ctx, binding); err != nil {
					return ctrl.Result{}, err
				}
			}
			return ctrl.Result{}, nil
		}
		// Fall through using binding's own APIServerRef
	} else if saErr != nil {
		r.updateServiceAccountBindingStatus(ctx, binding, saErr, "ServiceAccountNotFound",
			fmt.Sprintf("Failed to get ServiceAccount %s: %v", binding.Spec.ServiceAccountName, saErr))
		return ctrl.Result{}, saErr
	}

	// Check if the referenced ServiceAccount is ready (skip during deletion)
	if binding.DeletionTimestamp.IsZero() && saErr == nil {
		saReady := false
		for _, cond := range serviceAccount.Status.Conditions {
			if cond.Type == "Ready" && cond.Status == metav1.ConditionTrue {
				saReady = true
				break
			}
		}
		if !saReady {
			errMsg := fmt.Sprintf("ServiceAccount %s is not yet Ready. Will requeue.", serviceAccount.Name)
			logger.Info(errMsg)
			r.updateServiceAccountBindingStatus(ctx, binding, errors.New(errMsg), "ServiceAccountNotReady", errMsg)
			return ctrl.Result{RequeueAfter: requeueInterval}, nil
		}
	}

	// Determine which APIServerRef to use
	apiServerRefName := ""
	if binding.Spec.APIServerRef != nil && binding.Spec.APIServerRef.Name != "" {
		apiServerRefName = binding.Spec.APIServerRef.Name
	} else if saErr == nil && serviceAccount.Spec.APIServerRef.Name != "" {
		apiServerRefName = serviceAccount.Spec.APIServerRef.Name
	} else {
		err := fmt.Errorf("APIServerRef not found in ServiceAccountBinding spec or its referenced ServiceAccount spec")
		r.updateServiceAccountBindingStatus(ctx, binding, err, "ValidationFailed", err.Error())
		return ctrl.Result{}, err
	}

	// Get the APIServerConnection
	connection := &resourcev1alpha1.StreamNativeCloudConnection{}
	connErr := r.Get(ctx, types.NamespacedName{
		Namespace: req.Namespace,
		Name:      apiServerRefName,
	}, connection)
	if !binding.DeletionTimestamp.IsZero() && apierrors.IsNotFound(connErr) {
		logger.Info("Connection not found during deletion, removing finalizer without remote cleanup",
			"connection", apiServerRefName)
		if controllerutil.ContainsFinalizer(binding, ServiceAccountBindingFinalizer) {
			controllerutil.RemoveFinalizer(binding, ServiceAccountBindingFinalizer)
			if err := r.Update(ctx, binding); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}
	if connErr != nil {
		r.updateServiceAccountBindingStatus(ctx, binding, connErr, "ConnectionNotFound",
			fmt.Sprintf("Failed to get StreamNativeCloudConnection %s: %v", apiServerRefName, connErr))
		return ctrl.Result{}, connErr
	}

	// Get API connection
	apiConn, err := r.ConnectionManager.GetOrCreateConnection(connection, nil)
	if err != nil {
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
		err := fmt.Errorf("organization is required in StreamNativeCloudConnection %s but not specified", connection.Name)
		r.updateServiceAccountBindingStatus(ctx, binding, err, "ValidationFailed", err.Error())
		return ctrl.Result{}, err // No need to requeue if this is a permanent misconfiguration
	}

	// Create ServiceAccountBinding client
	bindingClient, err := controllers2.NewServiceAccountBindingClient(apiConn, organization)
	if err != nil {
		r.updateServiceAccountBindingStatus(ctx, binding, err, "ClientCreationFailed",
			fmt.Sprintf("Failed to create ServiceAccountBinding client: %v", err))
		return ctrl.Result{}, err
	}

	// Validate PoolMemberRefs
	if len(binding.Spec.PoolMemberRefs) == 0 {
		err := fmt.Errorf("at least one poolMemberRef is required in spec.poolMemberRefs")
		r.updateServiceAccountBindingStatus(ctx, binding, err, "ValidationFailed", err.Error())
		return ctrl.Result{}, err // No need to requeue if this is a permanent misconfiguration
	}

	// Handle deletion
	if !binding.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(binding, ServiceAccountBindingFinalizer) {
			for i, poolMemberRef := range binding.Spec.PoolMemberRefs {
				remoteName := fmt.Sprintf("%s.%s.%s", binding.Spec.ServiceAccountName, poolMemberRef.Namespace, poolMemberRef.Name)
				if err := bindingClient.DeleteServiceAccountBinding(ctx, &resourcev1alpha1.ServiceAccountBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name: remoteName, // This should ideally match the name used when creating/getting
					},
				}); err != nil {
					if !apierrors.IsNotFound(err) {
						r.updateServiceAccountBindingStatus(ctx, binding, err, "DeleteFailed",
							fmt.Sprintf("Failed to delete remote ServiceAccountBinding for PoolMemberRef %d (%s): %v", i, remoteName, err))
						return ctrl.Result{}, err
					}
					logger.Info("Remote ServiceAccountBinding already deleted or not found",
						"binding", remoteName, "poolMemberRef", poolMemberRef)
				}
			}

			controllerutil.RemoveFinalizer(binding, ServiceAccountBindingFinalizer)
			if err := r.Update(ctx, binding); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer
	if !controllerutil.ContainsFinalizer(binding, ServiceAccountBindingFinalizer) {
		controllerutil.AddFinalizer(binding, ServiceAccountBindingFinalizer)
		if err := r.Update(ctx, binding); err != nil {
			r.updateServiceAccountBindingStatus(ctx, binding, err, "UpdateFinalizerFailed",
				fmt.Sprintf("Failed to add finalizer: %v", err))
			return ctrl.Result{}, err
		}
		// Requeue after adding finalizer to ensure the update is processed before proceeding
		return ctrl.Result{Requeue: true}, nil
	}

	// Reconcile each PoolMemberRef
	allBindingsReady := true
	var lastError error
	for i, poolMemberRef := range binding.Spec.PoolMemberRefs {
		remoteName := fmt.Sprintf("%s.%s.%s", binding.Spec.ServiceAccountName, poolMemberRef.Namespace, poolMemberRef.Name)

		// This is the payload for the client's CreateServiceAccountBinding method
		// It should be of type *resourcev1alpha1.ServiceAccountBinding
		payloadForClient := &resourcev1alpha1.ServiceAccountBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      remoteName,        // This name will be used by the client's convertToCloudServiceAccountBinding
				Namespace: binding.Namespace, // Preserve original namespace, client might use it or organization
			},
			Spec: resourcev1alpha1.ServiceAccountBindingSpec{
				ServiceAccountName: binding.Spec.ServiceAccountName,
				PoolMemberRefs:     []resourcev1alpha1.PoolMemberReference{poolMemberRef}, // Single ref for this specific call
				// APIServerRef is not directly used by convertToCloudServiceAccountBinding for spec, but keep for completeness if other logic depends on it
				APIServerRef: binding.Spec.APIServerRef,
			},
		}

		// Check if the remote binding already exists
		existingRemoteBinding, err := bindingClient.GetServiceAccountBinding(ctx, remoteName)
		if err != nil {
			if apierrors.IsNotFound(err) {
				// Remote binding does not exist, so create it.
				if _, err := bindingClient.CreateServiceAccountBinding(ctx, payloadForClient); err != nil {
					errMsg := fmt.Sprintf("Failed to create remote ServiceAccountBinding for PoolMemberRef %d (%s): %v", i, remoteName, err)
					r.updateServiceAccountBindingStatus(ctx, binding, err, "CreateFailed", errMsg)
					allBindingsReady = false
					lastError = errors.Wrapf(err, "Failed to create remote ServiceAccountBinding for PoolMemberRef %d (%s)", i, remoteName)
					continue
				}
				logger.Info("Successfully created remote ServiceAccountBinding", "bindingName", remoteName, "poolMemberRef", poolMemberRef)
			} else {
				// Another error occurred while trying to get the remote binding.
				errMsg := fmt.Sprintf("Failed to get remote ServiceAccountBinding for PoolMemberRef %d (%s): %v", i, remoteName, err)
				r.updateServiceAccountBindingStatus(ctx, binding, err, "GetFailed", errMsg)
				allBindingsReady = false
				lastError = errors.Wrapf(err, "Failed to get remote ServiceAccountBinding for PoolMemberRef %d (%s)", i, remoteName)
				continue
			}
		} else {
			// Remote binding exists.
			logger.Info("Remote ServiceAccountBinding already exists", "bindingName", remoteName, "poolMemberRef", poolMemberRef, "existingRemoteName", existingRemoteBinding.ObjectMeta.Name)
			// TODO: Implement update logic if necessary.
			// Compare existingRemoteBinding.Spec with what payloadForClient would generate via conversion.
			// For now, we assume if it exists, it's correctly configured or updates are not handled here.
		}
	}

	if !allBindingsReady {
		// If any binding failed, the overall status is not Ready.
		// The status message will reflect the last error encountered during the loop for simplicity.
		// A more sophisticated approach might collect all errors.
		r.updateServiceAccountBindingStatus(ctx, binding, lastError, "Reconciling", "Some ServiceAccountBindings are still being processed or encountered errors. See last error for details.")
		return ctrl.Result{RequeueAfter: requeueInterval}, nil
	}

	r.updateServiceAccountBindingStatus(ctx, binding, nil, "Ready", "All ServiceAccountBindings synced successfully")
	logger.Info("Successfully reconciled ServiceAccountBinding")
	return ctrl.Result{RequeueAfter: requeueInterval}, nil
}

func (r *ServiceAccountBindingReconciler) updateServiceAccountBindingStatus(
	ctx context.Context,
	binding *resourcev1alpha1.ServiceAccountBinding,
	err error,
	reason string,
	message string,
) {
	logger := log.FromContext(ctx)
	binding.Status.ObservedGeneration = binding.Generation
	condition := metav1.Condition{
		Type:               "Ready",
		ObservedGeneration: binding.Generation,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	}
	if err != nil {
		logger.Error(err, message, "serviceAccountBindingName", binding.Name)
		condition.Status = metav1.ConditionFalse
	} else {
		condition.Status = metav1.ConditionTrue
	}

	newConditions := []metav1.Condition{}
	foundReady := false
	for _, c := range binding.Status.Conditions {
		if c.Type == "Ready" {
			// If we're providing an update for the current generation, replace the old Ready condition.
			if c.ObservedGeneration == condition.ObservedGeneration {
				newConditions = append(newConditions, condition)
				foundReady = true
			} else if c.ObservedGeneration < condition.ObservedGeneration {
				// New condition is for a newer generation, replace old one
				newConditions = append(newConditions, condition)
				foundReady = true
			} else {
				// Old condition is for a newer generation (should not happen if logic is correct, but keep it)
				newConditions = append(newConditions, c)
				foundReady = true // Mark as found, so we don't add the new one if old is newer
			}
		} else {
			newConditions = append(newConditions, c)
		}
	}
	if !foundReady {
		newConditions = append(newConditions, condition)
	}
	binding.Status.Conditions = newConditions

	if statusUpdateErr := r.Status().Update(ctx, binding); statusUpdateErr != nil {
		logger.Error(statusUpdateErr, "Failed to update ServiceAccountBinding status", "serviceAccountBindingName", binding.Name)
	}
}

func (r *ServiceAccountBindingReconciler) findResourcesForConnection(ctx context.Context, obj client.Object) []ctrl.Request {
	connName := obj.GetName()
	ns := obj.GetNamespace()

	bindings := &resourcev1alpha1.ServiceAccountBindingList{}
	if err := r.List(ctx, bindings, client.InNamespace(ns)); err != nil {
		return nil
	}

	// Find service accounts referencing this connection (for indirect lookups)
	serviceAccounts := &resourcev1alpha1.ServiceAccountList{}
	connServiceAccounts := map[string]bool{}
	if err := r.List(ctx, serviceAccounts, client.InNamespace(ns)); err == nil {
		for i := range serviceAccounts.Items {
			if serviceAccounts.Items[i].Spec.APIServerRef.Name == connName {
				connServiceAccounts[serviceAccounts.Items[i].Name] = true
			}
		}
	}

	requests := make([]ctrl.Request, 0, len(bindings.Items))
	for i := range bindings.Items {
		b := &bindings.Items[i]
		// Direct reference
		if b.Spec.APIServerRef != nil && b.Spec.APIServerRef.Name == connName {
			requests = append(requests, ctrl.Request{
				NamespacedName: types.NamespacedName{
					Namespace: b.Namespace,
					Name:      b.Name,
				},
			})
			continue
		}
		// Indirect reference via service account
		if (b.Spec.APIServerRef == nil || b.Spec.APIServerRef.Name == "") && connServiceAccounts[b.Spec.ServiceAccountName] {
			requests = append(requests, ctrl.Request{
				NamespacedName: types.NamespacedName{
					Namespace: b.Namespace,
					Name:      b.Name,
				},
			})
		}
	}
	return requests
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceAccountBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&resourcev1alpha1.ServiceAccountBinding{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Watches(&resourcev1alpha1.StreamNativeCloudConnection{},
			handler.EnqueueRequestsFromMapFunc(r.findResourcesForConnection)).
		Complete(r)
}
