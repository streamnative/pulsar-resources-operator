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
}

const APIKeyFinalizer = "apikey.resource.streamnative.io/finalizer"

//+kubebuilder:rbac:groups=resource.streamnative.io,resources=apikeys,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=apikeys/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=apikeys/finalizers,verbs=update
//+kubebuilder:rbac:groups=resource.streamnative.io,resources=streamnativecloudconnections,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

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
	logger.Info("Attempting to decrypt JWE token", "apiKey", localAPIKey.Name)
	token, err := DecryptToken(privateKey, *cloudAPIKey.Status.EncryptedToken)
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
			logger.Info("APIKey not found. Reconciliation will stop.", "namespace", req.Namespace, "name", req.Name)
			return ctrl.Result{}, nil
		}
		// Other error fetching APIKey
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
			if err := apiKeyClient.DeleteAPIKey(ctx, apiKey); err != nil {
				if !apierrors.IsNotFound(err) {
					r.updateAPIKeyStatus(ctx, apiKey, err, "DeleteFailed",
						fmt.Sprintf("Failed to delete external resources: %v", err))
					return ctrl.Result{}, err
				}
				logger.Info("Remote APIKey already deleted or not found",
					"apiKey", apiKey.Name)
			}
			apiKey.Finalizers = controllers2.RemoveString(apiKey.Finalizers, APIKeyFinalizer)
			if err := r.Update(ctx, apiKey); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	if !controllers2.ContainsString(apiKey.Finalizers, APIKeyFinalizer) {
		apiKey.Finalizers = append(apiKey.Finalizers, APIKeyFinalizer)
		if err := r.Update(ctx, apiKey); err != nil {
			return ctrl.Result{}, err
		}
	}

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
		if apiKey.Spec.EncryptionKey == nil || apiKey.Spec.EncryptionKey.PEM == "" {
			// Generate RSA key pair for token encryption
			privateKey, pKeyErr := crypto.GenerateRSAKeyPair()
			if pKeyErr != nil {
				r.updateAPIKeyStatus(ctx, apiKey, pKeyErr, "KeyGenerationFailed",
					fmt.Sprintf("Failed to generate RSA key pair: %v", pKeyErr))
				return ctrl.Result{}, pKeyErr
			}

			publicKeyPEM, pubKeyErr := crypto.ExportPublicKeyAsPEM(privateKey)
			if pubKeyErr != nil {
				r.updateAPIKeyStatus(ctx, apiKey, pubKeyErr, "KeyExportFailed",
					fmt.Sprintf("Failed to export public key: %v", pubKeyErr))
				return ctrl.Result{}, pubKeyErr
			}

			privateKeyPEM := crypto.ExportPrivateKeyAsPEM(privateKey)
			if err := utils.CreateOrUpdatePrivateKeySecret(ctx, r.Client, apiKey, apiKey.Namespace, apiKey.Name, privateKeyPEM); err != nil {
				r.updateAPIKeyStatus(ctx, apiKey, err, "SecretCreationFailed",
					fmt.Sprintf("Failed to create private key secret: %v", err))
				return ctrl.Result{}, err
			}

			if apiKey.Spec.EncryptionKey == nil {
				apiKey.Spec.EncryptionKey = &resourcev1alpha1.EncryptionKey{}
			}
			apiKey.Spec.EncryptionKey.PEM = publicKeyPEM
		}

		if err := r.Update(ctx, apiKey); err != nil { // Update spec with public key PEM
			logger.Error(err, "Failed to update APIKey spec with public key PEM")
			return ctrl.Result{}, err
		}

		createdAPIKey, createErr := apiKeyClient.CreateAPIKey(ctx, apiKey)
		if createErr != nil {
			r.updateAPIKeyStatus(ctx, apiKey, createErr, "CreateAPIKeyFailed",
				fmt.Sprintf("Failed to create APIKey: %v", createErr))
			return ctrl.Result{}, createErr
		}
		logger.Info("Successfully created remote APIKey", "apiKeyName", apiKey.Name)

		r.syncCloudStatusToLocal(apiKey, createdAPIKey)

		if apiKey.Spec.ExportPlaintextToken != nil && *apiKey.Spec.ExportPlaintextToken {
			if createdAPIKey.Status.EncryptedToken != nil && createdAPIKey.Status.EncryptedToken.JWE != nil {
				r.processEncryptedToken(ctx, apiKey, createdAPIKey)
			} else {
				r.updateAPIKeyStatus(ctx, apiKey, nil, "WaitingForToken",
					"Waiting for encrypted token from remote API server")
			}
		} else {
			r.updateAPIKeyStatus(ctx, apiKey, nil, "Ready", "APIKey created successfully (no token export requested)")
		}
		// The status might have been updated by processEncryptedToken or the else block above.
		// We requeue to ensure the latest status is reflected and to handle any eventual consistency.
		return ctrl.Result{RequeueAfter: requeueInterval}, nil
	}

	r.syncCloudStatusToLocal(apiKey, existingAPIKey)

	// If ExportPlaintextToken is true, ensure token is processed
	if apiKey.Spec.ExportPlaintextToken != nil && *apiKey.Spec.ExportPlaintextToken {
		if existingAPIKey.Status.EncryptedToken != nil && existingAPIKey.Status.EncryptedToken.JWE != nil {
			r.processEncryptedToken(ctx, apiKey, existingAPIKey)
		} else {
			r.updateAPIKeyStatus(ctx, apiKey, nil, "WaitingForToken", "Existing APIKey waiting for encrypted token from remote API server or token not available")
		}
	}

	// Sync spec fields like Revoke and Description if they differ
	// We compare the spec of the local apiKey CR with the spec of the fetched existingAPIKey from the cloud.
	// If they differ, we send the local apiKey (which represents the desired state) to UpdateAPIKey.
	needsRemoteUpdate := false
	if apiKey.Spec.Revoke != existingAPIKey.Spec.Revoke {
		logger.Info("Revoke status differs", "apiKeyName", apiKey.Name, "local", apiKey.Spec.Revoke, "remote", existingAPIKey.Spec.Revoke)
		needsRemoteUpdate = true
	}
	if apiKey.Spec.Description != existingAPIKey.Spec.Description {
		logger.Info("Description differs", "apiKeyName", apiKey.Name, "local", apiKey.Spec.Description, "remote", existingAPIKey.Spec.Description)
		needsRemoteUpdate = true
	}

	if needsRemoteUpdate {
		logger.Info("Updating remote APIKey spec due to detected differences", "apiKeyName", apiKey.Name)
		// Pass the local apiKey CR, as UpdateAPIKey expects *resourcev1alpha1.APIKey
		// It will be converted to cloudapi.APIKey by the client.
		_, err := apiKeyClient.UpdateAPIKey(ctx, apiKey)
		if err != nil {
			r.updateAPIKeyStatus(ctx, apiKey, err, "UpdateFailed", fmt.Sprintf("Failed to update remote APIKey spec: %v", err))
			return ctrl.Result{}, err
		}
		logger.Info("Successfully updated remote APIKey spec", "apiKeyName", apiKey.Name)
	}

	// Always update local status to reflect observed generation and potentially new conditions from processing token.
	r.updateAPIKeyStatus(ctx, apiKey, nil, "Ready", "APIKey synced successfully")
	return ctrl.Result{RequeueAfter: requeueInterval}, nil
}

