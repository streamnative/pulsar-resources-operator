# PulsarPackage

## Create PulsarPackage

1. Define a package named `test-pulsar-package` by using the YAML file and save the YAML file `package.yaml`.
```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarPackage
metadata:
  name: test-pulsar-package
  namespace: default
spec:
  packageURL: function://public/default/api-examples@v3.2.3.3
  fileURL: https://github.com/freeznet/pulsar-functions-api-examples/raw/main/api-examples.jar
  connectionRef:
    name: "test-pulsar-connection"
  description: api-examples.jar
  lifecyclePolicy: CleanUpAfterDeletion
```

This table lists specifications available for the `PulsarPackage` resource.

| Option | Description                                                                                                                                     | Required or not |
| ---|-------------------------------------------------------------------------------------------------------------------------------------------------|--- |
| `packageURL` | The package URL. The information you provide creates a URL for a package, in the format <type>://<tenant>/<namespace>/<package name>/<version>. | Yes |
| `fileURL` | The file URL that can be download from.                                                                                                         | Yes |
| `connectionRef` | The reference to a PulsarConnection.                                                                                                             | Yes |
| `description` | The description of the package.                                                                                                                  | Optional |
| `contact` | The contact information of the package.                                                                                                          | Optional |
| `properties` | A user-defined key/value map to store other information.                                                                                        | Optional |                                                                                         
| `lifecyclePolicy` | The resource lifecycle policy. Available options are `CleanUpAfterDeletion` and `KeepAfterDeletion`. By default, it is set to `CleanUpAfterDeletion`. | Optional |

2. Apply the YAML file to create the package.

```shell
kubectl apply -f package.yaml
```

3. Check the resource status. When column Ready is true, it indicates the resource is created successfully in the pulsar cluster

```shell
kubectl get pulsarpackage
```

```
NAME                   RESOURCE_NAME   GENERATION   OBSERVED_GENERATION   READY
test-pulsar-package                    1            1                     True
```

## Update PulsarPackage

You can update the package by editing the package.yaml, then apply it again. For example, if you want to update the contact of the package, you can edit the package.yaml as follows:

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarPackage
metadata:
  name: test-pulsar-package
  namespace: default
spec:
  packageURL: function://public/default/api-examples@v3.2.3.3
  fileURL: https://github.com/freeznet/pulsar-functions-api-examples/raw/main/api-examples.jar
  connectionRef:
    name: "test-pulsar-connection"
  description: api-examples.jar
  contact: streamnative
  lifecyclePolicy: CleanUpAfterDeletion
```

Please be noted that update will not overwrite the package content even if `fileURL` is changed. To change the package content, you need to create a new package.

## Delete PulsarPackage

You can delete the package by using the following command:

```shell
kubectl delete pulsarpackage test-pulsar-package
```

Please be noticed, when you delete the package, the real package will still exist if the `lifecyclePolicy` is `KeepAfterDeletion`.
