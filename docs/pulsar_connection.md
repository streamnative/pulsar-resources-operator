# PulsarConnection

## Overview

The `PulsarConnection` resource defines the connection details for a Pulsar cluster. It can be used to configure various connection parameters including service URLs, authentication methods, and cluster information.

## Specifications

| Field | Description | Required | Version |
|-------|-------------|----------|---------|
| `adminServiceURL` | The admin service URL of the Pulsar cluster (e.g., `http://cluster-broker.test.svc.cluster.local:8080`). | No | All |
| `adminServiceSecureURL` | The admin service URL for secure connection (HTTPS) to the Pulsar cluster (e.g., `https://cluster-broker.test.svc.cluster.local:443`). | No | ≥ 0.3.0 |
| `brokerServiceURL` | The broker service URL of the Pulsar cluster (e.g., `pulsar://cluster-broker.test.svc.cluster.local:6650`). | No | ≥ 0.3.0 |
| `brokerServiceSecureURL` | The broker service URL for secure connection (TLS) to the Pulsar cluster (e.g., `pulsar+ssl://cluster-broker.test.svc.cluster.local:6651`). | No | ≥ 0.3.0 |
| `clusterName` | The Pulsar cluster name. Use `pulsar-admin clusters list` to retrieve. Required for configuring Geo-Replication. | No | ≥ 0.3.0 |
| `authentication` | Authentication configuration. Required when authentication is enabled for the Pulsar cluster. Supports JWT Token and OAuth2 methods. | No | All |
| `brokerClientTrustCertsFilePath` | The file path to the trusted TLS certificate for outgoing connections to Pulsar brokers. Used for TLS verification. | No | ≥ 0.3.0 |

Note: Fields marked with version ≥ 0.3.0 are only available in that version and above.
 
## Authentication Methods

The `authentication` field supports two methods: JWT Token and OAuth2. Each method can use either a Kubernetes Secret reference or a direct value.

### Specification

| Field | Description | Type | Required |
|-------|-------------|------|----------|
| `token` | JWT Token authentication configuration | `ValueOrSecretRef` | No |
| `oauth2` | OAuth2 authentication configuration | `PulsarAuthenticationOAuth2` | No |

#### ValueOrSecretRef

| Field | Description | Type | Required |
|-------|-------------|------|----------|
| `value` | Direct string value | `*string` | No |
| `secretRef` | Reference to a Kubernetes Secret | `*SecretKeyRef` | No |
| `file` | Local file path whose contents are used as the value | `*string` | No |
| `file` | Local file path whose contents are used as the value | `*string` | No |

#### SecretKeyRef

| Field | Description | Type | Required |
|-------|-------------|------|----------|
| `name` | Name of the Kubernetes Secret | `string` | Yes |
| `key` | Key in the Kubernetes Secret | `string` | Yes |

#### PulsarAuthenticationOAuth2

| Field | Description | Type | Required |
|-------|-------------|------|----------|
| `issuerEndpoint` | URL of the OAuth2 authorization server | `string` | Yes |
| `clientID` | OAuth2 client identifier | `string` | Yes |
| `audience` | Intended recipient of the token | `string` | Yes |
| `key` | Client secret or path to JSON credentials file | `ValueOrSecretRef` | Yes |
| `scope` | Requested permissions from the OAuth2 server | `string` | No |

Note: Only one authentication method (either `token` or `oauth2`) should be specified at a time.

### JWT Token Authentication

JWT Token authentication can be configured using a direct value, a local file path, or a Kubernetes Secret reference.

#### Using a direct value

To use JWT Token authentication with a direct value, you can set the `token` field to the base64-encoded JWT token.

```yaml
authentication:
  token:
    value: <base64-encoded JWT token>
``` 

#### Using a Kubernetes Secret reference

To use JWT Token authentication with a Kubernetes Secret reference, you need to create a Kubernetes Secret containing the JWT token. The secret should have the following structure:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: <secret-name>
type: Opaque
stringData:
  <key-name>: <base64-encoded JWT token>
```

The `<key-name>` can be any name. It will be referenced in the `PulsarConnection` resource with the `token.secretRef` field.

```yaml
authentication:
  token:
    secretRef:
      name: <secret-name>
      key: <key-name>
```

### OAuth2 Authentication

OAuth2 authentication can be configured using a direct value, a local file path, or a Kubernetes Secret reference.

#### Using a direct value

To use OAuth2 authentication with a direct value, you can set the `issuerEndpoint`, `clientID`, `audience`, and `key` fields to the OAuth2 configuration.

```yaml
authentication:
  oauth2:
    issuerEndpoint: https://auth.streamnative.cloud
    clientID: <client-id>
    audience: urn:sn:pulsar:sndev:us-west
    key:
      value: |
        {
          "type":"sn_service_account",
          "client_id":"<client-id>",
          "grant_type":"client_credentials",
          "client_secret":"<client-secret>",
          "issuer_url":"https://auth.streamnative.cloud"
        }
    scope: <scope>
