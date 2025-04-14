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

// ServiceAccountClient handles ServiceAccount operations with the API server
type ServiceAccountClient struct {
	client cloudclient.Interface
	// organization is the organization to use for API calls
	organization string
}

// NewServiceAccountClient creates a new ServiceAccount client
func NewServiceAccountClient(apiConn *APIConnection, organization string) (*ServiceAccountClient, error) {
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

	return &ServiceAccountClient{
		client:       client,
		organization: organization,
	}, nil
}

// WatchServiceAccount watches a ServiceAccount by name
func (c *ServiceAccountClient) WatchServiceAccount(ctx context.Context, name string) (watch.Interface, error) {
	// Use ListOptions to watch a single object
	opts := metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", name),
		Watch:         true,
	}
	return c.client.CloudV1alpha1().ServiceAccounts(c.organization).Watch(ctx, opts)
}

// GetServiceAccount gets a ServiceAccount by name
func (c *ServiceAccountClient) GetServiceAccount(ctx context.Context, name string) (*cloudapi.ServiceAccount, error) {
	return c.client.CloudV1alpha1().ServiceAccounts(c.organization).Get(ctx, name, metav1.GetOptions{})
}

// convertToCloudServiceAccount converts a local ServiceAccount to a cloud ServiceAccount
func convertToCloudServiceAccount(sa *resourcev1alpha1.ServiceAccount) *cloudapi.ServiceAccount {
	// Convert to cloud API type
	cloudSA := &cloudapi.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: sa.Name,
		},
		Spec: cloudapi.ServiceAccountSpec{},
	}

	// Add fields from our ServiceAccount CRD
	// The cloud ServiceAccount spec is currently empty, but we may need to add fields here
	// if the cloud API requires specific fields in the future

	return cloudSA
}

// CreateServiceAccount creates a new ServiceAccount
func (c *ServiceAccountClient) CreateServiceAccount(ctx context.Context, sa *resourcev1alpha1.ServiceAccount) (*cloudapi.ServiceAccount, error) {
	cloudSA := convertToCloudServiceAccount(sa)

	// Create ServiceAccount
	created, err := c.client.CloudV1alpha1().ServiceAccounts(c.organization).Create(ctx, cloudSA, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return created, nil
}

// UpdateServiceAccount updates an existing ServiceAccount
func (c *ServiceAccountClient) UpdateServiceAccount(ctx context.Context, sa *resourcev1alpha1.ServiceAccount) (*cloudapi.ServiceAccount, error) {
	// Get existing ServiceAccount
	existing, err := c.GetServiceAccount(ctx, sa.Name)
	if err != nil {
		return nil, err
	}

	// Create updated version
	updated := convertToCloudServiceAccount(sa)
	// Preserve ResourceVersion for optimistic concurrency
	updated.ResourceVersion = existing.ResourceVersion

	// Update ServiceAccount
	return c.client.CloudV1alpha1().ServiceAccounts(c.organization).Update(ctx, updated, metav1.UpdateOptions{})
}

// DeleteServiceAccount deletes a ServiceAccount by name
func (c *ServiceAccountClient) DeleteServiceAccount(ctx context.Context, sa *resourcev1alpha1.ServiceAccount) error {
	return c.client.CloudV1alpha1().ServiceAccounts(c.organization).Delete(ctx, sa.Name, metav1.DeleteOptions{})
}
