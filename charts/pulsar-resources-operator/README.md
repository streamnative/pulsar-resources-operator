# pulsar-resources-operator

Pulsar Resources Operator Helm chart for Pulsar Resources Management on Kubernetes

![Version: v0.5.2](https://img.shields.io/badge/Version-v0.5.2-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v0.5.2](https://img.shields.io/badge/AppVersion-v0.5.2-informational?style=flat-square)

## Installing the Chart

To install the chart with the release name `my-release`:

```console
$ helm repo add streamnative https://charts.streamnative.io
$ helm -n <namespace> install my-release streamnative/pulsar-resources-operator
```

## Requirements

Kubernetes: `>= 1.16.0-0`

Pulsar: `>= 2.9.0.x`

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Add affinity for pod |
| features.alwaysUpdatePulsarResource | bool | `false` |  |
| fullnameOverride | string | `""` | It will override the name of deployment |
| image.kubeRbacProxy.registry | string | `"gcr.io"` | Specififies the registry of images, especially when user want to use a different image hub |
| image.kubeRbacProxy.repository | string | `"kubebuilder/kube-rbac-proxy"` | The full repo name for image. |
| image.kubeRbacProxy.tag | string | `"v0.14.4"` | Image tag, it can override the image tag whose default is the chart appVersion. |
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
| serviceAccount.annotations | object | `{}` |  |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `""` |  |
| terminationGracePeriodSeconds | int | `10` | The period seconds that pod will be termiated gracefully |
| tolerations | list | `[]` | Add tolerations |