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

package connection

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
)

const (
	// PropertyPrefix is the prefix for all properties managed by the operator
	PropertyPrefix = "pulsarpackages.resource.streamnative.io"

	// Property keys
	PropertyFileChecksum = PropertyPrefix + "/file-checksum"
	PropertyFileSize     = PropertyPrefix + "/file-size"
	PropertyManagedBy    = PropertyPrefix + "/managed-by"
	PropertyResourceNS   = PropertyPrefix + "/resource-namespace"
	PropertyResourceName = PropertyPrefix + "/resource-name"
	PropertyResourceUID  = PropertyPrefix + "/resource-uid"
	PropertyCluster      = PropertyPrefix + "/cluster"

	// Property values
	ValueManagedBy = "pulsar-resources-operator"
)

// PackageProperties contains the managed properties for a PulsarPackage
type PackageProperties struct {
	FileChecksum string
	FileSize     int64
	Namespace    string
	Name         string
	UID          string
	Cluster      string
}

// GeneratePackageProperties creates a new PackageProperties from a PulsarPackage and file
func GeneratePackageProperties(pkg *resourcev1alpha1.PulsarPackage, filePath string, clusterName string) (*PackageProperties, error) {
	// Calculate file checksum
	checksum, size, err := calculateFileChecksumAndSize(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate file properties: %w", err)
	}

	return &PackageProperties{
		FileChecksum: checksum,
		FileSize:     size,
		Namespace:    pkg.Namespace,
		Name:         pkg.Name,
		UID:          string(pkg.UID),
		Cluster:      clusterName,
	}, nil
}

// ToMap converts PackageProperties to a map for Pulsar package properties
func (p *PackageProperties) ToMap() map[string]string {
	return map[string]string{
		PropertyFileChecksum: p.FileChecksum,
		PropertyFileSize:     fmt.Sprintf("%d", p.FileSize),
		PropertyManagedBy:    ValueManagedBy,
		PropertyResourceNS:   p.Namespace,
		PropertyResourceName: p.Name,
		PropertyResourceUID:  p.UID,
		PropertyCluster:      p.Cluster,
	}
}

// FromMap creates PackageProperties from a map of Pulsar package properties
func FromMap(props map[string]string) *PackageProperties {
	if props == nil {
		return nil
	}

	// Convert file size string to int64, default to 0 if invalid
	var fileSize int64
	if sizeStr := props[PropertyFileSize]; sizeStr != "" {
		fmt.Sscanf(sizeStr, "%d", &fileSize)
	}

	return &PackageProperties{
		FileChecksum: props[PropertyFileChecksum],
		FileSize:     fileSize,
		Namespace:    props[PropertyResourceNS],
		Name:         props[PropertyResourceName],
		UID:          props[PropertyResourceUID],
		Cluster:      props[PropertyCluster],
	}
}

// IsManagedByOperator checks if the package is managed by this operator
func IsManagedByOperator(props map[string]string) bool {
	return props[PropertyManagedBy] == ValueManagedBy
}

// calculateFileChecksumAndSize calculates SHA256 checksum and size of a file
func calculateFileChecksumAndSize(filePath string) (string, int64, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	hash := sha256.New()
	size, err := io.Copy(hash, file)
	if err != nil {
		return "", 0, err
	}

	checksum := fmt.Sprintf("sha256:%s", hex.EncodeToString(hash.Sum(nil)))
	return checksum, size, nil
}

// MergeProperties merges package properties with user-defined properties
func MergeProperties(props *PackageProperties, userProps map[string]string) map[string]string {
	result := make(map[string]string)

	// Copy user properties first
	for k, v := range userProps {
		// Skip if user tries to override managed properties
		if _, isManagedProp := props.ToMap()[k]; !isManagedProp {
			result[k] = v
		}
	}

	// Add managed properties (will override any conflicting user properties)
	for k, v := range props.ToMap() {
		result[k] = v
	}

	return result
}
