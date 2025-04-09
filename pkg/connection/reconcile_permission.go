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
	"slices"

	"github.com/apache/pulsar-client-go/pulsaradmin/pkg/utils"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	"github.com/streamnative/pulsar-resources-operator/pkg/admin"
	"github.com/streamnative/pulsar-resources-operator/pkg/feature"
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
		log:  makeSubResourceLog(r, "PulsarPermission"),
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

	for i := range r.conn.permissions {
		if !resourcev1alpha1.IsPulsarResourceReady(&r.conn.permissions[i]) {
			r.conn.addUnreadyResource(&r.conn.permissions[i])
		}
	}

	r.log.V(1).Info("Observe Done")
	return nil
}

// Reconcile reconciles all permissions
func (r *PulsarPermissionReconciler) Reconcile(ctx context.Context) error {
	errs := []error{}
	for i := range r.conn.permissions {
		perm := &r.conn.permissions[i]
		if err := r.ReconcilePermission(ctx, r.conn.pulsarAdmin, perm); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("reconcile permission [%v]", errs)
	}

	return nil
}

// ReconcilePermission move the current state of the toic closer to the desired state
func (r *PulsarPermissionReconciler) ReconcilePermission(ctx context.Context, pulsarAdmin admin.PulsarAdmin,
	permission *resourcev1alpha1.PulsarPermission) error {
	log := r.log.WithValues("name", permission.Name, "namespace", permission.Namespace)
	log.Info("Start Reconcile")

	per := GetPermissioner(permission)

	if !permission.DeletionTimestamp.IsZero() {
		log.Info("Revoking permission", "LifecyclePolicy", permission.Spec.LifecyclePolicy)
		if permission.Spec.LifecyclePolicy != resourcev1alpha1.KeepAfterDeletion {
			if err := pulsarAdmin.RevokePermissions(per); err != nil && !admin.IsNotFound(err) {
				log.Error(err, "Failed to revoke permission")
				return err
			}
		}
		// TODO use otelcontroller until kube-instrumentation upgrade controller-runtime version to newer
		controllerutil.RemoveFinalizer(permission, resourcev1alpha1.FinalizerName)
		if err := r.conn.client.Update(ctx, permission); err != nil {
			log.Error(err, "Failed to remove finalizer")
			return err
		}
		return nil
	}

	if permission.Spec.LifecyclePolicy != resourcev1alpha1.KeepAfterDeletion {
		// TODO use otelcontroller until kube-instrumentation upgrade controller-runtime version to newer
		controllerutil.AddFinalizer(permission, resourcev1alpha1.FinalizerName)
		if err := r.conn.client.Update(ctx, permission); err != nil {
			log.Error(err, "Failed to add finalizer")
			return err
		}
	}

	if resourcev1alpha1.IsPulsarResourceReady(permission) &&
		!feature.DefaultFeatureGate.Enabled(feature.AlwaysUpdatePulsarResource) {
		log.Info("Skip reconcile, permission resource is ready")
		return nil
	}

	log.Info("Updating permission", "ResourceName", permission.Spec.ResourceName,
		"ResourceType", permission.Spec.ResoureType, "Roles", permission.Spec.Roles, "Actions", permission.Spec.Actions)

	var currentPermissions map[string][]utils.AuthAction
	var err error

	if permission.Spec.ResoureType == resourcev1alpha1.PulsarResourceTypeTopic {
		currentPermissions, err = pulsarAdmin.GetTopicPermissions(permission.Spec.ResourceName)
	} else {
		currentPermissions, err = pulsarAdmin.GetNamespacePermissions(permission.Spec.ResourceName)
	}

	if err != nil {
		log.Error(err, "Failed to get current permissions")
		meta.SetStatusCondition(&permission.Status.Conditions, *NewErrorCondition(permission.Generation, err.Error()))
		if err := r.conn.client.Status().Update(ctx, permission); err != nil {
			log.Error(err, "Failed to update permission status")
		}
		return err
	}

	currentRoles := []string{}
	incomingRoles := permission.Spec.Roles

	for role := range currentPermissions {
		currentRoles = append(currentRoles, role)
	}

	// revoking roles
	for _, role := range currentRoles {
		if !slices.Contains(incomingRoles, role) {
			permission.Spec.Roles = []string{role}
			per := GetPermissioner(permission)
			if err := pulsarAdmin.RevokePermissions(per); err != nil {
				log.Error(err, "Revoke permission failed")
				meta.SetStatusCondition(&permission.Status.Conditions, *NewErrorCondition(permission.Generation, err.Error()))
				if err := r.conn.client.Status().Update(ctx, permission); err != nil {
					log.Error(err, "Failed to update permission status")
					return err
				}
				return err
			}
		}
	}

	// granting roles
	for _, role := range incomingRoles {
		permission.Spec.Roles = []string{role}
		per := GetPermissioner(permission)
		if err := pulsarAdmin.GrantPermissions(per); err != nil {
			log.Error(err, "Grant permission failed")
			meta.SetStatusCondition(&permission.Status.Conditions, *NewErrorCondition(permission.Generation, err.Error()))
			if err := r.conn.client.Status().Update(ctx, permission); err != nil {
				log.Error(err, "Failed to update permission status")
				return err
			}
			return err
		}
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
