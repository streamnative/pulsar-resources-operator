# PulsarTenant

## Overview

The `PulsarTenant` resource defines a tenant in a Pulsar cluster. It allows you to configure tenant-level settings such as admin roles, allowed clusters, and lifecycle policies.

## Specifications

| Field | Description | Required |
|-------|-------------|----------|
| `name` | The tenant name. | Yes |
| `connectionRef` | Reference to the PulsarConnection resource used to connect to the Pulsar cluster for this tenant. | Yes |
| `adminRoles` | List of authentication principals allowed to manage the tenant. | No |
| `allowedClusters` | List of clusters the tenant is allowed to access. If not specified, tenant has access to all clusters. | No |
| `lifecyclePolicy` | Determines whether to keep or delete the Pulsar tenant when the Kubernetes resource is deleted. Options: `CleanUpAfterDeletion`, `KeepAfterDeletion`. | No |
| `geoReplicationRefs` | List of references to PulsarGeoReplication resources, used to grant permissions to the tenant for geo-replication. | No |

## Allowed Clusters vs GeoReplicationRefs

The `allowedClusters` and `geoReplicationRefs` fields in the PulsarTenant resource serve different purposes and are used in different scenarios:

1. `allowedClusters`:
   - Use this when you want to restrict a tenant's access to specific clusters within a single Pulsar instance.
   - It's a simple list of cluster names that the tenant is allowed to access.
   - This is suitable for scenarios where you have multiple clusters in a single Pulsar deployment and want to control tenant access.
   - If not specified, the tenant has access to all clusters in the Pulsar instance.
   - Example use case: Limiting a tenant to only use clusters in certain regions for data locality or compliance reasons.

2. `geoReplicationRefs`:
   - Use this when setting up geo-replication between separate Pulsar instances for a tenant.
   - It references PulsarGeoReplication resources, which contain more detailed configuration for connecting to external Pulsar clusters.
   - This is appropriate for scenarios involving separate Pulsar deployments, possibly in different data centers or cloud providers.
   - It grants the tenant permission to replicate data between the specified Pulsar instances.
   - Example use case: Allowing a tenant to replicate their data between two independent Pulsar instances in different geographical locations.

When to use which:
- Use `allowedClusters` when you want to restrict a tenant's access within a single Pulsar instance.
- Use `geoReplicationRefs` when you need to set up geo-replication for a tenant across separate Pulsar instances.

These fields can be used independently or together, depending on your specific requirements for tenant management and data replication.

## Create A Pulsar Tenant

1. Define a tenant named `test-tenant` by using the YAML file and save the YAML file `tenant.yaml`. 

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarTenant
metadata:
  name: test-pulsar-tenant
  namespace: test
spec:
  name: test-tenant
  connectionRef:
    name: test-pulsar-connection
  adminRoles:
  - admin
  - ops
  # lifecyclePolicy: CleanUpAfterDeletion
```

2. Apply the YAML file to create the tenant.

```shell
kubectl apply -f tenant.yaml
```

3. Check the resource status. When column Ready is true, it indicates the resource is created successfully in the pulsar cluster

```shell
kubectl -n test get pulsartenant.resource.streamnative.io
```

```shell
NAME                 RESOURCE_NAME   GENERATION   OBSERVED_GENERATION   READY
test-pulsar-tenant      test-tenant      1                1             True
```

## Update A Pulsar Tenant

You can update the tenant by editing the `tenant.yaml` file and then applying it again using kubectl. This allows you to modify various settings of the Pulsar tenant.

Here's an example of how to update a tenant:

1. Edit the `tenant.yaml` file:

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarTenant
metadata:
  name: test-pulsar-tenant
  namespace: test
spec:
  name: test-tenant
  connectionRef:
    name: test-pulsar-connection
  adminRoles:
  - admin
  - ops
  allowedClusters:
  - cluster1
  - cluster2
```

2. Apply the updated YAML file:

```shell
kubectl apply -f tenant.yaml
```

3. Check the updated status:

```shell
kubectl -n test get pulsartenant.resource.streamnative.io
```

After applying changes, always verify the status of the update using:
```shell
kubectl -n test get pulsartenant.resource.streamnative.io test-pulsar-tenant
```
The `OBSERVED_GENERATION` should increment, and `READY` should become `True` when the update is complete.

Please note the following important points when updating a Pulsar tenant:

1. The `name` field is immutable and cannot be changed after the tenant is created. If you need to rename a tenant, you'll need to create a new one and migrate the resources.

2. Changes to `adminRoles` will affect who has administrative access to the tenant. Be cautious when modifying this field to avoid accidentally revoking necessary permissions.

3. Updating `allowedClusters` will change which Pulsar clusters the tenant can use. Ensure that any namespaces and topics within the tenant are compatible with the new cluster list.

4. If you want to change the `connectionRef`, ensure that the new PulsarConnection resource exists and is properly configured. Changing the `connectionRef` can have significant implications:

   - If the new PulsarConnection refers to the same Pulsar cluster (i.e., the admin and broker URLs are the same), the tenant will remain in its original location. The operator will simply use the new connection details to manage the existing tenant.

   - If the new PulsarConnection points to a different Pulsar cluster (i.e., different admin and broker URLs), the operator will attempt to create a new tenant with the same configuration in the new cluster. The original tenant in the old cluster will not be automatically deleted.

   Be cautious when changing the `connectionRef`, especially if it points to a new cluster, as this can lead to tenant duplication across clusters. Always verify the intended behavior and manage any cleanup of the old tenant if necessary.

5. If you're adding the `lifecyclePolicy` field, remember that it only affects what happens when the PulsarTenant resource is deleted, not the current state of the tenant.

6. Be cautious when updating tenant configurations, as changes may affect existing namespaces and topics within the tenant. It's recommended to test changes in a non-production environment first.

7. If you're using quota management features, remember that changes to the tenant configuration might affect resource allocation across the Pulsar cluster.


## Delete A Pulsar Tenant

```shell
kubectl -n test delete pulsartenant.resource.streamnative.io test-pulsar-tenant
```

Please be noticed, when you delete the tenant, the real tenant will still exist if the `lifecyclePolicy` is `KeepAfterDeletion`.

If you want to delete the tenant in the pulsar cluster, you can use the following command:

```shell
pulsarctl tenants delete test-tenant
```
