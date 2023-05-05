# PulsarConnection

## Create PulsarConnection

1. Define a connection named `test-pulsar-connection` by using the YAML file and save the YAML file `connection.yaml`. 

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarConnection
metadata:
  name: test-pulsar-connection
  namespace: test
spec:
  adminServiceURL: http://test-pulsar-sn-platform-broker.test.svc.cluster.local:8080
  brokerServiceURL: pulsar://test-pulsar-sn-platform-broker.test.svc.cluster.local:6650
  clusterName: pulsar-cluster
```

Other `PulsarConnection` configuration examples:

* TLS connection
  
  ```yaml
  apiVersion: resource.streamnative.io/v1alpha1
  kind: PulsarConnection
  metadata:
    name: test-tls-pulsar-connection
    namespace: test
  spec:
    adminServiceSecureURL: https://test-pulsar-sn-platform-broker.test.svc.cluster.local:443
    brokerServiceSecureURL: pulsar+ssl//test-pulsar-sn-platform-broker.test.svc.cluster.local:6651
    clusterName: pulsar-cluster
  ```

* JWT Token Auth with Secret
  
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
    token:
      secretRef:
        name: test-pulsar-sn-platform-vault-secret-env-injection
        key: brokerClientAuthenticationParameters
  ```

* JWT Token Auth with value

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
    token:
      # JWT Token value should be base64 encoded to use
      value: ZXlKaGJHY2lPaUpJVXpJMU5pSjkuZXlKemRXSWlPaUowWlhOMExYVnpaWElpZlEuOU9IZ0U5WlVEZUJUWnM3blNNRUZJdUdORVgxOEZMUjNxdnk4bXF4U3hYdw==
  ```

* OAuth2 Auth

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
        # Encode the OAuth2 keyFile with base64 and create a secret to use
        secretRef:
          name: key-file-secret
          key: key-file
  ```

This table lists specifications available for the `PulsarConnection` resource.

| Option | Description | Required or not |
| ---| --- |--- |
| `adminServiceURL` | The admin service URL of the Pulsar cluster, such as `http://cluster-broker.test.svc.cluster.local:8080`. | No |
| `authentication` | A secret that stores authentication configurations. This option is required when you enable authentication for your Pulsar cluster. Supported authentication with JWT Token and OAuth2 | No |
| `brokerServiceURL` | The broker service URL of the Pulsar cluster, such as `pulsar://cluster-broker.test.svc.cluster.local:6650`. This option is required for configuring Geo-replication. This option is available for version `0.3.0` or above. | No |
| `brokerServiceSecureURL` | The broker service URL for secure connection of the Pulsar cluster, eg: `pulsar+ssl://cluster-broker.test.svc.cluster.local:6651`. This option is required for configuring Geo-replication with TLS enabled. Provided from `0.3.1` | No |
| `adminServiceSecureURL` | The admin service URL for secure connection of the Pulsar cluster, eg: `https://cluster-broker.test.svc.cluster.local:443`. Provided from `0.3.1` | No |
| `clusterName` | The Pulsar cluster name. You can use the `pulsar-admin clusters list` command to get the Pulsar cluster name. This option is required for configuring Geo-replication. Provided from `0.3.0` | No |
   

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
