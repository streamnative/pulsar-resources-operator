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
| `messageTTL` | The TTL time duration of messages, eg 10m, 1h.| Optional |
| `maxUnAckedMessagesPerConsumer` | The maximum number of unacked messages per consumer. | Optional |
| `maxUnAckedMessagesPerSubscription` | The maximum number of unacked messages per subscription. | Optional |
| `retentionTime` | The retention time (in minutes, hours, days, or weeks). | Optional |
| `retentionSize` | The retention size limit. | Optional |
|  `backlogQuotaLimitTime` | The Backlog quota time limit (in seconds). | Optional |
| `backlogQuotaLimitSize` | The Backlog quota size limit (such as 10Mi, 10Gi). | Optional |
| `backlogQuotaRetentionPolicy` | The Retention policy to be enforced when the limit is reached. | Optional |
| `lifecyclePolicy` | The resource lifecycle policy, CleanUpAfterDeletion or KeepAfterDeletion, the default is KeepAfterDeletion | Optional |

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

## Update PulsarTopic
You can update the topic policies by editing the topic.yaml, then apply if again. 

Please be noticed:
1. The field `name`, `persistent` and `partitions` couldn’t be updated.


## Delete PulsarTopic

```shell
kubectl -n test delete pulsartopic.resource.streamnative.io test-pulsar-topic
```

Please be noticed, when you delete the permission, the real permission will still exist if the `lifecyclePolicy` is `KeepAfterDeletion`


