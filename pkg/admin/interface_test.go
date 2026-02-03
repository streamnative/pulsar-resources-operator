// Copyright 2026 StreamNative
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

package admin

import (
	"errors"
	"net/http"
	"reflect"
	"testing"

	"github.com/apache/pulsar-client-go/oauth2"
	"github.com/apache/pulsar-client-go/pulsaradmin/pkg/admin/auth"
)

type fakeAuthProvider struct {
	transport http.RoundTripper
}

func (p *fakeAuthProvider) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, errors.New("not implemented")
}

func (p *fakeAuthProvider) Transport() http.RoundTripper {
	return p.transport
}

func (p *fakeAuthProvider) WithTransport(t http.RoundTripper) {
	p.transport = t
}

func TestNewPulsarAdminOAuth2IssuerURL(t *testing.T) {
	original := newAuthenticationOAuth2WithFlow
	t.Cleanup(func() {
		newAuthenticationOAuth2WithFlow = original
	})

	var gotIssuer oauth2.Issuer
	var gotFlow oauth2.ClientCredentialsFlowOptions
	called := false

	newAuthenticationOAuth2WithFlow = func(issuer oauth2.Issuer, flowOptions oauth2.ClientCredentialsFlowOptions) (auth.Provider, error) {
		called = true
		gotIssuer = issuer
		gotFlow = flowOptions
		return &fakeAuthProvider{}, nil
	}

	conf := PulsarAdminConfig{
		IssuerEndpoint: "https://issuer.example.com",
		ClientID:       "client-id",
		Audience:       "audience",
		Scope:          "scope-a scope-b",
		KeyFilePath:    "test-key-file.json",
	}

	_, err := NewPulsarAdmin(conf)
	if err != nil {
		t.Fatalf("NewPulsarAdmin() error = %v", err)
	}

	if !called {
		t.Fatalf("expected NewAuthenticationOAuth2WithFlow to be called")
	}

	if gotIssuer.IssuerEndpoint != conf.IssuerEndpoint {
		t.Fatalf("issuer endpoint = %q, want %q", gotIssuer.IssuerEndpoint, conf.IssuerEndpoint)
	}

	if gotFlow.IssuerURL != conf.IssuerEndpoint {
		t.Fatalf("flow issuer url = %q, want %q", gotFlow.IssuerURL, conf.IssuerEndpoint)
	}

	if gotFlow.KeyFile == "" {
		t.Fatalf("expected flow key file to be set")
	}

	expectedScopes := []string{"scope-a", "scope-b"}
	if !reflect.DeepEqual(gotFlow.AdditionalScopes, expectedScopes) {
		t.Fatalf("additional scopes = %v, want %v", gotFlow.AdditionalScopes, expectedScopes)
	}
}
