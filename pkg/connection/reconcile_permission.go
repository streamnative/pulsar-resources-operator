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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	"github.com/streamnative/pulsar-resources-operator/pkg/admin"
	"github.com/streamnative/pulsar-resources-operator/pkg/reconciler"
)

// PulsarPermissionReconciler reconciles a PulsarPermission object
type PulsarPermissionReconciler struct {
	conn *PulsarConnectionReconciler
	log  logr.Logger
}

func makePermissionsReconciler(r *PulsarConnectionReconciler) reconciler.Interface {
	return &PulsarPermissionReconciler{
		conn: r,
		log:  r.log.WithName("PulsarPermission"),
	}
}

// Observe checks the updates of object
func (r *PulsarPermissionReconciler) Observe(ctx context.Context) error {
	r.log.V(1).Info("Start Observe")

	permissionList := &resourcev1alpha1.PulsarPermissionList{}
	if err := r.conn.client.List(ctx, permissionList, client.InNamespace(r.conn.connection.Namespace),
		client.MatchingFields(map[string]string{
			".spec.connectionRef.name": r.conn.connection.Name,
		})); err != nil {
		return fmt.Errorf("list permission [%w]. connection[%s]", err, r.conn.connection.Name)
	}
	r.log.V(1).Info("Observed permissions items", "Count", len(permissionList.Items))
	r.conn.permissions = permissionList.Items

	if !r.conn.hasUnreadyResource {
		for i := range r.conn.permissions {
			if !resourcev1alpha1.IsPulsarResourceReady(&r.conn.permissions[i]) {
				r.conn.hasUnreadyResource = true
				break
			}
		}
	}

	r.log.V(1).Info("Observe Done")
	return nil
}

// Reconcile reconciles all permissions
func (r *PulsarPermissionReconciler) Reconcile(ctx context.Context) error {
	for i := range r.conn.permissions {
		perm := &r.conn.permissions[i]
		if err := r.ReconcilePermission(ctx, r.conn.pulsarAdmin, perm); err != nil {
			return fmt.Errorf("reconcile permission [%w]", err)
		}
	}

	return nil
}

// ReconcilePermission move the current state of the toic closer to the desired state
func (r *PulsarPermissionReconciler) ReconcilePermission(ctx context.Context, pulsarAdmin admin.PulsarAdmin,
	permission *resourcev1alpha1.PulsarPermission) error {
	log := r.log.WithValues("pulsarpermission", permission.Name, "namespace", permission.Namespace)
	log.V(1).Info("Start Reconcile")

	per := GetPermissioner(permission)

	if !permission.DeletionTimestamp.IsZero() {
		if permission.Spec.LifecyclePolicy == resourcev1alpha1.CleanUpAfterDeletion {
			log.Info("Revoking permission", "LifecyclePolicy", permission.Spec.LifecyclePolicy)

			if err := pulsarAdmin.RevokePermissions(per); err != nil && !admin.IsNotFound(err) {
				log.Error(err, "Failed to revoke permission")
				return err
			}
		}
		// TODO use otelcontroller until kube-instrumentation upgrade controller-runtime version to newer
		controllerutil.RemoveFinalizer(permission, resourcev1alpha1.FinalizerName)
		if err := r.conn.client.Update(ctx, permission); err != nil {
			r.conn.log.Error(err, "Failed to remove finalizer")
			return err
		}
		return nil
	}

	if permission.Spec.LifecyclePolicy == resourcev1alpha1.CleanUpAfterDeletion {
		// TODO use otelcontroller until kube-instrumentation upgrade controller-runtime version to newer
		controllerutil.AddFinalizer(permission, resourcev1alpha1.FinalizerName)
		if err := r.conn.client.Update(ctx, permission); err != nil {
			r.conn.log.Error(err, "Failed to add finalizer")
			return err
		}
	}

	if resourcev1alpha1.IsPulsarResourceReady(permission) {
		r.conn.log.V(1).Info("Resource is ready")
		return nil
	}

	log.V(1).Info("Granting permission", "ResourceName", permission.Spec.ResourceName,
		"ResourceType", permission.Spec.ResoureType, "Roles", permission.Spec.Roles, "Actions", permission.Spec.Actions)
	if err := pulsarAdmin.GrantPermissions(per); err != nil {
		log.Error(err, "Grant permission failed")
		meta.SetStatusCondition(&permission.Status.Conditions, *NewErrorCondition(permission.Generation, err.Error()))
		if err := r.conn.client.Status().Update(ctx, permission); err != nil {
			log.Error(err, "Failed to update permission status")
			return err
		}
		return err
	}

	permission.Status.ObservedGeneration = permission.Generation
	meta.SetStatusCondition(&permission.Status.Conditions, *NewReadyCondition(permission.Generation))
	if err := r.conn.client.Status().Update(ctx, permission); err != nil {
		log.Error(err, "Failed to update permission status")
		return err
	}

	log.V(1).Info("Reconcile Done")
	return nil
}

// GetPermissioner will return Permissioner according resource type
func GetPermissioner(p *resourcev1alpha1.PulsarPermission) admin.Permissioner {
	switch p.Spec.ResoureType {
	case resourcev1alpha1.PulsarResourceTypeNamespace:
		ns := &admin.NamespacePermission{
			ResourceName: p.Spec.ResourceName,
			Roles:        p.Spec.Roles,
			Actions:      p.Spec.Actions,
		}
		return ns

	case resourcev1alpha1.PulsarResourceTypeTopic:
		topic := admin.TopicPermission{
			ResourceName: p.Spec.ResourceName,
			Roles:        p.Spec.Roles,
			Actions:      p.Spec.Actions,
		}
		return &topic
	}
	return nil
}
