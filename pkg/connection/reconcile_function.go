// Copyright 2022 StreamNative
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
	"errors"
	"fmt"
	"strings"

	"github.com/apache/pulsar-client-go/pulsaradmin/pkg/utils"
	"github.com/go-logr/logr"
	"github.com/streamnative/pulsar-resources-operator/pkg/feature"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	"github.com/streamnative/pulsar-resources-operator/pkg/admin"
	"github.com/streamnative/pulsar-resources-operator/pkg/reconciler"
)

// PulsarFunctionReconciler reconciles a PulsarFunction object
type PulsarFunctionReconciler struct {
	conn *PulsarConnectionReconciler
	log  logr.Logger
}

func makeFunctionsReconciler(r *PulsarConnectionReconciler) reconciler.Interface {
	return &PulsarFunctionReconciler{
		conn: r,
		log:  makeSubResourceLog(r, "PulsarFunction"),
	}
}

// Observe checks the updates of object
func (r *PulsarFunctionReconciler) Observe(ctx context.Context) error {
	r.log.V(1).Info("Start Observe")

	functionsList := &resourcev1alpha1.PulsarFunctionList{}
	if err := r.conn.client.List(ctx, functionsList, client.InNamespace(r.conn.connection.Namespace),
		client.MatchingFields(map[string]string{
			".spec.connectionRef.name": r.conn.connection.Name,
		})); err != nil {
		return fmt.Errorf("list functions [%w]", err)
	}
	r.log.V(1).Info("Observed function items", "Count", len(functionsList.Items))

	r.conn.functions = functionsList.Items
	for i := range r.conn.functions {
		if !resourcev1alpha1.IsPulsarResourceReady(&r.conn.functions[i]) {
			r.conn.addUnreadyResource(&r.conn.functions[i])
		}
	}

	r.log.V(1).Info("Observe Done")
	return nil
}

// Reconcile reconciles all functions
func (r *PulsarFunctionReconciler) Reconcile(ctx context.Context) error {
	for i := range r.conn.functions {
		instance := &r.conn.functions[i]
		if err := r.ReconcileFunction(ctx, r.conn.pulsarAdminV3, instance); err != nil {
			return fmt.Errorf("reconcile pulsar function [%w]", err)
		}
	}
	return nil
}

// ReconcileFunction move the current state of the functions closer to the desired state
func (r *PulsarFunctionReconciler) ReconcileFunction(ctx context.Context, pulsarAdmin admin.PulsarAdmin,
	instance *resourcev1alpha1.PulsarFunction) error {
	log := r.log.WithValues("name", instance.Name, "namespace", instance.Namespace)
	log.Info("Start Reconcile")

	if !instance.DeletionTimestamp.IsZero() {
		log.Info("Deleting function", "LifecyclePolicy", instance.Spec.LifecyclePolicy)

		if instance.Spec.LifecyclePolicy != resourcev1alpha1.KeepAfterDeletion {
			if err := pulsarAdmin.DeletePulsarFunction(instance.Spec.Tenant, instance.Spec.Namespace, instance.Spec.Name); err != nil && !admin.IsNotFound(err) {
				log.Error(err, "Failed to delete function")
				return err
			}
		}

		// TODO use otelcontroller until kube-instrumentation upgrade controller-runtime version to newer
		controllerutil.RemoveFinalizer(instance, resourcev1alpha1.FinalizerName)
		if err := r.conn.client.Update(ctx, instance); err != nil {
			log.Error(err, "Failed to remove finalizer")
			return err
		}
		return nil
	}

	if instance.Spec.LifecyclePolicy != resourcev1alpha1.KeepAfterDeletion {
		// TODO use otelcontroller until kube-instrumentation upgrade controller-runtime version to newer
		controllerutil.AddFinalizer(instance, resourcev1alpha1.FinalizerName)
		if err := r.conn.client.Update(ctx, instance); err != nil {
			log.Error(err, "Failed to add finalizer")
			return err
		}
	}

	if resourcev1alpha1.IsPulsarResourceReady(instance) &&
		!feature.DefaultFeatureGate.Enabled(feature.AlwaysUpdatePulsarResource) {
		log.Info("Skip reconcile, function resource is ready")
		return nil
	}

	packageURL := ""
	if instance.Spec.Jar != nil && instance.Spec.Jar.URL != "" {
		packageURL = validateURL(instance.Spec.Jar.URL)
	} else if instance.Spec.Py != nil && instance.Spec.Py.URL != "" {
		packageURL = validateURL(instance.Spec.Py.URL)
	} else if instance.Spec.Go != nil && instance.Spec.Go.URL != "" {
		packageURL = validateURL(instance.Spec.Go.URL)
	} else {
		err := errors.New("no package URL found")
		return err
	}

	if packageURL == "" {
		err := errors.New("invalid package URL")
		return err
	}

	if err := pulsarAdmin.ApplyPulsarFunction(instance.Spec.Tenant, instance.Spec.Namespace, instance.Spec.Name, packageURL, &instance.Spec, instance.Status.ObservedGeneration > 1); err != nil {
		meta.SetStatusCondition(&instance.Status.Conditions, *NewErrorCondition(instance.Generation, err.Error()))
		log.Error(err, "Failed to apply function")
		if err := r.conn.client.Status().Update(ctx, instance); err != nil {
			log.Error(err, "Failed to update the function status")
			return nil
		}
		return err
	}

	instance.Status.ObservedGeneration = instance.Generation
	meta.SetStatusCondition(&instance.Status.Conditions, *NewReadyCondition(instance.Generation))
	if err := r.conn.client.Status().Update(ctx, instance); err != nil {
		log.Error(err, "Failed to update the function status")
		return err
	}
	return nil
}

func validateURL(url string) string {
	if url == "" {
		return ""
	}
	if lowerURL := strings.ToLower(url); strings.HasPrefix(lowerURL, "http://") ||
		strings.HasPrefix(lowerURL, "https://") ||
		strings.HasPrefix(lowerURL, "builtin://") ||
		strings.HasPrefix(lowerURL, "file://") {
		return url
	}
	if _, err := utils.GetPackageName(url); err != nil {
		return ""
	}
	return url
}
