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

// APIKeyClient handles APIKey operations with the API server
type APIKeyClient struct {
	client cloudclient.Interface
	// organization is the organization to use for API calls
	organization string
}

// NewAPIKeyClient creates a new APIKey client
func NewAPIKeyClient(apiConn *APIConnection, organization string) (*APIKeyClient, error) {
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

	return &APIKeyClient{
		client:       client,
		organization: organization,
	}, nil
}

// WatchAPIKey watches an APIKey by name
func (c *APIKeyClient) WatchAPIKey(ctx context.Context, name string) (watch.Interface, error) {
	// Use ListOptions to watch a single object
	opts := metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", name),
		Watch:         true,
	}
	return c.client.CloudV1alpha1().APIKeys(c.organization).Watch(ctx, opts)
}

// GetAPIKey gets an APIKey by name
func (c *APIKeyClient) GetAPIKey(ctx context.Context, name string) (*cloudapi.APIKey, error) {
	return c.client.CloudV1alpha1().APIKeys(c.organization).Get(ctx, name, metav1.GetOptions{})
}

// convertToCloudAPIKey converts a local APIKey to a cloud APIKey
func convertToCloudAPIKey(apiKey *resourcev1alpha1.APIKey) *cloudapi.APIKey {
	// Convert to cloud API type
	cloudAPIKey := &cloudapi.APIKey{
		ObjectMeta: metav1.ObjectMeta{
			Name: apiKey.Name,
		},
		Spec: cloudapi.APIKeySpec{},
	}

	// Add fields from our APIKey CRD
	if apiKey.Spec.InstanceName != "" {
		cloudAPIKey.Spec.InstanceName = apiKey.Spec.InstanceName
	}
	if apiKey.Spec.ServiceAccountName != "" {
		cloudAPIKey.Spec.ServiceAccountName = apiKey.Spec.ServiceAccountName
	}
	if apiKey.Spec.Description != "" {
		cloudAPIKey.Spec.Description = apiKey.Spec.Description
	}
	if apiKey.Spec.ExpirationTime != nil {
		cloudAPIKey.Spec.ExpirationTime = apiKey.Spec.ExpirationTime
	}
	cloudAPIKey.Spec.Revoke = apiKey.Spec.Revoke

	// Copy encryption key if provided
	if apiKey.Spec.EncryptionKey != nil {
		cloudAPIKey.Spec.EncryptionKey = &cloudapi.EncryptionKey{
			PEM: apiKey.Spec.EncryptionKey.PEM,
			JWK: apiKey.Spec.EncryptionKey.JWK,
		}
	}

	return cloudAPIKey
}

// CreateAPIKey creates a new APIKey
func (c *APIKeyClient) CreateAPIKey(ctx context.Context, apiKey *resourcev1alpha1.APIKey) (*cloudapi.APIKey, error) {
	cloudAPIKey := convertToCloudAPIKey(apiKey)

	// Create APIKey
	return c.client.CloudV1alpha1().APIKeys(c.organization).Create(ctx, cloudAPIKey, metav1.CreateOptions{})
}

// UpdateAPIKey updates an existing APIKey
func (c *APIKeyClient) UpdateAPIKey(ctx context.Context, apiKey *resourcev1alpha1.APIKey) (*cloudapi.APIKey, error) {
	// Get existing APIKey
	existing, err := c.GetAPIKey(ctx, apiKey.Name)
	if err != nil {
		return nil, err
	}

	// Create updated version
	updated := convertToCloudAPIKey(apiKey)
	// Preserve ResourceVersion for optimistic concurrency
	updated.ResourceVersion = existing.ResourceVersion

	// Update APIKey
	return c.client.CloudV1alpha1().APIKeys(c.organization).Update(ctx, updated, metav1.UpdateOptions{})
}

// DeleteAPIKey deletes an APIKey by name
func (c *APIKeyClient) DeleteAPIKey(ctx context.Context, apiKey *resourcev1alpha1.APIKey) error {
	return c.client.CloudV1alpha1().APIKeys(c.organization).Delete(ctx, apiKey.Name, metav1.DeleteOptions{})
}
