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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// StreamNativeCloudConnectionSpec defines the desired state of StreamNativeCloudConnection
type StreamNativeCloudConnectionSpec struct {
	// Server is the URL of the API server
	// +required
	Server string `json:"server"`

	// CertificateAuthorityData is the PEM-encoded certificate authority certificates
	// +optional
	CertificateAuthorityData []byte `json:"certificateAuthorityData,omitempty"`

	// InsecureSkipTLSVerify indicates whether to skip TLS verification
	// +optional
	InsecureSkipTLSVerify bool `json:"insecureSkipTLSVerify,omitempty"`

	// Auth defines the authentication configuration
	// +required
	Auth AuthConfig `json:"auth"`

	// Logs defines the logging service configuration
	// +optional
	Logs LogConfig `json:"logs,omitempty"`

	// Organization is the organization to use in the API server
	// If not specified, the operator will use the connection name as the organization
	// +optional
	Organization string `json:"organization,omitempty"`
}

// AuthConfig defines the authentication configuration
type AuthConfig struct {
	// CredentialsRef is the reference to the service account credentials secret
	// +required
	CredentialsRef corev1.LocalObjectReference `json:"credentialsRef"`
}

// ServiceAccountCredentials defines the structure of the service account credentials file
type ServiceAccountCredentials struct {
	// Type is the type of the credentials, must be "sn_service_account"
	Type string `json:"type"`

	// ClientID is the OAuth2 client ID
	ClientID string `json:"client_id"`

	// ClientSecret is the OAuth2 client secret
	ClientSecret string `json:"client_secret"`

	// ClientEmail is the email address of the service account
	ClientEmail string `json:"client_email"`

	// IssuerURL is the OpenID Connect issuer URL
	IssuerURL string `json:"issuer_url"`
}

// LogConfig defines the logging service configuration
type LogConfig struct {
	// ServiceURL is the URL of the logging service
	// +required
	ServiceURL string `json:"serviceUrl"`
}

// StreamNativeCloudConnectionStatus defines the observed state of StreamNativeCloudConnection
type StreamNativeCloudConnectionStatus struct {
	// Conditions represent the latest available observations of an object's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the last observed generation.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// LastConnectedTime is the last time we successfully connected to the API server
	// +optional
	LastConnectedTime *metav1.Time `json:"lastConnectedTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Namespaced,categories={streamnative,all}
//+kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status"
//+kubebuilder:printcolumn:name="SERVER",type="string",JSONPath=".spec.server"
//+kubebuilder:printcolumn:name="ORGANIZATION",type="string",JSONPath=".spec.organization"

// StreamNativeCloudConnection is the Schema for the StreamNativeCloudConnections API
type StreamNativeCloudConnection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StreamNativeCloudConnectionSpec   `json:"spec,omitempty"`
	Status StreamNativeCloudConnectionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// StreamNativeCloudConnectionList contains a list of StreamNativeCloudConnection
type StreamNativeCloudConnectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StreamNativeCloudConnection `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StreamNativeCloudConnection{}, &StreamNativeCloudConnectionList{})
}
