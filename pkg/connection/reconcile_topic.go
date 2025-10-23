// Copyright 2025 StreamNative
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
	"slices"

	"github.com/apache/pulsar-client-go/pulsaradmin/pkg/utils"
	"github.com/go-logr/logr"
	"github.com/streamnative/pulsar-resources-operator/pkg/feature"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
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
		if !isPulsarTopicResourceReady(&r.conn.topics[i]) {
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
			// return error will stop the other reconcile process
			r.log.Error(err, "Failed to reconcile topic", "topicName", topic.Spec.Name)
			continue
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
					if admin.IsNoSuchHostError(err) {
						log.Info("Pulsar cluster has been deleted")
					} else {
						log.Error(err, "Failed to reset the cluster for topic")
						return err
					}
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

	if isPulsarTopicResourceReady(topic) &&
		!feature.DefaultFeatureGate.Enabled(feature.AlwaysUpdatePulsarResource) {
		log.Info("Skip reconcile, topic resource is ready")
		return nil
	}

	var policyErrs []error
	var creationErr error
	defer func() {
		if creationErr != nil {
			meta.SetStatusCondition(&topic.Status.Conditions,
				NewTopicErrorCondition(topic.Generation, resourcev1alpha1.ConditionReady, creationErr.Error()))
		} else {
			meta.SetStatusCondition(&topic.Status.Conditions,
				NewTopicReadyCondition(topic.Generation, resourcev1alpha1.ConditionReady))
		}
		if len(policyErrs) != 0 || creationErr != nil {
			msg := ""
			for _, err := range policyErrs {
				msg += err.Error() + ";\n"
			}
			r.conn.retryer.CreateIfAbsent(topic)
			meta.SetStatusCondition(&topic.Status.Conditions,
				NewTopicErrorCondition(topic.Generation, resourcev1alpha1.ConditionTopicPolicyReady, msg))
		} else {
			meta.SetStatusCondition(&topic.Status.Conditions,
				NewTopicReadyCondition(topic.Generation, resourcev1alpha1.ConditionTopicPolicyReady))
		}

		if err := r.conn.client.Status().Update(ctx, topic); err != nil {
			log.Error(err, "Failed to update status")
		}
	}()

	params := createTopicParams(topic)

	r.applyDefault(params)

	if refs := topic.Spec.GeoReplicationRefs; len(refs) != 0 || len(topic.Spec.ReplicationClusters) > 0 {
		if len(refs) > 0 && len(topic.Spec.ReplicationClusters) > 0 {
			return fmt.Errorf("GeoReplicationRefs and ReplicationClusters cannot be set at the same time")
		}

		if len(refs) > 0 {
			for _, ref := range refs {
				if err := r.applyGeo(ctx, params, ref, topic); err != nil {
					log.Error(err, "Failed to get destination connection for geo replication "+ref.Name)
					policyErrs = append(policyErrs, err)
				}
			}
		} else if len(topic.Spec.ReplicationClusters) > 0 {
			topicNameStr := admin.MakeCompleteTopicName(topic.Spec.Name, topic.Spec.Persistent)
			topicName, err := utils.GetTopicName(topicNameStr)
			if err != nil {
				return err
			}
			tenantName := topicName.GetTenant()
			allowedClusters, err := pulsarAdmin.GetTenantAllowedClusters(tenantName)
			if err != nil {
				return err
			}
			for _, cluster := range topic.Spec.ReplicationClusters {
				if !slices.Contains(allowedClusters, cluster) {
					err := fmt.Errorf("cluster %s is not allowed in tenant %s", cluster, tenantName)
					meta.SetStatusCondition(&topic.Status.Conditions, *NewErrorCondition(topic.Generation, err.Error()))
					log.Error(err, "Failed to apply topic")
					if err := r.conn.client.Status().Update(ctx, topic); err != nil {
						log.Error(err, "Failed to update the topic status")
						return err
					}
					return err
				}
				params.ReplicationClusters = append(params.ReplicationClusters, topic.Spec.ReplicationClusters...)
			}
		}

		if len(params.ReplicationClusters) > 0 {
			log.Info("apply topic with replication clusters", "clusters", params.ReplicationClusters)
			topic.Status.GeoReplicationEnabled = true
		}
	} else if topic.Status.GeoReplicationEnabled {
		// when GeoReplicationRefs is removed, it should reset the topic clusters
		params.ReplicationClusters = []string{r.conn.connection.Spec.ClusterName}
		log.Info("Geo Replication disabled. Reset topic with local cluster", "cluster", params.ReplicationClusters)
		topic.Status.GeoReplicationEnabled = false
	}

	creationErr, policyErr := pulsarAdmin.ApplyTopic(topic.Spec.Name, params)
	log.Info("Apply topic", "creationErr", creationErr, "policyErr", policyErr)
	if policyErr != nil {
		policyErrs = append(policyErrs, policyErr)
	}
	if creationErr != nil {
		return creationErr
	}

	if err := r.reconcileTopicCompactionState(ctx, topic); err != nil {
		log.Error(err, "Failed to reconcile topic compaction state")
		policyErrs = append(policyErrs, err)
	}

	if err := applySchema(pulsarAdmin, topic, log); err != nil {
		policyErrs = append(policyErrs, err)
	}

	topic.Status.ObservedGeneration = topic.Generation
	return nil
}

