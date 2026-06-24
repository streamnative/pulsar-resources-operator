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
	"testing"

	"k8s.io/utils/ptr"

	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
)

func TestConvertOffloadPoliciesPreservesExplicitZeroThresholds(t *testing.T) {
	policies := convertOffloadPolicies(&resourcev1alpha1.OffloadPolicies{
		ManagedLedgerOffloadDriver:              "aws-s3",
		ManagedLedgerOffloadThresholdInBytes:    ptr.To[int64](0),
		ManagedLedgerOffloadThresholdInSeconds:  ptr.To[int64](0),
		ManagedLedgerOffloadDeletionLagInMillis: ptr.To[int64](0),
	})

	body, err := json.Marshal(policies)
	if err != nil {
		t.Fatalf("marshal offload policies: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal offload policies: %v", err)
	}

	for _, field := range []string{
		"managedLedgerOffloadThresholdInBytes",
		"managedLedgerOffloadThresholdInSeconds",
		"managedLedgerOffloadDeletionLagInMillis",
	} {
		if value, ok := got[field]; !ok || value != float64(0) {
			t.Fatalf("%s = %v, present = %t, want explicit zero", field, value, ok)
		}
	}
}
