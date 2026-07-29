# StreamNativeCloudConnection

## Overview

The `StreamNativeCloudConnection` resource defines a connection to the StreamNative Cloud API server. It allows you to configure authentication and connection details for interacting with StreamNative Cloud services.

## Specifications

| Field                           | Description                                                                                                     | Required |
|--------------------------------|-----------------------------------------------------------------------------------------------------------------|----------|
| `server`                        | URL of the API server. Defaults to `https://api.streamnative.cloud`.                                             | No       |
| `auth.credentialsRef`          | Reference to the service account credentials secret                                                             | Yes      |
| `logs.serviceUrl`              | Logging service URL. Required by the CRD when `logs` is present, but not consumed by the current connection or resource clients. | Conditional |
| `organization`                  | Organization namespace used by StreamNative Cloud resource clients. Required before reconciling any dependent cloud resource. | Conditional |

The connection health check itself does not require `organization`, but `ComputeWorkspace`, `ComputeFlinkDeployment`, `Secret`, `ServiceAccount`, `ServiceAccountBinding`, `APIKey`, and `RoleBinding` controllers reject an empty value. There is no fallback to the Kubernetes resource name in the current implementation.

## Status

| Field                | Description                                                                                     |
|----------------------|-------------------------------------------------------------------------------------------------|
| `conditions`         | List of status conditions for the connection                                                     |
| `observedGeneration` | The last observed generation of the resource                                                     |
| `lastConnectedTime`  | Timestamp of the last successful connection to the API server                                    |

## Service Account Credentials Structure

The service account credentials secret should contain a `credentials.json` file with the following structure:

```json
{
  "type": "sn_service_account",
  "client_id": "<client-id>",
  "client_secret": "<client-secret>",
  "client_email": "<client-email>",
  "issuer_url": "<issuer-url>"
}
```

## Example

1. Create a service account credentials secret:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: test-credentials
  namespace: default
type: Opaque
stringData:
  credentials.json: |
    {
      "type": "sn_service_account",
      "client_secret": "client_secret",
      "client_email": "client-email",
      "issuer_url": "issuer_url",
      "client_id": "client-id"
    }
```

2. Create a StreamNativeCloudConnection resource:

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: StreamNativeCloudConnection
metadata:
  name: test-connection
  namespace: default
spec:
  server: https://api.streamnative.dev
  auth:
    credentialsRef:
      name: test-credentials
  organization: org
```

3. Apply the YAML files:

```shell
kubectl apply -f credentials.yaml
kubectl apply -f connection.yaml
```

4. Check the connection status:

```shell
kubectl get streamnativecloudconnection test-connection
```

The connection is ready when the Ready condition is True:

```shell
NAME             READY   AGE
test-connection  True    1m
```

## Update Connection

You can update the connection by modifying the YAML file and reapplying it. Most fields can be updated, including:
- Server URL
- Organization
- Credentials reference

`logs` is currently stored by Kubernetes but does not affect controller behavior.

After applying changes, verify the status to ensure the connection is working properly.

## Delete Connection

To delete a StreamNativeCloudConnection resource:

```shell
kubectl delete streamnativecloudconnection test-connection
```

The controller keeps its finalizer while dependent cloud resources in the same namespace still reference this connection. Delete or repoint those resources first. This includes direct references and `ComputeFlinkDeployment` references inherited through `ComputeWorkspace`.
