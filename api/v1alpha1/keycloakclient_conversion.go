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
	"log"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"

	keycloakv1alpha2 "github.com/OSC/keycloak-cr-operator/api/v1alpha2"
)

// ConvertTo converts this KeycloakClient (v1alpha1) to the Hub version (v1alpha2).
func (src *KeycloakClient) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*keycloakv1alpha2.KeycloakClient)
	log.Printf("ConvertTo: Converting KeycloakClient from Spoke version v1alpha1 to Hub version v1alpha2;"+
		"source: %s/%s, target: %s/%s", src.Namespace, src.Name, dst.Namespace, dst.Name)

	// Copy ObjectMeta to preserve name, namespace, labels, etc.
	dst.ObjectMeta = src.ObjectMeta

	// Copy all spec fields from v1alpha1 to v1alpha2
	dst.Spec.AdminURL = src.Spec.AdminURL
	dst.Spec.AuthorizationServicesEnabled = src.Spec.AuthorizationServicesEnabled
	dst.Spec.BaseURL = src.Spec.BaseURL
	dst.Spec.BearerOnly = src.Spec.BearerOnly
	dst.Spec.ClientAuthenticatorType = src.Spec.ClientAuthenticatorType
	dst.Spec.ClientID = src.Spec.ClientID
	dst.Spec.ConsentRequired = src.Spec.ConsentRequired
	dst.Spec.DefaultClientScopes = src.Spec.DefaultClientScopes
	dst.Spec.DefaultRoles = src.Spec.DefaultRoles
	dst.Spec.Description = src.Spec.Description
	dst.Spec.DirectAccessGrantsEnabled = src.Spec.DirectAccessGrantsEnabled
	dst.Spec.Enabled = src.Spec.Enabled
	dst.Spec.FrontChannelLogout = src.Spec.FrontChannelLogout
	dst.Spec.FullScopeAllowed = src.Spec.FullScopeAllowed
	dst.Spec.ImplicitFlowEnabled = src.Spec.ImplicitFlowEnabled
	dst.Spec.Name = src.Spec.Name
	dst.Spec.NodeReRegistrationTimeout = src.Spec.NodeReRegistrationTimeout
	dst.Spec.NotBefore = src.Spec.NotBefore
	dst.Spec.OptionalClientScopes = src.Spec.OptionalClientScopes
	dst.Spec.Origin = src.Spec.Origin
	dst.Spec.Protocol = src.Spec.Protocol
	dst.Spec.PublicClient = src.Spec.PublicClient
	dst.Spec.RedirectURIs = src.Spec.RedirectURIs
	dst.Spec.RegistrationAccessToken = src.Spec.RegistrationAccessToken
	dst.Spec.RootURL = src.Spec.RootURL
	dst.Spec.ServiceAccountsEnabled = src.Spec.ServiceAccountsEnabled
	dst.Spec.StandardFlowEnabled = src.Spec.StandardFlowEnabled
	dst.Spec.SurrogateAuthRequired = src.Spec.SurrogateAuthRequired
	dst.Spec.WebOrigins = src.Spec.WebOrigins
	dst.Spec.LoginTheme = src.Spec.LoginTheme
	dst.Spec.Realm = src.Spec.Realm

	// Handle ClientSecretRef conversion
	if src.Spec.ClientSecretRef != nil {
		dst.Spec.ClientSecretRef = &keycloakv1alpha2.KeycloakClientSecret{
			SecretKeySelector: corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: src.Spec.ClientSecretRef.Name,
				},
				Key: src.Spec.ClientSecretRef.Key,
			},
			Create:     src.Spec.ClientSecretRef.Create,
			EnvVarKeys: src.Spec.ClientSecretRef.EnvVarKeys,
			KeyPrefix:  src.Spec.ClientSecretRef.KeyPrefix,
		}
	} else {
		dst.Spec.ClientSecretRef = nil
	}

	// Handle ConfigMap conversion
	if src.Spec.ConfigMap != nil {
		dst.Spec.ConfigMap = &keycloakv1alpha2.KeycloakClientConfigMap{
			Name:       src.Spec.ConfigMap.Name,
			EnvVarKeys: src.Spec.ConfigMap.EnvVarKeys,
		}
	} else {
		dst.Spec.ConfigMap = nil
	}

	// Handle ProtocolMappers conversion
	if src.Spec.ProtocolMappers != nil {
		dst.Spec.ProtocolMappers = make([]*keycloakv1alpha2.KeycloakClientProtocolMapper, len(src.Spec.ProtocolMappers))
		for i, mapper := range src.Spec.ProtocolMappers {
			if mapper != nil {
				dst.Spec.ProtocolMappers[i] = &keycloakv1alpha2.KeycloakClientProtocolMapper{
					Name:                   mapper.Name,
					Protocol:               mapper.Protocol,
					Type:                   mapper.Type,
					IDTokenClaim:           mapper.IDTokenClaim,
					AccessTokenClaim:       mapper.AccessTokenClaim,
					IncludedClientAudience: mapper.IncludedClientAudience,
					ConsentRequired:        mapper.ConsentRequired,
					Config:                 mapper.Config,
				}
			} else {
				dst.Spec.ProtocolMappers[i] = nil
			}
		}
	} else {
		dst.Spec.ProtocolMappers = nil
	}

	return nil
}

