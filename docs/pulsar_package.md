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
  syncPolicy: IfNotPresent
```

This table lists specifications available for the `PulsarPackage` resource.

| Option | Description | Required or not |
| ---|---|--- |
| `packageURL` | The package URL. The information you provide creates a URL for a package, in the format <type>://<tenant>/<namespace>/<package name>/<version>. | Yes |
| `fileURL` | The file URL that can be download from. | Yes |
| `connectionRef` | The reference to a PulsarConnection. | Yes |
| `description` | The description of the package. | Optional |
| `contact` | The contact information of the package. | Optional |
| `properties` | A user-defined key/value map to store other information. | Optional |
| `lifecyclePolicy` | The resource lifecycle policy. Available options are `CleanUpAfterDeletion` and `KeepAfterDeletion`. By default, it is set to `CleanUpAfterDeletion`. | Optional |
| `syncPolicy` | The sync policy for package updates. Available options are `Always`, `IfNotPresent`, and `Never`. If not specified, defaults to `Always` if @latest tag is used in the packageURL, or `IfNotPresent` otherwise. | Optional |

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
  syncPolicy: IfNotPresent
```

Please note:
1. The behavior of updating package content depends on the `syncPolicy`:
   - With `Always`: The operator will check the file content and update if changed
   - With `IfNotPresent` or `Never`: The operator will not update existing package content
2. To ensure getting the latest content, either:
   - Use `@latest` tag in packageURL
   - Set `syncPolicy: Always`
   - Create a new package with a different version

## Delete PulsarPackage

You can delete the package by using the following command:

```shell
kubectl delete pulsarpackage test-pulsar-package
```

Please note that when you delete the package:
1. The actual package will still exist in Pulsar if the `lifecyclePolicy` is set to `KeepAfterDeletion`
2. The package content remains unchanged regardless of the `syncPolicy`


## Sync Policy

The `syncPolicy` field determines how the operator handles package updates:

- `Always`: The operator will check and update the package content on each reconciliation if the content has changed.
- `IfNotPresent`: The operator will only upload the package if it doesn't exist in Pulsar.
- `Never`: The operator will never upload the package if it already exists, and will fail if the package doesn't exist.

If `syncPolicy` is not specified, the operator will:
- Use `Always` if the packageURL contains `@latest` tag (e.g., `function://public/default/api-examples@latest`)
- Use `IfNotPresent` for all other cases (e.g., `function://public/default/api-examples@v3.2.3.3`)

## Managed Properties

The operator automatically manages several properties for each package under the prefix `pulsarpackages.resource.streamnative.io`. These properties are used to track package state and ownership:

1. **File Properties**
   - `pulsarpackages.resource.streamnative.io/file-checksum`: SHA-256 checksum of the package file
   - `pulsarpackages.resource.streamnative.io/file-size`: Size of the package file in bytes

2. **Resource Reference Properties**
   - `pulsarpackages.resource.streamnative.io/resource-namespace`: Kubernetes namespace of the PulsarPackage resource
   - `pulsarpackages.resource.streamnative.io/resource-name`: Name of the PulsarPackage resource
   - `pulsarpackages.resource.streamnative.io/resource-uid`: UID of the PulsarPackage resource
   - `pulsarpackages.resource.streamnative.io/cluster`: Name of the Pulsar cluster
   - `pulsarpackages.resource.streamnative.io/managed-by`: Always set to "pulsar-resources-operator"

These properties are automatically set and managed by the operator. When specifying custom properties in the `properties` field, any property with the prefix `pulsarpackages.resource.streamnative.io` will be ignored to prevent conflicts with the managed properties.

Example of viewing managed properties:
```shell
kubectl get pulsarpackage test-pulsar-package -o jsonpath='{.status.properties}'
```

## Cloud Storage Support

The operator supports downloading package files from various cloud storage providers through the `fileURL` field. The following URL schemes are supported:

- `s3://` - Amazon S3 and S3-compatible storage
- `gs://` - Google Cloud Storage
- `azblob://` - Azure Blob Storage
- `https://` - HTTPS URLs (default)
- `file://` - Local file system (for testing)

### Amazon S3

To use S3 URLs (e.g., `s3://my-bucket/path/to/package.jar`), configure the following environment variables in the operator deployment:

```yaml
env:
  - name: AWS_REGION
    value: "us-west-1"  # Replace with your region
  - name: AWS_ACCESS_KEY_ID
    valueFrom:
      secretKeyRef:
        name: aws-credentials
        key: access-key-id
  - name: AWS_SECRET_ACCESS_KEY
    valueFrom:
      secretKeyRef:
        name: aws-credentials
        key: secret-access-key
```

Additional S3 URL parameters:
- `endpoint` - Custom endpoint for S3-compatible services (e.g., MinIO)
- `disableSSL` - Set to `true` to use HTTP instead of HTTPS
- `s3ForcePathStyle` - Set to `true` to use path-style addressing instead of virtual hosted-style

Example S3 URL with parameters:
```yaml
fileURL: s3://my-bucket/package.jar?endpoint=minio.example.com&disableSSL=true&s3ForcePathStyle=true
```

### Google Cloud Storage

To use GCS URLs (e.g., `gs://my-bucket/path/to/package.jar`), you need to:

1. Create a service account with appropriate permissions (requires Storage Object Viewer role)
2. Obtain the service account key file
3. Configure environment variables in the operator deployment:

```yaml
env:
  - name: GOOGLE_APPLICATION_CREDENTIALS
    value: "/path/to/service-account-key.json"  # Key file must be mounted in the container
```

Alternatively, use Workload Identity (recommended):
```yaml
serviceAccountName: my-k8s-sa  # Configured with GCP Workload Identity
```

### Azure Blob Storage

To use Azure Blob Storage URLs (e.g., `azblob://my-container/path/to/package.jar`), configure the following environment variables:

```yaml
env:
  - name: AZURE_STORAGE_ACCOUNT
    value: "your-storage-account-name"
  # Using Storage Account Key authentication
  - name: AZURE_STORAGE_KEY
    valueFrom:
      secretKeyRef:
        name: azure-credentials
        key: storage-key
  # Or using SAS Token authentication
  - name: AZURE_STORAGE_SAS_TOKEN
    valueFrom:
      secretKeyRef:
        name: azure-credentials
        key: sas-token
```

Azure Blob Storage URL parameters:
- `protocol` - Set to `http` to use HTTP instead of HTTPS (for local testing)
- `domain` - Custom domain name (for local emulator or private endpoints)

Example Azure URL with parameters:
```yaml
fileURL: azblob://my-container/package.jar?protocol=http&domain=localhost:10001
```

### Permission Requirements

Ensure appropriate permissions are configured for different storage providers:

- **S3**: Requires `s3:GetObject` permission
- **GCS**: Requires `storage.objects.get` permission
- **Azure**: Requires `Storage Blob Data Reader` role or equivalent SAS token permissions

### References for Cloud Storage Support

Pulsar Resources Operator uses [https://gocloud.dev/howto/blob](https://gocloud.dev/howto/blob/#services) to provides cloud storage provider feature, please refer to Go CDK documentation as well.
