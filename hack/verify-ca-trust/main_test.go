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
	"encoding/pem"
	"os"
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
	for _, data := range [][]byte{nil, []byte("not a PEM certificate"), pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("not DER")})} {
		if err := checkBundle(data); err == nil {
			t.Fatal("accepted an invalid bundle")
		}
	}
}
