# ComputeFlinkDeployment

## Overview

The `ComputeFlinkDeployment` resource defines a Flink deployment in StreamNative Cloud. It supports both Ververica Platform (VVP) and Community deployment templates, allowing you to deploy and manage Flink applications.

## Specifications

| Field                | Description                                                                                | Required |
|----------------------|--------------------------------------------------------------------------------------------|----------|
| `apiServerRef`       | Reference to the StreamNativeCloudConnection resource for API server access. If not specified, the APIServerRef from the referenced ComputeWorkspace will be used. | No       |
| `workspaceName`      | Name of the ComputeWorkspace where the Flink deployment will run                           | Yes      |
| `labels`             | Labels to add to the deployment                                                             | No       |
| `annotations`        | Annotations to add to the deployment                                                        | No       |
| `template`           | VVP deployment template configuration                                                       | No*      |
| `communityTemplate`  | Community deployment template configuration                                                 | No*      |
| `defaultPulsarCluster`| Default Pulsar cluster to use for the deployment                                          | No       |
| `configuration`      | Additional configuration for the Flink deployment, including environment variables and secrets | No       |
| `imagePullSecrets`   | List of image pull secrets to use for the deployment                                       | No       |

*Note: Either `template` or `communityTemplate` must be specified, but not both.

## APIServerRef Inheritance

The `ComputeFlinkDeployment` resource can inherit the `APIServerRef` from its referenced `ComputeWorkspace`. This simplifies configuration and reduces duplication. Here's how it works:

1. If `apiServerRef` is specified in the `ComputeFlinkDeployment`, that value will be used.
2. If `apiServerRef` is not specified, the operator will use the `APIServerRef` from the referenced `ComputeWorkspace`.
3. The `workspaceName` field is required and must reference a valid `ComputeWorkspace` in the same namespace.

This inheritance mechanism allows you to:
- Reduce configuration duplication
- Centralize API server configuration at the workspace level
- Easily change API server configuration for multiple deployments by updating the workspace

### Configuration Structure

| Field     | Description                                                                                | Required |
|-----------|--------------------------------------------------------------------------------------------|----------|
| `envs`    | List of environment variables to set in the Flink deployment                               | No       |
| `secrets` | List of secrets referenced to deploy with the Flink deployment                             | No       |

#### EnvVar Structure

| Field   | Description                                                                                | Required |
|---------|--------------------------------------------------------------------------------------------|----------|
| `name`  | Name of the environment variable                                                           | Yes      |
| `value` | Value of the environment variable                                                          | Yes      |

#### SecretReference Structure

| Field       | Description                                                                                | Required |
|-------------|--------------------------------------------------------------------------------------------|----------|
| `name`      | Name of the ENV variable                                                                   | Yes      |
| `valueFrom` | References a secret in the same namespace                                                  | Yes      |

### VVP Deployment Template

| Field           | Description                                                                                | Required |
|-----------------|--------------------------------------------------------------------------------------------|----------|
| `syncingMode`   | How the deployment should be synced (e.g., PATCH)                                          | No       |
| `deployment`    | VVP deployment configuration                                                                | Yes      |

#### VVP Deployment Configuration

| Field                      | Description                                                                                | Required |
|----------------------------|--------------------------------------------------------------------------------------------|----------|
| `userMetadata`             | Metadata for the deployment (name, namespace, displayName, etc.)                           | Yes      |
| `spec`                     | Deployment specification including state, target, resources, etc.                           | Yes      |

##### Deployment Spec Fields

| Field                          | Description                                                                            | Required |
|--------------------------------|----------------------------------------------------------------------------------------|----------|
| `deploymentTargetName`         | Target name for the deployment                                                         | No       |
| `state`                        | State of the deployment (RUNNING, SUSPENDED, CANCELLED)                                 | No       |
| `maxJobCreationAttempts`       | Maximum number of job creation attempts (minimum: 1)                                   | No       |
| `maxSavepointCreationAttempts` | Maximum number of savepoint creation attempts (minimum: 1)                             | No       |
| `template`                     | Deployment template configuration                                                       | Yes      |

##### Template Spec Fields

