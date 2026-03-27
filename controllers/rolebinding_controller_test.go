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
	"sync"
	"sync/atomic"
	"testing"
	"time"

	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	controllers2 "github.com/streamnative/pulsar-resources-operator/pkg/streamnativecloud"
	cloudapi "github.com/streamnative/pulsar-resources-operator/pkg/streamnativecloud/apis/cloud/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func TestRoleBindingReconcilerRequeuesAfterAddingFinalizer(t *testing.T) {
	t.Parallel()

	const (
		namespace       = "test-ns"
		connectionName  = "test-connection"
		organization    = "test-org"
		roleBindingName = "test-rolebinding"
	)

	var (
		tokenRequests  atomic.Int32
		getRequests    atomic.Int32
		createRequests atomic.Int32
		remoteCreated  atomic.Bool

		createdRoleBindingMu sync.Mutex
		createdRoleBinding   *cloudapi.RoleBinding
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/.well-known/openid-configuration":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{
				"token_endpoint": fmt.Sprintf("%s/oauth/token", serverURL(r)),
			})
		case r.Method == http.MethodPost && r.URL.Path == "/oauth/token":
			tokenRequests.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
		case r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf("/apis/cloud.streamnative.io/v1alpha1/namespaces/%s/rolebindings/%s", organization, roleBindingName):
			getRequests.Add(1)
			if !remoteCreated.Load() {
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(metav1.Status{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "Status",
					},
					Status:  metav1.StatusFailure,
					Reason:  metav1.StatusReasonNotFound,
					Code:    http.StatusNotFound,
					Message: fmt.Sprintf("rolebindings.cloud.streamnative.io %q not found", roleBindingName),
				})
				return
			}

			w.Header().Set("Content-Type", "application/json")
			createdRoleBindingMu.Lock()
			defer createdRoleBindingMu.Unlock()
			_ = json.NewEncoder(w).Encode(createdRoleBinding)
		case r.Method == http.MethodPost && r.URL.Path == fmt.Sprintf("/apis/cloud.streamnative.io/v1alpha1/namespaces/%s/rolebindings", organization):
			createRequests.Add(1)

			defer r.Body.Close()
			decoded := &cloudapi.RoleBinding{}
			if err := json.NewDecoder(r.Body).Decode(decoded); err != nil {
				t.Fatalf("decode create request: %v", err)
			}

			decoded.TypeMeta = metav1.TypeMeta{
				APIVersion: "cloud.streamnative.io/v1alpha1",
				Kind:       "RoleBinding",
			}

			createdRoleBindingMu.Lock()
			createdRoleBinding = decoded.DeepCopy()
			createdRoleBindingMu.Unlock()
			remoteCreated.Store(true)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(decoded)
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

	connection := &resourcev1alpha1.StreamNativeCloudConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      connectionName,
			Namespace: namespace,
		},
		Spec: resourcev1alpha1.StreamNativeCloudConnectionSpec{
			Server:       server.URL,
			Organization: organization,
			Auth: resourcev1alpha1.AuthConfig{
				CredentialsRef: corev1.LocalObjectReference{Name: "unused"},
			},
		},
	}

	roleBinding := &resourcev1alpha1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:       roleBindingName,
			Namespace:  namespace,
			Generation: 1,
		},
		Spec: resourcev1alpha1.RoleBindingSpec{
			APIServerRef: corev1.LocalObjectReference{Name: connectionName},
			Users:        []string{"user-1"},
			ClusterRole:  "super-admin",
			SRNTenant:    []string{"tenant-a"},
		},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&resourcev1alpha1.RoleBinding{}).
		WithObjects(connection, roleBinding).
		Build()

	apiConnection, err := controllers2.NewAPIConnection(connection, &resourcev1alpha1.ServiceAccountCredentials{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		IssuerURL:    server.URL,
	})
	if err != nil {
		t.Fatalf("create api connection: %v", err)
	}

	connectionManager := NewConnectionManager(client)
	connectionManager.connections[connectionName] = apiConnection

	reconciler := &RoleBindingReconciler{
		Client:            client,
		Scheme:            scheme,
		ConnectionManager: connectionManager,
	}

	request := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: namespace,
			Name:      roleBindingName,
		},
	}

	result, err := reconciler.Reconcile(context.Background(), request)
	if err != nil {
		t.Fatalf("first reconcile: %v", err)
	}
	if !result.Requeue {
		t.Fatalf("expected first reconcile to requeue after adding finalizer, got %#v", result)
	}
	if result.RequeueAfter != 0 {
		t.Fatalf("expected immediate requeue after adding finalizer, got %s", result.RequeueAfter)
	}
	if getRequests.Load() != 0 {
		t.Fatalf("expected no remote lookup before requeue, got %d", getRequests.Load())
	}

	updated := &resourcev1alpha1.RoleBinding{}
	if err := client.Get(context.Background(), request.NamespacedName, updated); err != nil {
		t.Fatalf("get updated rolebinding after first reconcile: %v", err)
	}
	if !controllerutil.ContainsFinalizer(updated, roleBindingFinalizer) {
		t.Fatalf("expected rolebinding finalizer to be added, got %v", updated.Finalizers)
	}

	result, err = reconciler.Reconcile(context.Background(), request)
	if err != nil {
		t.Fatalf("second reconcile: %v", err)
	}
	if result.Requeue {
		t.Fatalf("expected second reconcile to use periodic sync, got immediate requeue: %#v", result)
	}
	if result.RequeueAfter != 5*time.Minute {
		t.Fatalf("expected periodic requeue after successful reconcile, got %s", result.RequeueAfter)
	}
	if getRequests.Load() != 1 {
		t.Fatalf("expected one remote get on second reconcile, got %d", getRequests.Load())
	}
	if createRequests.Load() != 1 {
		t.Fatalf("expected one remote create on second reconcile, got %d", createRequests.Load())
	}
	if tokenRequests.Load() == 0 {
		t.Fatalf("expected oauth token request when contacting remote api")
	}

	if err := client.Get(context.Background(), request.NamespacedName, updated); err != nil {
		t.Fatalf("get updated rolebinding after second reconcile: %v", err)
	}
	if updated.Status.ObservedGeneration != updated.Generation {
		t.Fatalf("expected observed generation %d, got %d", updated.Generation, updated.Status.ObservedGeneration)
	}
	readyCondition := meta.FindStatusCondition(updated.Status.Conditions, "Ready")
	if readyCondition == nil {
		t.Fatalf("expected ready condition to be present")
	}
	if readyCondition.Status != metav1.ConditionTrue {
		t.Fatalf("expected ready condition true, got %s", readyCondition.Status)
	}

	createdRoleBindingMu.Lock()
	defer createdRoleBindingMu.Unlock()
	if createdRoleBinding == nil {
		t.Fatalf("expected remote rolebinding payload to be captured")
	}
	if createdRoleBinding.Spec.RoleRef.Name != "super-admin" {
		t.Fatalf("expected remote cluster role to be propagated, got %q", createdRoleBinding.Spec.RoleRef.Name)
	}
	if len(createdRoleBinding.Spec.Subjects) != 1 || createdRoleBinding.Spec.Subjects[0].Name != "user-1" {
		t.Fatalf("expected remote subjects to contain user-1, got %#v", createdRoleBinding.Spec.Subjects)
	}
}

func serverURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, r.Host)
}
