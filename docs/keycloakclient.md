# KeycloakClient

### Table of Contents
- [CRD Overview](#crd-overview)
- [Default Values Applied by Webhook](#default-values-applied-by-webhook)
- [Using an Existing Secret](#using-an-existing-secret)
- [Creating a Secret Automatically](#creating-a-secret-automatically)
- [Secret Creation](#secret-creation)
- [ConfigMap Creation](#configmap-creation)
- [ClientID Template Enforcement](#clientid-template-enforcement)

## CRD Overview
The operator manages Keycloak clients through the `KeycloakClient` Custom Resource Definition (CRD). This CRD supports various Keycloak client properties and configurations.

For detailed information about all available fields and their usage, please refer to the [KeycloakClient CRD documentation](docs/crds.md).

## Default Values Applied by Webhook
When creating or updating a KeycloakClient, the webhook automatically applies the following defaults:

- **ClientID**: If not specified, it will be auto-generated using the pattern:
  - If `--keycloak-client-id-required` template is used, that template is the value used for `clientID`. See [ClientID Template Enforcement](#clientid-template-enforcement) for details.
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
- `CLIENT_ID`: The client ID
- `CLIENT_SECRET`: The client secret value
- `COOKIE_SECRET`: A randomly generated secret used for OAuth2 Proxy cookie encryption

The `COOKIE_SECRET` is specifically intended to be used with OAuth2 Proxy for securing cookies. It is automatically generated upon Secret creation and not modified on updates.  If the cookie secret keys are removed from the Secret, a new random cookie secret will be added back to the Secret.

When `envVarKeys` is set to `false` in the ClientSecretRef configuration, the operator will use `client-id`, `client-secret` and `cookie-secret` keys.

Example Secret structure:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: example-client-secret
  namespace: default
type: Opaque
data:
  CLIENT_ID: "example-client"
  CLIENT_SECRET: <base64-encoded-secret>
  COOKIE_SECRET: <base64-encoded-cookie-secret>
```

## ConfigMap Creation
The operator creates Kubernetes ConfigMaps with Keycloak client configuration:
- `KEYCLOAK_URL`: The Keycloak server URL
- `KEYCLOAK_HOST`: The Keycloak host
- `ISSUER_URL`: The issuer URL for OpenID Connect

When `envVarKeys` is set to `false` in the ConfigMap configuration, the operator will use keys `keycloak-url`, `keycloak-host`, and `issuer-url`.

Example ConfigMap structure:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: example-client-config
  namespace: default
data:
  KEYCLOAK_URL: "https://keycloak.example.com"
  KEYCLOAK_HOST: "keycloak.example.com"
  ISSUER_URL: "https://keycloak.example.com/realms/my-realm"
```

## ClientID Template Enforcement
The operator supports enforcing client ID format using Go templates via the `--keycloak-client-id-required` flag. When this flag is set, all KeycloakClient resources must have a client ID that matches the provided template.

### How It Works:
1. When `--keycloak-client-id-required` is specified with a Go template, the operator will validate that all client IDs conform to the template.
2. During defaulting, if a client ID is not provided, the operator will evaluate the template to generate the required client ID.
3. During validation, if a client ID is explicitly provided, it must match the expected template output.

### Template Variables:
The template has access to the following data:
- `Obj` - The full KeycloakClient resource object
- `Config` - The operator configuration including:
  - `ClientIDPrefix` - The prefix used for client ID generation
  - `DefaultRealm` - The default realm used for clients

### Example Usage:
```bash
# Using a template that enforces a specific prefix
./keycloak-cr-operator --keycloak-client-id-prefix=kubernetes --keycloak-client-id-required='{{.Config.ClientIDPrefix}}-{{.Obj.Namespace}}-{{.Obj.Name}}'
```

This would enforce that all client IDs must follow the pattern: `prefix-namespace-name`.

### Example Template:
```yaml
apiVersion: keycloak.osc.edu/v1alpha1
kind: KeycloakClient
metadata:
  name: example-client
  namespace: default
spec:
  # This will be validated against the template
  clientID: "kubernetes-default-example-client"
  realm: "my-realm"
  # Other properties...
```

## OAuth2 Proxy Integration
The KeycloakClient operator integrates well with OAuth2 Proxy for secure authentication. When using OAuth2 Proxy, the operator generates the necessary `COOKIE_SECRET` in the client secret that OAuth2 Proxy can use for cookie encryption.

### Sample Configuration
For an example of how to configure a KeycloakClient for use with OAuth2 Proxy, see the sample configuration file:
- [`config/samples/keycloak_v1alpha1_keycloakclient_oauth2-proxy.yaml`](../config/samples/keycloak_v1alpha1_keycloakclient_oauth2-proxy.yaml)

### OAuth2 Proxy Configuration Values
The OAuth2 Proxy Helm configuration can be found in:
- [`config/samples/oauth2-proxy-values.yaml`](../config/samples/oauth2-proxy-values.yaml)

The minimal Helm configuration demonstrates how to reference the resources created by this operator.

These samples demonstrate how to:
1. Configure the KeycloakClient to work with OAuth2 Proxy
2. Set up the necessary secret references
3. Configure OAuth2 Proxy with proper environment variables and settings

The operator automatically creates the required secrets with `CLIENT_ID`, `CLIENT_SECRET`, and `COOKIE_SECRET` values that OAuth2 Proxy can consume.

The operator also automatically creates the ConfigMap used for things like `ISSUER_URL`, for exmaple.
