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
	"net/http"
	"net/http/httptest"
	"testing"

	pulsaradmin "github.com/apache/pulsar-client-go/pulsaradmin/pkg/admin"
	"github.com/apache/pulsar-client-go/pulsaradmin/pkg/admin/config"
	"k8s.io/utils/ptr"

	"github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
)

func TestApplyNamespacePoliciesBookieAffinityPermissions(t *testing.T) {
	for _, tt := range []struct {
		name     string
		explicit bool
		status   int
		wantErr  bool
	}{
		{name: "delete succeeds", status: http.StatusNoContent},
		{name: "delete unauthorized continues", status: http.StatusUnauthorized},
		{name: "delete forbidden continues", status: http.StatusForbidden},
		{name: "delete server failure propagates", status: http.StatusInternalServerError, wantErr: true},
		{name: "delete not found propagates", status: http.StatusNotFound, wantErr: true},
		{name: "set succeeds", explicit: true, status: http.StatusNoContent},
		{name: "set unauthorized propagates", explicit: true, status: http.StatusUnauthorized, wantErr: true},
		{name: "set forbidden propagates", explicit: true, status: http.StatusForbidden, wantErr: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var affinityCalled, subsequentPolicyCalled bool
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/admin/v2/namespaces/public/default/persistence/bookieAffinity":
					affinityCalled = true
					wantMethod := http.MethodDelete
					if tt.explicit {
						wantMethod = http.MethodPost
					}
					if r.Method != wantMethod {
						t.Errorf("affinity method = %s, want %s", r.Method, wantMethod)
					}
					if tt.status != http.StatusNoContent {
						http.Error(w, "policy operation failed", tt.status)
						return
					}
				case "/admin/v2/namespaces/public/default/schemaValidationEnforced":
					subsequentPolicyCalled = true
					if !affinityCalled {
						t.Error("subsequent policy applied before affinity operation")
					}
				}
				w.WriteHeader(http.StatusNoContent)
			}))
			defer server.Close()
			client, err := pulsaradmin.New(&config.Config{WebServiceURL: server.URL})
			if err != nil {
				t.Fatalf("create admin client: %v", err)
			}
			params := &NamespaceParams{SchemaValidationEnforced: ptr.To(true)}
			if tt.explicit {
				params.BookieAffinityGroup = &v1alpha1.BookieAffinityGroupData{BookkeeperAffinityGroupPrimary: "primary"}
			}
			adminClient := &PulsarAdminClient{adminClient: client}
			err = adminClient.applyNamespacePolicies("public/default", params)
			if (err != nil) != tt.wantErr {
				t.Fatalf("apply policies error = %v, want error %v", err, tt.wantErr)
			}
			if tt.wantErr && ErrorReason(err) != Reason(tt.status) {
				t.Errorf("error reason = %v, want %v", ErrorReason(err), tt.status)
			}
			if !affinityCalled {
				t.Error("affinity operation was not called")
			}
			if subsequentPolicyCalled == tt.wantErr {
				t.Errorf("subsequent policy called = %v, want %v", subsequentPolicyCalled, !tt.wantErr)
			}
		})
	}
}
