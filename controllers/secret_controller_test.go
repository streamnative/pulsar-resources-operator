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
	"testing"

	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestResolveSecretRefDataEncodesSelectedBinaryDataKeys(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("add core scheme: %v", err)
	}
	if err := resourcev1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add resource scheme: %v", err)
	}

	k8sSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "source-secret",
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"username":     []byte("admin"),
			"keystore.p12": {0x00, 0x01, 0x02, 0xff},
		},
	}
	secretCR := &resourcev1alpha1.Secret{
		Spec: resourcev1alpha1.SecretSpec{
			SecretRef: &resourcev1alpha1.KubernetesSecretReference{
				Namespace:      "default",
				Name:           "source-secret",
				BinaryDataKeys: []string{"keystore.p12"},
			},
		},
	}

	reconciler := &SecretReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(k8sSecret).Build(),
	}

	data, binaryData, secretType, err := reconciler.resolveSecretRefData(context.Background(), secretCR)
	if err != nil {
		t.Fatalf("resolve SecretRef data: %v", err)
	}
	if got := data["username"]; got != "admin" {
		t.Fatalf("resolved data username = %q, want admin", got)
	}
	if _, ok := data["keystore.p12"]; ok {
		t.Fatalf("keystore.p12 should not be resolved as text data")
	}
	if got := binaryData["keystore.p12"]; got != "AAEC/w==" {
		t.Fatalf("resolved binaryData keystore.p12 = %q, want AAEC/w==", got)
	}
	if secretType == nil || *secretType != corev1.SecretTypeOpaque {
		t.Fatalf("resolved type = %v, want Opaque", secretType)
	}
}

func TestResolveSecretRefDataErrorsForMissingBinaryDataKey(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("add core scheme: %v", err)
	}
	if err := resourcev1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add resource scheme: %v", err)
	}

	k8sSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "source-secret"},
		Data:       map[string][]byte{"username": []byte("admin")},
	}
	secretCR := &resourcev1alpha1.Secret{
		Spec: resourcev1alpha1.SecretSpec{
			SecretRef: &resourcev1alpha1.KubernetesSecretReference{
				Namespace:      "default",
				Name:           "source-secret",
				BinaryDataKeys: []string{"missing.p12"},
			},
		},
	}

	reconciler := &SecretReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(k8sSecret).Build(),
	}

	_, _, _, err := reconciler.resolveSecretRefData(context.Background(), secretCR)
	if err == nil {
		t.Fatal("expected missing binaryData key error")
	}
}

func TestValidateSecretDataKeysRejectsDuplicates(t *testing.T) {
	secretCR := &resourcev1alpha1.Secret{
		Spec: resourcev1alpha1.SecretSpec{
			Data:       map[string]string{"shared": "text"},
			BinaryData: map[string]string{"shared": "AAEC/w=="},
		},
	}

	if err := validateSecretDataKeys(secretCR); err == nil {
		t.Fatal("expected duplicate key validation error")
	}
}
