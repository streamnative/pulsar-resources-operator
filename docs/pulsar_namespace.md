# PulsarNamespace

## Overview

The `PulsarNamespace` resource defines a namespace in a Pulsar cluster. It allows you to configure various namespace-level settings such as backlog quotas, message TTL, and replication.

## Specifications

| Field                         | Description                                                                                                                                                                                                       | Required |
|-------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------|
| `name`                        | The fully qualified namespace name in the format "tenant/namespace".                                                                                                                                              | Yes      |
| `connectionRef`               | Reference to the PulsarConnection resource used to connect to the Pulsar cluster for this namespace.                                                                                                              | Yes      |
| `bundles`                     | Number of bundles to split the namespace into. This affects how the namespace is distributed across the cluster.                                                                                                  | No       |
| `lifecyclePolicy`             | Determines whether to keep or delete the Pulsar namespace when the Kubernetes resource is deleted. Options: `CleanUpAfterDeletion`, `KeepAfterDeletion`.                                                          | No       |
| `maxProducersPerTopic`        | Maximum number of producers allowed on a single topic in the namespace.                                                                                                                                           | No       |
| `maxConsumersPerTopic`        | Maximum number of consumers allowed on a single topic in the namespace.                                                                                                                                           | No       |
| `maxConsumersPerSubscription` | Maximum number of consumers allowed on a single subscription in the namespace.                                                                                                                                    | No       |
| `messageTTL`                  | Time to Live (TTL) for messages in the namespace. Messages older than this TTL will be automatically marked as consumed.                                                                                          | No       |
| `retentionTime`               | Minimum time to retain messages in the namespace. Should be set in conjunction with RetentionSize for effective retention policy.                                                                                 | No       |
| `retentionSize`               | Maximum size of backlog retained in the namespace. Should be set in conjunction with RetentionTime for effective retention policy.                                                                                | No       |
| `backlogQuotaLimitTime`       | Time limit for message backlog. Messages older than this limit will be removed or handled according to the retention policy.                                                                                      | No       |
| `backlogQuotaLimitSize`       | Size limit for message backlog. When the limit is reached, older messages will be removed or handled according to the retention policy.                                                                           | No       |
| `backlogQuotaRetentionPolicy` | Retention policy for messages when backlog quota is exceeded. Options: "producer_request_hold", "producer_exception", or "consumer_backlog_eviction".                                                             | No       |
| `backlogQuotaType`            | Controls how the backlog quota is enforced. Options: "destination_storage" (limits backlog by size in bytes), "message_age" (limits by time).                                                                     | No       |
| `offloadThresholdTime`        | Time limit for message offloading. Messages older than this limit will be offloaded to the tiered storage.                                                                                                        | No       |
| `offloadThresholdSize`        | Size limit for message offloading. When the limit is reached, older messages will be offloaded to the tiered storage.                                                                                             | No       |
| `geoReplicationRefs`          | List of references to PulsarGeoReplication resources, used to configure geo-replication for this namespace. Use only when using PulsarGeoReplication for setting up geo-replication between two Pulsar instances. | No       |
| `replicationClusters`         | List of clusters to which the namespace is replicated. Use only if replicating clusters within the same Pulsar instance.                                                                                          | No       |
| `deduplication`               | Whether to enable message deduplication for the namespace.                                                                                                                                                        | No       |
| `bookieAffinityGroup`         | Set the bookie-affinity group for the namespace, which has two sub fields: `bookkeeperAffinityGroupPrimary(String)` is required, and `bookkeeperAffinityGroupSecondary(String)` is optional.                      | No       |
| `topicAutoCreationConfig`     | Configures automatic topic creation behavior within this namespace. Contains settings for whether auto-creation is allowed, the type of topics created, and default number of partitions.                          | No       |
| `schemaCompatibilityStrategy` | Schema compatibility strategy for this namespace. Controls how schema evolution is handled for topics within this namespace. Options: `UNDEFINED`, `ALWAYS_INCOMPATIBLE`, `ALWAYS_COMPATIBLE`, `BACKWARD`, `FORWARD`, `FULL`, `BACKWARD_TRANSITIVE`, `FORWARD_TRANSITIVE`, `FULL_TRANSITIVE`.                                                          | No       |
| `schemaValidationEnforced`    | Controls whether schema validation is enforced for this namespace. When enabled, producers must provide a schema when publishing messages. If not specified, the cluster's default schema validation enforcement setting will be used.                                                                                                  | No       |

Note: Valid time units are "s" (seconds), "m" (minutes), "h" (hours), "d" (days), "w" (weeks).

## topicAutoCreationConfig

The `topicAutoCreationConfig` field allows you to control the automatic topic creation behavior at the namespace level:

| Field        | Description                                                                           | Required |
|-------------|---------------------------------------------------------------------------------------|----------|
| `allow`     | Whether automatic topic creation is allowed in this namespace.                         | No       |
| `type`      | The type of topics to create automatically. Options: "partitioned", "non-partitioned". | No       |
| `partitions`| The default number of partitions for automatically created topics when type is "partitioned". | No   |

This configuration overrides the broker's default topic auto-creation settings for the specific namespace. When a client attempts to produce messages to or consume messages from a non-existent topic, the broker can automatically create that topic based on these settings.

### Configuration examples

1. **Enable auto-creation with partitioned topics**:
   ```yaml
   topicAutoCreationConfig:
     allow: true
     type: "partitioned"
     partitions: 8
   ```
   This will automatically create partitioned topics with 8 partitions when clients attempt to use non-existent topics.

