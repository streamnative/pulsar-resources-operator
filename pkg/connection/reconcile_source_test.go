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

// finalizerCountingClient wraps a client.Client and counts full-object Update and Patch
// calls, so a test can assert that finalizer changes are persisted via a MergeFrom Patch
// (which carries no resourceVersion precondition) and never via a full-object Update.
type finalizerCountingClient struct {
	client.Client
	updateCalls int
	patchCalls  int
}

func (c *finalizerCountingClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	c.updateCalls++
	return c.Client.Update(ctx, obj, opts...)
}

func (c *finalizerCountingClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	c.patchCalls++
	return c.Client.Patch(ctx, obj, patch, opts...)
}

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

// staleSourceCopyAfterConcurrentWrite returns the source as the reconciler would have
// cached it, then advances the server-side resourceVersion out of band (simulating a
// concurrent writer such as ArgoCD) so the returned copy is stale relative to the server.
// This is the exact condition under which a full-object Update fails with a 409 conflict.
func staleSourceCopyAfterConcurrentWrite(t *testing.T, k8sClient client.Client) *resourcev1alpha1.PulsarSource {
	t.Helper()
	cached := &resourcev1alpha1.PulsarSource{}
	if err := k8sClient.Get(context.Background(),
		types.NamespacedName{Namespace: "test-ns", Name: "source"}, cached); err != nil {
		t.Fatalf("get cached source: %v", err)
	}

	live := cached.DeepCopy()
	live.Annotations = map[string]string{"out-of-band": "bump"}
	if err := k8sClient.Update(context.Background(), live); err != nil {
		t.Fatalf("out-of-band update to bump resourceVersion: %v", err)
	}
	return cached
}

// TestReconcileSourceDeletionRemovesFinalizerWithStaleResourceVersion is the regression
// test for the finalizer-conflict fix. When a PulsarSource is being deleted, the operator
// must remove its finalizer via a MergeFrom Patch (no resourceVersion precondition) even
// when the in-memory copy is stale, so the object is garbage-collected promptly without a
// 409 conflict — regardless of informer-cache lag or concurrent writers (e.g. GitOps).
//
// Against the pre-fix implementation (a full-object Update) this test fails: the Update
// carries the stale resourceVersion and the fake API server rejects it with a 409.
func TestReconcileSourceDeletionRemovesFinalizerWithStaleResourceVersion(t *testing.T) {
	scheme := newSourceTestScheme(t)
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&resourcev1alpha1.PulsarSource{}).
		WithObjects(newDeletingTestSource()).
		Build()

	// Capture the cached copy, then move the server resourceVersion ahead of it.
	stale := staleSourceCopyAfterConcurrentWrite(t, fakeClient)

	countingClient := &finalizerCountingClient{Client: fakeClient}
	r := &PulsarSourceReconciler{
		conn: &PulsarConnectionReconciler{
			connection: newReadyTestConnection(),
			log:        logr.Discard(),
			client:     countingClient,
		},
		log: logr.Discard(),
	}

	// DummyPulsarAdmin.DeletePulsarSource is a no-op; the assertion targets the finalizer path.
	if err := r.ReconcileSource(context.Background(), &admin.DummyPulsarAdmin{}, stale); err != nil {
		t.Fatalf("ReconcileSource on a deleting source with a stale resourceVersion returned an error "+
			"(want nil, no 409 conflict): %v", err)
	}

	if countingClient.updateCalls != 0 {
		t.Fatalf("finalizer removal must not use a full-object Update (it carries a resourceVersion "+
			"precondition); got %d Update call(s)", countingClient.updateCalls)
	}
	if countingClient.patchCalls == 0 {
		t.Fatal("expected finalizer removal to persist via a MergeFrom Patch, got 0 Patch calls")
	}

	got := &resourcev1alpha1.PulsarSource{}
	err := fakeClient.Get(context.Background(),
		types.NamespacedName{Namespace: "test-ns", Name: "source"}, got)
	if err == nil {
		t.Fatalf("expected source to be garbage-collected after finalizer removal, "+
			"but it still exists with finalizers %v", got.Finalizers)
	}
	if !apierrors.IsNotFound(err) {
		t.Fatalf("expected NotFound after finalizer removal, got: %v", err)
	}
}

// TestSourceStaleFinalizerUpdateConflicts documents the pre-fix failure mode and proves the
// test harness genuinely surfaces it: removing the finalizer with a full-object Update from a
// stale-resourceVersion copy returns a 409 conflict. This is precisely the conflict that the
// MergeFrom Patch fix avoids, and it guards against any regression back to Update-based
// finalizer writes.
func TestSourceStaleFinalizerUpdateConflicts(t *testing.T) {
	scheme := newSourceTestScheme(t)
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&resourcev1alpha1.PulsarSource{}).
		WithObjects(newDeletingTestSource()).
		Build()

	stale := staleSourceCopyAfterConcurrentWrite(t, fakeClient)

	controllerutil.RemoveFinalizer(stale, resourcev1alpha1.FinalizerName)
	err := fakeClient.Update(context.Background(), stale)
	if err == nil || !apierrors.IsConflict(err) {
		t.Fatalf("expected a 409 conflict from a stale full-object Update, got: %v", err)
	}
}
