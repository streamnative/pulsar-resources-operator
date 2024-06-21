# PulsarFunction

## Create PulsarFunction

1. Define a function named `test-pulsar-function` by using the YAML file and save the YAML file `function.yaml`.

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarFunction
metadata:
  name: test-pulsar-function
  namespace: default
spec:
  autoAck: true
  className: org.apache.pulsar.functions.api.examples.ExclamationFunction
  cleanupSubscription: true
  connectionRef:
    name: "test-pulsar-connection"
  customRuntimeOptions: {}
  deadLetterTopic: dl-topic
  exposePulsarAdminClientEnabled: false
  forwardSourceMessageProperty: true
  inputs:
  - input
  jar:
    url: function://public/default/api-examples@v3.2.3.3
  lifecyclePolicy: CleanUpAfterDeletion
  logTopic: func-log
  maxMessageRetries: 101
  name: test-pulsar-function
  namespace: default
  output: output
  parallelism: 1
  processingGuarantees: ATLEAST_ONCE
  retainKeyOrdering: true
  retainOrdering: false
  secrets:
    SECRET1:
      key: hello
      name: sectest
  skipToLatest: true
  subName: test-sub
  subscriptionPosition: Latest
  tenant: public
  timeoutMs: 6666
```

This table lists specifications available for the `PulsarFunction` resource.

| Option                           | Description                                                                                                                                           | Required or not |
|----------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------|
| `autoAck`                        | Whether to automatically acknowledge messages.                                                                                                        | Optional        |
| `className`                      | The class name of the function.                                                                                                                       | Yes             |
| `cleanupSubscription`            | Whether to clean up the subscription when the function is deleted.                                                                                    | Optional        |
| `connectionRef`                  | The reference to a PulsarConnection.                                                                                                                  | Yes             |
| `customRuntimeOptions`           | Custom runtime options.                                                                                                                               | Optional        |
| `deadLetterTopic`                | The dead letter topic.                                                                                                                                | Optional        |
| `exposePulsarAdminClientEnabled` | Whether to expose the Pulsar admin client.                                                                                                            | Optional        |
| `forwardSourceMessageProperty`   | Whether to forward the source message property.                                                                                                       | Optional        |
| `inputs`                         | The input topics.                                                                                                                                     | Yes             |
| `jar`                            | The JAR package URL, can be used by Java runtime.                                                                                                     | Optional        |
| `py`                             | The Python package URL, can be used by Java runtime.                                                                                                  | Optional        |
| `go`                             | The Go package URL, can be used by Java runtime.                                                                                                      | Optional        |
| `lifecyclePolicy`                | The resource lifecycle policy. Available options are `CleanUpAfterDeletion` and `KeepAfterDeletion`. By default, it is set to `CleanUpAfterDeletion`. | Optional        |
| `logTopic`                       | The log topic.                                                                                                                                        | Optional        |
| `maxMessageRetries`              | The maximum number of message retries.                                                                                                                | Optional        |
| `name`                           | The function name.                                                                                                                                    | Yes             |
| `namespace`                      | The namespace.                                                                                                                                        | Yes             |
| `output`                         | The output topic.                                                                                                                                     | Yes             |
| `parallelism`                    | The parallelism.                                                                                                                                      | Optional        |
| `processingGuarantees`           | The processing guarantees.                                                                                                                            | Optional        |
| `retainKeyOrdering`              | Whether to retain key ordering.                                                                                                                       | Optional        |
| `retainOrdering`                 | Whether to retain ordering.                                                                                                                           | Optional        |
| `secrets`                        | The secrets.                                                                                                                                          | Optional        |
| `skipToLatest`                   | Whether to skip to the latest.                                                                                                                        | Optional        |
| `subName`                        | The subscription name.                                                                                                                                | Optional        |
| `subscriptionPosition`           | The subscription position.                                                                                                                            | Optional        |
| `tenant`                         | The tenant.                                                                                                                                           | Yes             |
| `timeoutMs`                      | The timeout in milliseconds.                                                                                                                          | Optional        |
| `topicsPattern`                  | The topics pattern.                                                                                                                                   | Optional        |
| `batchBuilder`                   | The batch builder.                                                                                                                                    | Optional        |
| `producerConfig`                 | The producer configuration.                                                                                                                           | Optional        |
| `customSchemaOutputs`            | The custom schema outputs.                                                                                                                            | Optional        |
| `outputSerdeClassName`           | The output serde class name.                                                                                                                          | Optional        |
| `outputSchemaType`               | The output schema type.                                                                                                                               | Optional        |
| `outputTypeClassName`            | The output type class name of the function.                                                                                                           | Optional        |
| `runtimeFlags`                   | The runtime flags.                                                                                                                                    | Optional        |
| `resources`                      | The resources.                                                                                                                                        | Optional        |
| `windowConfig`                   | The window configuration.                                                                                                                             | Optional        |
| `userConfig`                     | The user configuration.                                                                                                                               | Optional        |
| `customSerdeInputs`              | The custom serde inputs.                                                                                                                              | Optional        |
| `customSchemaInputs`             | The custom schema inputs.                                                                                                                             | Optional        |
| `inputSpecs`                     | The input specs.                                                                                                                                      | Optional        |
| `inputTypeClassName`             | The input type class name of the function.                                                                                                            | Optional        |
| `maxPendingAsyncRequests`        | The maximum number of pending async requests.                                                                                                         | Optional        |
| `exposePulsarAdminClientEnabled` | Whether to expose the Pulsar admin client.                                                                                                            | Optional        |
| `skipToLatest`                   | Whether to skip to the latest.                                                                                                                        | Optional        |

2. Apply the YAML file to create the function.

```shell
kubectl apply -f function.yaml
```

3. Check the resource status. When column Ready is true, it indicates the resource is created successfully in the pulsar cluster

```shell
kubectl get pulsarfunction
```

```
NAME                   RESOURCE_NAME              GENERATION   OBSERVED_GENERATION   READY
test-pulsar-function   test-pulsar-function       1            1                     True
```

## Update PulsarFunction

You can update the function by editing the function.yaml, then apply it again. 

For example, if you want to update the parallelism of the function, you can edit the function.yaml as follows:

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarFunction
metadata:
  name: test-pulsar-function
  namespace: default
spec:
  autoAck: true
  className: org.apache.pulsar.functions.api.examples.ExclamationFunction
  cleanupSubscription: true
  connectionRef:
    name: "test-pulsar-connection"
  customRuntimeOptions: {}
  deadLetterTopic: dl-topic
  exposePulsarAdminClientEnabled: false
  forwardSourceMessageProperty: true
  inputs:
  - input
  jar:
    url: function://public/default/api-examples@v3.2.3.3
  lifecyclePolicy: CleanUpAfterDeletion
  logTopic: func-log
  maxMessageRetries: 101
  name: test-pulsar-function
  namespace: default
  output: output
  parallelism: 2
  processingGuarantees: ATLEAST_ONCE
  retainKeyOrdering: true
  retainOrdering: false
  secrets:
    SECRET1:
      key: hello
      name: sectest
  skipToLatest: true
  subName: test-sub
  subscriptionPosition: Latest
  tenant: public
  timeoutMs: 6666
```

```shell
kubectl apply -f function.yaml
```

## Delete PulsarFunction

```shell
kubectl delete pulsarfunction test-pulsar-function
```
