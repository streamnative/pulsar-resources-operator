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

package utils

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	// PrivateKeySecretSuffix is the suffix for private key secrets
	PrivateKeySecretSuffix = "-private-key"
	// TokenSecretSuffix is the suffix for token secrets
	TokenSecretSuffix = "-token"
	// PrivateKeyField is the key for the private key in the secret
	PrivateKeyField = "private-key"
	// TokenField is the key for the token in the secret
	TokenField = "token"
	// ServiceAccountCredentialsSecretSuffix is the suffix for service account credentials secrets
	ServiceAccountCredentialsSecretSuffix = "-credentials"
	// ServiceAccountCredentialsField is the key for the service account credentials in the secret
	ServiceAccountCredentialsField = "credentials.json"
	// ServiceAccountCredentialsType is the type for service account credentials
	//nolint:gosec
	ServiceAccountCredentialsType = "TYPE_SN_CREDENTIALS_FILE"
)

// CreateOrUpdatePrivateKeySecret creates or updates a secret containing a private key
func CreateOrUpdatePrivateKeySecret(ctx context.Context, c client.Client, owner metav1.Object, namespace, name, privateKey string) error {
	secretName := name + PrivateKeySecretSuffix
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, c, secret, func() error {
		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}
		secret.Data[PrivateKeyField] = []byte(privateKey)
		secret.Type = corev1.SecretTypeOpaque

		// Set owner reference
		return controllerutil.SetControllerReference(owner, secret, c.Scheme())
	})

	if err != nil {
		return fmt.Errorf("failed to create or update private key secret: %w", err)
	}

	return nil
}

// CreateOrUpdateTokenSecret creates or updates a secret containing a token
func CreateOrUpdateTokenSecret(ctx context.Context, c client.Client, owner metav1.Object, namespace, name, token, keyID string) error {
	secretName := name + TokenSecretSuffix
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, c, secret, func() error {
		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}
		secret.Data[TokenField] = []byte(token)
		secret.Type = corev1.SecretTypeOpaque

		// Add labels
		if secret.Labels == nil {
			secret.Labels = make(map[string]string)
		}
		secret.Labels["resources.streamnative.io/apikey"] = name
		if keyID != "" {
			secret.Labels["resources.streamnative.io/key-id"] = keyID
		}

		// Set owner reference
		return controllerutil.SetControllerReference(owner, secret, c.Scheme())
	})

	if err != nil {
		return fmt.Errorf("failed to create or update token secret: %w", err)
	}

	return nil
}

// CreateOrUpdateServiceAccountCredentialsSecret creates or updates a secret containing service account credentials
func CreateOrUpdateServiceAccountCredentialsSecret(ctx context.Context, c client.Client, owner metav1.Object, namespace, name, credentials string) error {
	secretName := name + ServiceAccountCredentialsSecretSuffix
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, c, secret, func() error {
		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}
		secret.Data[ServiceAccountCredentialsField] = []byte(credentials)
		secret.Type = corev1.SecretTypeOpaque

		// Add labels
		if secret.Labels == nil {
			secret.Labels = make(map[string]string)
		}
		secret.Labels["resources.streamnative.io/serviceaccount"] = name

		// Set owner reference
		return controllerutil.SetControllerReference(owner, secret, c.Scheme())
	})

	if err != nil {
		return fmt.Errorf("failed to create or update service account credentials secret: %w", err)
	}

	return nil
}

// GetPrivateKeyFromSecret retrieves the private key from a secret
func GetPrivateKeyFromSecret(ctx context.Context, c client.Client, namespace, name string) (string, error) {
	secretName := name + PrivateKeySecretSuffix
	var secret corev1.Secret
	if err := c.Get(ctx, types.NamespacedName{Namespace: namespace, Name: secretName}, &secret); err != nil {
		if apierrors.IsNotFound(err) {
			return "", fmt.Errorf("private key secret not found: %s", secretName)
		}
		return "", fmt.Errorf("failed to get private key secret: %w", err)
	}

	privateKey, ok := secret.Data[PrivateKeyField]
	if !ok {
		return "", fmt.Errorf("private key not found in secret: %s", secretName)
	}

	return string(privateKey), nil
}
