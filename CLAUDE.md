# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Pulsar Resources Operator is a Kubernetes operator that manages Apache Pulsar resources (tenants, namespaces, topics, permissions, functions, sinks, sources, etc.) using Custom Resource Definitions. Built with the Operator SDK and controller-runtime framework.

## Common Commands

### Build and Run
```bash
make build                    # Build manager binary
make run                      # Run controller locally (requires CRDs installed)
go run .                      # Alternative: run operator locally
```

### Code Generation
```bash
make generate                 # Generate DeepCopy methods for API types
make generate-internal        # Generate internal types and client code
make manifests                # Generate CRDs, RBAC, and webhook configurations
```

### Testing
```bash
make test                     # Run all tests with coverage

# Run single test
go test -v ./pkg/connection -run TestReconcileTenant

# E2E tests (separate module in tests/)
cd tests && ginkgo --trace --progress ./operator

# Run E2E tests against external Pulsar cluster
export ADMIN_SERVICE_URL=http://localhost:80
export NAMESPACE=pulsar
cd tests && ginkgo --trace --progress ./operator
```

### Linting and Formatting
```bash
make fmt                      # Run go fmt
make vet                      # Run go vet
make license-check            # Check license headers
make license-fix              # Fix license headers
```

### Deployment
```bash
make install                  # Install CRDs to cluster
make uninstall                # Remove CRDs from cluster
make deploy IMG=<image>       # Deploy operator to cluster
make undeploy                 # Remove operator from cluster
```

### Helm Chart
```bash
make copy-crds                # Sync CRDs to charts/pulsar-resources-operator/crds
```

## Architecture

### Entry Point and Controller Setup
- `main.go` - Initializes the controller manager and registers all reconcilers

### API Types (`api/v1alpha1/`)
Custom Resource types organized by domain:
- **Pulsar Resources**: `PulsarConnection`, `PulsarTenant`, `PulsarNamespace`, `PulsarTopic`, `PulsarPermission`, `PulsarFunction`, `PulsarSink`, `PulsarSource`, `PulsarPackage`, `PulsarGeoReplication`, `PulsarNSIsolationPolicy`
- **StreamNative Cloud**: `StreamNativeCloudConnection`, `ComputeWorkspace`, `ComputeFlinkDeployment`, `Secret`, `APIKey`, `ServiceAccount`, `ServiceAccountBinding`, `RoleBinding`

### Controllers (`controllers/`)
Two controller patterns:
1. **PulsarConnectionReconciler** - Central controller that manages Pulsar resources by watching `PulsarConnection` and all related resource types. Triggers reconciliation when any dependent resource changes.
2. **StreamNative Cloud Controllers** - Individual controllers for cloud resources (`APIServerConnectionReconciler`, `WorkspaceReconciler`, `FlinkDeploymentReconciler`, etc.) that share a `ConnectionManager` for API connectivity.

### Core Packages (`pkg/`)
- `connection/` - Pulsar resource reconciliation logic with sub-reconcilers for each resource type (tenant, namespace, topic, permission, etc.)
- `admin/` - Pulsar Admin API client wrapper
- `streamnativecloud/` - StreamNative Cloud API clients and converters
- `reconciler/` - Generic reconciler interface; `StatefulReconciler` provides annotation-based change detection using `SecretKeyHash` annotations
- `utils/` - Retry logic, event source, and helper utilities
- `feature/` - Feature gate management

### Reconciliation Pattern
The operator uses a hierarchical reconciliation pattern:
1. `PulsarConnection` serves as the parent resource linking to a Pulsar cluster
2. Child resources (tenants, namespaces, topics) reference a connection via `spec.connectionRef` with an indexed field (`spec.connectionRef.name`) for efficient lookup
3. When any child resource changes, it triggers the parent connection's reconciliation
4. The connection reconciler iterates through all associated resources and reconciles them in dependency order

**Reconciliation order (creation):** GeoReplication → Tenants → Namespaces → Topics → Permissions → Packages → Functions → Sinks → Sources → NSIsolationPolicies
*(Deletion uses reverse order)*

### Lifecycle Policy
Resources support `PulsarResourceLifeCyclePolicy`:
- `CleanUpAfterDeletion` (default) - Delete Pulsar resource when K8s CR is deleted
- `KeepAfterDeletion` - Preserve Pulsar resource after K8s CR deletion

## Conventions

- Follow [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) for PR titles
- All source files require Apache 2.0 license headers (use `make license-fix`)
- Filenames: lowercase with underscores for Go files, dashes for documentation
- Package names: avoid redundancy, match directory name
