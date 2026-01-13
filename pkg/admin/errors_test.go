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
