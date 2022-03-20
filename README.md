# Overview

The Pulsar Resources Operator is a controller that manages the Pulsar resources automatically using the manifest on Kubernetes. Therefore, you can manage the Pulsar resources without the help of `pulsar-admin` or `pulsarctl` CLI tool. It is useful for initializing basic resources when creating a new Pulsar cluster.

Currently, the Pulsar Resources Operator provides a full management life-cycle for the following Pulsar resources, including creation, update, deletion. 

- Tenants
- Namespaces
- Topics
- Permissions


# Installation

The Pulsar Resources Operator is an independent controller, it doesn’t need to be installed with the pulsar operator, you can install it when you need the feature. And it is built with the [Operator SDK](https://github.com/operator-framework/operator-sdk), which is part of the [Operator framework](https://github.com/operator-framework/).


You can install the Pulsar Resources Operator using the officially supported `pulsar-resources-operator` Helm [chart](https://github.com/streamnative/charts/tree/master/charts/pulsar-resources-operator). It provides Customer Resource Definitions (CRDs) and Controllers to manage the Pulsar resources.

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
    NAME                                              READY   STATUS    RESTARTS   AGE
    <release-name>-pulsar-resources-operator       1/1        Running            0           2m2s
    ```

## Upgrade Pulsar Resources Operator

To upgrade the operator, execute the following command.

```shell
helm repo update
helm -n <k8s-namespace> upgrade <release-name> streamnative/pulsar-resources-operator
```

## Uninstall Pulsar Resources Operator

To uninstall the operator, execute the following command.

```shell
helm -n <k8s-namespace> uninstall <release-name>
```

# Tutorial

This tutorial guides you through creating Pulsar resources. By applying resource manifest files  to the Kubernetes, you can create Pulsar resources automatically.

Before creating Pulsar resources, you need to create a resource called `PularConnection`. The `PulsarConnection` covers the address of the Pulsar cluster and the authentication information. You can use this information  to access a Pulsar cluster to create other resources.

In this tutorial, a Kubernetes namespace called `test` is used for examples, which is the namespace that the pulsar cluster installed.

- [PulsarConnection](docs/pulsar_connection.md)
- [PulsarTenant](docs/pulsar_tenant.md)
- [PulsarNamespace](docs/pulsar_namespace.md)
- [PulsarTopic](docs/pulsar_topic.md)
- [PulsarPermission](docs/pulsar_permission.md)