// ConvertFrom converts the Hub version (v1alpha2) to this KeycloakClient (v1alpha1).
func (dst *KeycloakClient) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*keycloakv1alpha2.KeycloakClient)
	log.Printf("ConvertFrom: Converting KeycloakClient from Hub version v1alpha2 to Spoke version v1alpha1;"+
		"source: %s/%s, target: %s/%s", src.Namespace, src.Name, dst.Namespace, dst.Name)

	// Copy ObjectMeta to preserve name, namespace, labels, etc.
	dst.ObjectMeta = src.ObjectMeta

	// Copy all spec fields from v1alpha2 to v1alpha1
	dst.Spec.AdminURL = src.Spec.AdminURL
	dst.Spec.AuthorizationServicesEnabled = src.Spec.AuthorizationServicesEnabled
	dst.Spec.BaseURL = src.Spec.BaseURL
	dst.Spec.BearerOnly = src.Spec.BearerOnly
	dst.Spec.ClientAuthenticatorType = src.Spec.ClientAuthenticatorType
	dst.Spec.ClientID = src.Spec.ClientID
	dst.Spec.ConsentRequired = src.Spec.ConsentRequired
	dst.Spec.DefaultClientScopes = src.Spec.DefaultClientScopes
	dst.Spec.DefaultRoles = src.Spec.DefaultRoles
	dst.Spec.Description = src.Spec.Description
	dst.Spec.DirectAccessGrantsEnabled = src.Spec.DirectAccessGrantsEnabled
	dst.Spec.Enabled = src.Spec.Enabled
	dst.Spec.FrontChannelLogout = src.Spec.FrontChannelLogout
	dst.Spec.FullScopeAllowed = src.Spec.FullScopeAllowed
	dst.Spec.ImplicitFlowEnabled = src.Spec.ImplicitFlowEnabled
	dst.Spec.Name = src.Spec.Name
	dst.Spec.NodeReRegistrationTimeout = src.Spec.NodeReRegistrationTimeout
	dst.Spec.NotBefore = src.Spec.NotBefore
	dst.Spec.OptionalClientScopes = src.Spec.OptionalClientScopes
	dst.Spec.Origin = src.Spec.Origin
	dst.Spec.Protocol = src.Spec.Protocol
	dst.Spec.PublicClient = src.Spec.PublicClient
	dst.Spec.RedirectURIs = src.Spec.RedirectURIs
	dst.Spec.RegistrationAccessToken = src.Spec.RegistrationAccessToken
	dst.Spec.RootURL = src.Spec.RootURL
	dst.Spec.ServiceAccountsEnabled = src.Spec.ServiceAccountsEnabled
	dst.Spec.StandardFlowEnabled = src.Spec.StandardFlowEnabled
	dst.Spec.SurrogateAuthRequired = src.Spec.SurrogateAuthRequired
	dst.Spec.WebOrigins = src.Spec.WebOrigins
	dst.Spec.LoginTheme = src.Spec.LoginTheme
	dst.Spec.Realm = src.Spec.Realm

	// Handle ClientSecretRef conversion
	if src.Spec.ClientSecretRef != nil {
		dst.Spec.ClientSecretRef = &KeycloakClientSecret{
			SecretKeySelector: corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: src.Spec.ClientSecretRef.Name,
				},
				Key: src.Spec.ClientSecretRef.Key,
			},
			Create:     src.Spec.ClientSecretRef.Create,
			EnvVarKeys: src.Spec.ClientSecretRef.EnvVarKeys,
			KeyPrefix:  src.Spec.ClientSecretRef.KeyPrefix,
		}
	} else {
		dst.Spec.ClientSecretRef = nil
	}

	// Handle ConfigMap conversion
	if src.Spec.ConfigMap != nil {
		dst.Spec.ConfigMap = &KeycloakClientConfigMap{
			Name:       src.Spec.ConfigMap.Name,
			EnvVarKeys: src.Spec.ConfigMap.EnvVarKeys,
		}
	} else {
		dst.Spec.ConfigMap = nil
	}

	// Handle ProtocolMappers conversion
	if src.Spec.ProtocolMappers != nil {
		dst.Spec.ProtocolMappers = make([]*KeycloakClientProtocolMapper, len(src.Spec.ProtocolMappers))
		for i, mapper := range src.Spec.ProtocolMappers {
			if mapper != nil {
				dst.Spec.ProtocolMappers[i] = &KeycloakClientProtocolMapper{
					Name:                   mapper.Name,
					Protocol:               mapper.Protocol,
					Type:                   mapper.Type,
					IDTokenClaim:           mapper.IDTokenClaim,
					AccessTokenClaim:       mapper.AccessTokenClaim,
					IncludedClientAudience: mapper.IncludedClientAudience,
					ConsentRequired:        mapper.ConsentRequired,
					Config:                 mapper.Config,
				}
			} else {
				dst.Spec.ProtocolMappers[i] = nil
			}
		}
	} else {
		dst.Spec.ProtocolMappers = nil
	}

	return nil
}
