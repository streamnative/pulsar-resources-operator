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
	"fmt"
	"testing"

	"github.com/apache/pulsar-client-go/pulsaradmin/pkg/rest"
)

func TestIsAlreadyExist(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "rest.Error with 409",
			err:      rest.Error{Code: 409, Reason: "already exists"},
			expected: true,
		},
		{
			name:     "rest.Error pointer with 409",
			err:      &rest.Error{Code: 409, Reason: "already exists"},
			expected: true,
		},
		{
			name:     "wrapped rest.Error with 409",
			err:      fmt.Errorf("wrapped: %w", rest.Error{Code: 409, Reason: "already exists"}),
			expected: true,
		},
		{
			name:     "rest.Error with 412 and already exists reason",
			err:      rest.Error{Code: 412, Reason: "This topic already exists"},
			expected: true,
		},
		{
			name:     "wrapped error with already exists message",
			err:      fmt.Errorf("wrapped: %w", errors.New("This topic already exists")),
			expected: true,
		},
		{
			name:     "plain error with already exists message",
			err:      errors.New("This topic already exists"),
			expected: true,
		},
		{
			name:     "unrelated error",
			err:      errors.New("permission denied"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAlreadyExist(tt.err)
			if result != tt.expected {
				t.Fatalf("IsAlreadyExist() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsPermissionDenied(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "rest.Error with 401",
			err:      rest.Error{Code: 401, Reason: "Need to authenticate to perform the request"},
			expected: true,
		},
		{
			name:     "rest.Error with 403",
			err:      rest.Error{Code: 403, Reason: "Don't have admin permission"},
			expected: true,
		},
		{
			name:     "rest.Error pointer with 403",
			err:      &rest.Error{Code: 403, Reason: "Don't have admin permission"},
			expected: true,
		},
		{
			name:     "wrapped rest.Error with 403",
			err:      fmt.Errorf("wrapped: %w", rest.Error{Code: 403, Reason: "Don't have admin permission"}),
			expected: true,
		},
		{
			// 404 is how Pulsar reports a namespace with no local policies, which is a
			// legitimate "nothing set" answer rather than a denial.
			name:     "rest.Error with 404",
			err:      rest.Error{Code: 404, Reason: "Namespace local-policies does not exist"},
			expected: false,
		},
		{
			name:     "rest.Error with 500",
			err:      rest.Error{Code: 500, Reason: "Internal server error"},
			expected: false,
		},
		{
			name:     "non-REST error",
			err:      errors.New("connection refused"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPermissionDenied(tt.err); got != tt.expected {
				t.Errorf("IsPermissionDenied(%v) = %v, want %v", tt.err, got, tt.expected)
			}
		})
	}
}
