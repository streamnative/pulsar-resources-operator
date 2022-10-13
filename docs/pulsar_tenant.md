# PulsarTenant

## Create PulsarTenant

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

This table lists specifications available for the `PulsarTenant` resource.

| Option | Description | Required or not |
| ---| --- |--- |
| `name` | The tenant name. | Yes |
| `connectionRef` | The reference to a PulsarConnection. | Yes |
| `adminRoles` | The list of authentication principals allowed to manage the tenant. | Optional |
| `allowedClusters` | The list of allowed clusters. If no cluster is specified, the tenant will have access to all clusters. If you specify a list of clusters, ensure that these clusters exist.| Optional | 
| `lifecyclePolicy` | The resource lifecycle policy, CleanUpAfterDeletion or KeepAfterDeletion, the default is KeepAfterDeletion | Optional |

2. Apply the YAML file to create the tenant.

```shell
kubectl apply -f tenant.yaml
```

3. Check the resource status. When column Ready is true, it indicates the resource is created successfully in the pulsar cluster

```shell
kubectl -n test get pulsartenant
```

```shell
NAME                 RESOURCE_NAME   GENERATION   OBSERVED_GENERATION   READY
test-pulsar-tenant      test-tenant      1                1             True
```

## Update PulsarTenant

You can update the tenant by editing the tenant.yaml, then apply if again


## Delete PulsarTenant
```shell
kubectl -n test delete pulsartenant test-pulsar-tenant
```

Please be noticed, when you delete the tenant, the real tenant will still exist if the `lifecyclePolicy` is `KeepAfterDeletion`
