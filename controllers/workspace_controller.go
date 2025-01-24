package controllers

import (
	"context"
	"fmt"
	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	controllers2 "github.com/streamnative/pulsar-resources-operator/pkg/streamnativecloud"
	"sync"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	computeapi "github.com/streamnative/cloud-api-server/pkg/apis/compute/v1alpha1"
)

// WorkspaceReconciler reconciles a Workspace object
type WorkspaceReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	ConnectionManager *ConnectionManager
	// watcherMap stores active watchers for Workspaces
	watcherMap map[types.NamespacedName]watch.Interface
	// watcherMutex protects watcherMap
	watcherMutex sync.RWMutex
}

//+kubebuilder:rbac:groups=resource.streamnative.io,resources=computeworkspaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=computeworkspaces/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=computeworkspaces/finalizers,verbs=update
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=streamnativecloudconnections,verbs=get;list;watch

// handleWatchEvents processes events from the watch interface
func (r *WorkspaceReconciler) handleWatchEvents(ctx context.Context, namespacedName types.NamespacedName, watcher watch.Interface) {
	logger := log.FromContext(ctx)
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-watcher.ResultChan():
			if !ok {
				logger.Info("Watch channel closed", "namespace", namespacedName.Namespace, "name", namespacedName.Name)
				// Remove the watcher from the map
				r.watcherMutex.Lock()
				delete(r.watcherMap, namespacedName)
				r.watcherMutex.Unlock()
				return
			}

			switch event.Type {
			case watch.Modified:
				remoteWorkspace, ok := event.Object.(*computeapi.Workspace)
				if !ok {
					logger.Error(fmt.Errorf("unexpected object type"), "Failed to convert object to Workspace")
					continue
				}

				// Get the local workspace
				localWorkspace := &resourcev1alpha1.ComputeWorkspace{}
				if err := r.Get(ctx, namespacedName, localWorkspace); err != nil {
					logger.Error(err, "Failed to get local Workspace")
					continue
				}

				// Update workspace ID and status
				localWorkspace.Status.WorkspaceID = remoteWorkspace.Name

				// Update status
				r.updateWorkspaceStatus(ctx, localWorkspace, nil, "Ready", "Workspace synced successfully")
			}
		}
	}
}

// setupWatch creates a new watcher for a Workspace
func (r *WorkspaceReconciler) setupWatch(ctx context.Context, workspace *resourcev1alpha1.ComputeWorkspace, workspaceClient *controllers2.WorkspaceClient) error {
	namespacedName := types.NamespacedName{
		Namespace: workspace.Namespace,
		Name:      workspace.Name,
	}

	// Check if we already have a watcher
	r.watcherMutex.RLock()
	_, exists := r.watcherMap[namespacedName]
	r.watcherMutex.RUnlock()
	if exists {
		return nil
	}

	// Create new watcher
	watcher, err := workspaceClient.WatchWorkspace(ctx, workspace.Name)
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}

	// Store watcher in map
	r.watcherMutex.Lock()
	r.watcherMap[namespacedName] = watcher
	r.watcherMutex.Unlock()

	// Start watching in a new goroutine
	go r.handleWatchEvents(ctx, namespacedName, watcher)
	return nil
}

