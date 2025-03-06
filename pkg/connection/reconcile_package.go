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
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/go-logr/logr"
	"github.com/streamnative/pulsar-resources-operator/pkg/feature"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	"github.com/streamnative/pulsar-resources-operator/pkg/admin"
	"github.com/streamnative/pulsar-resources-operator/pkg/reconciler"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/azureblob" // Azure support
	_ "gocloud.dev/blob/gcsblob"   // GCS support
	_ "gocloud.dev/blob/s3blob"    // S3 support
)

// PulsarPackageReconciler reconciles a PulsarPackage object
type PulsarPackageReconciler struct {
	conn *PulsarConnectionReconciler
	log  logr.Logger
}

func makePackagesReconciler(r *PulsarConnectionReconciler) reconciler.Interface {
	return &PulsarPackageReconciler{
		conn: r,
		log:  makeSubResourceLog(r, "PulsarPackage"),
	}
}

// Observe checks the updates of object
func (r *PulsarPackageReconciler) Observe(ctx context.Context) error {
	r.log.V(1).Info("Start Observe")

	packageList := &resourcev1alpha1.PulsarPackageList{}
	if err := r.conn.client.List(ctx, packageList, client.InNamespace(r.conn.connection.Namespace),
		client.MatchingFields(map[string]string{
			".spec.connectionRef.name": r.conn.connection.Name,
		})); err != nil {
		return fmt.Errorf("list packages [%w]", err)
	}
	r.log.V(1).Info("Observed package items", "Count", len(packageList.Items))

	r.conn.packages = packageList.Items
	for i := range r.conn.packages {
		if !resourcev1alpha1.IsPulsarResourceReady(&r.conn.packages[i]) {
			r.conn.addUnreadyResource(&r.conn.packages[i])
		}
	}

	r.log.V(1).Info("Observe Done")
	return nil
}

// Reconcile reconciles all topics
func (r *PulsarPackageReconciler) Reconcile(ctx context.Context) error {
	errs := []error{}
	for i := range r.conn.packages {
		pkg := &r.conn.packages[i]
		if err := r.ReconcilePackage(ctx, r.conn.pulsarAdminV3, pkg); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("reconcile pulsar package [%v]", errs)
	}
	return nil
}

// isLatestTag checks if the package URL contains @latest tag
func isLatestTag(packageURL string) bool {
	return strings.Contains(packageURL, "@latest")
}

// shouldSyncPackage determines if the package should be synced based on sync policy and current state
func (r *PulsarPackageReconciler) shouldSyncPackage(
	pkg *resourcev1alpha1.PulsarPackage,
	exists bool,
	currentProps map[string]string,
	newProps *PackageProperties) bool {

	// Determine the effective sync policy
	effectivePolicy := pkg.Spec.SyncPolicy
	if effectivePolicy == "" {
		// Default to Always if @latest tag is used, or IfNotPresent otherwise
		if isLatestTag(pkg.Spec.PackageURL) {
			effectivePolicy = resourcev1alpha1.PulsarPackageSyncAlways
		} else {
			effectivePolicy = resourcev1alpha1.PulsarPackageSyncIfNotPresent
		}
	}

	// If package doesn't exist, we need to sync unless the policy is Never
	if !exists {
		return effectivePolicy != resourcev1alpha1.PulsarPackageSyncNever
	}

	// For Never policy, we don't sync if package exists
	if effectivePolicy == resourcev1alpha1.PulsarPackageSyncNever {
		return false
	}

	// For IfNotPresent policy, we don't sync if package exists
	if effectivePolicy == resourcev1alpha1.PulsarPackageSyncIfNotPresent {
		return false
	}

	// For Always policy or if the package is managed by operator, check the checksum
	if effectivePolicy == resourcev1alpha1.PulsarPackageSyncAlways || IsManagedByOperator(currentProps) {
		currentChecksum := currentProps[PropertyFileChecksum]
		return currentChecksum != newProps.FileChecksum
	}

	return false
}

