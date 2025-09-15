# StreamNativeCloudConnection

## Overview

The `StreamNativeCloudConnection` resource defines a connection to the StreamNative Cloud API server. It allows you to configure authentication and connection details for interacting with StreamNative Cloud services.

## Specifications

| Field                           | Description                                                                                                     | Required |
|--------------------------------|-----------------------------------------------------------------------------------------------------------------|----------|
| `server`                        | The URL of the API server                                                                                       | Yes      |
| `auth.credentialsRef`          | Reference to the service account credentials secret                                                             | Yes      |
| `logs.serviceUrl`              | URL of the logging service. Required if logs configuration is specified.                                         | No*      |
| `organization`                  | The organization to use in the API server. If not specified, the connection name will be used                   | No       |

*Note: If `logs` configuration is specified, `serviceUrl` becomes required.

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

After applying changes, verify the status to ensure the connection is working properly.

## Delete Connection

To delete a StreamNativeCloudConnection resource:

```shell
kubectl delete streamnativecloudconnection test-connection
```

Note that deleting the connection will affect any resources that depend on it, such as ComputeWorkspaces or ComputeFlinkDeployments.
