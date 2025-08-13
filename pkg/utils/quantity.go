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

package utils

import (
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	// InfiniteQuantityValue represents the special value for infinite quantity
	InfiniteQuantityValue = "-1"
)

// IsInfiniteQuantity returns true if the quantity represents infinite size ("-1").
func IsInfiniteQuantity(q *resource.Quantity) bool {
	if q == nil {
		return false
	}
	return q.String() == InfiniteQuantityValue
}

// QuantityToBytes returns the quantity value in bytes.
// Returns -1 for infinite quantity, or the actual bytes for finite quantity.
func QuantityToBytes(q *resource.Quantity) int64 {
	if IsInfiniteQuantity(q) {
		return -1
	}

	// Convert quantity to bytes
	value, ok := q.AsInt64()
	if !ok {
		// If can't convert to int64, use approximation
		return int64(q.AsApproximateFloat64())
	}

	return value
}

// NewInfiniteQuantity creates a new quantity with infinite value ("-1").
func NewInfiniteQuantity() *resource.Quantity {
	q := resource.MustParse(InfiniteQuantityValue)
	return &q
}

// ValidateQuantityValue validates if a string is a valid quantity or "-1".
func ValidateQuantityValue(s string) error {
	if s == InfiniteQuantityValue {
		return nil
	}

	_, err := resource.ParseQuantity(s)
	return err
}
