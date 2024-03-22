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
	"reflect"

	"github.com/go-logr/logr"
	"github.com/streamnative/pulsar-resources-operator/pkg/feature"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	"github.com/streamnative/pulsar-resources-operator/pkg/admin"
	"github.com/streamnative/pulsar-resources-operator/pkg/reconciler"
)

// PulsarTopicReconciler reconciles a PulsarTopic object
type PulsarTopicReconciler struct {
	conn *PulsarConnectionReconciler
	log  logr.Logger
}

func makeTopicsReconciler(r *PulsarConnectionReconciler) reconciler.Interface {
	return &PulsarTopicReconciler{
		conn: r,
		log:  makeSubResourceLog(r, "PulsarTopic"),
	}
}

// Observe checks the updates of object
func (r *PulsarTopicReconciler) Observe(ctx context.Context) error {
	r.log.V(1).Info("Start Observe")

	topicsList := &resourcev1alpha1.PulsarTopicList{}
	if err := r.conn.client.List(ctx, topicsList, client.InNamespace(r.conn.connection.Namespace),
		client.MatchingFields(map[string]string{
			".spec.connectionRef.name": r.conn.connection.Name,
		})); err != nil {
		return fmt.Errorf("list topics [%w]", err)
	}
	r.log.V(1).Info("Observed topic items", "Count", len(topicsList.Items))

	r.conn.topics = topicsList.Items
	for i := range r.conn.topics {
		if !resourcev1alpha1.IsPulsarResourceReady(&r.conn.topics[i]) {
			r.conn.addUnreadyResource(&r.conn.topics[i])
		}
	}

	r.log.V(1).Info("Observe Done")
	return nil
}

// Reconcile reconciles all topics
func (r *PulsarTopicReconciler) Reconcile(ctx context.Context) error {
	for i := range r.conn.topics {
		topic := &r.conn.topics[i]
		if err := r.ReconcileTopic(ctx, r.conn.pulsarAdmin, topic); err != nil {
			return fmt.Errorf("reconcile topic [%w]", err)
		}
	}
	return nil
}

// ReconcileTopic move the current state of the toic closer to the desired state
func (r *PulsarTopicReconciler) ReconcileTopic(ctx context.Context, pulsarAdmin admin.PulsarAdmin,
	topic *resourcev1alpha1.PulsarTopic) error {
	log := r.log.WithValues("name", topic.Name, "namespace", topic.Namespace)
	log.Info("Start Reconcile")

	if !topic.DeletionTimestamp.IsZero() {
		log.Info("Deleting topic", "LifecyclePolicy", topic.Spec.LifecyclePolicy)

		if topic.Spec.LifecyclePolicy != resourcev1alpha1.KeepAfterDeletion {
			// TODO when geoReplicationRef is not nil, it should reset the replication clusters to
			// default local cluster for the topic
			if topic.Status.GeoReplicationEnabled {
				log.Info("GeoReplication is enabled. Reset topic cluster first", "LifecyclePolicy", topic.Spec.LifecyclePolicy, "ClusterName", r.conn.connection.Spec.ClusterName)
				if err := pulsarAdmin.SetTopicClusters(topic.Spec.Name, topic.Spec.Persistent, []string{r.conn.connection.Spec.ClusterName}); err != nil {
					log.Error(err, "Failed to reset the cluster for topic")
					return err
				}
			}

			// Delete the schema of the topic before the deletion
			if topic.Spec.SchemaInfo != nil {
				log.Info("Deleting topic schema")
				err := pulsarAdmin.DeleteSchema(topic.Spec.Name)
				if err != nil {
					return err
				}
			}

			if err := pulsarAdmin.DeleteTopic(topic.Spec.Name); err != nil && !admin.IsNotFound(err) {
				log.Error(err, "Failed to delete topic")
				return err
			}
		}

		// TODO use otelcontroller until kube-instrumentation upgrade controller-runtime version to newer
		controllerutil.RemoveFinalizer(topic, resourcev1alpha1.FinalizerName)
		if err := r.conn.client.Update(ctx, topic); err != nil {
			log.Error(err, "Failed to remove finalizer")
			return err
		}
		return nil
	}

	if topic.Spec.LifecyclePolicy != resourcev1alpha1.KeepAfterDeletion {
		// TODO use otelcontroller until kube-instrumentation upgrade controller-runtime version to newer
		controllerutil.AddFinalizer(topic, resourcev1alpha1.FinalizerName)
		if err := r.conn.client.Update(ctx, topic); err != nil {
			log.Error(err, "Failed to add finalizer")
			return err
		}
	}

	if resourcev1alpha1.IsPulsarResourceReady(topic) &&
		!feature.DefaultFeatureGate.Enabled(feature.AlwaysUpdatePulsarResource) {
		log.Info("Skip reconcile, topic resource is ready")
		return nil
	}

	params := createTopicParams(topic)

	r.applyDefault(params)

	if refs := topic.Spec.GeoReplicationRefs; len(refs) != 0 {
		for _, ref := range refs {
			if err := r.applyGeo(ctx, params, ref, topic); err != nil {
				log.Error(err, "Failed to get destination connection for geo replication")
				return err
			}
		}

		log.Info("apply topic with replication clusters", "clusters", params.ReplicationClusters)
		topic.Status.GeoReplicationEnabled = true
	} else if topic.Status.GeoReplicationEnabled {
		// when GeoReplicationRefs is removed, it should reset the topic clusters
		params.ReplicationClusters = []string{r.conn.connection.Spec.ClusterName}
		log.Info("Geo Replication disabled. Reset topic with local cluster", "cluster", params.ReplicationClusters)
		topic.Status.GeoReplicationEnabled = false
	}

	if err := pulsarAdmin.ApplyTopic(topic.Spec.Name, params); err != nil {
		meta.SetStatusCondition(&topic.Status.Conditions, *NewErrorCondition(topic.Generation, err.Error()))
		log.Error(err, "Failed to apply topic")
		if err := r.conn.client.Status().Update(ctx, topic); err != nil {
			log.Error(err, "Failed to update the topic status")
			return nil
		}
		return err
	}

	schema, serr := pulsarAdmin.GetSchema(topic.Spec.Name)
	if serr != nil && !admin.IsNotFound(serr) {
		return serr
	}
	if topic.Spec.SchemaInfo != nil {
		// Only upload the schema when schema doesn't exist or the schema has been updated
		if admin.IsNotFound(serr) || !reflect.DeepEqual(topic.Spec.SchemaInfo, schema) {
			info := topic.Spec.SchemaInfo
			param := &admin.SchemaParams{
				Type:       info.Type,
				Schema:     info.Schema,
				Properties: info.Properties,
			}
			log.Info("Upload schema for the topic", "name", topic.Spec.Name, "type", info.Type, "schema", info.Schema, "properties", info.Properties)
			if err := pulsarAdmin.UploadSchema(topic.Spec.Name, param); err != nil {
				log.Error(err, "Failed to upload schema")
				if err := r.conn.client.Status().Update(ctx, topic); err != nil {
					log.Error(err, "Failed to upload schema for the topic")
					return nil
				}
				return err
			}
		}
	} else if schema != nil {
		// Delete the schema when the schema exists and schema info is empty
		log.Info("Deleting topic schema", "name", topic.Spec.Name)
		err := pulsarAdmin.DeleteSchema(topic.Spec.Name)
		if err != nil {
			return err
		}
	}

	topic.Status.ObservedGeneration = topic.Generation
	meta.SetStatusCondition(&topic.Status.Conditions, *NewReadyCondition(topic.Generation))
	if err := r.conn.client.Status().Update(ctx, topic); err != nil {
		log.Error(err, "Failed to update the topic status")
		return err
	}
	return nil
}

