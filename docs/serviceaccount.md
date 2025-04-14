# ServiceAccount

The `ServiceAccount` resource allows you to create and manage service accounts in the StreamNative Cloud platform for programmatic access to resources.

## Example

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: ServiceAccount
metadata:
  name: my-service-account
  namespace: default
spec:
  apiServerRef:
    name: my-connection
```

## Specification

| Field | Type | Description | Required |
| --- | --- | --- | --- |
| `spec.apiServerRef` | [LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#localobjectreference-v1-core) | Reference to a StreamNativeCloudConnection in the same namespace | Yes |

## Status

| Field | Type | Description |
| --- | --- | --- |
| `status.conditions` | []Condition | Current state of the ServiceAccount |
| `status.observedGeneration` | int64 | Last observed generation |
| `status.privateKeyType` | string | Type of the private key data |
| `status.privateKeyData` | string | Private key data in base64 format |

## Secret Management

When a `ServiceAccount` resource is successfully created and contains credentials, the operator automatically creates a Kubernetes Secret to store these credentials:

- **Credentials Secret**: Contains the OAuth2 JSON credentials data
  - Name format: `<serviceaccount-name>-credentials`
  - Contains key: `credentials.json`
  - The Secret is automatically created when:
    - The ServiceAccount is in a Ready state
    - The `privateKeyType` is set to `TYPE_SN_CREDENTIALS_FILE`
    - The `privateKeyData` field contains valid base64-encoded credentials

This Secret is owned by the ServiceAccount resource and will be automatically deleted when the ServiceAccount is deleted.

## Usage

Service accounts provide a way to manage access to StreamNative Cloud resources programmatically. They are typically used for automation, CI/CD pipelines, or any system that needs to interact with StreamNative Cloud resources without using personal user credentials.

### Creating a Service Account

To create a service account, you need:

1. A StreamNativeCloudConnection resource configured with valid credentials
2. An instance in which to create the service account

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: ServiceAccount
metadata:
  name: automation-account
  namespace: default
spec:
  apiServerRef:
    name: my-connection
```

### Using Service Account Credentials in Applications

You can mount the credentials secret in your application pods:

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
    - name: credentials-volume
      mountPath: /etc/credentials
      readOnly: true
  volumes:
  - name: credentials-volume
    secret:
      secretName: automation-account-credentials
```

The application can then read the credentials from `/etc/credentials/credentials.json` and use them for OAuth2 authentication.

## Creating API Keys

After creating a service account, you can create API keys for it using the `APIKey` resource:

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: APIKey
metadata:
  name: my-api-key
  namespace: default
spec:
  apiServerRef:
    name: my-connection
  serviceAccountName: automation-account
  description: "API Key for automation"
  expirationTime: "2025-12-31T23:59:59Z"
  instanceName: my-pulsar-instance
```
