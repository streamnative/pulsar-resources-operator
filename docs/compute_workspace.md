# ComputeWorkspace

## Overview

The `ComputeWorkspace` resource defines a workspace in StreamNative Cloud for compute resources. It allows you to configure access to Pulsar clusters and compute pools for running Flink jobs.

## Specifications

| Field                    | Description                                                                                | Required |
|--------------------------|--------------------------------------------------------------------------------------------|----------|
| `apiServerRef`           | Reference to the StreamNativeCloudConnection resource for API server access                | Yes      |
| `pulsarClusterNames`     | List of Pulsar cluster names that the workspace will have access to                        | No       |
| `poolRef`                | Reference to the compute pool that the workspace will use                                   | No       |
| `useExternalAccess`      | Whether to use external access for the workspace                                           | No       |
| `flinkBlobStorage`       | Configuration for Flink blob storage                                                        | No       |

### PoolRef Structure

| Field       | Description                                                | Required |
|-------------|------------------------------------------------------------|----------|
| `namespace` | Namespace of the compute pool                               | No       |
| `name`      | Name of the compute pool                                    | Yes      |

### FlinkBlobStorage Structure

| Field    | Description                                                | Required |
|----------|------------------------------------------------------------|----------|
| `bucket` | Cloud storage bucket for Flink blob storage                 | Yes      |
| `path`   | Sub-path in the bucket (leave empty to use the whole bucket)| No       |

## Status

| Field                | Description                                                                                     |
|----------------------|-------------------------------------------------------------------------------------------------|
| `conditions`         | List of status conditions for the workspace                                                      |
| `observedGeneration` | The last observed generation of the resource                                                     |
| `workspaceId`        | The ID of the workspace in the StreamNative Cloud API server                                     |

## Example

1. Create a ComputeWorkspace resource:

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: ComputeWorkspace
metadata:
  name: test-operator-workspace
  namespace: default
spec:
  apiServerRef:
    name: test-connection
  pulsarClusterNames:
    - "test-pulsar"
  poolRef:
    name: shared
    namespace: streamnative
```

2. Apply the YAML file:

```shell
kubectl apply -f workspace.yaml
```

3. Check the workspace status:

```shell
kubectl get computeworkspace test-operator-workspace
```

The workspace is ready when the Ready condition is True:

```shell
NAME                    READY   AGE
test-operator-workspace True    1m
```

## Update Workspace

You can update the workspace by modifying the YAML file and reapplying it. Most fields can be updated, including:
- Pulsar cluster names
- Pool reference
- External access settings
- Flink blob storage configuration

After applying changes, verify the status to ensure the workspace is configured properly.

## Delete Workspace

To delete a ComputeWorkspace resource:

```shell
kubectl delete computeworkspace test-operator-workspace
```

Note that deleting the workspace will affect any resources that depend on it, such as ComputeFlinkDeployments. Make sure to handle any dependent resources appropriately before deletion. 