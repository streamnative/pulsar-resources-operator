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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	pulsaradmin "github.com/apache/pulsar-client-go/pulsaradmin/pkg/admin"
	"github.com/apache/pulsar-client-go/pulsaradmin/pkg/admin/config"
	pulsarutils "github.com/apache/pulsar-client-go/pulsaradmin/pkg/utils"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"

	rutils "github.com/streamnative/pulsar-resources-operator/pkg/utils"
)

func TestApplyNamespacePoliciesReplacesBacklogQuotaType(t *testing.T) {
	tests := []struct {
		name          string
		desiredType   pulsarutils.BacklogQuotaType
		existingType  pulsarutils.BacklogQuotaType
		staleType     pulsarutils.BacklogQuotaType
		limitSize     string
		retentionSize string
		wantDelete    bool
	}{
		{
			name:          "destination storage to message age while lowering retention",
			desiredType:   pulsarutils.MessageAge,
			existingType:  pulsarutils.DestinationStorage,
			staleType:     pulsarutils.DestinationStorage,
			limitSize:     "-1",
			retentionSize: "10Gi",
			wantDelete:    true,
		},
		{
			name:         "message age to destination storage",
			desiredType:  pulsarutils.DestinationStorage,
			existingType: pulsarutils.MessageAge,
			staleType:    pulsarutils.MessageAge,
			limitSize:    "10Gi",
			wantDelete:   true,
		},
		{
			name:         "opposite quota does not exist",
			desiredType:  pulsarutils.DestinationStorage,
			existingType: pulsarutils.DestinationStorage,
			staleType:    pulsarutils.MessageAge,
			limitSize:    "10Gi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type request struct {
				method    string
				path      string
				quotaType string
			}

			var requests []request
			staleQuotaExists := tt.existingType == tt.staleType
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.EscapedPath() {
				case "/admin/v2/namespaces/public/default/backlogQuotaMap":
					requests = append(requests, request{
						method: r.Method,
						path:   r.URL.EscapedPath(),
					})
					w.Header().Set("Content-Type", "application/json")
					_, _ = fmt.Fprintf(w, `{"%s":{}}`, tt.existingType)
					return
				case "/admin/v2/namespaces/public/default/backlogQuota":
					adminRequest := request{
						method:    r.Method,
						path:      r.URL.EscapedPath(),
						quotaType: r.URL.Query().Get("backlogQuotaType"),
					}
					requests = append(requests, adminRequest)
					if r.Method == http.MethodDelete && adminRequest.quotaType == string(tt.staleType) {
						staleQuotaExists = false
					}
				case "/admin/v2/namespaces/public/default/retention":
					requests = append(requests, request{
						method: r.Method,
						path:   r.URL.EscapedPath(),
					})
					if staleQuotaExists {
						http.Error(w, "Retention Quota must exceed configured backlog quota", http.StatusPreconditionFailed)
						return
					}
				}
				w.WriteHeader(http.StatusNoContent)
			}))
			t.Cleanup(server.Close)

			client, err := pulsaradmin.New(&config.Config{WebServiceURL: server.URL})
			if err != nil {
				t.Fatalf("create Pulsar admin client: %v", err)
			}

			limitTime := rutils.Duration("72h")
			limitSize := resource.MustParse(tt.limitSize)
			params := &NamespaceParams{
				BacklogQuotaLimitTime:       &limitTime,
				BacklogQuotaLimitSize:       &limitSize,
				BacklogQuotaRetentionPolicy: ptr.To("consumer_backlog_eviction"),
				BacklogQuotaType:            ptr.To(string(tt.desiredType)),
			}
			if tt.retentionSize != "" {
				retentionSize := resource.MustParse(tt.retentionSize)
				params.RetentionSize = &retentionSize
			}

			adminClient := &PulsarAdminClient{adminClient: client}
			err = adminClient.applyNamespacePolicies("public/default", params)
			if err != nil {
				t.Fatalf("apply namespace policies: %v", err)
			}

			want := []request{
				{
					method: http.MethodGet,
					path:   "/admin/v2/namespaces/public/default/backlogQuotaMap",
				},
			}
			if tt.retentionSize != "" && tt.wantDelete {
				want = append(want, request{
					method:    http.MethodDelete,
					path:      "/admin/v2/namespaces/public/default/backlogQuota",
					quotaType: string(tt.staleType),
				})
			}
			if tt.retentionSize != "" {
				want = append(want, request{
					method: http.MethodPost,
					path:   "/admin/v2/namespaces/public/default/retention",
				})
			}
			want = append(want, request{
				method:    http.MethodPost,
				path:      "/admin/v2/namespaces/public/default/backlogQuota",
				quotaType: string(tt.desiredType),
			})
			if tt.retentionSize == "" && tt.wantDelete {
				want = append(want, request{
					method:    http.MethodDelete,
					path:      "/admin/v2/namespaces/public/default/backlogQuota",
					quotaType: string(tt.staleType),
				})
			}
			if len(requests) != len(want) {
				t.Fatalf("requests = %#v, want %#v", requests, want)
			}
			for i := range want {
				if requests[i] != want[i] {
					t.Fatalf("request[%d] = %#v, want %#v", i, requests[i], want[i])
				}
			}
		})
	}
}
