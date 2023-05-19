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
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	"github.com/streamnative/pulsar-resources-operator/pkg/admin"
	"github.com/streamnative/pulsar-resources-operator/pkg/reconciler"
)

// PulsarNamespaceReconciler reconciles a PulsarNamespace object
type PulsarNamespaceReconciler struct {
	conn *PulsarConnectionReconciler
	log  logr.Logger
}

func makeNamespacesReconciler(r *PulsarConnectionReconciler) reconciler.Interface {
	return &PulsarNamespaceReconciler{
		conn: r,
		log:  r.log.WithName("PulsarNamespace"),
	}
}

// Observe checks the updates of object
func (r *PulsarNamespaceReconciler) Observe(ctx context.Context) error {
	r.log.V(1).Info("Start Observe")

	namespaceList := &resourcev1alpha1.PulsarNamespaceList{}
	if err := r.conn.client.List(ctx, namespaceList, client.InNamespace(r.conn.connection.Namespace),
		client.MatchingFields(map[string]string{
			".spec.connectionRef.name": r.conn.connection.Name,
		})); err != nil {
		return fmt.Errorf("list namespaces [%w]", err)
	}
	r.log.V(1).Info("Observed namespace items", "Count", len(namespaceList.Items))

	r.conn.namespaces = namespaceList.Items
	if !r.conn.hasUnreadyResource {
		for i := range r.conn.namespaces {
			if !resourcev1alpha1.IsPulsarResourceReady(&r.conn.namespaces[i]) {
				r.conn.hasUnreadyResource = true
				break
			}
		}
	}

	r.log.V(1).Info("Observe Done")
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
	log := r.log.WithValues("pulsarnamespace", namespace.Name, "namespace", namespace.Namespace)
	log.V(1).Info("Start Reconcile")

	if !namespace.DeletionTimestamp.IsZero() {
		log.Info("Deleting namespace", "LifecyclePolicy", namespace.Spec.LifecyclePolicy)

		// When geo replication is enabled, it should reset the replication clusters to
		// default local cluster for the namespace first.
		if namespace.Status.GeoReplicationEnabled {
			log.Info("GeoReplication is enabled. Reset namespace cluster", "LifecyclePolicy", namespace.Spec.LifecyclePolicy, "ClusterName", r.conn.connection.Spec.ClusterName)
			if err := pulsarAdmin.SetNamespaceClusters(namespace.Spec.Name, []string{r.conn.connection.Spec.ClusterName}); err != nil {
				log.Error(err, "Failed to reset the cluster for namespace")
				return err
			}
		}

		if namespace.Spec.LifecyclePolicy == resourcev1alpha1.CleanUpAfterDeletion {
			if err := pulsarAdmin.DeleteNamespace(namespace.Spec.Name); err != nil && !admin.IsNotFound(err) {
				log.Error(err, "Failed to delete namespace")
				meta.SetStatusCondition(&namespace.Status.Conditions, *NewErrorCondition(namespace.Generation, err.Error()))
				if err := r.conn.client.Status().Update(ctx, namespace); err != nil {
					log.Error(err, "Failed to update the geo replication status")
					return err
				}
				return err
			}
		}

		// TODO use otelcontroller until kube-instrumentation upgrade controller-runtime version to newer
		controllerutil.RemoveFinalizer(namespace, resourcev1alpha1.FinalizerName)
		if err := r.conn.client.Update(ctx, namespace); err != nil {
			log.Error(err, "Failed to remove finalizer")
			return err
		}

		return nil
	}

	if namespace.Spec.LifecyclePolicy == resourcev1alpha1.CleanUpAfterDeletion {
		// TODO use otelcontroller until kube-instrumentation upgrade controller-runtime version to newer
		controllerutil.AddFinalizer(namespace, resourcev1alpha1.FinalizerName)
		if err := r.conn.client.Update(ctx, namespace); err != nil {
			log.Error(err, "Failed to add finalizer")
			return err
		}
	}

	if resourcev1alpha1.IsPulsarResourceReady(namespace) {
		log.V(1).Info("Resource is ready")
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
		BacklogQuotaType:            namespace.Spec.BacklogQuotaType,
	}

	if refs := namespace.Spec.GeoReplicationRefs; len(refs) != 0 {
		for _, ref := range refs {
			geoReplication := &resourcev1alpha1.PulsarGeoReplication{}
			namespacedName := types.NamespacedName{
				Namespace: namespace.Namespace,
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
			params.ReplicationClusters = append(params.ReplicationClusters, destConnection.Spec.ClusterName)
			params.ReplicationClusters = append(params.ReplicationClusters, r.conn.connection.Spec.ClusterName)
		}

		log.Info("apply namespace with extra replication clusters", "clusters", params.ReplicationClusters)
		namespace.Status.GeoReplicationEnabled = true
	} else if namespace.Status.GeoReplicationEnabled {
		// when GeoReplicationRefs is removed, it should reset the namespace clusters
		// to the default local cluster
		params.ReplicationClusters = []string{r.conn.connection.Spec.ClusterName}
		log.Info("Geo Replication disabled. Reset namespace with local cluster", "cluster", params.ReplicationClusters)
		namespace.Status.GeoReplicationEnabled = false
	}

	if err := pulsarAdmin.ApplyNamespace(namespace.Spec.Name, params); err != nil {
		meta.SetStatusCondition(&namespace.Status.Conditions, *NewErrorCondition(namespace.Generation, err.Error()))
		log.Error(err, "Failed to apply namespace")
		if err := r.conn.client.Status().Update(ctx, namespace); err != nil {
			log.Error(err, "Failed to update the namespace status")
			return nil
		}
		return err
	}

	namespace.Status.ObservedGeneration = namespace.Generation
	meta.SetStatusCondition(&namespace.Status.Conditions, *NewReadyCondition(namespace.Generation))
	if err := r.conn.client.Status().Update(ctx, namespace); err != nil {
		log.Error(err, "Failed to update the namespace status")
		return err
	}

	return nil
}