// Reconcile handles the reconciliation of Workspace objects
func (r *WorkspaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling Workspace", "namespace", req.Namespace, "name", req.Name)

	// Add requeue interval for status sync
	requeueInterval := time.Minute

	// Get the Workspace resource
	workspace := &resourcev1alpha1.ComputeWorkspace{}
	if err := r.Get(ctx, req.NamespacedName, workspace); err != nil {
		if apierrors.IsNotFound(err) {
			// Stop and remove watcher if it exists
			r.watcherMutex.Lock()
			if watcher, exists := r.watcherMap[req.NamespacedName]; exists {
				watcher.Stop()
				delete(r.watcherMap, req.NamespacedName)
			}
			r.watcherMutex.Unlock()
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Get the APIServerConnection
	connection := &resourcev1alpha1.StreamNativeCloudConnection{}
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: req.Namespace,
		Name:      workspace.Spec.APIServerRef.Name,
	}, connection); err != nil {
		r.updateWorkspaceStatus(ctx, workspace, err, "ConnectionNotFound",
			fmt.Sprintf("Failed to get APIServerConnection: %v", err))
		return ctrl.Result{}, err
	}

	// Get API connection
	apiConn, err := r.ConnectionManager.GetOrCreateConnection(connection, nil)
	if err != nil {
		// If connection is not initialized, requeue the request
		if _, ok := err.(*NotInitializedError); ok {
			logger.Info("Connection not initialized, requeueing", "error", err.Error())
			return ctrl.Result{Requeue: true}, nil
		}
		r.updateWorkspaceStatus(ctx, workspace, err, "GetConnectionFailed",
			fmt.Sprintf("Failed to get connection: %v", err))
		return ctrl.Result{}, err
	}

	// Get organization from connection or workspace
	organization := connection.Spec.Organization
	if organization == "" {
		err := fmt.Errorf("organization is required but not specified")
		r.updateWorkspaceStatus(ctx, workspace, err, "ValidationFailed", err.Error())
		return ctrl.Result{}, err
	}

	// Create workspace client
	workspaceClient, err := controllers2.NewWorkspaceClient(apiConn, organization)
	if err != nil {
		r.updateWorkspaceStatus(ctx, workspace, err, "ClientCreationFailed",
			fmt.Sprintf("Failed to create workspace client: %v", err))
		return ctrl.Result{}, err
	}

	// Handle deletion
	if !workspace.DeletionTimestamp.IsZero() {
		if controllers2.ContainsString(workspace.Finalizers, controllers2.WorkspaceFinalizer) {
			// Try to delete remote workspace
			if err := workspaceClient.DeleteWorkspace(ctx, workspace); err != nil {
				if !apierrors.IsNotFound(err) {
					r.updateWorkspaceStatus(ctx, workspace, err, "DeleteFailed",
						fmt.Sprintf("Failed to delete external resources: %v", err))
					return ctrl.Result{}, err
				}
				// If the resource is already gone, that's fine
				logger.Info("Remote Workspace already deleted or not found",
					"workspace", workspace.Name)
			}

			// Remove finalizer after successful deletion
			workspace.Finalizers = controllers2.RemoveString(workspace.Finalizers, controllers2.WorkspaceFinalizer)
			if err := r.Update(ctx, workspace); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer if it doesn't exist
	if !controllers2.ContainsString(workspace.Finalizers, controllers2.WorkspaceFinalizer) {
		workspace.Finalizers = append(workspace.Finalizers, controllers2.WorkspaceFinalizer)
		if err := r.Update(ctx, workspace); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Check if workspace exists
	existingWorkspace, err := workspaceClient.GetWorkspace(ctx, workspace.Name)
	if err != nil {
		logger.Info("Failed to get workspace", "error", err, "existingWorkspace", existingWorkspace)
		if !apierrors.IsNotFound(err) {
			r.updateWorkspaceStatus(ctx, workspace, err, "GetWorkspaceFailed",
				fmt.Sprintf("Failed to get workspace: %v", err))
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		existingWorkspace = nil
	}

	if existingWorkspace == nil {
		// Create workspace
		resp, err := workspaceClient.CreateWorkspace(ctx, workspace)
		if err != nil {
			r.updateWorkspaceStatus(ctx, workspace, err, "CreateWorkspaceFailed",
				fmt.Sprintf("Failed to create workspace: %v", err))
			return ctrl.Result{}, err
		}

		// Update status
		workspace.Status.WorkspaceID = resp.Name
		r.updateWorkspaceStatus(ctx, workspace, nil, "Ready", "Workspace created successfully")
	} else {
		// Update workspace
		resp, err := workspaceClient.UpdateWorkspace(ctx, workspace)
		if err != nil {
			r.updateWorkspaceStatus(ctx, workspace, err, "UpdateWorkspaceFailed",
				fmt.Sprintf("Failed to update workspace: %v", err))
			return ctrl.Result{}, err
		}
		// Update status
		workspace.Status.WorkspaceID = resp.Name
		r.updateWorkspaceStatus(ctx, workspace, nil, "Ready", "Workspace updated successfully")
	}

	// Setup watch after workspace is created/updated
	if err := r.setupWatch(ctx, workspace, workspaceClient); err != nil {
		logger.Error(err, "Failed to setup watch")
		// Don't return error, just log it
	}

	return ctrl.Result{RequeueAfter: requeueInterval}, nil
}

func (r *WorkspaceReconciler) updateWorkspaceStatus(
	ctx context.Context,
	workspace *resourcev1alpha1.ComputeWorkspace,
	err error,
	reason string,
	message string,
) {
	// Create new condition
	condition := metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: workspace.Generation,
		LastTransitionTime: metav1.Now(),
	}

	if err != nil {
		condition.Status = metav1.ConditionFalse
	}

	// Create new conditions slice for comparison
	newConditions := make([]metav1.Condition, 0)
	for _, c := range workspace.Status.Conditions {
		if c.Type != condition.Type {
			newConditions = append(newConditions, c)
		}
	}
	newConditions = append(newConditions, condition)

	// Check if status has actually changed
	if !controllers2.StatusHasChanged(workspace.Status.Conditions, newConditions) {
		return
	}

	// Update conditions
	workspace.Status.Conditions = newConditions

	// Update observed generation
	workspace.Status.ObservedGeneration = workspace.Generation

	// Update status
	if err := r.Status().Update(ctx, workspace); err != nil {
		log.FromContext(ctx).Error(err, "Failed to update Workspace status")
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *WorkspaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Initialize the watcher map
	r.watcherMap = make(map[types.NamespacedName]watch.Interface)

	return ctrl.NewControllerManagedBy(mgr).
		For(&resourcev1alpha1.ComputeWorkspace{}).
		WithEventFilter(predicate.Or(
			// Trigger on spec changes
			predicate.GenerationChangedPredicate{},
			// Trigger periodically to sync status
			predicate.NewPredicateFuncs(func(object client.Object) bool {
				// Trigger every minute to sync status
				return true
			}),
		)).
		Complete(r)
}
