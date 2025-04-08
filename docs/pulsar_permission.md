# PulsarPermission

## Overview

The `PulsarPermission` resource is a custom resource in the Pulsar Resources Operator that allows you to manage access control for Pulsar resources declaratively using Kubernetes. It provides a way to grant or revoke permissions for specific roles on Pulsar resources such as namespaces and topics.

With `PulsarPermission`, you can:

1. Define fine-grained access control policies for your Pulsar resources.
2. Specify which roles have permissions to perform certain actions (like produce, consume, or manage functions) on specific namespaces or topics.
3. Manage permissions across your entire Pulsar cluster using Kubernetes-native tools and workflows.
4. Automate the process of granting and revoking permissions as part of your CI/CD pipeline or GitOps practices.

This resource is particularly useful for maintaining consistent and auditable access control across complex Pulsar deployments, ensuring that the right roles have the appropriate levels of access to the right resources.

## Specifications

The `PulsarPermission` resource has the following specifications:

| Field | Description | Required |
|-------|-------------|----------|
| `connectionRef` | Reference to the PulsarConnection resource used to connect to the Pulsar cluster. | Yes |
| `resourceType` | The type of Pulsar resource to which the permission applies. Can be either "namespace" or "topic". | Yes |
| `resourceName` | The name of the Pulsar resource. For namespaces, use the format "tenant/namespace". For topics, use the full topic name including persistence and namespace, e.g., "persistent://tenant/namespace/topic". | Yes |
| `roles` | A list of roles to which the permissions will be granted. | Yes |
| `actions` | A list of actions to be permitted. Can include "produce", "consume", "functions", "sinks", "sources", and "packages". | No |
| `lifecyclePolicy` | Determines whether to keep or delete the Pulsar permissions when the Kubernetes resource is deleted. Options: `CleanUpAfterDeletion`, `KeepAfterDeletion`. Default is `CleanUpAfterDeletion`. | No |

### Actions

The `actions` field can include one or more of the following (for more details, see the [Pulsar documentation on authorization and ACLs](https://pulsar.apache.org/docs/security-authorization/)):

- `produce`: Allows the role to produce messages to the specified resource.
- `consume`: Allows the role to consume messages from the specified resource.
- `functions`: Allows the role to manage Pulsar Functions on the specified resource.
- `sinks`: Allows the role to manage Pulsar IO Sinks on the specified resource.
- `sources`: Allows the role to manage Pulsar IO Sources on the specified resource.
- `packages`: Allows the role to manage Pulsar Packages on the specified resource.

If the `actions` field is omitted, no specific actions will be granted, effectively revoking all permissions for the specified roles on the resource.

### Lifecycle Policy

The `lifecyclePolicy` field determines what happens to the Pulsar permissions when the Kubernetes `PulsarPermission` resource is deleted:

- `CleanUpAfterDeletion` (default): The permissions will be removed from the Pulsar cluster when the Kubernetes resource is deleted.
- `KeepAfterDeletion`: The permissions will remain in the Pulsar cluster even after the Kubernetes resource is deleted.

For more information about lifecycle policies, refer to the [PulsarResourceLifeCyclePolicy documentation](pulsar_resource_lifecycle.md).

## Create a Pulsar Permission

1. Define a permission by using the YAML file and save the YAML file `permission.yaml`.t
This example grants the `ironman` with `consume`, `produce`, `functions`, and `sink` permissions on the namespace `test-tenant/testns`.
```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarPermission
metadata:
  name: test-pulsar-namespace-permission
  namespace: test
spec:
  connectionRef:
    name: "test-pulsar-connection"
  resourceType: namespace
  resourceName: test-tenant/testn1
  roles:
  - ironman
  actions:
  - consume
  - produce
  - functions
  - sinks
  # lifecyclePolicy: CleanUpAfterDeletion
```

2. Apply the YAML file to create the permission.

```shell
kubectl apply -f permission.yaml
```

3. Check the resource status. When column Ready is true, it indicates the resource is created successfully in the pulsar cluster.

```shell
kubectl -n test get pulsarpermission.resource.streamnative.io
```

```shell
NAME                                  RESOURCE NAME        RESOURCE TYPE   ROLES         ACTIONS                                               GENERATION   OBSERVED GENERATION   READY
pulsarpermission-sample-topic-error   test-tenant/testn1   namespace       ["ironman"]   ["consume","produce","functions","sinks","sources"]   2            2                     True
```

## Update A Pulsar Permission

You can update the permission by editing the `permission.yaml` file and then applying it again using `kubectl apply -f permission.yaml`. This allows you to modify various settings of the Pulsar permission.

Important notes about side effects of updating a Pulsar permission:

1. Changing permissions can immediately affect access to Pulsar resources. Be cautious when modifying permissions in production environments.

2. Removing permissions may disrupt ongoing operations for affected roles. Ensure all stakeholders are aware of permission changes.

3. Adding new permissions doesn't automatically grant access to existing data. Users may need to reconnect or refresh their sessions to utilize new permissions.

4. Modifying the `resourceType` or `resourceName` effectively creates a new permission set rather than updating the existing one. The old permissions will remain unless explicitly removed.

5. If you want to change the `connectionRef`, ensure that the new PulsarConnection resource exists and is properly configured. Changing the `connectionRef` can have significant implications:

   - If the new PulsarConnection refers to the same Pulsar cluster (i.e., the admin and broker URLs are the same), the permissions will remain in their original location. The operator will simply use the new connection details to manage the existing permissions.

   - If the new PulsarConnection points to a different Pulsar cluster (i.e., different admin and broker URLs), the operator will attempt to create new permissions with the same configuration in the new cluster. The original permissions in the old cluster will not be automatically deleted.

   Be cautious when changing the `connectionRef`, especially if it points to a new cluster, as this can lead to permission duplication across clusters. Always verify the intended behavior and manage any cleanup of the old permissions if necessary.

6. Updating `lifecyclePolicy` only affects future deletion behavior, not the current state of the permission. For more detailed information about the lifecycle policies and their implications, please refer to the [PulsarResourceLifeCyclePolicy documentation](pulsar_resource_lifecycle.md).

7. If you're using role-based access control (RBAC) in conjunction with these permissions, ensure that changes here don't conflict with RBAC policies.

Here's an example of how to update a permission:

1. Edit the `permission.yaml` file:

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarPermission
metadata:
  name: test-pulsar-namespace-permission
  namespace: test
spec: 
  connectionRef:
    name: "test-pulsar-connection"
  resourceType: namespace
  resourceName: test-tenant/testn1
  roles:
  - ironman
  actions:
  - consume # add consume action
  - produce # add produce action
  - functions # add functions action
  - sinks # add sinks action
  - sources # add sources action
  - packages # add packages action
```

2. Apply the updated YAML file:

```shell
kubectl apply -f permission.yaml
```

3. Check the resource status:

```shell
kubectl -n test get pulsarpermission.resource.streamnative.io
```

```shell
NAME                                  RESOURCE NAME        RESOURCE TYPE   ROLES         ACTIONS                                               GENERATION   OBSERVED GENERATION   READY
pulsarpermission-sample-topic-error   test-tenant/testn1   namespace       ["ironman"]   ["consume","produce","functions","sinks","sources"]   2            2                     True
```

## Delete a Pulsar Permission

```
kubectl -n test delete pulsarpermission.resource.streamnative.io test-pulsar-permission
```

Please be noticed, when you delete the permission, the real permission will still exist if the `lifecyclePolicy` is `KeepAfterDeletion`.