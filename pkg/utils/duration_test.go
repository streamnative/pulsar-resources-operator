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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Duration", func() {
	Context("Parse", func() {
		It("should parse normal duration formats correctly", func() {
			testCases := []struct {
				input    string
				expected time.Duration
			}{
				{"1h", time.Hour},
				{"30m", 30 * time.Minute},
				{"5s", 5 * time.Second},
				{"1d", 24 * time.Hour},
				{"2w", 14 * 24 * time.Hour},
			}

			for _, tc := range testCases {
				duration := Duration(tc.input)
				result, err := duration.Parse()
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(tc.expected))
			}
		})

		It("should parse infinite duration (-1) correctly", func() {
			duration := Duration("-1")
			result, err := duration.Parse()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(time.Duration(-1)))
		})

		It("should return error for invalid duration formats", func() {
			testCases := []string{
				"invalid",
				"",
				"abc",
				"-2",  // Only -1 is allowed for negative values
				"-5m", // Negative durations other than -1 are not allowed
			}

			for _, tc := range testCases {
				duration := Duration(tc)
				_, err := duration.Parse()
				Expect(err).To(HaveOccurred())
			}
		})
	})

	Context("IsInfinite", func() {
		It("should return true for -1 value", func() {
			duration := Duration("-1")
			Expect(duration.IsInfinite()).To(BeTrue())
		})

		It("should return false for normal duration values", func() {
			testCases := []string{
				"1h",
				"30m",
				"5s",
				"0",
			}

			for _, tc := range testCases {
				duration := Duration(tc)
				Expect(duration.IsInfinite()).To(BeFalse())
			}
		})
	})

	Context("ToSeconds", func() {
		It("should return -1 for infinite duration", func() {
			duration := Duration("-1")
			seconds, err := duration.ToSeconds()
			Expect(err).NotTo(HaveOccurred())
			Expect(seconds).To(Equal(int64(-1)))
		})

		It("should return correct seconds for normal durations", func() {
			testCases := []struct {
				input    string
				expected int64
			}{
				{"1h", 3600},
				{"30m", 1800},
				{"5s", 5},
			}

			for _, tc := range testCases {
				duration := Duration(tc.input)
				seconds, err := duration.ToSeconds()
				Expect(err).NotTo(HaveOccurred())
				Expect(seconds).To(Equal(tc.expected))
			}
		})

		It("should return error for invalid duration formats", func() {
			duration := Duration("invalid")
			_, err := duration.ToSeconds()
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Validate", func() {
		It("should validate normal duration formats", func() {
			testCases := []string{
				"1h",
				"30m",
				"5s",
				"-1",
			}

			for _, tc := range testCases {
				duration := Duration(tc)
				err := duration.Validate()
				Expect(err).NotTo(HaveOccurred())
			}
		})

		It("should return error for invalid duration formats", func() {
			testCases := []string{
				"invalid",
				"",
				"abc",
				"-2",
			}

			for _, tc := range testCases {
				duration := Duration(tc)
				err := duration.Validate()
				Expect(err).To(HaveOccurred())
			}
		})
	})

	Context("String", func() {
		It("should return original string representation", func() {
			testCases := []string{
				"1h",
				"30m",
				"-1",
			}

			for _, tc := range testCases {
				duration := Duration(tc)
				Expect(duration.String()).To(Equal(tc))
			}
		})
	})
})