| Field                | Description                                                                            | Required |
|----------------------|----------------------------------------------------------------------------------------|----------|
| `artifact`           | Deployment artifact configuration                                                       | Yes      |
| `flinkConfiguration` | Flink configuration key-value pairs                                                    | No       |
| `parallelism`        | Parallelism of the Flink job                                                          | No       |
| `numberOfTaskManagers`| Number of task managers                                                               | No       |
| `resources`          | Resource requirements for jobmanager and taskmanager                                   | No       |
| `logging`            | Logging configuration                                                                  | No       |

##### Artifact Configuration

| Field                    | Description                                                                            | Required |
|--------------------------|----------------------------------------------------------------------------------------|----------|
| `kind`                   | Type of artifact (JAR, PYTHON, sqlscript)                                              | Yes      |
| `jarUri`                 | URI of the JAR file                                                                    | No*      |
| `pythonArtifactUri`      | URI of the Python artifact                                                             | No*      |
| `sqlScript`              | SQL script content                                                                      | No*      |
| `flinkVersion`           | Flink version to use                                                                   | No       |
| `flinkImageTag`          | Flink image tag to use                                                                 | No       |
| `mainArgs`               | Arguments for the main class/method                                                     | No       |
| `entryClass`             | Entry class for JAR artifacts                                                          | No       |

*Note: One of `jarUri`, `pythonArtifactUri`, or `sqlScript` must be specified based on the `kind`.

### Community Deployment Template

| Field                    | Description                                                                            | Required |
|--------------------------|----------------------------------------------------------------------------------------|----------|
| `metadata`               | Metadata for the deployment (annotations, labels)                                       | No       |
| `spec`                   | Community deployment specification                                                      | Yes      |

#### Community Deployment Spec

| Field                    | Description                                                                            | Required |
|--------------------------|----------------------------------------------------------------------------------------|----------|
| `image`                  | Flink image to use                                                                     | Yes      |
| `jarUri`                 | URI of the JAR file                                                                    | Yes      |
| `entryClass`             | Entry class of the JAR                                                                 | No       |
| `mainArgs`               | Main arguments for the application                                                      | No       |
| `flinkConfiguration`     | Flink configuration key-value pairs                                                    | No       |
| `jobManagerPodTemplate`  | Pod template for the job manager                                                       | No       |
| `taskManagerPodTemplate` | Pod template for the task manager                                                      | No       |

## Status

| Field                | Description                                                                                     |
|----------------------|-------------------------------------------------------------------------------------------------|
| `conditions`         | List of status conditions for the deployment                                                     |
| `observedGeneration` | The last observed generation of the resource                                                     |
| `deploymentStatus`   | Raw deployment status from the API server                                                        |

## Example

1. Create a ComputeFlinkDeployment with explicit APIServerRef:

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: ComputeFlinkDeployment
metadata:
  name: operator-test-v1
  namespace: default
spec:
  apiServerRef:
    name: test-connection
  workspaceName: test-operator-workspace
  template:
    syncingMode: PATCH
    deployment:
      userMetadata:
        name: operator-test-v1
        namespace: default
        displayName: operator-test-v1
      spec:
        state: RUNNING
        deploymentTargetName: default
        maxJobCreationAttempts: 99
        template:
          metadata:
            annotations:
              flink.queryable-state.enabled: 'false'
              flink.security.ssl.enabled: 'false'
          spec:
            artifact:
              jarUri: function://public/default/flink-operator-test-beam-pulsar-io@1.19-snapshot
              mainArgs: --runner=FlinkRunner --attachedMode=false --checkpointingInterval=60000
              entryClass: org.apache.beam.examples.WordCount
              kind: JAR
              flinkVersion: "1.18.1"
              flinkImageTag: "1.18.1-stream3-scala_2.12-java17"
            flinkConfiguration:
              execution.checkpointing.externalized-checkpoint-retention: RETAIN_ON_CANCELLATION
              execution.checkpointing.interval: 1min
              execution.checkpointing.timeout: 10min
              high-availability.type: kubernetes
              state.backend: filesystem
              taskmanager.memory.managed.fraction: '0.2'
            parallelism: 1
            numberOfTaskManagers: 1
            resources:
              jobmanager:
                cpu: "1"
                memory: 2G
              taskmanager:
                cpu: "1"
                memory: 2G
            logging:
              loggingProfile: default
              log4jLoggers:
                "": DEBUG
                com.company: DEBUG
