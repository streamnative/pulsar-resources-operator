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
	"github.com/streamnative/pulsar-resources-operator/pkg/feature"
	"github.com/streamnative/pulsar-resources-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
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
		log:  makeSubResourceLog(r, "PulsarGeoReplication"),
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

	if len(geoList.Items) == 0 {
		if err := r.conn.client.List(ctx, geoList, client.InNamespace(r.conn.connection.Namespace),
			client.MatchingFields(map[string]string{
				".spec.destinationConnectionRef.name": r.conn.connection.Name,
			})); err != nil {
			return fmt.Errorf("list GeoReplication [%w]", err)
		}
	}
	r.log.V(1).Info("Observed geo replication items", "Count", len(geoList.Items))

	r.conn.geoReplications = geoList.Items
	// Force the `hasUnreadyResource` to be `true`` to trigger the PulsarConnection reload the auth config
	for i := range geoList.Items {
		r.conn.addUnreadyResource(&geoList.Items[i])
	}

	r.log.V(1).Info("Observe Done")
	return nil
}

// Reconcile reconciles all the geo replication objects
func (r *PulsarGeoReplicationReconciler) Reconcile(ctx context.Context) error {
	for i := range r.conn.geoReplications {
		r.log.V(1).Info("Reconcile Geo")
		geoReplication := &r.conn.geoReplications[i]
		pulsarAdmin := r.conn.pulsarAdmin
		if geoReplication.Spec.ConnectionRef.Name != r.conn.connection.Name {
			// If the connectionRef is the remote connection, we need to create a new pulsarAdmin for it
			localConnection := &resourcev1alpha1.PulsarConnection{}
			namespacedName := types.NamespacedName{
				Name:      geoReplication.Spec.ConnectionRef.Name,
				Namespace: geoReplication.Namespace,
			}
			if err := r.conn.client.Get(ctx, namespacedName, localConnection); err != nil {
				return fmt.Errorf("get local pulsarConnection [%w]", err)
			}
			cfg, err := MakePulsarAdminConfig(ctx, localConnection, r.conn.client)
			if err != nil {
				return fmt.Errorf("make pulsar admin config [%w]", err)
			}
			pulsarAdmin, err = r.conn.creator(*cfg)
			if err != nil {
				return fmt.Errorf("make pulsar admin [%w]", err)
			}
		}

		if err := r.ReconcileGeoReplication(ctx, pulsarAdmin, geoReplication); err != nil {
			return fmt.Errorf("reconcile geo replication [%w]", err)
		}
	}
	return nil
}

// ReconcileGeoReplication handle the current state of the geo replication to the desired state
func (r *PulsarGeoReplicationReconciler) ReconcileGeoReplication(ctx context.Context, pulsarAdmin admin.PulsarAdmin,
	geoReplication *resourcev1alpha1.PulsarGeoReplication) error {
	log := r.log.WithValues("name", geoReplication.Name, "namespace", geoReplication.Namespace)
	log.Info("Start Reconcile")

	destConnection := &resourcev1alpha1.PulsarConnection{}
	namespacedName := types.NamespacedName{
		Name:      geoReplication.Spec.DestinationConnectionRef.Name,
		Namespace: geoReplication.Namespace,
	}
	if err := r.conn.client.Get(ctx, namespacedName, destConnection); err != nil {
		log.Error(err, "Failed to get destination connection for geo replication")
		return err
	}

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
			if geoReplication.Spec.LifecyclePolicy != resourcev1alpha1.KeepAfterDeletion {
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

	secretUpdated, err := r.checkSecretRefUpdate(*destConnection)
	if err != nil {
		return err
	}

	// if destConnection update, let's update the cluster info
	if !secretUpdated && destConnection.Generation == destConnection.Status.ObservedGeneration &&
		resourcev1alpha1.IsPulsarResourceReady(geoReplication) &&
		!feature.DefaultFeatureGate.Enabled(feature.AlwaysUpdatePulsarResource) {
		// After the previous reconcile succeed, the cluster will be created successfully,
		// it will update the condition Ready to true, and update the observedGeneration to metadata.generation
		// If there is no new changes in the object, there is no need to run the left code again.
		log.Info("Skip reconcile, geo resource is ready",
			"Name", geoReplication.Name, "Namespace", geoReplication.Namespace)
		return nil
	}

	clusterParam, err2 := createParams(ctx, destConnection, r.conn.client)
	if err2 != nil {
		return err2
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
		log.Info("Update cluster success", "ClusterName", destClusterName)
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
		log.Info("Create cluster success", "ClusterName", destClusterName)
	}

	destConnection.Status.ObservedGeneration = destConnection.Generation
	meta.SetStatusCondition(&destConnection.Status.Conditions, *NewReadyCondition(destConnection.Generation))
	if err := r.conn.client.Status().Update(ctx, destConnection); err != nil {
		log.Error(err, "Failed to update the destination connection status")
		return err
	}

	geoReplication.Status.ObservedGeneration = geoReplication.Generation
	meta.SetStatusCondition(&geoReplication.Status.Conditions, *NewReadyCondition(geoReplication.Generation))
	if err := r.conn.client.Status().Update(ctx, geoReplication); err != nil {
		log.Error(err, "Failed to update the geoReplication status")
		return err
	}

	return nil
}

func (r *PulsarGeoReplicationReconciler) checkSecretRefUpdate(connection resourcev1alpha1.PulsarConnection) (bool, error) {
	auth := connection.Spec.Authentication
	if auth == nil || auth.Token.SecretRef == nil {
		return false, nil
	}
	secret := &corev1.Secret{}
	namespacedName := types.NamespacedName{
		Name:      auth.Token.SecretRef.Name,
		Namespace: connection.Namespace,
	}
	if err := r.conn.client.Get(context.Background(), namespacedName, secret); err != nil {
		return false, err
	}
	secretHash, err := utils.CalculateSecretKeyMd5(secret, auth.Token.SecretRef.Key)
	if err != nil {
		return false, err
	}
	return connection.Status.SecretKeyHash != secretHash, nil
}

func createParams(ctx context.Context, destConnection *resourcev1alpha1.PulsarConnection, client client.Client) (*admin.ClusterParams, error) {
	clusterParam := &admin.ClusterParams{
		ServiceURL:                     destConnection.Spec.AdminServiceURL,
		BrokerServiceURL:               destConnection.Spec.BrokerServiceURL,
		ServiceSecureURL:               destConnection.Spec.AdminServiceSecureURL,
		BrokerServiceSecureURL:         destConnection.Spec.BrokerServiceSecureURL,
		BrokerClientTrustCertsFilePath: destConnection.Spec.BrokerClientTrustCertsFilePath,
	}

	if auth := destConnection.Spec.Authentication; auth != nil {
		if auth.Token != nil {
			value, err := GetValue(ctx, client, destConnection.Namespace, auth.Token)
			if err != nil {
				return nil, err
			}
			if value != nil {
				clusterParam.AuthPlugin = resourcev1alpha1.AuthPluginToken
				clusterParam.AuthParameters = "token:" + *value
			}
		}
		// TODO support oauth2
		// if auth.OAuth2 != nil && !hasAuth {
		// }
	}
	return clusterParam, nil
}
