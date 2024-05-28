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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PulsarConnectionSpec defines the desired state of PulsarConnection
type PulsarConnectionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// AdminServiceURL is the admin service url of the pulsar cluster
	// +optional
	// +kubebuilder:validation:Pattern="^https?://.+$"
	AdminServiceURL string `json:"adminServiceURL"`

	// Authentication defines authentication configurations
	// +optional
	Authentication *PulsarAuthentication `json:"authentication,omitempty"`

	// BrokerServiceURL is the broker service url of the pulsar cluster
	// +optional
	// +kubebuilder:validation:Pattern="^pulsar?://.+$"
	BrokerServiceURL string `json:"brokerServiceURL,omitempty"`

	// BrokerServiceSecureURL is the broker service url for secure connection.
	// +optional
	// +kubebuilder:validation:Pattern="^pulsar\\+ssl://.+$"
	BrokerServiceSecureURL string `json:"brokerServiceSecureURL,omitempty"`

	// AdminServiceSecureURL is the admin service url for secure connection.
	// +optional
	// +kubebuilder:validation:Pattern="^https://.+$"
	AdminServiceSecureURL string `json:"adminServiceSecureURL,omitempty"`

	// BrokerClientTrustCertsFilePath Path for the trusted TLS certificate file for outgoing connection to a server (broker)
	// +optional
	BrokerClientTrustCertsFilePath string `json:"brokerClientTrustCertsFilePath,omitempty"`

	// ClusterName indicates the local cluster name of the pulsar cluster. It should
	// set when enabling the Geo Replication
	// +optional
	ClusterName string `json:"clusterName,omitempty"`

	// TLSEnableHostnameVerification indicates whether to verify the hostname of the broker.
	// Only used when using secure urls.
	// +optional
	TLSEnableHostnameVerification bool `json:"tlsEnableHostnameVerification,omitempty"`

	// TLSAllowInsecureConnection indicates whether to allow insecure connection to the broker.
	// +optional
	TLSAllowInsecureConnection bool `json:"tlsAllowInsecureConnection,omitempty"`

	// TLSTrustCertsFilePath Path for the TLS certificate used to validate the broker endpoint when using TLS.
	// +optional
	TLSTrustCertsFilePath string `json:"tlsTrustCertsFilePath,omitempty"`
}

// PulsarConnectionStatus defines the observed state of PulsarConnection
type PulsarConnectionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ObservedGeneration is the most recent generation observed for this resource.
	// It corresponds to the metadata generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// SecretKeyHash is the hash of the secret ref
	// +optional
	SecretKeyHash string `json:"secretKeyHash,omitempty"`

	// Represents the observations of a connection's current state.
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:categories=pulsar;pulsarres,shortName=pconn
//+kubebuilder:printcolumn:name="ADMIN_SERVICE_URL",type=string,JSONPath=`.spec.adminServiceURL`
//+kubebuilder:printcolumn:name="ADMIN_SERVICE_SECURE_URL",type=string,JSONPath=`.spec.adminServiceSecureURL`,priority=1
//+kubebuilder:printcolumn:name="BROKER_SERVICE_URL",type=string,JSONPath=`.spec.brokerServiceURL`
//+kubebuilder:printcolumn:name="BROKER_SERVICE_SECURE_URL",type=string,JSONPath=`.spec.brokerServiceSecureURL`,priority=1
//+kubebuilder:printcolumn:name="GENERATION",type=string,JSONPath=`.metadata.generation`,priority=1
//+kubebuilder:printcolumn:name="OBSERVED_GENERATION",type=string,JSONPath=`.status.observedGeneration`,priority=1
//+kubebuilder:printcolumn:name="READY",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`

// PulsarConnection is the Schema for the pulsarconnections API
type PulsarConnection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PulsarConnectionSpec   `json:"spec,omitempty"`
	Status PulsarConnectionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PulsarConnectionList contains a list of PulsarConnection
type PulsarConnectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PulsarConnection `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PulsarConnection{}, &PulsarConnectionList{})
}
