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
	cloudapi "github.com/streamnative/pulsar-resources-operator/pkg/streamnativecloud/apis/cloud/v1alpha1"
	cloudclient "github.com/streamnative/pulsar-resources-operator/pkg/streamnativecloud/client/clientset_generated/clientset"
)

// ServiceAccountBindingClient handles ServiceAccountBinding operations with the API server
type ServiceAccountBindingClient struct {
	client cloudclient.Interface
	// organization is the organization to use for API calls
	organization string
}

// NewServiceAccountBindingClient creates a new ServiceAccountBinding client
func NewServiceAccountBindingClient(apiConn *APIConnection, organization string) (*ServiceAccountBindingClient, error) {
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

	return &ServiceAccountBindingClient{
		client:       client,
		organization: organization,
	}, nil
}

// WatchServiceAccountBinding watches a ServiceAccountBinding by name
func (c *ServiceAccountBindingClient) WatchServiceAccountBinding(ctx context.Context, name string) (watch.Interface, error) {
	// Use ListOptions to watch a single object
	opts := metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", name),
		Watch:         true,
	}
	return c.client.CloudV1alpha1().ServiceAccountBindings(c.organization).Watch(ctx, opts)
}

// GetServiceAccountBinding gets a ServiceAccountBinding by name
func (c *ServiceAccountBindingClient) GetServiceAccountBinding(ctx context.Context, name string) (*cloudapi.ServiceAccountBinding, error) {
	return c.client.CloudV1alpha1().ServiceAccountBindings(c.organization).Get(ctx, name, metav1.GetOptions{})
}

// convertToCloudServiceAccountBinding converts a local ServiceAccountBinding to a cloud ServiceAccountBinding
func convertToCloudServiceAccountBinding(binding *resourcev1alpha1.ServiceAccountBinding, organization string) *cloudapi.ServiceAccountBinding {
	// Convert to cloud API type
	cloudBinding := &cloudapi.ServiceAccountBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: binding.Name,
		},
		Spec: cloudapi.ServiceAccountBindingSpec{},
	}

	// Add fields from our ServiceAccountBinding CRD
	if binding.Spec.ServiceAccountName != "" {
		cloudBinding.Spec.ServiceAccountName = binding.Spec.ServiceAccountName
	}

	// Convert PoolMemberRef - use the first one if multiple are provided
	if len(binding.Spec.PoolMemberRefs) > 0 {
		cloudBinding.Spec.PoolMemberRef = cloudapi.PoolMemberReference{
			Namespace: binding.Spec.PoolMemberRefs[0].Namespace,
			Name:      binding.Spec.PoolMemberRefs[0].Name,
		}

		if binding.Spec.PoolMemberRefs[0].Namespace == "" {
			cloudBinding.Spec.PoolMemberRef.Namespace = organization
		}
	}

	return cloudBinding
}

// CreateServiceAccountBinding creates a new ServiceAccountBinding
func (c *ServiceAccountBindingClient) CreateServiceAccountBinding(ctx context.Context, binding *resourcev1alpha1.ServiceAccountBinding) (*cloudapi.ServiceAccountBinding, error) {
	cloudBinding := convertToCloudServiceAccountBinding(binding, c.organization)

	// Create ServiceAccountBinding
	return c.client.CloudV1alpha1().ServiceAccountBindings(c.organization).Create(ctx, cloudBinding, metav1.CreateOptions{})
}

// UpdateServiceAccountBinding updates an existing ServiceAccountBinding
func (c *ServiceAccountBindingClient) UpdateServiceAccountBinding(ctx context.Context, binding *resourcev1alpha1.ServiceAccountBinding) (*cloudapi.ServiceAccountBinding, error) {
	// Get existing ServiceAccountBinding
	existing, err := c.GetServiceAccountBinding(ctx, binding.Name)
	if err != nil {
		return nil, err
	}

	// Create updated version
	updated := convertToCloudServiceAccountBinding(binding, c.organization)
	// Preserve ResourceVersion for optimistic concurrency
	updated.ResourceVersion = existing.ResourceVersion

	// Update ServiceAccountBinding
	return c.client.CloudV1alpha1().ServiceAccountBindings(c.organization).Update(ctx, updated, metav1.UpdateOptions{})
}

// DeleteServiceAccountBinding deletes a ServiceAccountBinding by name
func (c *ServiceAccountBindingClient) DeleteServiceAccountBinding(ctx context.Context, binding *resourcev1alpha1.ServiceAccountBinding) error {
	return c.client.CloudV1alpha1().ServiceAccountBindings(c.organization).Delete(ctx, binding.Name, metav1.DeleteOptions{})
}
