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

package utils

import (
	"fmt"
	"net/http"

	"github.com/apache/pulsar-client-go/pulsaradmin/pkg/admin/auth"
)

// tokenSupplierAuthProvider is a custom auth provider for the Pulsar Admin client that can dynamically obtain an
// auth token, e.g. for Service Accounts
type tokenSupplierAuthProvider struct {
	T             http.RoundTripper
	tokenSupplier func() (string, error)
}

// NewPulsarAdminAuthProviderWithTokenSupplier creates a new dynamic authentication provider for the pulsar admin client
func NewPulsarAdminAuthProviderWithTokenSupplier(supplier func() (string, error), t http.RoundTripper) auth.Provider {
	return &tokenSupplierAuthProvider{
		T:             t,
		tokenSupplier: supplier,
	}
}

// Transport returns the underlying transport of the provider
func (t *tokenSupplierAuthProvider) Transport() http.RoundTripper {
	return t.T
}

// WithTransport replaces the underlying transport of the provider
func (t *tokenSupplierAuthProvider) WithTransport(tripper http.RoundTripper) {
	t.T = tripper
}

// RoundTrip executes the actual http request enriched with the dynamic token
func (t *tokenSupplierAuthProvider) RoundTrip(req *http.Request) (*http.Response, error) {
	token, err := t.tokenSupplier()
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	return t.T.RoundTrip(req)
}
