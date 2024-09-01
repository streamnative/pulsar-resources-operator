# PulsarResourceLifeCyclePolicy

## Overview

The `PulsarResourceLifeCyclePolicy` is a configuration option that determines the behavior of Pulsar resources when their corresponding Kubernetes custom resources are deleted. This policy helps manage the lifecycle of Pulsar resources in relation to their Kubernetes representations.

## Available Options

The `PulsarResourceLifeCyclePolicy` can be set to one of two values:

1. `CleanUpAfterDeletion`
2. `KeepAfterDeletion`

### CleanUpAfterDeletion

When set to `CleanUpAfterDeletion`, the Pulsar resource (such as a tenant, namespace, or topic) will be deleted from the Pulsar cluster when its corresponding Kubernetes custom resource is deleted.

Example:

```yaml
apiVersion: pulsar.streamnative.io/v1alpha1
kind: PulsarTenant
metadata:
  name: my-tenant
spec:
  lifecyclePolicy: CleanUpAfterDeletion
  <...>
```

In this example, when the Kubernetes custom resource for the Pulsar tenant is deleted, the corresponding Pulsar tenant will be deleted from the Pulsar cluster.

### KeepAfterDeletion

When set to `KeepAfterDeletion`, the Pulsar resource will not be deleted from the Pulsar cluster when its corresponding Kubernetes custom resource is deleted. The resource will remain in the Pulsar cluster after the Kubernetes custom resource is deleted.

Example:

```yaml
apiVersion: pulsar.streamnative.io/v1alpha1
kind: PulsarNamespace
metadata:
  name: my-namespace
spec:
  lifecyclePolicy: KeepAfterDeletion
  <...>
```

In this example, when the Kubernetes custom resource for the Pulsar namespace is deleted, the corresponding Pulsar namespace will not be deleted from the Pulsar cluster. The namespace will remain in the Pulsar cluster after the Kubernetes custom resource is deleted.

## Default Policy

The default policy for Pulsar resources is `CleanUpAfterDeletion`. This means that the Pulsar resource will be deleted from the Pulsar cluster when its corresponding Kubernetes custom resource is deleted.

## Deleting the Actual Pulsar Resource

When you need to delete the actual Pulsar resource (tenant, namespace, or topic) from the Pulsar cluster, regardless of the `lifecyclePolicy` setting, you can follow these steps:

1. **For resources with `CleanUpAfterDeletion` policy:**
   Simply delete the Kubernetes custom resource, and the corresponding Pulsar resource will be automatically deleted from the Pulsar cluster.

   ```shell
   kubectl delete pulsartenant my-tenant
   ```

2. **For resources with `KeepAfterDeletion` policy:**
   a. First, update the custom resource to change the policy to `CleanUpAfterDeletion`:

   ```yaml
   apiVersion: pulsar.streamnative.io/v1alpha1
   kind: PulsarTenant
   metadata:
     name: my-tenant
   spec:
     lifecyclePolicy: CleanUpAfterDeletion
     # ... other fields ...
   ```

   Apply the updated resource:

   ```shell
   kubectl apply -f updated-tenant.yaml
   ```

   b. Then, delete the Kubernetes custom resource:

   ```shell
   kubectl delete pulsartenant my-tenant
   ```

   This two-step process ensures that the Pulsar resource is deleted from both Kubernetes and the Pulsar cluster.

3. **Manual deletion using Pulsar admin tools:**
   If you need to delete the Pulsar resource directly without involving Kubernetes, you can use Pulsar's admin tools. For more detailed information on using these tools, refer to the [Pulsar Admin CLI documentation](https://pulsar.apache.org/docs/admin-api-overview/) or [pulsarctl documentation](https://github.com/streamnative/pulsarctl). For example:

   ```shell
   # Delete a tenant
   pulsarctl tenants delete my-tenant

   # Delete a namespace
   pulsarctl namespaces delete my-tenant/my-namespace

   # Delete a topic
   pulsarctl topics delete persistent://my-tenant/my-namespace/my-topic
   ```

   Note: Be cautious when using this method, as it may create inconsistencies between Kubernetes and Pulsar if the corresponding Kubernetes resources are not also deleted.

Always ensure you have the necessary permissions and have considered the implications of deleting resources before proceeding with any deletion operation.

## Changing the Policy

You can change the policy of a Pulsar resource by updating the `lifecyclePolicy` field in the corresponding Kubernetes custom resource. However, there are important considerations to keep in mind when changing the policy:

1. **Timing**: The policy change takes effect immediately upon updating the custom resource. However, it only affects future deletion attempts, not any ongoing deletion processes.

2. **From CleanUpAfterDeletion to KeepAfterDeletion**: 
   - This change prevents the Pulsar resource from being deleted when the Kubernetes resource is removed.
   - Ensure you have a plan to manage the retained Pulsar resource outside of Kubernetes.

3. **From KeepAfterDeletion to CleanUpAfterDeletion**:
   - This change means the Pulsar resource will be deleted when the Kubernetes resource is removed.
   - Be cautious, as this could lead to unintended data loss if not managed properly.

4. **Consistency**: After changing the policy, verify that the behavior aligns with your expectations by attempting a delete operation.

5. **Resource Management**: When changing to `KeepAfterDeletion`, consider how you will manage and potentially clean up these resources in the future to avoid clutter in your Pulsar cluster.

6. **Documentation**: It's advisable to document any policy changes, especially in production environments, to maintain clarity on resource management strategies.

Always test policy changes in a non-production environment first to understand their full implications.