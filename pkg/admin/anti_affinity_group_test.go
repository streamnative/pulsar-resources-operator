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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	pulsaradmin "github.com/apache/pulsar-client-go/pulsaradmin/pkg/admin"
	"github.com/apache/pulsar-client-go/pulsaradmin/pkg/admin/config"
	"k8s.io/utils/ptr"
)

const antiAffinityPath = "/admin/v2/namespaces/public/default/antiAffinity"

// antiAffinityServer answers the anti-affinity endpoint the way Pulsar 4.0 does: GET is 204
// with no body when no group is set and the group as a JSON string otherwise. DELETE fails
// with a 500 when the namespace has no local policies yet, as on a namespace without topics.
func antiAffinityServer(t *testing.T, current string, hasLocalPolicies bool,
	methods *[]string, posted *string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.EscapedPath() != antiAffinityPath {
			t.Errorf("unexpected request to %s", r.URL.EscapedPath())
			w.WriteHeader(http.StatusNotImplemented)
			return
		}
		*methods = append(*methods, r.Method)
		switch r.Method {
		case http.MethodGet:
			if current == "" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(current)
		case http.MethodPost:
			if err := json.NewDecoder(r.Body).Decode(posted); err != nil {
				t.Errorf("decode anti-affinity payload: %v", err)
			}
			w.WriteHeader(http.StatusNoContent)
		case http.MethodDelete:
			if !hasLocalPolicies {
				http.Error(w, "MetadataStoreException$NotFoundException", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

func TestApplyNamespaceAntiAffinityGroup(t *testing.T) {
	tests := []struct {
		name             string
		desired          *string
		current          string
		hasLocalPolicies bool
		wantMethods      []string
		wantPosted       string
	}{
		{
			// The case that kept a new namespace from becoming Ready: the broker cannot
			// clear a group on a namespace that has no local policies yet.
			name:             "unset stays unset without writing on a new namespace",
			desired:          nil,
			current:          "",
			hasLocalPolicies: false,
			wantMethods:      []string{http.MethodGet},
		},
		{
			name:             "unset stays unset without writing",
			desired:          nil,
			current:          "",
			hasLocalPolicies: true,
			wantMethods:      []string{http.MethodGet},
		},
		{
			name:             "removing the setting deletes the group",
			desired:          nil,
			current:          "group-a",
			hasLocalPolicies: true,
			wantMethods:      []string{http.MethodGet, http.MethodDelete},
		},
		{
			name:             "group is set when none exists",
			desired:          ptr.To("group-a"),
			current:          "",
			hasLocalPolicies: false,
			wantMethods:      []string{http.MethodGet, http.MethodPost},
			wantPosted:       "group-a",
		},
		{
			name:             "matching group is left alone",
			desired:          ptr.To("group-a"),
			current:          "group-a",
			hasLocalPolicies: true,
			wantMethods:      []string{http.MethodGet},
		},
		{
			name:             "changed group is rewritten",
			desired:          ptr.To("group-b"),
			current:          "group-a",
			hasLocalPolicies: true,
			wantMethods:      []string{http.MethodGet, http.MethodPost},
			wantPosted:       "group-b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var methods []string
			var posted string
			server := antiAffinityServer(t, tt.current, tt.hasLocalPolicies, &methods, &posted)

			client, err := pulsaradmin.New(&config.Config{WebServiceURL: server.URL})
			if err != nil {
				t.Fatalf("create Pulsar admin client: %v", err)
			}

			adminClient := &PulsarAdminClient{adminClient: client}
			if err := adminClient.applyNamespaceAntiAffinityGroup("public/default", tt.desired); err != nil {
				t.Fatalf("apply anti-affinity group: %v", err)
			}

			if len(methods) != len(tt.wantMethods) {
				t.Fatalf("requests = %v, want %v", methods, tt.wantMethods)
			}
			for i := range tt.wantMethods {
				if methods[i] != tt.wantMethods[i] {
					t.Fatalf("request[%d] = %s, want %s", i, methods[i], tt.wantMethods[i])
				}
			}
			if posted != tt.wantPosted {
				t.Fatalf("posted group = %q, want %q", posted, tt.wantPosted)
			}
		})
	}
}

// A GET failure must surface rather than be mistaken for "no group set", which would
// otherwise let the operator silently skip a required write.
func TestApplyNamespaceAntiAffinityGroupPropagatesReadFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("unexpected %s request after a failed read", r.Method)
		}
		http.Error(w, "Don't have admin permission", http.StatusForbidden)
	}))
	t.Cleanup(server.Close)

	client, err := pulsaradmin.New(&config.Config{WebServiceURL: server.URL})
	if err != nil {
		t.Fatalf("create Pulsar admin client: %v", err)
	}

	adminClient := &PulsarAdminClient{adminClient: client}
	err = adminClient.applyNamespaceAntiAffinityGroup("public/default", ptr.To("group-a"))
	if err == nil {
		t.Fatal("apply anti-affinity group succeeded, want error")
	}
	if reason := ErrorReason(err); reason != ReasonForbidden {
		t.Fatalf("error reason = %v, want %v", reason, ReasonForbidden)
	}
}
