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
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	controllers2 "github.com/streamnative/pulsar-resources-operator/pkg/streamnativecloud"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestServiceAccountReconcilerIgnoresRemoteNotFoundDuringDeletion(t *testing.T) {
	t.Parallel()

	const (
		namespace          = "test-ns"
		connectionName     = "test-connection"
		organization       = "test-org"
		serviceAccountName = "test-service-account"
	)

	var deleteRequests atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/.well-known/openid-configuration":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{
				"token_endpoint": fmt.Sprintf("%s/oauth/token", serverURL(r)),
			})
		case r.Method == http.MethodPost && r.URL.Path == "/oauth/token":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
		case r.Method == http.MethodDelete && r.URL.Path == fmt.Sprintf("/apis/cloud.streamnative.io/v1alpha1/namespaces/%s/serviceaccounts/%s", organization, serviceAccountName):
			deleteRequests.Add(1)
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(metav1.Status{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Status",
				},
				Status:  metav1.StatusFailure,
				Reason:  metav1.StatusReasonNotFound,
				Code:    http.StatusNotFound,
				Message: fmt.Sprintf("serviceaccounts.cloud.streamnative.io %q not found", serviceAccountName),
			})
		default:
			w.WriteHeader(http.StatusInternalServerError)
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	scheme := runtime.NewScheme()
	if err := resourcev1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add resource scheme: %v", err)
	}

	connection := newTestCloudConnection(namespace, connectionName, organization, server.URL)
	now := metav1.NewTime(time.Now())
	serviceAccount := &resourcev1alpha1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:              serviceAccountName,
			Namespace:         namespace,
			DeletionTimestamp: &now,
			Finalizers:        []string{ServiceAccountFinalizer},
		},
		Spec: resourcev1alpha1.ServiceAccountSpec{
			APIServerRef: corev1.LocalObjectReference{Name: connectionName},
		},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(connection, serviceAccount).
		Build()

	reconciler := &ServiceAccountReconciler{
		Client:            client,
		Scheme:            scheme,
		ConnectionManager: newTestConnectionManager(t, client, connection, server.URL),
	}

	request := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: namespace,
			Name:      serviceAccountName,
		},
	}

	result, err := reconciler.Reconcile(context.Background(), request)
	if err != nil {
		t.Fatalf("reconcile service account deletion: %v", err)
	}
	if result != (ctrl.Result{}) {
		t.Fatalf("expected no requeue after successful deletion, got %#v", result)
	}
	if deleteRequests.Load() != 1 {
		t.Fatalf("expected one remote delete request, got %d", deleteRequests.Load())
	}

	updated := &resourcev1alpha1.ServiceAccount{}
	if err := client.Get(context.Background(), request.NamespacedName, updated); !apierrors.IsNotFound(err) {
		t.Fatalf("expected service account to be fully removed after finalizer cleanup, got err=%v object=%#v", err, updated)
	}
}

