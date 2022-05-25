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

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	"github.com/streamnative/pulsar-resources-operator/pkg/admin"
	"github.com/streamnative/pulsar-resources-operator/pkg/connection"
	"github.com/streamnative/pulsar-resources-operator/pkg/utils"
)

// PulsarConnectionReconciler reconciles a PulsarConnection object
type PulsarConnectionReconciler struct {
	client.Client
	Scheme             *runtime.Scheme
	Log                logr.Logger
	Recorder           record.EventRecorder
	PulsarAdminCreator admin.PulsarAdminCreator
}

//nolint:lll
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=pulsarconnections,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=pulsarconnections/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=pulsarconnections/finalizers,verbs=update
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=pulsartenants,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=pulsartenants/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=pulsartenants/finalizers,verbs=update
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=pulsarnamespaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=pulsarnamespaces/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=pulsarnamespaces/finalizers,verbs=update
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=pulsartopics,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=pulsartopics/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=pulsartopics/finalizers,verbs=update
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=pulsarpermissions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=pulsarpermissions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=pulsarpermissions/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PulsarConnection object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *PulsarConnectionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("pulsarconnection", req.NamespacedName)

	pulsarConnection := &resourcev1alpha1.PulsarConnection{}
	if err := r.Get(ctx, req.NamespacedName, pulsarConnection); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("get PulsarConnection [%s]: %w", req.String(), err)
	}

	if !utils.IsManaged(pulsarConnection) {
		log.Info("Skipping the object not managed by the controller", "Name", req.String())
		return reconcile.Result{}, nil
	}

	reconciler := connection.MakeReconciler(log, r.Client, r.PulsarAdminCreator, pulsarConnection)
	if err := reconciler.Observe(ctx); err != nil {
		return ctrl.Result{}, err
	}
	if err := reconciler.Reconcile(ctx); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PulsarConnectionReconciler) SetupWithManager(mgr ctrl.Manager, options controller.Options) error {
	if err := mgr.GetCache().IndexField(context.TODO(), &resourcev1alpha1.PulsarTenant{}, ".spec.connectionRef.name",
		func(object client.Object) []string {
			return []string{
				object.(*resourcev1alpha1.PulsarTenant).Spec.ConnectionRef.Name,
			}
		}); err != nil {
		return err
	}
	if err := mgr.GetCache().IndexField(context.TODO(), &resourcev1alpha1.PulsarNamespace{}, ".spec.connectionRef.name",
		func(object client.Object) []string {
			return []string{
				object.(*resourcev1alpha1.PulsarNamespace).Spec.ConnectionRef.Name,
			}
		}); err != nil {
		return err
	}
	if err := mgr.GetCache().IndexField(context.TODO(), &resourcev1alpha1.PulsarTopic{}, ".spec.connectionRef.name",
		func(object client.Object) []string {
			return []string{
				object.(*resourcev1alpha1.PulsarTopic).Spec.ConnectionRef.Name,
			}
		}); err != nil {
		return err
	}
	if err := mgr.GetCache().IndexField(context.TODO(), &resourcev1alpha1.PulsarPermission{}, ".spec.connectionRef.name",
		func(object client.Object) []string {
			return []string{
				object.(*resourcev1alpha1.PulsarPermission).Spec.ConnectionRef.Name,
			}
		}); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&resourcev1alpha1.PulsarConnection{}).
		Watches(&source.Kind{Type: &resourcev1alpha1.PulsarTenant{}},
			handler.EnqueueRequestsFromMapFunc(ConnectionRefMapper),
			builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Watches(&source.Kind{Type: &resourcev1alpha1.PulsarNamespace{}},
			handler.EnqueueRequestsFromMapFunc(ConnectionRefMapper),
			builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Watches(&source.Kind{Type: &resourcev1alpha1.PulsarTopic{}},
			handler.EnqueueRequestsFromMapFunc(ConnectionRefMapper),
			builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Watches(&source.Kind{Type: &resourcev1alpha1.PulsarPermission{}},
			handler.EnqueueRequestsFromMapFunc(ConnectionRefMapper),
			builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		WithOptions(options).
		Complete(r)
}