```

2. Create a ComputeFlinkDeployment with configuration and imagePullSecrets:

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: ComputeFlinkDeployment
metadata:
  name: resource-operator-v4
  namespace: default
spec:
  apiServerRef:
    name: test-connection
  workspaceName: o-nn5f0-vvp
  configuration:
    envs:
    - name: ENV_TEST
      value: "test"
    secrets:
    - name: SECRET_PASSWORD
      valueFrom:
        name: my-secret
        key: password
  imagePullSecrets:
  - name: resource-operator-secret-docker-hub
  template:
    syncingMode: PATCH
    deployment:
      userMetadata:
        name: resource-operator-v4
        namespace: default
        displayName: resource-operator-v4
      spec:
        state: RUNNING
        deploymentTargetName: o-nn5f0
        maxJobCreationAttempts: 99
        template:
          metadata:
            annotations:
              flink.queryable-state.enabled: 'false'
              flink.security.ssl.enabled: 'false'
          spec:
            artifact:
              mainArgs: --runner=FlinkRunner --pulsarCluster=wstest --attachedMode=false
              entryClass: com.example.DataTransformer
              kind: JAR
              flinkVersion: "1.18.1"
              flinkImageTag: "1.18.1-stream3-scala_2.12-java17"
              artifactImage: example/private:latest
            flinkConfiguration:
              classloader.resolve-order: parent-first
              execution.checkpointing.externalized-checkpoint-retention: RETAIN_ON_CANCELLATION
              execution.checkpointing.interval: 1min
              execution.checkpointing.timeout: 10min
              high-availability.type: kubernetes
              state.backend: filesystem
              taskmanager.memory.managed.fraction: '0.2'
            parallelism: 1
            numberOfTaskManagers: 1
            resources:
              jobmanager:
                cpu: "1"
                memory: 2G
              taskmanager:
                cpu: "1"
                memory: 2G
            logging:
              loggingProfile: default
              log4jLoggers:
                "": DEBUG
                com.company: DEBUG
```

3. Apply the YAML file:

```shell
kubectl apply -f deployment.yaml
```

4. Check the deployment status:

```shell
kubectl get computeflinkdeployment operator-test-v1
```

The deployment is ready when the Ready condition is True:

```shell
NAME             READY   AGE
operator-test-v1 True    1m
```

5. Create a ComputeFlinkDeployment using Workspace's APIServerRef:

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: ComputeFlinkDeployment
metadata:
  name: operator-test-v2
  namespace: default
spec:
  workspaceName: test-operator-workspace  # Will use APIServerRef from this workspace
  template:
    syncingMode: PATCH
    deployment:
      userMetadata:
        name: operator-test-v2
        namespace: default
        displayName: operator-test-v2
      spec:
        state: RUNNING
        deploymentTargetName: default
        maxJobCreationAttempts: 99
        template:
          metadata:
            annotations:
              flink.queryable-state.enabled: 'false'
              flink.security.ssl.enabled: 'false'
          spec:
            artifact:
              jarUri: function://public/default/flink-operator-test-beam-pulsar-io@1.19-snapshot
              mainArgs: --runner=FlinkRunner --attachedMode=false --checkpointingInterval=60000
              entryClass: org.apache.beam.examples.WordCount
              kind: JAR
              flinkVersion: "1.18.1"
              flinkImageTag: "1.18.1-stream3-scala_2.12-java17"
            flinkConfiguration:
              execution.checkpointing.externalized-checkpoint-retention: RETAIN_ON_CANCELLATION
              execution.checkpointing.interval: 1min
              execution.checkpointing.timeout: 10min
              high-availability.type: kubernetes
              state.backend: filesystem
              taskmanager.memory.managed.fraction: '0.2'
            parallelism: 1
            numberOfTaskManagers: 1
            resources:
              jobmanager:
                cpu: "1"
                memory: 2G
              taskmanager:
                cpu: "1"
                memory: 2G
            logging:
              loggingProfile: default
              log4jLoggers:
                "": DEBUG
                com.company: DEBUG
```

## Update Deployment

You can update the deployment by modifying the YAML file and reapplying it. Most fields can be updated, including:
- Flink configuration
- Resources
- Parallelism
- Logging settings
- Artifact configuration
- Environment variables and secrets
- Image pull secrets

After applying changes, verify the status to ensure the deployment is updated properly.

## Delete Deployment

To delete a ComputeFlinkDeployment resource:

```shell
kubectl delete computeflinkdeployment operator-test-v1
```

This will stop the Flink job and clean up all associated resources in StreamNative Cloud.