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
	"testing"
	"time"

	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	streamnativecloud "github.com/streamnative/pulsar-resources-operator/pkg/streamnativecloud"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func TestKeepAfterDeletionRemovesFinalizerWithoutRemoteDependencies(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	if err := resourcev1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add resource scheme: %v", err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("add core scheme: %v", err)
	}

	deletionTime := metav1.NewTime(time.Unix(1700000000, 0))
	objectMeta := func(name, finalizer string) metav1.ObjectMeta {
		return metav1.ObjectMeta{
			Name:              name,
			Namespace:         "test-ns",
			Finalizers:        []string{finalizer},
			DeletionTimestamp: &deletionTime,
		}
	}

	testCases := []struct {
		name      string
		object    client.Object
		reconcile func(client.Client) (ctrl.Result, error)
		load      func() client.Object
	}{
		{
			name: "workspace",
			object: &resourcev1alpha1.ComputeWorkspace{
				ObjectMeta: objectMeta("workspace", streamnativecloud.WorkspaceFinalizer),
				Spec: resourcev1alpha1.ComputeWorkspaceSpec{
					APIServerRef:    corev1.LocalObjectReference{Name: "conn"},
					LifecyclePolicy: resourcev1alpha1.KeepAfterDeletion,
				},
			},
			reconcile: func(cl client.Client) (ctrl.Result, error) {
				return (&WorkspaceReconciler{Client: cl}).Reconcile(context.Background(), ctrl.Request{
					NamespacedName: types.NamespacedName{Namespace: "test-ns", Name: "workspace"},
				})
			},
			load: func() client.Object { return &resourcev1alpha1.ComputeWorkspace{} },
		},
		{
			name: "flink deployment",
			object: &resourcev1alpha1.ComputeFlinkDeployment{
				ObjectMeta: objectMeta("flink", streamnativecloud.FlinkDeploymentFinalizer),
				Spec: resourcev1alpha1.ComputeFlinkDeploymentSpec{
					WorkspaceName:   "workspace",
					LifecyclePolicy: resourcev1alpha1.KeepAfterDeletion,
				},
			},
			reconcile: func(cl client.Client) (ctrl.Result, error) {
				return (&FlinkDeploymentReconciler{Client: cl}).Reconcile(context.Background(), ctrl.Request{
					NamespacedName: types.NamespacedName{Namespace: "test-ns", Name: "flink"},
				})
			},
			load: func() client.Object { return &resourcev1alpha1.ComputeFlinkDeployment{} },
		},
		{
			name: "secret",
			object: &resourcev1alpha1.Secret{
				ObjectMeta: objectMeta("secret", streamnativecloud.SecretFinalizer),
				Spec: resourcev1alpha1.SecretSpec{
					LifecyclePolicy: resourcev1alpha1.KeepAfterDeletion,
				},
			},
			reconcile: func(cl client.Client) (ctrl.Result, error) {
				return (&SecretReconciler{Client: cl}).Reconcile(context.Background(), ctrl.Request{
					NamespacedName: types.NamespacedName{Namespace: "test-ns", Name: "secret"},
				})
			},
			load: func() client.Object { return &resourcev1alpha1.Secret{} },
		},
		{
			name: "service account",
			object: &resourcev1alpha1.ServiceAccount{
				ObjectMeta: objectMeta("service-account", ServiceAccountFinalizer),
				Spec: resourcev1alpha1.ServiceAccountSpec{
					LifecyclePolicy: resourcev1alpha1.KeepAfterDeletion,
				},
			},
			reconcile: func(cl client.Client) (ctrl.Result, error) {
				return (&ServiceAccountReconciler{Client: cl}).Reconcile(context.Background(), ctrl.Request{
					NamespacedName: types.NamespacedName{Namespace: "test-ns", Name: "service-account"},
				})
			},
			load: func() client.Object { return &resourcev1alpha1.ServiceAccount{} },
		},
		{
			name: "service account binding",
			object: &resourcev1alpha1.ServiceAccountBinding{
				ObjectMeta: objectMeta("service-account-binding", ServiceAccountBindingFinalizer),
				Spec: resourcev1alpha1.ServiceAccountBindingSpec{
					ServiceAccountName: "service-account",
					LifecyclePolicy:    resourcev1alpha1.KeepAfterDeletion,
				},
			},
			reconcile: func(cl client.Client) (ctrl.Result, error) {
				return (&ServiceAccountBindingReconciler{Client: cl}).Reconcile(context.Background(), ctrl.Request{
					NamespacedName: types.NamespacedName{Namespace: "test-ns", Name: "service-account-binding"},
				})
			},
			load: func() client.Object { return &resourcev1alpha1.ServiceAccountBinding{} },
		},
		{
			name: "api key",
			object: &resourcev1alpha1.APIKey{
				ObjectMeta: objectMeta("api-key", APIKeyFinalizer),
				Spec: resourcev1alpha1.APIKeySpec{
					LifecyclePolicy: resourcev1alpha1.KeepAfterDeletion,
				},
			},
			reconcile: func(cl client.Client) (ctrl.Result, error) {
				return (&APIKeyReconciler{Client: cl}).Reconcile(context.Background(), ctrl.Request{
					NamespacedName: types.NamespacedName{Namespace: "test-ns", Name: "api-key"},
				})
			},
			load: func() client.Object { return &resourcev1alpha1.APIKey{} },
		},
		{
			name: "role binding",
			object: &resourcev1alpha1.RoleBinding{
				ObjectMeta: objectMeta("role-binding", roleBindingFinalizer),
				Spec: resourcev1alpha1.RoleBindingSpec{
					LifecyclePolicy: resourcev1alpha1.KeepAfterDeletion,
				},
			},
			reconcile: func(cl client.Client) (ctrl.Result, error) {
				return (&RoleBindingReconciler{Client: cl}).Reconcile(context.Background(), ctrl.Request{
					NamespacedName: types.NamespacedName{Namespace: "test-ns", Name: "role-binding"},
				})
			},
			load: func() client.Object { return &resourcev1alpha1.RoleBinding{} },
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cl := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(testCase.object.DeepCopyObject().(client.Object)).
				Build()

			if _, err := testCase.reconcile(cl); err != nil {
				t.Fatalf("reconcile: %v", err)
			}

			actual := testCase.load()
			if err := cl.Get(context.Background(), client.ObjectKeyFromObject(testCase.object), actual); err != nil {
				if apierrors.IsNotFound(err) {
					return
				}
				t.Fatalf("get reconciled object: %v", err)
			}

			if controllerutil.ContainsFinalizer(actual, testCase.object.GetFinalizers()[0]) {
				t.Fatalf("expected finalizer %q to be removed, got %v", testCase.object.GetFinalizers()[0], actual.GetFinalizers())
			}
		})
	}
}
