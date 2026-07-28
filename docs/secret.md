# Secret

## Overview

The `Secret` resource defines a secret in StreamNative Cloud. It allows you to create and manage secrets in StreamNative Cloud that can be referenced and used by other resources, such as ComputeFlinkDeployment.

## Specifications

| Field | Description | Required |
|-------|-------------|----------|
| `apiServerRef` | Reference to the StreamNativeCloudConnection resource for API server access | Yes |
| `lifecyclePolicy` | Whether to delete the remote secret or keep it when the Kubernetes resource is deleted. Defaults to cleanup when omitted. | No |
| `instanceName` | Name of the instance this secret is for (e.g. pulsar-instance) | No |
| `location` | Location of the secret | No |
| `data` | Secret data as plain string values passed to the StreamNative Cloud API. | No* |
| `secretRef` | Reference to a Kubernetes Secret whose decoded `data` values are copied into this resource's `spec.data`. | No* |
| `poolMemberName` | Pool member to deploy the secret | No |
| `tolerations` | Tolerations for the secret | No |
| `type` | Used to facilitate programmatic handling of secret data | No |

*Note: Either `data` or `secretRef` must be specified. When both are present, `data` takes precedence.

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
| `effect` | Taint effect to match. Supported controller values include `NoSchedule`, `PreferNoSchedule`, `NoCleanup`, and `NoConnect`; empty matches all effects. | No |

### Kubernetes Secret Reference Behavior

On the first reconciliation with an empty `spec.data`, the controller reads the referenced Kubernetes Secret, decodes each byte value to a string, copies the result and Secret type into the custom resource spec, then sends that copied data to StreamNative Cloud.

This is a snapshot, not a live reference. After `spec.data` has been populated, later changes to the referenced Kubernetes Secret are not copied automatically because direct data takes precedence. Remove `spec.data` explicitly to import the reference again, for example with a JSON Patch:

```shell
kubectl -n default patch secret.resource.streamnative.io test-secret \
  --type=json -p='[{"op":"remove","path":"/spec/data"}]'
```

Because copied values are stored in the custom resource, protect access to both the source Kubernetes Secret and the `Secret.resource.streamnative.io` object.

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

Changing `secretRef` alone does not refresh copied data. Remove `spec.data` as shown above so the next reconciliation reads the new reference.

After applying changes, verify the status to ensure the secret is configured properly.

## Delete Secret

To delete a Secret resource:

```shell
kubectl delete secret.resource.streamnative.io resource-operator-secret
```

Note that deleting the secret will affect any resources that depend on it, such as ComputeFlinkDeployments. Make sure to handle any dependent resources appropriately before deletion.

Set `spec.lifecyclePolicy: KeepAfterDeletion` if you want to keep the remote StreamNative Cloud secret after the Kubernetes resource is removed.
