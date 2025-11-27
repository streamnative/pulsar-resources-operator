# PulsarConnection

## Overview

The `PulsarConnection` resource defines the connection details for a Pulsar cluster. It covers service URLs, authentication, TLS verification, and cluster metadata used by the operator.

## Specifications

| Field | Description | Required | Version |
|-------|-------------|----------|---------|
| `adminServiceURL` | Admin service URL (e.g., `http://cluster-broker.test.svc.cluster.local:8080`). | No | All |
| `adminServiceSecureURL` | HTTPS admin service URL. | No | ≥ 0.3.0 |
| `brokerServiceURL` | Broker service URL (e.g., `pulsar://pulsar-sn-platform-broker.test.svc.cluster.local:6650`). | No | ≥ 0.3.0 |
| `brokerServiceSecureURL` | TLS broker service URL (e.g., `pulsar+ssl://pulsar-sn-platform-broker.test.svc.cluster.local:6651`). | No | ≥ 0.3.0 |
| `clusterName` | Pulsar cluster name (used for Geo-Replication). | No | ≥ 0.3.0 |
| `authentication` | Authentication configuration (`token`, `oauth2`, or `tls`). | No | All |
| `brokerClientTrustCertsFilePath` | Path to trusted TLS cert for broker connections. | No | ≥ 0.3.0 |
| `tlsAllowInsecureConnection` | Allow insecure TLS connection to brokers. | No | ≥ 0.5.0 |
| `tlsEnableHostnameVerification` | Enable hostname verification for TLS. | No | ≥ 0.5.0 |
| `tlsTrustCertsFilePath` | CA certificate path for TLS verification. | No | ≥ 0.5.0 |

Fields with a listed version are available only from that version onward.

## Authentication Methods

The `authentication` field supports JWT token, OAuth2, or TLS client certificate authentication. Configure only one method per connection.

### Subfields

| Field | Description | Type | Required |
|-------|-------------|------|----------|
| `token` | JWT Token authentication configuration. | `ValueOrSecretRef` | No |
| `oauth2` | OAuth2 authentication configuration. | `PulsarAuthenticationOAuth2` | No |
| `tls` | TLS client certificate authentication. | `PulsarAuthenticationTLS` | No |

#### ValueOrSecretRef

| Field | Description | Type | Required |
|-------|-------------|------|----------|
| `value` | Direct string value. | `*string` | No |
| `secretRef` | Reference to a Kubernetes Secret. | `*SecretKeyRef` | No |
| `file` | Local file path whose contents are used as the value. | `*string` | No |

#### SecretKeyRef

| Field | Description | Type | Required |
|-------|-------------|------|----------|
| `name` | Name of the Kubernetes Secret. | `string` | Yes |
| `key` | Key in the Kubernetes Secret. | `string` | Yes |

#### PulsarAuthenticationOAuth2

| Field | Description | Type | Required |
|-------|-------------|------|----------|
| `issuerEndpoint` | URL of the OAuth2 authorization server. | `string` | Yes |
| `clientID` | OAuth2 client identifier. | `string` | Yes |
| `audience` | Intended recipient of the token. | `string` | Yes |
| `key` | Client secret or path to JSON credentials file. | `ValueOrSecretRef` | Yes |
| `scope` | Requested permissions from the OAuth2 server. | `string` | No |

#### PulsarAuthenticationTLS

| Field | Description | Type | Required |
|-------|-------------|------|----------|
| `clientCertificatePath` | Path to the client certificate. | `string` | Yes |
| `clientCertificateKeyPath` | Path to the client private key. | `string` | Yes |

## Connection Examples

### Plaintext connection

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

### TLS connection

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarConnection
metadata:
  name: pulsar-connection-tls
  namespace: test
spec:
  adminServiceSecureURL: https://pulsar-sn-platform-broker.test.svc.cluster.local:443
  brokerServiceSecureURL: pulsar+ssl://pulsar-sn-platform-broker.test.svc.cluster.local:6651
  clusterName: pulsar-cluster
  brokerClientTrustCertsFilePath: /etc/pulsar/ca.crt
  tlsEnableHostnameVerification: true
```

## Authentication Examples

### JWT token with Secret

```bash
kubectl create secret generic pulsar-jwt-secret \
  --from-literal=brokerClientAuthenticationParameters=<base64-encoded-JWT>
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
      secretRef:
        name: pulsar-jwt-secret
        key: brokerClientAuthenticationParameters
```

### JWT token with direct value

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
      value: <base64-encoded-JWT-token>
```

### OAuth2 with Secret

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
      scope: <scope>
```

### OAuth2 with direct value

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
        value: |
          {
            "type":"sn_service_account",
            "client_id":"zvex72oGvFQMBQGZ2ozMxOus2s4tQASJ",
            "client_secret":"<client-secret>",
            "client_email":"contoso@sndev.auth.streamnative.cloud",
            "issuer_url":"https://auth.streamnative.cloud"
          }
      scope: <scope>
```

### OAuth2 with file-based `ValueOrSecretRef`

When you want the controller to read OAuth2 credentials from a mounted file instead of embedding them in the CR or a Secret reference, mount the secret into the operator pod and point `key.file` at the mounted path.

1) Create a secret from the credentials file:

```bash
kubectl create secret generic admin-credentials \
  --from-file=credentials.json=credentials.json \
  -n default
```

2) Mount the secret into the operator (example Helm values):

```yaml
extraVolumes:
  - name: admin-credentials
    secret:
      secretName: admin-credentials
      items:
        - key: credentials.json
          path: credentials.json

extraVolumeMounts:
  - name: admin-credentials
    mountPath: /etc/streamnative/admin-credentials
    readOnly: true
```

3) Reference the mounted file in the `PulsarConnection`:

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarConnection
metadata:
  name: pc-file
  namespace: default
spec:
  adminServiceSecureURL: https://<admin-host>
  brokerServiceSecureURL: pulsar+ssl://<broker-host>:6651
  clusterName: <cluster-name>
  authentication:
    oauth2:
      issuerEndpoint: https://auth.example.com/
      clientID: <client-id>
      audience: <audience>
      key:
        file: /etc/streamnative/admin-credentials/credentials.json
```

### TLS client-certificate authentication

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarConnection
metadata:
  name: test-tls-auth-pulsar-connection
  namespace: test
spec:
  adminServiceSecureURL: https://test-pulsar-sn-platform-broker.test.svc.cluster.local:443
  brokerServiceSecureURL: pulsar+ssl://test-pulsar-sn-platform-broker.test.svc.cluster.local:6651
  clusterName: pulsar-cluster
  authentication:
    tls:
      clientCertificateKeyPath: /certs/tls.key
      clientCertificatePath: /certs/tls.crt
  tlsTrustCertsFilePath: /certs/ca.crt
```

## Common Operations

- Create: `kubectl apply -f connection.yaml`
- Check status: `kubectl -n <namespace> get pulsarconnection.resource.streamnative.io`
- Update: edit `connection.yaml` and re-apply (for example, remove `authentication` if the cluster is unauthenticated).
- Delete: `kubectl -n <namespace> delete pulsarconnection.resource.streamnative.io <name>`; the CR is removed after dependent Pulsar resources are cleaned up.