```

#### Using a Kubernetes Secret reference

To use OAuth2 authentication with a Kubernetes Secret reference, you need to create a Kubernetes Secret containing the OAuth2 configuration. The secret should have the following structure:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: <secret-name>
type: Opaque
stringData:
  <key-name>: |
    {
      "type":"sn_service_account",
      "client_id":"<client-id>",
      "grant_type":"client_credentials",
      "client_secret":"<client-secret>",
      "issuer_url":"https://auth.streamnative.cloud"
    } 
```

The `<key-name>` should contain the OAuth2 configuration.

```yaml
authentication:
  oauth2:
    issuerEndpoint: https://auth.streamnative.cloud
    clientID: <client-id>
    audience: urn:sn:pulsar:sndev:us-west
    key:
      secretRef:
        name: <secret-name>
        key: <key-name> 
    scope: <scope>
```

## PlainText Connection vs TLS Connection

The `PulsarConnection` resource supports both plaintext and TLS connections. Plaintext connections are used for non-secure connections to the Pulsar cluster, while TLS connections are used for secure connections with TLS enabled.

### Plaintext Connection

To create a plaintext connection, you need to set the `adminServiceURL`, `brokerServiceURL`, and `clusterName` fields.  

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarConnection
metadata:
  name: pulsar-connection
spec:
  adminServiceURL: http://pulsar-sn-platform-broker.test.svc.cluster.local:8080
  brokerServiceURL: pulsar://pulsar-sn-platform-broker.test.svc.cluster.local:6650
  clusterName: pulsar-cluster
```

### TLS Connection

To create a TLS connection, you need to set the `adminServiceSecureURL`, `brokerServiceSecureURL`, and `clusterName` fields.
## Overview

The `PulsarConnection` resource defines the connection details for a Pulsar cluster. It can be used to configure various connection parameters including service URLs, authentication methods, and cluster information.

## Specifications

| Field | Description | Required | Version |
|-------|-------------|----------|---------|
| `adminServiceURL` | The admin service URL of the Pulsar cluster (e.g., `http://cluster-broker.test.svc.cluster.local:8080`). | No | All |
| `adminServiceSecureURL` | The admin service URL for secure connection (HTTPS) to the Pulsar cluster (e.g., `https://cluster-broker.test.svc.cluster.local:443`). | No | ≥ 0.3.0 |
| `brokerServiceURL` | The broker service URL of the Pulsar cluster (e.g., `pulsar://cluster-broker.test.svc.cluster.local:6650`). | No | ≥ 0.3.0 |
| `brokerServiceSecureURL` | The broker service URL for secure connection (TLS) to the Pulsar cluster (e.g., `pulsar+ssl://cluster-broker.test.svc.cluster.local:6651`). | No | ≥ 0.3.0 |
| `clusterName` | The Pulsar cluster name. Use `pulsar-admin clusters list` to retrieve. Required for configuring Geo-Replication. | No | ≥ 0.3.0 |
| `authentication` | Authentication configuration. Required when authentication is enabled for the Pulsar cluster. Supports JWT Token and OAuth2 methods. | No | All |
| `brokerClientTrustCertsFilePath` | The file path to the trusted TLS certificate for outgoing connections to Pulsar brokers. Used for TLS verification. | No | ≥ 0.3.0 |

Note: Fields marked with version ≥ 0.3.0 are only available in that version and above.
 
## Authentication Methods

The `authentication` field supports two methods: JWT Token and OAuth2. Each method can use either a Kubernetes Secret reference or a direct value.

### Specification

| Field | Description | Type | Required |
|-------|-------------|------|----------|
| `token` | JWT Token authentication configuration | `ValueOrSecretRef` | No |
| `oauth2` | OAuth2 authentication configuration | `PulsarAuthenticationOAuth2` | No |

#### ValueOrSecretRef

| Field | Description | Type | Required |
|-------|-------------|------|----------|
| `value` | Direct string value | `*string` | No |
| `secretRef` | Reference to a Kubernetes Secret | `*SecretKeyRef` | No |

#### SecretKeyRef

| Field | Description | Type | Required |
|-------|-------------|------|----------|
| `name` | Name of the Kubernetes Secret | `string` | Yes |
| `key` | Key in the Kubernetes Secret | `string` | Yes |

