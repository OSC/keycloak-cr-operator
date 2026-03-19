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
    - [CRD Overview](#crd-overview)
    - [Default Values Applied by Webhook](#default-values-applied-by-webhook)
    - [Using an Existing Secret](#using-an-existing-secret)
    - [Creating a Secret Automatically](#creating-a-secret-automatically)
    - [Secret Creation](#secret-creation)
    - [ConfigMap Creation](#configmap-creation)
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

#### CRD Overview
The operator manages Keycloak clients through the `KeycloakClient` Custom Resource Definition (CRD). This CRD supports various Keycloak client properties and configurations.

For detailed information about all available fields and their usage, please refer to the [KeycloakClient CRD documentation](docs/crds.md).

#### Default Values Applied by Webhook
When creating or updating a KeycloakClient, the webhook automatically applies the following defaults:

- **ClientID**: If not specified, it will be auto-generated using the pattern:
  - If `ClientIDPrefix` is set in the operator configuration: `prefix-namespace-name`
  - If no prefix: `namespace-name`

- **Realm**: If not specified, it will default to the operator's configured `DefaultRealm`

- **ClientSecretRef**: When `clientAuthenticatorType` is "client-secret" and `publicClient` is false:
  - If `clientSecretRef` is not set, it will be auto-created with:
    - `name`: `name-secret` (where `name` is the KeycloakClient resource name)
    - `key`: `client-secret`
    - `create`: `true` (indicating the secret should be auto-created)

- **ConfigMapName**: If not specified, it will default to `name-config` (where `name` is the KeycloakClient resource name)

#### Using an Existing Secret
To reference an existing Kubernetes Secret for the client secret:

```yaml
apiVersion: keycloak.osc.edu/v1alpha1
kind: KeycloakClient
metadata:
  name: example-client
  namespace: default
spec:
  # Define the Keycloak realm
  realm: "my-realm"

  # Define the client ID (optional - will default to namespace-name)
  clientID: "my-client-id"

  # Reference an existing secret
  clientSecretRef:
    name: "existing-secret"
    key: "client-secret"
    # Set to false if you want to use an existing secret without creating it
    create: false

  # Other client properties
  name: "Example Client"
  description: "An example Keycloak client"
  enabled: true
  publicClient: false
  redirectUris:
    - "https://example.com/callback"
  webOrigins:
    - "https://example.com"
```

When referencing an existing secret, the operator uses the label `keycloak.osc.edu/keycloakclient` to identify which KeycloakClient owns the secret. This allows the operator to properly manage the relationship between Keycloak clients and their secrets.

Example of a secret with the required label:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: existing-secret
  namespace: default
  labels:
    keycloak.osc.edu/keycloakclient: "example-client"
type: Opaque
data:
  client-secret: <base64-encoded-secret>
```

#### Creating a Secret Automatically
To have the operator automatically create a secret with a generated client secret:

```yaml
apiVersion: keycloak.osc.edu/v1alpha1
kind: KeycloakClient
metadata:
  name: example-client-auto-secret
  namespace: default
spec:
  # Define the Keycloak realm
  realm: "my-realm"

  # Define the client ID (optional - will default to namespace-name)
  clientID: "my-client-id"

  # Configure secret creation
  clientSecretRef:
    # Name will default to name-secret if not specified
    name: "my-client-secret"
    key: "client-secret"
    # Set to true (default) to create the secret automatically
    create: true

  # Other client properties
  name: "Example Client with Auto Secret"
  description: "An example Keycloak client with auto-created secret"
  enabled: true
  publicClient: false
  redirectUris:
    - "https://example.com/callback"
  webOrigins:
    - "https://example.com"
```

#### Secret Creation
The operator creates Kubernetes Secrets containing client credentials with the following structure:
- `client-secret`: The raw client secret value
- `CLIENT_SECRET`: The client secret value in uppercase snake_case format
- `cookie-secret`: A randomly generated secret used for OAuth2 Proxy cookie encryption
- `COOKIE_SECRET`: The cookie secret value in uppercase snake_case format

The `cookie-secret` is specifically intended to be used with OAuth2 Proxy for securing cookies. It is automatically generated upon Secret creation and not modified on updates.  If the cookie secret keys are removed from the Secret, a new random cookie secret will be added back to the Secret.

Example Secret structure:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: example-client-secret
  namespace: default
type: Opaque
data:
  client-secret: <base64-encoded-secret>
  CLIENT_SECRET: <base64-encoded-secret>
  cookie-secret: <base64-encoded-cookie-secret>
  COOKIE_SECRET: <base64-encoded-cookie-secret>
```

#### ConfigMap Creation
The operator creates Kubernetes ConfigMaps with Keycloak client configuration:
- `client-id`: The client ID
- `CLIENT_ID`: The client ID in uppercase snake_case format
- `keycloak-url`: The Keycloak server URL
- `KEYCLOAK_URL`: The Keycloak server URL in uppercase snake_case format
- `issuer-url`: The issuer URL for OpenID Connect
- `ISSUER_URL`: The issuer URL in uppercase snake_case format

Example ConfigMap structure:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: example-client-config
  namespace: default
data:
  client-id: "example-client"
  CLIENT_ID: "example-client"
  keycloak-url: "https://keycloak.example.com"
  KEYCLOAK_URL: "https://keycloak.example.com"
  issuer-url: "https://keycloak.example.com/realms/my-realm"
  ISSUER_URL: "https://keycloak.example.com/realms/my-realm"
```

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

