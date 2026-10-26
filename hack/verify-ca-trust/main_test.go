// Copyright 2024 StreamNative
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

package main

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBundleRejectsEachAffectedRoot(t *testing.T) {
	data, err := os.ReadFile("testdata/roots.pem")
	if err != nil {
		t.Fatal(err)
	}
	certs, err := certificates(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(certs) != 4 {
		t.Fatalf("expected four root fixtures, got %d", len(certs))
	}
	for _, cert := range certs {
		t.Run(cert.Subject.CommonName, func(t *testing.T) {
			if err := checkBundle(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})); err == nil {
				t.Fatal("accepted an affected root")
			}
		})
	}
}

func TestBundleRejectsExcludedCustomerRoots(t *testing.T) {
	data, err := os.ReadFile("testdata/excluded-emsign-roots.pem")
	if err != nil {
		t.Fatal(err)
	}
	certs, err := certificates(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(certs) != 2 {
		t.Fatalf("expected C1 and C3 fixtures, got %d certificates", len(certs))
	}
	for _, cert := range certs {
		t.Run(cert.Subject.CommonName, func(t *testing.T) {
			data := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
			if err := checkBundle(data); err == nil {
				t.Fatal("accepted a root excluded by the customer trust-store baseline")
			}
		})
	}
}

func TestBundleAcceptsUnaffectedCertificates(t *testing.T) {
	data, err := os.ReadFile("testdata/valid-isrg-chain.pem")
	if err != nil {
		t.Fatal(err)
	}
	if err := checkBundle(data); err != nil {
		t.Fatal(err)
	}
}

func TestBundleRejectsEmptyOrInvalidData(t *testing.T) {
	// An absent bundle must fail closed just like malformed certificate data.
	for _, data := range [][]byte{nil, []byte("not a PEM certificate"), pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("not DER")})} {
		if err := checkBundle(data); err == nil {
			t.Fatal("accepted an invalid bundle")
		}
	}
}

func unaffectedRoots(t *testing.T) *x509.CertPool {
	t.Helper()
	data, err := os.ReadFile("testdata/isrg-root-x1.pem")
	if err != nil {
		t.Fatal(err)
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(data) {
		t.Fatal("failed to load the unaffected ISRG root")
	}
	return roots
}

func TestVerifyChainRejectsCompromisedChains(t *testing.T) {
	roots := unaffectedRoots(t)
	paths, err := filepath.Glob("testdata/*.we-have-a-problem.com.pem")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != expectedCompromisedChainCount {
		t.Fatalf("expected %d fixture chains, found %d", expectedCompromisedChainCount, len(paths))
	}
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			err := verifyChain(path, roots)
			var unknown x509.UnknownAuthorityError
			// Parse failures, a missing intermediate, or a missing serverAuth EKU
			// must not masquerade as rejecting a compromised chain.
			if !errors.As(err, &unknown) {
				t.Fatalf("expected an untrusted issuer, got %v", err)
			}
		})
	}
}

func TestVerifyChainAcceptsValidChain(t *testing.T) {
	if err := verifyChain("testdata/valid-isrg-chain.pem", unaffectedRoots(t)); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyChainRejectsIncompleteFixture(t *testing.T) {
	data, err := os.ReadFile("testdata/valid-isrg-chain.pem")
	if err != nil {
		t.Fatal(err)
	}
	leaf, _ := pem.Decode(data)
	if leaf == nil {
		t.Fatal("fixture is missing its leaf")
	}
	path := filepath.Join(t.TempDir(), "leaf-only.pem")
	if err := os.WriteFile(path, pem.EncodeToMemory(leaf), 0600); err != nil { // #nosec G703 -- Fixed filename inside the private per-test t.TempDir.
		t.Fatal(err)
	}
	if err := verifyChain(path, unaffectedRoots(t)); err == nil || !strings.Contains(err.Error(), "at least one intermediate") {
		t.Fatalf("expected an incomplete-fixture error, got %v", err)
	}
}

func TestVerifyChainRejectsFixtureWithoutServerAuth(t *testing.T) {
	// A CA certificate is not a valid TLS leaf fixture: ISRG Root X1 has no
	// serverAuth EKU. Include a second certificate to exercise the EKU guard.
	root, err := os.ReadFile("testdata/isrg-root-x1.pem")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "not-a-server-leaf.pem")
	if err := os.WriteFile(path, append(root, root...), 0600); err != nil { // #nosec G703 -- Fixed filename inside the private per-test t.TempDir.
		t.Fatal(err)
	}
	if err := verifyChain(path, unaffectedRoots(t)); err == nil || !strings.Contains(err.Error(), "serverAuth") {
		t.Fatalf("expected an EKU-fixture error, got %v", err)
	}
}