// ReconcilePackage move the current state of the package closer to the desired state
func (r *PulsarPackageReconciler) ReconcilePackage(ctx context.Context, pulsarAdmin admin.PulsarAdmin,
	pkg *resourcev1alpha1.PulsarPackage) error {
	log := r.log.WithValues("name", pkg.Name, "namespace", pkg.Namespace)
	log.Info("Start Reconcile")

	if !pkg.DeletionTimestamp.IsZero() {
		log.Info("Deleting package", "LifecyclePolicy", pkg.Spec.LifecyclePolicy)

		if pkg.Spec.LifecyclePolicy != resourcev1alpha1.KeepAfterDeletion {
			if err := pulsarAdmin.DeletePulsarPackage(pkg.Spec.PackageURL); err != nil && !admin.IsNotFound(err) {
				log.Error(err, "Failed to delete package")
				return err
			}
		}

		controllerutil.RemoveFinalizer(pkg, resourcev1alpha1.FinalizerName)
		if err := r.conn.client.Update(ctx, pkg); err != nil {
			log.Error(err, "Failed to remove finalizer")
			return err
		}
		return nil
	}

	if pkg.Spec.LifecyclePolicy != resourcev1alpha1.KeepAfterDeletion {
		controllerutil.AddFinalizer(pkg, resourcev1alpha1.FinalizerName)
		if err := r.conn.client.Update(ctx, pkg); err != nil {
			log.Error(err, "Failed to add finalizer")
			return err
		}
	}

	// Check if package exists and get current properties
	exists := false
	var currentProps map[string]string
	if existingPkg, err := pulsarAdmin.GetPulsarPackageMetadata(pkg.Spec.PackageURL); err != nil {
		if !admin.IsNotFound(err) {
			log.Error(err, "Failed to get pulsar package")
			meta.SetStatusCondition(&pkg.Status.Conditions, *NewErrorCondition(pkg.Generation, fmt.Sprintf("failed to get pulsar package: %s", err.Error())))
			return err
		}
	} else {
		exists = true
		currentProps = existingPkg.Properties
	}

	// Skip if resource is ready and no sync is needed
	if resourcev1alpha1.IsPulsarResourceReady(pkg) &&
		!feature.DefaultFeatureGate.Enabled(feature.AlwaysUpdatePulsarResource) {
		// Download and check properties only if sync policy is Always
		if pkg.Spec.SyncPolicy != resourcev1alpha1.PulsarPackageSyncAlways {
			log.Info("Skip reconcile, package resource is ready")
			return nil
		}
	}

	// Download the file
	filePath, err := downloadToTempFile(ctx, pkg.Spec.FileURL)
	if err != nil {
		log.Error(err, "Failed to download the package file")
		meta.SetStatusCondition(&pkg.Status.Conditions, *NewErrorCondition(pkg.Generation, fmt.Sprintf("failed to download package file: %s", err.Error())))
		return err
	}
	defer func() {
		err = os.Remove(filePath)
		if err != nil {
			fmt.Printf("failed to remove temp file: %v", err)
		}
	}()

	// Generate new properties
	newProps, err := GeneratePackageProperties(pkg, filePath, r.conn.connection.Name)
	if err != nil {
		log.Error(err, "Failed to generate package properties")
		meta.SetStatusCondition(&pkg.Status.Conditions, *NewErrorCondition(pkg.Generation, fmt.Sprintf("failed to generate package properties: %s", err.Error())))
		return err
	}

	// Check if we should sync the package
	if !r.shouldSyncPackage(pkg, exists, currentProps, newProps) {
		log.Info("Skip sync based on policy and state", "SyncPolicy", pkg.Spec.SyncPolicy)
		pkg.Status.ObservedGeneration = pkg.Generation
		meta.SetStatusCondition(&pkg.Status.Conditions, *NewReadyCondition(pkg.Generation))
		if err := r.conn.client.Status().Update(ctx, pkg); err != nil {
			log.Error(err, "Failed to update the package status")
			return err
		}
		return nil
	}

	// Merge properties
	mergedProps := MergeProperties(newProps, pkg.Spec.Properties)

	// Apply the package
	if err := pulsarAdmin.ApplyPulsarPackage(pkg.Spec.PackageURL, filePath, pkg.Spec.Description, pkg.Spec.Contact, mergedProps, exists); err != nil {
		meta.SetStatusCondition(&pkg.Status.Conditions, *NewErrorCondition(pkg.Generation, err.Error()))
		log.Error(err, "Failed to apply package")
		if err := r.conn.client.Status().Update(ctx, pkg); err != nil {
			log.Error(err, "Failed to update the package status")
			return nil
		}
		return err
	}

	pkg.Status.ObservedGeneration = pkg.Generation
	meta.SetStatusCondition(&pkg.Status.Conditions, *NewReadyCondition(pkg.Generation))
	if err := r.conn.client.Status().Update(ctx, pkg); err != nil {
		log.Error(err, "Failed to update the package status")
		return err
	}

	log.Info("Successfully reconciled package", "SyncPolicy", pkg.Spec.SyncPolicy)
	return nil
}

// downloadToTempFile downloads file from various sources (http/https/cloud storage) to a temp file
func downloadToTempFile(ctx context.Context, fileURL string) (string, error) {
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %v", err)
	}

	// Create a temporary file
	tmpFile, err := os.CreateTemp("/tmp/packages", "downloaded-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %v", err)
	}
	defer func() {
		err = tmpFile.Close()
		if err != nil {
			fmt.Printf("failed to close temp file: %v", err)
		}
	}()

	switch parsedURL.Scheme {
	case "http", "https":
		return downloadHTTP(fileURL)
	case "file":
		return parsedURL.Path, nil
	case "s3", "gs", "azblob":
		return downloadCloudStorage(ctx, fileURL, tmpFile)
	default:
		return "", fmt.Errorf("unsupported scheme: %s", parsedURL.Scheme)
	}
}

func downloadCloudStorage(ctx context.Context, fileURL string, tmpFile *os.File) (string, error) {
	// Open a connection to the blob
	bucket, err := blob.OpenBucket(ctx, fileURL)
	if err != nil {
		return "", fmt.Errorf("failed to open bucket: %v", err)
	}
	defer bucket.Close()

	// Parse the URL to get the key (path in the bucket)
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %v", err)
	}

	objectKey := parsedURL.Path
	if len(objectKey) > 0 && objectKey[0] == '/' {
		objectKey = objectKey[1:] // Remove the leading slash
	}

	// Get the object
	reader, err := bucket.NewReader(ctx, objectKey, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create reader: %v", err)
	}
	defer reader.Close()

	// Copy the content to the temp file
	if _, err := io.Copy(tmpFile, reader); err != nil {
		return "", fmt.Errorf("failed to copy content to temp file: %v", err)
	}

	return tmpFile.Name(), nil
}

// downloadHTTP downloads file from http/https URL to a temp file
func downloadHTTP(fileURL string) (string, error) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("/tmp", "downloaded-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %v", err)
	}
	defer tmpFile.Close()

	// Make HTTP GET request to the URL
	resp, err := http.Get(fileURL) //nolint:gosec
	if err != nil {
		return "", fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status: %v", resp.Status)
	}

	// Copy the response body to the temp file
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write to temp file: %v", err)
	}

	return tmpFile.Name(), nil
}
