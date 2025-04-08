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

package feature

import (
	"os"
	"strconv"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/component-base/featuregate"
)

var (

	// DefaultMutableFeatureGate is a mutable version of DefaultFeatureGate.
	DefaultMutableFeatureGate featuregate.MutableFeatureGate = featuregate.NewFeatureGate()

	// DefaultFeatureGate is a shared global FeatureGate.
	DefaultFeatureGate featuregate.FeatureGate = DefaultMutableFeatureGate
)

func init() {
	runtime.Must(DefaultMutableFeatureGate.Add(defaultFeatureGates))
}

// SetFeatureGates sets the provided feature gates.
func SetFeatureGates() error {
	envFlags := map[string]featuregate.Feature{
		"ALWAYS_UPDATE_PULSAR_RESOURCE": AlwaysUpdatePulsarResource,
	}

	m := map[string]bool{}
	for envVar, feature := range envFlags {
		if v, ok := os.LookupEnv(envVar); ok {
			val, err := strconv.ParseBool(v)
			if err != nil {
				return err
			}
			m[string(feature)] = val
		}
	}

	return DefaultMutableFeatureGate.SetFromMap(m)
}
