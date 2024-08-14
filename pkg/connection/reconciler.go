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
	"fmt"

	"github.com/apache/pulsar-client-go/pulsaradmin/pkg/admin/config"

	"github.com/go-logr/logr"
	"github.com/streamnative/pulsar-resources-operator/pkg/utils"
	authenticationv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	"github.com/streamnative/pulsar-resources-operator/pkg/admin"
	"github.com/streamnative/pulsar-resources-operator/pkg/reconciler"
)

// PulsarConnectionReconciler reconciles a PulsarConnection object
type PulsarConnectionReconciler struct {
	connection       *resourcev1alpha1.PulsarConnection
	log              logr.Logger
	client           client.Client
	creator          admin.PulsarAdminCreator
	tenants          []resourcev1alpha1.PulsarTenant
	namespaces       []resourcev1alpha1.PulsarNamespace
	topics           []resourcev1alpha1.PulsarTopic
	permissions      []resourcev1alpha1.PulsarPermission
	geoReplications  []resourcev1alpha1.PulsarGeoReplication
	packages         []resourcev1alpha1.PulsarPackage
	sinks            []resourcev1alpha1.PulsarSink
	sources          []resourcev1alpha1.PulsarSource
	functions        []resourcev1alpha1.PulsarFunction
	unreadyResources []string

	pulsarAdmin   admin.PulsarAdmin
	pulsarAdminV3 admin.PulsarAdmin
	reconcilers   []reconciler.Interface
}

var _ reconciler.Interface = &PulsarConnectionReconciler{}

// MakeReconciler creates resource reconcilers
func MakeReconciler(log logr.Logger, k8sClient client.Client, creator admin.PulsarAdminCreator,
	connection *resourcev1alpha1.PulsarConnection) reconciler.Interface {
	r := &PulsarConnectionReconciler{
		log:        log,
		connection: connection,
		creator:    creator,
		client:     k8sClient,
	}
	r.reconcilers = []reconciler.Interface{
		makeGeoReplicationReconciler(r),
		makeTenantsReconciler(r),
		makeNamespacesReconciler(r),
		makeTopicsReconciler(r),
		makePermissionsReconciler(r),
		makePackagesReconciler(r),
		makeFunctionsReconciler(r),
		makeSinksReconciler(r),
		makeSourcesReconciler(r),
	}
	return r
}

func makeSubResourceLog(r *PulsarConnectionReconciler, name string) logr.Logger {
	return r.log.WithName(name).WithValues("connectionRef",
		fmt.Sprintf("%s/%s", r.connection.Namespace, r.connection.Name))
}

