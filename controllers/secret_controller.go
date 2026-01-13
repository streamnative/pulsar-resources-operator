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
	cloudapi "github.com/streamnative/pulsar-resources-operator/pkg/streamnativecloud" // Standard alias for cloud client package

	corev1 "k8s.io/api/core/v1"
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

// SecretReconciler reconciles a StreamNative Cloud Secret object
type SecretReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	ConnectionManager *ConnectionManager
}

//+kubebuilder:rbac:groups=resource.streamnative.io,resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=secrets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=secrets/finalizers,verbs=update
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=streamnativecloudconnections,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

// getSecretData obtains the Secret data either from direct Data field or from SecretRef
func (r *SecretReconciler) getSecretData(ctx context.Context, secretCR *resourcev1alpha1.Secret) (map[string]string, *corev1.SecretType, error) {
	// If direct data is provided, use it
	if len(secretCR.Spec.Data) > 0 {
		return secretCR.Spec.Data, secretCR.Spec.Type, nil
	}

	// If SecretRef is provided, fetch from the referenced Kubernetes Secret
	if secretCR.Spec.SecretRef != nil {
		nsName := secretCR.Spec.SecretRef.ToNamespacedName()
		k8sSecret := &corev1.Secret{}
		if err := r.Get(ctx, nsName, k8sSecret); err != nil {
			return nil, nil, fmt.Errorf("failed to get referenced Secret %s/%s: %w", nsName.Namespace, nsName.Name, err)
		}

		stringData := make(map[string]string)
		for k, v := range k8sSecret.Data {
			stringData[k] = string(v)
		}
		return stringData, &k8sSecret.Type, nil
	}

	return nil, nil, fmt.Errorf("neither Data nor SecretRef is specified in the Secret spec")
}

// Reconcile handles the reconciliation of Secret objects
func (r *SecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling Secret", "namespace", req.Namespace, "name", req.Name)

	requeueInterval := 1 * time.Minute // Default requeue interval

	secretCR := &resourcev1alpha1.Secret{}
	if err := r.Get(ctx, req.NamespacedName, secretCR); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Secret resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get Secret resource")
		return ctrl.Result{}, err
	}

	// Validate APIServerRef
	if secretCR.Spec.APIServerRef.Name == "" {
		err := fmt.Errorf("APIServerRef.Name is required in Secret spec but not specified")
		r.updateSecretStatus(ctx, secretCR, err, "ValidationFailed", err.Error())
		return ctrl.Result{}, err // No requeue for permanent misconfiguration
	}

	// Get StreamNativeCloudConnection
	connection := &resourcev1alpha1.StreamNativeCloudConnection{}
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: req.Namespace, // Assuming connection is in the same namespace
		Name:      secretCR.Spec.APIServerRef.Name,
	}, connection); err != nil {
		r.updateSecretStatus(ctx, secretCR, err, "ConnectionNotFound",
			fmt.Sprintf("Failed to get StreamNativeCloudConnection %s: %v", secretCR.Spec.APIServerRef.Name, err))
		return ctrl.Result{}, err
	}

	// Get or create API connection
	apiConn, err := r.ConnectionManager.GetOrCreateConnection(connection, nil)
	if err != nil {
		if _, ok := err.(*NotInitializedError); ok {
			logger.Info("Connection not initialized, requeueing", "connectionName", connection.Name, "error", err.Error())
			return ctrl.Result{Requeue: true}, nil
		}
		r.updateSecretStatus(ctx, secretCR, err, "GetConnectionFailed",
			fmt.Sprintf("Failed to get active connection for %s: %v", connection.Name, err))
		return ctrl.Result{}, err
	}

	// Validate Organization in connection spec
	if connection.Spec.Organization == "" {
		err := fmt.Errorf("organization is required in StreamNativeCloudConnection %s but not specified", connection.Name)
		r.updateSecretStatus(ctx, secretCR, err, "ValidationFailed", err.Error())
		return ctrl.Result{}, err
	}

	// Create Secret client
	secretClient, err := cloudapi.NewSecretClient(apiConn, connection.Spec.Organization)
	if err != nil {
		r.updateSecretStatus(ctx, secretCR, err, "ClientCreationFailed",
			fmt.Sprintf("Failed to create Secret client: %v", err))
		return ctrl.Result{}, err
	}

	finalizerName := cloudapi.SecretFinalizer
	// Handle deletion
	if !secretCR.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(secretCR, finalizerName) {
			if err := secretClient.DeleteSecret(ctx, secretCR); err != nil {
				// If the remote secret is already gone, that's okay.
				if !apierrors.IsNotFound(err) {
					r.updateSecretStatus(ctx, secretCR, err, "DeleteRemoteSecretFailed", fmt.Sprintf("Failed to delete remote Secret: %v", err))
					return ctrl.Result{}, err
				}
			}

			controllerutil.RemoveFinalizer(secretCR, finalizerName)
			if err := r.Update(ctx, secretCR); err != nil {
				logger.Error(err, "Failed to remove finalizer from Secret")
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(secretCR, finalizerName) {
		controllerutil.AddFinalizer(secretCR, finalizerName)
		if err := r.Update(ctx, secretCR); err != nil {
			r.updateSecretStatus(ctx, secretCR, err, "AddFinalizerFailed", fmt.Sprintf("Failed to add finalizer: %v", err))
			return ctrl.Result{}, err
		}
		// Requeue after adding finalizer to ensure the update is processed before proceeding
		return ctrl.Result{Requeue: true}, nil
	}

	// Resolve secret data from Spec.Data or Spec.SecretRef
	// The secretCR passed to cloud client methods should have its Spec.Data and Spec.Type populated.
	currentSpecData := make(map[string]string)
	for k, v := range secretCR.Spec.Data {
		currentSpecData[k] = v
	}
	currentSpecType := secretCR.Spec.Type

	// If SecretRef is used, we resolve it and update the CR if necessary.
	// This ensures that the CR in etcd reflects the data being sent to the cloud API.
	if secretCR.Spec.SecretRef != nil {
		resolvedData, resolvedType, err := r.getSecretData(ctx, secretCR)
		if err != nil {
			r.updateSecretStatus(ctx, secretCR, err, "GetSecretDataFailed", fmt.Sprintf("Failed to get secret data from SecretRef: %v", err))
			return ctrl.Result{}, err
		}

		// Check if an update to the local CR is needed
		updateLocalCR := false
		if len(secretCR.Spec.Data) == 0 { // Only populate from SecretRef if direct data is not set
			secretCR.Spec.Data = resolvedData
			updateLocalCR = true
		}
		// Only update type if original spec type was nil or empty, and resolvedType is not nil
		if (secretCR.Spec.Type == nil || *secretCR.Spec.Type == "") && resolvedType != nil {
			secretCR.Spec.Type = resolvedType
			updateLocalCR = true
		}

		if updateLocalCR {
			if err := r.Update(ctx, secretCR); err != nil {
				// Restore original spec data before status update to avoid inconsistent state reporting
				secretCR.Spec.Data = currentSpecData
				secretCR.Spec.Type = currentSpecType
				r.updateSecretStatus(ctx, secretCR, err, "UpdateLocalSecretFailed",
					fmt.Sprintf("Failed to update local Secret CR with resolved data: %v", err))
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil // Requeue to use the updated CR
		}
	}

	existingRemoteSecret, err := secretClient.GetSecret(ctx, secretCR.Name)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			r.updateSecretStatus(ctx, secretCR, err, "GetRemoteSecretFailed", fmt.Sprintf("Failed to get remote Secret: %v", err))
			return ctrl.Result{}, err
		}
		existingRemoteSecret = nil
	}

	if existingRemoteSecret == nil {
		if _, err := secretClient.CreateSecret(ctx, secretCR); err != nil {
			r.updateSecretStatus(ctx, secretCR, err, "CreateRemoteSecretFailed", fmt.Sprintf("Failed to create remote Secret: %v", err))
			return ctrl.Result{}, err
		}
		r.updateSecretStatus(ctx, secretCR, nil, "Ready", "Secret created and synced successfully")
	} else {
		if _, err := secretClient.UpdateSecret(ctx, secretCR); err != nil { // Pass the K8s CR
			r.updateSecretStatus(ctx, secretCR, err, "UpdateRemoteSecretFailed", fmt.Sprintf("Failed to update remote Secret: %v", err))
			return ctrl.Result{}, err
		}
		r.updateSecretStatus(ctx, secretCR, nil, "Ready", "Secret updated and synced successfully")
	}

	logger.Info("Successfully reconciled Secret", "namespace", req.Namespace, "name", req.Name)
	return ctrl.Result{RequeueAfter: requeueInterval}, nil
}

