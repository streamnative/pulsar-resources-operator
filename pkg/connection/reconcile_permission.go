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

	// Extract current desired state
	currentState := r.extractCurrentState(permission)

	// Get previous state from annotation
	previousState, err := r.getPreviousState(permission)
	if err != nil {
		log.Error(err, "Failed to get previous state from annotation")
		return err
	}

	// Check for context changes (ResourceType or ResourceName)
	contextChanged := false
	if previousState != nil {
		contextChanged = previousState.ResourceType != currentState.ResourceType ||
			previousState.ResourceName != currentState.ResourceName
	}

	if contextChanged {
		log.Info("Context change detected, cleaning up previous permissions",
			"previousResourceType", previousState.ResourceType,
			"currentResourceType", currentState.ResourceType,
			"previousResourceName", previousState.ResourceName,
			"currentResourceName", currentState.ResourceName)

		// Clean up previous context
		if err := r.cleanupPreviousContext(permission, *previousState); err != nil {
			log.Error(err, "Failed to cleanup previous context, continuing with current operations")
		}
	}

	// Determine roles to manage
	var previouslyManagedRoles []string
	if previousState != nil && !contextChanged {
		previouslyManagedRoles = previousState.Roles
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

	// Update the state annotation
	if err := r.updateStateAnnotation(permission, currentState); err != nil {
		log.Error(err, "Failed to update state annotation")
		return err
	}

	// Update the resource with new annotations
	if err := r.conn.client.Update(ctx, permission); err != nil {
		log.Error(err, "Failed to update permission annotations")
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

const (
	// PulsarPermissionStateAnnotation is the annotation key used to store the previous state
	// of PulsarPermission resources for stateful reconciliation
	PulsarPermissionStateAnnotation = "pulsarpermissions.resource.streamnative.io/managed-state"
)

// PulsarPermissionState represents the state that needs to be tracked for PulsarPermission resources
type PulsarPermissionState struct {
	ResourceType string   `json:"resourceType"`
	ResourceName string   `json:"resourceName"`
	Roles        []string `json:"roles"`
	Actions      []string `json:"actions"`
}


// extractCurrentState extracts the current desired state from the PulsarPermission spec
func (r *PulsarPermissionReconciler) extractCurrentState(permission *resourcev1alpha1.PulsarPermission) PulsarPermissionState {
	// Sort roles and actions for consistent comparison
	roles := make([]string, len(permission.Spec.Roles))
	copy(roles, permission.Spec.Roles)
	slices.Sort(roles)

	actions := make([]string, len(permission.Spec.Actions))
	copy(actions, permission.Spec.Actions)
	slices.Sort(actions)

	return PulsarPermissionState{
		ResourceType: string(permission.Spec.ResoureType),
		ResourceName: permission.Spec.ResourceName,
		Roles:        roles,
		Actions:      actions,
	}
}

// getPreviousState retrieves the previous state from the resource annotation
func (r *PulsarPermissionReconciler) getPreviousState(permission *resourcev1alpha1.PulsarPermission) (*PulsarPermissionState, error) {
	annotations := permission.GetAnnotations()
	if annotations == nil {
		r.log.V(1).Info("No annotations found, treating as first reconciliation")
		return nil, nil
	}

	stateJSON, exists := annotations[PulsarPermissionStateAnnotation]
	if !exists {
		r.log.V(1).Info("No previous state annotation found, treating as first reconciliation")
		return nil, nil
	}

	// Try to unmarshal as PulsarPermissionState
	var previousState PulsarPermissionState
	if err := json.Unmarshal([]byte(stateJSON), &previousState); err != nil {
		r.log.Error(err, "Failed to unmarshal previous state annotation, treating as first reconciliation",
			"annotation", stateJSON)
		return nil, nil
	}

	return &previousState, nil
}

// updateStateAnnotation updates the annotation with the current state after successful reconciliation
func (r *PulsarPermissionReconciler) updateStateAnnotation(permission *resourcev1alpha1.PulsarPermission, currentState PulsarPermissionState) error {
	stateJSON, err := json.Marshal(currentState)
	if err != nil {
		return err
	}

	annotations := permission.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
		permission.SetAnnotations(annotations)
	}

	// Only update if the value has changed
	currentValue := annotations[PulsarPermissionStateAnnotation]
	newValue := string(stateJSON)

	if currentValue != newValue {
		annotations[PulsarPermissionStateAnnotation] = newValue
		r.log.V(1).Info("Updated state annotation", "state", currentState)
	}

	return nil
}

// cleanupPreviousContext cleans up permissions from the previous resource context
func (r *PulsarPermissionReconciler) cleanupPreviousContext(permission *resourcev1alpha1.PulsarPermission, prevState PulsarPermissionState) error {
	if len(prevState.Roles) == 0 {
		return nil
	}

	r.log.Info("Cleaning up permissions from previous context",
		"previousResourceType", prevState.ResourceType,
		"previousResourceName", prevState.ResourceName,
		"rolesToCleanup", prevState.Roles)

	// Create a temporary permission resource for the previous context
	tempPermission := permission.DeepCopy()
	tempPermission.Spec.ResourceName = prevState.ResourceName
	tempPermission.Spec.ResoureType = resourcev1alpha1.PulsarResourceType(prevState.ResourceType)
	tempPermission.Spec.Roles = prevState.Roles
	tempPermission.Spec.Actions = prevState.Actions

	// Get permissioner for the previous context
	permissioner := GetPermissioner(tempPermission)
	if permissioner == nil {
		return fmt.Errorf("failed to get permissioner for previous context")
	}

	// Get the pulsar admin instance
	pulsarAdmin := r.conn.pulsarAdmin

	// Revoke all roles from the previous context
	for _, role := range prevState.Roles {
		r.log.Info("Revoking permission from previous context", "role", role)
		
		// Create a temporary permission for this specific role
		rolePermission := tempPermission.DeepCopy()
		rolePermission.Spec.Roles = []string{role}
		rolePermissioner := GetPermissioner(rolePermission)
		
		if err := pulsarAdmin.RevokePermissions(rolePermissioner); err != nil {
			r.log.Error(err, "Failed to revoke permission from previous context, continuing",
				"role", role,
				"previousResourceType", prevState.ResourceType,
				"previousResourceName", prevState.ResourceName)
			// Continue with other roles even if one fails
		}
	}

	return nil
}