// Observe checks the updates of object
func (r *PulsarConnectionReconciler) Observe(ctx context.Context) error {
	r.log.Info("Start PulsarConnectionReconciler Observe")
	for _, reconciler := range r.reconcilers {
		if err := reconciler.Observe(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Reconcile reconciles all resources
func (r *PulsarConnectionReconciler) Reconcile(ctx context.Context) error {
	var err error
	log := r.log.WithValues("name", r.connection.Name, "namespace", r.connection.Namespace)

	if !r.hasUnreadyResource() {
		if !r.connection.DeletionTimestamp.IsZero() {
			if len(r.tenants) == 0 && len(r.namespaces) == 0 && len(r.topics) == 0 && len(r.geoReplications) == 0 {
				// keep the connection until all resources has been removed

				// TODO use otelcontroller until kube-instrumentation upgrade controller-runtime version to newer
				controllerutil.RemoveFinalizer(r.connection, resourcev1alpha1.FinalizerName)
				if err := r.client.Update(ctx, r.connection); err != nil {
					return err
				}
			} else {
				r.log.Info("There are still remaining resources before deleting the connection", "tenants", len(r.tenants), "namespaces",
					len(r.namespaces), "topics", len(r.topics), "geo", len(r.geoReplications))
				msg := fmt.Sprintf("remaining resources: tenants [%d], namespaces [%d], topics [%d], geoReplications [%d]",
					len(r.tenants), len(r.namespaces), len(r.topics), len(r.geoReplications))
				meta.SetStatusCondition(&r.connection.Status.Conditions, *NewErrorCondition(r.connection.Generation, msg))
				if err := r.client.Status().Update(ctx, r.connection); err != nil {
					return err
				}
				return nil
			}
			return nil
		}
		log.Info("Doesn't have associated unready resource, reconcile completed")
		return nil
	}
	log.Info("Reconciling pulsar resources", "resources", r.unreadyResources)

	if r.connection.Spec.AdminServiceURL == "" && r.connection.Spec.AdminServiceSecureURL != "" {
		r.connection.Spec.AdminServiceURL = r.connection.Spec.AdminServiceSecureURL
	}

	// TODO use otelcontroller until kube-instrumentation upgrade controller-runtime version to newer
	controllerutil.AddFinalizer(r.connection, resourcev1alpha1.FinalizerName)
	if err := r.client.Update(ctx, r.connection); err != nil {
		return err
	}

	pulsarConfig, err := r.MakePulsarAdminConfig(ctx)
	if err != nil {
		return err
	}
	r.pulsarAdmin, err = r.creator(*pulsarConfig)
	if err != nil {
		log.Error(err, "create pulsar admin")
		return err
	}
	defer func() {
		if err := r.pulsarAdmin.Close(); err != nil {
			log.Error(err, "close pulsar admin")
		}
		r.pulsarAdmin = nil
	}()

	pulsarConfig, err = r.MakePulsarAdminConfigWithAPIVersion(ctx, config.V3)
	if err != nil {
		return err
	}
	r.pulsarAdminV3, err = r.creator(*pulsarConfig)
	if err != nil {
		log.Error(err, "create pulsar admin v3")
		return err
	}
	defer func() {
		if err := r.pulsarAdminV3.Close(); err != nil {
			log.Error(err, "close pulsar admin v3")
		}
		r.pulsarAdminV3 = nil
	}()

	if r.connection.DeletionTimestamp.IsZero() {
		for _, reconciler := range r.reconcilers {
			if err = reconciler.Reconcile(ctx); err != nil {
				return err
			}
		}
	} else {
		// delete children resources first
		for i := len(r.reconcilers) - 1; i >= 0; i-- {
			if err = r.reconcilers[i].Reconcile(ctx); err != nil {
				return err
			}
		}
	}

	auth := r.connection.Spec.Authentication
	if auth != nil && auth.Token != nil && auth.Token.SecretRef != nil {
		// calculate secret key hash
		secret := &corev1.Secret{}
		if err := r.client.Get(ctx, types.NamespacedName{
			Namespace: r.connection.Namespace,
			Name:      auth.Token.SecretRef.Name,
		}, secret); err != nil {
			return err
		}
		hash, err := utils.CalculateSecretKeyMd5(secret, auth.Token.SecretRef.Key)
		if err != nil {
			return err
		}
		r.connection.Status.SecretKeyHash = hash
	}
	r.connection.Status.ObservedGeneration = r.connection.Generation
	meta.SetStatusCondition(&r.connection.Status.Conditions, *NewReadyCondition(r.connection.Generation))
	if err := r.client.Status().Update(ctx, r.connection); err != nil {
		return err
	}

	return nil
}

func (r *PulsarConnectionReconciler) hasUnreadyResource() bool {
	return len(r.unreadyResources) > 0
}

func (r *PulsarConnectionReconciler) addUnreadyResource(obj reconciler.Object) {
	if len(r.unreadyResources) == 30 {
		// avoid add too many unready resources
		r.unreadyResources = append(r.unreadyResources, "...")
	}
	if len(r.unreadyResources) > 30 {
		return
	}
	r.unreadyResources = append(r.unreadyResources, fmt.Sprintf("%s:%s/%s",
		obj.GetObjectKind().GroupVersionKind().Kind, obj.GetNamespace(), obj.GetName()))
}

// NewErrorCondition create a condition with error
func NewErrorCondition(generation int64, msg string) *metav1.Condition {
	return &metav1.Condition{
		Type:               resourcev1alpha1.ConditionReady,
		Status:             metav1.ConditionFalse,
		ObservedGeneration: generation,
		Reason:             "ReconcileError",
		Message:            msg,
	}
}

// GetValue get the authentication token value or secret
func GetValue(ctx context.Context, k8sClient client.Client, namespace string,
	vRef *resourcev1alpha1.ValueOrSecretRef) (*string, error) {
	if value := vRef.Value; value != nil {
		return value, nil
	} else if ref := vRef.SecretRef; ref != nil {
		secret := &corev1.Secret{}
		if err := k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: ref.Name}, secret); err != nil {
			return nil, err
		}
		if value, exists := secret.Data[ref.Key]; exists {
			return pointer.StringPtr(string(value)), nil
		}
	}
	return nil, nil
}

// GetServiceAccountToken obtains a service account token from a service account
func GetServiceAccountToken(ctx context.Context, k8sClient client.Client, namespace string, serviceAccount string, brokerAudience string) (string, error) {
	sa := &corev1.ServiceAccount{}
	sa.ObjectMeta = metav1.ObjectMeta{
		Name:      serviceAccount,
		Namespace: namespace,
	}

	request := &authenticationv1.TokenRequest{}
	request.Spec.Audiences = []string{brokerAudience}
	request.Spec.ExpirationSeconds = pointer.Int64(3600)

	if err := k8sClient.SubResource("token").Create(ctx, sa, request); err != nil {
		return "", err
	}
	return request.Status.Token, nil
}

// MakePulsarAdminConfig create pulsar admin configuration
func (r *PulsarConnectionReconciler) MakePulsarAdminConfig(ctx context.Context) (*admin.PulsarAdminConfig, error) {
	return MakePulsarAdminConfig(ctx, r.connection, r.client)
}

// MakePulsarAdminConfigWithAPIVersion create pulsar admin configuration with api version
func (r *PulsarConnectionReconciler) MakePulsarAdminConfigWithAPIVersion(ctx context.Context, ver config.APIVersion) (*admin.PulsarAdminConfig, error) {
	c, e := MakePulsarAdminConfig(ctx, r.connection, r.client)
	if e != nil {
		return nil, e
	}
	c.PulsarAPIVersion = &ver
	return c, nil
}

// MakePulsarAdminConfig create pulsar admin configuration
func MakePulsarAdminConfig(ctx context.Context, connection *resourcev1alpha1.PulsarConnection,
	k8sClient client.Client) (*admin.PulsarAdminConfig, error) {
	if connection.Spec.AdminServiceURL == "" && connection.Spec.AdminServiceSecureURL == "" {
		return nil, fmt.Errorf("adminServiceURL or adminServiceSecureURL must not be empty")
	}
	cfg := admin.PulsarAdminConfig{
		WebServiceURL: connection.Spec.AdminServiceURL,
	}
	hasAuth := false
	if authn := connection.Spec.Authentication; authn != nil {
		if token := authn.Token; token != nil {
			value, err := GetValue(ctx, k8sClient, connection.Namespace, token)
			if err != nil {
				return nil, err
			}
			if value != nil {
				cfg.Token = *value
				hasAuth = true
			}
		}
		if serviceAccount := authn.ServiceAccount; !hasAuth && serviceAccount != nil {
			// service account auth
			cfg.TokenSupplier = func() (string, error) {
				return GetServiceAccountToken(ctx, k8sClient, connection.Namespace, serviceAccount.Name, serviceAccount.Audience)
			}
		}
		if oauth2 := authn.OAuth2; !hasAuth && oauth2 != nil {
			cfg.IssuerEndpoint = oauth2.IssuerEndpoint
			cfg.ClientID = oauth2.ClientID
			cfg.Audience = oauth2.Audience
			cfg.Scope = oauth2.Scope
			value, err := GetValue(ctx, k8sClient, connection.Namespace, &oauth2.Key)
			if err != nil {
				return nil, err
			}
			if value != nil {
				cfg.Key = *value
			}
		}
	}
	return &cfg, nil
}

// NewReadyCondition make condition with ready info
func NewReadyCondition(generation int64) *metav1.Condition {
	return &metav1.Condition{
		Type:               resourcev1alpha1.ConditionReady,
		Status:             metav1.ConditionTrue,
		ObservedGeneration: generation,
		Reason:             "Reconciled",
		Message:            "",
	}
}