func createTopicParams(topic *resourcev1alpha1.PulsarTopic) *admin.TopicParams {
	return &admin.TopicParams{
		Persistent:                        topic.Spec.Persistent,
		Partitions:                        topic.Spec.Partitions,
		MaxProducers:                      topic.Spec.MaxProducers,
		MaxConsumers:                      topic.Spec.MaxConsumers,
		MessageTTL:                        topic.Spec.MessageTTL,
		MaxUnAckedMessagesPerConsumer:     topic.Spec.MaxUnAckedMessagesPerConsumer,
		MaxUnAckedMessagesPerSubscription: topic.Spec.MaxUnAckedMessagesPerSubscription,
		RetentionTime:                     topic.Spec.RetentionTime,
		RetentionSize:                     topic.Spec.RetentionSize,
		BacklogQuotaLimitTime:             topic.Spec.BacklogQuotaLimitTime,
		BacklogQuotaLimitSize:             topic.Spec.BacklogQuotaLimitSize,
		BacklogQuotaRetentionPolicy:       topic.Spec.BacklogQuotaRetentionPolicy,
	}
}

func (r *PulsarTopicReconciler) applyDefault(params *admin.TopicParams) {
	if params.Persistent == nil {
		// by default create persistent topic
		params.Persistent = pointer.BoolPtr(true)
	}

	if params.Partitions == nil {
		// by default create non-partitioned topic
		params.Partitions = pointer.Int32Ptr(0)
	}
}

func (r *PulsarTopicReconciler) applyGeo(ctx context.Context, params *admin.TopicParams,
	ref *corev1.LocalObjectReference, topic *resourcev1alpha1.PulsarTopic) error {
	geoReplication := &resourcev1alpha1.PulsarGeoReplication{}
	if err := r.conn.client.Get(ctx, types.NamespacedName{
		Namespace: topic.Namespace,
		Name:      ref.Name,
	}, geoReplication); err != nil {
		return err
	}
	r.log.V(1).Info("Found geo replication", "GEO Replication", geoReplication.Name)
	destConnection := &resourcev1alpha1.PulsarConnection{}
	if err := r.conn.client.Get(ctx, types.NamespacedName{
		Name:      geoReplication.Spec.DestinationConnectionRef.Name,
		Namespace: geoReplication.Namespace,
	}, destConnection); err != nil {
		return err
	}

	params.ReplicationClusters = append(params.ReplicationClusters, destConnection.Spec.ClusterName)
	params.ReplicationClusters = append(params.ReplicationClusters, r.conn.connection.Spec.ClusterName)
	return nil
}
