// Copyright 2023 StreamNative
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
	"github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	"github.com/streamnative/pulsar-resources-operator/pkg/admin"
	"github.com/streamnative/pulsar-resources-operator/pkg/reconciler"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// PulsarGeoReplicationReconciler reconciles a PulsarGeoReplication object
type PulsarGeoReplicationReconciler struct {
	conn *PulsarConnectionReconciler
	log  logr.Logger
}

func makeGeoReplicationReconciler(r *PulsarConnectionReconciler) reconciler.Interface {
	return &PulsarGeoReplicationReconciler{
		conn: r,
		log:  r.log.WithName("PulsarTenant"),
	}
}

// Observe checks the updates of object
func (r *PulsarGeoReplicationReconciler) Observe(ctx context.Context) error {
	geoList := &resourcev1alpha1.PulsarGeoReplicationList{}
	if err := r.conn.client.List(ctx, geoList, client.InNamespace(r.conn.connection.Namespace),
		client.MatchingFields(map[string]string{
			".spec.connectionRef.name": r.conn.connection.Name,
		})); err != nil {
		return fmt.Errorf("list GeoReplication [%w]", err)
	}
	r.log.V(1).Info("Observed geo replication items", "Count", len(geoList.Items))

	r.conn.geoReplications = geoList.Items
	if !r.conn.hasUnreadyResource {
		for i := range r.conn.geoReplications {
			if !resourcev1alpha1.IsPulsarResourceReady(&r.conn.geoReplications[i]) {
				r.conn.hasUnreadyResource = true
				break
			}
		}
	}

	r.log.V(1).Info("Observe Done")
	return nil
}

// Observe checks the updates of object
func (r *PulsarGeoReplicationReconciler) Reconcile(ctx context.Context) error {
	for i := range r.conn.geoReplications {
		geoReplication := &r.conn.geoReplications[i]
		if err := r.ReconcileGeoReplication(ctx, r.conn.pulsarAdmin, geoReplication); err != nil {
			return fmt.Errorf("reconcile tenant [%w]", err)
		}
	}
	return nil
}

func (r *PulsarGeoReplicationReconciler) ReconcileGeoReplication(ctx context.Context, pulsarAdmin admin.PulsarAdmin,
	geoReplication *resourcev1alpha1.PulsarGeoReplication) error {

	log := r.log.WithValues("pulsargeoreplication", geoReplication.Name, "namespace", geoReplication.Namespace)
	log.V(1).Info("Start Reconcile")

	if !geoReplication.DeletionTimestamp.IsZero() {
		log.Info("Deleting GeoReplication", "LifecyclePolicy", geoReplication.Spec.LifecyclePolicy)
		if geoReplication.Spec.LifecyclePolicy == resourcev1alpha1.CleanUpAfterDeletion {
			for _, cluster := range geoReplication.Spec.Clusters {
				if err := pulsarAdmin.DeleteCluster(cluster.Name); err != nil && admin.IsNotFound(err) {
					log.Error(err, "Failed to delete geo replication")
					return err
				}
			}
		}

		controllerutil.RemoveFinalizer(geoReplication, resourcev1alpha1.FinalizerName)
		if err := r.conn.client.Update(ctx, geoReplication); err != nil {
			log.Error(err, "Failed to remove finalizer")
			return err
		}
	}
	controllerutil.AddFinalizer(geoReplication, resourcev1alpha1.FinalizerName)
	if err := r.conn.client.Update(ctx, geoReplication); err != nil {
		log.Error(err, "Failed to remove finalizer")
		return err
	}

	if resourcev1alpha1.IsPulsarResourceReady(geoReplication) {
		log.V(1).Info("Resource is ready")
		return nil
	}
	for _, cluster := range geoReplication.Spec.Clusters {

		destConnection := &resourcev1alpha1.PulsarConnection{}
		namespacedName := types.NamespacedName{
			Name:      cluster.DestinationConnectionRef.Name,
			Namespace: geoReplication.Namespace,
		}
		if err := r.conn.client.Get(ctx, namespacedName, destConnection); err != nil {
			log.Error(err, "Failed to get destination connection for geo replication")
			if apierrors.IsNotFound(err) {
				return err
			}
		}

		clusterParam := &admin.ClusterParams{
			ServiceURL:             destConnection.Spec.AdminServiceURL,
			BrokerServiceURL:       destConnection.Spec.BrokerServiceURL,
			ServiceSecureURL:       destConnection.Spec.AdminServiceSecureURL,
			BrokerServiceSecureURL: destConnection.Spec.BrokerServiceSecureURL,
		}

		hasAuth := true
		if auth := destConnection.Spec.Authentication; auth != nil {
			if auth.Token != nil {
				value, err := GetValue(ctx, r.conn.client, destConnection.Namespace, auth.Token)
				if err != nil {
					return err
				}
				if value != nil {
					clusterParam.AuthPlugin = v1alpha1.AuthPluginToken
					clusterParam.AuthParameters = "token:" + *value
					hasAuth = true
				}
			}
			if auth.OAuth2 != nil && !hasAuth {
				// TODO
			}

		}

		// If the cluster already exists, skip it
		if pulsarAdmin.CheckClusterExist(cluster.Name) {
			if err := pulsarAdmin.UpdateCluster(cluster.Name, clusterParam); err != nil {
				meta.SetStatusCondition(&geoReplication.Status.Conditions, *NewErrorCondition(geoReplication.Generation, err.Error()))
				log.Error(err, "Failed to create geo replication cluster")
				if err := r.conn.client.Status().Update(ctx, geoReplication); err != nil {
					log.Error(err, "Failed to update the geo replication status")
					return err
				}
				return err
			}
			continue
		}

		// Create Clusters
		if err := pulsarAdmin.CreateCluster(cluster.Name, clusterParam); err != nil {
			meta.SetStatusCondition(&geoReplication.Status.Conditions, *NewErrorCondition(geoReplication.Generation, err.Error()))
			log.Error(err, "Failed to create geo replication cluster")
			if err := r.conn.client.Status().Update(ctx, geoReplication); err != nil {
				log.Error(err, "Failed to update the geo replication status")
				return err
			}
			return err
		}
	}

	return nil
}