func (r *SecretReconciler) updateSecretStatus(
	ctx context.Context,
	secretCR *resourcev1alpha1.Secret, // Changed to secretCR for clarity
	errEncountered error,
	reason string,
	message string,
) {
	logger := log.FromContext(ctx)
	statusChanged := false

	// Prepare the new condition
	newCondition := metav1.Condition{
		Type:               resourcev1alpha1.ConditionReady,
		Status:             metav1.ConditionFalse,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: secretCR.Generation,
		LastTransitionTime: metav1.Now(),
	}
	if errEncountered == nil {
		newCondition.Status = metav1.ConditionTrue
	} else {
		logger.Error(errEncountered, message, "secretName", secretCR.Name)
	}

	// Check existing conditions
	existingCondition := findCondition(secretCR.Status.Conditions, resourcev1alpha1.ConditionReady)
	if existingCondition == nil {
		secretCR.Status.Conditions = append(secretCR.Status.Conditions, newCondition)
		statusChanged = true
	} else if existingCondition.Status != newCondition.Status ||
		existingCondition.Reason != newCondition.Reason ||
		existingCondition.Message != newCondition.Message ||
		existingCondition.ObservedGeneration != newCondition.ObservedGeneration {
		existingCondition.Status = newCondition.Status
		existingCondition.Reason = newCondition.Reason
		existingCondition.Message = newCondition.Message
		existingCondition.ObservedGeneration = newCondition.ObservedGeneration
		existingCondition.LastTransitionTime = newCondition.LastTransitionTime
		statusChanged = true
	}

	if statusChanged {
		err := r.Status().Update(ctx, secretCR)
		if err != nil {
			logger.Error(err, "Failed to update Secret status", "secretName", secretCR.Name)
		} else {
			logger.Info("Successfully updated Secret status", "secretName", secretCR.Name, "reason", reason)
		}
	}
}

// findCondition finds a condition of a specific type in a list of conditions.
// Returns nil if not found.
func findCondition(conditions []metav1.Condition, conditionType string) *metav1.Condition {
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return &conditions[i]
		}
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&resourcev1alpha1.Secret{}).
		Owns(&corev1.Secret{}).                                  // If we want to react to changes in owned k8s secrets
		WithEventFilter(predicate.GenerationChangedPredicate{}). // Only reconcile on spec changes or finalizer changes
		Complete(r)
}
