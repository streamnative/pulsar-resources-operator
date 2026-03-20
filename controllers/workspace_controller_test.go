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

	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestWorkspaceReconcilerFindResourcesForConnection(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	if err := resourcev1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add resource scheme: %v", err)
	}

	connection := &resourcev1alpha1.StreamNativeCloudConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "conn-a",
			Namespace: "test-ns",
		},
	}

	otherNamespaceConnection := &resourcev1alpha1.StreamNativeCloudConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "conn-a",
			Namespace: "other-ns",
		},
	}

	referencedWorkspace := &resourcev1alpha1.ComputeWorkspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workspace-a",
			Namespace: "test-ns",
		},
		Spec: resourcev1alpha1.ComputeWorkspaceSpec{
			APIServerRef: corev1.LocalObjectReference{Name: "conn-a"},
		},
	}

	anotherReferencedWorkspace := &resourcev1alpha1.ComputeWorkspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workspace-b",
			Namespace: "test-ns",
		},
		Spec: resourcev1alpha1.ComputeWorkspaceSpec{
			APIServerRef: corev1.LocalObjectReference{Name: "conn-a"},
		},
	}

	differentConnectionWorkspace := &resourcev1alpha1.ComputeWorkspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workspace-other-conn",
			Namespace: "test-ns",
		},
		Spec: resourcev1alpha1.ComputeWorkspaceSpec{
			APIServerRef: corev1.LocalObjectReference{Name: "conn-b"},
		},
	}

	differentNamespaceWorkspace := &resourcev1alpha1.ComputeWorkspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workspace-other-ns",
			Namespace: "other-ns",
		},
		Spec: resourcev1alpha1.ComputeWorkspaceSpec{
			APIServerRef: corev1.LocalObjectReference{Name: "conn-a"},
		},
	}

	cl := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(
			connection,
			otherNamespaceConnection,
			referencedWorkspace,
			anotherReferencedWorkspace,
			differentConnectionWorkspace,
			differentNamespaceWorkspace,
		).
		WithIndex(&resourcev1alpha1.ComputeWorkspace{}, workspaceAPIServerRefNameField, workspaceAPIServerRefNameIndex).
		Build()

	reconciler := &WorkspaceReconciler{Client: cl}

	requests := reconciler.findResourcesForConnection(context.Background(), connection)

	expected := map[ctrl.Request]struct{}{
		{NamespacedName: types.NamespacedName{Namespace: "test-ns", Name: "workspace-a"}}: {},
		{NamespacedName: types.NamespacedName{Namespace: "test-ns", Name: "workspace-b"}}: {},
	}

	if len(requests) != len(expected) {
		t.Fatalf("expected %d requests, got %d: %#v", len(expected), len(requests), requests)
	}

	for _, request := range requests {
		if _, ok := expected[request]; !ok {
			t.Fatalf("unexpected request returned: %#v", request)
		}
		delete(expected, request)
	}

	if len(expected) != 0 {
		t.Fatalf("missing requests: %#v", expected)
	}
}
