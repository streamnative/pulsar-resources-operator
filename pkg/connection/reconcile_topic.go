// Copyright (c) 2022 StreamNative, Inc.. All Rights Reserved.

package connection

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	commonsreconciler "github.com/streamnative/pulsar-operators/commons/pkg/controller/reconciler"
	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	"github.com/streamnative/pulsar-resources-operator/pkg/admin"
)

// PulsarTopicReconciler reconciles a PulsarTopic object
type PulsarTopicReconciler struct {
	conn *PulsarConnectionReconciler
}

func makeTopicsReconciler(r *PulsarConnectionReconciler) commonsreconciler.Interface {
	return &PulsarTopicReconciler{
		conn: r,
	}
}

// Observe checks the updates of object
func (r *PulsarTopicReconciler) Observe(ctx context.Context) error {
	topicsList := &resourcev1alpha1.PulsarTopicList{}
	if err := r.conn.client.List(ctx, topicsList, client.InNamespace(r.conn.connection.Namespace),
		client.MatchingFields(map[string]string{
			".spec.connectionRef.name": r.conn.connection.Name,
		})); err != nil {
		return fmt.Errorf("list topics [%w]", err)
	}
	r.conn.topics = topicsList.Items
	if !r.conn.hasUnreadyResource {
		for i := range r.conn.topics {
			if !resourcev1alpha1.IsPulsarResourceReady(&r.conn.topics[i]) {
				r.conn.hasUnreadyResource = true
				break
			}
		}
	}
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
	if !topic.DeletionTimestamp.IsZero() {
		if topic.Spec.LifecyclePolicy == resourcev1alpha1.CleanUpAfterDeletion {
			if err := pulsarAdmin.DeleteTopic(topic.Spec.Name); err != nil && admin.IsNotFound(err) {
				return err
			}
		}

		// TODO use otelcontroller until kube-instrumentation upgrade controller-runtime version to newer
		controllerutil.RemoveFinalizer(topic, resourcev1alpha1.FinalizerName)
		if err := r.conn.client.Update(ctx, topic); err != nil {
			return err
		}
		return nil
	}

	if topic.Spec.LifecyclePolicy == resourcev1alpha1.CleanUpAfterDeletion {
		// TODO use otelcontroller until kube-instrumentation upgrade controller-runtime version to newer
		controllerutil.AddFinalizer(topic, resourcev1alpha1.FinalizerName)
		if err := r.conn.client.Update(ctx, topic); err != nil {
			return err
		}
	}

	if resourcev1alpha1.IsPulsarResourceReady(topic) {
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
		if err := r.conn.client.Status().Update(ctx, topic); err != nil {
			r.conn.log.Error(err, "Failed to update connection status", "Namespace", r.conn.connection.Namespace,
				"Name", r.conn.connection.Name)
			return nil
		}
		return err
	}

	topic.Status.ObservedGeneration = topic.Generation
	meta.SetStatusCondition(&topic.Status.Conditions, *NewReadyCondition(topic.Generation))
	if err := r.conn.client.Status().Update(ctx, topic); err != nil {
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
