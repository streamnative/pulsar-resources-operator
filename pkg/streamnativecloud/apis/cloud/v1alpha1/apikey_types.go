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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// APIKey is a resource that represents an API key for a service account
// +k8s:openapi-gen=true
// +resource:path=apikeys,strategy=APIKeyStrategy
// +kubebuilder:categories=all
type APIKey struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   APIKeySpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status APIKeyStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// APIKeySpec defines the desired state of APIKey
type APIKeySpec struct {
	// InstanceName is the name of the instance this API key is for
	InstanceName string `json:"instanceName,omitempty" protobuf:"bytes,1,opt,name=instanceName"`

	// ServiceAccountName is the name of the service account this API key is for
	ServiceAccountName string `json:"serviceAccountName,omitempty" protobuf:"bytes,2,opt,name=serviceAccountName"`

	// Description is a user defined description of the key
	// +optional
	Description string `json:"description,omitempty" protobuf:"bytes,3,opt,name=description"`

	// Expiration is a duration (as a golang duration string) that defines how long this API key is valid for.
	// This can only be set on initial creation and not updated later
	// +optional
	ExpirationTime *metav1.Time `json:"expirationTime,omitempty" protobuf:"bytes,4,opt,name=expiration"`

	// Revoke is a boolean that defines if the token of this API key should be revoked.
	// +kubebuilder:default=false
	// +optional
	Revoke bool `json:"revoke,omitempty" protobuf:"bytes,5,opt,name=revoke"`

	// EncryptionKey is a public key to encrypt the API Key token.
	// Please provide an RSA key with modulus length of at least 2048 bits.
	// See: https://developer.mozilla.org/en-US/docs/Web/API/SubtleCrypto/generateKey#rsa_key_pair_generation
	// See: https://developer.mozilla.org/en-US/docs/Web/API/SubtleCrypto/exportKey#subjectpublickeyinfo_export
	// +optional
	EncryptionKey *EncryptionKey `json:"encryptionKey,omitempty" protobuf:"bytes,6,opt,name=encryptionKey"`
}

// EncryptionKey specifies a public key for encryption purposes.
// Either a PEM or JWK must be specified.
type EncryptionKey struct {
	// PEM is a PEM-encoded public key in PKIX, ASN.1 DER form ("spki" format).
	// +optional
	PEM string `json:"pem,omitempty" protobuf:"bytes,1,opt,name=pem"`

	// JWK is a JWK-encoded public key.
	// +optional
	JWK string `json:"jwk,omitempty" protobuf:"bytes,2,opt,name=jwk"`
}

// APIKeyStatus defines the observed state of ServiceAccount
type APIKeyStatus struct {
	// ObservedGeneration is the most recent generation observed by the controller.
	// ObservedGeneration int64 `json:"observedGeneration" protobuf:"varint,1,opt,name=observedGeneration"`

	// KeyID is a generated field that is a uid for the token
	// +optional
	KeyID *string `json:"keyId,omitempty" protobuf:"bytes,1,opt,name=keyId"`

	// IssuedAt is a timestamp of when the key was issued, stored as an epoch in seconds
	// +optional
	IssuedAt *metav1.Time `json:"issuedAt,omitempty" protobuf:"bytes,2,opt,name=issuedAt"`

	// ExpiresAt is a timestamp of when the key expires
	// +optional
	ExpiresAt *metav1.Time `json:"expiresAt,omitempty" protobuf:"bytes,3,opt,name=expiresAt"`

	// Token is the plaintext security token issued for the key.
	// +optional
	Token *string `json:"token,omitempty" protobuf:"bytes,4,opt,name=keySecret"`

	// Token is the encrypted security token issued for the key.
	// +optional
	EncryptedToken *EncryptedToken `json:"encryptedToken,omitempty" protobuf:"bytes,7,opt,name=encryptedToken"`

	// Conditions is an array of current observed service account conditions.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,5,rep,name=conditions"`

	// ExpiresAt is a timestamp of when the key was revoked, it triggers revocation action
	// +optional
	RevokedAt *metav1.Time `json:"revokedAt,omitempty" protobuf:"bytes,6,opt,name=revokedAt"`
}

// EncryptedToken is token that is encrypted using an encryption key.
// +structType=atomic
type EncryptedToken struct {
	// JWE is the token as a JSON Web Encryption (JWE) message.
	// For RSA public keys, the key encryption algorithm is RSA-OAEP, and the content encryption algorithm is AES GCM.
	// +optional
	JWE *string `json:"jwe,omitempty" protobuf:"bytes,7,opt,name=jwe"`
}

//+kubebuilder:object:root=true

// APIKeyList contains a list of APIKey
type APIKeyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []APIKey `json:"items"`
}
