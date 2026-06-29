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

package connection

import (
	"context"

	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ensureFinalizer adds finalizer to obj, persisting the change against the live API object
// and retrying on optimistic-lock conflicts.
//
// reader must be an uncached reader (e.g. mgr.GetAPIReader()): the manager's default client
// reads through the informer cache, which can lag behind a recent write and hand back a stale
// resourceVersion, causing the Update to 409 and the retry to keep re-reading the same stale
// copy. Reading live guarantees a fresh resourceVersion so the write succeeds on the first
// reconcile regardless of cache lag. Writes use the normal writer client.
//
// Guard: when the cached copy of obj already contains finalizer this returns immediately with
// no API calls, so the common steady-state path (finalizer already present) pays nothing —
// only the first add issues the uncached read.
//
// Only this operator's finalizer is mutated, so finalizers added concurrently by other
// controllers are preserved. On success obj holds the server's latest state (including the
// refreshed resourceVersion), so callers may safely issue a follow-up Status().Update.
func ensureFinalizer(ctx context.Context, reader client.Reader, writer client.Client, obj client.Object, finalizer string) error {
	if controllerutil.ContainsFinalizer(obj, finalizer) {
		return nil
	}
	key := client.ObjectKeyFromObject(obj)
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := reader.Get(ctx, key, obj); err != nil {
			return err
		}
		if !controllerutil.AddFinalizer(obj, finalizer) {
			// Already present on the live object; nothing to write.
			return nil
		}
		return writer.Update(ctx, obj)
	})
}

// removeFinalizer removes finalizer from obj, persisting the change against the live API
// object and retrying on optimistic-lock conflicts.
//
// reader must be an uncached reader (e.g. mgr.GetAPIReader()). Re-reading the live object
// before each attempt yields a fresh resourceVersion, so deletion is not wedged by a stale
// informer cache (the original stuck-Terminating symptom), and only this operator's finalizer
// is removed — finalizers owned by other controllers are preserved and the object is not
// garbage-collected prematurely. Writes use the normal writer client.
//
// Guard: when the cached copy of obj does not contain finalizer this returns immediately with
// no API calls. A NotFound result (the object was garbage-collected once its last finalizer
// was removed) is treated as success.
func removeFinalizer(ctx context.Context, reader client.Reader, writer client.Client, obj client.Object, finalizer string) error {
	if !controllerutil.ContainsFinalizer(obj, finalizer) {
		return nil
	}
	key := client.ObjectKeyFromObject(obj)
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := reader.Get(ctx, key, obj); err != nil {
			return err
		}
		if !controllerutil.RemoveFinalizer(obj, finalizer) {
			// Already absent on the live object; nothing to write.
			return nil
		}
		return writer.Update(ctx, obj)
	})
	return client.IgnoreNotFound(err)
}
