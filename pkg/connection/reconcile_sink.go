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

// PulsarSinkReconciler reconciles a PulsarSink object
type PulsarSinkReconciler struct {
	conn *PulsarConnectionReconciler
	log  logr.Logger
}

func makeSinksReconciler(r *PulsarConnectionReconciler) reconciler.Interface {
	return &PulsarSinkReconciler{
		conn: r,
		log:  makeSubResourceLog(r, "PulsarSink"),
	}
}

// Observe checks the updates of object
func (r *PulsarSinkReconciler) Observe(ctx context.Context) error {
	r.log.V(1).Info("Start Observe")

	sinksList := &resourcev1alpha1.PulsarSinkList{}
	if err := r.conn.client.List(ctx, sinksList, client.InNamespace(r.conn.connection.Namespace),
		client.MatchingFields(map[string]string{
			".spec.connectionRef.name": r.conn.connection.Name,
		})); err != nil {
		return fmt.Errorf("list sinks [%w]", err)
	}
	r.log.V(1).Info("Observed sinks items", "Count", len(sinksList.Items))

	r.conn.sinks = sinksList.Items
	for i := range r.conn.sinks {
		if !resourcev1alpha1.IsPulsarResourceReady(&r.conn.sinks[i]) {
			r.conn.addUnreadyResource(&r.conn.sinks[i])
		}
	}

	r.log.V(1).Info("Observe Done")
	return nil
}

// Reconcile reconciles the object
func (r *PulsarSinkReconciler) Reconcile(ctx context.Context) error {
	r.log.V(1).Info("Start Reconcile")

	for i := range r.conn.sinks {
		sink := &r.conn.sinks[i]
		if err := r.ReconcileSink(ctx, r.conn.pulsarAdminV3, sink); err != nil {
			return fmt.Errorf("reconcile sink [%s] [%w]", sink.Name, err)
		}
	}

	return nil
}

// ReconcileSink reconciles the sink
func (r *PulsarSinkReconciler) ReconcileSink(ctx context.Context, pulsarAdmin admin.PulsarAdmin, sink *resourcev1alpha1.PulsarSink) error {
	log := r.log.WithValues("name", sink.Name, "namespace", sink.Namespace)
	log.Info("Start Reconcile")

	if !sink.DeletionTimestamp.IsZero() {
		log.Info("Deleting sink", "LifecyclePolicy", sink.Spec.LifecyclePolicy)

		if sink.Spec.LifecyclePolicy != resourcev1alpha1.KeepAfterDeletion {
			if err := pulsarAdmin.DeletePulsarSink(sink.Spec.Tenant, sink.Spec.Namespace, sink.Spec.Name); err != nil && !admin.IsNotFound(err) {
				log.Error(err, "Failed to delete sink")
				return err
			}
		}

		// TODO use otelcontroller until kube-instrumentation upgrade controller-runtime version to newer
		controllerutil.RemoveFinalizer(sink, resourcev1alpha1.FinalizerName)
		if err := r.conn.client.Update(ctx, sink); err != nil {
			log.Error(err, "Failed to remove finalizer")
			return err
		}
		return nil
	}

	if sink.Spec.LifecyclePolicy != resourcev1alpha1.KeepAfterDeletion {
		// TODO use otelcontroller until kube-instrumentation upgrade controller-runtime version to newer
		controllerutil.AddFinalizer(sink, resourcev1alpha1.FinalizerName)
		if err := r.conn.client.Update(ctx, sink); err != nil {
			log.Error(err, "Failed to add finalizer")
			return err
		}
	}

	if resourcev1alpha1.IsPulsarResourceReady(sink) &&
		!feature.DefaultFeatureGate.Enabled(feature.AlwaysUpdatePulsarResource) {
		log.Info("Skip reconcile, function resource is ready")
		return nil
	}

	if sink.Spec.Archive == nil && sink.Spec.SinkType == "" {
		err := errors.New("no package URL or sink type provided")
		return err
	}

	packageURL := ""

	if sink.Spec.Archive.URL != "" {
		packageURL = validateURL(sink.Spec.Archive.URL)
		if packageURL == "" {
			err := errors.New("invalid package URL")
			return err
		}
	}

	if err := pulsarAdmin.ApplyPulsarSink(sink.Spec.Tenant, sink.Spec.Namespace, sink.Spec.Name, packageURL, &sink.Spec, sink.Status.ObservedGeneration > 1); err != nil {
		meta.SetStatusCondition(&sink.Status.Conditions, *NewErrorCondition(sink.Generation, err.Error()))
		log.Error(err, "Failed to apply sink")
		if err := r.conn.client.Status().Update(ctx, sink); err != nil {
			log.Error(err, "Failed to update the sink status")
			return nil
		}
		return err
	}

	sink.Status.ObservedGeneration = sink.Generation
	meta.SetStatusCondition(&sink.Status.Conditions, *NewReadyCondition(sink.Generation))
	if err := r.conn.client.Status().Update(ctx, sink); err != nil {
		log.Error(err, "Failed to update the sink status")
		return err
	}
	return nil
}