#### PulsarAuthenticationOAuth2

| Field | Description | Type | Required |
|-------|-------------|------|----------|
| `issuerEndpoint` | URL of the OAuth2 authorization server | `string` | Yes |
| `clientID` | OAuth2 client identifier | `string` | Yes |
| `audience` | Intended recipient of the token | `string` | Yes |
| `key` | Client secret or path to JSON credentials file | `ValueOrSecretRef` | Yes |
| `scope` | Requested permissions from the OAuth2 server | `string` | No |

Note: Only one authentication method (either `token` or `oauth2`) should be specified at a time.

### JWT Token Authentication

JWT Token authentication can be configured using a direct value, a local file path, or a Kubernetes Secret reference.

#### Using a direct value

To use JWT Token authentication with a direct value, you can set the `token` field to the base64-encoded JWT token.

```yaml
authentication:
  token:
    value: <base64-encoded JWT token>
``` 

#### Using a Kubernetes Secret reference

To use JWT Token authentication with a Kubernetes Secret reference, you need to create a Kubernetes Secret containing the JWT token. The secret should have the following structure:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: <secret-name>
type: Opaque
stringData:
  <key-name>: <base64-encoded JWT token>
```

The `<key-name>` can be any name. It will be referenced in the `PulsarConnection` resource with the `token.secretRef` field.

```yaml
authentication:
  token:
    secretRef:
      name: <secret-name>
      key: <key-name>
```

### OAuth2 Authentication

OAuth2 authentication can be configured using a direct value, a local file path, or a Kubernetes Secret reference.

#### Using a direct value

To use OAuth2 authentication with a direct value, you can set the `issuerEndpoint`, `clientID`, `audience`, and `key` fields to the OAuth2 configuration.

```yaml
authentication:
  oauth2:
    issuerEndpoint: https://auth.streamnative.cloud
    clientID: <client-id>
    audience: urn:sn:pulsar:sndev:us-west
    key:
      value: |
        {
          "type":"sn_service_account",
          "client_id":"<client-id>",
          "grant_type":"client_credentials",
          "client_secret":"<client-secret>",
          "issuer_url":"https://auth.streamnative.cloud"
        }
    scope: <scope>
```

#### Using a Kubernetes Secret reference

To use OAuth2 authentication with a Kubernetes Secret reference, you need to create a Kubernetes Secret containing the OAuth2 configuration. The secret should have the following structure:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: <secret-name>
type: Opaque
stringData:
  <key-name>: |
    {
      "type":"sn_service_account",
      "client_id":"<client-id>",
      "grant_type":"client_credentials",
      "client_secret":"<client-secret>",
      "issuer_url":"https://auth.streamnative.cloud"
    } 
```

The `<key-name>` should contain the OAuth2 configuration.

```yaml
authentication:
  oauth2:
    issuerEndpoint: https://auth.streamnative.cloud
    clientID: <client-id>
    audience: urn:sn:pulsar:sndev:us-west
    key:
      secretRef:
        name: <secret-name>
        key: <key-name> 
    scope: <scope>
```

## PlainText Connection vs TLS Connection

The `PulsarConnection` resource supports both plaintext and TLS connections. Plaintext connections are used for non-secure connections to the Pulsar cluster, while TLS connections are used for secure connections with TLS enabled.

### Plaintext Connection

To create a plaintext connection, you need to set the `adminServiceURL`, `brokerServiceURL`, and `clusterName` fields.  

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarConnection
metadata:
  name: pulsar-connection
spec:
  adminServiceURL: http://pulsar-sn-platform-broker.test.svc.cluster.local:8080
  brokerServiceURL: pulsar://pulsar-sn-platform-broker.test.svc.cluster.local:6650
  clusterName: pulsar-cluster
```

### TLS Connection

To create a TLS connection, you need to set the `adminServiceSecureURL`, `brokerServiceSecureURL`, and `clusterName` fields.

```yaml 
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarConnection
metadata:
  name: pulsar-connection-tls
  namespace: test
spec:
  adminServiceSecureURL: https://pulsar-sn-platform-broker.test.svc.cluster.local:443
  brokerServiceSecureURL: pulsar+ssl://pulsar-sn-platform-broker.test.svc.cluster.local:6651
```

## Create A Pulsar Connection

1. Create a YAML file for the Pulsar Connection.

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarConnection
metadata:
  name: pulsar-connection
spec:
  adminServiceURL: http://pulsar-sn-platform-broker.test.svc.cluster.local:8080
  brokerServiceURL: pulsar://pulsar-sn-platform-broker.test.svc.cluster.local:6650
  clusterName: pulsar-cluster
```

