// Copyright (c) 2021 StreamNative, Inc.. All Rights Reserved.

package controllers

import (
	"context"
	"fmt"

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

	"github.com/go-logr/logr"
	pulsarv1alpha1 "github.com/streamnative/resources-operator/api/v1alpha1"
	"github.com/streamnative/resources-operator/pkg/admin"
	"github.com/streamnative/resources-operator/pkg/connection"
	"github.com/streamnative/resources-operator/pkg/utils"
)

// PulsarConnectionReconciler reconciles a PulsarConnection object
type PulsarConnectionReconciler struct {
	client.Client
	Scheme             *runtime.Scheme
	Log                logr.Logger
	Recorder           record.EventRecorder
	PulsarAdminCreator admin.PulsarAdminCreator
}

//+kubebuilder:rbac:groups=pulsar.streamnative.io,resources=pulsarconnections,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pulsar.streamnative.io,resources=pulsarconnections/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pulsar.streamnative.io,resources=pulsarconnections/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PulsarConnection object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *PulsarConnectionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("pulsarconnection", req.NamespacedName)

	pulsarConnection := &pulsarv1alpha1.PulsarConnection{}
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
	if err := mgr.GetCache().IndexField(context.TODO(), &pulsarv1alpha1.PulsarTenant{}, ".spec.connectionRef.name",
		func(object client.Object) []string {
			return []string{
				object.(*pulsarv1alpha1.PulsarTenant).Spec.ConnectionRef.Name,
			}
		}); err != nil {
		return err
	}
	if err := mgr.GetCache().IndexField(context.TODO(), &pulsarv1alpha1.PulsarNamespace{}, ".spec.connectionRef.name",
		func(object client.Object) []string {
			return []string{
				object.(*pulsarv1alpha1.PulsarNamespace).Spec.ConnectionRef.Name,
			}
		}); err != nil {
		return err
	}
	if err := mgr.GetCache().IndexField(context.TODO(), &pulsarv1alpha1.PulsarTopic{}, ".spec.connectionRef.name",
		func(object client.Object) []string {
			return []string{
				object.(*pulsarv1alpha1.PulsarTopic).Spec.ConnectionRef.Name,
			}
		}); err != nil {
		return err
	}
	if err := mgr.GetCache().IndexField(context.TODO(), &pulsarv1alpha1.PulsarPermission{}, ".spec.connectionRef.name",
		func(object client.Object) []string {
			return []string{
				object.(*pulsarv1alpha1.PulsarPermission).Spec.ConnectionRef.Name,
			}
		}); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&pulsarv1alpha1.PulsarConnection{}).
		Watches(&source.Kind{Type: &pulsarv1alpha1.PulsarTenant{}},
			handler.EnqueueRequestsFromMapFunc(ConnectionRefMapper),
			builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Watches(&source.Kind{Type: &pulsarv1alpha1.PulsarNamespace{}},
			handler.EnqueueRequestsFromMapFunc(ConnectionRefMapper),
			builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Watches(&source.Kind{Type: &pulsarv1alpha1.PulsarTopic{}},
			handler.EnqueueRequestsFromMapFunc(ConnectionRefMapper),
			builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Watches(&source.Kind{Type: &pulsarv1alpha1.PulsarPermission{}},
			handler.EnqueueRequestsFromMapFunc(ConnectionRefMapper),
			builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		WithOptions(options).
		Complete(r)
}
