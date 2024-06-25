# Overview

Authored by [StreamNative](https://streamnative.io), this Pulsar Resources Operator is a controller that manages the Pulsar resources automatically using the manifest on Kubernetes. Therefore, you can manage the Pulsar resources without the help of `pulsar-admin` or `pulsarctl` CLI tool. It is useful for initializing basic resources when creating a new Pulsar cluster.

Currently, the Pulsar Resources Operator provides full lifecycle management for the following Pulsar resources, including creation, update, and deletion. 

- Tenants
- Namespaces
- Topics
- Permissions
- Packages
- Functions
- Sinks
- Sources


# Installation

The Pulsar Resources Operator is an independent controller, it doesn’t need to be installed with the pulsar operator. You can install it when you need the feature. And it is built with the [Operator SDK](https://github.com/operator-framework/operator-sdk), which is part of the [Operator framework](https://github.com/operator-framework/).


You can install the Pulsar Resources Operator using the officially supported `pulsar-resources-operator` Helm [chart](https://github.com/streamnative/charts/tree/master/charts/pulsar-resources-operator). It provides Custom Resource Definitions (CRDs) and Controllers to manage the Pulsar resources.

## Prerequisites

- Install [`kubectl`](https://kubernetes.io/docs/tasks/tools/#kubectl) (v1.16 - v1.24), compatible with your cluster (+/- 1 minor release from your cluster).
- Install [`Helm`](https://helm.sh/docs/intro/install/) (v3.0.2 or higher).
- Prepare a Kubernetes cluster (v1.16 - v1.24).
- Prepare a [Pulsar cluster](https://docs.streamnative.io/operators/pulsar-operator/tutorial/deploy-pulsar)


## Install Pulsar Resources Operator

To install the Pulsar Resources Operator, follow these steps.
1. Add the StreamNative chart repository.
    
    ```shell
    helm repo add streamnative https://charts.streamnative.io
    helm repo update
    ```

2. Create a Kubernetes namespace.
    
    ```shell
    kubectl create namespace <k8s-namespace>
    ```
    >**Note**
    >
    > You can skip this step if you specify a Kubernetes namespace via the `-- create-namespace <k8s-namespace>` option when you install the operator.

3. Install the operator using the `pulsar-resources-operator` Helm chart.
    
    ```shell
    helm -n <k8s-namespace> install <release-name> streamnative/pulsar-resources-operator
    ```
4. Verify that the operator is installed successfully
    
    ```shell
    kubectl -n <k8s-namespace> get pods
    ```

    Expected outputs:

    ```shell
    NAME                                          READY       STATUS           RESTARTS      AGE
    <release-name>-pulsar-resources-operator      1/1         Running          0             2m2s
    ```

## Upgrade Pulsar Resources Operator

To upgrade the operator, execute the following command.

```shell
helm repo update
helm -n <k8s-namespace> upgrade <release-name> streamnative/pulsar-resources-operator
```

>**Note**
>
> Don not forget to apply the latest crd files. Because there is no support for upgrading or deleting CRDs using Helm
> https://helm.sh/docs/chart_best_practices/custom_resource_definitions/#some-caveats-and-explanations
> You can use `helm pull streamnative/pulsar-resources-operator` to download the chart and unpack it, then apply the crds

## Uninstall Pulsar Resources Operator

To uninstall the operator, execute the following command.

```shell
helm -n <k8s-namespace> uninstall <release-name>
```

# Tutorial

This tutorial guides you through creating Pulsar resources. You can create Pulsar resources automatically by applying resource manifest files  to the Kubernetes.

Before creating Pulsar resources, you must create a resource called `PulsarConnection`. The `PulsarConnection` covers the address of the Pulsar cluster and the authentication information. You can use this information to access a Pulsar cluster to create other resources.

In this tutorial, a Kubernetes namespace called `test` is used for examples, which is the namespace that the pulsar cluster installed.

- [PulsarConnection](docs/pulsar_connection.md)
- [PulsarTenant](docs/pulsar_tenant.md)
- [PulsarNamespace](docs/pulsar_namespace.md)
- [PulsarTopic](docs/pulsar_topic.md)
- [PulsarPermission](docs/pulsar_permission.md)
- [PulsarPackage](docs/pulsar_package.md)
- [PulsarFunction](docs/pulsar_function.md)
- [PulsarSink](docs/pulsar_sink.md)
- [PulsarSource](docs/pulsar_source.md)

## Contributing

Contributions are warmly welcomed and greatly appreciated! 
The project follows the typical GitHub pull request model.
Please read the [contribution guidelines](CONTRIBUTING.md) for more details.

Before starting any work, please either comment on an existing issue, or file a new one.

## License

This library is licensed under the terms of the [Apache License 2.0](LICENSE) and may include packages written by third parties which carry their own copyright notices and license terms.

## About StreamNative

Founded in 2019 by the original creators of Apache Pulsar, [StreamNative](https://streamnative.io) is one of the leading contributors to the open-source Apache Pulsar project. We have helped engineering teams worldwide make the move to Pulsar with [StreamNative Cloud](https://streamnative.io/product), a fully managed service to help teams accelerate time-to-production.

