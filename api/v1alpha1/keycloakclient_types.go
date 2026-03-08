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

	// Attributes is the map of additional attributes
	// +optional
	Attributes map[string]string `json:"attributes,omitempty"`

	// AuthenticationFlowBindingOverrides is the map of authentication flow binding overrides
	// +optional
	AuthenticationFlowBindingOverrides map[string]string `json:"authenticationFlowBindingOverrides,omitempty"`

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
	// +required
	ClientID string `json:"clientID"`

	// ClientTemplate is the client template
	// +optional
	ClientTemplate *string `json:"clientTemplate,omitempty"`

	// ConsentRequired indicates if consent is required
	// +optional
	ConsentRequired *bool `json:"consentRequired,omitempty"`

	// DefaultClientScopes is the default client scopes
	// +optional
	DefaultClientScopes []string `json:"defaultClientScopes,omitempty"`

	// DefaultRoles is the default roles
	// +optional
	DefaultRoles []string `json:"defaultRoles,omitempty"`

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

	// ManagementURL is the management URL for the client
	// +optional
	ManagementURL *string `json:"managementUrl,omitempty"`

	// Name is the display name for the client
	// +optional
	Name *string `json:"name,omitempty"`

	// NodeReRegistrationTimeout is the node re-registration timeout
	// +optional
	NodeReRegistrationTimeout *int `json:"nodeReRegistrationTimeout,omitempty"`

	// NotBefore is the not before setting
	// +optional
	NotBefore *int `json:"notBefore,omitempty"`

	// OptionalClientScopes is the optional client scopes
	// +optional
	OptionalClientScopes []string `json:"optionalClientScopes,omitempty"`

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
	// +required
	RedirectURIs []string `json:"redirectUris"`

	// RegisteredNodes is the registered nodes
	// +optional
	RegisteredNodes map[string]int `json:"registeredNodes,omitempty"`

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

	// UseTemplateConfig indicates if template config is used
	// +optional
	UseTemplateConfig *bool `json:"useTemplateConfig,omitempty"`

	// UseTemplateMappers indicates if template mappers are used
	// +optional
	UseTemplateMappers *bool `json:"useTemplateMappers,omitempty"`

	// UseTemplateScope indicates if template scope is used
	// +optional
	UseTemplateScope *bool `json:"useTemplateScope,omitempty"`

	// WebOrigins is the list of valid web origins
	// +optional
	WebOrigins []string `json:"webOrigins,omitempty"`
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
