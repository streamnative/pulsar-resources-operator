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

	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func shouldKeepRemoteResource(policy resourcev1alpha1.PulsarResourceLifeCyclePolicy) bool {
	return policy == resourcev1alpha1.KeepAfterDeletion
}

func finalizeKeptResource(
	ctx context.Context,
	k8sClient client.Client,
	obj client.Object,
	finalizer string,
	resourceKind string,
) (ctrl.Result, error) {
	if !controllerutil.ContainsFinalizer(obj, finalizer) {
		return ctrl.Result{}, nil
	}

	log.FromContext(ctx).Info(
		"Skipping remote cleanup and removing finalizer due to lifecycle policy",
		"resource", resourceKind,
		"namespace", obj.GetNamespace(),
		"name", obj.GetName(),
		"lifecyclePolicy", resourcev1alpha1.KeepAfterDeletion,
	)

	controllerutil.RemoveFinalizer(obj, finalizer)
	if err := k8sClient.Update(ctx, obj); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}
