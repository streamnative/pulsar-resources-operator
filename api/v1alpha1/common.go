// Copyright (c) 2021 StreamNative, Inc.. All Rights Reserved.

package v1alpha1

type SecretKeyRef struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

type ValueOrSecretRef struct {
	// +optional
	Value *string `json:"value,omitempty"`

	// +optional
	SecretRef *SecretKeyRef `json:"secretRef,omitempty"`
}

type PulsarAuthentication struct {
	// +optional
	Token *ValueOrSecretRef `json:"token,omitempty"`

	// +optional
	OAuth2 *PulsarAuthenticationOAuth2 `json:"oauth2,omitempty"`
}

type PulsarResourceLifeCyclePolicy string

const (
	KeepAfterDeletion    PulsarResourceLifeCyclePolicy = "KeepAfterDeletion"
	CleanUpAfterDeletion PulsarResourceLifeCyclePolicy = "CleanUpAfterDeletion"
)

type PulsarAuthenticationOAuth2 struct {
	IssuerEndpoint string           `json:"issuerEndpoint"`
	ClientID       string           `json:"clientID"`
	Audience       string           `json:"audience"`
	Key            ValueOrSecretRef `json:"key"`
}
