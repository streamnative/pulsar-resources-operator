# APIKey

The `APIKey` resource allows you to create and manage API keys for authenticating with the StreamNative Cloud API.

## Example

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: APIKey
metadata:
  name: my-apikey
  namespace: default
spec:
  apiServerRef:
    name: my-connection
  instanceName: my-pulsar-instance
  serviceAccountName: my-service-account
  description: "API Key for automation"
  expirationTime: "2025-12-31T23:59:59Z"
```

## Specification

| Field | Type | Description | Required |
| --- | --- | --- | --- |
| `spec.apiServerRef` | [LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#localobjectreference-v1-core) | Reference to a StreamNativeCloudConnection in the same namespace | Yes |
| `spec.instanceName` | string | Name of the instance this API key is for | No |
| `spec.serviceAccountName` | string | Name of the service account this API key is for | Yes |
| `spec.description` | string | User-defined description of the API key | No |
| `spec.expirationTime` | string | Timestamp defining when this API key will expire | No |
| `spec.revoke` | boolean | Indicates whether this API key should be revoked | No |
| `spec.encryptionKey` | object | Contains the public key used to encrypt the token | No |

## Status

| Field | Type | Description |
| --- | --- | --- |
| `status.conditions` | []Condition | Current state of the APIKey |
| `status.observedGeneration` | int64 | Last observed generation |
| `status.keyId` | string | Unique identifier for the API key |
| `status.issuedAt` | string | Timestamp when the key was issued |
| `status.expiresAt` | string | Timestamp when the key expires |
| `status.token` | string | The plaintext security token issued for the key (available only after creation) |
| `status.encryptedToken` | object | The encrypted security token if an encryption key was provided |
| `status.revokedAt` | string | Timestamp when the key was revoked, if applicable |

## Secret Management

For each APIKey resource, the operator creates and manages two types of Secrets:

1. **Private Key Secret**: Contains the RSA private key used for decrypting tokens
   - Name format: `<apikey-name>-private-key`
   - Contains key: `private-key`

2. **Token Secret**: Contains the decrypted API token for authentication
   - Name format: `<apikey-name>-token`
   - Contains key: `token`
   - Includes labels:
     - `resources.streamnative.io/apikey`: Name of the APIKey
     - `resources.streamnative.io/key-id`: Unique identifier of the APIKey

## Usage

API keys provide authentication credentials for service accounts to access the StreamNative Cloud API. They can be used in automated workflows, CI/CD pipelines, or any system that needs to interact with StreamNative Cloud resources.

### Creating an API Key

To create an API key, you need:

1. A StreamNativeCloudConnection resource configured with valid credentials
2. An existing service account for which the API key will be created

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: APIKey
metadata:
  name: my-automation-key
  namespace: default
spec:
  apiServerRef:
    name: my-connection
  instanceName: my-pulsar-instance
  serviceAccountName: my-service-account
  description: "API Key for CI/CD pipeline"
  expirationTime: "2026-01-01T00:00:00Z"
```

### Using API Keys in Applications

You can mount the token secret in your application pods:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: my-app
spec:
  containers:
  - name: app
    image: my-app-image
    volumeMounts:
    - name: apikey-volume
      mountPath: /etc/apikey
      readOnly: true
  volumes:
  - name: apikey-volume
    secret:
      secretName: my-automation-key-token
```

The application can then read the token from `/etc/apikey/token`.

### Revoking an API Key

To revoke an API key, update the `spec.revoke` field to `true`:

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: APIKey
metadata:
  name: my-automation-key
  namespace: default
spec:
  apiServerRef:
    name: my-connection
  instanceName: my-pulsar-instance
  serviceAccountName: my-service-account
  revoke: true
```

Once revoked, the API key can no longer be used for authentication.
