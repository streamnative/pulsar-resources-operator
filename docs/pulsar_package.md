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

1. Apply the YAML file to create the package.

```shell

