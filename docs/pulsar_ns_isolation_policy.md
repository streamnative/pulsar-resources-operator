# PulsarNSIsolationPolicy

## Overview

The `PulsarNSIsolationPolicy` resource defines a ns-isolation-policy in a Pulsar cluster. It allows you to configure namespace isolation policies to limit the set of brokers that can be used for assignment.

## Specifications

| Field                      | Description                                                                                          | Required |
|----------------------------|------------------------------------------------------------------------------------------------------|----------|
| `name`                     | The name of the policy.                                                                              | Yes      |
| `connectionRef`            | Reference to the PulsarConnection resource used to connect to the Pulsar cluster for this namespace. | Yes      |
| `cluster`                  | The name of the cluster.                                                                             | Yes      |
| `namespaces`               | A list of fully qualified namespace name in the format "tenant/namespace".                           | Yes      |
| `primary`                  | A list of primary-broker-regex.                                                                      | Yes      |
| `secondary`                | A list of secondary-broker-regex.                                                                    | No       |
| `autoFailoverPolicyType`   | Auto failover policy type name, only support `min_available` now.                                    | Yes      |
| `autoFailoverPolicyParams` | A map of auto failover policy parameters.                                                            | Yes      |


## Create A Pulsar ns-isolation-policy

1. Define a isolation policy named `test-policy` by using the YAML file and save the YAML file `policy.yaml`.
```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: PulsarNSIsolationPolicy
metadata:
  name: test-pulsar-ns-isolation-policy
  namespace: test
spec:
  name: test-policy
  cluster: standalone
  connectionRef:
    name: test-pulsar-connection
  namespaces:
    - test-tenant/test-ns
  primary:
    - test-pulsar-broker-0.*
  secondary:
    - test-pulsar-broker-1.*
  autoFailoverPolicyType: min_available
  autoFailoverPolicyParams:
    min_limit: "1"
    usage_threshold: "80"
```

2. Apply the YAML file to create the ns-isolation-policy.

```shell
kubectl apply -f policy.yaml
```

3. Check the resource status. When column Ready is true, it indicates the resource is created successfully in the pulsar cluster

```shell
kubectl -n test get pulsarnsisolationpolicy.resource.streamnative.io
```

```shell
NAME                                        RESOURCE_NAME        GENERATION   OBSERVED_GENERATION   READY
test-pulsar-ns-isolation-policy             test-policy          1            1                     True
```

## Update A Pulsar ns-isolation-policy

You can update the ns-isolation-policy by editing the `policy.yaml` file and then applying it again using `kubectl apply -f policy.yaml`. This allows you to modify various settings of the Pulsar ns-isolation-policy.

After applying changes, you can check the status of the update using:

```shell
kubectl -n test get pulsarnsisolationpolicy.resource.streamnative.io
```
The `OBSERVED_GENERATION` should increment, and `READY` should become `True` when the update is complete.

## Delete A Pulsar ns-isolation-policy

To delete a PulsarNSIsolationPolicy resource, use the following kubectl command:

```shell
kubectl -n test delete pulsarnsisolationpolicy.resource.streamnative.io test-pulsar-ns-isolation-policy
```
