// Copyright (c) 2022 StreamNative, Inc.. All Rights Reserved.

package connection

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	commonsreconciler "github.com/streamnative/pulsar-operators/commons/pkg/controller/reconciler"
	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	"github.com/streamnative/pulsar-resources-operator/pkg/admin"
)

// PulsarNamespaceReconciler reconciles a PulsarNamespace object
type PulsarNamespaceReconciler struct {
	conn *PulsarConnectionReconciler
}

func makeNamespacesReconciler(r *PulsarConnectionReconciler) commonsreconciler.Interface {
	return &PulsarNamespaceReconciler{
		conn: r,
	}
}

// Observe checks the updates of object
func (r *PulsarNamespaceReconciler) Observe(ctx context.Context) error {
	namespaceList := &resourcev1alpha1.PulsarNamespaceList{}
	if err := r.conn.client.List(ctx, namespaceList, client.InNamespace(r.conn.connection.Namespace),
		client.MatchingFields(map[string]string{
			".spec.connectionRef.name": r.conn.connection.Name,
		})); err != nil {
		return fmt.Errorf("list namespaces [%w]", err)
	}
	r.conn.namespaces = namespaceList.Items
	if !r.conn.hasUnreadyResource {
		for i := range r.conn.namespaces {
			if !resourcev1alpha1.IsPulsarResourceReady(&r.conn.namespaces[i]) {
				r.conn.hasUnreadyResource = true
				break
			}
		}
	}
	return nil
}

// Reconcile reconciles all namespaces
func (r *PulsarNamespaceReconciler) Reconcile(ctx context.Context) error {
	for i := range r.conn.namespaces {
		namespace := &r.conn.namespaces[i]
		if err := r.ReconcileNamespace(ctx, r.conn.pulsarAdmin, namespace); err != nil {
			return fmt.Errorf("reconcile namespace [%w]", err)
		}
	}
	return nil
}

// ReconcileNamespace move the current state of the toic closer to the desired state
func (r *PulsarNamespaceReconciler) ReconcileNamespace(ctx context.Context, pulsarAdmin admin.PulsarAdmin,
	namespace *resourcev1alpha1.PulsarNamespace) error {
	if !namespace.DeletionTimestamp.IsZero() {
		if namespace.Spec.LifecyclePolicy == resourcev1alpha1.CleanUpAfterDeletion {
			if err := pulsarAdmin.DeleteNamespace(namespace.Spec.Name); err != nil && admin.IsNotFound(err) {
				r.conn.log.Error(err, "Failed to delete namespace", "Namespace", namespace.Namespace, "Name", namespace.Name)
				return err
			}
		}
		// TODO use otelcontroller until kube-instrumentation upgrade controller-runtime version to newer
		controllerutil.RemoveFinalizer(namespace, resourcev1alpha1.FinalizerName)
		if err := r.conn.client.Update(ctx, namespace); err != nil {
			r.conn.log.Error(err, "Failed to remove finalizer", "Namespace", namespace.Namespace, "Name", namespace.Name)
			return err
		}

		return nil
	}

	if namespace.Spec.LifecyclePolicy == resourcev1alpha1.CleanUpAfterDeletion {
		// TODO use otelcontroller until kube-instrumentation upgrade controller-runtime version to newer
		controllerutil.AddFinalizer(namespace, resourcev1alpha1.FinalizerName)
		if err := r.conn.client.Update(ctx, namespace); err != nil {
			r.conn.log.Error(err, "Failed to add finalizer", "Namespace", namespace.Namespace, "Name", namespace.Name)
			return err
		}
	}

	if resourcev1alpha1.IsPulsarResourceReady(namespace) {
		r.conn.log.V(1).Info("Resource is ready", "Namespace", namespace.Namespace, "Name", namespace.Name)
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
		r.conn.log.Error(err, "Failed to apply namespace", "Namespace", namespace.Namespace, "Name", namespace.Name)
		if err := r.conn.client.Status().Update(ctx, namespace); err != nil {
			r.conn.log.Error(err, "Failed to update the namespace status", "Namespace", namespace.Namespace,
				"Name", namespace.Name)
			return nil
		}
		return err
	}

	namespace.Status.ObservedGeneration = namespace.Generation
	meta.SetStatusCondition(&namespace.Status.Conditions, *NewReadyCondition(namespace.Generation))
	if err := r.conn.client.Status().Update(ctx, namespace); err != nil {
		r.conn.log.Error(err, "Failed to update the namespace status", "Namespace", namespace.Namespace,
			"Name", namespace.Name)
		return err
	}

	return nil
}
