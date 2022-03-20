# PulsarNamespace

## Create PulsarNamespace

1. Define a namespace named `test-tenant/testns` by using the YAML file and save the YAML file `namespace.yaml`.
```yaml
apiVersion: pulsar.streamnative.io/v1alpha1
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

This table lists specifications available for the `PulsarNamespace` resource.

| Option | Description | Required or not |
| ---| --- |--- |
| `name` | The namespace name. | Yes |
| `connectionRef` | The reference to a PulsarConnection. | Yes |
|  `backlogQuotaLimitTime` | The Backlog quota time limit (in seconds). | Optional |
| `backlogQuotaLimitSize` | The Backlog quota size limit (such as 10Mi, 10Gi). | Optional |
| `backlogQuotaRetentionPolicy` | The Retention policy to be enforced when the limit is reached. options: producer_exception,producer_request_hold,consumer_backlog_eviction | Optional |
| `messageTTL` | The TTL time duration of messages. | Optional |
| `bundles` | The number of activated bundles. By default, it is set to 4. It couldn’t be updated after namespace is created | Optional |
| `maxProducersPerTopic` | The maximum number of producers per topic for a namespace.| Optional |
| `maxConsumersPerTopic` | The maximum number of consumers per topic for a namespace. | Optional |
| `maxConsumersPerSubscription` | The maximum number of consumers per subscription for a namespace. | Optional |
| `retentionTime` | The retention time (in minutes, hours, days, or weeks). | Optional |
| `retentionSize` | The retention size limit(such as 800Mi, 10Gi). | Optional |
| `lifecyclePolicy` | The resource lifecycle policy, CleanUpAfterDeletion or KeepAfterDeletion, the default is KeepAfterDeletion | Optional |


2. Apply the YAML file to create the namespace.

```shell
kubectl apply -f namespace.yaml
```

3. Check the resource status. When column Ready is true, it indicates the resource is created successfully in the pulsar cluster

```shell
kubectl -n test get pns/pulsarnamespace
```

```shell
NAME                    RESOURCE_NAME        GENERATION   OBSERVED_GENERATION   READY
test-pulsar-namespace   test-tenant/testns              1                                   1                               True
```

## Update PulsarNamespace

You can update the namespace policies by editing the namespace.yaml, then apply if again. Please be noticed:
The field `name` and `bundles` couldn’t be updated.

## Delete PulsarNamespace

```shell
kubectl -n test delete pns test-pulsar-namespace
```

Please be noticed, when you delete the namespace, the real namespace will still exist if the `lifecyclePolicy` is `KeepAfterDeletion`
