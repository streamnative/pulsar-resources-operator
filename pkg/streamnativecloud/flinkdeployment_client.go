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

package streamnativecloud

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"

	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	computeapi "github.com/streamnative/pulsar-resources-operator/pkg/streamnativecloud/apis/compute/v1alpha1"
	cloudclient "github.com/streamnative/pulsar-resources-operator/pkg/streamnativecloud/client/clientset_generated/clientset"
)

// FlinkDeploymentClient handles FlinkDeployment operations with the API server
type FlinkDeploymentClient struct {
	client cloudclient.Interface
	// organization is the organization to use for API calls
	organization string
}

// NewFlinkDeploymentClient creates a new FlinkDeployment client
func NewFlinkDeploymentClient(apiConn *APIConnection, organization string) (*FlinkDeploymentClient, error) {
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

	return &FlinkDeploymentClient{
		client:       client,
		organization: organization,
	}, nil
}

// WatchFlinkDeployment watches a FlinkDeployment by name
func (c *FlinkDeploymentClient) WatchFlinkDeployment(ctx context.Context, name string) (watch.Interface, error) {
	// Use ListOptions to watch a single object
	opts := metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", name),
		Watch:         true,
	}
	return c.client.ComputeV1alpha1().FlinkDeployments(c.organization).Watch(ctx, opts)
}

// GetFlinkDeployment gets a FlinkDeployment by name
func (c *FlinkDeploymentClient) GetFlinkDeployment(ctx context.Context, name string) (*computeapi.FlinkDeployment, error) {
	return c.client.ComputeV1alpha1().FlinkDeployments(c.organization).Get(ctx, name, metav1.GetOptions{})
}

// CreateFlinkDeployment creates a new FlinkDeployment
func (c *FlinkDeploymentClient) CreateFlinkDeployment(ctx context.Context, deployment *resourcev1alpha1.ComputeFlinkDeployment) (*computeapi.FlinkDeployment, error) {
	// Convert to cloud API type
	cloudDeployment := &computeapi.FlinkDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deployment.Name,
		},
		Spec: computeapi.FlinkDeploymentSpec{
			WorkspaceName:        deployment.Spec.WorkspaceName,
			DefaultPulsarCluster: deployment.Spec.DefaultPulsarCluster,
			Configuration:        convertDeploymentConfiguration(deployment.Spec.Configuration),
		},
	}

	if deployment.Spec.Template != nil {
		cloudDeployment.Spec.Template = &computeapi.VvpDeploymentTemplate{
			SyncingMode: deployment.Spec.Template.SyncingMode,
			Deployment:  convertVvpDeploymentTemplateSpec(deployment, deployment.Spec.Template.Deployment),
		}
	}

	// Create deployment
	return c.client.ComputeV1alpha1().FlinkDeployments(c.organization).Create(ctx, cloudDeployment, metav1.CreateOptions{})
}

// UpdateFlinkDeployment updates an existing FlinkDeployment
func (c *FlinkDeploymentClient) UpdateFlinkDeployment(ctx context.Context, deployment *resourcev1alpha1.ComputeFlinkDeployment) (*computeapi.FlinkDeployment, error) {
	// Get existing deployment
	existing, err := c.GetFlinkDeployment(ctx, deployment.Name)
	if err != nil {
		return nil, err
	}

	// Update spec
	existing.Spec.WorkspaceName = deployment.Spec.WorkspaceName
	existing.Spec.DefaultPulsarCluster = deployment.Spec.DefaultPulsarCluster

	if deployment.Spec.Template != nil {
		existing.Spec.Template = &computeapi.VvpDeploymentTemplate{
			SyncingMode: deployment.Spec.Template.SyncingMode,
			Deployment:  convertVvpDeploymentTemplateSpec(deployment, deployment.Spec.Template.Deployment),
		}
	} else {
		existing.Spec.Template = nil
	}

	// Update deployment
	return c.client.ComputeV1alpha1().FlinkDeployments(c.organization).Update(ctx, existing, metav1.UpdateOptions{})
}

// DeleteFlinkDeployment deletes a FlinkDeployment by name
func (c *FlinkDeploymentClient) DeleteFlinkDeployment(ctx context.Context, deployment *resourcev1alpha1.ComputeFlinkDeployment) error {
	return c.client.ComputeV1alpha1().FlinkDeployments(c.organization).Delete(ctx, deployment.Name, metav1.DeleteOptions{})
}
