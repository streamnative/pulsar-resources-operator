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

// Package crypto implements the crypto functions
package crypto

import (
	"crypto/rsa"
	"fmt"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwe"
)

// EncryptTokenToJWE encrypts a token using JWE with the provided public key
func EncryptTokenToJWE(token string, publicKey *rsa.PublicKey) (string, error) {
	// Encrypt the token using RSA-OAEP for key encryption and A256GCM for content encryption
	encrypted, err := jwe.Encrypt([]byte(token), jwe.WithKey(jwa.RSA_OAEP, publicKey))
	if err != nil {
		return "", fmt.Errorf("failed to encrypt token: %w", err)
	}

	return string(encrypted), nil
}

// DecryptJWEToken decrypts a JWE token using the provided private key
func DecryptJWEToken(jweToken string, privateKey *rsa.PrivateKey) (string, error) {
	// Decrypt the JWE token
	decrypted, err := jwe.Decrypt([]byte(jweToken), jwe.WithKey(jwa.RSA_OAEP, privateKey))
	if err != nil {
		return "", fmt.Errorf("failed to decrypt token: %w", err)
	}

	return string(decrypted), nil
}
