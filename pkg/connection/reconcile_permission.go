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
	"encoding/json"
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

	// Get previously managed roles from the last observed state to enable proper cleanup
	// The second return value indicates if a resource type/name change was detected and cleaned up
	previouslyManagedRoles, wasCleanedUp := r.getPreviouslyManagedRoles(permission)
	if wasCleanedUp {
		log.Info("Old permissions were cleaned up due to resource type/name change, proceeding with fresh reconciliation")
	}

	currentRoles := make([]string, 0, len(currentPermissions))
	incomingRoles := permission.Spec.Roles

	for role := range currentPermissions {
		currentRoles = append(currentRoles, role)
	}

	// Only revoke roles that were previously managed by this PulsarPermission resource
	// This prevents conflicts with other PulsarPermission resources managing the same target
	for _, role := range previouslyManagedRoles {
		// If this role is no longer in the incoming roles, and it currently exists, revoke it
		if !slices.Contains(incomingRoles, role) && slices.Contains(currentRoles, role) {
			log.Info("Revoking previously managed role", "role", role)
			tempPermission := permission.DeepCopy()
			tempPermission.Spec.Roles = []string{role}
			per := GetPermissioner(tempPermission)
			if err := pulsarAdmin.RevokePermissions(per); err != nil {
				log.Error(err, "Revoke permission failed", "role", role)
				meta.SetStatusCondition(&permission.Status.Conditions, *NewErrorCondition(permission.Generation, err.Error()))
				if err := r.conn.client.Status().Update(ctx, permission); err != nil {
					log.Error(err, "Failed to update permission status")
					return err
				}
				return err
			}
		}
	}

	// Grant permissions for all incoming roles
	for _, role := range incomingRoles {
		tempPermission := permission.DeepCopy()
		tempPermission.Spec.Roles = []string{role}
		per := GetPermissioner(tempPermission)
		if err := pulsarAdmin.GrantPermissions(per); err != nil {
			log.Error(err, "Grant permission failed", "role", role)
			meta.SetStatusCondition(&permission.Status.Conditions, *NewErrorCondition(permission.Generation, err.Error()))
			if err := r.conn.client.Status().Update(ctx, permission); err != nil {
				log.Error(err, "Failed to update permission status")
				return err
			}
			return err
		}
	}

	// Update the managed roles tracking for future reconciliations
	annotationUpdated := r.updateManagedRolesTracking(permission, incomingRoles)

	// Update the resource with new annotations if they changed
	if annotationUpdated {
		if err := r.conn.client.Update(ctx, permission); err != nil {
			log.Error(err, "Failed to update permission annotations")
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

const (
	// ManagedRolesAnnotation stores the roles that this PulsarPermission resource manages
	ManagedRolesAnnotation = "pulsarpermissions.resource.streamnative.io/managed-roles"
)

// ManagedRolesData represents the structure stored in the managed roles annotation.
// It includes resource type and name context to handle ResourceType changes properly.
type ManagedRolesData struct {
	ResourceType string   `json:"resourceType"`
	ResourceName string   `json:"resourceName"`
	Roles        []string `json:"roles"`
}

// getPreviouslyManagedRoles retrieves the list of roles that this PulsarPermission
// resource managed in its previous reconciliation. This is used to properly clean up
// roles that are removed from the spec while avoiding conflicts with other PulsarPermission resources.
// It also handles ResourceType changes by cleaning up old permissions when the type changes.
func (r *PulsarPermissionReconciler) getPreviouslyManagedRoles(permission *resourcev1alpha1.PulsarPermission) ([]string, bool) {
	if permission.Annotations == nil {
		// If no annotation exists, assume this is the first reconciliation
		// In this case, we don't have previous state, so return empty list
		return []string{}, false
	}

	managedRolesJSON, exists := permission.Annotations[ManagedRolesAnnotation]
	if !exists {
		return []string{}, false
	}

	// Try to unmarshal as new format (ManagedRolesData) first
	var managedData ManagedRolesData
	if err := json.Unmarshal([]byte(managedRolesJSON), &managedData); err == nil {
		// Check if ResourceType or ResourceName changed
		typeChanged := managedData.ResourceType != string(permission.Spec.ResoureType)
		nameChanged := managedData.ResourceName != permission.Spec.ResourceName

		if typeChanged || nameChanged {
			// ResourceType or ResourceName changed, need to clean up old permissions
			r.log.Info("Detected resource type or name change, will clean up old permissions",
				"oldType", managedData.ResourceType,
				"newType", permission.Spec.ResoureType,
				"oldName", managedData.ResourceName,
				"newName", permission.Spec.ResourceName)

			// Clean up old permissions
			r.cleanupOldPermissions(permission, managedData)

			// Return empty list for current reconciliation (fresh start)
			return []string{}, true
		}

		// Same resource type and name, return the roles
		return managedData.Roles, false
	}

	// Try to unmarshal as legacy format ([]string) for backward compatibility
	var legacyRoles []string
	if err := json.Unmarshal([]byte(managedRolesJSON), &legacyRoles); err != nil {
		r.log.Error(err, "Failed to unmarshal managed roles annotation, treating as empty", "annotation", managedRolesJSON)
		return []string{}, false
	}

	// Legacy format doesn't have resource type info, assume no change needed
	r.log.V(1).Info("Found legacy managed roles annotation, will upgrade to new format")
	return legacyRoles, false
}

// updateManagedRolesTracking updates the annotation that tracks which roles this
// PulsarPermission resource is responsible for managing. This enables proper cleanup
// in future reconciliations when roles are removed from the spec.
// Returns true if the annotation was updated, false otherwise.
func (r *PulsarPermissionReconciler) updateManagedRolesTracking(permission *resourcev1alpha1.PulsarPermission, managedRoles []string) bool {
	if permission.Annotations == nil {
		permission.Annotations = make(map[string]string)
	}

	// Create the new format with resource type context
	managedData := ManagedRolesData{
		ResourceType: string(permission.Spec.ResoureType),
		ResourceName: permission.Spec.ResourceName,
		Roles:        managedRoles,
	}

	managedRolesJSON, err := json.Marshal(managedData)
	if err != nil {
		r.log.Error(err, "Failed to marshal managed roles for annotation, skipping tracking update")
		return false
	}

	// Only update the annotation if it has changed to avoid unnecessary updates
	currentValue, exists := permission.Annotations[ManagedRolesAnnotation]
	newValue := string(managedRolesJSON)

	if !exists || currentValue != newValue {
		permission.Annotations[ManagedRolesAnnotation] = newValue
		r.log.V(1).Info("Updated managed roles tracking", "managedRoles", managedRoles, "resourceType", managedData.ResourceType, "resourceName", managedData.ResourceName)
		return true
	}

	return false
}

// cleanupOldPermissions removes permissions that were managed under the old resource type/name
// when a ResourceType or ResourceName change is detected.
func (r *PulsarPermissionReconciler) cleanupOldPermissions(permission *resourcev1alpha1.PulsarPermission, oldManagedData ManagedRolesData) {
	if len(oldManagedData.Roles) == 0 {
		return
	}

	// Get the old resource type to determine which API to use
	var oldResourceType resourcev1alpha1.PulsarResourceType
	switch oldManagedData.ResourceType {
	case string(resourcev1alpha1.PulsarResourceTypeNamespace):
		oldResourceType = resourcev1alpha1.PulsarResourceTypeNamespace
	case string(resourcev1alpha1.PulsarResourceTypeTopic):
		oldResourceType = resourcev1alpha1.PulsarResourceTypeTopic
	default:
		r.log.Error(nil, "Unknown old resource type, skipping cleanup", "resourceType", oldManagedData.ResourceType)
		return
	}

	r.log.Info("Cleaning up old permissions due to resource type/name change",
		"oldResourceType", oldManagedData.ResourceType,
		"oldResourceName", oldManagedData.ResourceName,
		"rolesToCleanup", oldManagedData.Roles)

	// Create a temporary permission spec for the old resource to perform cleanup
	tempPermission := &resourcev1alpha1.PulsarPermission{
		Spec: resourcev1alpha1.PulsarPermissionSpec{
			ConnectionRef: permission.Spec.ConnectionRef,
			ResourceName:  oldManagedData.ResourceName,
			ResoureType:   oldResourceType,
			Roles:         oldManagedData.Roles,
			Actions:       permission.Spec.Actions, // Use current actions for cleanup
		},
	}

	// Get the appropriate permissioner for the old resource type
	permissioner := GetPermissioner(tempPermission)
	if permissioner == nil {
		r.log.Error(nil, "Failed to get permissioner for old resource type cleanup",
			"resourceType", oldManagedData.ResourceType,
			"resourceName", oldManagedData.ResourceName)
		return
	}

	// Revoke all old permissions
	for _, role := range oldManagedData.Roles {
		// Create a temporary permission for this specific role
		rolePermission := &resourcev1alpha1.PulsarPermission{
			Spec: resourcev1alpha1.PulsarPermissionSpec{
				ConnectionRef: permission.Spec.ConnectionRef,
				ResourceName:  oldManagedData.ResourceName,
				ResoureType:   oldResourceType,
				Roles:         []string{role},
				Actions:       permission.Spec.Actions,
			},
		}
		rolePermissioner := GetPermissioner(rolePermission)
		if err := r.conn.pulsarAdmin.RevokePermissions(rolePermissioner); err != nil {
			r.log.Error(err, "Failed to revoke old permission during cleanup, continuing with other roles",
				"role", role,
				"oldResourceType", oldManagedData.ResourceType,
				"oldResourceName", oldManagedData.ResourceName)
		} else {
			r.log.Info("Successfully revoked old permission during cleanup",
				"role", role,
				"oldResourceType", oldManagedData.ResourceType,
				"oldResourceName", oldManagedData.ResourceName)
		}
	}
}