2. Apply the YAML file to create the Pulsar Connection. 

```shell
kubectl  apply -f connection.yaml
```

3. Check the resource status.

```shell
kubectl -n test get pulsarconnection.resource.streamnative.io
```

```shell
NAME                     ADMIN_SERVICE_URL                                        GENERATION   OBSERVED_GENERATION   READY
test-pulsar-connection   http://ok-sn-platform-broker.test.svc.cluster.local:8080 1            1                     True
```

## Update A Pulsar Connection

You can update the connection by editing the connection.yaml, then apply it again. For example, if pulsar cluster doesn’t setup the authentication, then you don’t need the authentication part in the spec

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarConnection
metadata:
  name: test-pulsar-connection
  namespace: test
spec:
  adminServiceURL: http://test-pulsar-sn-platform-broker.test.svc.cluster.local:8080
```

```shell
kubectl apply -f connection.yaml
```

## Delete A Pulsar Connection

```shell
kubectl -n test delete pulsarconnection.resource.streamnative.io test-pulsar-connection
```

Please be noticed, because the Pulsar Resources Operator are using the connection to manage pulsar resources, If you delete the pulsar connection, it will only be deleted after the resources CRs are deleted


## More PulsarConnection Examples

### PlainText Connection

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarConnection
metadata:
  name: pulsar-connection
spec:
  adminServiceURL: http://pulsar-sn-platform-broker.test.svc.cluster.local:8080
  brokerServiceURL: pulsar://pulsar-sn-platform-broker.test.svc.cluster.local:6650
  clusterName: pulsar-cluster
```

### TLS Connection

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarConnection
metadata:
  name: pulsar-connection-tls
  namespace: test
spec:
  adminServiceSecureURL: https://pulsar-sn-platform-broker.test.svc.cluster.local:443
  brokerServiceSecureURL: pulsar+ssl//pulsar-sn-platform-broker.test.svc.cluster.local:6651
  clusterName: pulsar-cluster
```

### PlainText Connection with JWT Token Authentication with Secret

```bash
kubectl create secret generic test-pulsar-sn-platform-vault-secret-env-injection --from-literal=brokerClientAuthenticationParameters=eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJKb2UifQ.ipevRNuRP6HflG8cFKnmUPtypruRC4fb1DWtoLL62SY
```

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarConnection
metadata:
  name: pulsar-connection-jwt-secret
spec:
  adminServiceURL: http://pulsar-sn-platform-broker.test.svc.cluster.local:8080
  brokerServiceURL: pulsar://pulsar-sn-platform-broker.test.svc.cluster.local:6650
  clusterName: pulsar-cluster
  authentication:
    token:
      # Use a Kubernetes Secret to store the JWT Token. https://kubernetes.io/docs/concepts/configuration/secret/
      # Secret data field have to be base64-encoded strings. https://kubernetes.io/docs/concepts/configuration/secret/#restriction-names-data
      secretRef:
        name: test-pulsar-sn-platform-vault-secret-env-injection
        key: brokerClientAuthenticationParameters
```

### PlainText Connection with JWT Token Authentication with value

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarConnection
metadata:
  name: pulsar-connection-jwt-value
spec:
  adminServiceURL: http://pulsar-sn-platform-broker.test.svc.cluster.local:8080
  brokerServiceURL: pulsar://pulsar-sn-platform-broker.test.svc.cluster.local:6650
  clusterName: pulsar-cluster
  authentication:
    token:
      # Use the JWT Token raw data as the token value
      value: eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJKb2UifQ.ipevRNuRP6HflG8cFKnmUPtypruRC4fb1DWtoLL62SY
```

### PlainText Connection with OAuth2 Authentication with Secret
  
```bash
kubectl create secret generic oauth2-key-file --from-file=sndev-admin.json
```

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarConnection
metadata:
  name: pulsar-connection-oauth2-secret
spec:
  adminServiceURL: http://pulsar-sn-platform-broker.test.svc.cluster.local:8080
  brokerServiceURL: pulsar://pulsar-sn-platform-broker.test.svc.cluster.local:6650
  clusterName: pulsar-cluster
  authentication:
    oauth2:
      issuerEndpoint: https://auth.streamnative.cloud
      clientID: pvqx76oGvWQMIGGP2ozMfOus2s4tDQAJ
      audience: urn:sn:pulsar:sndev:us-west
      key: 
        secretRef:
          name: oauth2-key-file
          key: sndev-admin.json
```

