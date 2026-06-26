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

// ensureFinalizer adds finalizer to obj, persisting the change against the live API
// object and retrying on optimistic-lock conflicts.
//
// It re-reads obj from the API server before each attempt and mutates only this
// operator's finalizer, so finalizers added concurrently by other controllers are
// preserved. A full-object Update with optimistic concurrency is used (not a JSON
// merge patch, which replaces the whole finalizers list and could drop a concurrent
// finalizer when applied from a stale base). On success obj holds the server's latest
// state, including the refreshed resourceVersion, so callers may safely issue a
// follow-up Status().Update(ctx, obj).
func ensureFinalizer(ctx context.Context, c client.Client, obj client.Object, finalizer string) error {
	key := client.ObjectKeyFromObject(obj)
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := c.Get(ctx, key, obj); err != nil {
			return err
		}
		if !controllerutil.AddFinalizer(obj, finalizer) {
			// Already present; nothing to write.
			return nil
		}
		return c.Update(ctx, obj)
	})
}

// removeFinalizer removes finalizer from obj, persisting the change against the live
// API object and retrying on optimistic-lock conflicts.
//
// It re-reads obj before each attempt and removes only this operator's finalizer, so it
// never clobbers finalizers owned by other controllers and never deletes the object
// prematurely while a foreign finalizer is still pending. A NotFound result (the object
// was garbage-collected once its last finalizer was gone) is treated as success.
func removeFinalizer(ctx context.Context, c client.Client, obj client.Object, finalizer string) error {
	key := client.ObjectKeyFromObject(obj)
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := c.Get(ctx, key, obj); err != nil {
			return err
		}
		if !controllerutil.RemoveFinalizer(obj, finalizer) {
			// Already absent; nothing to write.
			return nil
		}
		return c.Update(ctx, obj)
	})
	return client.IgnoreNotFound(err)
}