func TestServiceAccountBindingReconcilerIgnoresRemoteNotFoundDuringDeletion(t *testing.T) {
	t.Parallel()

	const (
		namespace          = "test-ns"
		connectionName     = "test-connection"
		organization       = "test-org"
		serviceAccountName = "test-service-account"
		bindingName        = "test-binding"
		poolNamespace      = "pool-ns"
		poolName           = "pool-a"
	)

	remoteBindingName := fmt.Sprintf("%s.%s.%s", serviceAccountName, poolNamespace, poolName)
	var deleteRequests atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/.well-known/openid-configuration":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{
				"token_endpoint": fmt.Sprintf("%s/oauth/token", serverURL(r)),
			})
		case r.Method == http.MethodPost && r.URL.Path == "/oauth/token":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
		case r.Method == http.MethodDelete && r.URL.Path == fmt.Sprintf("/apis/cloud.streamnative.io/v1alpha1/namespaces/%s/serviceaccountbindings/%s", organization, remoteBindingName):
			deleteRequests.Add(1)
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(metav1.Status{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Status",
				},
				Status:  metav1.StatusFailure,
				Reason:  metav1.StatusReasonNotFound,
				Code:    http.StatusNotFound,
				Message: fmt.Sprintf("serviceaccountbindings.cloud.streamnative.io %q not found", remoteBindingName),
			})
		default:
			w.WriteHeader(http.StatusInternalServerError)
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	scheme := runtime.NewScheme()
	if err := resourcev1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add resource scheme: %v", err)
	}

	connection := newTestCloudConnection(namespace, connectionName, organization, server.URL)
	serviceAccount := &resourcev1alpha1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceAccountName,
			Namespace: namespace,
		},
		Spec: resourcev1alpha1.ServiceAccountSpec{
			APIServerRef: corev1.LocalObjectReference{Name: connectionName},
		},
	}

	now := metav1.NewTime(time.Now())
	binding := &resourcev1alpha1.ServiceAccountBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:              bindingName,
			Namespace:         namespace,
			DeletionTimestamp: &now,
			Finalizers:        []string{ServiceAccountBindingFinalizer},
		},
		Spec: resourcev1alpha1.ServiceAccountBindingSpec{
			ServiceAccountName: serviceAccountName,
			PoolMemberRefs: []resourcev1alpha1.PoolMemberReference{
				{
					Namespace: poolNamespace,
					Name:      poolName,
				},
			},
		},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(connection, serviceAccount, binding).
		Build()

	reconciler := &ServiceAccountBindingReconciler{
		Client:            client,
		Scheme:            scheme,
		ConnectionManager: newTestConnectionManager(t, client, connection, server.URL),
	}

	request := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: namespace,
			Name:      bindingName,
		},
	}

	result, err := reconciler.Reconcile(context.Background(), request)
	if err != nil {
		t.Fatalf("reconcile service account binding deletion: %v", err)
	}
	if result != (ctrl.Result{}) {
		t.Fatalf("expected no requeue after successful deletion, got %#v", result)
	}
	if deleteRequests.Load() != 1 {
		t.Fatalf("expected one remote delete request, got %d", deleteRequests.Load())
	}

	updated := &resourcev1alpha1.ServiceAccountBinding{}
	if err := client.Get(context.Background(), request.NamespacedName, updated); !apierrors.IsNotFound(err) {
		t.Fatalf("expected service account binding to be fully removed after finalizer cleanup, got err=%v object=%#v", err, updated)
	}
}

func TestAPIKeyReconcilerIgnoresRemoteNotFoundDuringDeletion(t *testing.T) {
	t.Parallel()

	const (
		namespace      = "test-ns"
		connectionName = "test-connection"
		organization   = "test-org"
		apiKeyName     = "test-api-key"
	)

	server, deleteRequests := newDeleteNotFoundTestServer(
		t,
		fmt.Sprintf("/apis/cloud.streamnative.io/v1alpha1/namespaces/%s/apikeys/%s", organization, apiKeyName),
		fmt.Sprintf("apikeys.cloud.streamnative.io %q not found", apiKeyName),
	)
	defer server.Close()

	scheme := runtime.NewScheme()
	if err := resourcev1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add resource scheme: %v", err)
	}

	connection := newTestCloudConnection(namespace, connectionName, organization, server.URL)
	now := metav1.NewTime(time.Now())
	apiKey := &resourcev1alpha1.APIKey{
		ObjectMeta: metav1.ObjectMeta{
			Name:              apiKeyName,
			Namespace:         namespace,
			DeletionTimestamp: &now,
			Finalizers:        []string{APIKeyFinalizer},
		},
		Spec: resourcev1alpha1.APIKeySpec{
			APIServerRef: corev1.LocalObjectReference{Name: connectionName},
		},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(connection, apiKey).
		Build()

	reconciler := &APIKeyReconciler{
		Client:            client,
		Scheme:            scheme,
		ConnectionManager: newTestConnectionManager(t, client, connection, server.URL),
	}

	request := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: namespace,
			Name:      apiKeyName,
		},
	}

	result, err := reconciler.Reconcile(context.Background(), request)
	assertDeleteNotFoundCleanup(t, result, err, deleteRequests, client, request.NamespacedName, &resourcev1alpha1.APIKey{})
}

