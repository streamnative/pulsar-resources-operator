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

		// TODO use otelcontroller until kube-instrumentation upgrade controller-runtime version to newer
		controllerutil.RemoveFinalizer(pkg, resourcev1alpha1.FinalizerName)
		if err := r.conn.client.Update(ctx, pkg); err != nil {
			log.Error(err, "Failed to remove finalizer")
			return err
		}
		return nil
	}

	if pkg.Spec.LifecyclePolicy != resourcev1alpha1.KeepAfterDeletion {
		// TODO use otelcontroller until kube-instrumentation upgrade controller-runtime version to newer
		controllerutil.AddFinalizer(pkg, resourcev1alpha1.FinalizerName)
		if err := r.conn.client.Update(ctx, pkg); err != nil {
			log.Error(err, "Failed to add finalizer")
			return err
		}
	}

	if resourcev1alpha1.IsPulsarResourceReady(pkg) &&
		!feature.DefaultFeatureGate.Enabled(feature.AlwaysUpdatePulsarResource) {
		log.Info("Skip reconcile, package resource is ready")
		return nil
	}

	filePath, err := downloadToTempFile(ctx, pkg.Spec.FileURL)
	if err != nil {
		log.Error(err, "Failed to download the package file")
		return err
	}
	defer os.Remove(filePath)

	updated := false
	if exist, err := pulsarAdmin.CheckPulsarPackageExist(pkg.Spec.PackageURL); err != nil {
		log.Error(err, "Failed to check pulsar package existence")
		meta.SetStatusCondition(&pkg.Status.Conditions, *NewErrorCondition(pkg.Generation, fmt.Sprintf("failed to check pulsar package existence: %s", err.Error())))
		return err
	} else if exist {
		updated = true
	}

	if err := pulsarAdmin.ApplyPulsarPackage(pkg.Spec.PackageURL, filePath, pkg.Spec.Description, pkg.Spec.Contact, pkg.Spec.Properties, updated); err != nil {
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
	return nil
}

// downloadToTempFile downloads file from various sources (http/https/cloud storage) to a temp file
func downloadToTempFile(ctx context.Context, fileURL string) (string, error) {
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %v", err)
	}

	// Create a temporary file
	tmpFile, err := os.CreateTemp("/tmp", "downloaded-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %v", err)
	}
	defer tmpFile.Close()

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

	// Get the object
	reader, err := bucket.NewReader(ctx, parsedURL.Path, nil)
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

func isSupportedScheme(scheme string) bool {
	supportedSchemes := map[string]bool{
		"http":   true,
		"https":  true,
		"file":   true,
		"s3":     true,
		"gs":     true,
		"azblob": true,
	}
	return supportedSchemes[scheme]
}
