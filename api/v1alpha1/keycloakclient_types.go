/*
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
*/

package v1alpha1

import (
	"fmt"

	"github.com/Nerzal/gocloak/v13"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KeycloakClientSpec defines the desired state of KeycloakClient
type KeycloakClientSpec struct {
	// Access TODO: Add Access property later
	// +optional
	// Access interface{} `json:"access,omitempty"`

	// AdminURL is the URL for the admin console
	// +optional
	AdminURL *string `json:"adminUrl,omitempty"`

	// Do not implement attributes, instead populate as needed
	// Attributes map[string]string `json:"attributes,omitempty"`

	// AuthenticationFlowBindingOverrides is the map of authentication flow binding overrides
	// TODO: Implement if needed
	// +optional
	// AuthenticationFlowBindingOverrides *map[string]string `json:"authenticationFlowBindingOverrides,omitempty"`

	// AuthorizationServicesEnabled indicates if authorization services are enabled
	// +kubebuilder:default=false
	// +optional
	AuthorizationServicesEnabled *bool `json:"authorizationServicesEnabled,omitempty"`

	// AuthorizationSettings TODO: Add AuthorizationSettings property later
	// +optional
	// AuthorizationSettings interface{} `json:"authorizationSettings,omitempty"`

	// BaseURL is the base URL for the client
	// +optional
	BaseURL *string `json:"baseUrl,omitempty"`

	// BearerOnly indicates if the client is bearer-only
	// +kubebuilder:default=false
	// +optional
	BearerOnly *bool `json:"bearerOnly,omitempty"`

	// ClientAuthenticatorType is the client authenticator type
	// +kubebuilder:default="client-secret"
	// +optional
	ClientAuthenticatorType *string `json:"clientAuthenticatorType,omitempty"`

	// ClientID is the unique identifier for the client
	// +optional
	ClientID *string `json:"clientID"`

	// ConsentRequired indicates if consent is required
	// +optional
	ConsentRequired *bool `json:"consentRequired,omitempty"`

	// DefaultClientScopes is the default client scopes
	// +optional
	DefaultClientScopes *[]string `json:"defaultClientScopes,omitempty"`

	// DefaultRoles is the default roles
	// +optional
	DefaultRoles *[]string `json:"defaultRoles,omitempty"`

	// Description is the description of the client
	// +optional
	Description *string `json:"description,omitempty"`

	// DirectAccessGrantsEnabled indicates if direct access grants are enabled
	// +kubebuilder:default=true
	// +optional
	DirectAccessGrantsEnabled *bool `json:"directAccessGrantsEnabled,omitempty"`

	// Enabled indicates if the client is enabled
	// +kubebuilder:default=true
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// FrontChannelLogout is the front channel logout setting
	// +optional
	FrontChannelLogout *bool `json:"frontChannelLogout,omitempty"`

	// FullScopeAllowed indicates if full scope is allowed
	// +kubebuilder:default=true
	// +optional
	FullScopeAllowed *bool `json:"fullScopeAllowed,omitempty"`

	// ID is the unique identifier for the client
	// +optional
	ID *string `json:"id,omitempty"`

	// ImplicitFlowEnabled indicates if implicit flow is enabled
	// +kubebuilder:default=false
	// +optional
	ImplicitFlowEnabled *bool `json:"implicitFlowEnabled,omitempty"`

	// Name is the display name for the client
	// +optional
	Name *string `json:"name,omitempty"`

	// NodeReRegistrationTimeout is the node re-registration timeout
	// +optional
	NodeReRegistrationTimeout *int32 `json:"nodeReRegistrationTimeout,omitempty"`

	// NotBefore is the not before setting
	// +optional
	NotBefore *int32 `json:"notBefore,omitempty"`

	// OptionalClientScopes is the optional client scopes
	// +optional
	OptionalClientScopes *[]string `json:"optionalClientScopes,omitempty"`

	// Origin is the origin of the client
	// +optional
	Origin *string `json:"origin,omitempty"`

	// Protocol is the protocol type
	// +kubebuilder:validation:Enum=openid-connect
	// +kubebuilder:default="openid-connect"
	// +optional
	Protocol *string `json:"protocol,omitempty"`

	// ProtocolMappers TODO: Add ProtocolMappers property later
	// +optional
	// ProtocolMappers interface{} `json:"protocolMappers,omitempty"`

	// PublicClient indicates if the client is public
	// +kubebuilder:default=false
	// +optional
	PublicClient *bool `json:"publicClient,omitempty"`

	// RedirectURIs is the list of valid redirect URIs
	// +optional
	RedirectURIs *[]string `json:"redirectUris"`

	// RegisteredNodes is the registered nodes, TODO: Add RegisteredNodes later
	// +optional
	// RegisteredNodes *map[string]int `json:"registeredNodes,omitempty"`

	// RegistrationAccessToken is the registration access token
	// +optional
	RegistrationAccessToken *string `json:"registrationAccessToken,omitempty"`

	// RootURL is the root URL for the client
	// +optional
	RootURL *string `json:"rootUrl,omitempty"`

	// Secret is the client secret
	// +optional
	Secret *string `json:"secret,omitempty"`

	// ServiceAccountsEnabled indicates if service accounts are enabled
	// +kubebuilder:default=false
	// +optional
	ServiceAccountsEnabled *bool `json:"serviceAccountsEnabled,omitempty"`

	// StandardFlowEnabled indicates if standard flow is enabled
	// +optional
	StandardFlowEnabled *bool `json:"standardFlowEnabled,omitempty"`

	// SurrogateAuthRequired indicates if surrogate authentication is required
	// +optional
	SurrogateAuthRequired *bool `json:"surrogateAuthRequired,omitempty"`

	// WebOrigins is the list of valid web origins
	// +optional
	WebOrigins *[]string `json:"webOrigins,omitempty"`

	// BEGIN ATTRIBUTES

	// The client's login theme
	// +optional
	LoginTheme *string `json:"loginTheme,omitempty"`

	// END ATTRIBUTES

	// The Realm for the Keycloak Client
	// +optional
	Realm *string `json:"realm,omitempty"`

	// Reference to the secret holding the ClientSecret
	// +optional
	ClientSecretRef *corev1.SecretKeySelector `json:"clientSecretRef,omitempty"`
}

// KeycloakClientStatus defines the observed state of KeycloakClient.
type KeycloakClientStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the KeycloakClient resource.
	// Each condition has a unique type and reflects the status of a specific aspect of the resource.
	//
	// Standard condition types include:
	// - "Available": the resource is fully functional
	// - "Progressing": the resource is being created or updated
	// - "Degraded": the resource failed to reach or maintain its desired state
	//
	// The status of each condition is one of True, False, or Unknown.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// KeycloakClient is the Schema for the keycloakclients API
type KeycloakClient struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of KeycloakClient
	// +required
	Spec KeycloakClientSpec `json:"spec"`

	// status defines the observed state of KeycloakClient
	// +optional
	Status KeycloakClientStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// KeycloakClientList contains a list of KeycloakClient
type KeycloakClientList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []KeycloakClient `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KeycloakClient{}, &KeycloakClientList{})
}

func (k *KeycloakClient) GetClient(prefix string, clientSecret string) *gocloak.Client {
	client := &gocloak.Client{}

	if k.Spec.ClientID == nil || *k.Spec.ClientID == "" {
		clientID := fmt.Sprintf("%s-%s-%s", prefix, k.Namespace, k.Name)
		client.ClientID = &clientID
	} else {
		client.ClientID = k.Spec.ClientID
	}
	if k.Spec.ID == nil || *k.Spec.ID == "" {
		client.ID = client.ClientID
	} else {
		client.ID = k.Spec.ID
	}
	attributes := make(map[string]string)
	if k.Spec.LoginTheme != nil && *k.Spec.LoginTheme != "" {
		attributes["login_theme"] = *k.Spec.LoginTheme
	}
	if len(attributes) > 0 {
		client.Attributes = &attributes
	}

	client.AdminURL = k.Spec.AdminURL
	client.AuthorizationServicesEnabled = k.Spec.AuthorizationServicesEnabled
	client.BaseURL = k.Spec.BaseURL
	client.BearerOnly = k.Spec.BearerOnly
	client.ClientAuthenticatorType = k.Spec.ClientAuthenticatorType
	client.ClientSecret = &clientSecret
	client.ConsentRequired = k.Spec.ConsentRequired
	client.DefaultClientScopes = k.Spec.DefaultClientScopes
	client.DefaultRoles = k.Spec.DefaultRoles
	client.Description = k.Spec.Description
	client.DirectAccessGrantsEnabled = k.Spec.DirectAccessGrantsEnabled
	client.Enabled = k.Spec.Enabled
	client.FrontChannelLogout = k.Spec.FrontChannelLogout
	client.FullScopeAllowed = k.Spec.FullScopeAllowed
	client.ImplicitFlowEnabled = k.Spec.ImplicitFlowEnabled
	client.Name = k.Spec.Name
	client.NodeReRegistrationTimeout = k.Spec.NodeReRegistrationTimeout
	client.NotBefore = k.Spec.NotBefore
	client.OptionalClientScopes = k.Spec.OptionalClientScopes
	client.Origin = k.Spec.Origin
	client.Protocol = k.Spec.Protocol
	client.PublicClient = k.Spec.PublicClient
	client.RedirectURIs = k.Spec.RedirectURIs
	client.RegistrationAccessToken = k.Spec.RegistrationAccessToken
	client.RootURL = k.Spec.RootURL
	client.ServiceAccountsEnabled = k.Spec.ServiceAccountsEnabled
	client.StandardFlowEnabled = k.Spec.StandardFlowEnabled
	client.SurrogateAuthRequired = k.Spec.SurrogateAuthRequired
	client.WebOrigins = k.Spec.WebOrigins

	return client
}
