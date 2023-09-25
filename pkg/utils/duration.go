// Copyright 2022 StreamNative
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

	str2duration "github.com/xhit/go-str2duration/v2"
)

// Duration represents a elapsed time in string.
type Duration string

// Parse parses a duration from string.
// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h", "d", "w".
func (d *Duration) Parse() (time.Duration, error) {
	var res time.Duration
	res, err := str2duration.ParseDuration(string(*d))
	if err != nil {
		return res, err
	}
	return res, nil
}
