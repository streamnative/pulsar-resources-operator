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

	"github.com/go-logr/logr"
	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	"github.com/streamnative/pulsar-resources-operator/pkg/admin"
	"github.com/streamnative/pulsar-resources-operator/pkg/feature"
	"github.com/streamnative/pulsar-resources-operator/pkg/reconciler"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// PulsarSourceReconciler reconciles a PulsarSource object
type PulsarSourceReconciler struct {
	conn *PulsarConnectionReconciler
	log  logr.Logger
}

func makeSourcesReconciler(r *PulsarConnectionReconciler) reconciler.Interface {
	return &PulsarSourceReconciler{
		conn: r,
		log:  makeSubResourceLog(r, "PulsarSource"),
	}
}

// Observe checks the updates of object
func (r *PulsarSourceReconciler) Observe(ctx context.Context) error {
	r.log.V(1).Info("Start Observe")

	sourcesList := &resourcev1alpha1.PulsarSourceList{}
	if err := r.conn.client.List(ctx, sourcesList, client.InNamespace(r.conn.connection.Namespace),
		client.MatchingFields(map[string]string{
			".spec.connectionRef.name": r.conn.connection.Name,
		})); err != nil {
		return fmt.Errorf("list sources [%w]", err)
	}
	r.log.V(1).Info("Observed sources items", "Count", len(sourcesList.Items))

	r.conn.sources = sourcesList.Items
	r.log.Info("Observe sources", "Count", len(r.conn.sources))
	for i := range r.conn.sources {
		if !resourcev1alpha1.IsPulsarResourceReady(&r.conn.sources[i]) {
			r.conn.addUnreadyResource(&r.conn.sources[i])
		}
	}

	r.log.V(1).Info("Observe Done")
	return nil
}

// Reconcile reconciles the object
func (r *PulsarSourceReconciler) Reconcile(ctx context.Context) error {
	r.log.V(1).Info("Start Reconcile")

	for i := range r.conn.sources {
		source := &r.conn.sources[i]
		r.log.Info("Reconcile source", "Name", source.Name)
		if err := r.ReconcileSource(ctx, r.conn.pulsarAdminV3, source); err != nil {
			return fmt.Errorf("reconcile source [%s] [%w]", source.Name, err)
		}
	}

	return nil
}

// ReconcileSource reconciles the source
func (r *PulsarSourceReconciler) ReconcileSource(ctx context.Context, pulsarAdmin admin.PulsarAdmin, source *resourcev1alpha1.PulsarSource) error {
	log := r.log.WithValues("name", source.Name, "namespace", source.Namespace)
	log.Info("Start Reconcile")

	if !source.DeletionTimestamp.IsZero() {
		log.Info("Deleting source", "LifecyclePolicy", source.Spec.LifecyclePolicy)

		if source.Spec.LifecyclePolicy != resourcev1alpha1.KeepAfterDeletion {
			if err := pulsarAdmin.DeletePulsarSource(source.Spec.Tenant, source.Spec.Namespace, source.Spec.Name); err != nil && !admin.IsNotFound(err) {
				log.Error(err, "Failed to delete source")
				return err
			}
		}

		// TODO use otelcontroller until kube-instrumentation upgrade controller-runtime version to newer
		controllerutil.RemoveFinalizer(source, resourcev1alpha1.FinalizerName)
		if err := r.conn.client.Update(ctx, source); err != nil {
			log.Error(err, "Failed to remove finalizer")
			return err
		}
		return nil
	}

	if source.Spec.LifecyclePolicy != resourcev1alpha1.KeepAfterDeletion {
		// TODO use otelcontroller until kube-instrumentation upgrade controller-runtime version to newer
		controllerutil.AddFinalizer(source, resourcev1alpha1.FinalizerName)
		if err := r.conn.client.Update(ctx, source); err != nil {
			log.Error(err, "Failed to add finalizer")
			return err
		}
	}

	if resourcev1alpha1.IsPulsarResourceReady(source) &&
		!feature.DefaultFeatureGate.Enabled(feature.AlwaysUpdatePulsarResource) {
		log.Info("Skip reconcile, function resource is ready")
		return nil
	}

	if source.Spec.Archive == nil {
		err := errors.New("invalid package URL")
		return err
	}

	packageURL := validateURL(source.Spec.Archive.URL)

	if packageURL == "" {
		err := errors.New("invalid package URL")
		return err
	}

	if err := pulsarAdmin.ApplyPulsarSource(source.Spec.Tenant, source.Spec.Namespace, source.Spec.Name, packageURL, &source.Spec, source.Status.ObservedGeneration > 1); err != nil {
		meta.SetStatusCondition(&source.Status.Conditions, *NewErrorCondition(source.Generation, err.Error()))
		log.Error(err, "Failed to apply source")
		if err := r.conn.client.Status().Update(ctx, source); err != nil {
			log.Error(err, "Failed to update the source status")
			return nil
		}
		return err
	}

	source.Status.ObservedGeneration = source.Generation
	meta.SetStatusCondition(&source.Status.Conditions, *NewReadyCondition(source.Generation))
	if err := r.conn.client.Status().Update(ctx, source); err != nil {
		log.Error(err, "Failed to update the source status")
		return err
	}
	return nil
}
