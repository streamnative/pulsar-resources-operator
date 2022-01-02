// Copyright (c) 2020 StreamNative, Inc.. All Rights Reserved.

package connection

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	commonsreconciler "github.com/streamnative/pulsar-operators/commons/pkg/controller/reconciler"
	pulsarv1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	"github.com/streamnative/pulsar-resources-operator/pkg/admin"
)

// PulsarPermissionReconciler reconciles a PulsarPermission object
type PulsarPermissionReconciler struct {
	conn *PulsarConnectionReconciler
	log  logr.Logger
}

func makePermissionsReconciler(r *PulsarConnectionReconciler) commonsreconciler.Interface {
	return &PulsarPermissionReconciler{
		conn: r,
		log:  r.log.WithName("PulsarPermission"),
	}
}

func (r *PulsarPermissionReconciler) Observe(ctx context.Context) error {
	log := r.log.WithValues("Observe Namespace", r.conn.connection.Namespace)
	log.V(1).Info("Start Observe")

	permissionList := &pulsarv1alpha1.PulsarPermissionList{}
	if err := r.conn.client.List(ctx, permissionList, client.InNamespace(r.conn.connection.Namespace),
		client.MatchingFields(map[string]string{
			".spec.connectionRef.name": r.conn.connection.Name,
		})); err != nil {
		return fmt.Errorf("list permission [%w]. connection[%s]", err, r.conn.connection.Name)
	}
	log.V(1).Info("Observed permissions items", "Count", len(permissionList.Items))
	r.conn.permissions = permissionList.Items

	if !r.conn.hasUnreadyResource {
		for i := range r.conn.permissions {
			if !pulsarv1alpha1.IsPulsarResourceReady(&r.conn.permissions[i]) {
				r.conn.hasUnreadyResource = true
				break
			}
		}
	}

	log.V(1).Info("Observe Done")
	return nil
}

func (r *PulsarPermissionReconciler) Reconcile(ctx context.Context) error {
	for i := range r.conn.permissions {
		perm := &r.conn.permissions[i]
		if err := r.ReconcilePermission(ctx, r.conn.pulsarAdmin, perm); err != nil {
			return fmt.Errorf("reconcile permission [%w]", err)
		}
	}

	return nil
}

func (r *PulsarPermissionReconciler) ReconcilePermission(ctx context.Context, pulsarAdmin admin.PulsarAdmin,
	permission *pulsarv1alpha1.PulsarPermission) error {
	log := r.log.WithValues("PulsarPermission", permission.Name)
	log.V(1).Info("Start Reconcile")

	per := GetPermissioner(permission)

	if !permission.DeletionTimestamp.IsZero() {
		if permission.Spec.LifecyclePolicy == pulsarv1alpha1.CleanUpAfterDeletion {
			log.Info("Revoking permission", "LifecyclePolicy", permission.Spec.LifecyclePolicy)

			if err := pulsarAdmin.RevokePermissions(per); err != nil && admin.IsNotFound(err) {
				return err
			}
		}
		// TODO use otelcontroller until kube-instrumentation upgrade controller-runtime version to newer
		controllerutil.RemoveFinalizer(permission, pulsarv1alpha1.FinalizerName)
		if err := r.conn.client.Update(ctx, permission); err != nil {
			return err
		}
		return nil
	}

	if permission.Spec.LifecyclePolicy == pulsarv1alpha1.CleanUpAfterDeletion {
		// TODO use otelcontroller until kube-instrumentation upgrade controller-runtime version to newer
		controllerutil.AddFinalizer(permission, pulsarv1alpha1.FinalizerName)
		if err := r.conn.client.Update(ctx, permission); err != nil {
			return err
		}
	}

	if pulsarv1alpha1.IsPulsarResourceReady(permission) {
		return nil
	}

	log.V(1).Info("Granting permission", "ResourceName", permission.Spec.ResourceName,
		"ResourceType", permission.Spec.ResoureType, "Roles", permission.Spec.Roles, "Actions", permission.Spec.Actions)
	if err := pulsarAdmin.GrantPermissions(per); err != nil {
		log.Error(err, "Grant permission failed", "Name", permission.Name, "Error", err)
		meta.SetStatusCondition(&permission.Status.Conditions, *NewErrorCondition(permission.Generation, err.Error()))
		if err := r.conn.client.Status().Update(ctx, permission); err != nil {
			log.Error(err, "Failed to update PulsarPermission status", "Namespace", r.conn.connection.Namespace,
				"Name", permission.Name)
			return err
		}
		return err
	}

	permission.Status.ObservedGeneration = permission.Generation
	meta.SetStatusCondition(&permission.Status.Conditions, *NewReadyCondition(permission.Generation))
	if err := r.conn.client.Status().Update(ctx, permission); err != nil {
		return err
	}

	log.V(1).Info("Reconcile Done")
	return nil
}

// GetPermissioner will return Permissioner according resource type
func GetPermissioner(p *pulsarv1alpha1.PulsarPermission) admin.Permissioner {
	switch p.Spec.ResoureType {
	case pulsarv1alpha1.PulsarResourceTypeNamespace:
		ns := &admin.NamespacePermission{
			ResourceName: p.Spec.ResourceName,
			Roles:        p.Spec.Roles,
			Actions:      p.Spec.Actions,
		}
		return ns

	case pulsarv1alpha1.PulsarResourceTypeTopic:
		topic := admin.TopicPermission{
			ResourceName: p.Spec.ResourceName,
			Roles:        p.Spec.Roles,
			Actions:      p.Spec.Actions,
		}
		return &topic
	}
	return nil
}
