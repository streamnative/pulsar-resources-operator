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
	"crypto/rsa"
	"fmt"
	"sync"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwe"
	"github.com/pkg/errors"
	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	"github.com/streamnative/pulsar-resources-operator/pkg/crypto"
	controllers2 "github.com/streamnative/pulsar-resources-operator/pkg/streamnativecloud"
	"github.com/streamnative/pulsar-resources-operator/pkg/utils"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	cloudapi "github.com/streamnative/pulsar-resources-operator/pkg/streamnativecloud/apis/cloud/v1alpha1"
)

// APIKeyReconciler reconciles a StreamNative Cloud APIKey object
type APIKeyReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	ConnectionManager *ConnectionManager
	// watcherMap stores active watchers for APIKeys
	watcherMap map[types.NamespacedName]watch.Interface
	// watcherMutex protects watcherMap
	watcherMutex sync.RWMutex
}

const APIKeyFinalizer = "apikey.resource.streamnative.io/finalizer"

//+kubebuilder:rbac:groups=resource.streamnative.io,resources=apikeys,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=apikeys/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=apikeys/finalizers,verbs=update
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=streamnativecloudconnections,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

// handleWatchEvents processes events from the watch interface
func (r *APIKeyReconciler) handleWatchEvents(ctx context.Context, namespacedName types.NamespacedName, watcher watch.Interface) {
	logger := log.FromContext(ctx)
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-watcher.ResultChan():
			if !ok {
				logger.Info("Watch channel closed", "namespace", namespacedName.Namespace, "name", namespacedName.Name)
				// Remove the watcher from the map
				r.watcherMutex.Lock()
				delete(r.watcherMap, namespacedName)
				r.watcherMutex.Unlock()
				return
			}

			logger.Info("Received watch event",
				"apiKey", namespacedName.Name,
				"eventType", event.Type)

			if event.Type == watch.Modified {
				// Check if the object is an APIKey
				cloudAPIKey, ok := event.Object.(*cloudapi.APIKey)
				if !ok {
					logger.Error(fmt.Errorf("unexpected object type"), "Failed to convert object to APIKey")
					continue
				}

				// Log token and encryption status
				hasEncryptedToken := cloudAPIKey.Status.EncryptedToken != nil
				hasJWE := hasEncryptedToken && cloudAPIKey.Status.EncryptedToken.JWE != nil
				logger.Info("Received APIKey update",
					"apiKey", namespacedName.Name,
					"hasEncryptedToken", hasEncryptedToken,
					"hasJWE", hasJWE)

				// Get the local APIKey
				localAPIKey := &resourcev1alpha1.APIKey{}
				if err := r.Get(ctx, namespacedName, localAPIKey); err != nil {
					logger.Error(err, "Failed to get local APIKey")
					continue
				}

				// Check remote status and handle abnormal situations
				if isRemoteStatusAbnormal(cloudAPIKey) {
					logger.Info("Detected abnormal remote APIKey status",
						"apiKey", namespacedName.Name,
						"hasEncryptedToken", hasEncryptedToken,
						"hasJWE", hasJWE)
					r.updateAPIKeyStatus(ctx, localAPIKey,
						fmt.Errorf("remote API server returned incomplete or invalid status"),
						"RemoteStatusAbnormal",
						"Remote API server returned incomplete or invalid status for the APIKey")
					continue
				}

				if localAPIKey.Spec.ExportPlaintextToken != nil && *localAPIKey.Spec.ExportPlaintextToken {
					// Process encrypted token and update Secret if needed
					if cloudAPIKey.Status.EncryptedToken != nil && cloudAPIKey.Status.EncryptedToken.JWE != nil {
						logger.Info("Found encrypted token in watch event, processing", "apiKey", namespacedName.Name)
						r.processEncryptedToken(ctx, localAPIKey, cloudAPIKey)
					} else {
						logger.Info("No encrypted token found in watch event", "apiKey", namespacedName.Name)
						// Update status to reflect that we're waiting for token
						r.updateAPIKeyStatus(ctx, localAPIKey, nil, "WaitingForToken",
							"Waiting for encrypted token from remote API server")
					}
				} else {
					r.updateAPIKeyStatus(ctx, localAPIKey, nil, "Ready", "APIKey created successfully")
				}
			}
		}
	}
}

