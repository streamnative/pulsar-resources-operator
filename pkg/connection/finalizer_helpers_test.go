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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
)

// TestRemoveFinalizerUsesLiveReaderResourceVersion proves the finalizer write is driven by the
// reader, not by the caller's (possibly stale) in-hand object. A reader that keeps returning a
// stale resourceVersion — as the manager's cached client can under informer lag — makes every
// retry 409, so the removal fails (the original stuck-Terminating symptom). The same removal
// with a live reader reads a fresh resourceVersion and succeeds. In production the reader is the
// uncached mgr.GetAPIReader(), so a lagging informer cache cannot wedge deletion.
//
// This is the case ordinary single-fake-client tests cannot exercise: here the reader and the
// write store are deliberately decoupled to simulate cache lag.
func TestRemoveFinalizerUsesLiveReaderResourceVersion(t *testing.T) {
	scheme := newSourceTestScheme(t)
	topic := &resourcev1alpha1.PulsarTopic{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "t",
			Namespace:  "ns",
			Finalizers: []string{resourcev1alpha1.FinalizerName},
		},
	}
	writer := fake.NewClientBuilder().WithScheme(scheme).WithObjects(topic).Build()
	key := client.ObjectKeyFromObject(topic)

	// Snapshot the object at its current resourceVersion, then advance the live RV out of band.
	staleObj := &resourcev1alpha1.PulsarTopic{}
	if err := writer.Get(context.Background(), key, staleObj); err != nil {
		t.Fatalf("get snapshot: %v", err)
	}
	bumped := staleObj.DeepCopy()
	bumped.Annotations = map[string]string{"bump": "1"}
	if err := writer.Update(context.Background(), bumped); err != nil {
		t.Fatalf("bump live resourceVersion: %v", err)
	}

	// staleReader always returns the pre-bump snapshot — a persistently stale resourceVersion.
	staleReader := interceptor.NewClient(writer, interceptor.Funcs{
		Get: func(_ context.Context, _ client.WithWatch, _ client.ObjectKey, obj client.Object, _ ...client.GetOption) error {
			staleObj.DeepCopyInto(obj.(*resourcev1alpha1.PulsarTopic))
			return nil
		},
	})

	// Persistently stale reader: every retry re-reads the stale RV and the Update conflicts.
	if err := removeFinalizer(context.Background(), staleReader, writer, staleObj.DeepCopy(),
		resourcev1alpha1.FinalizerName); !apierrors.IsConflict(err) {
		t.Fatalf("expected a 409 conflict from a persistently stale reader, got: %v", err)
	}

	// Live (uncached) reader: the removal reads a fresh RV and succeeds.
	if err := removeFinalizer(context.Background(), writer, writer, staleObj.DeepCopy(),
		resourcev1alpha1.FinalizerName); err != nil {
		t.Fatalf("removeFinalizer with a live reader: %v", err)
	}
	got := &resourcev1alpha1.PulsarTopic{}
	if err := writer.Get(context.Background(), key, got); err != nil {
		t.Fatalf("get after removal: %v", err)
	}
	if controllerutil.ContainsFinalizer(got, resourcev1alpha1.FinalizerName) {
		t.Fatalf("operator finalizer should have been removed, got %v", got.Finalizers)
	}
}

// TestEnsureFinalizerGuardSkipsReadWhenPresent verifies the steady-state fast path: when the
// cached object already carries the finalizer, ensureFinalizer issues no API calls at all, so
// routing the read through the uncached API reader adds no per-reconcile apiserver load on the
// common path.
func TestEnsureFinalizerGuardSkipsReadWhenPresent(t *testing.T) {
	scheme := newSourceTestScheme(t)
	writer := fake.NewClientBuilder().WithScheme(scheme).Build()
	panicReader := interceptor.NewClient(writer, interceptor.Funcs{
		Get: func(_ context.Context, _ client.WithWatch, _ client.ObjectKey, _ client.Object, _ ...client.GetOption) error {
			panic("reader.Get must not be called when the finalizer is already present")
		},
	})

	obj := &resourcev1alpha1.PulsarTopic{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "t",
			Namespace:  "ns",
			Finalizers: []string{resourcev1alpha1.FinalizerName},
		},
	}
	if err := ensureFinalizer(context.Background(), panicReader, writer, obj, resourcev1alpha1.FinalizerName); err != nil {
		t.Fatalf("ensureFinalizer (guard fast path) returned error: %v", err)
	}
}
