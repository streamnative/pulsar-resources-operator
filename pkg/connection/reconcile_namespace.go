// Copyright (c) 2020 StreamNative, Inc.. All Rights Reserved.

package connection

import (
	"context"
	"fmt"

	commonsreconciler "github.com/streamnative/pulsar-operators/commons/pkg/controller/reconciler"
	pulsarv1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	"github.com/streamnative/pulsar-resources-operator/pkg/admin"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type PulsarNamespaceReconciler struct {
	conn *PulsarConnectionReconciler
}

func makeNamespacesReconciler(r *PulsarConnectionReconciler) commonsreconciler.Interface {
	return &PulsarNamespaceReconciler{
		conn: r,
	}
}

func (r *PulsarNamespaceReconciler) Observe(ctx context.Context) error {
	namespaceList := &pulsarv1alpha1.PulsarNamespaceList{}
	if err := r.conn.client.List(ctx, namespaceList, client.InNamespace(r.conn.connection.Namespace),
		client.MatchingFields(map[string]string{
			".spec.connectionRef.name": r.conn.connection.Name,
		})); err != nil {
		return fmt.Errorf("list namespaces [%w]", err)
	}
	r.conn.namespaces = namespaceList.Items
	if !r.conn.hasUnreadyResource {
		for i := range r.conn.namespaces {
			if !pulsarv1alpha1.IsPulsarResourceReady(&r.conn.namespaces[i]) {
				r.conn.hasUnreadyResource = true
				break
			}
		}
	}
	return nil
}

func (r *PulsarNamespaceReconciler) Reconcile(ctx context.Context) error {
	for i := range r.conn.namespaces {
		namespace := &r.conn.namespaces[i]
		if err := r.ReconcileNamespace(ctx, r.conn.pulsarAdmin, namespace); err != nil {
			return fmt.Errorf("reconcile namespace [%w]", err)
		}
	}
	return nil
}

func (r *PulsarNamespaceReconciler) ReconcileNamespace(ctx context.Context, pulsarAdmin admin.PulsarAdmin,
	namespace *pulsarv1alpha1.PulsarNamespace) error {
	if !namespace.DeletionTimestamp.IsZero() {
		if namespace.Spec.LifecyclePolicy == pulsarv1alpha1.CleanUpAfterDeletion {
			if err := pulsarAdmin.DeleteNamespace(namespace.Spec.Name); err != nil && admin.IsNotFound(err) {
				return err
			}
		}
		// TODO use otelcontroller until kube-instrumentation upgrade controller-runtime version to newer
		controllerutil.RemoveFinalizer(namespace, pulsarv1alpha1.FinalizerName)
		if err := r.conn.client.Update(ctx, namespace); err != nil {
			return err
		}

		return nil
	}

	if namespace.Spec.LifecyclePolicy == pulsarv1alpha1.CleanUpAfterDeletion {
		// TODO use otelcontroller until kube-instrumentation upgrade controller-runtime version to newer
		controllerutil.AddFinalizer(namespace, pulsarv1alpha1.FinalizerName)
		if err := r.conn.client.Update(ctx, namespace); err != nil {
			return err
		}
	}

	if pulsarv1alpha1.IsPulsarResourceReady(namespace) {
		return nil
	}

	params := &admin.NamespaceParams{
		Bundles:                     namespace.Spec.Bundles,
		MaxProducersPerTopic:        namespace.Spec.MaxProducersPerTopic,
		MaxConsumersPerTopic:        namespace.Spec.MaxConsumersPerTopic,
		MaxConsumersPerSubscription: namespace.Spec.MaxConsumersPerSubscription,
		MessageTTL:                  namespace.Spec.MessageTTL,
		RetentionTime:               namespace.Spec.RetentionTime,
		RetentionSize:               namespace.Spec.RetentionSize,
		BacklogQuotaLimitTime:       namespace.Spec.BacklogQuotaLimitTime,
		BacklogQuotaLimitSize:       namespace.Spec.BacklogQuotaLimitSize,
		BacklogQuotaRetentionPolicy: namespace.Spec.BacklogQuotaRetentionPolicy,
	}

	if err := pulsarAdmin.ApplyNamespace(namespace.Spec.Name, params); err != nil {
		meta.SetStatusCondition(&namespace.Status.Conditions, *NewErrorCondition(namespace.Generation, err.Error()))
		if err := r.conn.client.Status().Update(ctx, namespace); err != nil {
			r.conn.log.Error(err, "Failed to update connection status", "Namespace", r.conn.connection.Namespace,
				"Name", r.conn.connection.Name)
			return nil
		}
		return err
	}

	namespace.Status.ObservedGeneration = namespace.Generation
	meta.SetStatusCondition(&namespace.Status.Conditions, *NewReadyCondition(namespace.Generation))
	if err := r.conn.client.Status().Update(ctx, namespace); err != nil {
		return err
	}

	return nil
}