func TestSecretReconcilerIgnoresRemoteNotFoundDuringDeletion(t *testing.T) {
	t.Parallel()

	const (
		namespace      = "test-ns"
		connectionName = "test-connection"
		organization   = "test-org"
		secretName     = "test-secret"
	)

	server, deleteRequests := newDeleteNotFoundTestServer(
		t,
		fmt.Sprintf("/apis/cloud.streamnative.io/v1alpha1/namespaces/%s/secrets/%s", organization, secretName),
		fmt.Sprintf("secrets.cloud.streamnative.io %q not found", secretName),
	)
	defer server.Close()

	scheme := runtime.NewScheme()
	if err := resourcev1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add resource scheme: %v", err)
	}

	connection := newTestCloudConnection(namespace, connectionName, organization, server.URL)
	now := metav1.NewTime(time.Now())
	secretCR := &resourcev1alpha1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:              secretName,
			Namespace:         namespace,
			DeletionTimestamp: &now,
			Finalizers:        []string{controllers2.SecretFinalizer},
		},
		Spec: resourcev1alpha1.SecretSpec{
			APIServerRef: corev1.LocalObjectReference{Name: connectionName},
		},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(connection, secretCR).
		Build()

	reconciler := &SecretReconciler{
		Client:            client,
		Scheme:            scheme,
		ConnectionManager: newTestConnectionManager(t, client, connection, server.URL),
	}

	request := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: namespace,
			Name:      secretName,
		},
	}

	result, err := reconciler.Reconcile(context.Background(), request)
	assertDeleteNotFoundCleanup(t, result, err, deleteRequests, client, request.NamespacedName, &resourcev1alpha1.Secret{})
}

func TestWorkspaceReconcilerIgnoresRemoteNotFoundDuringDeletion(t *testing.T) {
	t.Parallel()

	const (
		namespace      = "test-ns"
		connectionName = "test-connection"
		organization   = "test-org"
		workspaceName  = "test-workspace"
	)

	server, deleteRequests := newDeleteNotFoundTestServer(
		t,
		fmt.Sprintf("/apis/compute.streamnative.io/v1alpha1/namespaces/%s/workspaces/%s", organization, workspaceName),
		fmt.Sprintf("workspaces.compute.streamnative.io %q not found", workspaceName),
	)
	defer server.Close()

	scheme := runtime.NewScheme()
	if err := resourcev1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add resource scheme: %v", err)
	}

	connection := newTestCloudConnection(namespace, connectionName, organization, server.URL)
	now := metav1.NewTime(time.Now())
	workspace := &resourcev1alpha1.ComputeWorkspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:              workspaceName,
			Namespace:         namespace,
			DeletionTimestamp: &now,
			Finalizers:        []string{controllers2.WorkspaceFinalizer},
		},
		Spec: resourcev1alpha1.ComputeWorkspaceSpec{
			APIServerRef: corev1.LocalObjectReference{Name: connectionName},
		},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(connection, workspace).
		Build()

	reconciler := &WorkspaceReconciler{
		Client:            client,
		Scheme:            scheme,
		ConnectionManager: newTestConnectionManager(t, client, connection, server.URL),
	}

	request := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: namespace,
			Name:      workspaceName,
		},
	}

	result, err := reconciler.Reconcile(context.Background(), request)
	assertDeleteNotFoundCleanup(t, result, err, deleteRequests, client, request.NamespacedName, &resourcev1alpha1.ComputeWorkspace{})
}

