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
	"encoding/json"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// ConvertMap converts a map[string]string to a map[string]interface{}
func ConvertMap(input map[string]string) map[string]interface{} {
	// Create an empty map[string]interface{}
	result := make(map[string]interface{})

	// Loop through each key-value pair in the input map
	for key, value := range input {
		// Assign the value to the result map with the same key
		result[key] = value
	}

	return result
}

// ConvertJSONToMapStringInterface converts a JSON object to a map[string]interface{}
func ConvertJSONToMapStringInterface(raw *apiextensionsv1.JSON) (map[string]interface{}, error) {
	// Create an empty map[string]interface{}
	result := make(map[string]interface{})

	// Unmarshal the raw JSON object into the result map
	if err := json.Unmarshal(raw.Raw, &result); err != nil {
		return nil, err
	}

	return result, nil
}