2. **Enable auto-creation with non-partitioned topics**:
   ```yaml
   topicAutoCreationConfig:
     allow: true
     type: "non-partitioned"
   ```
   This will automatically create non-partitioned topics when clients attempt to use non-existent topics.

3. **Disable auto-creation**:
   ```yaml
   topicAutoCreationConfig:
     allow: false
   ```
   This explicitly disables topic auto-creation for the namespace, overriding any broker-level settings that might enable it.

## Schema Management

Pulsar provides powerful schema management capabilities at the namespace level, allowing you to control how schema evolution is handled and whether schema validation is enforced. This feature consists of two complementary settings: schema compatibility strategy and schema validation enforcement.

### Schema Compatibility Strategy

The `schemaCompatibilityStrategy` field controls how Pulsar handles schema evolution for topics within the namespace. This allows you to configure different compatibility requirements for different namespaces based on your use case.

#### Available Strategies

1. **UNDEFINED**: Uses the cluster's default schema compatibility strategy. When this value is set, the namespace inherits the schema compatibility settings from the Pulsar cluster configuration. This is useful when you want to maintain consistency with cluster-wide policies.

2. **ALWAYS_INCOMPATIBLE**: Disallows any schema changes. This is the most restrictive strategy, suitable for ultra-stable environments where no schema evolution is desired and strict schema immutability is required.

3. **ALWAYS_COMPATIBLE**: Allows any schema changes without validation. This is the most permissive strategy but may lead to compatibility issues. Suitable for development/testing environments where rapid iteration is needed.

4. **BACKWARD**: New schema can read data written with the previous schema. This strategy supports consumer-driven schema evolution, such as adding optional fields or removing fields. Consumers with new schemas can process data from producers with older schemas.

5. **FORWARD**: Previous schema can read data written with the new schema. This strategy supports producer-driven schema evolution, such as adding fields that older consumers can ignore. Consumers with older schemas can still process data from producers with newer schemas.

6. **FULL**: Schema changes are both forward and backward compatible. Both new and previous schemas can read data written by either schema. This provides strict compatibility requirements in both directions and is suitable for environments requiring maximum interoperability.

7. **BACKWARD_TRANSITIVE**: New schema can read data written with any previous schema in the chain. This provides long-term backward compatibility across multiple schema versions, ensuring consumers can always read historical data regardless of schema evolution.

8. **FORWARD_TRANSITIVE**: Any previous schema can read data written with the new schema. This ensures new data is readable by any older schema version, providing comprehensive forward compatibility across the entire schema evolution chain.

9. **FULL_TRANSITIVE**: Schema changes are forward and backward compatible with all schemas. Any schema in the chain can read data written by any other schema in the chain. This provides maximum compatibility guarantees across all schema versions.

#### Usage Examples

**Development Environment**:
```yaml
schemaCompatibilityStrategy: ALWAYS_COMPATIBLE
```

**Production Environment**:
```yaml
schemaCompatibilityStrategy: BACKWARD
```

**Critical Systems**:
```yaml
schemaCompatibilityStrategy: FULL_TRANSITIVE
```

### Schema Validation Enforcement

The `schemaValidationEnforced` field controls whether producers must provide a schema when publishing messages to topics within the namespace.

- **When enabled (`true`)**: Producers must provide a schema when publishing messages. Messages without schemas will be rejected. This ensures all data in the namespace has a defined structure and is recommended for production environments where data consistency is critical.

- **When disabled (`false`)**: Producers can publish messages with or without schemas. This allows for more flexibility in message publishing and is useful for development/testing environments or legacy integrations.

- **Default behavior**: If `schemaValidationEnforced` is not specified, the cluster's default schema validation enforcement setting will be used.

### Configuration Examples by Use Case

#### Development/Testing Environment
```yaml
schemaCompatibilityStrategy: ALWAYS_COMPATIBLE
schemaValidationEnforced: false
```
This configuration allows rapid schema iteration and flexible schema validation for experimentation.

#### Standard Production Environment
```yaml
schemaCompatibilityStrategy: BACKWARD
schemaValidationEnforced: true
```
This provides a good balance between flexibility and safety, ensuring consumers can handle schema changes while enforcing schema validation for data consistency.

#### Mission-Critical Systems
```yaml
schemaCompatibilityStrategy: FULL_TRANSITIVE
schemaValidationEnforced: true
```
This configuration provides maximum compatibility guarantees with strict schema validation enforcement.

#### Legacy System Integration
```yaml
schemaCompatibilityStrategy: FORWARD_TRANSITIVE
schemaValidationEnforced: false
```
This ensures older systems can consume new data while allowing gradual migration without strict schema requirements.

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
  # Schema management configuration
  # schemaCompatibilityStrategy: BACKWARD
  # schemaValidationEnforced: true
  # backlogQuotaRetentionPolicy: producer_request_hold
  # maxProducersPerTopic: 2
  # maxConsumersPerTopic: 2
  # optional
  # maxConsumersPerSubscription: 2
  # retentionTime: 20h
  # retentionSize: 2Gi
  # lifecyclePolicy: CleanUpAfterDeletion
  # topicAutoCreationConfig:
  #  allow: true
  #  type: "partitioned"
  #  partitions: 4
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

2. Other fields such as `backlogQuotaLimitSize`, `backlogQuotaLimitTime`, `messageTTL`, `maxProducersPerTopic`, `maxConsumersPerTopic`, `maxConsumersPerSubscription`, `retentionTime`, `retentionSize`, `topicAutoCreationConfig`, `schemaCompatibilityStrategy`, and `schemaValidationEnforced` can be modified.

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
