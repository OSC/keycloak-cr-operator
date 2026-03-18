# keycloak-cr-operator

![Version: 0.0.1](https://img.shields.io/badge/Version-0.0.1-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v0.0.1](https://img.shields.io/badge/AppVersion-v0.0.1-informational?style=flat-square)

A Helm chart to distribute keycloak-cr-operator

## Installing the Chart

```console
helm repo add osc https://osc.github.io/keycloak-cr-operator
helm install keycloak-cr-operator osc/keycloak-cr-operator \
  --namespace keycloak-cr-operator-system \
  --create-namespace \
  --set manager.config.keycloakURL="https://keycloak.example.com" \
  --set manager.config.adminPassword="your-admin-password"
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| nameOverride | string | `""` | String to partially override chart.fullname template (will maintain the release name) |
| fullnameOverride | string | `""` | String to fully override chart.fullname template |
| manager.replicas | int | `1` | Number of manager replicas |
| manager.image.repository | string | `"quay.io/ohiosupercomputercenter/keycloak-cr-operator"` | Manager image repository |
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
| manager.podAnnotations | object | `{"kubectl.kubernetes.io/default-container":"manager"}` | Pod annotations to add to manager pods |
| manager.healthPort | int | `8081` | Health check port |
| manager.env | list | `[]` | Environment variables to add to manager pods |
| manager.imagePullSecrets | list | `[]` | Image pull secrets |
| manager.podSecurityContext | object | unprivileged | Pod-level security settings |
| manager.securityContext | object | unprivileged | Container-level security settings |
| manager.resources.limits.cpu | int | `1` | CPU limit |
| manager.resources.limits.memory | string | `"256Mi"` | Memory limit |
| manager.resources.requests.cpu | string | `"100m"` | CPU request |
| manager.resources.requests.memory | string | `"64Mi"` | Memory request |
| manager.affinity | object | `{}` | Manager pod's affinity |
| manager.nodeSelector | object | `{}` | Manager pod's node selector |
| manager.tolerations | list | `[]` | Manager pod's tolerations |
| rbacHelpers.enable | bool | `false` | Install convenience admin/editor/viewer roles for CRDs |
| crd.enable | bool | `true` | Install CRDs with the chart |
| crd.keep | bool | `true` | Keep CRDs when uninstalling |
| metrics.enable | bool | `true` | Enable to expose /metrics endpoint with RBAC protection |
| metrics.protocol | string | `"https"` | Metrics protocol (http or https) |
| metrics.ports | object | `{"http":8080,"https":8443}` | Metrics server ports.  Only supports http and https keys |
| metrics.ports.http | int | `8080` | HTTP port |
| metrics.ports.https | int | `8443` | HTTPS port |
| metrics.annotations | object | `{}` | Annotations to add to metrics endpoint |
| certManager.enable | bool | `true` | Enable cert-manager integration. Required for webhook certificates and metrics endpoint certificates |
| webhook.enable | bool | `true` | Enable webhook server |
| webhook.port | int | `9443` | Webhook server port |
| webhook.annotations | object | `{}` | Annotations to add to webhook server |
| networkPolicy.enable | bool | `true` | Enable NetworkPolicy resources for this operator |
| networkPolicy.allowMetricsFromPods | bool | `false` | Allow all pods in operator's namespace to access the operator's metrics |
| networkPolicy.prometheusLabels | object | `{"app.kubernetes.io/name":"prometheus"}` | The Prometheus namespace to allow access |
| networkPolicy.apiServerNamespace | string | `"kube-system"` | The API server namespace name |
| networkPolicy.apiServerPodLabels | object | `{"tier":"control-plane"}` | The API server pod labels to allow |
| prometheus.enable | bool | `false` | Enable Prometheus ServiceMonitor. Requires prometheus-operator to be installed in the cluster |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.11.0](https://github.com/norwoodj/helm-docs/releases/v1.11.0)