func applySchema(pulsarAdmin admin.PulsarAdmin, topic *resourcev1alpha1.PulsarTopic, log logr.Logger) error {
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
				return err
			}
		}
	}
	// Note: We intentionally do NOT delete existing schemas when schemaInfo is not specified.
	// This preserves existing schemas that may have been created by producers/consumers,
	// which is the expected behavior for most users. If schema deletion is needed,
	// users should explicitly manage it through the Pulsar admin APIs.
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
		Deduplication:                     topic.Spec.Deduplication,
		CompactionThreshold:               topic.Spec.CompactionThreshold,
		PersistencePolicies:               topic.Spec.PersistencePolicies,
		DelayedDelivery:                   topic.Spec.DelayedDelivery,
		DispatchRate:                      topic.Spec.DispatchRate,
		PublishRate:                       topic.Spec.PublishRate,
		InactiveTopicPolicies:             topic.Spec.InactiveTopicPolicies,
		SubscribeRate:                     topic.Spec.SubscribeRate,
		MaxMessageSize:                    topic.Spec.MaxMessageSize,
		MaxConsumersPerSubscription:       topic.Spec.MaxConsumersPerSubscription,
		MaxSubscriptionsPerTopic:          topic.Spec.MaxSubscriptionsPerTopic,
		SchemaValidationEnforced:          topic.Spec.SchemaValidationEnforced,
		SubscriptionDispatchRate:          topic.Spec.SubscriptionDispatchRate,
		ReplicatorDispatchRate:            topic.Spec.ReplicatorDispatchRate,
		DeduplicationSnapshotInterval:     topic.Spec.DeduplicationSnapshotInterval,
		OffloadPolicies:                   topic.Spec.OffloadPolicies,
		AutoSubscriptionCreation:          topic.Spec.AutoSubscriptionCreation,
		SchemaCompatibilityStrategy:       topic.Spec.SchemaCompatibilityStrategy,
		Properties:                        topic.Spec.Properties,
	}
}

func (r *PulsarTopicReconciler) reconcileTopicCompactionState(ctx context.Context, topic *resourcev1alpha1.PulsarTopic) error {
	original := topic.DeepCopy()
	reconciler := newTopicCompactionStateReconciler(r.log, r.conn.pulsarAdmin)
	changed, err := reconciler.reconcile(ctx, topic)
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}

	if err := r.conn.client.Patch(ctx, topic, client.MergeFrom(original)); err != nil {
		return err
	}
	return nil
}

func (r *PulsarTopicReconciler) applyDefault(params *admin.TopicParams) {
	if params.Persistent == nil {
		// by default create persistent topic
		params.Persistent = ptr.To(true)
	}

	if params.Partitions == nil {
		// by default create non-partitioned topic
		params.Partitions = ptr.To(int32(0))
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

func isPulsarTopicResourceReady(topic *resourcev1alpha1.PulsarTopic) bool {
	condition := meta.FindStatusCondition(topic.Status.Conditions, resourcev1alpha1.ConditionTopicPolicyReady)
	return resourcev1alpha1.IsPulsarResourceReady(topic) && condition != nil && condition.Status == metav1.ConditionTrue
}

// NewTopicReadyCondition make condition with ready info
func NewTopicReadyCondition(generation int64, conditionType string) metav1.Condition {
	return metav1.Condition{
		Type:               conditionType,
		Status:             metav1.ConditionTrue,
		ObservedGeneration: generation,
		Reason:             "Reconciled",
		Message:            "",
	}
}

// NewTopicErrorCondition make condition with ready info
func NewTopicErrorCondition(generation int64, conditionType, msg string) metav1.Condition {
	return metav1.Condition{
		Type:               conditionType,
		Status:             metav1.ConditionFalse,
		ObservedGeneration: generation,
		Reason:             "ReconcileError",
		Message:            msg,
	}
}
