# pulsar-resources-operator

Pulsar Resources Operator Helm chart for Pulsar Resources Management on Kubernetes

![Version: 0.19.0](https://img.shields.io/badge/Version-0.19.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v0.19.0](https://img.shields.io/badge/AppVersion-v0.19.0-informational?style=flat-square)

## Installing the Chart

To install the chart with the release name `my-release`:

```console
$ helm repo add streamnative https://charts.streamnative.io
$ helm -n <namespace> install my-release streamnative/pulsar-resources-operator --create-namespace
```

## Requirements

Kubernetes: `>= 1.18.0-0`

Pulsar: `>= 2.9.0.x`

## CRD Upgrade

Helm installs files from `crds/` during `helm install`, but does not update them during `helm upgrade`. Pull the target chart and apply its CRDs explicitly:

```console
$ helm pull streamnative/pulsar-resources-operator --version <chart-version> --untar
$ kubectl apply -f pulsar-resources-operator/crds
$ helm -n <namespace> upgrade my-release streamnative/pulsar-resources-operator --version <chart-version>
```

Uninstalling the release does not remove CRDs or existing custom resources.

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Add affinity for pod |
| annotations | object | `{}` | Add annotations for the deployment |
| cloudStorage | object | `{"azure":{"accountName":"","credentials":{"accountKey":"","create":false,"sasToken":"","secretName":"azure-credentials","useAccountKey":true},"enabled":false},"gcs":{"enabled":false,"serviceAccount":{"key":{"create":false,"json":"","mountPath":"/var/secrets/google","secretName":"gcs-credentials"},"name":"","useWorkloadIdentity":false}},"s3":{"credentials":{"accessKeyId":"","create":false,"secretAccessKey":"","secretName":"aws-credentials"},"enabled":false,"region":""}}` | Cloud storage configuration used when downloading files for PulsarPackage resources. |
| cloudStorage.azure | object | `{"accountName":"","credentials":{"accountKey":"","create":false,"sasToken":"","secretName":"azure-credentials","useAccountKey":true},"enabled":false}` | Azure Blob Storage configuration |
| cloudStorage.azure.accountName | string | `""` | Azure storage account name |
| cloudStorage.azure.credentials | object | `{"accountKey":"","create":false,"sasToken":"","secretName":"azure-credentials","useAccountKey":true}` | Azure credentials configuration |
| cloudStorage.azure.credentials.accountKey | string | `""` | Storage account key (only used if create is true) |
| cloudStorage.azure.credentials.create | bool | `false` | Create a new secret for Azure credentials |
| cloudStorage.azure.credentials.sasToken | string | `""` | SAS token (only used if create is true) |
| cloudStorage.azure.credentials.secretName | string | `"azure-credentials"` | Existing secret name |
| cloudStorage.azure.credentials.useAccountKey | bool | `true` | Use account key for authentication (if false, will use SAS token) |
| cloudStorage.azure.enabled | bool | `false` | Enable Azure Blob Storage support |
| cloudStorage.gcs | object | `{"enabled":false,"serviceAccount":{"key":{"create":false,"json":"","mountPath":"/var/secrets/google","secretName":"gcs-credentials"},"name":"","useWorkloadIdentity":false}}` | Google Cloud Storage configuration |
| cloudStorage.gcs.enabled | bool | `false` | Enable Google Cloud Storage support |
| cloudStorage.gcs.serviceAccount | object | `{"key":{"create":false,"json":"","mountPath":"/var/secrets/google","secretName":"gcs-credentials"},"name":"","useWorkloadIdentity":false}` | Service account configuration |
| cloudStorage.gcs.serviceAccount.key | object | `{"create":false,"json":"","mountPath":"/var/secrets/google","secretName":"gcs-credentials"}` | Service account key configuration (only used if useWorkloadIdentity is false) |
| cloudStorage.gcs.serviceAccount.key.create | bool | `false` | Create a new secret for service account key |
| cloudStorage.gcs.serviceAccount.key.json | string | `""` | Service account key JSON content (only used if create is true) |
| cloudStorage.gcs.serviceAccount.key.mountPath | string | `"/var/secrets/google"` | Mount path of the service account key file |
| cloudStorage.gcs.serviceAccount.key.secretName | string | `"gcs-credentials"` | Existing secret name containing the service account key |
| cloudStorage.gcs.serviceAccount.name | string | `""` | GCP service account email for Workload Identity Format: GSA_NAME@PROJECT_ID.iam.gserviceaccount.com |
| cloudStorage.gcs.serviceAccount.useWorkloadIdentity | bool | `false` | Use GCP Workload Identity (recommended for GKE) |
| cloudStorage.s3 | object | `{"credentials":{"accessKeyId":"","create":false,"secretAccessKey":"","secretName":"aws-credentials"},"enabled":false,"region":""}` | AWS S3 configuration |
| cloudStorage.s3.credentials | object | `{"accessKeyId":"","create":false,"secretAccessKey":"","secretName":"aws-credentials"}` | AWS credentials secret configuration |
| cloudStorage.s3.credentials.accessKeyId | string | `""` | AWS access key ID (only used if create is true) |
| cloudStorage.s3.credentials.create | bool | `false` | Create a new secret for AWS credentials |
| cloudStorage.s3.credentials.secretAccessKey | string | `""` | AWS secret access key (only used if create is true) |
| cloudStorage.s3.credentials.secretName | string | `"aws-credentials"` | Existing secret name |
| cloudStorage.s3.enabled | bool | `false` | Enable AWS S3 support |
| cloudStorage.s3.region | string | `""` | AWS region |
| containerName | string | `"manager"` | Name of the operator container |
| extraVolumeMounts | list | `[]` | Additional volume mounts for the operator container. Paths are available to file-based PulsarConnection authentication. |
| extraVolumes | list | `[]` | Additional pod volumes for the operator deployment. |
| features.alwaysUpdatePulsarResource | bool | `false` | Re-apply observed managed Pulsar resources even when their Kubernetes resources are already Ready. Prefer temporary use for upgrade remediation because it increases Pulsar admin API load on reconciliations and resyncs. |
| features.maxConcurrentReconciles | int | `1` | Maximum concurrent reconciles for the PulsarConnection and RoleBinding controllers. Values of 0 or 1 leave the flags unset and use the binary default. |
| features.resyncPeriod | int | `10` | Base informer resync period in hours. |
| features.retryCount | int | `5` | Number of retries used by the PulsarConnection-managed resource reconciler. |
| fullnameOverride | string | `""` | It will override the name of deployment |
| image.manager.registry | string | `"docker.io"` | Container image registry. |
| image.manager.repository | string | `"streamnative/pulsar-resources-operator"` | Container image repository. |
| image.manager.tag | string | `""` | Container image tag. Defaults to chart appVersion when empty. |
| image.pullPolicy | string | `"IfNotPresent"` | Image pull policy for the container. |
| imagePullSecrets | list | `[]` | Image pull secrets for private registries, for example `- name: gcr`. |
| labels | object | `{}` | Add labels for the deployment |
| nameOverride | string | `""` | It will override the value of label `app.kubernetes.io/name` on pod |
| namespace | string | `""` | Namespace for chart resources. When empty, use Helm release namespace. |
| nodeSelector | object | `{}` | Add NodeSelector for pod schedule |
| podAnnotations | object | `{}` | Add annotations for the deployment pod |
| podLabels | object | `{}` | Add labels for the deployment pod |
| podSecurityContext | object | `{}` | Add security context for pod |
| replicaCount | int | `1` | Number of operator replicas. |
| resources | object | `{}` | Add resource limits and requests |
| securityContext | object | `{}` | Add security context for container |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `""` | Name of the service account to use. When empty and create=true, the chart generates a name. |
| terminationGracePeriodSeconds | int | `10` | Graceful termination period in seconds. |
| tolerations | list | `[]` | Add tolerations |
