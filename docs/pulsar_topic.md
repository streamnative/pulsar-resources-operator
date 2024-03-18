# PulsarTopic

## Create PulsarTopic

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

This table lists specifications available for the `PulsarTopic` resource.

| Option | Description | Required or not |
| ---| --- |--- |
| `name` | The topic name. | Yes |
| `connectionRef` | The reference to a PulsarConnection. | Yes |
| `persistent` | Set whether it is a persistent or non-persistent topic, you also can set it by topic name prefix with persistent or non-persistent. default is false| Optional |
| `partitions` | The number of partitions for the topic. default is 0 | Optional |
| `maxProducers` | The maximum number of  producers for a topic. | Optional |
| `messageTTL` | The TTL time duration of messages (valid time units are "s", "m", "h", "d", "w").| Optional |
| `maxUnAckedMessagesPerConsumer` | The maximum number of unacked messages per consumer. | Optional |
| `maxUnAckedMessagesPerSubscription` | The maximum number of unacked messages per subscription. | Optional |
| `retentionTime` | The retention time (valid time units are "s", "m", "h", "d", "w"). | Optional |
| `retentionSize` | The retention size limit. | Optional |
| `backlogQuotaLimitTime` | The Backlog quota time limit (valid time units are "s", "m", "h", "d", "w"). | Optional |
| `backlogQuotaLimitSize` | The Backlog quota size limit (such as 10Mi, 10Gi). | Optional |
| `backlogQuotaRetentionPolicy` | The Retention policy to be enforced when the limit is reached. | Optional |
| `lifecyclePolicy` | The resource lifecycle policy. Available options are `CleanUpAfterDeletion` and `KeepAfterDeletion`. By default, it is set to `CleanUpAfterDeletion`. | Optional |
| `schemaInfo` | The schema of pulsar topic, default is nil. More details you can find in [schemaInfo](#schemainfo) Optional |
| `geoReplicationRefs` | The reference list of the PulsarGeoReplication. Enable Geo-replication at the topic level. It will add the topic to the clusters. | No |

1. Apply the YAML file to create the topic.

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

## Update PulsarTopic
You can update the topic policies by editing the topic.yaml, then apply if again. 

Please be noticed:
1. The field `name`, `persistent` couldn’t be updated.


## Delete PulsarTopic

```shell
kubectl -n test delete pulsartopic.resource.streamnative.io test-pulsar-topic
```

Please be noticed, when you delete the permission, the real permission will still exist if the `lifecyclePolicy` is `KeepAfterDeletion`



## SchemaInfo

More details about pulsar schema, you can check the [official document](https://pulsar.apache.org/docs/2.10.x/schema-understand/)

| Option | Description |
| ---| --- |
| `type` | Schema type, which determines how to interpret the schema data |
| `schema` | Schema data, which is a sequence of 8-bit unsigned bytes and schema-type specific |
| `properties` | It is a user defined properties as a string/string map. Applications can use this bag for carrying any application specific logics. |

### Example

This is a demo that uses the json type schema.

```golang
type testJSON struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
```

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