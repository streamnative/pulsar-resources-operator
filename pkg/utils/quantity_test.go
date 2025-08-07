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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("Quantity Utils", func() {
	Context("IsInfiniteQuantity", func() {
		It("should return true for -1 quantity", func() {
			q := resource.MustParse("-1")
			Expect(IsInfiniteQuantity(&q)).To(BeTrue())
		})

		It("should return false for normal quantities", func() {
			testCases := []string{
				"1Gi",
				"500Mi",
				"100M",
				"0",
			}

			for _, tc := range testCases {
				q := resource.MustParse(tc)
				Expect(IsInfiniteQuantity(&q)).To(BeFalse())
			}
		})

		It("should return false for nil quantity", func() {
			Expect(IsInfiniteQuantity(nil)).To(BeFalse())
		})
	})

	Context("QuantityToBytes", func() {
		It("should return -1 for infinite quantity", func() {
			q := resource.MustParse("-1")
			bytes := QuantityToBytes(&q)
			Expect(bytes).To(Equal(int64(-1)))
		})

		It("should return correct bytes for normal quantities", func() {
			testCases := []struct {
				input    string
				expected int64
			}{
				{"1Ki", 1024},
				{"1Mi", 1024 * 1024},
				{"1Gi", 1024 * 1024 * 1024},
				{"100", 100},
			}

			for _, tc := range testCases {
				q := resource.MustParse(tc.input)
				bytes := QuantityToBytes(&q)
				Expect(bytes).To(Equal(tc.expected))
			}
		})
	})

	Context("NewInfiniteQuantity", func() {
		It("should create a quantity with -1 value", func() {
			q := NewInfiniteQuantity()
			Expect(q).NotTo(BeNil())
			Expect(q.String()).To(Equal("-1"))
			Expect(IsInfiniteQuantity(q)).To(BeTrue())
		})
	})

	Context("ValidateQuantityValue", func() {
		It("should validate -1 as infinite quantity", func() {
			err := ValidateQuantityValue("-1")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should validate normal quantity formats", func() {
			testCases := []string{
				"1Gi",
				"500Mi",
				"100M",
				"1Ki",
				"0",
			}

			for _, tc := range testCases {
				err := ValidateQuantityValue(tc)
				Expect(err).NotTo(HaveOccurred())
			}
		})

		It("should return error for invalid quantity formats", func() {
			testCases := []string{
				"invalid",
				"abc",
				"1.5.2",
			}

			for _, tc := range testCases {
				err := ValidateQuantityValue(tc)
				Expect(err).To(HaveOccurred())
			}
		})
	})
})
