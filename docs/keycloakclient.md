# KeycloakClient

### Table of Contents
- [CRD Overview](#crd-overview)
- [Default Values Applied by Webhook](#default-values-applied-by-webhook)
- [Using an Existing Secret](#using-an-existing-secret)
- [Creating a Secret Automatically](#creating-a-secret-automatically)
- [Secret Creation](#secret-creation)
- [ConfigMap Creation](#configmap-creation)

## CRD Overview
The operator manages Keycloak clients through the `KeycloakClient` Custom Resource Definition (CRD). This CRD supports various Keycloak client properties and configurations.

For detailed information about all available fields and their usage, please refer to the [KeycloakClient CRD documentation](docs/crds.md).

## Default Values Applied by Webhook
When creating or updating a KeycloakClient, the webhook automatically applies the following defaults:

- **ClientID**: If not specified, it will be auto-generated using the pattern:
  - If `--keycloak-client-id-required` template is used, that template is the value used for `clientID`
  - If `--keycloak-client-id-prefix` is set in the operator configuration: `prefix-namespace-name`
  - If no prefix: `namespace-name`

- **Realm**: If not specified, it will default to the operator's configured `--keycloak-default-realm`

- **ClientSecretRef**: When `clientAuthenticatorType` is "client-secret" and `publicClient` is false:
  - If `clientSecretRef` is not set, it will be auto-created with:
    - `name`: `name-secret` (where `name` is the KeycloakClient resource name)
    - `key`: `client-secret`
    - `create`: `true` (indicating the secret should be auto-created)
    - `envVarKeys`: `true` (use EnvVar keys for Secret data)

- **ConfigMap**:
  - `name`: If not specified, it will default to `name-config` (where `name` is the KeycloakClient resource name)
  - `envVarKeys`: `true` (use EnvVar keys for ConfigMap data)

## Using an Existing Secret
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
    # Do not force EnvVar keys in Secret
    envVarKeys: false
  configMap:
    name: my-client-config
    # Use EnvVar keys in ConfigMap
    envVarKeys: true

  # Other client properties
  name: "Example Client"
  description: "An example Keycloak client"
  enabled: true
  publicClient: false
  redirectUris:
    - "https://example.com/callback"
  webOrigins:
    - "https://example.com"
  defaultClientScopes:
    - web-origins
    - profile
    - email
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

## Creating a Secret Automatically
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
    # Use EnvVar keys in Secret
    envVarKeys: true
  configMap:
    name: my-client-config
    # Use EnvVar keys in ConfigMap
    envVarKeys: true

  # Other client properties
  name: "Example Client with Auto Secret"
  description: "An example Keycloak client with auto-created secret"
  enabled: true
  publicClient: false
  redirectUris:
    - "https://example.com/callback"
  webOrigins:
    - "https://example.com"
  defaultClientScopes:
    - web-origins
    - profile
    - email
```

## Secret Creation
The operator creates Kubernetes Secrets containing client credentials with the following structure:
- `CLIENT_SECRET`: The client secret value
- `COOKIE_SECRET`: A randomly generated secret used for OAuth2 Proxy cookie encryption

The `COOKIE_SECRET` is specifically intended to be used with OAuth2 Proxy for securing cookies. It is automatically generated upon Secret creation and not modified on updates.  If the cookie secret keys are removed from the Secret, a new random cookie secret will be added back to the Secret.

When `envVarKeys` is set to `false` in the ClientSecretRef configuration, the operator will use `client-secret` and `cookie-secret` keys.

Example Secret structure:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: example-client-secret
  namespace: default
type: Opaque
data:
  CLIENT_SECRET: <base64-encoded-secret>
  COOKIE_SECRET: <base64-encoded-cookie-secret>
```

## ConfigMap Creation
The operator creates Kubernetes ConfigMaps with Keycloak client configuration:
- `CLIENT_ID`: The client ID
- `KEYCLOAK_URL`: The Keycloak server URL
- `ISSUER_URL`: The issuer URL for OpenID Connect

When `envVarKeys` is set to `false` in the ConfigMap configuration, the operator will use keys `client-id`, `keycloak-url` and `issuer-url`.

Example ConfigMap structure:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: example-client-config
  namespace: default
data:
  CLIENT_ID: "example-client"
  KEYCLOAK_URL: "https://keycloak.example.com"
  ISSUER_URL: "https://keycloak.example.com/realms/my-realm"
```
