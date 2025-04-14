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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// APIKeySpec defines the desired state of APIKey
type APIKeySpec struct {
	// APIServerRef is the reference to the StreamNativeCloudConnection
	// +required
	APIServerRef corev1.LocalObjectReference `json:"apiServerRef"`

	// InstanceName is the name of the instance this API key is for
	// +optional
	InstanceName string `json:"instanceName,omitempty"`

	// ServiceAccountName is the name of the service account this API key is for
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// Description is a user defined description of the key
	// +optional
	Description string `json:"description,omitempty"`

	// ExpirationTime is a timestamp that defines when this API key will expire
	// This can only be set on initial creation and not updated later
	// +optional
	ExpirationTime *metav1.Time `json:"expirationTime,omitempty"`

	// Revoke indicates whether this API key should be revoked
	// +optional
	Revoke bool `json:"revoke,omitempty"`

	// EncryptionKey contains the public key used to encrypt the token
	// +optional
	EncryptionKey *EncryptionKey `json:"encryptionKey,omitempty"`
}

// EncryptionKey contains a public key used for encryption
// +structType=atomic
type EncryptionKey struct {
	// PEM is the public key in PEM format
	// +optional
	PEM string `json:"pem,omitempty"`
}

// APIKeyStatus defines the observed state of APIKey
type APIKeyStatus struct {
	// Conditions represent the latest available observations of an object's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the last observed generation
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// KeyID is a generated field that is a uid for the token
	// +optional
	KeyID *string `json:"keyId,omitempty"`

	// IssuedAt is a timestamp of when the key was issued
	// +optional
	IssuedAt *metav1.Time `json:"issuedAt,omitempty"`

	// ExpiresAt is a timestamp of when the key expires
	// +optional
	ExpiresAt *metav1.Time `json:"expiresAt,omitempty"`

	// Token is the plaintext security token issued for the key
	// +optional
	Token *string `json:"token,omitempty"`

	// EncryptedToken is the encrypted security token issued for the key
	// +optional
	EncryptedToken *EncryptedToken `json:"encryptedToken,omitempty"`

	// RevokedAt is a timestamp of when the key was revoked, it triggers revocation action
	// +optional
	RevokedAt *metav1.Time `json:"revokedAt,omitempty"`
}

// EncryptedToken is token that is encrypted using an encryption key
// +structType=atomic
type EncryptedToken struct {
	// JWE is the token as a JSON Web Encryption (JWE) message
	// For RSA public keys, the key encryption algorithm is RSA-OAEP, and the content encryption algorithm is AES GCM
	// +optional
	JWE *string `json:"jwe,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Namespaced,categories={streamnative,all}
//+kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status"
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// APIKey is the Schema for the APIKeys API
type APIKey struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   APIKeySpec   `json:"spec,omitempty"`
	Status APIKeyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// APIKeyList contains a list of APIKey
type APIKeyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []APIKey `json:"items"`
}

func init() {
	SchemeBuilder.Register(&APIKey{}, &APIKeyList{})
}
