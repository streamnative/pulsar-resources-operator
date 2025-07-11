# RoleBinding

The `RoleBinding` resource grants a [`ClusterRole`](https://docs.streamnative.io/cloud/security/access/rbac/manage-rbac-roles) to a set of subjects (users, identity pools, or service accounts). It connects a role's permissions with the entities that receive those permissions and scopes them to specific resources.

This resource provides a Kubernetes-native way to manage [Role Bindings on StreamNative Cloud](https://docs.streamnative.io/cloud/security/access/rbac/manage-rbac-role-bindings).

## Example

This example grants the `org-admin` role to a user and a service account, scoped to a specific Pulsar instance.

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: RoleBinding
metadata:
  name: my-rolebinding
  namespace: default
spec:
  apiServerRef:
    name: my-connection
  clusterRole: org-admin
  users:
  - my-user@example.com
  serviceAccounts:
  - my-service-account
  srnInstance:
  - "my-pulsar-instance"
```

## Specification

| Field | Type | Description | Required |
| --- | --- | --- | --- |
| `spec.apiServerRef` | [LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#localobjectreference-v1-core) | Reference to a StreamNativeCloudConnection in the same namespace. | Yes |
| `spec.clusterRole` | string | The name of the `ClusterRole` to be granted. See [Predefined RBAC Roles](https://docs.streamnative.io/cloud/security/access/rbac/predefined-rbac-roles). | Yes |
| `spec.users` | []string | A list of user emails that will be granted the role. | No |
| `spec.identityPools` | []string | A list of identity pools that will be granted the role. | No |
| `spec.serviceAccounts` | []string | A list of service accounts that will be granted the role. | No |
| `spec.cel` | string | An optional CEL (Common Expression Language) expression for conditional role binding. | No |
| `spec.srnOrganization` | []string | The organization scope for the SRN. | No |
| `spec.srnInstance` | []string | The Pulsar instance scope for the SRN. | No |
| `spec.srnCluster` | []string | The cluster scope for the SRN. | No |
| `spec.srnTenant` | []string | The tenant scope for the SRN. | No |
| `spec.srnNamespace` | []string | The namespace scope for the SRN. | No |
| `spec.srnTopicDomain` | []string | The topic domain scope for the SRN. | No |
| `spec.srnTopicName` | []string | The topic name scope for the SRN. | No |
| `spec.srnSubscription` | []string | The subscription scope for the SRN. | No |
| `spec.srnServiceAccount`| []string | The service account scope for the SRN. | No |
| `spec.srnSecret` | []string | The secret scope for the SRN. | No |


## Status

| Field | Type | Description |
| --- | --- | --- |
| `status.conditions` | []Condition | Represents the latest available observations of the `RoleBinding`'s state. |
| `status.observedGeneration`| int64 | The last generation of the resource that was observed by the controller. |
| `status.failedClusters` | []string | A list of clusters where applying the role binding failed. |
| `status.syncedClusters` | map[string]string | A map of clusters where the role binding has been successfully synced. The key is the cluster name and the value is the sync status. |


## Conditional Role Bindings

While basic role bindings associate a role with a subject, conditional role bindings provide more granular control by scoping permissions based on resource attributes. This operator supports two ways to define these conditions, mirroring the functionality available in StreamNative Cloud.

### Using Resource Names (SRN Fields)

You can scope permissions by specifying one or more `spec.srn*` fields. This is the simplest way to limit a role to specific resources like tenants, namespaces, or topics.

The SRN fields are provided as arrays to allow granting the same role across multiple resources of the same type in a single `RoleBinding`.

For example, to grant the `tenant-admin` role to a user for two specific tenants (`finance` and `marketing`) within an instance:

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: RoleBinding
metadata:
  name: tenant-admin-binding
  namespace: default
spec:
  apiServerRef:
    name: my-connection
  clusterRole: tenant-admin
  users:
  - "jane.doe@example.com"
  # Define the scope of this binding
  srnInstance:
  - "my-cloud-instance"
  srnTenant:
  - "finance"
  - "marketing"
```
The controller will create a separate binding in the cloud for each combination of SRN values provided.

### Using CEL Expressions

For more complex conditions, you can use a [Common Expression Language (CEL)](https://github.com/google/cel-spec) expression in the `spec.cel` field. The expression has access to the `srn` variable, which contains the fields of the StreamNative Resource Name.

This example grants the `topic-producer` role to a service account, but only for topics within the `finance` tenant on a specific cluster.

```yaml
apiVersion: resource.streamnative.io/v1alpha1
kind: RoleBinding
metadata:
  name: producer-finance-binding
spec:
  apiServerRef:
    name: my-connection
  clusterRole: topic-producer
  serviceAccounts:
  - "sa-producer-finance"
  cel: "srn.instance == 'my-cloud-instance' && srn.cluster == 'us-west' && srn.tenant == 'finance'"
```

## Deleting a RoleBinding

When you delete a `RoleBinding` resource from Kubernetes, the controller will automatically remove the corresponding binding from StreamNative Cloud, effectively revoking the granted permissions. 