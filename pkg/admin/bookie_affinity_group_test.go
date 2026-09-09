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
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	pulsaradmin "github.com/apache/pulsar-client-go/pulsaradmin/pkg/admin"
	"github.com/apache/pulsar-client-go/pulsaradmin/pkg/admin/config"

	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
)

const bookieAffinityPath = "/admin/v2/namespaces/public/default/persistence/bookieAffinity"

// currentAffinity models what Pulsar reports for a namespace's bookie affinity group.
type currentAffinity struct {
	// status is the response code for GET, 404 meaning the namespace has no local policies.
	status int
	// body is the JSON payload returned when status is 200.
	body string
}

var (
	affinityAbsent = currentAffinity{status: http.StatusNotFound}
	// Pulsar answers 200 with an empty group when local policies exist but the affinity
	// group was cleared.
	affinityCleared = currentAffinity{status: http.StatusOK, body: `{}`}
	affinityGroupA  = currentAffinity{
		status: http.StatusOK,
		body:   `{"bookkeeperAffinityGroupPrimary":"group-a","bookkeeperAffinityGroupSecondary":"group-b"}`,
	}
)

func TestApplyBookieAffinityGroup(t *testing.T) {
	tests := []struct {
		name         string
		desired      *resourcev1alpha1.BookieAffinityGroupData
		current      currentAffinity
		wantMethods  []string
		wantPostBody map[string]string
	}{
		{
			// Setting and clearing the affinity group both require superuser access in
			// Pulsar, so a namespace that never uses the feature must not write at all.
			name:        "unset stays unset without writing",
			desired:     nil,
			current:     affinityAbsent,
			wantMethods: []string{http.MethodGet},
		},
		{
			name:        "cleared group is not deleted again",
			desired:     nil,
			current:     affinityCleared,
			wantMethods: []string{http.MethodGet},
		},
		{
			name:        "removing the setting deletes the group",
			desired:     nil,
			current:     affinityGroupA,
			wantMethods: []string{http.MethodGet, http.MethodDelete},
		},
		{
			name: "group is set when none exists",
			desired: &resourcev1alpha1.BookieAffinityGroupData{
				BookkeeperAffinityGroupPrimary:   "group-a",
				BookkeeperAffinityGroupSecondary: "group-b",
			},
			current:     affinityAbsent,
			wantMethods: []string{http.MethodGet, http.MethodPost},
			wantPostBody: map[string]string{
				"bookkeeperAffinityGroupPrimary":   "group-a",
				"bookkeeperAffinityGroupSecondary": "group-b",
			},
		},
		{
			name: "matching group is left alone",
			desired: &resourcev1alpha1.BookieAffinityGroupData{
				BookkeeperAffinityGroupPrimary:   "group-a",
				BookkeeperAffinityGroupSecondary: "group-b",
			},
			current:     affinityGroupA,
			wantMethods: []string{http.MethodGet},
		},
		{
			name: "changed group is rewritten",
			desired: &resourcev1alpha1.BookieAffinityGroupData{
				BookkeeperAffinityGroupPrimary: "group-c",
			},
			current:     affinityGroupA,
			wantMethods: []string{http.MethodGet, http.MethodPost},
			wantPostBody: map[string]string{
				"bookkeeperAffinityGroupPrimary":   "group-c",
				"bookkeeperAffinityGroupSecondary": "",
			},
		},
		{
			// Dropping the secondary group is a real change even though the primary matches.
			name: "dropping the secondary group is rewritten",
			desired: &resourcev1alpha1.BookieAffinityGroupData{
				BookkeeperAffinityGroupPrimary: "group-a",
			},
			current:     affinityGroupA,
			wantMethods: []string{http.MethodGet, http.MethodPost},
			wantPostBody: map[string]string{
				"bookkeeperAffinityGroupPrimary":   "group-a",
				"bookkeeperAffinityGroupSecondary": "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var methods []string
			var postBody map[string]string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.EscapedPath() != bookieAffinityPath {
					t.Errorf("unexpected request to %s", r.URL.EscapedPath())
					w.WriteHeader(http.StatusNotImplemented)
					return
				}
				methods = append(methods, r.Method)

				switch r.Method {
				case http.MethodGet:
					if tt.current.status != http.StatusOK {
						http.Error(w, "Namespace local-policies does not exist", tt.current.status)
						return
					}
					w.Header().Set("Content-Type", "application/json")
					_, _ = io.WriteString(w, tt.current.body)
				case http.MethodPost:
					if err := json.NewDecoder(r.Body).Decode(&postBody); err != nil {
						t.Errorf("decode bookie affinity payload: %v", err)
					}
					w.WriteHeader(http.StatusNoContent)
				default:
					w.WriteHeader(http.StatusNoContent)
				}
			}))
			t.Cleanup(server.Close)

			client, err := pulsaradmin.New(&config.Config{WebServiceURL: server.URL})
			if err != nil {
				t.Fatalf("create Pulsar admin client: %v", err)
			}

			adminClient := &PulsarAdminClient{adminClient: client}
			if err := adminClient.applyBookieAffinityGroup("public/default", tt.desired); err != nil {
				t.Fatalf("apply bookie affinity group: %v", err)
			}

			if len(methods) != len(tt.wantMethods) {
				t.Fatalf("requests = %v, want %v", methods, tt.wantMethods)
			}
			for i := range tt.wantMethods {
				if methods[i] != tt.wantMethods[i] {
					t.Fatalf("request[%d] = %s, want %s", i, methods[i], tt.wantMethods[i])
				}
			}

			if tt.wantPostBody == nil {
				return
			}
			for key, want := range tt.wantPostBody {
				if got := postBody[key]; got != want {
					t.Fatalf("post body %q = %q, want %q", key, got, want)
				}
			}
		})
	}
}

