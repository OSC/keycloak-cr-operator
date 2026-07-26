# keycloak-cr-operator

![Version: v0.0.12](https://img.shields.io/badge/Version-v0.0.12-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v0.0.12](https://img.shields.io/badge/AppVersion-v0.0.12-informational?style=flat-square)

A Helm chart to distribute keycloak-cr-operator

## Installing the Chart

```console
helm repo add keycloak-cr-operator https://osc.github.io/keycloak-cr-operator
helm install keycloak-cr-operator keycloak-cr-operator/keycloak-cr-operator \
  --namespace keycloak-cr-operator-system \
  --create-namespace \
  --set manager.config.keycloakURL="https://keycloak.example.com" \
  --set manager.config.adminPassword="your-admin-password"
```

## Migration from v0.0.x to v0.1.0

Version 0.1.0 introduces several changes to the `values.yaml` structure. If you're upgrading from v0.0.x, please update your values accordingly:

The following values have been renamed for consistency:

| Old Path | New Path |
|----------|----------|
| `rbacHelpers.enable` | `rbac.helpers.enabled` |
| `crd.enable` | `crd.enabled` |
| `metrics.enable` | `metrics.enabled` |
| `certManager.enable` | `certManager.enabled` |
| `webhook.enable` | `webhook.enabled` |
| `networkPolicy.enable` | `networkPolicy.enabled` |
| `hooks.enable` | `hooks.enabled` |
| `prometheus.enable` | `prometheus.enabled` |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| nameOverride | string | `""` | String to partially override chart.fullname template (will maintain the release name) |
| fullnameOverride | string | `""` | String to fully override chart.fullname template |
| manager.enabled | bool | `true` | Set to false to skip manager installation |
| manager.replicas | int | `1` | Number of manager replicas |
| manager.image.registry | string | `"quay.io"` | Image registry |
| manager.image.repository | string | `"ohiosupercomputercenter/keycloak-cr-operator"` | Manager image repository |
| manager.image.tag | string | The chart's app version | Manager image tag |
| manager.image.pullPolicy | string | `"IfNotPresent"` | Manager image pull policy |
| manager.config.keycloakURL | string | `""` | Keycloak server URL, eg: https://keycloak.example.com. **required** |
| manager.config.adminUsername | string | `"admin"` | Keycloak admin username |
| manager.config.adminPassword | string | `""` | Keycloak admin password. **required** |
| manager.config.adminRealm | string | `"master"` | Keycloak admin realm |
| manager.config.defaultRealm | string | `nil` | Default Keycloak realm for new resources |
| manager.config.allowedRealms | list | `[]` | Realms that can be used for custom resources |
| manager.config.clientIdPrefix | string | `"kubernetes"` | Prefix for generated client IDs |
| manager.config.clientIdRequired | string | `""` | Required ClientID template |
| manager.extraArgs | list | `[]` | Extra arguments to pass to the manager |
| manager.annotations | object | `{}` | Annotations to add to manager Deployment |
| manager.labels | object | `{}` | Custom Deployment labels |
| manager.podAnnotations | object | `{"kubectl.kubernetes.io/default-container":"manager"}` | Pod annotations to add to manager pods |
| manager.podLabels | object | `{}` | Pod labels to add to manager pods |
| manager.healthPort | int | `8081` | Health check port |
| manager.env | list | `[]` | Environment variables to add to manager pods |
| manager.useImagePullSecret | bool | `false` | Use the imagePullSecret resource created by this chart |
| manager.imagePullSecrets | list | `[]` | imagePullSecrets to use for existing secrets |
| manager.podSecurityContext | object | unprivileged | Pod-level security settings |
| manager.securityContext | object | unprivileged | Container-level security settings |
| manager.resources.limits.cpu | int | `1` | CPU limit |
| manager.resources.limits.memory | string | `"256Mi"` | Memory limit |
| manager.resources.requests.cpu | string | `"100m"` | CPU request |
| manager.resources.requests.memory | string | `"64Mi"` | Memory request |
| manager.affinity | object | `{}` | Manager pod's affinity |
| manager.nodeSelector | object | `{}` | Manager pod's node selector |
| manager.tolerations | list | `[]` | Manager pod's tolerations |
| manager.strategy | object | RollingUpdate | Deployment strategy |
| manager.priorityClassName | string | `""` | Priority class name |
| manager.topologySpreadConstraints | list | `[]` | Topology spread constraints |
| manager.terminationGracePeriodSeconds | int | `10` | Termination grace period seconds |
| rbac.namespaced | bool | `false` | RBAC resource scope - false (default): ClusterRole/ClusterRoleBinding (all namespaces) - true: Role/RoleBinding (release namespace only) |
| rbac.helpers | object | `{"enabled":false}` | Helper roles for CRD management (admin/editor/viewer) |
| rbac.helpers.enabled | bool | `false` | Install convenience admin/editor/viewer roles for CRDs |
| serviceAccount.enabled | bool | `true` | Install default ServiceAccount provided |
| serviceAccount.name | string | `""` | Existing ServiceAccount name (only when enabled=false) Note: When enabled=true, respects nameOverride/fullnameOverride |
| serviceAccount.annotations | object | `{}` | Custom ServiceAccount annotations |
| serviceAccount.labels | object | `{}` | Custom ServiceAccount labels |
| crd.enabled | bool | `true` | Install CRDs with the chart |
| crd.keep | bool | `true` | Keep CRDs when uninstalling |
| metrics.enabled | bool | `true` | Enable to expose /metrics endpoint with RBAC protection |
| metrics.protocol | string | `"https"` | Metrics protocol (http or https) |
| metrics.ports | object | `{"http":8080,"https":8443}` | Metrics server ports.  Only supports http and https keys |
| metrics.ports.http | int | `8080` | HTTP port |
| metrics.ports.https | int | `8443` | HTTPS port |
| metrics.annotations | object | `{}` | Annotations to add to metrics endpoint |
| certManager.enabled | bool | `true` | Enable cert-manager integration. Required for webhook certificates and metrics endpoint certificates |
| webhook.enabled | bool | `true` | Enable webhook server |
| webhook.port | int | `9443` | Webhook server port |
| webhook.annotations | object | `{}` | Annotations to add to webhook server |
| prometheus.enabled | bool | `false` | Enable Prometheus ServiceMonitor. Requires prometheus-operator to be installed in the cluster |
| networkPolicy.enabled | bool | `true` | Enable NetworkPolicy resources for this operator |
| networkPolicy.allowMetricsFromPods | bool | `false` | Allow all pods in operator's namespace to access the operator's metrics |
| networkPolicy.prometheusLabels | object | `{"app.kubernetes.io/name":"prometheus"}` | The Prometheus namespace to allow access |
| networkPolicy.apiServerNamespace | string | `"kube-system"` | The API server namespace name |
| networkPolicy.apiServerPodLabels | object | `{"tier":"control-plane"}` | The API server pod labels to allow |
| imagePullSecret.create | bool | `false` | Create the image pull secret |
| imagePullSecret.registry | string | `""` | imagePullSecret registry |
| imagePullSecret.username | string | `""` | imagePullSecret username |
| imagePullSecret.password | string | `""` | imagePullSecret password |
| hooks.enabled | bool | `true` | Enable post-install hooks |
| hooks.image.registry | string | `"docker.io"` | hook image registry |
| hooks.image.repository | string | `"portainer/kubectl-shell"` | hook image repository |
| hooks.image.tag | string | `"2.39.0"` | hook image tag |
| hooks.image.pullPolicy | string | `"IfNotPresent"` | hook image pull policy |
| hooks.useImagePullSecret | bool | `false` | Use the imagePullSecret resource created by this chart |
| hooks.imagePullSecrets | list | `[]` | imagePullSecrets to use for existing secrets |
| hooks.jobLabels | object | `{}` | Job labels to add to the hook job |
| hooks.podLabels | object | `{}` | Pod labels to add to the hook pod |
| hooks.podAnnotations | object | `{}` | Pod annotations to add to the hook pod |
| hooks.podSecurityContext | object | unprivileged | Pod-level security settings |
| hooks.securityContext | object | unprivileged | Security context for the hook pod |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.11.0](https://github.com/norwoodj/helm-docs/releases/v1.11.0)
