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

// +kubebuilder:object:generate=true
package common

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KeycloakClientProtocolMapper struct {
	// Name of the protocol mapper
	// +required
	Name *string `json:"name,omitempty"`
	// The protocol mapper protocol
	// +kubebuilder:default="openid-connect"
	// +optional
	Protocol *string `json:"protocol,omitempty"`
	// The protocol mapper type
	// +required
	Type *string `json:"type,omitempty"`
	// Is this ID claim token
	// +kubebuilder:default=true
	// +optional
	IDTokenClaim *bool `json:"idTokenClaim,omitempty"`
	// Is this access claim token
	// +kubebuilder:default=true
	// +optional
	AccessTokenClaim *bool `json:"accessTokenClaim,omitempty"`
	// Included client audience
	// Required for type=oidc-audience-mapper
	// +optional
	IncludedClientAudience *string `json:"includedClientAudience,omitempty"`
	// Consent required
	// +kubebuilder:default=false
	// +optional
	ConsentRequired *bool `json:"consentRequired,omitempty"`
	// Additional configs
	// +optional
	Config *map[string]string `json:"config,omitempty"`
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

	// The ID of the Keycloak Client object in Keycloak
	// +optional
	ID *string `json:"id"`
}
