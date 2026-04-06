# API Reference

Packages:

- [keycloak.osc.edu/v1alpha1](#keycloakosceduv1alpha1)

# keycloak.osc.edu/v1alpha1

Resource Types:

- [KeycloakClient](#keycloakclient)




## KeycloakClient
<sup><sup>[↩ Parent](#keycloakosceduv1alpha1 )</sup></sup>






KeycloakClient is the Schema for the keycloakclients API

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>keycloak.osc.edu/v1alpha1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>KeycloakClient</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#keycloakclientspec">spec</a></b></td>
        <td>object</td>
        <td>
          spec defines the desired state of KeycloakClient<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#keycloakclientstatus">status</a></b></td>
        <td>object</td>
        <td>
          status defines the observed state of KeycloakClient<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### KeycloakClient.spec
<sup><sup>[↩ Parent](#keycloakclient)</sup></sup>



spec defines the desired state of KeycloakClient

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>adminUrl</b></td>
        <td>string</td>
        <td>
          AdminURL is the URL for the admin console<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>authorizationServicesEnabled</b></td>
        <td>boolean</td>
        <td>
          AuthorizationServicesEnabled indicates if authorization services are enabled<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>baseUrl</b></td>
        <td>string</td>
        <td>
          BaseURL is the base URL for the client<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>bearerOnly</b></td>
        <td>boolean</td>
        <td>
          BearerOnly indicates if the client is bearer-only<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#keycloakclientspecchecksumref">checksumRef</a></b></td>
        <td>object</td>
        <td>
          The Checksum reference configuration<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>clientAuthenticatorType</b></td>
        <td>string</td>
        <td>
          ClientAuthenticatorType is the client authenticator type<br/>
          <br/>
            <i>Default</i>: client-secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>clientID</b></td>
        <td>string</td>
        <td>
          ClientID is the unique identifier for the client<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#keycloakclientspecclientsecretref">clientSecretRef</a></b></td>
        <td>object</td>
        <td>
          Reference to the secret holding the ClientSecret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#keycloakclientspecconfigmap">configMap</a></b></td>
        <td>object</td>
        <td>
          The ConfigMap configuration<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>consentRequired</b></td>
        <td>boolean</td>
        <td>
          ConsentRequired indicates if consent is required<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>defaultClientScopes</b></td>
        <td>[]string</td>
        <td>
          DefaultClientScopes is the default client scopes<br/>
          <br/>
            <i>Default</i>: []<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>defaultRoles</b></td>
        <td>[]string</td>
        <td>
          DefaultRoles is the default roles<br/>
          <br/>
            <i>Default</i>: []<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>description</b></td>
        <td>string</td>
        <td>
          Description is the description of the client<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>directAccessGrantsEnabled</b></td>
        <td>boolean</td>
        <td>
          DirectAccessGrantsEnabled indicates if direct access grants are enabled<br/>
          <br/>
            <i>Default</i>: true<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>enabled</b></td>
        <td>boolean</td>
        <td>
          Enabled indicates if the client is enabled<br/>
          <br/>
            <i>Default</i>: true<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>frontChannelLogout</b></td>
        <td>boolean</td>
        <td>
          FrontChannelLogout is the front channel logout setting<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>fullScopeAllowed</b></td>
        <td>boolean</td>
        <td>
          FullScopeAllowed indicates if full scope is allowed<br/>
          <br/>
            <i>Default</i>: true<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>implicitFlowEnabled</b></td>
        <td>boolean</td>
        <td>
          ImplicitFlowEnabled indicates if implicit flow is enabled<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>loginTheme</b></td>
        <td>string</td>
        <td>
          The client's login theme<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is the display name for the client<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>nodeReRegistrationTimeout</b></td>
        <td>integer</td>
        <td>
          NodeReRegistrationTimeout is the node re-registration timeout<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>notBefore</b></td>
        <td>integer</td>
        <td>
          NotBefore is the not before setting<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optionalClientScopes</b></td>
        <td>[]string</td>
        <td>
          OptionalClientScopes is the optional client scopes<br/>
          <br/>
            <i>Default</i>: []<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>origin</b></td>
        <td>string</td>
        <td>
          Origin is the origin of the client<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>protocol</b></td>
        <td>enum</td>
        <td>
          Protocol is the protocol type<br/>
          <br/>
            <i>Enum</i>: openid-connect<br/>
            <i>Default</i>: openid-connect<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>publicClient</b></td>
        <td>boolean</td>
        <td>
          PublicClient indicates if the client is public<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>realm</b></td>
        <td>string</td>
        <td>
          The Realm for the Keycloak Client<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>redirectUris</b></td>
        <td>[]string</td>
        <td>
          RedirectURIs is the list of valid redirect URIs<br/>
          <br/>
            <i>Default</i>: []<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>registrationAccessToken</b></td>
        <td>string</td>
        <td>
          RegistrationAccessToken is the registration access token<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>rootUrl</b></td>
        <td>string</td>
        <td>
          RootURL is the root URL for the client<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>serviceAccountsEnabled</b></td>
        <td>boolean</td>
        <td>
          ServiceAccountsEnabled indicates if service accounts are enabled<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>standardFlowEnabled</b></td>
        <td>boolean</td>
        <td>
          StandardFlowEnabled indicates if standard flow is enabled<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>surrogateAuthRequired</b></td>
        <td>boolean</td>
        <td>
          SurrogateAuthRequired indicates if surrogate authentication is required<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>webOrigins</b></td>
        <td>[]string</td>
        <td>
          WebOrigins is the list of valid web origins<br/>
          <br/>
            <i>Default</i>: []<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### KeycloakClient.spec.checksumRef
<sup><sup>[↩ Parent](#keycloakclientspec)</sup></sup>



The Checksum reference configuration

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>kind</b></td>
        <td>enum</td>
        <td>
          The type of resource to restart<br/>
          <br/>
            <i>Enum</i>: Deployment, StatefulSet<br/>
            <i>Default</i>: Deployment<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the resource to restart<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### KeycloakClient.spec.clientSecretRef
<sup><sup>[↩ Parent](#keycloakclientspec)</sup></sup>



Reference to the secret holding the ClientSecret

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>create</b></td>
        <td>boolean</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: true<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>envVarKeys</b></td>
        <td>boolean</td>
        <td>
          Whether to use envVar keys<br/>
          <br/>
            <i>Default</i>: true<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
This field is effectively required, but due to backwards compatibility is
allowed to be empty. Instances of this type with an empty value here are
almost certainly wrong.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### KeycloakClient.spec.configMap
<sup><sup>[↩ Parent](#keycloakclientspec)</sup></sup>



The ConfigMap configuration

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>envVarKeys</b></td>
        <td>boolean</td>
        <td>
          Whether to use envVar keys<br/>
          <br/>
            <i>Default</i>: true<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          The ConfigMap name, will default to "<name>-config"<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### KeycloakClient.status
<sup><sup>[↩ Parent](#keycloakclient)</sup></sup>



status defines the observed state of KeycloakClient

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#keycloakclientstatusconditionsindex">conditions</a></b></td>
        <td>[]object</td>
        <td>
          conditions represent the current state of the KeycloakClient resource.
Each condition has a unique type and reflects the status of a specific aspect of the resource.

Standard condition types include:
- "Available": the resource is fully functional
- "Progressing": the resource is being created or updated
- "Degraded": the resource failed to reach or maintain its desired state

The status of each condition is one of True, False, or Unknown.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>id</b></td>
        <td>string</td>
        <td>
          The ID of the Keycloak Client object in Keycloak<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### KeycloakClient.status.conditions[index]
<sup><sup>[↩ Parent](#keycloakclientstatus)</sup></sup>



Condition contains details for one aspect of the current state of this API Resource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>lastTransitionTime</b></td>
        <td>string</td>
        <td>
          lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          message is a human readable message indicating details about the transition.
This may be an empty string.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>reason</b></td>
        <td>string</td>
        <td>
          reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          status of the condition, one of True, False, Unknown.<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type of condition in CamelCase or in foo.example.com/CamelCase.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>