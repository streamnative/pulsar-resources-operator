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

package streamnativecloud

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newTestSecretClient(t *testing.T, handler http.Handler) *SecretClient {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	httpClient := server.Client()
	if httpClient.Transport == nil {
		httpClient.Transport = http.DefaultTransport
	}
	apiConn := &APIConnection{
		config: &resourcev1alpha1.StreamNativeCloudConnection{
			Spec: resourcev1alpha1.StreamNativeCloudConnectionSpec{Server: server.URL},
		},
		client:      httpClient,
		initialized: true,
	}
	secretClient, err := NewSecretClient(apiConn, "test-org")
	if err != nil {
		t.Fatalf("create secret client: %v", err)
	}
	return secretClient
}

func newSecretWithBinaryData() *resourcev1alpha1.Secret {
	return &resourcev1alpha1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "test-secret"},
		Spec: resourcev1alpha1.SecretSpec{
			Data: map[string]string{
				"username": "admin",
			},
			BinaryData: map[string]string{
				"keystore.p12": "AAEC/w==",
			},
		},
	}
}

func assertSecretPayloadHasBinaryData(t *testing.T, payload map[string]any) {
	t.Helper()

	data, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("payload data = %#v, want object", payload["data"])
	}
	if data["username"] != "admin" {
		t.Fatalf("payload data username = %#v, want admin", data["username"])
	}

	binaryData, ok := payload["binaryData"].(map[string]any)
	if !ok {
		t.Fatalf("payload binaryData = %#v, want object", payload["binaryData"])
	}
	if binaryData["keystore.p12"] != "AAEC/w==" {
		t.Fatalf("payload binaryData keystore.p12 = %#v, want AAEC/w==", binaryData["keystore.p12"])
	}
}

func TestConvertToCloudSecretIncludesBinaryData(t *testing.T) {
	cloudSecret := convertToCloudSecret(newSecretWithBinaryData())
	if got := cloudSecret.Data["username"]; got != "admin" {
		t.Fatalf("cloud data username = %q, want admin", got)
	}
	if got := cloudSecret.BinaryData["keystore.p12"]; got != "AAEC/w==" {
		t.Fatalf("cloud binaryData keystore.p12 = %q, want AAEC/w==", got)
	}

	payload, err := json.Marshal(cloudSecret)
	if err != nil {
		t.Fatalf("marshal cloud secret: %v", err)
	}

	var decoded struct {
		Data       map[string]string `json:"data"`
		BinaryData map[string]string `json:"binaryData"`
	}
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshal cloud secret payload: %v", err)
	}
	if decoded.Data["username"] != "admin" {
		t.Fatalf("payload data username = %q, want admin", decoded.Data["username"])
	}
	if decoded.BinaryData["keystore.p12"] != "AAEC/w==" {
		t.Fatalf("payload binaryData keystore.p12 = %q, want AAEC/w==", decoded.BinaryData["keystore.p12"])
	}
}

func TestCreateSecretPayloadIncludesBinaryData(t *testing.T) {
	var payload map[string]any
	secretClient := newTestSecretClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/apis/cloud.streamnative.io/v1alpha1/namespaces/test-org/secrets" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode create request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"apiVersion":"cloud.streamnative.io/v1alpha1","kind":"Secret","metadata":{"name":"test-secret"}}`))
	}))

	if _, err := secretClient.CreateSecret(context.Background(), newSecretWithBinaryData()); err != nil {
		t.Fatalf("create secret: %v", err)
	}
	assertSecretPayloadHasBinaryData(t, payload)
}

func TestUpdateSecretPayloadIncludesBinaryData(t *testing.T) {
	var payload map[string]any
	secretClient := newTestSecretClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/apis/cloud.streamnative.io/v1alpha1/namespaces/test-org/secrets/test-secret":
			_, _ = w.Write([]byte(`{"apiVersion":"cloud.streamnative.io/v1alpha1","kind":"Secret","metadata":{"name":"test-secret","resourceVersion":"123"}}`))
		case r.Method == http.MethodPut && r.URL.Path == "/apis/cloud.streamnative.io/v1alpha1/namespaces/test-org/secrets/test-secret":
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode update request body: %v", err)
			}
			_, _ = w.Write([]byte(`{"apiVersion":"cloud.streamnative.io/v1alpha1","kind":"Secret","metadata":{"name":"test-secret","resourceVersion":"124"}}`))
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))

	if _, err := secretClient.UpdateSecret(context.Background(), newSecretWithBinaryData()); err != nil {
		t.Fatalf("update secret: %v", err)
	}
	assertSecretPayloadHasBinaryData(t, payload)
}
