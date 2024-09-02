# PulsarNamespace

## Overview

The `PulsarNamespace` resource defines a namespace in a Pulsar cluster. It allows you to configure various namespace-level settings such as backlog quotas, message TTL, and replication.

## Specifications

| Field | Description | Required |
|-------|-------------|----------|
| `name` | The fully qualified namespace name in the format "tenant/namespace". | Yes |
| `connectionRef` | Reference to the PulsarConnection resource used to connect to the Pulsar cluster for this namespace. | Yes |
| `bundles` | Number of bundles to split the namespace into. This affects how the namespace is distributed across the cluster. | No |
| `lifecyclePolicy` | Determines whether to keep or delete the Pulsar namespace when the Kubernetes resource is deleted. Options: `CleanUpAfterDeletion`, `KeepAfterDeletion`. | No |
| `maxProducersPerTopic` | Maximum number of producers allowed on a single topic in the namespace. | No |
| `maxConsumersPerTopic` | Maximum number of consumers allowed on a single topic in the namespace. | No |
| `maxConsumersPerSubscription` | Maximum number of consumers allowed on a single subscription in the namespace. | No |
| `messageTTL` | Time to Live (TTL) for messages in the namespace. Messages older than this TTL will be automatically marked as consumed. | No |
| `retentionTime` | Minimum time to retain messages in the namespace. Should be set in conjunction with RetentionSize for effective retention policy. | No |
| `retentionSize` | Maximum size of backlog retained in the namespace. Should be set in conjunction with RetentionTime for effective retention policy. | No |
| `backlogQuotaLimitTime` | Time limit for message backlog. Messages older than this limit will be removed or handled according to the retention policy. | No |
| `backlogQuotaLimitSize` | Size limit for message backlog. When the limit is reached, older messages will be removed or handled according to the retention policy. | No |
| `backlogQuotaRetentionPolicy` | Retention policy for messages when backlog quota is exceeded. Options: "producer_request_hold", "producer_exception", or "consumer_backlog_eviction". | No |
| `backlogQuotaType` | Controls how the backlog quota is enforced. Options: "destination_storage" (limits backlog by size in bytes), "message_age" (limits by time). | No |
| `offloadThresholdTime` | Time limit for message offloading. Messages older than this limit will be offloaded to the tiered storage. | No |
| `offloadThresholdSize` | Size limit for message offloading. When the limit is reached, older messages will be offloaded to the tiered storage. | No |
| `geoReplicationRefs` | List of references to PulsarGeoReplication resources, used to configure geo-replication for this namespace. Use only when using PulsarGeoReplication for setting up geo-replication between two Pulsar instances. | No |
| `replicationClusters` | List of clusters to which the namespace is replicated. Use only if replicating clusters within the same Pulsar instance. | No |

Note: Valid time units are "s" (seconds), "m" (minutes), "h" (hours), "d" (days), "w" (weeks).

## replicationClusters vs geoReplicationRefs

The `replicationClusters` and `geoReplicationRefs` fields serve different purposes in configuring replication for a Pulsar namespace:

1. `replicationClusters`:
   - Use this when replicating data between clusters within the same Pulsar instance.
   - It's a simple list of cluster names to which the namespace should be replicated.
   - This is suitable for scenarios where all clusters are managed by the same Pulsar instance and have direct connectivity.
   - Example use case: Replicating data between regions within a single Pulsar instance.

2. `geoReplicationRefs`:
   - Use this when setting up geo-replication between separate Pulsar instances.
   - It references PulsarGeoReplication resources, which contain more detailed configuration for connecting to external Pulsar clusters.
   - This is appropriate for scenarios involving separate Pulsar deployments, possibly in different data centers or cloud providers.
   - Example use case: Replicating data between two independent Pulsar instancesin different geographical locations.

Choose `replicationClusters` for simpler, intra-instance replication, and `geoReplicationRefs` for more complex, inter-instance geo-replication scenarios. These fields are mutually exclusive; use only one depending on your replication requirements.

## Create A Pulsar Namespace

1. Define a namespace named `test-tenant/testns` by using the YAML file and save the YAML file `namespace.yaml`.
```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarNamespace
metadata:
  name: test-pulsar-namespace
  namespace: test
spec:
  name: test-tenant/testns
  connectionRef:
    name: test-pulsar-connection
  backlogQuotaLimitSize: 1Gi
  backlogQuotaLimitTime: 24h
  bundles: 16
  messageTTL: 1h
  # backlogQuotaRetentionPolicy: producer_request_hold
  # maxProducersPerTopic: 2
  # maxConsumersPerTopic: 2
  # optional
  # maxConsumersPerSubscription: 2
  # retentionTime: 20h
  # retentionSize: 2Gi
  # lifecyclePolicy: CleanUpAfterDeletion
```

2. Apply the YAML file to create the namespace.

```shell
kubectl apply -f namespace.yaml
```

3. Check the resource status. When column Ready is true, it indicates the resource is created successfully in the pulsar cluster

```shell
kubectl -n test get pulsarnamespace.resource.streamnative.io
```

```shell
NAME                    RESOURCE_NAME        GENERATION   OBSERVED_GENERATION   READY
test-pulsar-namespace   test-tenant/testns   1            1                     True
```

## Update A Pulsar Namespace

You can update the namespace policies by editing the `namespace.yaml` file and then applying it again using `kubectl apply -f namespace.yaml`. This allows you to modify various settings of the Pulsar namespace.

Please note the following important points:

1. The fields `name` and `bundles` cannot be updated after the namespace is created. These are immutable properties of the namespace.

2. Other fields such as `backlogQuotaLimitSize`, `backlogQuotaLimitTime`, `messageTTL`, `maxProducersPerTopic`, `maxConsumersPerTopic`, `maxConsumersPerSubscription`, `retentionTime`, and `retentionSize` can be modified.

3. If you want to change the `connectionRef`, ensure that the new PulsarConnection resource exists and is properly configured. Changing the `connectionRef` can have significant implications:

   - If the new PulsarConnection refers to the same Pulsar cluster (i.e., the admin and broker URLs are the same), the namespace will remain in its original location. The operator will simply use the new connection details to manage the existing namespace.

   - If the new PulsarConnection points to a different Pulsar cluster (i.e., different admin and broker URLs), the operator will attempt to create a new namespace with the same configuration in the new cluster. The original namespace in the old cluster will not be automatically deleted.

   Be cautious when changing the `connectionRef`, especially if it points to a new cluster, as this can lead to namespace duplication across clusters. Always verify the intended behavior and manage any cleanup of the old namespace if necessary.

4. Changes to `lifecyclePolicy` will only affect what happens when the PulsarNamespace resource is deleted, not the current state of the namespace.

5. After applying changes, you can check the status of the update using:
   ```shell
   kubectl -n test get pulsarnamespace.resource.streamnative.io test-pulsar-namespace
   ```
   The `OBSERVED_GENERATION` should increment, and `READY` should become `True` when the update is complete.

6. Be cautious when updating namespace policies, as changes may affect existing producers and consumers. It's recommended to test changes in a non-production environment first.

## Delete A Pulsar Namespace

To delete a PulsarNamespace resource, use the following kubectl command:

```shell
kubectl -n test delete pulsarnamespace.resource.streamnative.io test-pulsar-namespace
```

Please be noticed, when you delete the namespace, the real namespace will still exist if the `lifecyclePolicy` is `KeepAfterDeletion`
