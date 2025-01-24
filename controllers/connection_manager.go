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

package controllers

import (
	"context"
	"fmt"
	"sync"

	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	controllers2 "github.com/streamnative/pulsar-resources-operator/pkg/streamnativecloud"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ConnectionManager manages API connections
type ConnectionManager struct {
	client.Client
	mu          sync.RWMutex
	connections map[string]*controllers2.APIConnection
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(client client.Client) *ConnectionManager {
	return &ConnectionManager{
		Client:      client,
		connections: make(map[string]*controllers2.APIConnection),
	}
}

// GetOrCreateConnection gets or creates a connection
func (m *ConnectionManager) GetOrCreateConnection(
	apiConn *resourcev1alpha1.StreamNativeCloudConnection,
	creds *resourcev1alpha1.ServiceAccountCredentials,
) (*controllers2.APIConnection, error) {
	m.mu.RLock()
	conn, exists := m.connections[apiConn.Name]
	m.mu.RUnlock()

	if exists {
		// Check if connection needs update
		if conn.NeedsUpdate(apiConn, creds) {
			if err := conn.Update(apiConn, creds); err != nil {
				return nil, fmt.Errorf("failed to update connection: %w", err)
			}
		}

		// If connection exists but not initialized, return it with a special error
		if !conn.IsInitialized() {
			err := conn.Test(context.Background())
			if err != nil {
				return conn, &NotInitializedError{message: fmt.Sprintf("connection not fully initialized: %s", err)}
			}
		}
		return conn, nil
	}

	// Create new connection
	newConn, err := controllers2.NewAPIConnection(apiConn, creds)
	if err != nil {
		// Store uninitialized connection for future retry
		m.mu.Lock()
		m.connections[apiConn.Name] = newConn
		m.mu.Unlock()
		return newConn, &NotInitializedError{message: err.Error()}
	}

	// Store connection
	m.mu.Lock()
	m.connections[apiConn.Name] = newConn
	m.mu.Unlock()

	return newConn, nil
}

// CloseConnection closes and removes a connection
func (m *ConnectionManager) CloseConnection(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if conn, exists := m.connections[name]; exists {
		if err := conn.Close(); err != nil {
			return err
		}
		delete(m.connections, name)
	}
	return nil
}

// Close closes all connections
func (m *ConnectionManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, conn := range m.connections {
		if err := conn.Close(); err != nil {
			return err
		}
		delete(m.connections, name)
	}
	return nil
}

// NotInitializedError represents an error when the connection is not fully initialized
type NotInitializedError struct {
	message string
}

func (e *NotInitializedError) Error() string {
	return e.message
}