// isRemoteStatusAbnormal checks if the remote API key status is incomplete or invalid
func isRemoteStatusAbnormal(cloudAPIKey *cloudapi.APIKey) bool {
	// Consider status abnormal if:
	// 1. Status is completely empty/nil (very unlikely but possible)
	if cloudAPIKey == nil {
		return true
	}

	// 2. We expect some basic fields to be present in a normal response
	// KeyID should be present in a successful API key
	if cloudAPIKey.Status.KeyID == nil {
		return true
	}

	// 3. For an APIKey that's been processed by the server, we should have either:
	//    - An encrypted token
	//    - Or explicit conditions indicating why token creation failed
	if cloudAPIKey.Status.EncryptedToken == nil && len(cloudAPIKey.Status.Conditions) == 0 {
		return true
	}

	// 4. If there are conditions, check if they indicate failure
	for _, condition := range cloudAPIKey.Status.Conditions {
		if condition.Type == "Ready" && condition.Status == metav1.ConditionFalse {
			return true
		}
	}

	return false
}

// processEncryptedToken handles decrypting the token and storing it in a Secret
func (r *APIKeyReconciler) processEncryptedToken(ctx context.Context, localAPIKey *resourcev1alpha1.APIKey, cloudAPIKey *cloudapi.APIKey) {
	logger := log.FromContext(ctx)

	logger.Info("Processing encrypted token",
		"apiKey", localAPIKey.Name,
		"hasEncryptionKey", localAPIKey.Spec.EncryptionKey != nil,
		"hasJWE", cloudAPIKey.Status.EncryptedToken != nil && cloudAPIKey.Status.EncryptedToken.JWE != nil)

	// Skip if we don't have an encryption key in our local APIKey
	if localAPIKey.Spec.EncryptionKey == nil || localAPIKey.Spec.EncryptionKey.PEM == "" {
		logger.Info("No encryption key found in local APIKey, skipping token processing")
		r.updateAPIKeyStatus(ctx, localAPIKey,
			fmt.Errorf("missing encryption key in local APIKey"),
			"MissingEncryptionKey",
			"Cannot decrypt token: no encryption key found in local APIKey")
		return
	}

	// Get private key from secret
	privateKeyStr, err := utils.GetPrivateKeyFromSecret(ctx, r.Client, localAPIKey.Namespace, localAPIKey.Name)
	if err != nil {
		logger.Error(err, "Failed to get private key from secret")
		r.updateAPIKeyStatus(ctx, localAPIKey,
			err,
			"PrivateKeyError",
			fmt.Sprintf("Failed to get private key from secret: %v", err))
		return
	}
	logger.Info("Retrieved private key from secret", "apiKey", localAPIKey.Name)

	// Import private key
	privateKey, err := crypto.ImportPrivateKeyFromPEM(privateKeyStr)
	if err != nil {
		logger.Error(err, "Failed to import private key")
		r.updateAPIKeyStatus(ctx, localAPIKey,
			err,
			"PrivateKeyImportError",
			fmt.Sprintf("Failed to import private key: %v", err))
		return
	}
	logger.Info("Successfully imported private key", "apiKey", localAPIKey.Name)

	// Decrypt token
	jweToken := *cloudAPIKey.Status.EncryptedToken.JWE
	logger.Info("Attempting to decrypt JWE token", "apiKey", localAPIKey.Name, "jweLength", len(jweToken))
	token, err := crypto.DecryptJWEToken(jweToken, privateKey)
	if err != nil {
		logger.Error(err, "Failed to decrypt token")
		r.updateAPIKeyStatus(ctx, localAPIKey,
			err,
			"TokenDecryptionFailed",
			fmt.Sprintf("Failed to decrypt token: %v", err))
		return
	}
	logger.Info("Successfully decrypted token", "apiKey", localAPIKey.Name)

	// Store token in secret
	var keyID string
	if cloudAPIKey.Status.KeyID != nil {
		keyID = *cloudAPIKey.Status.KeyID
	}

	if err := utils.CreateOrUpdateTokenSecret(ctx, r.Client, localAPIKey, localAPIKey.Namespace, localAPIKey.Name, token, keyID); err != nil {
		logger.Error(err, "Failed to create or update token secret")
		r.updateAPIKeyStatus(ctx, localAPIKey,
			err,
			"SecretUpdateFailed",
			fmt.Sprintf("Failed to create or update token secret: %v", err))
		return
	}

	logger.Info("Successfully processed encrypted token and stored in secret", "apiKey", localAPIKey.Name)

	// Update status to indicate successful token processing
	r.updateAPIKeyStatus(ctx, localAPIKey, nil, "Ready", "APIKey token successfully decrypted and stored in secret")
}

