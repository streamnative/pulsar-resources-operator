# Pulsar Resources Operator

![Version: v0.21.0](https://img.shields.io/badge/Version-v0.21.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v0.21.0](https://img.shields.io/badge/AppVersion-v0.21.0-informational?style=flat-square)

## Installing the Chart

To install the chart with the release name `my-release`:

```console
$ helm repo add streamnative https://charts.streamnative.io
$ helm -n <namespace> install my-release streamnative/pulsar-resources-operator
```

## Requirements

Kubernetes: `>= 1.16.0-0`

Pulsar: `>= 2.9.0.x`

## CRD Upgrade

Helm installs CRDs from `crds/` only on `helm install`. A `helm upgrade` does not update CRDs.
To upgrade CRDs, apply them explicitly before or after upgrading the chart.

Apply from the local chart directory:

```console
$ kubectl apply -f charts/pulsar-resources-operator/crds
```

Apply from raw URLs for a specific version (v0.21.0 shown below):

```console
kubectl apply -f https://raw.githubusercontent.com/streamnative/pulsar-resources-operator/refs/tags/pulsar-resources-operator-v0.21.0/charts/pulsar-resources-operator/crds/resource.streamnative.io_pulsarfunctions.yaml
kubectl apply -f https://raw.githubusercontent.com/streamnative/pulsar-resources-operator/refs/tags/pulsar-resources-operator-v0.21.0/charts/pulsar-resources-operator/crds/resource.streamnative.io_serviceaccounts.yaml
kubectl apply -f https://raw.githubusercontent.com/streamnative/pulsar-resources-operator/refs/tags/pulsar-resources-operator-v0.21.0/charts/pulsar-resources-operator/crds/resource.streamnative.io_pulsarpackages.yaml
kubectl apply -f https://raw.githubusercontent.com/streamnative/pulsar-resources-operator/refs/tags/pulsar-resources-operator-v0.21.0/charts/pulsar-resources-operator/crds/resource.streamnative.io_serviceaccountbindings.yaml
kubectl apply -f https://raw.githubusercontent.com/streamnative/pulsar-resources-operator/refs/tags/pulsar-resources-operator-v0.21.0/charts/pulsar-resources-operator/crds/resource.streamnative.io_computeworkspaces.yaml
kubectl apply -f https://raw.githubusercontent.com/streamnative/pulsar-resources-operator/refs/tags/pulsar-resources-operator-v0.21.0/charts/pulsar-resources-operator/crds/resource.streamnative.io_pulsargeoreplications.yaml
kubectl apply -f https://raw.githubusercontent.com/streamnative/pulsar-resources-operator/refs/tags/pulsar-resources-operator-v0.21.0/charts/pulsar-resources-operator/crds/resource.streamnative.io_apikeys.yaml
kubectl apply -f https://raw.githubusercontent.com/streamnative/pulsar-resources-operator/refs/tags/pulsar-resources-operator-v0.21.0/charts/pulsar-resources-operator/crds/resource.streamnative.io_computeflinkdeployments.yaml
kubectl apply -f https://raw.githubusercontent.com/streamnative/pulsar-resources-operator/refs/tags/pulsar-resources-operator-v0.21.0/charts/pulsar-resources-operator/crds/resource.streamnative.io_pulsarconnections.yaml
kubectl apply -f https://raw.githubusercontent.com/streamnative/pulsar-resources-operator/refs/tags/pulsar-resources-operator-v0.21.0/charts/pulsar-resources-operator/crds/resource.streamnative.io_pulsarpermissions.yaml
kubectl apply -f https://raw.githubusercontent.com/streamnative/pulsar-resources-operator/refs/tags/pulsar-resources-operator-v0.21.0/charts/pulsar-resources-operator/crds/resource.streamnative.io_pulsarnamespaces.yaml
kubectl apply -f https://raw.githubusercontent.com/streamnative/pulsar-resources-operator/refs/tags/pulsar-resources-operator-v0.21.0/charts/pulsar-resources-operator/crds/resource.streamnative.io_pulsarsinks.yaml
kubectl apply -f https://raw.githubusercontent.com/streamnative/pulsar-resources-operator/refs/tags/pulsar-resources-operator-v0.21.0/charts/pulsar-resources-operator/crds/resource.streamnative.io_pulsartopics.yaml
kubectl apply -f https://raw.githubusercontent.com/streamnative/pulsar-resources-operator/refs/tags/pulsar-resources-operator-v0.21.0/charts/pulsar-resources-operator/crds/resource.streamnative.io_secrets.yaml
kubectl apply -f https://raw.githubusercontent.com/streamnative/pulsar-resources-operator/refs/tags/pulsar-resources-operator-v0.21.0/charts/pulsar-resources-operator/crds/resource.streamnative.io_rolebindings.yaml
kubectl apply -f https://raw.githubusercontent.com/streamnative/pulsar-resources-operator/refs/tags/pulsar-resources-operator-v0.21.0/charts/pulsar-resources-operator/crds/resource.streamnative.io_pulsarnsisolationpolicies.yaml
kubectl apply -f https://raw.githubusercontent.com/streamnative/pulsar-resources-operator/refs/tags/pulsar-resources-operator-v0.21.0/charts/pulsar-resources-operator/crds/resource.streamnative.io_pulsartenants.yaml
kubectl apply -f https://raw.githubusercontent.com/streamnative/pulsar-resources-operator/refs/tags/pulsar-resources-operator-v0.21.0/charts/pulsar-resources-operator/crds/resource.streamnative.io_pulsarsources.yaml
kubectl apply -f https://raw.githubusercontent.com/streamnative/pulsar-resources-operator/refs/tags/pulsar-resources-operator-v0.21.0/charts/pulsar-resources-operator/crds/resource.streamnative.io_streamnativecloudconnections.yaml
```

## Hardened Configuration

The chart leaves ServiceAccount token automount and security contexts unchanged by default. Set
`serviceAccount.automountServiceAccountToken` to `true` or `false` to configure the chart-created
ServiceAccount explicitly; omit it or use `null` to leave the field unset. When `serviceAccount.create=false`,
configure automount on the externally managed ServiceAccount instead. Changing only the ServiceAccount
setting does not trigger a Deployment rollout; recreate the operator Pods for the change to take effect.

The operator needs Kubernetes API credentials for watches, status updates, and leader election.
Setting automount to `false` without supplying credentials breaks the standard in-cluster configuration.
To disable automatic mounting while retaining API access, save the following as `values-hardened.yaml`.
It explicitly projects a short-lived ServiceAccount token at client-go's default path and provides
writable temporary storage while keeping the container root filesystem read-only:

```yaml
serviceAccount:
  automountServiceAccountToken: false

securityContext:
  readOnlyRootFilesystem: true
  allowPrivilegeEscalation: false
  runAsNonRoot: true
  capabilities:
    drop:
      - ALL

extraVolumes:
  - name: kube-api-access
    projected:
      sources:
        - serviceAccountToken:
            path: token
            expirationSeconds: 3600
        - configMap:
            name: kube-root-ca.crt
            items:
              - key: ca.crt
                path: ca.crt
        - downwardAPI:
            items:
              - path: namespace
                fieldRef:
                  apiVersion: v1
                  fieldPath: metadata.namespace
  - name: tmp
    emptyDir: {}

extraVolumeMounts:
  - name: kube-api-access
    mountPath: /var/run/secrets/kubernetes.io/serviceaccount
    readOnly: true
  - name: tmp
    mountPath: /tmp
```

Apply the override alongside your existing release values:

```console
$ helm -n <namespace> upgrade --install my-release streamnative/pulsar-resources-operator -f values.yaml -f values-hardened.yaml
```

- Merge any existing `extraVolumes` and `extraVolumeMounts` into the override before applying it: Helm
  replaces lists rather than appending to them. Preserve existing authentication and cloud storage mounts.
- The cluster must support ServiceAccount token projection and provide the `kube-root-ca.crt` ConfigMap
  in the release namespace. With no explicit audience, the token uses the API server's default audience.
  The kubelet rotates the token, and client-go refreshes credentials from the mounted file. Do not use a
  `subPath` mount for the token because it would prevent projected updates from reaching the container.
- Keep `/tmp` writable by the container's runtime UID: inline OAuth2 credentials and PulsarPackage
  downloads create temporary files there. Verify these operations, not just operator startup, when
  enabling a read-only root filesystem. Size temporary storage according to package sizes and concurrency.
- This example does not pin a UID or GID, so the image or cluster can supply the appropriate non-root
  identity. Validate the actual image and any additional security requirements in your environment.
- Explicit token projection does not remove API credentials or reduce the ServiceAccount's RBAC
  permissions. Check the rendered ServiceAccount and Deployment, the admitted Pod, and your policy reports.
  A policy forbidding all API tokens still requires an approved exception or another authentication method.
- Before rollout, verify leader election, resource reconciliation, token rotation, and the OAuth2 and
  package operations you use in a test environment. Rendering this example is not a runtime or policy check.

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Add affinity for pod |
| annotations | object | `{}` | Add annotations for the deployment |
| containerName | string | `"manager"` | Name of the operator container |
| features.alwaysUpdatePulsarResource | bool | `false` | Re-apply observed managed Pulsar resources even when their Kubernetes resources are already Ready. Prefer temporary use for upgrade remediation because it increases Pulsar admin API load on reconciliations and resyncs. |
| fullnameOverride | string | `""` | It will override the name of deployment |
| image.manager.registry | string | `"docker.io"` | Specififies the registry of images, especially when user want to use a different image hub |
| image.manager.repository | string | `"streamnative/pulsar-resources-operator"` | The full repo name for image. |
| image.manager.tag | string | `""` | Image tag, it can override the image tag whose default is the chart appVersion. |
| image.pullPolicy | string | `"IfNotPresent"` | Image pull policy for the container. |
| imagePullSecrets | list | `[]` | Specifies image pull secrets for private registry, the format is `- name: gcr` |
| labels | object | `{}` | Add labels for the deployment |
| nameOverride | string | `""` | It will override the value of label `app.kubernetes.io/name` on pod |
| namespace | string | `""` | Specifies namespace for the release, it will override the `-n` parameter when it's not empty |
| nodeSelector | object | `{}` | Add NodeSelector for pod schedule |
| podAnnotations | object | `{}` | Add annotations for the deployment pod |
| podLabels | object | `{}` | Add labels for the deployment pod |
| podSecurityContext | object | `{}` | Add security context for pod |
| replicaCount | int | `1` | The replicas of pod |
| resources | object | `{}` | Add resource limits and requests |
| securityContext | object | `{}` | Add security context for container |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account |
| serviceAccount.automountServiceAccountToken | bool/null | `nil` | Optional token automount setting for the chart-created ServiceAccount. Null preserves Kubernetes' default behavior. When false, explicitly mount Kubernetes API credentials; see the hardened configuration example in the README. |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `""` |  |
| terminationGracePeriodSeconds | int | `10` | The period seconds that pod will be termiated gracefully |
| tolerations | list | `[]` | Add tolerations |
