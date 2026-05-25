// Copyright 2026 StreamNative
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

package connection

import (
	"context"
	"testing"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	"github.com/streamnative/pulsar-resources-operator/pkg/admin"
	"github.com/streamnative/pulsar-resources-operator/pkg/feature"
	"github.com/streamnative/pulsar-resources-operator/pkg/reconciler"
)

type countingConnectionClient struct {
	client.Client
	updateCalls       int
	statusUpdateCalls int
}

func (c *countingConnectionClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	c.updateCalls++
	return c.Client.Update(ctx, obj, opts...)
}

func (c *countingConnectionClient) Status() client.SubResourceWriter {
	return &countingConnectionStatusWriter{
		SubResourceWriter: c.Client.Status(),
		updateCalls:       &c.statusUpdateCalls,
	}
}

type countingConnectionStatusWriter struct {
	client.SubResourceWriter
	updateCalls *int
}

func (w *countingConnectionStatusWriter) Update(ctx context.Context, obj client.Object, opts ...client.SubResourceUpdateOption) error {
	(*w.updateCalls)++
	return w.SubResourceWriter.Update(ctx, obj, opts...)
}

type countingConnectionChildReconciler struct {
	reconcileCalls int
}

func (r *countingConnectionChildReconciler) Observe(context.Context) error {
	return nil
}

func (r *countingConnectionChildReconciler) Reconcile(context.Context) error {
	r.reconcileCalls++
	return nil
}

func TestPulsarConnectionReconcileSkipsReadyChildrenWhenAlwaysUpdateDisabled(t *testing.T) {
	setAlwaysUpdatePulsarResourceForTest(t, false)

	reconciler, childReconciler, adminCreateCalls := newConnectionReconcilerForReadyChildrenTest(t, newReadyTestConnection())
	reconciler.tenants = []resourcev1alpha1.PulsarTenant{newReadyTestTenant()}

	if err := reconciler.Reconcile(context.Background()); err != nil {
		t.Fatalf("reconcile connection: %v", err)
	}
	if *adminCreateCalls != 0 {
		t.Fatalf("expected no pulsar admin creation, got %d", *adminCreateCalls)
	}
	if childReconciler.reconcileCalls != 0 {
		t.Fatalf("expected no child reconcile, got %d", childReconciler.reconcileCalls)
	}
}

func TestPulsarConnectionReconcileReadyChildrenWhenAlwaysUpdateEnabled(t *testing.T) {
	setAlwaysUpdatePulsarResourceForTest(t, true)

	connection := newReadyTestConnection()
	reconciler, childReconciler, adminCreateCalls := newConnectionReconcilerForReadyChildrenTest(t, connection)
	reconciler.tenants = []resourcev1alpha1.PulsarTenant{newReadyTestTenant()}

	if err := reconciler.Reconcile(context.Background()); err != nil {
		t.Fatalf("reconcile connection: %v", err)
	}
	if *adminCreateCalls != 2 {
		t.Fatalf("expected two pulsar admin creations, got %d", *adminCreateCalls)
	}
	if childReconciler.reconcileCalls != 1 {
		t.Fatalf("expected one child reconcile, got %d", childReconciler.reconcileCalls)
	}

	updated := &resourcev1alpha1.PulsarConnection{}
	if err := reconciler.client.Get(context.Background(), types.NamespacedName{Namespace: connection.Namespace, Name: connection.Name}, updated); err != nil {
		t.Fatalf("get updated connection: %v", err)
	}
	if !containsString(updated.Finalizers, resourcev1alpha1.FinalizerName) {
		t.Fatalf("expected connection finalizer %q, got %v", resourcev1alpha1.FinalizerName, updated.Finalizers)
	}
}

func TestPulsarConnectionReconcileReadyChildrenSkipsNoopConnectionUpdatesWhenAlwaysUpdateEnabled(t *testing.T) {
	setAlwaysUpdatePulsarResourceForTest(t, true)

	connection := newReadyTestConnection()
	connection.Finalizers = []string{resourcev1alpha1.FinalizerName}
	connection.Status.ObservedGeneration = connection.Generation
	connection.Status.Conditions = []metav1.Condition{*NewReadyCondition(connection.Generation)}

	reconciler, childReconciler, adminCreateCalls := newConnectionReconcilerForReadyChildrenTest(t, connection)
	countingClient := &countingConnectionClient{Client: reconciler.client}
	reconciler.client = countingClient
	reconciler.tenants = []resourcev1alpha1.PulsarTenant{newReadyTestTenant()}

	if err := reconciler.Reconcile(context.Background()); err != nil {
		t.Fatalf("reconcile connection: %v", err)
	}
	if *adminCreateCalls != 2 {
		t.Fatalf("expected two pulsar admin creations, got %d", *adminCreateCalls)
	}
	if childReconciler.reconcileCalls != 1 {
		t.Fatalf("expected one child reconcile, got %d", childReconciler.reconcileCalls)
	}
	if countingClient.updateCalls != 0 {
		t.Fatalf("expected no no-op connection update, got %d", countingClient.updateCalls)
	}
	if countingClient.statusUpdateCalls != 0 {
		t.Fatalf("expected no no-op connection status update, got %d", countingClient.statusUpdateCalls)
	}
}

