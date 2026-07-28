# Overview

Authored by [StreamNative](https://streamnative.io), this Pulsar Resources Operator is a controller that manages the Pulsar resources automatically using the manifest on Kubernetes. Therefore, you can manage the Pulsar resources without the help of `pulsar-admin` or `pulsarctl` CLI tool. It is useful for initializing basic resources when creating a new Pulsar cluster.

The operator manages these resource groups:

- Pulsar connectivity: [PulsarConnection](docs/pulsar_connection.md)
- Pulsar resources: [Tenants](docs/pulsar_tenant.md), [Namespaces](docs/pulsar_namespace.md), [Topics](docs/pulsar_topic.md), [Permissions](docs/pulsar_permission.md), [Packages](docs/pulsar_package.md), [Functions](docs/pulsar_function.md), [Sinks](docs/pulsar_sink.md), [Sources](docs/pulsar_source.md), [Geo-Replication](docs/pulsar_geo_replication.md), and [NS-Isolation-Policy](docs/pulsar_ns_isolation_policy.md)
- StreamNative Cloud connectivity and resources: [StreamNativeCloudConnection](docs/streamnative_cloud_connection.md), [ComputeWorkspace](docs/compute_workspace.md), [ComputeFlinkDeployment](docs/compute_flink_deployment.md), [Secret](docs/secret.md), [ServiceAccount](docs/serviceaccount.md), [ServiceAccountBinding](docs/serviceaccountbinding.md), [APIKey](docs/apikey.md), and [RoleBinding](docs/rolebinding.md)

## Lifecycle Management

The Pulsar Resources Operator provides a flexible approach to managing remote-resource lifecycle through `PulsarResourceLifeCyclePolicy`. This policy determines how supported Pulsar and StreamNative Cloud resources are handled when their Kubernetes custom resources are deleted. For details and the supported-resource list, see [PulsarResourceLifeCyclePolicy](docs/pulsar_resource_lifecycle.md).

There are two available options for the lifecycle policy:

1. `CleanUpAfterDeletion`: The remote resource is deleted when its Kubernetes custom resource is deleted. This is the default policy.

2. `KeepAfterDeletion`: The remote resource remains after its Kubernetes custom resource is deleted.

You can specify the lifecycle policy in the custom resource definition:

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarTenant
metadata:
  name: my-tenant
spec:
  name: my-tenant
  connectionRef:
    name: my-pulsar-connection
  lifecyclePolicy: KeepAfterDeletion
```

# Installation

The Pulsar Resources Operator is an independent controller, it doesn’t need to be installed with the pulsar operator. You can install it when you need the feature. And it is built with the [Operator SDK](https://github.com/operator-framework/operator-sdk), which is part of the [Operator framework](https://github.com/operator-framework/).

You can install the Pulsar Resources Operator using the officially supported `pulsar-resources-operator` Helm [chart](https://github.com/streamnative/charts/tree/master/charts/pulsar-resources-operator). It provides Custom Resource Definitions (CRDs) and Controllers to manage the Pulsar resources.

## Prerequisites

- Install [`kubectl`](https://kubernetes.io/docs/tasks/tools/#kubectl), compatible with your cluster (+/- 1 minor release from your cluster).
- Install [`Helm`](https://helm.sh/docs/intro/install/) v3.
- Prepare a Kubernetes cluster v1.18 or newer, matching the Helm chart's `kubeVersion` constraint.
- Prepare a [Pulsar cluster](https://docs.streamnative.io/operators/pulsar-operator/tutorial/deploy-pulsar) when managing Pulsar resources.
- Prepare StreamNative Cloud service-account credentials and an organization name when managing StreamNative Cloud resources.


## Install Pulsar Resources Operator

To install the Pulsar Resources Operator, follow these steps.
1. Add the StreamNative chart repository.
    
    ```shell
    helm repo add streamnative https://charts.streamnative.io
    helm repo update
    ```

2. Install the operator using the `pulsar-resources-operator` Helm chart. Helm creates the namespace when needed.
    
    ```shell
    helm -n <k8s-namespace> install <release-name> streamnative/pulsar-resources-operator \
      --create-namespace
    ```
3. Verify that the operator is installed successfully.
    
    ```shell
    kubectl -n <k8s-namespace> get pods
    ```

    Expected outputs:

    ```shell
    NAME                                          READY       STATUS           RESTARTS      AGE
    <release-name>-pulsar-resources-operator      1/1         Running          0             2m2s
    ```

## Upgrade Pulsar Resources Operator

Helm does not upgrade CRDs from a chart's `crds/` directory. Pull the target chart, apply its CRDs, then upgrade the release:

```shell
helm repo update
helm pull streamnative/pulsar-resources-operator --version <chart-version> --untar
kubectl apply -f pulsar-resources-operator/crds
helm -n <k8s-namespace> upgrade <release-name> streamnative/pulsar-resources-operator \
  --version <chart-version>
```

See [Helm CRD caveats](https://helm.sh/docs/chart_best_practices/custom_resource_definitions/#some-caveats-and-explanations).

## Uninstall Pulsar Resources Operator

To uninstall the operator, execute the following command.

```shell
helm -n <k8s-namespace> uninstall <release-name>
```

Helm leaves CRDs and existing custom resources in place. Remove CRDs separately only after deleting or preserving all managed resources intentionally.

# Tutorial

This tutorial guides you through creating Pulsar resources. You can create Pulsar resources automatically by applying resource manifest files  to the Kubernetes.

Before creating Pulsar resources, you must create a resource called `PulsarConnection`. The `PulsarConnection` covers the address of the Pulsar cluster and the authentication information. You can use this information to access a Pulsar cluster to create other resources.

In this tutorial, a Kubernetes namespace called `test` is used for examples, which is the namespace that the pulsar cluster installed.

- [PulsarConnection](docs/pulsar_connection.md)
- [PulsarResourceLifeCyclePolicy](docs/pulsar_resource_lifecycle.md)
- [PulsarTenant](docs/pulsar_tenant.md)
- [PulsarNamespace](docs/pulsar_namespace.md)
- [PulsarTopic](docs/pulsar_topic.md)
- [PulsarPermission](docs/pulsar_permission.md)
- [PulsarPackage](docs/pulsar_package.md)
- [PulsarFunction](docs/pulsar_function.md)
- [PulsarSink](docs/pulsar_sink.md)
- [PulsarSource](docs/pulsar_source.md)
- [PulsarGeoReplication](docs/pulsar_geo_replication.md)
- [NS-Isolation-Policy](docs/pulsar_ns_isolation_policy.md)
- [StreamNativeCloudConnection](docs/streamnative_cloud_connection.md)
- [ComputeWorkspace](docs/compute_workspace.md)
- [ComputeFlinkDeployment](docs/compute_flink_deployment.md)
- [StreamNative Cloud Secret](docs/secret.md)
- [StreamNative Cloud APIKey](docs/apikey.md)
- [StreamNative Cloud ServiceAccount](docs/serviceaccount.md)
- [StreamNative Cloud ServiceAccountBinding](docs/serviceaccountbinding.md)
- [StreamNative Cloud RBAC RoleBinding](docs/rolebinding.md)

# Contributing

Contributions are warmly welcomed and greatly appreciated! 
The project follows the typical GitHub pull request model.
Please read the [contribution guidelines](CONTRIBUTING.md) for more details.

Before starting any work, please either comment on an existing issue, or file a new one.

## License

This library is licensed under the terms of the [Apache License 2.0](LICENSE) and may include packages written by third parties which carry their own copyright notices and license terms.

## About StreamNative

Founded in 2019 by the original creators of Apache Pulsar, [StreamNative](https://streamnative.io) is one of the leading contributors to the open-source Apache Pulsar project. We have helped engineering teams worldwide make the move to Pulsar with [StreamNative Cloud](https://streamnative.io/product), a fully managed service to help teams accelerate time-to-production.
