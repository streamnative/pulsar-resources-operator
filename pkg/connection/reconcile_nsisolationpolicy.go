// Copyright 2024 StreamNative
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

	utils2 "github.com/apache/pulsar-client-go/pulsaradmin/pkg/utils"
	"github.com/go-logr/logr"
	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	"github.com/streamnative/pulsar-resources-operator/pkg/admin"
	"github.com/streamnative/pulsar-resources-operator/pkg/feature"
	"github.com/streamnative/pulsar-resources-operator/pkg/reconciler"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// PulsarNSIsolationPolicyReconciler reconciles a PulsarNSIsolationPolicy object
type PulsarNSIsolationPolicyReconciler struct {
	conn *PulsarConnectionReconciler
	log  logr.Logger
}

func makeNSIsolationPoliciesReconciler(r *PulsarConnectionReconciler) reconciler.Interface {
	return &PulsarNSIsolationPolicyReconciler{
		conn: r,
		log:  makeSubResourceLog(r, "PulsarNSIsolationPolicy"),
	}
}

// Observe checks the updates of object
func (r *PulsarNSIsolationPolicyReconciler) Observe(ctx context.Context) error {
	r.log.V(1).Info("Start Observe")

	nsIsolationPolicyList := &resourcev1alpha1.PulsarNSIsolationPolicyList{}
	if err := r.conn.client.List(ctx, nsIsolationPolicyList, client.InNamespace(r.conn.connection.Namespace),
		client.MatchingFields(map[string]string{
			".spec.connectionRef.name": r.conn.connection.Name,
		})); err != nil {
		return fmt.Errorf("list ns-isolation-policies [%w]", err)
	}
	r.log.V(1).Info("Observed ns-isolation-policy items", "Count", len(nsIsolationPolicyList.Items))

	r.conn.nsIsolationPolicies = nsIsolationPolicyList.Items
	for i := range r.conn.nsIsolationPolicies {
		if !resourcev1alpha1.IsPulsarResourceReady(&r.conn.nsIsolationPolicies[i]) {
			r.conn.addUnreadyResource(&r.conn.nsIsolationPolicies[i])
		}
	}

	r.log.V(1).Info("Observe Done")
	return nil
}

// Reconcile reconciles all ns-isolation-policies
func (r *PulsarNSIsolationPolicyReconciler) Reconcile(ctx context.Context) error {
	errs := []error{}
	for i := range r.conn.nsIsolationPolicies {
		policy := &r.conn.nsIsolationPolicies[i]
		if err := r.ReconcilePolicy(ctx, r.conn.pulsarAdmin, policy); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("reconcile ns-isolation-policies error: [%v]", errs)
	}
	return nil
}

// ReconcilePolicy move the current state of the ns-isolation-policy closer to the desired state
func (r *PulsarNSIsolationPolicyReconciler) ReconcilePolicy(ctx context.Context, pulsarAdmin admin.PulsarAdmin,
	policy *resourcev1alpha1.PulsarNSIsolationPolicy) error {
	log := r.log.WithValues("name", policy.Name, "namespace", policy.Namespace)
	log.Info("Start Reconcile")

	if !policy.DeletionTimestamp.IsZero() {
		log.Info("Deleting ns-isolation-policy")

		if err := pulsarAdmin.DeleteNSIsolationPolicy(policy.Spec.Name, policy.Spec.Cluster); err != nil {
			if admin.IsNoSuchHostError(err) {
				log.Info("Pulsar cluster has been deleted")
			} else {
				log.Error(err, "Failed to delete ns-isolation-policy")
				meta.SetStatusCondition(&policy.Status.Conditions, *NewErrorCondition(policy.Generation, err.Error()))
				if err := r.conn.client.Status().Update(ctx, policy); err != nil {
					log.Error(err, "Failed to update the ns-isolation-policy status")
					return err
				}
				return err
			}
		}

		// TODO use otelcontroller until kube-instrumentation upgrade controller-runtime version to newer
		controllerutil.RemoveFinalizer(policy, resourcev1alpha1.FinalizerName)
		if err := r.conn.client.Update(ctx, policy); err != nil {
			log.Error(err, "Failed to remove finalizer")
			return err
		}

		return nil
	}

	// TODO use otelcontroller until kube-instrumentation upgrade controller-runtime version to newer
	controllerutil.AddFinalizer(policy, resourcev1alpha1.FinalizerName)
	if err := r.conn.client.Update(ctx, policy); err != nil {
		log.Error(err, "Failed to add finalizer")
		return err
	}

	if resourcev1alpha1.IsPulsarResourceReady(policy) &&
		!feature.DefaultFeatureGate.Enabled(feature.AlwaysUpdatePulsarResource) {
		log.Info("Skip reconcile, ns-isolation-policy resource is ready")
		return nil
	}

	if err := pulsarAdmin.CreateNSIsolationPolicy(policy.Spec.Name, policy.Spec.Cluster, utils2.NamespaceIsolationData{
		Namespaces: policy.Spec.Namespaces,
		Primary:    policy.Spec.Primary,
		Secondary:  policy.Spec.Secondary,
		AutoFailoverPolicy: utils2.AutoFailoverPolicyData{
			PolicyType: convertPolicyType(policy.Spec.AutoFailoverPolicyType),
			Parameters: policy.Spec.AutoFailoverPolicyParams,
		},
	}); err != nil {
		meta.SetStatusCondition(&policy.Status.Conditions, *NewErrorCondition(policy.Generation, err.Error()))
		log.Error(err, "Failed to apply ns isolation policy")
		if err := r.conn.client.Status().Update(ctx, policy); err != nil {
			log.Error(err, "Failed to update the ns isolation policy")
			return err
		}
		return err
	}

	policy.Status.ObservedGeneration = policy.Generation
	meta.SetStatusCondition(&policy.Status.Conditions, *NewReadyCondition(policy.Generation))
	if err := r.conn.client.Status().Update(ctx, policy); err != nil {
		log.Error(err, "Failed to update the ns-isolation-policy status")
		return err
	}

	return nil
}

func convertPolicyType(policyType resourcev1alpha1.AutoFailoverPolicyType) utils2.AutoFailoverPolicyType {
	switch policyType {
	case resourcev1alpha1.MinAvailable:
		return utils2.MinAvailable
	default:
		return ""
	}
}