func TestFlinkDeploymentReconcilerIgnoresRemoteNotFoundDuringDeletion(t *testing.T) {
	t.Parallel()

	const (
		namespace      = "test-ns"
		connectionName = "test-connection"
		organization   = "test-org"
		deploymentName = "test-deployment"
		workspaceName  = "unused-workspace"
	)

	server, deleteRequests := newDeleteNotFoundTestServer(
		t,
		fmt.Sprintf("/apis/compute.streamnative.io/v1alpha1/namespaces/%s/flinkdeployments/%s", organization, deploymentName),
		fmt.Sprintf("flinkdeployments.compute.streamnative.io %q not found", deploymentName),
	)
	defer server.Close()

	scheme := runtime.NewScheme()
	if err := resourcev1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add resource scheme: %v", err)
	}

	connection := newTestCloudConnection(namespace, connectionName, organization, server.URL)
	now := metav1.NewTime(time.Now())
	deployment := &resourcev1alpha1.ComputeFlinkDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:              deploymentName,
			Namespace:         namespace,
			DeletionTimestamp: &now,
			Finalizers:        []string{controllers2.FlinkDeploymentFinalizer},
		},
		Spec: resourcev1alpha1.ComputeFlinkDeploymentSpec{
			APIServerRef:  corev1.LocalObjectReference{Name: connectionName},
			WorkspaceName: workspaceName,
		},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(connection, deployment).
		Build()

	reconciler := &FlinkDeploymentReconciler{
		Client:            client,
		Scheme:            scheme,
		ConnectionManager: newTestConnectionManager(t, client, connection, server.URL),
	}

	request := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: namespace,
			Name:      deploymentName,
		},
	}

	result, err := reconciler.Reconcile(context.Background(), request)
	assertDeleteNotFoundCleanup(t, result, err, deleteRequests, client, request.NamespacedName, &resourcev1alpha1.ComputeFlinkDeployment{})
}

func newDeleteNotFoundTestServer(t *testing.T, deletePath, notFoundMessage string) (*httptest.Server, *atomic.Int32) {
	t.Helper()

	var deleteRequests atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/.well-known/openid-configuration":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{
				"token_endpoint": fmt.Sprintf("%s/oauth/token", serverURL(r)),
			})
		case r.Method == http.MethodPost && r.URL.Path == "/oauth/token":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
		case r.Method == http.MethodDelete && r.URL.Path == deletePath:
			deleteRequests.Add(1)
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(metav1.Status{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Status",
				},
				Status:  metav1.StatusFailure,
				Reason:  metav1.StatusReasonNotFound,
				Code:    http.StatusNotFound,
				Message: notFoundMessage,
			})
		default:
			w.WriteHeader(http.StatusInternalServerError)
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))

	return server, &deleteRequests
}

func assertDeleteNotFoundCleanup(
	t *testing.T,
	result ctrl.Result,
	err error,
	deleteRequests *atomic.Int32,
	k8sClient client.Client,
	key types.NamespacedName,
	obj client.Object,
) {
	t.Helper()

	if err != nil {
		t.Fatalf("reconcile deletion: %v", err)
	}
	if result != (ctrl.Result{}) {
		t.Fatalf("expected no requeue after successful deletion, got %#v", result)
	}
	if deleteRequests.Load() != 1 {
		t.Fatalf("expected one remote delete request, got %d", deleteRequests.Load())
	}
	if err := k8sClient.Get(context.Background(), key, obj); !apierrors.IsNotFound(err) {
		t.Fatalf("expected object to be fully removed after finalizer cleanup, got err=%v object=%#v", err, obj)
	}
}

func newTestCloudConnection(namespace, name, organization, serverURL string) *resourcev1alpha1.StreamNativeCloudConnection {
	return &resourcev1alpha1.StreamNativeCloudConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: resourcev1alpha1.StreamNativeCloudConnectionSpec{
			Server:       serverURL,
			Organization: organization,
			Auth: resourcev1alpha1.AuthConfig{
				CredentialsRef: corev1.LocalObjectReference{Name: "unused"},
			},
		},
	}
}

func newTestConnectionManager(
	t *testing.T,
	k8sClient client.Client,
	connection *resourcev1alpha1.StreamNativeCloudConnection,
	serverURL string,
) *ConnectionManager {
	t.Helper()

	apiConnection, err := controllers2.NewAPIConnection(connection, &resourcev1alpha1.ServiceAccountCredentials{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		IssuerURL:    serverURL,
	})
	if err != nil {
		t.Fatalf("create api connection: %v", err)
	}

	connectionManager := NewConnectionManager(k8sClient)
	connectionManager.connections[connection.Name] = apiConnection
	return connectionManager
}
