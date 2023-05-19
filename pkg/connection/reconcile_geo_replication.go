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
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	"github.com/streamnative/pulsar-resources-operator/pkg/admin"
	"github.com/streamnative/pulsar-resources-operator/pkg/reconciler"
)

// PulsarGeoReplicationReconciler reconciles a PulsarGeoReplication object
type PulsarGeoReplicationReconciler struct {
	conn *PulsarConnectionReconciler
	log  logr.Logger
}

func makeGeoReplicationReconciler(r *PulsarConnectionReconciler) reconciler.Interface {
	return &PulsarGeoReplicationReconciler{
		conn: r,
		log:  r.log.WithName("PulsarGeoReplication"),
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

// Reconcile reconciles all the geo replication objects
func (r *PulsarGeoReplicationReconciler) Reconcile(ctx context.Context) error {
	for i := range r.conn.geoReplications {
		geoReplication := &r.conn.geoReplications[i]
		if err := r.ReconcileGeoReplication(ctx, r.conn.pulsarAdmin, geoReplication); err != nil {
			return fmt.Errorf("reconcile geo replication [%w]", err)
		}
	}
	return nil
}

// ReconcileGeoReplication handle the current state of the geo replication to the desired state
func (r *PulsarGeoReplicationReconciler) ReconcileGeoReplication(ctx context.Context, pulsarAdmin admin.PulsarAdmin,
	geoReplication *resourcev1alpha1.PulsarGeoReplication) error {
	log := r.log.WithValues("pulsargeoreplication", geoReplication.Name, "namespace", geoReplication.Namespace)
	log.V(1).Info("Start Reconcile")

	destConnection := &resourcev1alpha1.PulsarConnection{}
	namespacedName := types.NamespacedName{
		Name:      geoReplication.Spec.DestinationConnectionRef.Name,
		Namespace: geoReplication.Namespace,
	}
	if err := r.conn.client.Get(ctx, namespacedName, destConnection); err != nil {
		log.Error(err, "Failed to get destination connection for geo replication")
		return err
	}
	// TODO Currently, if the destination pulsarconnection is updated, the this reconcile won't notice it
	// Need to fix it in the future work.

	destClusterName := destConnection.Spec.ClusterName
	if destClusterName == "" {
		err := fmt.Errorf("ClusterName is empty in destination connection")
		meta.SetStatusCondition(&geoReplication.Status.Conditions, *NewErrorCondition(geoReplication.Generation, err.Error()))
		log.Error(err, "Failed to validate geo replication cluster")
		if err := r.conn.client.Status().Update(ctx, geoReplication); err != nil {
			log.Error(err, "Failed to update the geo replication status")
			return err
		}
		return err
	}

	if !geoReplication.DeletionTimestamp.IsZero() {
		log.Info("Deleting GeoReplication", "LifecyclePolicy", geoReplication.Spec.LifecyclePolicy)
		skip := false
		for _, i := range r.conn.tenants {
			if len(i.Spec.GeoReplicationRefs) != 0 {
				for _, j := range i.Spec.GeoReplicationRefs {
					if j.Name == geoReplication.Name {
						skip = true
						log.Info("There is still a tenant used this geo replication. The GeoReplication couldn't be deleted", "tenant", i.Name)
					}
				}
			}
		}
		if !skip {
			if geoReplication.Spec.LifecyclePolicy == resourcev1alpha1.CleanUpAfterDeletion {
				// Delete the cluster that created with destination cluster info.
				// TODO it can only be deleted after the cluster has been removed from the tenant, namespace, and topic
				if err := pulsarAdmin.DeleteCluster(destClusterName); err != nil && !admin.IsNotFound(err) {
					log.Error(err, "Failed to delete geo replication cluster")
					return err
				}
			}
			controllerutil.RemoveFinalizer(geoReplication, resourcev1alpha1.FinalizerName)
			if err := r.conn.client.Update(ctx, geoReplication); err != nil {
				log.Error(err, "Failed to remove finalizer")
				return err
			}
		}
	}
	controllerutil.AddFinalizer(geoReplication, resourcev1alpha1.FinalizerName)
	if err := r.conn.client.Update(ctx, geoReplication); err != nil {
		log.Error(err, "Failed to add finalizer")
		return err
	}

	if resourcev1alpha1.IsPulsarResourceReady(geoReplication) {
		// After the previous reconcile succeed, the cluster will be created successfully,
		// it will update the condition Ready to true, and update the observedGeneration to metadata.generation
		// If there is no new changes in the object, there is no need to run the left code again.
		log.V(1).Info("Resource is ready")
		return nil
	}

	clusterParam := &admin.ClusterParams{
		ServiceURL:                     destConnection.Spec.AdminServiceURL,
		BrokerServiceURL:               destConnection.Spec.BrokerServiceURL,
		ServiceSecureURL:               destConnection.Spec.AdminServiceSecureURL,
		BrokerServiceSecureURL:         destConnection.Spec.BrokerServiceSecureURL,
		BrokerClientTrustCertsFilePath: destConnection.Spec.BrokerClientTrustCertsFilePath,
	}

	hasAuth := true
	if auth := destConnection.Spec.Authentication; auth != nil {
		if auth.Token != nil {
			value, err := GetValue(ctx, r.conn.client, destConnection.Namespace, auth.Token)
			if err != nil {
				return err
			}
			if value != nil {
				clusterParam.AuthPlugin = resourcev1alpha1.AuthPluginToken
				clusterParam.AuthParameters = "token:" + *value
				hasAuth = true
			}
		}
		if auth.OAuth2 != nil && !hasAuth {
			// TODO support oauth2
			log.Info("Oauth2 will support later")
		}
	}

	// If the cluster already exists, only update it
	exist, err := pulsarAdmin.CheckClusterExist(destClusterName)
	if err != nil {
		return err
	}
	if exist {
		log.V(1).Info("Update cluster", "ClusterName", destClusterName, "params", clusterParam)
		if err := pulsarAdmin.UpdateCluster(destClusterName, clusterParam); err != nil {
			meta.SetStatusCondition(&geoReplication.Status.Conditions, *NewErrorCondition(geoReplication.Generation, err.Error()))
			log.Error(err, "Failed to update existing geo replication cluster")
			if err := r.conn.client.Status().Update(ctx, geoReplication); err != nil {
				log.Error(err, "Failed to update the geo replication status")
				return err
			}
			return err
		}
	} else {
		// Create Clusters
		log.V(1).Info("Create cluster", "ClusterName", destClusterName, "params", clusterParam)
		if err := pulsarAdmin.CreateCluster(destClusterName, clusterParam); err != nil {
			meta.SetStatusCondition(&geoReplication.Status.Conditions, *NewErrorCondition(geoReplication.Generation, err.Error()))
			log.Error(err, "Failed to create geo replication cluster")
			if err := r.conn.client.Status().Update(ctx, geoReplication); err != nil {
				log.Error(err, "Failed to update the geo replication status")
				return err
			}
			return err
		}
	}

	geoReplication.Status.ObservedGeneration = geoReplication.Generation
	meta.SetStatusCondition(&geoReplication.Status.Conditions, *NewReadyCondition(geoReplication.Generation))
	if err := r.conn.client.Status().Update(ctx, geoReplication); err != nil {
		log.Error(err, "Failed to update the geoReplication status")
		return err
	}

	return nil
}
