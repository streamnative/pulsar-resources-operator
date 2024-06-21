// Copyright 2022 StreamNative
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
	"errors"
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
	for i := range r.conn.packages {
		pkg := &r.conn.packages[i]
		if err := r.ReconcilePackage(ctx, r.conn.pulsarAdminV3, pkg); err != nil {
			return fmt.Errorf("reconcile pulsar package [%w]", err)
		}
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

	filePath, err := createTmpFile(pkg.Spec.FileURL)
	if err != nil {
		log.Error(err, "Failed to download the package file")
		return err
	}
	defer os.Remove(filePath)

	if err := pulsarAdmin.ApplyPulsarPackage(pkg.Spec.PackageURL, filePath, pkg.Spec.Description, pkg.Spec.Contact, pkg.Spec.Properties, pkg.Status.ObservedGeneration > 1); err != nil {
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

func createTmpFile(fileURL string) (string, error) {
	// get the file from the url and save it to a temp file
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %v", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", errors.New("URL must be HTTP or HTTPS")
	}

	// create a temporary file
	tmpFile, err := os.CreateTemp("/tmp", "downloaded-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %v", err)
	}
	defer tmpFile.Close()

	// make HTTP GET request to the URL
	resp, err := http.Get(fileURL) //nolint:gosec
	if err != nil {
		return "", fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status: %v", resp.Status)
	}

	// copy the response body to the temp file
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write to temp file: %v", err)
	}

	return tmpFile.Name(), nil
}