### PlainText Connection with OAuth2 Authentication with value

  ```yaml
  apiVersion: resource.streamnative.io/v1alpha1
  kind: PulsarConnection
  metadata:
    name: test-tls-pulsar-connection
    namespace: test
  spec:
    adminServiceURL: http://test-pulsar-sn-platform-broker.test.svc.cluster.local:8080
    brokerServiceURL: pulsar://test-pulsar-sn-platform-broker.test.svc.cluster.local:6650
    clusterName: pulsar-cluster
    authentication:
      oauth2:
        issuerEndpoint: https://auth.streamnative.cloud
        clientID: pvqx76oGvWQMIGGP2ozMfOus2s4tDQAJ
        audience: urn:sn:pulsar:sndev:us-west
        key: 
          # Use the keyFile contents as the oauth2 key value
          value: {"type":"sn_service_account","client_id":"zvex72oGvFQMBQGZ2ozMxOus2s4tQASJ","client_secret":"60J6fo81j-h69_vVvYvqFOHs2NfOyy6pqGqwIhTgnxpQ7O3UH8PdCbVtdm_SJjIf","client_email":"contoso@sndev.auth.streamnative.cloud","issuer_url":"https://auth.streamnative.cloud"}

* TLS authentication

  ```yaml
  apiVersion: resource.streamnative.io/v1alpha1
  kind: PulsarConnection
  metadata:
    name: test-tls-auth-pulsar-connection
    namespace: test
  spec:
    adminServiceURL: http://test-pulsar-sn-platform-broker.test.svc.cluster.local:8080
  brokerServiceURL: pulsar://test-pulsar-sn-platform-broker.test.svc.cluster.local:6650
  clusterName: pulsar-cluster
  authentication:
    tls:
      clientCertificateKeyPath: /certs/tls.key
      clientCertificatePath: /certs/tls.crt
  ```

This table lists specifications available for the `PulsarConnection` resource.

| Option | Description | Required or not |
| ---| --- |--- |
| `adminServiceURL` | The admin service URL of the Pulsar cluster, such as `http://cluster-broker.test.svc.cluster.local:8080`. | No |
| `authentication` | A secret that stores authentication configurations. This option is required when you enable authentication for your Pulsar cluster. Support JWT Token and OAuth2 authentication methods. | No |
| `brokerServiceURL` | The broker service URL of the Pulsar cluster, such as `pulsar://cluster-broker.test.svc.cluster.local:6650`. This option is required for configuring Geo-replication. This option is available for version `0.3.0` or above. | No |
| `brokerServiceSecureURL` | The broker service URL for secure connection to the Pulsar cluster, such as `pulsar+ssl://cluster-broker.test.svc.cluster.local:6651`. This option is required for configuring Geo-replication when TLS is enabled. This option is available for version `0.3.0` or above. | No |
| `adminServiceSecureURL` | The admin service URL for secure connection to the Pulsar cluster, such as `https://cluster-broker.test.svc.cluster.local:443`. This option is available for version `0.3.0` or above. | No |
| `clusterName` | The Pulsar cluster name. You can use the `pulsar-admin clusters list` command to get the Pulsar cluster name. This option is required for configuring Geo-replication. Provided from `0.3.0` | No |
| `tlsAllowInsecureConnection` | A flag that indicates whether to allow insecure connection to the broker. Provided from `0.5.0` | No |
| `tlsEnableHostnameVerification` | A flag that indicates wether hostname verification is enabled. Provided from `0.5.0` | No |
| `tlsTrustCertsFilePath` | The path to the certificate used during hostname verfification. Provided from `0.5.0` | No |

1. Apply the YAML file to create the Pulsar Connection. 

```shell
kubectl  apply -f connection.yaml
```

3. Check the resource status.

```shell
kubectl -n test get pulsarconnection.resource.streamnative.io
```
```shell
NAME                     ADMIN_SERVICE_URL                                        GENERATION   OBSERVED_GENERATION   READY
test-pulsar-connection   http://ok-sn-platform-broker.test.svc.cluster.local:8080 1            1                     True
```

## Update PulsarConnection

You can update the connection by editing the connection.yaml, then apply it again. For example, if pulsar cluster doesn’t setup the authentication, then you don’t need the authentication part in the spec

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarConnection
metadata:
  name: test-pulsar-connection
  namespace: test
spec:
  adminServiceURL: http://test-pulsar-sn-platform-broker.test.svc.cluster.local:8080
```

```shell
kubectl apply -f connection.yaml
```

## Delete PulsarConnection

```shell
kubectl -n test delete pulsarconnection.resource.streamnative.io test-pulsar-connection
```

Please be noticed, because the Pulsar Resources Operator are using the connection to manage pulsar resources, If you delete the pulsar connection, it will only be deleted after the resources CRs are deleted
