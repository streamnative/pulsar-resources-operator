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

package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

// GenerateRSAKeyPair generates a new RSA private and public key pair
func GenerateRSAKeyPair() (*rsa.PrivateKey, error) {
	// Generate a 2048-bit RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key pair: %w", err)
	}
	return privateKey, nil
}

// ExportPublicKeyAsPEM exports the public key in PEM format
func ExportPublicKeyAsPEM(privateKey *rsa.PrivateKey) (string, error) {
	publicKey := &privateKey.PublicKey
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %w", err)
	}

	publicKeyBlock := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	publicKeyPEM := pem.EncodeToMemory(publicKeyBlock)
	return string(publicKeyPEM), nil
}

// ExportPrivateKeyAsPEM exports the private key in PEM format
func ExportPrivateKeyAsPEM(privateKey *rsa.PrivateKey) string {
	privKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privKeyBytes,
	}
	privKeyPEM := pem.EncodeToMemory(privKeyBlock)
	return string(privKeyPEM)
}

// ImportPrivateKeyFromPEM imports a private key from PEM format
func ImportPrivateKeyFromPEM(privateKeyPEM string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return privateKey, nil
}

// EncodeToBase64 encodes a string to base64
func EncodeToBase64(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}

// DecodeFromBase64 decodes a base64 string
func DecodeFromBase64(data string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 string: %w", err)
	}
	return string(decoded), nil
}
