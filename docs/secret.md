# Secret

## Overview

The `Secret` resource defines a secret in StreamNative Cloud. It allows you to create and manage secrets in StreamNative Cloud that can be referenced and used by other resources, such as ComputeFlinkDeployment.

## Specifications

| Field | Description | Required |
|-------|-------------|----------|
| `apiServerRef` | Reference to the StreamNativeCloudConnection resource for API server access | Yes |
| `instanceName` | Name of the instance this secret is for (e.g. pulsar-instance) | No |
| `location` | Location of the secret | No |
| `data` | Secret data, values should be base64 encoded | No* |
| `secretRef` | Reference to a Kubernetes secret. When secretRef is set, it will be used to fetch the secret data, and data field will be ignored | No* |
| `poolMemberName` | Pool member to deploy the secret | No |
| `tolerations` | Tolerations for the secret | No |
| `type` | Used to facilitate programmatic handling of secret data | No |

*Note: Either `data` or `secretRef` must be specified.

### KubernetesSecretReference Structure

| Field | Description | Required |
|-------|-------------|----------|
| `namespace` | Namespace of the Kubernetes secret | Yes |
| `name` | Name of the Kubernetes secret | Yes |

### Toleration Structure

| Field | Description | Required |
|-------|-------------|----------|
| `key` | Taint key that the toleration applies to. Empty means match all taint keys | No |
| `operator` | Represents a key's relationship to the value. Valid operators are Exists and Equal. Defaults to Equal | No |
| `value` | Taint value the toleration matches to | No |
| `effect` | Indicates the taint effect to match. Empty means match all taint effects | No |

## Status

| Field | Description |
|-------|-------------|
| `conditions` | List of status conditions for the secret |
| `observedGeneration` | The last observed generation of the resource |

## Example

1. Create a Secret resource with direct data:

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: Secret
metadata:
  name: resource-operator-secret
  namespace: default
spec:
  apiServerRef:
    name: test-connection
  data:
    test-key: test-value
  instanceName: wstest
  location: us-central1
```

2. Create a Secret resource with Kubernetes Secret reference:

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: Secret
metadata:
  name: resource-operator-secret-docker-hub
  namespace: default
spec:
  apiServerRef:
    name: test-connection
  secretRef:
    name: regcred
    namespace: default
  instanceName: wstest
  location: us-central1
```

3. Apply the YAML file:

```shell
kubectl apply -f secret.yaml
```

4. Check the secret status:

```shell
kubectl get secret.resource.streamnative.io resource-operator-secret
```

The secret is ready when the Ready condition is True:

```shell
NAME                     READY   AGE
resource-operator-secret True    1m
```

## Update Secret

You can update the secret by modifying the YAML file and reapplying it. Most fields can be updated, including:
- Secret data
- Kubernetes secret reference
- Tolerations

After applying changes, verify the status to ensure the secret is configured properly.

## Delete Secret

To delete a Secret resource:

```shell
kubectl delete secret.resource.streamnative.io resource-operator-secret
```

Note that deleting the secret will affect any resources that depend on it, such as ComputeFlinkDeployments. Make sure to handle any dependent resources appropriately before deletion. 