# ServiceAccountBinding

The `ServiceAccountBinding` resource allows you to bind a service account to a pool member in the StreamNative Cloud, granting the service account access to the resources managed by that pool member, such as Pulsar Functions, Pulsar Connectors, Kafka Connect Connectors, Flink Jobs, etc.

## Example

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: ServiceAccountBinding
metadata:
  name: my-service-account-binding
  namespace: default
spec:
  serviceAccountName: my-service-account
  poolMemberRef:
    namespace: default
    name: my-pool-member
```

## Specification

| Field | Type | Description | Required |
| --- | --- | --- | --- |
| `spec.serviceAccountName` | string | Reference to a ServiceAccount in the same namespace as this binding object | Yes |
| `spec.poolMemberRef.name` | string | Name of the pool member this service account will be bound to | Yes |
| `spec.poolMemberRef.namespace` | string | Namespace of the pool member | Yes |

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
2. A pool member to which the service account will be bound

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: ServiceAccountBinding
metadata:
  name: app-service-binding
  namespace: default
spec:
  serviceAccountName: app-service-account
  poolMemberRef:
    namespace: default
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
