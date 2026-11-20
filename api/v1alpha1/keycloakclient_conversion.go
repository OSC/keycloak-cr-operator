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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilconversion "sigs.k8s.io/cluster-api/util/conversion"
	"sigs.k8s.io/controller-runtime/pkg/conversion"

	keycloakcommon "github.com/OSC/keycloak-cr-operator/api/common"
	keycloakv1alpha2 "github.com/OSC/keycloak-cr-operator/api/v1alpha2"
)

// ConvertTo converts this KeycloakClient (v1alpha1) to the Hub version (v1alpha2).
func (src *KeycloakClient) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*keycloakv1alpha2.KeycloakClient)
	// Copy ObjectMeta to preserve name, namespace, labels, etc.
	dst.ObjectMeta = src.ObjectMeta
	log.Printf("ConvertTo: Converting KeycloakClient from Spoke version v1alpha1 to Hub version v1alpha2;"+
		"source: %s/%s, target: %s/%s", src.Namespace, src.Name, dst.Namespace, dst.Name)

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
		var name, key *string
		if src.Spec.ClientSecretRef.Name != "" {
			name = new(src.Spec.ClientSecretRef.Name)
		}
		if src.Spec.ClientSecretRef.Key != "" {
			key = new(src.Spec.ClientSecretRef.Key)
		}
		dst.Spec.Secret = &keycloakv1alpha2.KeycloakClientSecret{
			Name:            name,
			ClientSecretKey: key,
			Create:          src.Spec.ClientSecretRef.Create,
			EnvVarKeys:      src.Spec.ClientSecretRef.EnvVarKeys,
			KeyPrefix:       src.Spec.ClientSecretRef.KeyPrefix,
		}
	} else {
		dst.Spec.Secret = nil
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
		dst.Spec.ProtocolMappers = make([]*keycloakcommon.KeycloakClientProtocolMapper, len(src.Spec.ProtocolMappers))
		for i := range src.Spec.ProtocolMappers {
			if src.Spec.ProtocolMappers[i] != nil {
				dst.Spec.ProtocolMappers[i] = src.Spec.ProtocolMappers[i].DeepCopy()
			}
		}
	} else {
		dst.Spec.ProtocolMappers = nil
	}

	if src.Status.Conditions != nil {
		dst.Status.Conditions = make([]metav1.Condition, len(src.Status.Conditions))
		copy(dst.Status.Conditions, src.Status.Conditions)
	}
	dst.Status.ID = src.Status.ID

	// Restore dropped fields from the annotation into dst
	if _, err := utilconversion.UnmarshalData(src, dst); err != nil {
		return err
	}

	if len(dst.Annotations) == 0 {
		dst.Annotations = nil
	}

	return nil
}

// ConvertFrom converts the Hub version (v1alpha2) to this KeycloakClient (v1alpha1).
func (dst *KeycloakClient) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*keycloakv1alpha2.KeycloakClient)
	// Copy ObjectMeta to preserve name, namespace, labels, etc.
	dst.ObjectMeta = src.ObjectMeta
	log.Printf("ConvertFrom: Converting KeycloakClient from Hub version v1alpha2 to Spoke version v1alpha1;"+
		"source: %s/%s, target: %s/%s", src.Namespace, src.Name, dst.Namespace, dst.Name)

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
	if src.Spec.Secret != nil {
		var name, key string
		if src.Spec.Secret.Name != nil {
			name = *src.Spec.Secret.Name
		}
		if src.Spec.Secret.ClientSecretKey != nil {
			key = *src.Spec.Secret.ClientSecretKey
		}
		dst.Spec.ClientSecretRef = &KeycloakClientSecret{
			SecretKeySelector: corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: name,
				},
				Key: key,
			},
			Create:     src.Spec.Secret.Create,
			EnvVarKeys: src.Spec.Secret.EnvVarKeys,
			KeyPrefix:  src.Spec.Secret.KeyPrefix,
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
		dst.Spec.ProtocolMappers = make([]*keycloakcommon.KeycloakClientProtocolMapper, len(src.Spec.ProtocolMappers))
		for i := range src.Spec.ProtocolMappers {
			if src.Spec.ProtocolMappers[i] != nil {
				dst.Spec.ProtocolMappers[i] = src.Spec.ProtocolMappers[i].DeepCopy()
			}
		}
	} else {
		dst.Spec.ProtocolMappers = nil
	}

	if src.Status.Conditions != nil {
		dst.Status.Conditions = make([]metav1.Condition, len(src.Status.Conditions))
		copy(dst.Status.Conditions, src.Status.Conditions)
	}
	dst.Status.ID = src.Status.ID

	if err := utilconversion.MarshalData(src, dst); err != nil {
		return err
	}

	return nil
}
