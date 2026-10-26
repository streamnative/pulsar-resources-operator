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

// Command verify-ca-trust checks the operator image's default Linux trust store
// without contacting external servers or requiring Kubernetes credentials.
package main

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Direct issuers and both alternate root paths from the incident tracker
// (https://emudhra.we-have-a-problem.com/roots.json), plus C1/C3 exclusions
// preserving the customer's current trust-store baseline.
var distrustedRoots = map[string]string{
	"7dd78d5f4f13459a83dff9abbba62edbaf6f2d102bf257fd712f4d9f2746ed8d": "emSign Root TLS CA - G3",
	"f2edaaff3d9b701fb4d14c9f285f8c0bad90fc3bcc1fa07d47f4718edaba34a5": "emSign Root CS CA - G3",
	"86a1ecba089c4a8d3bbe2734c612ba341d813e043cf9e8a862cd5c57a36bbe6b": "emSign ECC Root CA - G3",
	"d7a7a0fb5d7e2731d771e9484ebcdef71d5f0c3e0a2948782bc83ee0ea699ef4": "AAA Certificate Services",
	// These two entries enforce the customer baseline separately from the
	// compromised intermediate keys covered by the four trust paths above.
	"125609aa301da0a249b97a8239cb6a34216f44dcac9f3954b14292f2e8c8608f": "emSign Root CA - C1",
	"bc4d809b15189d78db3e1d8cf4f9726a795da1643ca5f1358e1ddb0edc0d7eb3": "emSign ECC Root CA - C3",
}

// Six affected TLS intermediates each have eMudhra and AAA signing paths,
// plus two extra DV chains without Certificate Transparency records.
const expectedCompromisedChainCount = 6*2 + 2

func certificates(data []byte) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate
	for {
		block, rest := pem.Decode(data)
		if block == nil {
			break
		}
		data = rest
		if block.Type != "CERTIFICATE" {
			continue
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}
		certs = append(certs, cert)
	}
	if len(certs) == 0 {
		return nil, errors.New("no certificates found")
	}
	return certs, nil
}

func checkBundle(data []byte) error {
	certs, err := certificates(data)
	if err != nil {
		return err
	}
	for _, cert := range certs {
		digest := sha256.Sum256(cert.Raw)
		if name, found := distrustedRoots[hex.EncodeToString(digest[:])]; found {
			return fmt.Errorf("CA bundle still trusts %s", name)
		}
	}
	return nil
}

func verifyChain(path string, roots *x509.CertPool) error {
	data, err := os.ReadFile(path) // #nosec G304 -- Explicit local fixture path supplied by the offline test runner.
	if err != nil {
		return err
	}
	chain, err := certificates(data)
	if err != nil {
		return err
	}
	if len(chain) < 2 {
		return errors.New("fixture must contain a leaf and at least one intermediate certificate")
	}
	serverAuth := false
	for _, usage := range chain[0].ExtKeyUsage {
		if usage == x509.ExtKeyUsageServerAuth {
			serverAuth = true
			break
		}
	}
	if !serverAuth {
		return errors.New("fixture leaf must include the serverAuth extended key usage")
	}
	intermediates := x509.NewCertPool()
	for _, cert := range chain[1:] {
		intermediates.AddCert(cert)
	}
	_, err = chain[0].Verify(x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediates,
		// Freeze the fixture capture period so expiry cannot create a false pass.
		CurrentTime: time.Date(2026, time.October, 24, 0, 0, 0, 0, time.UTC),
		KeyUsages:   []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	})
	return err
}

func verify(bundle, fixtures string) error {
	data, err := os.ReadFile(bundle) // #nosec G304 -- Explicit local CA bundle selected by the verifier's CLI flag.
	if err != nil {
		return err
	}
	if err = checkBundle(data); err != nil {
		return err
	}
	// Check the actual Go default pool, including the distribution's hashed CA
	// directory, rather than only constructing a pool from the checked file.
	roots, err := x509.SystemCertPool()
	if err != nil {
		return err
	}
	paths, err := filepath.Glob(filepath.Join(fixtures, "*.we-have-a-problem.com.pem"))
	if err != nil {
		return err
	}
	if len(paths) != expectedCompromisedChainCount {
		return fmt.Errorf("expected %d compromised test chains, found %d", expectedCompromisedChainCount, len(paths))
	}
	for _, path := range paths {
		err = verifyChain(path, roots)
		if err == nil {
			return fmt.Errorf("default Go trust store still accepts %s", filepath.Base(path))
		}
		var unknown x509.UnknownAuthorityError
		if !errors.As(err, &unknown) {
			return fmt.Errorf("unexpected verification failure for %s: %w", filepath.Base(path), err)
		}
		fmt.Println("REJECT", filepath.Base(path))
	}
	if err = verifyChain(filepath.Join(fixtures, "valid-isrg-chain.pem"), roots); err != nil {
		return fmt.Errorf("normal ISRG certificate chain failed: %w", err)
	}
	fmt.Printf("PASS: %d compromised chains rejected; normal ISRG chain accepted\n", expectedCompromisedChainCount)
	return nil
}

func main() {
	bundle := flag.String("bundle", "/etc/ssl/certs/ca-certificates.crt", "OS CA bundle")
	fixtures := flag.String("chains-dir", "hack/verify-ca-trust/testdata", "offline certificate fixtures")
	flag.Parse()
	if err := verify(*bundle, *fixtures); err != nil {
		fmt.Fprintln(os.Stderr, strings.TrimSpace(err.Error()))
		os.Exit(1)
	}
}