// setupWatch creates a new watcher for an APIKey
func (r *APIKeyReconciler) setupWatch(ctx context.Context, apiKey *resourcev1alpha1.APIKey, apiKeyClient *controllers2.APIKeyClient) error {
	namespacedName := types.NamespacedName{
		Namespace: apiKey.Namespace,
		Name:      apiKey.Name,
	}

	// Check if we already have a watcher
	r.watcherMutex.RLock()
	_, exists := r.watcherMap[namespacedName]
	r.watcherMutex.RUnlock()
	if exists {
		return nil
	}

	// Create new watcher
	watcher, err := apiKeyClient.WatchAPIKey(ctx, apiKey.Name)
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}

	// Store watcher in map
	r.watcherMutex.Lock()
	r.watcherMap[namespacedName] = watcher
	r.watcherMutex.Unlock()

	// Start watching in a new goroutine
	go r.handleWatchEvents(ctx, namespacedName, watcher)
	return nil
}

// Reconcile handles the reconciliation of APIKey objects
func (r *APIKeyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling APIKey", "namespace", req.Namespace, "name", req.Name)

	// Add requeue interval for status sync
	requeueInterval := time.Minute

	// Get the APIKey resource
	apiKey := &resourcev1alpha1.APIKey{}
	if err := r.Get(ctx, req.NamespacedName, apiKey); err != nil {
		if apierrors.IsNotFound(err) {
			// Stop and remove watcher if it exists
			r.watcherMutex.Lock()
			if watcher, exists := r.watcherMap[req.NamespacedName]; exists {
				watcher.Stop()
				delete(r.watcherMap, req.NamespacedName)
			}
			r.watcherMutex.Unlock()
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Get the APIServerConnection
	connection := &resourcev1alpha1.StreamNativeCloudConnection{}
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: req.Namespace,
		Name:      apiKey.Spec.APIServerRef.Name,
	}, connection); err != nil {
		r.updateAPIKeyStatus(ctx, apiKey, err, "ConnectionNotFound",
			fmt.Sprintf("Failed to get APIServerConnection: %v", err))
		return ctrl.Result{}, err
	}

	// Get API connection
	apiConn, err := r.ConnectionManager.GetOrCreateConnection(connection, nil)
	if err != nil {
		// If connection is not initialized, requeue the request
		if _, ok := err.(*NotInitializedError); ok {
			logger.Info("Connection not initialized, requeueing", "error", err.Error())
			return ctrl.Result{Requeue: true}, nil
		}
		r.updateAPIKeyStatus(ctx, apiKey, err, "GetConnectionFailed",
			fmt.Sprintf("Failed to get connection: %v", err))
		return ctrl.Result{}, err
	}

	// Get organization from connection
	organization := connection.Spec.Organization
	if organization == "" {
		err := fmt.Errorf("organization is required but not specified")
		r.updateAPIKeyStatus(ctx, apiKey, err, "ValidationFailed", err.Error())
		return ctrl.Result{}, err
	}

	// Create APIKey client
	apiKeyClient, err := controllers2.NewAPIKeyClient(apiConn, organization)
	if err != nil {
		r.updateAPIKeyStatus(ctx, apiKey, err, "ClientCreationFailed",
			fmt.Sprintf("Failed to create APIKey client: %v", err))
		return ctrl.Result{}, err
	}

	// Handle deletion
	if !apiKey.DeletionTimestamp.IsZero() {
		if controllers2.ContainsString(apiKey.Finalizers, APIKeyFinalizer) {
			// Try to delete remote APIKey
			if err := apiKeyClient.DeleteAPIKey(ctx, apiKey); err != nil {
				if !apierrors.IsNotFound(err) {
					r.updateAPIKeyStatus(ctx, apiKey, err, "DeleteFailed",
						fmt.Sprintf("Failed to delete external resources: %v", err))
					return ctrl.Result{}, err
				}
				// If the resource is already gone, that's fine
				logger.Info("Remote APIKey already deleted or not found",
					"apiKey", apiKey.Name)
			}

			// Remove finalizer after successful deletion
			apiKey.Finalizers = controllers2.RemoveString(apiKey.Finalizers, APIKeyFinalizer)
			if err := r.Update(ctx, apiKey); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer if it doesn't exist
	if !controllers2.ContainsString(apiKey.Finalizers, APIKeyFinalizer) {
		apiKey.Finalizers = append(apiKey.Finalizers, APIKeyFinalizer)
		if err := r.Update(ctx, apiKey); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Check if ApiKey exists
	existingAPIKey, err := apiKeyClient.GetAPIKey(ctx, apiKey.Name)
	if err != nil {
		logger.Info("Failed to get APIKey", "error", err, "existingAPIKey", existingAPIKey)
		if !apierrors.IsNotFound(err) {
			r.updateAPIKeyStatus(ctx, apiKey, err, "GetAPIKeyFailed",
				fmt.Sprintf("Failed to get APIKey: %v", err))
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		existingAPIKey = nil
	}

	if existingAPIKey == nil {
		// Generate RSA key pair for token encryption
		privateKey, err := crypto.GenerateRSAKeyPair()
		if err != nil {
			r.updateAPIKeyStatus(ctx, apiKey, err, "KeyGenerationFailed",
				fmt.Sprintf("Failed to generate RSA key pair: %v", err))
			return ctrl.Result{}, err
		}

		// Export public key in PEM format
		publicKeyPEM, err := crypto.ExportPublicKeyAsPEM(privateKey)
		if err != nil {
			r.updateAPIKeyStatus(ctx, apiKey, err, "KeyExportFailed",
				fmt.Sprintf("Failed to export public key: %v", err))
			return ctrl.Result{}, err
		}

		// Store private key in Secret
		privateKeyPEM := crypto.ExportPrivateKeyAsPEM(privateKey)
		if err := utils.CreateOrUpdatePrivateKeySecret(ctx, r.Client, apiKey, apiKey.Namespace, apiKey.Name, privateKeyPEM); err != nil {
			r.updateAPIKeyStatus(ctx, apiKey, err, "SecretCreationFailed",
				fmt.Sprintf("Failed to create private key secret: %v", err))
			return ctrl.Result{}, err
		}

		// Add encryption key to spec
		if apiKey.Spec.EncryptionKey == nil {
			apiKey.Spec.EncryptionKey = &resourcev1alpha1.EncryptionKey{}
		}
		apiKey.Spec.EncryptionKey.PEM = publicKeyPEM

		// Update APIKey with encryption key
		if err := r.Update(ctx, apiKey); err != nil {
			return ctrl.Result{}, err
		}

		// Create APIKey
		resultAPIKey, err := apiKeyClient.CreateAPIKey(ctx, apiKey)
		if err != nil {
			r.updateAPIKeyStatus(ctx, apiKey, err, "CreateAPIKeyFailed",
				fmt.Sprintf("Failed to create APIKey: %v", err))
			return ctrl.Result{}, err
		}

		if apiKey.Spec.ExportPlaintextToken != nil && *apiKey.Spec.ExportPlaintextToken {
			// Update status with token information
			if resultAPIKey.Status.EncryptedToken != nil && resultAPIKey.Status.EncryptedToken.JWE != nil {
				// Process the encrypted token
				r.processEncryptedToken(ctx, apiKey, resultAPIKey)
			} else {
				r.updateAPIKeyStatus(ctx, apiKey, nil, "WaitingForToken",
					"Waiting for encrypted token from remote API server")
			}
		} else {
			// Update status with token information
			r.updateAPIKeyStatus(ctx, apiKey, nil, "Ready", "APIKey created successfully")
		}

		// Set up watch for APIKey
		if err := r.setupWatch(ctx, apiKey, apiKeyClient); err != nil {
			logger.Error(err, "Failed to set up watch", "apiKey", apiKey.Name)
		}

		// Update status
		r.updateAPIKeyStatus(ctx, apiKey, nil, "Ready", "APIKey created successfully")
		return ctrl.Result{RequeueAfter: requeueInterval}, nil
	}

	// Update revocation status if needed
	if apiKey.Spec.Revoke != existingAPIKey.Spec.Revoke {
		// Create a new local copy with updated values
		updatedAPIKey := apiKey.DeepCopy()
		if err := r.Update(ctx, updatedAPIKey); err != nil {
			r.updateAPIKeyStatus(ctx, apiKey, err, "UpdateFailed",
				fmt.Sprintf("Failed to update APIKey: %v", err))
			return ctrl.Result{}, err
		}
	}

	// Update description if needed
	if apiKey.Spec.Description != existingAPIKey.Spec.Description {
		// Create a new local copy with updated values
		updatedAPIKey := apiKey.DeepCopy()
		if err := r.Update(ctx, updatedAPIKey); err != nil {
			r.updateAPIKeyStatus(ctx, apiKey, err, "UpdateFailed",
				fmt.Sprintf("Failed to update APIKey: %v", err))
			return ctrl.Result{}, err
		}
	}

	// Set up watch for APIKey
	if err := r.setupWatch(ctx, apiKey, apiKeyClient); err != nil {
		logger.Error(err, "Failed to set up watch", "apiKey", apiKey.Name)
	}

	// Update status
	r.updateAPIKeyStatus(ctx, apiKey, nil, "Ready", "APIKey synced successfully")
	return ctrl.Result{RequeueAfter: requeueInterval}, nil
}

// updateAPIKeyStatus updates the status of the ApiKey resource
func (r *APIKeyReconciler) updateAPIKeyStatus(
	ctx context.Context,
	apiKey *resourcev1alpha1.APIKey,
	err error,
	reason string,
	message string,
) {
	logger := log.FromContext(ctx)
	apiKey.Status.ObservedGeneration = apiKey.Generation

	// Set ready condition based on error
	meta := metav1.Now()
	readyCondition := metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: meta,
	}

	if err != nil {
		readyCondition.Status = metav1.ConditionFalse
		logger.Error(err, "APIKey reconciliation failed",
			"reason", reason, "message", message)
	}

	// Update ready condition
	meta = metav1.Now()
	if len(apiKey.Status.Conditions) == 0 {
		apiKey.Status.Conditions = []metav1.Condition{readyCondition}
	} else {
		for i, condition := range apiKey.Status.Conditions {
			if condition.Type == "Ready" {
				// Only update time if status changes
				if condition.Status != readyCondition.Status {
					readyCondition.LastTransitionTime = meta
				} else {
					readyCondition.LastTransitionTime = condition.LastTransitionTime
				}
				apiKey.Status.Conditions[i] = readyCondition
				break
			}
		}
	}

	// Update APIKey status
	if err := r.Status().Update(ctx, apiKey); err != nil {
		logger.Error(err, "Failed to update APIKey status")
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *APIKeyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.watcherMap = make(map[types.NamespacedName]watch.Interface)
	return ctrl.NewControllerManagedBy(mgr).
		For(&resourcev1alpha1.APIKey{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}

// DecryptToken decrypts an encrypted token using the provided private key
func DecryptToken(priv *rsa.PrivateKey, encryptedToken cloudapi.EncryptedToken) (string, error) {
	if encryptedToken.JWE == nil || *encryptedToken.JWE == "" {
		return "", errors.New("encrypted token is empty")
	}

	// Decrypt the JWE token
	token, err := jwe.Decrypt([]byte(*encryptedToken.JWE), jwe.WithKey(jwa.RSA_OAEP, priv))
	if err != nil {
		return "", errors.Wrap(err, "failed to decrypt JWE token")
	}

	return string(token), nil
}
