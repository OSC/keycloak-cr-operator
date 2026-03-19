[![CI Status](https://github.com/OSC/keycloak-cr-operator/actions/workflows/test-e2e.yml/badge.svg?branch=main)](https://github.com/OSC/keycloak-cr-operator/actions)
[![GitHub release](https://img.shields.io/github/v/release/OSC/keycloak-cr-operator?include_prereleases&sort=semver)](https://github.com/OSC/keycloak-cr-operator/releases/latest)
![GitHub All Releases](https://img.shields.io/github/downloads/OSC/keycloak-cr-operator/total)

# keycloak-cr-operator
Keycloak Custom Resource Operator for Kubernetes

## Table of Contents
- [Description](#description)
- [Install](#install)
- [Usage](#usage)
  - [KeycloakClient](#keycloakclient)
- [Development](#development)
- [Working with Kubebuilder](#working-with-kubebuilder)
- [License](#license)

## Description
The keycloak-cr-operator is a Kubernetes operator that manages Keycloak resources based on Custom Resources defined in Kubernetes. The following types of resources can be managed:

* Keycloak Clients using [KeycloakClient](#keycloakclient)

The keycloak-cr-operator is designed to work with existing Keycloak deployments that can be deployed outside Kubernetes or within Kubernetes.

## Install
The primary method to install the keycloak-cr-operator is with Helm.

### Prerequisites
- Helm 3.x
- Kubernetes cluster
- cert-manager (required by default)

### Installation Steps
1. Add the OSC Helm repository:
```bash
helm repo add keycloak-cr-operator https://osc.github.io/keycloak-cr-operator
```

2. Install the operator with required configuration:
```bash
helm install keycloak-cr-operator keycloak-cr-operator/keycloak-cr-operator \
  --namespace keycloak-cr-operator-system \
  --create-namespace \
  --set manager.config.keycloakURL="https://keycloak.example.com" \
  --set manager.config.adminPassword="your-admin-password"
```

### Required Parameters
When installing with Helm, the following parameters must be set:
- `manager.config.keycloakURL`: The URL of your Keycloak server
- `manager.config.adminPassword`: The admin password for Keycloak

### Optional Configuration
The operator can be configured with additional parameters:
- `manager.config.defaultRealm`: The default Keycloak realm (defaults to "master")
- `manager.config.clientIdPrefix`: Prefix for generated client IDs (defaults to "kubernetes")
- `manager.config.adminUsername`: Admin username (defaults to "admin")
- `manager.config.adminRealm`: Admin realm (defaults to "master")

### Cert-manager Dependency
The operator requires cert-manager for metric and webhook certificate management. Cert-manager is enabled by default. If you're not using cert-manager, you can disable it:
```bash
helm install keycloak-cr-operator osc/keycloak-cr-operator \
  --namespace keycloak-cr-operator-system \
  --create-namespace \
  --set manager.config.keycloakURL="https://keycloak.example.com" \
  --set manager.config.adminPassword="your-admin-password" \
  --set certManager.enable=false \
  --set metrics.protocol=http
```

## Usage

### KeycloakClient

See [KeycloakClient docs](./docs/keycloakclient.md)

## Development

**Requires**

* Kind
* kubectl
* Helm

The following outlines the steps to setup a development environment:

```
make setup-test-e2e
make install-cert-manager
make install-keycloak

make docker-build IMG=quay.io/ohiosupercomputercenter/keycloak-cr-operator:latest
kind load docker-image quay.io/ohiosupercomputercenter/keycloak-cr-operator:latest --name keycloak-cr-operator-test-e2e

make helm-deploy IMG=quay.io/ohiosupercomputercenter/keycloak-cr-operator:latest HELM_EXTRA_ARGS="-f charts/keycloak-cr-operator/ci/test-values.yaml --cleanup-on-fail=false"

kubectl apply -f config/samples/keycloak_v1alpha1_keycloakclient.yaml

kubectl logs -n keycloak-cr-operator -l app.kubernetes.io/name=keycloak-cr-operator
```

## Working with Kubebuilder

Refer to [Kubebuilder Usage](./docs/kubebuilder_usage.md) for additional information about interacting with this project via Kubebuilder.

## License

Copyright 2026 Ohio Supercomputer Center.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