// A GET failure that is not a 404 must surface rather than be mistaken for "no group set",
// which would otherwise let the operator silently skip a required write.
func TestApplyBookieAffinityGroupPropagatesReadFailure(t *testing.T) {
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
	err = adminClient.applyBookieAffinityGroup("public/default", &resourcev1alpha1.BookieAffinityGroupData{
		BookkeeperAffinityGroupPrimary: "group-a",
	})
	if err == nil {
		t.Fatal("apply bookie affinity group succeeded, want error")
	}
	if reason := ErrorReason(err); reason != ReasonForbidden {
		t.Fatalf("error reason = %v, want %v", reason, ReasonForbidden)
	}
}

const autoTopicCreationPath = "/admin/v2/namespaces/public/default/autoTopicCreation"

// affinityDenialServer answers the bookie affinity endpoint with status and records every
// path the reconcile touches, so a test can assert what still ran after the denial.
func affinityDenialServer(t *testing.T, status int, paths *[]string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*paths = append(*paths, r.URL.EscapedPath())
		if r.URL.EscapedPath() == bookieAffinityPath {
			http.Error(w, "Don't have admin permission", status)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)
	return server
}

func contains(paths []string, want string) bool {
	for _, p := range paths {
		if p == want {
			return true
		}
	}
	return false
}

// Reading the bookie affinity group is superuser-only in Pulsar, so a tenant-admin
// connection is denied on it even for a namespace that never asked for a group. That
// denial must not abort the reconcile: every policy after it would be skipped and the
// namespace would never reach Ready.
func TestApplyNamespacePoliciesToleratesDeniedAffinityReadWhenUnset(t *testing.T) {
	for _, status := range []int{http.StatusUnauthorized, http.StatusForbidden} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			var paths []string
			server := affinityDenialServer(t, status, &paths)

			client, err := pulsaradmin.New(&config.Config{WebServiceURL: server.URL})
			if err != nil {
				t.Fatalf("create Pulsar admin client: %v", err)
			}

			adminClient := &PulsarAdminClient{adminClient: client}
			err = adminClient.applyNamespacePolicies("public/default", &NamespaceParams{
				BookieAffinityGroup: nil,
				TopicAutoCreationConfig: &resourcev1alpha1.TopicAutoCreationConfig{
					Allow: true,
					Type:  "non-partitioned",
				},
			})
			if err != nil {
				t.Fatalf("apply namespace policies: %v", err)
			}

			if !contains(paths, bookieAffinityPath) {
				t.Fatalf("bookie affinity was never read, paths = %v", paths)
			}
			// The policy that follows the affinity block must still have been applied.
			if !contains(paths, autoTopicCreationPath) {
				t.Fatalf("a later policy was skipped after the denied read, paths = %v", paths)
			}
		})
	}
}

// A group the user explicitly asked for is the opposite case: the operator has been told
// to write it, so a permission failure has to surface rather than be silently skipped.
func TestApplyNamespacePoliciesFailsOnDeniedAffinityReadWhenRequested(t *testing.T) {
	for _, status := range []int{http.StatusUnauthorized, http.StatusForbidden} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			var paths []string
			server := affinityDenialServer(t, status, &paths)

			client, err := pulsaradmin.New(&config.Config{WebServiceURL: server.URL})
			if err != nil {
				t.Fatalf("create Pulsar admin client: %v", err)
			}

			adminClient := &PulsarAdminClient{adminClient: client}
			err = adminClient.applyNamespacePolicies("public/default", &NamespaceParams{
				BookieAffinityGroup: &resourcev1alpha1.BookieAffinityGroupData{
					BookkeeperAffinityGroupPrimary: "group-a",
				},
			})
			if err == nil {
				t.Fatal("apply namespace policies succeeded, want error")
			}
			if !IsPermissionDenied(err) {
				t.Fatalf("error reason = %v, want a permission denial", ErrorReason(err))
			}
			if contains(paths, autoTopicCreationPath) {
				t.Fatalf("reconcile continued past a required write it could not make, paths = %v", paths)
			}
		})
	}
}
