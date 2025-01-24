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

	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// APIServerConnectionReconciler reconciles a APIServerConnection object
type APIServerConnectionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	// ConnectionManager manages the API server connections
	ConnectionManager *ConnectionManager
}

//+kubebuilder:rbac:groups=resource.streamnative.io,resources=streamnativecloudconnections,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=streamnativecloudconnections/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=streamnativecloudconnections/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

func (r *APIServerConnectionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling APIServerConnection", "namespace", req.Namespace, "name", req.Name)

	// Get the APIServerConnection resource
	connection := &resourcev1alpha1.StreamNativeCloudConnection{}
	if err := r.Get(ctx, req.NamespacedName, connection); err != nil {
		if apierrors.IsNotFound(err) {
			// Delete connection if it exists
			if err := r.ConnectionManager.CloseConnection(req.Name); err != nil {
				logger.Error(err, "Failed to close connection")
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Get the credentials secret
	secret := &corev1.Secret{}
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: connection.Namespace,
		Name:      connection.Spec.Auth.CredentialsRef.Name,
	}, secret); err != nil {
		if apierrors.IsNotFound(err) {
			msg := fmt.Sprintf("Credentials secret %s/%s not found", connection.Namespace, connection.Spec.Auth.CredentialsRef.Name)
			r.updateConnectionStatus(ctx, connection, err, "CredentialsNotFound", msg)
			return ctrl.Result{}, fmt.Errorf(msg)
		}
		return ctrl.Result{}, err
	}

	// Get the credentials from secret
	credentialsData, ok := secret.Data["credentials.json"]
	if !ok {
		err := fmt.Errorf("credentials.json not found in secret %s", secret.Name)
		r.updateConnectionStatus(ctx, connection, err, "InvalidCredentials",
			"credentials.json not found in secret")
		return ctrl.Result{}, err
	}

	// Parse credentials
	credentials := &resourcev1alpha1.ServiceAccountCredentials{}
	if err := json.Unmarshal(credentialsData, credentials); err != nil {
		r.updateConnectionStatus(ctx, connection, err, "InvalidCredentials",
			fmt.Sprintf("Failed to parse credentials: %v", err))
		return ctrl.Result{}, err
	}

	// Validate credentials
	if err := r.validateCredentials(credentials); err != nil {
		r.updateConnectionStatus(ctx, connection, err, "InvalidCredentials",
			fmt.Sprintf("Invalid credentials: %v", err))
		return ctrl.Result{}, err
	}

	logger.Info("Validated credentials", "client_id", credentials.ClientID)

	// Get or create connection
	conn, err := r.ConnectionManager.GetOrCreateConnection(connection, credentials)
	if err != nil {
		// If connection is not initialized, requeue the request
		if _, ok := err.(*NotInitializedError); ok {
			logger.Info("Connection not initialized, requeueing", "error", err.Error())
			return ctrl.Result{Requeue: true}, nil
		}
		r.updateConnectionStatus(ctx, connection, err, "ConnectionFailed",
			fmt.Sprintf("Failed to create connection: %v", err))
		return ctrl.Result{}, err
	}

	// Test connection
	if err := conn.Test(ctx); err != nil {
		r.updateConnectionStatus(ctx, connection, err, "TestFailed",
			fmt.Sprintf("Failed to test connection: %v", err))
		return ctrl.Result{}, err
	}

	// Update status
	r.updateConnectionStatus(ctx, connection, nil, "Ready", "Connection established successfully")

	return ctrl.Result{}, nil
}

func (r *APIServerConnectionReconciler) validateCredentials(creds *resourcev1alpha1.ServiceAccountCredentials) error {
	if creds.Type != "sn_service_account" {
		return fmt.Errorf("invalid credentials type: %s", creds.Type)
	}
	if creds.ClientID == "" {
		return fmt.Errorf("client_id is required")
	}
	if creds.ClientSecret == "" {
		return fmt.Errorf("client_secret is required")
	}
	if creds.ClientEmail == "" {
		return fmt.Errorf("client_email is required")
	}
	if creds.IssuerURL == "" {
		return fmt.Errorf("issuer_url is required")
	}
	return nil
}

func (r *APIServerConnectionReconciler) updateConnectionStatus(
	ctx context.Context,
	connection *resourcev1alpha1.StreamNativeCloudConnection,
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
		ObservedGeneration: connection.Generation,
		LastTransitionTime: metav1.Now(),
	}

	if err != nil {
		condition.Status = metav1.ConditionFalse
	}

	// Update conditions
	// Find and update existing condition or append new one
	found := false
	for i, c := range connection.Status.Conditions {
		if c.Type == condition.Type {
			connection.Status.Conditions[i] = condition
			found = true
			break
		}
	}
	if !found {
		connection.Status.Conditions = append(connection.Status.Conditions, condition)
	}

	// Update observed generation
	connection.Status.ObservedGeneration = connection.Generation

	// Update last connected time if successful
	if err == nil {
		now := metav1.Now()
		connection.Status.LastConnectedTime = &now
	}

	// Update status
	if err := r.Status().Update(ctx, connection); err != nil {
		log.FromContext(ctx).Error(err, "Failed to update APIServerConnection status")
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *APIServerConnectionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&resourcev1alpha1.StreamNativeCloudConnection{}).
		Complete(r)
}
