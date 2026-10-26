# Pulsar Resources Operator image CA trust regression

Support issue: https://github.com/streamnative/eng-support-tickets/issues/5137

The final Alpine 3.23 stage matches the builder's Alpine release and explicitly
installs and updates `ca-certificates` and
`ca-certificates-bundle` to at least `20261020-r0`, then regenerates the trust
store. Installing certificates in the Go builder does not update the runtime.
This fixes the Alpine `pulsar-resources-operator` image build in `Dockerfile`;
it does not change the separate Red Hat image build.

Run the production runtime-stage build and offline checks:

```sh
PLATFORM=linux/amd64 bash .ci/verify-image-ca-trust.sh
PLATFORM=linux/arm64 bash .ci/verify-image-ca-trust.sh
```

Docker must support the selected architecture (native or through QEMU).
The script compiles the standalone Go checker package with no external Go package
dependencies, builds the real runtime stage, and runs as UID 65532 with a read-only
filesystem and no network. No Kubernetes cluster or private-module credentials
are needed. The PR workflow runs both architectures with Go 1.26.9, matching the
production builder, so a newer Go certificate blocklist cannot mask a stale OS
bundle. The checker also explicitly tests certificate fingerprints in the bundle.

To check an existing image without building it:

```sh
PLATFORM=linux/amd64 bash .ci/verify-image-ca-trust.sh streamnative/pulsar-resources-operator:v0.21.0
```

This old image should fail because its trust bundle still includes excluded
emSign roots. The updated runtime must reject all 14 compromised chains, accept the
valid ISRG chain. Additional fingerprint checks keep C1/C3 out of the bundle as
requested by the customer's current baseline. Those two exclusions are separate
from the compromised intermediate keys and their four root trust paths.
No specific emSign root is required to remain trusted; upstream may legitimately
remove G1 or other roots in a future update. The ISRG chain is the positive control.

The Docker Hub release job runs this same checker against the assembled `$IMG`
after `make docker-build` and before `make docker-push`. A failed verification
therefore stops publication of the Alpine image.

Run the checker unit tests:

```sh
GOWORK=off go test ./hack/verify-ca-trust
```

## Public certificate fixtures

Copied from the certificate fixtures in `streamnative-ci` PR 8023, retrieved on
2026-10-24 from:

- https://emudhra.we-have-a-problem.com/chains.zip
  SHA-256: `7b949956955fc501a30f4d98d2f513d34c0bbbe92e29a968194c96e097ce2345`
- https://emudhra.we-have-a-problem.com/roots.json
- `valid-isrgrootx1.letsencrypt.org:443`, using `openssl s_client -showcerts`.
- ISRG Root X1 from the Alpine CA bundle, SHA-256:
  `96BCEC06264976F37460779ACF28C5A7CFE8A3C0AAE11A8FFCEE05C0BDDF08C6`.
- C1/C3 from the published `streamnative/pulsar-resources-operator:v0.21.0`
  bundle; SHA-256 fingerprints:
  `125609AA301DA0A249B97A8239CB6A34216F44DCAC9F3954B14292F2E8C8608F`
  and `BC4D809B15189D78DB3E1D8CF4F9726A795DA1643CA5F1358E1DDB0EDC0D7EB3`.

The upstream incident fixtures are Apache-2.0 licensed:
https://github.com/galadran/we-have-a-problem.com
Only public certificates are included. No private keys are included.

Verification fixes the date to 2026-10-24. Certificate expiry, DNS outages, and
network timeouts therefore cannot masquerade as successfully removing trust.
Fixtures must contain a leaf and an intermediate with explicit serverAuth EKU on
the leaf. Unit tests cover chain verification against an explicit unaffected
pool, valid chains, malformed fixture profiles, and each excluded root.
The Go default system pool is tested, including the hashed certificate directory.
Custom CA files supplied by users at deployment time remain their responsibility.

The checks cover both Alpine runtime architectures. The existing manager build
remains amd64; this change does not add multi-architecture manager publishing.
