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
  # optional
  authentication:
    token:
      secretRef:
        name: test-pulsar-sn-platform-vault-secret-env-injection
        key: brokerClientAuthenticationParameters
```

This table lists specifications available for the `PulsarConnection` resource.

| Option | Description | Required or not |
| ---| --- |--- |
| `adminServiceURL` | The admin service URL of the Pulsar cluster, eg: `cluster-broker.test.svc.cluster.local:8080`. | Yes |
| `authentication` | A secret that stores authentication configurations.This option is required when you enable authentication for your Pulsar cluster. | No |
| `brokerServiceURL` | The broker service URL of the Pulsar cluster, eg: `cluster-broker.test.svc.cluster.local:8080`. This option is required when you want to setup geo replication| No |
| `clusterName` | The name of the local pulsar cluster, you can use `pulsar-admin clusters list` to get it. This option is required when you want to setup geo replication| No |
   

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
