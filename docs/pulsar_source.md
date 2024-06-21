# PulsarSource

## Create PulsarSource

1. Define a source named `test-pulsar-source` by using the YAML file and save the YAML file `source.yaml`.

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarSource
metadata:
  name: test-pulsar-source
  namespace: default
spec:
  className: org.apache.pulsar.io.datagenerator.DataGeneratorSource
  connectionRef:
    name: "test-pulsar-connection"
  customRuntimeOptions: 
    sleepBetweenMessages: "1000"
  topicName: sink-input
  archive:
    url: builtin://data-generator
  configs:
    sleepBetweenMessages: "1000"
  lifecyclePolicy: CleanUpAfterDeletion
  name: test-pulsar-source
  namespace: default
  parallelism: 1
  processingGuarantees: ATLEAST_ONCE
  secrets:
    SECRET1:
      key: hello
      name: sectest
  tenant: public
```

This table lists specifications available for the `PulsarSource` resource.

| Option                         | Description                                                                                                                                           | Required or not |
|--------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------|
| `tenant`                       | The tenant.                                                                                                                                           | Required        |
| `namespace`                    | The namespace.                                                                                                                                        | Required        |
| `name`                         | The name.                                                                                                                                             | Required        |
| `className`                    | The class name.                                                                                                                                       | Required        |
| `producerConfig`               | The producer config.                                                                                                                                  | Optional        |
| `topicName`                    | The topic name.                                                                                                                                       | Required        |
| `serdeClassName`               | The serde class name.                                                                                                                                 | Optional        |
| `schemaType`                   | The schema type.                                                                                                                                      | Optional        |
| `configs`                      | The map of configs of the PulsarSource.                                                                                                                                          | Optional        |
| `secrets`                      | The map of secrets of the PulsarSource.                                                                                                                                          | Optional        |
| `parallelism`                  | The parallelism.                                                                                                                                      | Optional        |
| `processingGuarantees`         | The processing guarantees.                                                                                                                            | Optional        |
| `resources`                    | The resources of the source.                                                                                                                         | Optional        |
| `archive`                      | The package url of source.                                                                                                                           | Required        |
| `runtimeFlags`                 | The runtime flags.                                                                                                                                    | Optional        |
| `customRuntimeOptions`         | The custom runtime options.                                                                                                                          | Optional        |
| `batchSourceConfig`            | The batch source config.                                                                                                                             | Optional        |
| `batchBuilder`                 | The batch builder.                                                                                                                                    | Optional        |
| `connectionRef`                | The reference to a PulsarConnection.                                                                                                                 | Required        |
| `lifecyclePolicy`              | The resource lifecycle policy. Available options are `CleanUpAfterDeletion` and `KeepAfterDeletion`. By default, it is set to `CleanUpAfterDeletion`. | Optional        |

2. Apply the YAML file to create the source.

```shell
kubectl apply -f source.yaml
```

3. Check the resource status. When column Ready is true, it indicates the resource is created successfully in the pulsar cluster

```shell
kubectl get pulsarsource
```

```
NAME                 RESOURCE_NAME         GENERATION   OBSERVED_GENERATION   READY
test-pulsar-source   test-pulsar-source    1            1                     True
```

## Update PulsarSource

You can update the source by editing the source.yaml, then apply it again. For example, if you want to update the parallelism of the source, you can edit the source.yaml as follows:

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarSource
metadata:
  name: test-pulsar-source
  namespace: default
spec:
  className: org.apache.pulsar.io.datagenerator.DataGeneratorSource
  connectionRef:
    name: "test-pulsar-connection"
  customRuntimeOptions: 
    sleepBetweenMessages: "1000"
  topicName: sink-input
  archive:
    url: builtin://data-generator
  configs:
    sleepBetweenMessages: "1000"
  lifecyclePolicy: CleanUpAfterDeletion
  name: test-pulsar-source
  namespace: default
  parallelism: 2
  processingGuarantees: ATLEAST_ONCE
  secrets:
    SECRET1:
      key: hello
      name: sectest
  tenant: public
```

```shell
kubectl apply -f source.yaml
```

## Delete PulsarSource
    
```shell
kubectl delete pulsarsource test-pulsar-source
```