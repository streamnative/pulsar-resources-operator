# PulsarSink

## Create PulsarSink

1. Define a sink named `test-pulsar-sink` by using the YAML file and save the YAML file `sink.yaml`.

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarSink
metadata:
  name: test-pulsar-sink
  namespace: default
spec:
  autoAck: true
  className: org.apache.pulsar.io.datagenerator.DataGeneratorPrintSink
  cleanupSubscription: false
  connectionRef:
    name: "test-pulsar-connection"
  customRuntimeOptions: {}
  inputs:
  - sink-input
  archive:
    url: builtin://data-generator
  lifecyclePolicy: CleanUpAfterDeletion
  name: test-pulsar-sink
  namespace: default
  parallelism: 1
  processingGuarantees: EFFECTIVELY_ONCE
  secrets:
    SECRET1:
      key: hello
      name: sectest
  sourceSubscriptionPosition: Latest
  tenant: public
```

This table lists specifications available for the `PulsarSink` resource.

| Option                         | Description                                                                                                                                           | Required or not |
|--------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------|
| `topicsPattern`                | The topic pattern.                                                                                                                                    | Optional        |
| `resources`                    | The resources of the sink.                                                                                                                            | Optional        |
| `timeoutMs`                    | The timeout in milliseconds.                                                                                                                          | Optional        |
| `cleanupSubscription`          | The cleanup subscription.                                                                                                                             | Optional        |
| `retainOrdering`               | The retain ordering.                                                                                                                                  | Optional        |
| `retainKeyOrdering`            | The retain key ordering.                                                                                                                              | Optional        |
| `autoAck`                      | The auto ack.                                                                                                                                         | Optional        |
| `parallelism`                  | The parallelism.                                                                                                                                      | Optional        |
| `tenant`                       | The tenant.                                                                                                                                           | Required        |
| `namespace`                    | The namespace.                                                                                                                                        | Required        |
| `name`                         | The name.                                                                                                                                             | Required        |
| `className`                    | The class name.                                                                                                                                       | Required        |
| `sinkType`                     | The sink type.                                                                                                                                        | Optional        |
| `archive`                      | The package url of sink.                                                                                                                              | Required        |
| `processingGuarantees`         | The processing guarantees.                                                                                                                            | Optional        |
| `sourceSubscriptionName`       | The source subscription name.                                                                                                                         | Optional        |
| `sourceSubscriptionPosition`   | The source subscription position.                                                                                                                     | Optional        |
| `runtimeFlags`                 | The runtime flags.                                                                                                                                    | Optional        |
| `inputs`                       | The input topics.                                                                                                                                     | Optional        |
| `topicToSerdeClassName`        | the map of topic to serde class name of the PulsarSink.                                                                                               | Optional        |
| `topicToSchemaType`            | the map of topic to schema type of the PulsarSink.                                                                                                    | Optional        |
| `inputSpecs`                   | the map of input specs of the PulsarSink.                                                                                                             | Optional        |
| `configs`                      | the map of configs of the PulsarSink.                                                                                                                 | Optional        |
| `topicToSchemaProperties`      | the map of topic to schema properties of the PulsarSink.                                                                                              | Optional        |
| `customRuntimeOptions`         | the map of custom runtime options of the PulsarSink.                                                                                                  | Optional        |
| `secrets`                      | the map of secrets of the PulsarSink.                                                                                                                 | Optional        |
| `maxMessageRetries`            | the max message retries of the PulsarSink.                                                                                                            | Optional        |
| `deadLetterTopic`              | the dead letter topic of the PulsarSink.                                                                                                              | Optional        |
| `negativeAckRedeliveryDelayMs` | the negative ack redelivery delay in milliseconds of the PulsarSink.                                                                                  | Optional        |
| `transformFunction`            | the transform function of the PulsarSink.                                                                                                             | Optional        |
| `transformFunctionClassName`   | the transform function class name of the PulsarSink.                                                                                                  | Optional        |
| `transformFunctionConfig`      | the transform function config of the PulsarSink.                                                                                                      | Optional        |
| `connectionRef`                | The reference to a PulsarConnection.                                                                                                                  | Required        |
| `lifecyclePolicy`              | The resource lifecycle policy. Available options are `CleanUpAfterDeletion` and `KeepAfterDeletion`. By default, it is set to `CleanUpAfterDeletion`. | Optional        |

2. Apply the YAML file to create the sink.

```shell
kubectl apply -f sink.yaml
```

3. Check the resource status. When column Ready is true, it indicates the resource is created successfully in the pulsar cluster

```shell
kubectl get pulsarsink
```

```
NAME               RESOURCE_NAME          GENERATION   OBSERVED_GENERATION   READY
test-pulsar-sink   test-pulsar-sink       2            2                     True
```

## Update PulsarSink

You can update the sink by editing the sink.yaml, then apply it again. For example, if you want to update the parallelism of the sink, you can edit the sink.yaml as follows:

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarSink
metadata:
  name: test-pulsar-sink
  namespace: default
spec:
  autoAck: true
  className: org.apache.pulsar.io.datagenerator.DataGeneratorPrintSink
  cleanupSubscription: false
  connectionRef:
    name: "test-pulsar-connection"
  customRuntimeOptions: {}
  inputs:
  - sink-input
  archive:
    url: builtin://data-generator
  lifecyclePolicy: CleanUpAfterDeletion
  name: test-pulsar-sink
  namespace: default
  parallelism: 2
  processingGuarantees: EFFECTIVELY_ONCE
  secrets:
    SECRET1:
      key: hello
      name: sectest
  sourceSubscriptionPosition: Latest
  tenant: public
```

```shell
kubectl apply -f sink.yaml
```

## Delete PulsarSink

```shell
kubectl delete pulsarsink test-pulsar-sink
```
