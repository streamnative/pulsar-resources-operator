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
	"k8s.io/apimachinery/pkg/api/meta"
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
		log:  r.log.WithName("PulsarTopic"),
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
	if !r.conn.hasUnreadyResource {
		for i := range r.conn.topics {
			if !resourcev1alpha1.IsPulsarResourceReady(&r.conn.topics[i]) {
				r.conn.hasUnreadyResource = true
				break
			}
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
	log := r.log.WithValues("pulsartopic", topic.Name, "namespace", topic.Namespace)
	log.V(1).Info("Start Reconcile")

	if !topic.DeletionTimestamp.IsZero() {
		log.Info("Deleting topic", "LifecyclePolicy", topic.Spec.LifecyclePolicy)
		if topic.Spec.LifecyclePolicy == resourcev1alpha1.CleanUpAfterDeletion {
			// Delete the schema of the topic before the deletion
			if topic.Spec.SchemaInfo != nil {
				log.Info("Deleting topic schema")
				err := pulsarAdmin.DeleteSchema(topic.Spec.Name)
				if err != nil {
					return err
				}
			}

			if err := pulsarAdmin.DeleteTopic(topic.Spec.Name); err != nil && admin.IsNotFound(err) {
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

	if topic.Spec.LifecyclePolicy == resourcev1alpha1.CleanUpAfterDeletion {
		// TODO use otelcontroller until kube-instrumentation upgrade controller-runtime version to newer
		controllerutil.AddFinalizer(topic, resourcev1alpha1.FinalizerName)
		if err := r.conn.client.Update(ctx, topic); err != nil {
			log.Error(err, "Failed to add finalizer")
			return err
		}
	}

	if resourcev1alpha1.IsPulsarResourceReady(topic) {
		log.V(1).Info("Resource is ready")
		return nil
	}

	params := &admin.TopicParams{
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

	r.applyDefault(params)
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
