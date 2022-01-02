// Copyright (c) 2020 StreamNative, Inc.. All Rights Reserved.

package connection

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	commonsreconciler "github.com/streamnative/pulsar-operators/commons/pkg/controller/reconciler"
	pulsarv1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	"github.com/streamnative/pulsar-resources-operator/pkg/admin"
)

type PulsarTenantReconciler struct {
	conn *PulsarConnectionReconciler
}

func makeTenantsReconciler(r *PulsarConnectionReconciler) commonsreconciler.Interface {
	return &PulsarTenantReconciler{
		conn: r,
	}
}

func (r *PulsarTenantReconciler) Observe(ctx context.Context) error {
	tenantList := &pulsarv1alpha1.PulsarTenantList{}
	if err := r.conn.client.List(ctx, tenantList, client.InNamespace(r.conn.connection.Namespace),
		client.MatchingFields(map[string]string{
			".spec.connectionRef.name": r.conn.connection.Name,
		})); err != nil {
		return fmt.Errorf("list tenants [%w]", err)
	}
	r.conn.tenants = tenantList.Items
	if !r.conn.hasUnreadyResource {
		for i := range r.conn.tenants {
			if !pulsarv1alpha1.IsPulsarResourceReady(&r.conn.tenants[i]) {
				r.conn.hasUnreadyResource = true
				break
			}
		}
	}
	return nil
}

func (r *PulsarTenantReconciler) Reconcile(ctx context.Context) error {
	for i := range r.conn.tenants {
		tenant := &r.conn.tenants[i]
		if err := r.ReconcileTenant(ctx, r.conn.pulsarAdmin, tenant); err != nil {
			return fmt.Errorf("reconcile tenant [%w]", err)
		}
	}
	return nil
}

func (r *PulsarTenantReconciler) ReconcileTenant(ctx context.Context, pulsarAdmin admin.PulsarAdmin,
	tenant *pulsarv1alpha1.PulsarTenant) error {
	if !tenant.DeletionTimestamp.IsZero() {
		if tenant.Spec.LifecyclePolicy == pulsarv1alpha1.CleanUpAfterDeletion {
			if err := pulsarAdmin.DeleteTenant(tenant.Spec.Name); err != nil && admin.IsNotFound(err) {
				return err
			}
		}

		// TODO use otelcontroller until kube-instrumentation upgrade controller-runtime version to newer
		controllerutil.RemoveFinalizer(tenant, pulsarv1alpha1.FinalizerName)
		if err := r.conn.client.Update(ctx, tenant); err != nil {
			return err
		}

		return nil
	}

	// TODO use otelcontroller until kube-instrumentation upgrade controller-runtime version to newer
	controllerutil.AddFinalizer(tenant, pulsarv1alpha1.FinalizerName)
	if err := r.conn.client.Update(ctx, tenant); err != nil {
		return err
	}

	if pulsarv1alpha1.IsPulsarResourceReady(tenant) {
		return nil
	}

	tenantParams := &admin.TenantParams{
		AdminRoles:      tenant.Spec.AdminRoles,
		AllowedClusters: tenant.Spec.AllowedClusters,
	}
	if err := pulsarAdmin.ApplyTenant(tenant.Spec.Name, tenantParams); err != nil {
		meta.SetStatusCondition(&tenant.Status.Conditions, *NewErrorCondition(tenant.Generation, err.Error()))
		if err := r.conn.client.Status().Update(ctx, tenant); err != nil {
			r.conn.log.Error(err, "Failed to update connection status", "Namespace", r.conn.connection.Namespace,
				"Name", r.conn.connection.Name)
			return nil
		}
		return err
	}

	tenant.Status.ObservedGeneration = tenant.Generation
	meta.SetStatusCondition(&tenant.Status.Conditions, *NewReadyCondition(tenant.Generation))
	if err := r.conn.client.Status().Update(ctx, tenant); err != nil {
		return err
	}

	return nil
}
