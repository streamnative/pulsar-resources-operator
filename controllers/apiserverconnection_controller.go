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
	"errors"
	"fmt"

	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const CloudConnectionFinalizer = "resource.streamnative.io/cloud-connection-finalizer"

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
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=computeworkspaces;computeflinkdeployments;secrets;serviceaccounts;serviceaccountbindings;apikeys;rolebindings,verbs=get;list;watch
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

	// Handle deletion
	if !connection.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(connection, CloudConnectionFinalizer) {
			remaining, err := r.listDependentResources(ctx, req.Namespace, req.Name)
			if err != nil {
				return ctrl.Result{}, err
			}
			if len(remaining) > 0 {
				msg := fmt.Sprintf("waiting for dependent resources to be deleted: %v", remaining)
				logger.Info(msg)
				r.updateConnectionStatus(ctx, connection, fmt.Errorf("dependent resources remain"), "WaitingForDependents", msg)
				return ctrl.Result{}, nil
			}
			_ = r.ConnectionManager.CloseConnection(req.Name)
			controllerutil.RemoveFinalizer(connection, CloudConnectionFinalizer)
			if err := r.Update(ctx, connection); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer
	if !controllerutil.ContainsFinalizer(connection, CloudConnectionFinalizer) {
		controllerutil.AddFinalizer(connection, CloudConnectionFinalizer)
		if err := r.Update(ctx, connection); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
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
			return ctrl.Result{}, errors.New(msg)
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

// listDependentResources returns a list of dependent resources that reference the given connection.
func (r *APIServerConnectionReconciler) listDependentResources(ctx context.Context, namespace, connectionName string) ([]string, error) {
	var remaining []string

	workspaceList := &resourcev1alpha1.ComputeWorkspaceList{}
	if err := r.List(ctx, workspaceList, client.InNamespace(namespace)); err != nil {
		return nil, fmt.Errorf("failed to list ComputeWorkspaces: %w", err)
	}
	connWorkspaces := make(map[string]struct{}, len(workspaceList.Items))
	for i := range workspaceList.Items {
		if workspaceList.Items[i].Spec.APIServerRef.Name == connectionName {
			connWorkspaces[workspaceList.Items[i].Name] = struct{}{}
			remaining = append(remaining, fmt.Sprintf("ComputeWorkspace/%s", workspaceList.Items[i].Name))
		}
	}

	flinkList := &resourcev1alpha1.ComputeFlinkDeploymentList{}
	if err := r.List(ctx, flinkList, client.InNamespace(namespace)); err != nil {
		return nil, fmt.Errorf("failed to list ComputeFlinkDeployments: %w", err)
	}
	for i := range flinkList.Items {
		if flinkList.Items[i].Spec.APIServerRef.Name == connectionName {
			remaining = append(remaining, fmt.Sprintf("ComputeFlinkDeployment/%s", flinkList.Items[i].Name))
			continue
		}
		if flinkList.Items[i].Spec.APIServerRef.Name == "" {
			if _, ok := connWorkspaces[flinkList.Items[i].Spec.WorkspaceName]; ok {
				remaining = append(remaining, fmt.Sprintf("ComputeFlinkDeployment/%s", flinkList.Items[i].Name))
			}
		}
	}

	secretList := &resourcev1alpha1.SecretList{}
	if err := r.List(ctx, secretList, client.InNamespace(namespace)); err != nil {
		return nil, fmt.Errorf("failed to list Secrets: %w", err)
	}
	for i := range secretList.Items {
		if secretList.Items[i].Spec.APIServerRef.Name == connectionName {
			remaining = append(remaining, fmt.Sprintf("Secret/%s", secretList.Items[i].Name))
		}
	}

	saList := &resourcev1alpha1.ServiceAccountList{}
	if err := r.List(ctx, saList, client.InNamespace(namespace)); err != nil {
		return nil, fmt.Errorf("failed to list ServiceAccounts: %w", err)
	}
	connServiceAccounts := make(map[string]struct{}, len(saList.Items))
	for i := range saList.Items {
		if saList.Items[i].Spec.APIServerRef.Name == connectionName {
			connServiceAccounts[saList.Items[i].Name] = struct{}{}
			remaining = append(remaining, fmt.Sprintf("ServiceAccount/%s", saList.Items[i].Name))
		}
	}

	sabList := &resourcev1alpha1.ServiceAccountBindingList{}
	if err := r.List(ctx, sabList, client.InNamespace(namespace)); err != nil {
		return nil, fmt.Errorf("failed to list ServiceAccountBindings: %w", err)
	}
	for i := range sabList.Items {
		if sabList.Items[i].Spec.APIServerRef != nil && sabList.Items[i].Spec.APIServerRef.Name == connectionName {
			remaining = append(remaining, fmt.Sprintf("ServiceAccountBinding/%s", sabList.Items[i].Name))
			continue
		}
		if sabList.Items[i].Spec.APIServerRef == nil || sabList.Items[i].Spec.APIServerRef.Name == "" {
			if _, ok := connServiceAccounts[sabList.Items[i].Spec.ServiceAccountName]; ok {
				remaining = append(remaining, fmt.Sprintf("ServiceAccountBinding/%s", sabList.Items[i].Name))
			}
		}
	}

	akList := &resourcev1alpha1.APIKeyList{}
	if err := r.List(ctx, akList, client.InNamespace(namespace)); err != nil {
		return nil, fmt.Errorf("failed to list APIKeys: %w", err)
	}
	for i := range akList.Items {
		if akList.Items[i].Spec.APIServerRef.Name == connectionName {
			remaining = append(remaining, fmt.Sprintf("APIKey/%s", akList.Items[i].Name))
		}
	}

	rbList := &resourcev1alpha1.RoleBindingList{}
	if err := r.List(ctx, rbList, client.InNamespace(namespace)); err != nil {
		return nil, fmt.Errorf("failed to list RoleBindings: %w", err)
	}
	for i := range rbList.Items {
		if rbList.Items[i].Spec.APIServerRef.Name == connectionName {
			remaining = append(remaining, fmt.Sprintf("RoleBinding/%s", rbList.Items[i].Name))
		}
	}

	return remaining, nil
}

// connectionNameForDependentResource resolves the APIServerRef used by a dependent resource.
func (r *APIServerConnectionReconciler) connectionNameForDependentResource(ctx context.Context, obj client.Object) string {
	switch o := obj.(type) {
	case *resourcev1alpha1.ComputeWorkspace:
		return o.Spec.APIServerRef.Name
	case *resourcev1alpha1.ComputeFlinkDeployment:
		if o.Spec.APIServerRef.Name != "" {
			return o.Spec.APIServerRef.Name
		}
		if o.Spec.WorkspaceName == "" {
			return ""
		}
		workspace := &resourcev1alpha1.ComputeWorkspace{}
		if err := r.Get(ctx, types.NamespacedName{
			Namespace: o.Namespace,
			Name:      o.Spec.WorkspaceName,
		}, workspace); err != nil {
			if !apierrors.IsNotFound(err) {
				log.FromContext(ctx).Error(err, "Failed to resolve ComputeWorkspace for dependent FlinkDeployment",
					"namespace", o.Namespace, "deployment", o.Name, "workspace", o.Spec.WorkspaceName)
			}
			return ""
		}
		return workspace.Spec.APIServerRef.Name
	case *resourcev1alpha1.Secret:
		return o.Spec.APIServerRef.Name
	case *resourcev1alpha1.ServiceAccount:
		return o.Spec.APIServerRef.Name
	case *resourcev1alpha1.ServiceAccountBinding:
		if o.Spec.APIServerRef != nil && o.Spec.APIServerRef.Name != "" {
			return o.Spec.APIServerRef.Name
		}
		if o.Spec.ServiceAccountName == "" {
			return ""
		}
		serviceAccount := &resourcev1alpha1.ServiceAccount{}
		if err := r.Get(ctx, types.NamespacedName{
			Namespace: o.Namespace,
			Name:      o.Spec.ServiceAccountName,
		}, serviceAccount); err != nil {
			if !apierrors.IsNotFound(err) {
				log.FromContext(ctx).Error(err, "Failed to resolve ServiceAccount for dependent ServiceAccountBinding",
					"namespace", o.Namespace, "binding", o.Name, "serviceAccount", o.Spec.ServiceAccountName)
			}
			return ""
		}
		return serviceAccount.Spec.APIServerRef.Name
	case *resourcev1alpha1.APIKey:
		return o.Spec.APIServerRef.Name
	case *resourcev1alpha1.RoleBinding:
		return o.Spec.APIServerRef.Name
	default:
		return ""
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *APIServerConnectionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	mapFunc := func(ctx context.Context, obj client.Object) []ctrl.Request {
		name := r.connectionNameForDependentResource(ctx, obj)
		if name == "" {
			return nil
		}
		return []ctrl.Request{
			{NamespacedName: types.NamespacedName{
				Namespace: obj.GetNamespace(),
				Name:      name,
			}},
		}
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&resourcev1alpha1.StreamNativeCloudConnection{}).
		Watches(&resourcev1alpha1.ComputeWorkspace{}, handler.EnqueueRequestsFromMapFunc(mapFunc)).
		Watches(&resourcev1alpha1.ComputeFlinkDeployment{}, handler.EnqueueRequestsFromMapFunc(mapFunc)).
		Watches(&resourcev1alpha1.Secret{}, handler.EnqueueRequestsFromMapFunc(mapFunc)).
		Watches(&resourcev1alpha1.ServiceAccount{}, handler.EnqueueRequestsFromMapFunc(mapFunc)).
		Watches(&resourcev1alpha1.ServiceAccountBinding{}, handler.EnqueueRequestsFromMapFunc(mapFunc)).
		Watches(&resourcev1alpha1.APIKey{}, handler.EnqueueRequestsFromMapFunc(mapFunc)).
		Watches(&resourcev1alpha1.RoleBinding{}, handler.EnqueueRequestsFromMapFunc(mapFunc)).
		Complete(r)
}