func TestPulsarConnectionDeletionKeepsRemainingResourceGuardWhenAlwaysUpdateEnabled(t *testing.T) {
	setAlwaysUpdatePulsarResourceForTest(t, true)

	connection := newReadyTestConnection()
	now := metav1.NewTime(time.Unix(1700000000, 0))
	connection.DeletionTimestamp = &now
	connection.Finalizers = []string{resourcev1alpha1.FinalizerName}

	reconciler, childReconciler, adminCreateCalls := newConnectionReconcilerForReadyChildrenTest(t, connection)
	reconciler.tenants = []resourcev1alpha1.PulsarTenant{newReadyTestTenant()}

	if err := reconciler.Reconcile(context.Background()); err != nil {
		t.Fatalf("reconcile deleting connection: %v", err)
	}
	if *adminCreateCalls != 0 {
		t.Fatalf("expected no pulsar admin creation for deleting connection with remaining children, got %d", *adminCreateCalls)
	}
	if childReconciler.reconcileCalls != 0 {
		t.Fatalf("expected no child reconcile for deleting connection with remaining children, got %d", childReconciler.reconcileCalls)
	}

	updated := &resourcev1alpha1.PulsarConnection{}
	if err := reconciler.client.Get(context.Background(), types.NamespacedName{Namespace: connection.Namespace, Name: connection.Name}, updated); err != nil {
		t.Fatalf("get updated connection: %v", err)
	}
	if !containsString(updated.Finalizers, resourcev1alpha1.FinalizerName) {
		t.Fatalf("expected remaining-resource guard to keep finalizer %q, got %v", resourcev1alpha1.FinalizerName, updated.Finalizers)
	}
}

func newConnectionReconcilerForReadyChildrenTest(t *testing.T, connection *resourcev1alpha1.PulsarConnection) (*PulsarConnectionReconciler, *countingConnectionChildReconciler, *int) {
	t.Helper()

	scheme := runtime.NewScheme()
	if err := resourcev1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add resource scheme: %v", err)
	}

	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&resourcev1alpha1.PulsarConnection{}).
		WithObjects(connection).
		Build()

	childReconciler := &countingConnectionChildReconciler{}
	adminCreateCalls := 0
	reconciler := &PulsarConnectionReconciler{
		connection: connection,
		log:        logr.Discard(),
		client:     k8sClient,
		creator: func(admin.PulsarAdminConfig) (admin.PulsarAdmin, error) {
			adminCreateCalls++
			return &admin.DummyPulsarAdmin{}, nil
		},
		reconcilers: []reconciler.Interface{childReconciler},
	}

	return reconciler, childReconciler, &adminCreateCalls
}

func newReadyTestConnection() *resourcev1alpha1.PulsarConnection {
	return &resourcev1alpha1.PulsarConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "connection",
			Namespace:  "test-ns",
			Generation: 1,
		},
		Spec: resourcev1alpha1.PulsarConnectionSpec{
			AdminServiceURL: "http://pulsar-admin:8080",
		},
	}
}

func newReadyTestTenant() resourcev1alpha1.PulsarTenant {
	return resourcev1alpha1.PulsarTenant{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "tenant",
			Namespace:  "test-ns",
			Generation: 1,
		},
		Spec: resourcev1alpha1.PulsarTenantSpec{
			Name:          "tenant",
			ConnectionRef: corev1.LocalObjectReference{Name: "connection"},
		},
		Status: resourcev1alpha1.PulsarTenantStatus{
			ObservedGeneration: 1,
			Conditions:         []metav1.Condition{*NewReadyCondition(1)},
		},
	}
}

func setAlwaysUpdatePulsarResourceForTest(t *testing.T, enabled bool) {
	t.Helper()
	setAlwaysUpdatePulsarResource(t, enabled)
	t.Cleanup(func() {
		setAlwaysUpdatePulsarResource(t, false)
	})
}

func setAlwaysUpdatePulsarResource(t *testing.T, enabled bool) {
	t.Helper()
	if err := feature.DefaultMutableFeatureGate.SetFromMap(map[string]bool{
		string(feature.AlwaysUpdatePulsarResource): enabled,
	}); err != nil {
		t.Fatalf("set AlwaysUpdatePulsarResource=%t: %v", enabled, err)
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
