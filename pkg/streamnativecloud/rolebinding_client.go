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

// RoleBindingClient handles RoleBinding operations with the API server
type RoleBindingClient struct {
	client cloudclient.Interface
	// organization is the organization to use for API calls
	organization string
}

// NewRoleBindingClient creates a new RoleBinding client
func NewRoleBindingClient(apiConn *APIConnection, organization string) (*RoleBindingClient, error) {
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

	return &RoleBindingClient{
		client:       client,
		organization: organization,
	}, nil
}

// WatchRoleBinding watches a RoleBinding by name
func (c *RoleBindingClient) WatchRoleBinding(ctx context.Context, name string) (watch.Interface, error) {
	// Use ListOptions to watch a single object
	opts := metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", name),
		Watch:         true,
	}
	return c.client.CloudV1alpha1().RoleBindings(c.organization).Watch(ctx, opts)
}

// GetRoleBinding gets a RoleBinding by name
func (c *RoleBindingClient) GetRoleBinding(ctx context.Context, name string) (*cloudapi.RoleBinding, error) {
	return c.client.CloudV1alpha1().RoleBindings(c.organization).Get(ctx, name, metav1.GetOptions{})
}

// convertToCloudRoleBinding converts a local RoleBinding to a cloud RoleBinding
func convertToCloudRoleBinding(roleBinding *resourcev1alpha1.RoleBinding, organization string) *cloudapi.RoleBinding {
	// Convert to cloud API type
	cloudRoleBinding := &cloudapi.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: roleBinding.Name,
		},
		Spec: cloudapi.RoleBindingSpec{
			Subjects: makeSubjects(roleBinding.Spec.Users, roleBinding.Spec.IdentityPools, roleBinding.Spec.ServiceAccounts),
			RoleRef: cloudapi.RoleRef{
				Kind:     "ClusterRole",
				Name:     roleBinding.Spec.ClusterRole,
				APIGroup: "cloud.streamnative.io",
			},
			ResourceNames: makeResourceNames(roleBinding.Spec, organization),
		},
	}

	// Handle optional fields
	if roleBinding.Spec.CEL != nil {
		cloudRoleBinding.Spec.CEL = roleBinding.Spec.CEL
	}

	return cloudRoleBinding
}

// makeSubjects creates Subject array from Users, IdentityPools, and ServiceAccounts
func makeSubjects(users []string, identityPools []string, serviceAccounts []string) []cloudapi.Subject {
	var subjects []cloudapi.Subject

	// Add users
	for _, user := range users {
		subjects = append(subjects, cloudapi.Subject{
			Kind:     "User",
			Name:     user,
			APIGroup: "cloud.streamnative.io",
		})
	}

	// Add identity pools
	for _, pool := range identityPools {
		subjects = append(subjects, cloudapi.Subject{
			Kind:     "IdentityPool",
			Name:     pool,
			APIGroup: "cloud.streamnative.io",
		})
	}

	// Add service accounts
	for _, sa := range serviceAccounts {
		subjects = append(subjects, cloudapi.Subject{
			Kind:     "ServiceAccount",
			Name:     sa,
			APIGroup: "cloud.streamnative.io",
		})
	}

	return subjects
}

// makeResourceNames creates ResourceName array from SRN fields
func makeResourceNames(spec resourcev1alpha1.RoleBindingSpec, organization string) []cloudapi.ResourceName {
	var resourceNames []cloudapi.ResourceName

	for idx := range findLongestStrArray(spec.SRNOrganization, spec.SRNInstance, spec.SRNCluster, spec.SRNTenant, spec.SRNNamespace, spec.SRNTopicDomain, spec.SRNTopicName, spec.SRNSubscription, spec.SRNServiceAccount, spec.SRNSecret) {
		resourceName := cloudapi.ResourceName{}
		resourceName.Organization = organization
		if len(spec.SRNInstance) > idx {
			resourceName.Instance = spec.SRNInstance[idx]
		}
		if len(spec.SRNCluster) > idx {
			resourceName.Cluster = spec.SRNCluster[idx]
		}
		if len(spec.SRNTenant) > idx {
			resourceName.Tenant = spec.SRNTenant[idx]
		}
		if len(spec.SRNNamespace) > idx {
			resourceName.Namespace = spec.SRNNamespace[idx]
		}
		if len(spec.SRNTopicDomain) > idx {
			resourceName.TopicDomain = spec.SRNTopicDomain[idx]
		}
		if len(spec.SRNTopicName) > idx {
			resourceName.TopicName = spec.SRNTopicName[idx]
		}
		if len(spec.SRNSubscription) > idx {
			resourceName.Subscription = spec.SRNSubscription[idx]
		}
		if len(spec.SRNServiceAccount) > idx {
			resourceName.ServiceAccount = spec.SRNServiceAccount[idx]
		}
		if len(spec.SRNSecret) > idx {
			resourceName.Secret = spec.SRNSecret[idx]
		}
		resourceNames = append(resourceNames, resourceName)
	}

	return resourceNames
}

func findLongestStrArray(arrays ...[]string) []string {
	var longest []string
	for _, arr := range arrays {
		if len(arr) > len(longest) {
			longest = arr
		}
	}
	return longest
}

// CreateRoleBinding creates a new RoleBinding
func (c *RoleBindingClient) CreateRoleBinding(ctx context.Context, roleBinding *resourcev1alpha1.RoleBinding) (*cloudapi.RoleBinding, error) {
	cloudRoleBinding := convertToCloudRoleBinding(roleBinding, c.organization)

	// Create RoleBinding
	return c.client.CloudV1alpha1().RoleBindings(c.organization).Create(ctx, cloudRoleBinding, metav1.CreateOptions{})
}

// UpdateRoleBinding updates an existing RoleBinding
func (c *RoleBindingClient) UpdateRoleBinding(ctx context.Context, roleBinding *resourcev1alpha1.RoleBinding) (*cloudapi.RoleBinding, error) {
	// Get existing RoleBinding
	existing, err := c.GetRoleBinding(ctx, roleBinding.Name)
	if err != nil {
		return nil, err
	}

	// Create updated version
	updated := convertToCloudRoleBinding(roleBinding, c.organization)
	// Preserve ResourceVersion for optimistic concurrency
	updated.ResourceVersion = existing.ResourceVersion

	// Update RoleBinding
	return c.client.CloudV1alpha1().RoleBindings(c.organization).Update(ctx, updated, metav1.UpdateOptions{})
}

// DeleteRoleBinding deletes a RoleBinding by name
func (c *RoleBindingClient) DeleteRoleBinding(ctx context.Context, roleBinding *resourcev1alpha1.RoleBinding) error {
	return c.client.CloudV1alpha1().RoleBindings(c.organization).Delete(ctx, roleBinding.Name, metav1.DeleteOptions{})
}
