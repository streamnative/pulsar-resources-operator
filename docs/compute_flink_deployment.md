# ComputeFlinkDeployment

## Overview

The `ComputeFlinkDeployment` resource defines a Flink deployment in StreamNative Cloud. The current client conversion implements Ververica Platform (VVP) templates. `communityTemplate` remains in the CRD but is not sent to StreamNative Cloud.

## Specifications

| Field                | Description                                                                                | Required |
|----------------------|--------------------------------------------------------------------------------------------|----------|
| `apiServerRef`       | Reference to the StreamNativeCloudConnection resource for API server access. If not specified, the APIServerRef from the referenced ComputeWorkspace will be used. | No       |
| `lifecyclePolicy`    | Whether to delete the remote Flink deployment or keep it when the Kubernetes resource is deleted. Defaults to cleanup when omitted. | No |
| `workspaceName`      | Name of the ComputeWorkspace where the Flink deployment will run                           | Yes      |
| `labels`             | Present in the CRD but not propagated by the current client. Use `template.deployment.userMetadata.labels` instead. | No |
| `annotations`        | Present in the CRD but not propagated by the current client. Use `template.deployment.userMetadata.annotations` instead. | No |
| `template`           | VVP deployment template configuration. This is the only template type currently propagated. | Conditional |
| `communityTemplate`  | Present in the CRD but ignored by the current create/update conversion. | No |
| `defaultPulsarCluster`| Default Pulsar cluster to use for the deployment                                          | No       |
| `configuration`      | Environment variables and Secret references. Propagated during remote creation; current update conversion leaves existing remote configuration unchanged. | No |
| `imagePullSecrets`   | Image pull secrets injected into VVP JobManager and TaskManager pod templates. | No |

Use `template` for managed deployments. A resource containing only `communityTemplate` is accepted by Kubernetes but reaches the remote API without a deployment template.

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
| `valueFrom` | Secret key selector sent to the remote deployment. Optional in the CRD, but needed for a useful secret-backed value. | No |

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
| `jobFailureExpirationTime`     | Expiration setting for failed jobs                                                      | No       |
| `state`                        | State of the deployment (RUNNING, SUSPENDED, CANCELLED)                                 | No       |
| `maxJobCreationAttempts`       | Maximum number of job creation attempts (minimum: 1)                                   | No       |
| `maxSavepointCreationAttempts` | Maximum number of savepoint creation attempts (minimum: 1)                             | No       |
| `restoreStrategy`              | Restore strategy containing `kind` and `allowNonRestoredState`                          | No       |
| `sessionClusterName`           | Session cluster used by the deployment                                                  | No       |
| `template`                     | Deployment template configuration                                                       | Yes      |

##### Template Spec Fields

| Field                | Description                                                                            | Required |
|----------------------|----------------------------------------------------------------------------------------|----------|
| `artifact`           | Deployment artifact configuration                                                       | Yes      |
| `flinkConfiguration` | Flink configuration key-value pairs                                                    | No       |
| `kubernetes`         | VVP Kubernetes settings. The current local type propagates `labels`; top-level `imagePullSecrets` injects pod template image-pull secrets. | No |
| `latestCheckpointFetchInterval` | Checkpoint status fetch interval                                                       | No       |
| `parallelism`        | Parallelism of the Flink job                                                          | No       |
| `numberOfTaskManagers`| Number of task managers                                                               | No       |
| `resources`          | Resource requirements for jobmanager and taskmanager                                   | No       |
| `logging`            | Logging configuration                                                                  | No       |

##### Artifact Configuration

| Field                    | Description                                                                            | Required |
|--------------------------|----------------------------------------------------------------------------------------|----------|
| `kind`                   | Type of artifact (for example `JAR`, `PYTHON`, or `sqlscript`). The current checked-in CRD does not mark it required, but set it for a usable remote deployment. | No* |
| `jarUri`                 | URI of the JAR file                                                                    | No*      |
| `pythonArtifactUri`      | URI of the Python artifact                                                             | No*      |
| `sqlScript`              | SQL script content                                                                      | No*      |
| `additionalDependencies` | Additional artifact dependencies                                                        | No       |
| `flinkVersion`           | Flink version to use                                                                   | No       |
| `flinkImageRegistry`     | Flink image registry                                                                    | No       |
| `flinkImageRepository`   | Flink image repository                                                                  | No       |
| `flinkImageTag`          | Flink image tag to use                                                                 | No       |
| `mainArgs`               | Arguments for the main class/method                                                     | No       |
| `entryClass`             | Entry class for JAR artifacts                                                          | No       |
| `uri`                    | Generic artifact URI                                                                    | No       |
| `artifactImage`          | Container image containing the artifact                                                 | No       |

*The current CRD does not enforce the artifact kind/URI combination. Supply `kind` and the matching artifact field expected by StreamNative Cloud, such as `jarUri`, `pythonArtifactUri`, or `sqlScript`.

`additionalPythonArchives`, `additionalPythonLibraries`, `artifactKind`, and `entryModule` exist in the CRD but are not copied by the current converter.

### Community Deployment Template

`communityTemplate` is defined by the CRD, but `pkg/streamnativecloud/flinkdeployment_client.go` currently copies only `template`. Do not use `communityTemplate` until client conversion support is implemented.

## Status

| Field                | Description                                                                                     |
|----------------------|-------------------------------------------------------------------------------------------------|
| `conditions`         | List of status conditions for the deployment                                                     |
| `observedGeneration` | Reserved field; the current controller records generation on the `Ready` condition but does not populate this top-level status field. |
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
- VVP template Flink configuration
- Resources
- Parallelism
- Logging settings
- Artifact configuration
- Image pull secrets

The current update client replaces the VVP template, workspace name, and default Pulsar cluster. It does not copy top-level `configuration`, `labels`, or `annotations` during update. Environment variables and Secret references supplied at creation therefore remain unchanged until update support is added or the remote deployment is recreated.

After applying changes, verify the `Ready` condition and `status.deploymentStatus` to ensure the remote deployment accepted the update.

## Delete Deployment

To delete a ComputeFlinkDeployment resource:

```shell
kubectl delete computeflinkdeployment operator-test-v1
```

This will stop the Flink job and clean up all associated resources in StreamNative Cloud.

Set `spec.lifecyclePolicy: KeepAfterDeletion` to remove only the Kubernetes custom resource and retain the remote deployment.
