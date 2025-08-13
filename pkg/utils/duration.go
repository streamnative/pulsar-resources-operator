// Copyright 2023 StreamNative
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
	"fmt"
	"time"

	str2duration "github.com/xhit/go-str2duration/v2"
)

// Duration represents a elapsed time in string.
// Supports standard duration formats (e.g., "1h", "30m", "5s") and special value "-1" for infinite duration.
type Duration string

const (
	// InfiniteDurationValue represents the special value for infinite duration
	InfiniteDurationValue = "-1"
)

// Parse parses a duration from string.
// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h", "d", "w".
// Special value "-1" represents infinite duration and returns -1 nanosecond.
func (d *Duration) Parse() (time.Duration, error) {
	s := string(*d)

	// Handle infinite duration special case
	if s == InfiniteDurationValue {
		return time.Duration(-1), nil
	}

	// Parse normal duration
	res, err := str2duration.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid duration format '%s': %w", s, err)
	}

	// Ensure duration is positive for non-infinite values
	if res < 0 {
		return 0, fmt.Errorf("duration must be positive or -1 for infinite, got: %s", s)
	}

	return res, nil
}

// IsInfinite returns true if the duration represents infinite duration ("-1").
func (d *Duration) IsInfinite() bool {
	return string(*d) == InfiniteDurationValue
}

// ToSeconds returns the duration in seconds.
// Returns -1 for infinite duration, or the actual seconds for finite duration.
func (d *Duration) ToSeconds() (int64, error) {
	if d.IsInfinite() {
		return -1, nil
	}

	duration, err := d.Parse()
	if err != nil {
		return 0, err
	}

	return int64(duration.Seconds()), nil
}

// String returns the string representation of the duration.
func (d Duration) String() string {
	return string(d)
}

// Validate validates the duration format.
// Accepts standard duration formats and the special value "-1".
func (d *Duration) Validate() error {
	_, err := d.Parse()
	return err
}
