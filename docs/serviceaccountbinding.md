# ServiceAccountBinding

The `ServiceAccountBinding` resource allows you to bind a service account to one or more pool members in the StreamNative Cloud, granting the service account access to the resources managed by those pool members, such as Pulsar Functions, Pulsar Connectors, Kafka Connect Connectors, Flink Jobs, etc.

## Example

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: ServiceAccountBinding
metadata:
  name: my-service-account-binding
  namespace: default
spec:
  serviceAccountName: my-service-account
  # Optional: Override APIServer connection from ServiceAccount
  # apiServerRef:
  #   name: custom-api-connection
  poolMemberRefs:
  - namespace: default
    name: pool-member-1
  - namespace: production
    name: pool-member-2
```

## Specification

| Field | Type | Description | Required |
| --- | --- | --- | --- |
| `spec.serviceAccountName` | string | Reference to a ServiceAccount in the same namespace as this binding object | Yes |
| `spec.apiServerRef.name` | string | Optional reference to a StreamNativeCloudConnection. If not provided, the connection from the ServiceAccount will be used | No |
| `spec.poolMemberRefs` | []PoolMemberReference | List of pool members this service account will be bound to | Yes |
| `spec.poolMemberRefs[].name` | string | Name of the pool member | Yes |
| `spec.poolMemberRefs[].namespace` | string | Namespace of the pool member | Yes |

## Status

| Field | Type | Description |
| --- | --- | --- |
| `status.conditions` | []Condition | Current state of the ServiceAccountBinding |
| `status.observedGeneration` | int64 | Last observed generation |

## Usage

Service account bindings provide a way to grant service accounts access to specific pool members in StreamNative Cloud. This allows for fine-grained control over which service accounts can access which resources.

### Creating a Service Account Binding

To create a service account binding, you need:

1. An existing ServiceAccount resource in the same namespace
2. One or more pool members to which the service account will be bound

#### Basic Example

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: ServiceAccountBinding
metadata:
  name: app-service-binding
  namespace: default
spec:
  serviceAccountName: app-service-account
  poolMemberRefs:
  - namespace: default
    name: production-pool-member
```

#### Advanced Example with Multiple Pool Members

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: ServiceAccountBinding
metadata:
  name: multi-cluster-binding
  namespace: default
spec:
  serviceAccountName: app-service-account
  # Optional: Use a different API connection than the one in ServiceAccount
  apiServerRef:
    name: custom-api-connection
  poolMemberRefs:
  - namespace: dev
    name: dev-pool-member
  - namespace: staging
    name: staging-pool-member
  - namespace: production
    name: production-pool-member
```

### Checking Binding Status

You can check the status of a service account binding using kubectl:

```bash
kubectl get serviceaccountbinding app-service-binding -n default
```

The `READY` column will show `True` when the binding is successfully established.

For more detailed status information:

```bash
kubectl describe serviceaccountbinding app-service-binding -n default
```
