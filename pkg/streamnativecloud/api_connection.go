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
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/oauth2/clientcredentials"

	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
)

// APIConnection represents a connection to the API server
type APIConnection struct {
	config      *resourcev1alpha1.StreamNativeCloudConnection
	credentials *resourcev1alpha1.ServiceAccountCredentials
	client      *http.Client
	// initialized indicates whether the connection has been fully initialized
	initialized bool
}

// WellKnownConfig represents the OpenID Connect configuration
type WellKnownConfig struct {
	TokenEndpoint string `json:"token_endpoint"`
}

// NewAPIConnection creates a new API connection
func NewAPIConnection(
	config *resourcev1alpha1.StreamNativeCloudConnection,
	creds *resourcev1alpha1.ServiceAccountCredentials,
) (*APIConnection, error) {
	conn := &APIConnection{
		config:      config.DeepCopy(),
		credentials: creds,
		initialized: false,
	}

	if err := conn.connect(); err != nil {
		return conn, err // Return connection even if initialization fails
	}

	conn.initialized = true
	return conn, nil
}

// connect establishes the connection
func (c *APIConnection) connect() error {
	if c.credentials == nil {
		return fmt.Errorf("waiting for credentials")
	}
	// Get well-known configuration
	wellKnownURL := fmt.Sprintf("%s/.well-known/openid-configuration", c.credentials.IssuerURL)
	resp, err := http.Get(wellKnownURL)
	if err != nil {
		return fmt.Errorf("failed to get well-known config: %w", err)
	}
	defer resp.Body.Close()

	var config WellKnownConfig
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return fmt.Errorf("failed to decode well-known config: %w", err)
	}

	// Configure OAuth2 client credentials flow
	oauth2Config := &clientcredentials.Config{
		ClientID:     c.credentials.ClientID,
		ClientSecret: c.credentials.ClientSecret,
		TokenURL:     config.TokenEndpoint,
		EndpointParams: url.Values{
			"audience": []string{c.config.Spec.Server},
		},
	}

	// Create HTTP client with token source
	c.client = oauth2Config.Client(context.Background())

	return nil
}

// Test tests the connection
func (c *APIConnection) Test(ctx context.Context) error {
	if c.credentials == nil {
		return fmt.Errorf("waiting for credentials")
	}
	if c.client == nil {
		return fmt.Errorf("waiting for client")
	}
	req, err := http.NewRequestWithContext(ctx, "GET", c.config.Spec.Server+"/healthz", nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if !c.IsInitialized() {
		c.initialized = true
	}

	return nil
}

// NeedsUpdate checks if the connection needs to be updated
func (c *APIConnection) NeedsUpdate(
	config *resourcev1alpha1.StreamNativeCloudConnection,
	creds *resourcev1alpha1.ServiceAccountCredentials,
) bool {
	if creds == nil {
		return false
	}
	if c.credentials == nil {
		return true
	}
	return c.config.Spec.Server != config.Spec.Server ||
		c.credentials.ClientID != creds.ClientID ||
		c.credentials.ClientSecret != creds.ClientSecret ||
		c.credentials.IssuerURL != creds.IssuerURL
}

// Update updates the connection configuration
func (c *APIConnection) Update(
	config *resourcev1alpha1.StreamNativeCloudConnection,
	creds *resourcev1alpha1.ServiceAccountCredentials,
) error {
	c.config = config.DeepCopy()
	c.credentials = creds
	return c.connect()
}

// Close closes the connection and cleans up any resources
func (c *APIConnection) Close() error {
	if c.client != nil {
		// Clean up any transport resources
		if transport, ok := c.client.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
	}
	return nil
}

// IsInitialized returns whether the connection has been fully initialized
func (c *APIConnection) IsInitialized() bool {
	return c.initialized
}
