# PulsarGeoReplication

## Create PulsarGeoReplication

The PulsarGeoReplication is a unidirectional setup. When you create a PulsarGeoReplication for `cluster1` only, data will be replicated from `cluster1` to `cluster2`. If you need to replicate data between clusters `cluster1` and `cluster2`, you need to create PulsarGeoReplication for both `cluster1` and `cluster2`.


The Pulsar Resource Operator will create a new cluster named `clusterName` in the destination connection for each PulsarGeoReplication.


1. Define a Geo-replication named `cluster1-geo` and save the YAML file as `cluster1-geo.yaml`. 

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarGeoReplication
metadata:
  name: cluster1-geo
  namespace: cluster1
spec:
  connectionRef:
    name: cluster1-connection
  destinationConnectionRef:
    name: cluster1-dest-connection
  lifecyclePolicy: CleanUpAfterDeletion
```

This table lists specifications available for the `PulsarGeoReplication ` resource.

| Option | Description | Required or not |
| ---| --- |--- |
| `connectionRef` | The reference to a PulsarConnection. | Yes |
| `destinationConnectionRef` | The reference to a destination PulsarConnection. | Yes |
| `lifecyclePolicy` | The resource lifecycle policy. Available options are `CleanUpAfterDeletion` and `KeepAfterDeletion`. By default, it is set to `KeepAfterDeletion`. | Optional |


## How to configure Geo-replication

This section describes how to configure Geo-replication between clusters `cluster1-sn-platform` and `cluster2-sn-platform` in different namespaces of the same Kubernetes cluster.

The relation is shown below.
```mermaid
graph TD;
   cluster1-->cluster2;
   cluster2-->cluster1;
```

### Prerequisites
1. Deploy two pulsar clusters.
2. Ensure that both clusters can access each other.

### Update the existing PulsarConnection
Add the Pulsar cluster `cluster1-sn-platform` information through the `clusterName` and `brokerServiceURL` fields to the existing PulsarConnection.

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarConnection
metadata:
  name: cluster1-connection
  namespace: cluster1
spec:
  clusterName: cluster1-sn-platform # the local cluster name
  adminServiceURL: http://cluster1-sn-platform-broker.cluster1.svc.cluster.local:8080
  brokerServiceURL: pulsar://cluster1-sn-platform-broker.cluster1.svc.cluster.local:6650
  authentication:
    token:
      secretRef:
        name: cluster1-sn-platform-vault-secret-env-injection
        key: brokerClientAuthenticationParameters
```

### Create a destination PulsarConnection
The destination PulsarConnection has the information of the Pulsar cluster`cluster2-sn-platform`.
Add the Pulsar cluster `cluster2-sn-platform` information through the `clusterName` and `brokerServiceURL` fields to the destination PulsarConnection.

> **Note**
>
> The token only supports the `value` format.

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarConnection
metadata:
  name: cluster1-dest-connection
  namespace: cluster1
spec:
  clusterName: cluster2-sn-platform # the remote cluster name
  adminServiceURL: http://cluster2-sn-platform-broker.cluster2.svc.cluster.local:8080 # the remote pulsar admin service
  brokerServiceURL: pulsar://cluster2-sn-platform-broker.cluster2.svc.cluster.local:6650 # the remote pulsar broker service
  authentication:
    token:
      value: YzBkMzVjOGEtNWRiNi0yZGM3LTI4NjktZTM1NGU1YmJiNTY4OjAwZTk2MzgzLWFjY2YtMGMzOC1kNmRlLTY1ODc0ODkxNmE1ZQo=

```


### Create a PulsarGeoReplication
This section enabled Geo-replication on `cluster1`, which replicates data from `cluster1` to `cluster2`. The operator will create a new cluster called `cluster2-sn-platform`.

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarGeoReplication
metadata:
  name: cluster1-geo
  namespace: cluster1
spec:
  connectionRef:
    name: cluster1-connection
  destinationConnectionRef:
    name: cluster1-dest-connection
  lifecyclePolicy: CleanUpAfterDeletion
```

### Grant the permission for the tenant

You can create a new tenant or update an existing tenant by adding the field `geoReplicationRefs`. It will add the cluster `cluster2-sn-platform` to the tenant.

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarTenant
metadata:
  name: cluster1-geo-tenant
  namespace: cluster1
spec:
  name: geo-test
  connectionRef:
    name: cluster1-connection
  geoReplicationRefs:
  - name: cluster1-geo
  lifecyclePolicy: CleanUpAfterDeletion
```

### Enable Geo-replication at the namespace level

You can create a new namespace or update an existing namespace by adding the field `geoReplicationRefs`. It will add the namespace to `cluster2-sn-platform`.

> **Note**
>
> Once you enable Geo-replication at the namespace level, messages to all topics within that namespace are replicated across clusters.

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarNamespace
metadata:
  name: cluster1-geo-namespace
  namespace: cluster1
spec:
  name: geo-test/testn1
  geoReplicationRefs:
  - name: cluster1-geo
  connectionRef:
    name: cluster1-connection
  backlogQuotaLimitSize: 1Gi
  backlogQuotaLimitTime: 24h
  bundles: 16
  messageTTL: 1h
  lifecyclePolicy: CleanUpAfterDeletion

```


### Enable Geo-replication at the topic level

You can create a new topic or update an existing topic by adding the field `geoReplicationRefs`. It will add the topic to `cluster2-sn-platform`.

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarTopic
metadata:
  name: cluster1-geo-topic1
  namespace: cluster1
spec:
  name: persistent://geo-test/testn1/topic1
  partitions: 1
  connectionRef:
    name: cluster1-connection
  geoReplicationRefs:
  - name: cluster1-geo
  lifecyclePolicy: CleanUpAfterDeletion
```

#### Test

After the resources are ready, you can test Geo-replication by producing and consuming messages.
- Open a terminal and run the command `./bin/pulsar-client produce geo-test/testn1/topic1 -m "hello" -n 10` to produce messages to `cluster1`.
- Open another terminal and run the command `./bin/pulsar-client consume geo-test/testn1/topic1 -s sub -n 0` to consume messages from `cluster2`.