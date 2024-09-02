// Copyright 2023 StreamNative
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

// ClientCredentials represents the client credentials for OAuth2
type ClientCredentials struct {
	IssuerURL  string `json:"issuerUrl,omitempty"`
	Audience   string `json:"audience,omitempty"`
	Scope      string `json:"scope,omitempty"`
	PrivateKey string `json:"privateKey,omitempty"`
	ClientID   string `json:"clientId,omitempty"`
}
