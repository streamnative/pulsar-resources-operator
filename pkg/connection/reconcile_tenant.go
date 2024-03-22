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
	"fmt"

	"github.com/go-logr/logr"
	"github.com/streamnative/pulsar-resources-operator/pkg/feature"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	"github.com/streamnative/pulsar-resources-operator/pkg/admin"
	"github.com/streamnative/pulsar-resources-operator/pkg/reconciler"
)

// PulsarTenantReconciler reconciles a PulsarTenant object
type PulsarTenantReconciler struct {
	conn *PulsarConnectionReconciler
	log  logr.Logger
}

func makeTenantsReconciler(r *PulsarConnectionReconciler) reconciler.Interface {
	return &PulsarTenantReconciler{
		conn: r,
		log:  makeSubResourceLog(r, "PulsarTenant"),
	}
}

// Observe checks the updates of object
func (r *PulsarTenantReconciler) Observe(ctx context.Context) error {
	r.log.V(1).Info("Start Observe")

	tenantList := &resourcev1alpha1.PulsarTenantList{}
	if err := r.conn.client.List(ctx, tenantList, client.InNamespace(r.conn.connection.Namespace),
		client.MatchingFields(map[string]string{
			".spec.connectionRef.name": r.conn.connection.Name,
		})); err != nil {
		return fmt.Errorf("list tenants [%w]", err)
	}
	r.log.V(1).Info("Observed tenants items", "Count", len(tenantList.Items))

	r.conn.tenants = tenantList.Items
	for i := range r.conn.tenants {
		if !resourcev1alpha1.IsPulsarResourceReady(&r.conn.tenants[i]) {
			r.conn.addUnreadyResource(&r.conn.tenants[i])
		}
	}

	r.log.V(1).Info("Observe Done")
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
	log := r.log.WithValues("name", tenant.Name, "namespace", tenant.Namespace)
	log.Info("Start Reconcile")

	if !tenant.DeletionTimestamp.IsZero() {
		log.Info("Deleting tenant", "LifecyclePolicy", tenant.Spec.LifecyclePolicy)
		if tenant.Spec.LifecyclePolicy != resourcev1alpha1.KeepAfterDeletion {
			if err := pulsarAdmin.DeleteTenant(tenant.Spec.Name); err != nil && !admin.IsNotFound(err) {
				log.Error(err, "Failed to delete tenant")
				return err
			}
		}

		// TODO use otelcontroller until kube-instrumentation upgrade controller-runtime version to newer
		controllerutil.RemoveFinalizer(tenant, resourcev1alpha1.FinalizerName)
		if err := r.conn.client.Update(ctx, tenant); err != nil {
			log.Error(err, "Failed to remove finalizer")
			return err
		}

		return nil
	}

	// TODO use otelcontroller until kube-instrumentation upgrade controller-runtime version to newer
	controllerutil.AddFinalizer(tenant, resourcev1alpha1.FinalizerName)
	if err := r.conn.client.Update(ctx, tenant); err != nil {
		log.Error(err, "Failed to add finalizer")
		return err
	}

	if resourcev1alpha1.IsPulsarResourceReady(tenant) &&
		!feature.DefaultFeatureGate.Enabled(feature.AlwaysUpdatePulsarResource) {
		log.Info("Skip reconcile, tenant resource is ready")
		return nil
	}

	tenantParams := &admin.TenantParams{
		AdminRoles:      tenant.Spec.AdminRoles,
		AllowedClusters: tenant.Spec.AllowedClusters,
	}
	// If ObeservedGeneration is greater than 1, means the tenant alreday been created
	// So it will been updated in ApplyTenant
	if tenant.Status.ObservedGeneration > 1 {
		tenantParams.Changed = true
	}

	if refs := tenant.Spec.GeoReplicationRefs; len(refs) != 0 {
		for _, ref := range refs {
			geoReplication := &resourcev1alpha1.PulsarGeoReplication{}
			namespacedName := types.NamespacedName{
				Namespace: tenant.Namespace,
				Name:      ref.Name,
			}
			if err := r.conn.client.Get(ctx, namespacedName, geoReplication); err != nil {
				return err
			}
			log.V(1).Info("Found geo replication", "GEO Replication", geoReplication.Name)

			destConnection := &resourcev1alpha1.PulsarConnection{}
			namespacedName = types.NamespacedName{
				Name:      geoReplication.Spec.DestinationConnectionRef.Name,
				Namespace: geoReplication.Namespace,
			}
			if err := r.conn.client.Get(ctx, namespacedName, destConnection); err != nil {
				log.Error(err, "Failed to get destination connection for geo replication")
				return err
			}
			tenantParams.AllowedClusters = append(tenantParams.AllowedClusters, destConnection.Spec.ClusterName)
			tenantParams.AllowedClusters = append(tenantParams.AllowedClusters, r.conn.connection.Spec.ClusterName)
		}
		log.Info("Geo Replication is enabled. Apply tenant with allowed clusters", "clusters", tenantParams.AllowedClusters)
	}

	if err := pulsarAdmin.ApplyTenant(tenant.Spec.Name, tenantParams); err != nil {
		meta.SetStatusCondition(&tenant.Status.Conditions, *NewErrorCondition(tenant.Generation, err.Error()))
		log.Error(err, "Failed to apply tenant")
		if err := r.conn.client.Status().Update(ctx, tenant); err != nil {
			log.Error(err, "Failed to update the tenant status")
			return err
		}
		return err
	}

	tenant.Status.ObservedGeneration = tenant.Generation
	meta.SetStatusCondition(&tenant.Status.Conditions, *NewReadyCondition(tenant.Generation))
	if err := r.conn.client.Status().Update(ctx, tenant); err != nil {
		log.Error(err, "Failed to update the tenant status")
		return err
	}

	return nil
}
