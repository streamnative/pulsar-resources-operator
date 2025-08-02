# PulsarTopic

## Overview

The `PulsarTopic` resource defines a topic in a Pulsar cluster. It allows you to configure various topic-level settings such as persistence, partitions, retention policies, and schema information. This resource is part of the Pulsar Resources Operator, which enables declarative management of Pulsar resources using Kubernetes custom resources.

## Sepcifications

## Specifications

| Field                               | Description                                                                                                                                                                             | Required |
|-------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------|
| `name`                              | The fully qualified topic name in the format "persistent://tenant/namespace/topic" or "non-persistent://tenant/namespace/topic".                                                        | Yes      |
| `connectionRef`                     | Reference to the PulsarConnection resource used to connect to the Pulsar cluster for this topic.                                                                                        | Yes      |
| `persistent`                        | Whether the topic is persistent or non-persistent. Default is false. Can also be set by topic name prefix.                                                                              | No       |
| `partitions`                        | Number of partitions for the topic. Default is 0.                                                                                                                                       | No       |
| `maxProducers`                      | Maximum number of producers allowed on the topic.                                                                                                                                       | No       |
| `maxConsumers`                      | Maximum number of consumers allowed on the topic.                                                                                                                                       | No       |
| `messageTTL`                        | Time to Live (TTL) for messages in the topic. Messages older than this TTL will be automatically marked as consumed.                                                                    | No       |
| `maxUnAckedMessagesPerConsumer`     | Maximum number of unacknowledged messages allowed per consumer.                                                                                                                         | No       |
| `maxUnAckedMessagesPerSubscription` | Maximum number of unacknowledged messages allowed per subscription.                                                                                                                     | No       |
| `retentionTime`                     | Minimum time to retain messages in the topic. Should be set in conjunction with retentionSize for effective retention policy.                                                           | No       |
| `retentionSize`                     | Maximum size of backlog retained in the topic. Should be set in conjunction with retentionTime for effective retention policy.                                                          | No       |
| `backlogQuotaLimitTime`             | Time limit for message backlog. Messages older than this limit will be removed or handled according to the retention policy.                                                            | No       |
| `backlogQuotaLimitSize`             | Size limit for message backlog. When the limit is reached, older messages will be removed or handled according to the retention policy.                                                 | No       |
| `backlogQuotaRetentionPolicy`       | Retention policy for messages when backlog quota is exceeded. Options: "producer_request_hold", "producer_exception", or "consumer_backlog_eviction".                                   | No       |
| `lifecyclePolicy`                   | Determines whether to keep or delete the Pulsar topic when the Kubernetes resource is deleted. Options: `CleanUpAfterDeletion`, `KeepAfterDeletion`. Default is `CleanUpAfterDeletion`. | No       |
| `schemaInfo`                        | Schema information for the topic. See [schemaInfo](#schemainfo) for more details.                                                                                                       | No       |
| `geoReplicationRefs`                | List of references to PulsarGeoReplication resources, used to enable geo-replication at the topic level.                                                                                | No       |
| `replicationClusters`               | List of clusters to which the topic is replicated. Use only if replicating clusters within the same Pulsar instance.                                                                    | No       |
| `deduplication`                     | whether to enable message deduplication for the topic.                                                                                                                                  | No       |
| `compactionThreshold`               | Size threshold in bytes for automatic topic compaction. When the topic reaches this size, compaction will be triggered automatically.                                                   | No       |
| `persistencePolicies`               | Persistence configuration for the topic, controlling how data is stored and replicated in BookKeeper. See [persistencePolicies](#persistencePolicies) for more details.               | No       |
| `delayedDelivery`                   | Delayed delivery policy for the topic, allowing messages to be delivered with a delay. See [delayedDelivery](#delayedDelivery) for more details.                                       | No       |
| `dispatchRate`                      | Message dispatch rate limiting policy for the topic, controlling the rate at which messages are delivered to consumers. See [dispatchRate](#dispatchRate) for more details.            | No       |
| `publishRate`                       | Message publish rate limiting policy for the topic, controlling the rate at which producers can publish messages. See [publishRate](#publishRate) for more details.                    | No       |
| `inactiveTopicPolicies`             | Inactive topic cleanup policy for the topic, controlling how inactive topics are automatically cleaned up. See [inactiveTopicPolicies](#inactiveTopicPolicies) for more details.       | No       |
| `subscribeRate`                     | Subscription rate limiting policy for the topic, controlling the rate at which new subscriptions can be created. See [subscribeRate](#subscribeRate) for more details.                 | No       |
| `maxMessageSize`                    | Maximum size of messages that can be published to the topic. Messages larger than this size will be rejected.                                                                           | No       |
| `maxConsumersPerSubscription`       | Maximum number of consumers allowed per subscription.                                                                                                                                    | No       |
| `maxSubscriptionsPerTopic`          | Maximum number of subscriptions allowed on the topic.                                                                                                                                   | No       |
| `schemaValidationEnforced`          | Whether schema validation is enforced for the topic. When enabled, only messages that conform to the topic's schema will be accepted.                                                  | No       |
| `subscriptionDispatchRate`          | Message dispatch rate limiting policy for subscriptions, controlling the rate at which messages are delivered to consumers per subscription. Uses same format as [dispatchRate](#dispatchRate). | No       |
| `replicatorDispatchRate`            | Message dispatch rate limiting policy for replicators, controlling the rate at which messages are replicated to other clusters. Uses same format as [dispatchRate](#dispatchRate).    | No       |
| `deduplicationSnapshotInterval`     | Interval for taking deduplication snapshots. This affects the deduplication performance and storage overhead.                                                                           | No       |
| `offloadPolicies`                   | Offload policies for the topic, controlling how data is offloaded to external storage systems.                                                                                          | No       |
| `autoSubscriptionCreation`          | Auto subscription creation override for the topic, controlling whether subscriptions can be created automatically.                                                                       | No       |
| `schemaCompatibilityStrategy`       | Schema compatibility strategy for the topic, controlling how schema evolution is handled. Options: UNDEFINED, ALWAYS_INCOMPATIBLE, ALWAYS_COMPATIBLE, BACKWARD, FORWARD, FULL, BACKWARD_TRANSITIVE, FORWARD_TRANSITIVE, FULL_TRANSITIVE. | No       |
| `properties`                        | Map of user-defined properties associated with the topic. These can be used to store additional metadata about the topic.                                                              | No       |

Note: Valid time units for duration fields are "s" (seconds), "m" (minutes), "h" (hours), "d" (days), "w" (weeks).

## replicationClusters vs geoReplicationRefs

The `replicationClusters` and `geoReplicationRefs` fields serve different purposes in configuring replication for a Pulsar topic:

1. `replicationClusters`:
   - Use this when replicating data between clusters within the same Pulsar instance.
   - It's a simple list of cluster names to which the topic should be replicated.
   - This is suitable for scenarios where all clusters are managed by the same Pulsar instance and have direct connectivity.
   - Example use case: Replicating data between regions within a single Pulsar instance.

2. `geoReplicationRefs`:
   - Use this when setting up geo-replication between separate Pulsar instances.
   - It references PulsarGeoReplication resources, which contain more detailed configuration for connecting to external Pulsar clusters.
   - This is appropriate for scenarios involving separate Pulsar deployments, possibly in different data centers or cloud providers.
   - Example use case: Replicating data between two independent Pulsar instancesin different geographical locations.

Choose `replicationClusters` for simpler, intra-instance replication, and `geoReplicationRefs` for more complex, inter-instance geo-replication scenarios. These fields are mutually exclusive; use only one depending on your replication requirements.


## Create A Pulsar Topic

1. Define a topic named `persistent://test-tenant/testns/topic123` by using the YAML file and save the YAML file `topic.yaml`.
```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarTopic
metadata:
  name: "test-pulsar-topic123"
  namespace: test
spec:
  name: persistent://test-tenant/testns/topic123
  connectionRef:
    name: "test-pulsar-connection"
# persistent: true
# partitions: 8
# maxProducers: 8
# maxConsumers: 8
# messageTTL:
# maxUnAckedMessagesPerConsumer:
# maxUnAckedMessagesPerSubscription:
# retentionTime: 20h
# retentionSize: 2Gi
# backlogQuotaLimitTime: 24h
# backlogQuotaLimitSize: 1Gi
# backlogQuotaRetentionPolicy: producer_request_hold
# lifecyclePolicy: CleanUpAfterDeletion
# compactionThreshold: 104857600  # 100MB
# persistencePolicies:
#   bookkeeperEnsemble: 3
#   bookkeeperWriteQuorum: 2
#   bookkeeperAckQuorum: 2
#   managedLedgerMaxMarkDeleteRate: "1.0"
# delayedDelivery:
#   active: true
#   tickTimeMillis: 1000
# dispatchRate:
#   dispatchThrottlingRateInMsg: 1000
#   dispatchThrottlingRateInByte: 1048576
#   ratePeriodInSecond: 1
# publishRate:
#   publishThrottlingRateInMsg: 2000
#   publishThrottlingRateInByte: 2097152
# inactiveTopicPolicies:
#   inactiveTopicDeleteMode: "delete_when_no_subscriptions"
#   maxInactiveDurationInSeconds: 3600
#   deleteWhileInactive: true
```

2. Apply the YAML file to create the topic.

```shell
kubectl apply -f topic.yaml
```

3. Check the resource status. When column Ready is true, it indicates the resource is created successfully in the pulsar cluster

```shell
kubectl -n test get pulsartopic.resource.streamnative.io
```

```shell
NAME                   RESOURCE_NAME                              GENERATION   OBSERVED_GENERATION   READY
test-pulsar-topic123   persistent://test-tenant/testns/topic123   1            1                     True
```

## Update A Pulsar Topic

You can update the topic policies by editing the `topic.yaml` file and then applying it again using `kubectl apply -f topic.yaml`. This allows you to modify various settings of the Pulsar topic.

Important notes when updating a Pulsar topic:

1. The fields `name` and `persistent` are immutable and cannot be updated after the topic is created.

2. Other fields such as `partitions`, `maxProducers`, `maxConsumers`, `messageTTL`, `retentionTime`, `retentionSize`, `backlogQuotaLimitTime`, `backlogQuotaLimitSize`, `backlogQuotaRetentionPolicy`, `compactionThreshold`, `persistencePolicies`, `delayedDelivery`, `dispatchRate`, `publishRate`, and `inactiveTopicPolicies` can be modified.

3. If you want to change the `connectionRef`, ensure that the new PulsarConnection resource exists and is properly configured. Changing the `connectionRef` can have significant implications:

   - If the new PulsarConnection refers to the same Pulsar cluster (i.e., the admin and broker URLs are the same), the topic will remain in its original location. The operator will simply use the new connection details to manage the existing topic.

   - If the new PulsarConnection points to a different Pulsar cluster (i.e., different admin and broker URLs), the operator will attempt to create a new topic with the same configuration in the new cluster. The original topic in the old cluster will not be automatically deleted.

   Be cautious when changing the `connectionRef`, especially if it points to a new cluster, as this can lead to topic duplication across clusters. Always verify the intended behavior and manage any cleanup of the old topic if necessary.

4. Changes to `lifecyclePolicy` will only affect what happens when the PulsarTopic resource is deleted, not the current state of the topic.

5. Be cautious when updating topic policies, as changes may affect existing producers and consumers. It's recommended to test changes in a non-production environment first.

6. After applying changes, you can check the status of the update using:
   ```shell
   kubectl -n test get pulsartopic.resource.streamnative.io test-pulsar-topic123
   ```
   The `OBSERVED_GENERATION` should increment, and `READY` should become `True` when the update is complete.

7. Updating the `schemaInfo` field may have implications for existing producers and consumers. Ensure that any schema changes adhere to Pulsar's schema compatibility strategies. For more information on schema evolution and compatibility, refer to the [Pulsar Schema Evolution and Compatibility](https://pulsar.apache.org/docs/schema-understand#schema-evolution) documentation.

## Delete A PulsarTopic

To delete a PulsarTopic resource, use the following kubectl command:

```shell
kubectl delete pulsartopic.resource.streamnative.io test-pulsar-topic123
```

Please be aware that when you delete the topic, the actual topic will still exist in the Pulsar cluster if the `lifecyclePolicy` is set to `KeepAfterDeletion`. For more detailed information about the lifecycle policies and their implications, please refer to the [PulsarResourceLifeCyclePolicy documentation](pulsar_resource_lifecycle.md).

If you want to delete the topic in the pulsar cluster, you can use the following command:

```shell
pulsarctl topics delete persistent://test-tenant/testns/topic123
```

## Topic-Level Policies

The PulsarTopic resource supports several advanced topic-level policies that provide fine-grained control over topic behavior.

### persistencePolicies

The `persistencePolicies` field configures how data is stored and replicated in BookKeeper, the storage layer for Pulsar.

| Field | Description | Type | Required |
|-------|-------------|------|----------|
| `bookkeeperEnsemble` | Number of bookies to store ledger data across | int32 | No |
| `bookkeeperWriteQuorum` | Number of replicas to write for each ledger entry | int32 | No |
| `bookkeeperAckQuorum` | Number of replicas that must acknowledge writes | int32 | No |
| `managedLedgerMaxMarkDeleteRate` | Rate limit for mark-delete operations as a string (e.g., "1.0", "2.5") | string | No |

**Example:**
```yaml
spec:
  persistencePolicies:
    bookkeeperEnsemble: 3
    bookkeeperWriteQuorum: 2
    bookkeeperAckQuorum: 2
    managedLedgerMaxMarkDeleteRate: "1.5"
```

### delayedDelivery

The `delayedDelivery` field configures delayed message delivery, allowing messages to be delivered after a specified delay.

| Field | Description | Type | Required |
|-------|-------------|------|----------|
| `active` | Whether delayed delivery is enabled for the topic | bool | No |
| `tickTimeMillis` | Tick time for delayed message delivery in milliseconds | int64 | No |

**Example:**
```yaml
spec:
  delayedDelivery:
    active: true
    tickTimeMillis: 1000  # 1 second
```

### dispatchRate

The `dispatchRate` field configures rate limiting for message delivery to consumers.

| Field | Description | Type | Required |
|-------|-------------|------|----------|
| `dispatchThrottlingRateInMsg` | Maximum number of messages dispatched per rate period | int32 | No |
| `dispatchThrottlingRateInByte` | Maximum number of bytes dispatched per rate period | int64 | No |
| `ratePeriodInSecond` | Rate period in seconds | int32 | No |

**Example:**
```yaml
spec:
  dispatchRate:
    dispatchThrottlingRateInMsg: 1000
    dispatchThrottlingRateInByte: 1048576  # 1MB
    ratePeriodInSecond: 1
```

### publishRate

The `publishRate` field configures rate limiting for message publishing from producers.

| Field | Description | Type | Required |
|-------|-------------|------|----------|
| `publishThrottlingRateInMsg` | Maximum number of messages published per rate period | int32 | No |
| `publishThrottlingRateInByte` | Maximum number of bytes published per rate period | int64 | No |

**Example:**
```yaml
spec:
  publishRate:
    publishThrottlingRateInMsg: 2000
    publishThrottlingRateInByte: 2097152  # 2MB
```

### inactiveTopicPolicies

The `inactiveTopicPolicies` field configures automatic cleanup of inactive topics.

| Field | Description | Type | Required |
|-------|-------------|------|----------|
| `inactiveTopicDeleteMode` | How to delete inactive topics: "delete_when_no_subscriptions" or "delete_when_subscriptions_caught_up" | string | No |
| `maxInactiveDurationInSeconds` | Maximum time in seconds a topic can be inactive before deletion | int32 | No |
| `deleteWhileInactive` | Whether to delete the topic while it's inactive | bool | No |

**Example:**
```yaml
spec:
  inactiveTopicPolicies:
    inactiveTopicDeleteMode: "delete_when_no_subscriptions"
    maxInactiveDurationInSeconds: 3600  # 1 hour
    deleteWhileInactive: true
```

### subscribeRate

The `subscribeRate` field configures rate limiting for new subscription creation on the topic.

| Field | Description | Type | Required |
|-------|-------------|------|----------|
| `subscribeThrottlingRatePerConsumer` | Maximum subscribe rate per consumer (-1 means unlimited) | int32 | No |
| `ratePeriodInSecond` | Time window in seconds for rate calculation (default: 30) | int32 | No |

**Example:**
```yaml
spec:
  subscribeRate:
    subscribeThrottlingRatePerConsumer: 10
    ratePeriodInSecond: 30  # 30 seconds
```

### schemaCompatibilityStrategy

The `schemaCompatibilityStrategy` field defines how schema evolution is handled for the topic. This controls whether and how schemas can be updated while maintaining compatibility with existing producers and consumers.

**Available Strategies:**

| Strategy | Description |
|----------|-------------|
| `UNDEFINED` | Inherit the cluster's default schema compatibility strategy |
| `ALWAYS_INCOMPATIBLE` | New schemas are always incompatible with existing schemas |
| `ALWAYS_COMPATIBLE` | New schemas are always compatible with existing schemas |
| `BACKWARD` | New schema is compatible with the previous schema |
| `FORWARD` | Previous schema is compatible with the new schema |
| `FULL` | New schema is both backward and forward compatible |
| `BACKWARD_TRANSITIVE` | New schema is compatible with all previous schemas |
| `FORWARD_TRANSITIVE` | All previous schemas are compatible with the new schema |
| `FULL_TRANSITIVE` | New schema is backward and forward compatible with all schemas |

**Example:**
```yaml
spec:
  schemaCompatibilityStrategy: BACKWARD
```

### autoSubscriptionCreation

The `autoSubscriptionCreation` field controls whether subscriptions can be created automatically when consumers connect to the topic.

**Example:**
```yaml
spec:
  autoSubscriptionCreation:
    allowAutoSubscriptionCreation: true
```

## SchemaInfo

The `schemaInfo` field in the PulsarTopic specification allows you to define the schema for the topic. For more details about Pulsar schemas, refer to the [official documentation](https://pulsar.apache.org/docs/schema-understand/).

The `schemaInfo` field has the following structure:

| Field | Description | Required |
|-------|-------------|----------|
| `type` | The schema type, which determines how to interpret the schema data. | Yes |
| `schema` | The schema definition. For AVRO and JSON schemas, this should be a JSON string of the schema definition. For PROTOBUF_NATIVE schemas, this should be the JSON string of protobuf definition string from [ProtobufNativeSchemaData](https://github.com/apache/pulsar/blob/master/pulsar-common/src/main/java/org/apache/pulsar/common/protocol/schema/ProtobufNativeSchemaData.java). | Yes |
| `properties` | A map of user-defined properties as key-value pairs. Applications can use this for carrying any application-specific logic. | No |

### Supported Schema Types

Pulsar supports various schema types, including:

- AVRO
- JSON
- PROTOBUF
- PROTOBUF_NATIVE
- THRIFT
- BOOLEAN
- INT8
- INT16
- INT32
- INT64
- FLOAT
- DOUBLE
- STRING
- BYTES
- DATE
- TIME
- TIMESTAMP
- INSTANT
- LOCAL_DATE
- LOCAL_TIME
- LOCAL_DATE_TIME

For more detailed information about these schema types and their usage, please refer to the [Pulsar Schema documentation](https://pulsar.apache.org/docs/schema-understand/).

### Example

Here's an example of a PulsarTopic resource with a JSON schema:

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarTopic
metadata:
  name: "test-pulsar-topic123"
  namespace: test
spec:
  name: persistent://test-tenant/testns/topic123
  connectionRef:
    name: "test-pulsar-connection"
  partitions: 1
  schemaInfo:
    type: "JSON"
    schema: "{\"type\":\"record\",\"name\":\"Example\",\"namespace\":\"test\",\"fields\":[{\"name\":\"ID\",\"type\":\"int\"},{\"name\":\"Name\",\"type\":\"string\"}]}"
    properties:
      "owner": "pulsar"
```

This example defines a JSON schema with two fields, `ID` and `Name`, both of which are required. The `type` field is set to `JSON`, indicating that the schema is in JSON format. The `schema` field contains the actual JSON schema definition. The `properties` field is optional and can be used to add any application-specific logic.

## Complete Example with Advanced Policies

Here's a comprehensive example of a PulsarTopic resource that demonstrates the use of the new topic-level policies:

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarTopic
metadata:
  name: "advanced-pulsar-topic"
  namespace: production
spec:
  name: persistent://production-tenant/high-throughput/events
  connectionRef:
    name: "production-pulsar-connection"
  persistent: true
  partitions: 4
  maxProducers: 10
  maxConsumers: 20
  messageTTL: 7d
  retentionTime: 30d
  retentionSize: 100Gi
  compactionThreshold: 1073741824  # 1GB
  
  # Advanced persistence configuration for high availability
  persistencePolicies:
    bookkeeperEnsemble: 5
    bookkeeperWriteQuorum: 3
    bookkeeperAckQuorum: 2
    managedLedgerMaxMarkDeleteRate: "2.0"
  
  # Enable delayed delivery for scheduled messages
  delayedDelivery:
    active: true
    tickTimeMillis: 1000
  
  # Rate limiting for consumer message delivery
  dispatchRate:
    dispatchThrottlingRateInMsg: 10000
    dispatchThrottlingRateInByte: 10485760  # 10MB
    ratePeriodInSecond: 1
  
  # Rate limiting for producer message publishing
  publishRate:
    publishThrottlingRateInMsg: 5000
    publishThrottlingRateInByte: 5242880   # 5MB
  
  # Automatic cleanup of inactive topics
  inactiveTopicPolicies:
    inactiveTopicDeleteMode: "delete_when_no_subscriptions"
    maxInactiveDurationInSeconds: 86400  # 24 hours
    deleteWhileInactive: false
  
  # Message deduplication
  deduplication: true
  
  # Subscription rate limiting
  subscribeRate:
    subscribeThrottlingRatePerConsumer: 5
    ratePeriodInSecond: 30
  
  # Message size limits and consumer/subscription limits
  maxMessageSize: 1048576  # 1MB
  maxConsumersPerSubscription: 5
  maxSubscriptionsPerTopic: 100
  
  # Schema validation and compatibility
  schemaValidationEnforced: true
  schemaCompatibilityStrategy: BACKWARD
  
  # Additional dispatch rate controls
  subscriptionDispatchRate:
    dispatchThrottlingRateInMsg: 8000
    dispatchThrottlingRateInByte: 8388608  # 8MB
    ratePeriodInSecond: 1
    
  replicatorDispatchRate:
    dispatchThrottlingRateInMsg: 3000
    dispatchThrottlingRateInByte: 3145728  # 3MB
    ratePeriodInSecond: 1
  
  # Deduplication configuration
  deduplicationSnapshotInterval: 1000
  
  # Auto subscription creation
  autoSubscriptionCreation:
    allowAutoSubscriptionCreation: false
  
  # Custom properties
  properties:
    "environment": "production"
    "team": "data-platform"
    "cost-center": "engineering"
  
  # Schema definition
  schemaInfo:
    type: "JSON"
    schema: "{\"type\":\"record\",\"name\":\"Event\",\"fields\":[{\"name\":\"id\",\"type\":\"string\"},{\"name\":\"timestamp\",\"type\":\"long\"},{\"name\":\"data\",\"type\":\"string\"}]}"
    properties:
      "version": "1.0"
      "owner": "data-platform-team"

  lifecyclePolicy: CleanUpAfterDeletion
```

This example demonstrates a production-ready topic configuration with:
- High availability persistence settings (5 bookies ensemble, 3 write quorum)
- Rate limiting for producers, consumers, subscriptions, and replicators
- Delayed delivery capability for scheduled messages  
- Automatic topic compaction at 1GB
- Inactive topic cleanup after 24 hours
- Message deduplication enabled with snapshot intervals
- JSON schema enforcement with backward compatibility
- Subscription rate limiting (5 subscriptions per consumer per 30 seconds)
- Message size limits (1MB maximum)
- Consumer and subscription limits per topic
- Custom properties for metadata tracking
- Controlled auto-subscription creation (disabled for production)