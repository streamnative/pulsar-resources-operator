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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	"github.com/streamnative/pulsar-resources-operator/pkg/admin"
)

// foreignFinalizer stands in for a finalizer owned by another controller (e.g. the
// Kubernetes foregroundDeletion finalizer, or a GitOps tool's finalizer).
const foreignFinalizer = "other.io/protect"

var sourceKey = types.NamespacedName{Namespace: "test-ns", Name: "source"}

func newSourceTestScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := resourcev1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add resource scheme: %v", err)
	}
	return scheme
}

// newDeletingTestSource returns a PulsarSource that is mid-deletion: it carries a
// deletionTimestamp and the operator finalizer. The fake client permits creating an object
// with a deletionTimestamp as long as it still has at least one finalizer.
func newDeletingTestSource() *resourcev1alpha1.PulsarSource {
	deletionTime := metav1.NewTime(time.Unix(1700000000, 0))
	return &resourcev1alpha1.PulsarSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "source",
			Namespace:         "test-ns",
			Generation:        1,
			Finalizers:        []string{resourcev1alpha1.FinalizerName},
			DeletionTimestamp: &deletionTime,
		},
		Spec: resourcev1alpha1.PulsarSourceSpec{
			Tenant:          "public",
			Namespace:       "default",
			Name:            "source",
			LifecyclePolicy: resourcev1alpha1.CleanUpAfterDeletion,
		},
	}
}

func newSourceReconciler(t *testing.T, k8sClient client.Client) *PulsarSourceReconciler {
	t.Helper()
	return &PulsarSourceReconciler{
		conn: &PulsarConnectionReconciler{
			connection: newReadyTestConnection(),
			log:        logr.Discard(),
			client:     k8sClient,
			apiReader:  k8sClient,
		},
		log: logr.Discard(),
	}
}

// cachedSourceAfterConcurrentWrite returns the source as the reconciler would have cached
// it, then advances the server-side object out of band (simulating informer-cache lag or a
// concurrent writer such as ArgoCD) so the returned copy is stale relative to the server.
func cachedSourceAfterConcurrentWrite(t *testing.T, k8sClient client.Client, mutateLive func(*resourcev1alpha1.PulsarSource)) *resourcev1alpha1.PulsarSource {
	t.Helper()
	cached := &resourcev1alpha1.PulsarSource{}
	if err := k8sClient.Get(context.Background(), sourceKey, cached); err != nil {
		t.Fatalf("get cached source: %v", err)
	}

	live := cached.DeepCopy()
	mutateLive(live)
	if err := k8sClient.Update(context.Background(), live); err != nil {
		t.Fatalf("out-of-band update to advance server state: %v", err)
	}
	return cached
}

// TestReconcileSourceDeletionRemovesFinalizerWithStaleResourceVersion is the regression
// test for the original stuck-Terminating bug: deleting a PulsarSource must remove the
// operator finalizer and let the object be garbage-collected promptly, even when the
// reconciler's in-memory copy is stale relative to the server (informer-cache lag or a
// concurrent writer). The reconcile reads the live object under conflict-retry, so it does
// not get wedged on a 409.
func TestReconcileSourceDeletionRemovesFinalizerWithStaleResourceVersion(t *testing.T) {
	scheme := newSourceTestScheme(t)
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&resourcev1alpha1.PulsarSource{}).
		WithObjects(newDeletingTestSource()).
		Build()

	// Capture the cached copy, then advance the live object so the cached copy is stale.
	stale := cachedSourceAfterConcurrentWrite(t, fakeClient, func(live *resourcev1alpha1.PulsarSource) {
		live.Annotations = map[string]string{"out-of-band": "bump"}
	})

	r := newSourceReconciler(t, fakeClient)
	// DummyPulsarAdmin.DeletePulsarSource is a no-op; the assertion targets the finalizer path.
	if err := r.ReconcileSource(context.Background(), &admin.DummyPulsarAdmin{}, stale); err != nil {
		t.Fatalf("ReconcileSource on a deleting source with a stale resourceVersion returned an error "+
			"(want nil, no 409 conflict): %v", err)
	}

	err := fakeClient.Get(context.Background(), sourceKey, &resourcev1alpha1.PulsarSource{})
	if !apierrors.IsNotFound(err) {
		t.Fatalf("expected source to be garbage-collected after finalizer removal, got err: %v", err)
	}
}

// TestReconcileSourceDeletionPreservesForeignFinalizer is the regression test for the
// finalizer-clobbering hazard. The reconciler's cached copy only knows about the operator
// finalizer, while another controller has concurrently added its own finalizer to the live
// object. Removing the operator finalizer must remove ONLY the operator finalizer — it must
// not clobber the foreign finalizer (a JSON merge patch from a stale base would replace the
// whole list and drop it) and must not delete the object while the foreign finalizer is
// still pending.
func TestReconcileSourceDeletionPreservesForeignFinalizer(t *testing.T) {
	scheme := newSourceTestScheme(t)
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&resourcev1alpha1.PulsarSource{}).
		WithObjects(newDeletingTestSource()).
		Build()

	// The reconciler's cached copy has only the operator finalizer; a concurrent controller
	// adds a foreign finalizer to the live object, so the cached copy is now stale.
	stale := cachedSourceAfterConcurrentWrite(t, fakeClient, func(live *resourcev1alpha1.PulsarSource) {
		controllerutil.AddFinalizer(live, foreignFinalizer)
	})

	r := newSourceReconciler(t, fakeClient)
	if err := r.ReconcileSource(context.Background(), &admin.DummyPulsarAdmin{}, stale); err != nil {
		t.Fatalf("ReconcileSource on a deleting source returned an error: %v", err)
	}

	got := &resourcev1alpha1.PulsarSource{}
	if err := fakeClient.Get(context.Background(), sourceKey, got); err != nil {
		t.Fatalf("source must survive while the foreign finalizer is pending, but Get failed: %v", err)
	}
	if containsString(got.Finalizers, resourcev1alpha1.FinalizerName) {
		t.Fatalf("operator finalizer should have been removed, got %v", got.Finalizers)
	}
	if !containsString(got.Finalizers, foreignFinalizer) {
		t.Fatalf("foreign finalizer must be preserved, got %v", got.Finalizers)
	}
}
