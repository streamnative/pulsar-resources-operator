package controllers

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"

	computeapi "github.com/streamnative/cloud-api-server/pkg/apis/compute/v1alpha1"
	cloudclient "github.com/streamnative/cloud-api-server/pkg/client/clientset_generated/clientset"
	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
)

// WorkspaceClient handles workspace operations with the API server
type WorkspaceClient struct {
	client cloudclient.Interface
	// organization is the organization to use for API calls
	organization string
}

// NewWorkspaceClient creates a new workspace client
func NewWorkspaceClient(apiConn *APIConnection, organization string) (*WorkspaceClient, error) {
	if apiConn == nil || apiConn.config == nil || apiConn.client == nil {
		return nil, fmt.Errorf("api connection is nil")
	}
	if !apiConn.IsInitialized() {
		return nil, fmt.Errorf("api connection is not initialized")
	}
	// Create REST config
	config := &rest.Config{
		Host:      apiConn.config.Spec.Server,
		Transport: apiConn.client.Transport,
	}

	// Create client
	client, err := cloudclient.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &WorkspaceClient{
		client:       client,
		organization: organization,
	}, nil
}

// WatchWorkspace watches a workspace by name
func (c *WorkspaceClient) WatchWorkspace(ctx context.Context, name string) (watch.Interface, error) {
	// Use ListOptions to watch a single object
	opts := metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", name),
		Watch:         true,
	}
	return c.client.ComputeV1alpha1().Workspaces(c.organization).Watch(ctx, opts)
}

// GetWorkspace gets a workspace by ID
func (c *WorkspaceClient) GetWorkspace(ctx context.Context, id string) (*computeapi.Workspace, error) {
	return c.client.ComputeV1alpha1().Workspaces(c.organization).Get(ctx, id, metav1.GetOptions{})
}

// CreateWorkspace creates a new workspace
func (c *WorkspaceClient) CreateWorkspace(ctx context.Context, workspace *resourcev1alpha1.ComputeWorkspace) (*computeapi.Workspace, error) {
	// Convert to cloud API type
	cloudWorkspace := &computeapi.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name: workspace.Name,
		},
		Spec: computeapi.WorkspaceSpec{
			PulsarClusterNames: workspace.Spec.PulsarClusterNames,
			UseExternalAccess:  workspace.Spec.UseExternalAccess,
		},
	}

	if workspace.Spec.PoolRef != nil {
		cloudWorkspace.Spec.PoolRef = &computeapi.PoolRef{
			Name:      workspace.Spec.PoolRef.Name,
			Namespace: workspace.Spec.PoolRef.Namespace,
		}
	}

	if workspace.Spec.FlinkBlobStorage != nil {
		cloudWorkspace.Spec.FlinkBlobStorage = &computeapi.FlinkBlobStorage{
			Bucket: workspace.Spec.FlinkBlobStorage.Bucket,
			Path:   workspace.Spec.FlinkBlobStorage.Path,
		}
	}

	// Create workspace
	return c.client.ComputeV1alpha1().Workspaces(c.organization).Create(ctx, cloudWorkspace, metav1.CreateOptions{})
}

// UpdateWorkspace updates a workspace
func (c *WorkspaceClient) UpdateWorkspace(ctx context.Context, workspace *resourcev1alpha1.ComputeWorkspace) (*computeapi.Workspace, error) {
	// Get existing workspace
	existing, err := c.GetWorkspace(ctx, workspace.Name)
	if err != nil {
		return nil, err
	}

	// Update spec
	existing.Spec.PulsarClusterNames = workspace.Spec.PulsarClusterNames
	existing.Spec.UseExternalAccess = workspace.Spec.UseExternalAccess

	if workspace.Spec.PoolRef != nil {
		existing.Spec.PoolRef = &computeapi.PoolRef{
			Name:      workspace.Spec.PoolRef.Name,
			Namespace: workspace.Spec.PoolRef.Namespace,
		}
	} else {
		existing.Spec.PoolRef = nil
	}

	if workspace.Spec.FlinkBlobStorage != nil {
		existing.Spec.FlinkBlobStorage = &computeapi.FlinkBlobStorage{
			Bucket: workspace.Spec.FlinkBlobStorage.Bucket,
			Path:   workspace.Spec.FlinkBlobStorage.Path,
		}
	} else {
		existing.Spec.FlinkBlobStorage = nil
	}

	// Update workspace
	return c.client.ComputeV1alpha1().Workspaces(c.organization).Update(ctx, existing, metav1.UpdateOptions{})
}

// DeleteWorkspace deletes a workspace
func (c *WorkspaceClient) DeleteWorkspace(ctx context.Context, workspace *resourcev1alpha1.ComputeWorkspace) error {
	return c.client.ComputeV1alpha1().Workspaces(c.organization).Delete(ctx, workspace.Name, metav1.DeleteOptions{})
}
