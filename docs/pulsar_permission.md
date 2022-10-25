# PulsarPermission

## Create PulsarPermission

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

This table lists specifications available for the `PulsarPermission` resource.
| Option | Description | Required or not |
| ---| --- |--- |
| `connectionRef` | The reference to a PulsarConnection. | Yes |
| `resourceType` | The type of the target resource which will be granted the permissions.. It can be at the namespace or topic level. | Yes |
| `resourceName` | The name of the target resource which will be granted the permissions. | Yes |
| `roles` | The list of roles which will be granted with the same permissions. | Yes
| `actions` | The list of actions to be granted. The options include `produce`,`consume`,`functions`, `sinks`, `sources`, `packages`. | Optional |
| `lifecyclePolicy` | The resource lifecycle policy, CleanUpAfterDeletion or KeepAfterDeletion, the default is KeepAfterDeletion | Optional |

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

## Update PulsarPermission

You can update the permission roles or actions by editing the permission.yaml, then apply if again. For example, add action `sources`.

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
  # lifecyclePolicy: CleanUpAfterDeletion
  roles:
  - ironman
  actions:
  - consume
  - produce
  - functions
  - sinks
  - sources
# lifecyclePolicy: CleanUpAfterDeletion
```
    
```shell
kubectl apply -f permission.yaml
```


## Delete PulsarPermission

```
kubectl -n test delete pulsarpermission.resource.streamnative.io test-pulsar-permission
```

Please be noticed, when you delete the permission, the real permission will still exist if the `lifecyclePolicy` is `KeepAfterDeletion`
