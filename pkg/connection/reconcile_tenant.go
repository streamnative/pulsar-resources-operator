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

// PulsarTenantReconciler reconciles a PulsarTenant object
type PulsarTenantReconciler struct {
	conn *PulsarConnectionReconciler
}

func makeTenantsReconciler(r *PulsarConnectionReconciler) commonsreconciler.Interface {
	return &PulsarTenantReconciler{
		conn: r,
	}
}

// Observe checks the updates of object
func (r *PulsarTenantReconciler) Observe(ctx context.Context) error {
	tenantList := &resourcev1alpha1.PulsarTenantList{}
	if err := r.conn.client.List(ctx, tenantList, client.InNamespace(r.conn.connection.Namespace),
		client.MatchingFields(map[string]string{
			".spec.connectionRef.name": r.conn.connection.Name,
		})); err != nil {
		return fmt.Errorf("list tenants [%w]", err)
	}
	r.conn.tenants = tenantList.Items
	if !r.conn.hasUnreadyResource {
		for i := range r.conn.tenants {
			if !resourcev1alpha1.IsPulsarResourceReady(&r.conn.tenants[i]) {
				r.conn.hasUnreadyResource = true
				break
			}
		}
	}
	return nil
}

// Reconcile reconciles all tenants
func (r *PulsarTenantReconciler) Reconcile(ctx context.Context) error {
	for i := range r.conn.tenants {
		tenant := &r.conn.tenants[i]
		if err := r.ReconcileTenant(ctx, r.conn.pulsarAdmin, tenant); err != nil {
			return fmt.Errorf("reconcile tenant [%w]", err)
		}
	}
	return nil
}

// ReconcileTenant move the current state of the toic closer to the desired state
func (r *PulsarTenantReconciler) ReconcileTenant(ctx context.Context, pulsarAdmin admin.PulsarAdmin,
	tenant *resourcev1alpha1.PulsarTenant) error {
	if !tenant.DeletionTimestamp.IsZero() {
		if tenant.Spec.LifecyclePolicy == resourcev1alpha1.CleanUpAfterDeletion {
			if err := pulsarAdmin.DeleteTenant(tenant.Spec.Name); err != nil && admin.IsNotFound(err) {
				r.conn.log.Error(err, "Failed to delete tenant", "Namespace", tenant.Namespace, "Name", tenant.Name)
				return err
			}
		}

		// TODO use otelcontroller until kube-instrumentation upgrade controller-runtime version to newer
		controllerutil.RemoveFinalizer(tenant, resourcev1alpha1.FinalizerName)
		if err := r.conn.client.Update(ctx, tenant); err != nil {
			r.conn.log.Error(err, "Failed to remove finalizer", "Namespace", tenant.Namespace, "Name", tenant.Name)
			return err
		}

		return nil
	}

	// TODO use otelcontroller until kube-instrumentation upgrade controller-runtime version to newer
	controllerutil.AddFinalizer(tenant, resourcev1alpha1.FinalizerName)
	if err := r.conn.client.Update(ctx, tenant); err != nil {
		r.conn.log.Error(err, "Failed to add finalizer", "Namespace", tenant.Namespace, "Name", tenant.Name)
		return err
	}

	if resourcev1alpha1.IsPulsarResourceReady(tenant) {
		r.conn.log.V(1).Info("Resource is ready", "Namespace", tenant.Namespace, "Name", tenant.Name)
		return nil
	}

	tenantParams := &admin.TenantParams{
		AdminRoles:      tenant.Spec.AdminRoles,
		AllowedClusters: tenant.Spec.AllowedClusters,
	}
	if err := pulsarAdmin.ApplyTenant(tenant.Spec.Name, tenantParams); err != nil {
		meta.SetStatusCondition(&tenant.Status.Conditions, *NewErrorCondition(tenant.Generation, err.Error()))
		r.conn.log.Error(err, "Failed to apply tenant", "Namespace", tenant.Namespace, "Name", tenant.Name)
		if err := r.conn.client.Status().Update(ctx, tenant); err != nil {
			r.conn.log.Error(err, "Failed to update the tenant status", "Namespace", tenant.Namespace,
				"Name", tenant.Name)
			return nil
		}
		return err
	}

	tenant.Status.ObservedGeneration = tenant.Generation
	meta.SetStatusCondition(&tenant.Status.Conditions, *NewReadyCondition(tenant.Generation))
	if err := r.conn.client.Status().Update(ctx, tenant); err != nil {
		r.conn.log.Error(err, "Failed to update the tenant status", "Namespace", tenant.Namespace,
			"Name", tenant.Name)
		return err
	}

	return nil
}
