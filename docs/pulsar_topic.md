# PulsarTopic

## Overview

The `PulsarTopic` resource defines a topic in a Pulsar cluster. It allows you to configure various topic-level settings such as persistence, partitions, retention policies, and schema information. This resource is part of the Pulsar Resources Operator, which enables declarative management of Pulsar resources using Kubernetes custom resources.

## Sepcifications

## Specifications

| Field | Description | Required |
|-------|-------------|----------|
| `name` | The fully qualified topic name in the format "persistent://tenant/namespace/topic" or "non-persistent://tenant/namespace/topic". | Yes |
| `connectionRef` | Reference to the PulsarConnection resource used to connect to the Pulsar cluster for this topic. | Yes |
| `persistent` | Whether the topic is persistent or non-persistent. Default is false. Can also be set by topic name prefix. | No |
| `partitions` | Number of partitions for the topic. Default is 0. | No |
| `maxProducers` | Maximum number of producers allowed on the topic. | No |
| `maxConsumers` | Maximum number of consumers allowed on the topic. | No |
| `messageTTL` | Time to Live (TTL) for messages in the topic. Messages older than this TTL will be automatically marked as consumed. | No |
| `maxUnAckedMessagesPerConsumer` | Maximum number of unacknowledged messages allowed per consumer. | No |
| `maxUnAckedMessagesPerSubscription` | Maximum number of unacknowledged messages allowed per subscription. | No |
| `retentionTime` | Minimum time to retain messages in the topic. Should be set in conjunction with retentionSize for effective retention policy. | No |
| `retentionSize` | Maximum size of backlog retained in the topic. Should be set in conjunction with retentionTime for effective retention policy. | No |
| `backlogQuotaLimitTime` | Time limit for message backlog. Messages older than this limit will be removed or handled according to the retention policy. | No |
| `backlogQuotaLimitSize` | Size limit for message backlog. When the limit is reached, older messages will be removed or handled according to the retention policy. | No |
| `backlogQuotaRetentionPolicy` | Retention policy for messages when backlog quota is exceeded. Options: "producer_request_hold", "producer_exception", or "consumer_backlog_eviction". | No |
| `lifecyclePolicy` | Determines whether to keep or delete the Pulsar topic when the Kubernetes resource is deleted. Options: `CleanUpAfterDeletion`, `KeepAfterDeletion`. Default is `CleanUpAfterDeletion`. | No |
| `schemaInfo` | Schema information for the topic. See [schemaInfo](#schemainfo) for more details. | No |
| `geoReplicationRefs` | List of references to PulsarGeoReplication resources, used to enable geo-replication at the topic level. | No |

Note: Valid time units for duration fields are "s" (seconds), "m" (minutes), "h" (hours), "d" (days), "w" (weeks).

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

2. Other fields such as `partitions`, `maxProducers`, `maxConsumers`, `messageTTL`, `retentionTime`, `retentionSize`, `backlogQuotaLimitTime`, `backlogQuotaLimitSize`, and `backlogQuotaRetentionPolicy` can be modified.

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

## SchemaInfo

The `schemaInfo` field in the PulsarTopic specification allows you to define the schema for the topic. For more details about Pulsar schemas, refer to the [official documentation](https://pulsar.apache.org/docs/schema-understand/).

The `schemaInfo` field has the following structure:

| Field | Description | Required |
|-------|-------------|----------|
| `type` | The schema type, which determines how to interpret the schema data. | Yes |
| `schema` | The schema definition, which is a base64 encoded string representing the schema. | Yes |
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