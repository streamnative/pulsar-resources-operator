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
	"crypto/md5" //nolint:gosec
	"encoding/hex"
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

// CalculateSecretKeyMd5 calculates the hash of the secret key.
func CalculateSecretKeyMd5(secret *corev1.Secret, key string) (string, error) {
	data, ok := secret.Data[key]
	if !ok {
		return "", fmt.Errorf("key %s not found in secret", key)
	}

	hasher := md5.New()
	if _, err := hasher.Write(data); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
