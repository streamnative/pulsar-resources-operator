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

	cloudapi "github.com/streamnative/cloud-api-server/pkg/apis/cloud/v1alpha1"
	cloudclient "github.com/streamnative/cloud-api-server/pkg/client/clientset_generated/clientset"
	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
)

// SecretClient handles Secret operations with the API server
type SecretClient struct {
	client cloudclient.Interface
	// organization is the organization to use for API calls
	organization string
}

// NewSecretClient creates a new Secret client
func NewSecretClient(apiConn *APIConnection, organization string) (*SecretClient, error) {
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

	return &SecretClient{
		client:       client,
		organization: organization,
	}, nil
}

// WatchSecret watches a Secret by name
func (c *SecretClient) WatchSecret(ctx context.Context, name string) (watch.Interface, error) {
	// Use ListOptions to watch a single object
	opts := metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", name),
		Watch:         true,
	}
	return c.client.CloudV1alpha1().Secrets(c.organization).Watch(ctx, opts)
}

// GetSecret gets a Secret by name
func (c *SecretClient) GetSecret(ctx context.Context, name string) (*cloudapi.Secret, error) {
	return c.client.CloudV1alpha1().Secrets(c.organization).Get(ctx, name, metav1.GetOptions{})
}

// convertToCloudSecret converts a local Secret to a cloud Secret
func convertToCloudSecret(secret *resourcev1alpha1.Secret) *cloudapi.Secret {
	// Convert to cloud API type
	cloudSecret := &cloudapi.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secret.Name,
		},
		// Create the spec from scratch using the data from our CRD
		Spec: cloudapi.SecretSpec{},
	}

	// We'll use reflection or structured generic mapping to convert our CRD to cloud API
	// Add data fields where the API expects them

	// Add fields from our Secret CRD
	if secret.Spec.Data != nil {
		cloudSecret.Data = secret.Spec.Data
	}
	if secret.Spec.InstanceName != "" {
		cloudSecret.InstanceName = secret.Spec.InstanceName
	}
	if secret.Spec.Location != "" {
		cloudSecret.Location = secret.Spec.Location
	}
	if secret.Spec.PoolMemberName != nil && *secret.Spec.PoolMemberName != "" {
		cloudSecret.PoolMemberRef = &cloudapi.PoolMemberReference{Namespace: secret.Namespace, Name: *secret.Spec.PoolMemberName}
	}

	// Convert tolerations if they exist
	if len(secret.Spec.Tolerations) > 0 {
		cloudSecret.Tolerations = make([]cloudapi.Toleration, 0, len(secret.Spec.Tolerations))
		for _, t := range secret.Spec.Tolerations {
			cloudSecret.Tolerations = append(cloudSecret.Tolerations, cloudapi.Toleration{
				Key:      t.Key,
				Operator: t.Operator,
				Value:    t.Value,
				Effect:   t.Effect,
			})
		}
	}

	return cloudSecret
}

// CreateSecret creates a new Secret
func (c *SecretClient) CreateSecret(ctx context.Context, secret *resourcev1alpha1.Secret) (*cloudapi.Secret, error) {
	cloudSecret := convertToCloudSecret(secret)

	// Create secret
	return c.client.CloudV1alpha1().Secrets(c.organization).Create(ctx, cloudSecret, metav1.CreateOptions{})
}

// UpdateSecret updates an existing Secret
func (c *SecretClient) UpdateSecret(ctx context.Context, secret *resourcev1alpha1.Secret) (*cloudapi.Secret, error) {
	// Get existing secret
	existing, err := c.GetSecret(ctx, secret.Name)
	if err != nil {
		return nil, err
	}

	// Create updated version
	updated := convertToCloudSecret(secret)
	// Preserve ResourceVersion for optimistic concurrency
	updated.ResourceVersion = existing.ResourceVersion

	// Update secret
	return c.client.CloudV1alpha1().Secrets(c.organization).Update(ctx, updated, metav1.UpdateOptions{})
}

// DeleteSecret deletes a Secret by name
func (c *SecretClient) DeleteSecret(ctx context.Context, secret *resourcev1alpha1.Secret) error {
	return c.client.CloudV1alpha1().Secrets(c.organization).Delete(ctx, secret.Name, metav1.DeleteOptions{})
}