// syncCloudStatusToLocal copies relevant status fields from the cloud APIKey (cloudapi.APIKey)
// to the local APIKey resource's status (resourcev1alpha1.APIKey).
// This is used to ensure that the local representation reflects the state observed from the cloud.
func (r *APIKeyReconciler) syncCloudStatusToLocal(localAPIKey *resourcev1alpha1.APIKey, cloudAPIKey *cloudapi.APIKey) {
	if cloudAPIKey == nil {
		// If there's no cloud APIKey (e.g., not found), clear relevant local status fields.
		localAPIKey.Status.KeyID = nil
		localAPIKey.Status.IssuedAt = nil
		localAPIKey.Status.ExpiresAt = nil
		localAPIKey.Status.RevokedAt = nil
		localAPIKey.Status.EncryptedToken = nil
		return
	}

	localAPIKey.Status.KeyID = cloudAPIKey.Status.KeyID // Direct assignment for *string

	if cloudAPIKey.Status.IssuedAt != nil {
		localAPIKey.Status.IssuedAt = cloudAPIKey.Status.IssuedAt.DeepCopy()
	} else {
		localAPIKey.Status.IssuedAt = nil
	}

	// Sync ExpiresAt
	if cloudAPIKey.Status.ExpiresAt != nil {
		localAPIKey.Status.ExpiresAt = cloudAPIKey.Status.ExpiresAt.DeepCopy()
	} else {
		localAPIKey.Status.ExpiresAt = nil
	}

	// Sync RevokedAt
	if cloudAPIKey.Status.RevokedAt != nil {
		localAPIKey.Status.RevokedAt = cloudAPIKey.Status.RevokedAt.DeepCopy()
	} else {
		localAPIKey.Status.RevokedAt = nil
	}

	if cloudAPIKey.Status.EncryptedToken != nil {
		if localAPIKey.Status.EncryptedToken == nil {
			localAPIKey.Status.EncryptedToken = &resourcev1alpha1.EncryptedToken{}
		}
		if cloudAPIKey.Status.EncryptedToken.JWE != nil {
			// Create a new string pointer for JWE to ensure no aliasing issues.
			jweCopy := *cloudAPIKey.Status.EncryptedToken.JWE
			localAPIKey.Status.EncryptedToken.JWE = &jweCopy
		} else {
			localAPIKey.Status.EncryptedToken.JWE = nil
		}
	} else {
		localAPIKey.Status.EncryptedToken = nil
	}
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

	newCondition := metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.Now(),
	}

	if err != nil {
		newCondition.Status = metav1.ConditionFalse
		// Do not log error here again if err is not nil, as it's usually logged by the caller
	}

	// Update ready condition
	found := false
	for i, condition := range apiKey.Status.Conditions {
		if condition.Type == "Ready" {
			// Only update LastTransitionTime if Status or Reason or Message changes
			if condition.Status != newCondition.Status || condition.Reason != newCondition.Reason || condition.Message != newCondition.Message {
				apiKey.Status.Conditions[i] = newCondition
			} else {
				// If nothing changed, keep the old LastTransitionTime
				newCondition.LastTransitionTime = condition.LastTransitionTime
				apiKey.Status.Conditions[i] = newCondition
			}
			found = true
			break
		}
	}
	if !found {
		apiKey.Status.Conditions = append(apiKey.Status.Conditions, newCondition)
	}

	// Persist status update
	if statusUpdateErr := r.Status().Update(ctx, apiKey); statusUpdateErr != nil {
		logger.Error(statusUpdateErr, "Failed to update APIKey status")
		// Potentially requeue if status update fails, but be careful of loops
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *APIKeyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&resourcev1alpha1.APIKey{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}

// DecryptToken decrypts an encrypted token using the provided private key
// This function is a utility and not part of the Reconciler methods.
// It's kept here for potential direct use or if it was previously used by other parts of the codebase.
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
