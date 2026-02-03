# Repository Guidelines

## Project Structure & Module Organization
- `api/` CRD types (e.g., `v1alpha1`), generated code via controller-tools.
- `controllers/` reconcilers and test suite (`suite_test.go`). Files follow `<resource>_controller.go`.
- `pkg/` shared libraries and utilities.
- `config/` Kustomize manifests, CRDs in `config/crd/bases`.
- `charts/` Helm chart; sync CRDs with `make copy-crds`.
- `tests/` integration/E2E (Ginkgo/Gomega) against a cluster.
- `scripts/` lint/format helpers; `bin/` tool cache.

## Build, Test, and Development Commands
- `make help` list targets.
- `make build` compile manager to `bin/manager`.
- `make run` run controller locally.
- `make test` run unit/integration tests with envtest; writes `cover.out`.
- `make manifests` and `make generate` regenerate CRDs and deepcopy code.
- `make install`/`make uninstall` apply/remove CRDs from current kubeconfig.
- `make deploy IMG=...` deploy to cluster; `make undeploy` remove.
- `make docker-build IMG=...` and `make docker-push` build/push images.
- `make copy-crds` sync CRDs into the Helm chart.

## Coding Style & Naming Conventions
- Go 1.24+. Format with `make fmt`; imports via `goimports`.
- Lint with `./scripts/lint.sh ./...` (golangci-lint). Fix license headers with `make license-fix`.
- Use canonical aliases: `corev1`, `metav1`, `apierrors`, `ctrl` (see `.golangci.yml importas`).
- Package and file names are lowercase; tests end with `_test.go`.

## Testing Guidelines
- Frameworks: Ginkgo v2 + Gomega; controller-runtime envtest.
- Run all tests with `make test`. Integration tests live under `tests/operator/` and expect a reachable Pulsar cluster; set `NAMESPACE`, `BROKER_NAME`, `PROXY_URL` as needed.
- Keep tests deterministic and fast; prefer envtest for controller logic.

## Commit & Pull Request Guidelines
- Conventional Commits for titles/messages (e.g., `feat(controller): add topic policy`, `fix(api): correct validation`).
- Install hooks with `make husky`; pre-commit runs license, lint, gofmt, and govet.
- PRs should: describe motivation, link issues, include test updates, and update `docs/` or `charts/` when user-visible.

## Security & Configuration Tips
- Never commit secrets. Pass tokens via env (e.g., `ACCESS_TOKEN` for image builds).
- After API changes, run `make manifests generate` and then `make copy-crds`.

## Agent-Specific Notes
- Keep changes scoped; do not modify generated files manually. Regenerate with the Make targets above.
