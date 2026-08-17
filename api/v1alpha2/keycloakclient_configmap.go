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

package v1alpha2

import (
	"fmt"

	"github.com/OSC/keycloak-cr-operator/internal/models"
	"github.com/stoewer/go-strcase"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetConfigMap creates a corev1.ConfigMap object using data from KeycloakClient for name and namespace
// and the provided KeycloakConfig for host information. It sets data based on the client information
// and KeycloakConfig.
func (k *KeycloakClient) GetConfigMap(config *models.KeycloakConfig) *corev1.ConfigMap {
	name := k.ConfigMapName()
	realm := config.DefaultRealm
	if k.Spec.Realm != nil && *k.Spec.Realm != "" {
		realm = *k.Spec.Realm
	}

	url := config.KeycloakURL.String()
	host := config.KeycloakURL.Host
	issuerUrl := config.KeycloakURL.JoinPath("realms", realm)
	providerUrl := issuerUrl.JoinPath(".well-known/openid-configuration")

	data := make(map[string]string)
	data[k.ConfigMapKeycloakUrlKey()] = url
	data[k.ConfigMapKeycloakHostKey()] = host
	data[k.ConfigMapIssuerUrlKey()] = issuerUrl.String()
	data[k.ConfigMapProviderUrlKey()] = providerUrl.String()

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: k.Namespace,
			Name:      name,
		},
		Data: data,
	}
}

func (k *KeycloakClient) ConfigMapName() string {
	name := fmt.Sprintf("%s-config", k.Name)
	if k.Spec.ConfigMap != nil {
		if k.Spec.ConfigMap.Name != nil && *k.Spec.ConfigMap.Name != "" {
			name = *k.Spec.ConfigMap.Name
		}
	}
	return name
}

func (k *KeycloakClient) ConfigMapEnvVarKeys() bool {
	envVarKeys := true
	if k.Spec.ConfigMap != nil {
		if k.Spec.ConfigMap.EnvVarKeys != nil {
			envVarKeys = *k.Spec.ConfigMap.EnvVarKeys
		}
	}
	return envVarKeys
}

func (k *KeycloakClient) ConfigMapKeycloakUrlKey() string {
	key := "keycloak-url"
	envVarKeys := k.ConfigMapEnvVarKeys()
	if k.Spec.ConfigMap != nil {
		if k.Spec.ConfigMap.KeycloakUrlKey != nil && *k.Spec.ConfigMap.KeycloakUrlKey != "" {
			key = *k.Spec.ConfigMap.KeycloakUrlKey
		}
	}
	if envVarKeys {
		key = strcase.UpperSnakeCase(key)
	}
	return key
}

func (k *KeycloakClient) ConfigMapKeycloakHostKey() string {
	key := "keycloak-host"
	envVarKeys := k.ConfigMapEnvVarKeys()
	if k.Spec.ConfigMap != nil {
		if k.Spec.ConfigMap.KeycloakHostKey != nil && *k.Spec.ConfigMap.KeycloakHostKey != "" {
			key = *k.Spec.ConfigMap.KeycloakHostKey
		}
	}
	if envVarKeys {
		key = strcase.UpperSnakeCase(key)
	}
	return key
}

func (k *KeycloakClient) ConfigMapIssuerUrlKey() string {
	key := "issuer-url"
	envVarKeys := k.ConfigMapEnvVarKeys()
	if k.Spec.ConfigMap != nil {
		if k.Spec.ConfigMap.IssuerUrlKey != nil && *k.Spec.ConfigMap.IssuerUrlKey != "" {
			key = *k.Spec.ConfigMap.IssuerUrlKey
		}
	}
	if envVarKeys {
		key = strcase.UpperSnakeCase(key)
	}
	return key
}

func (k *KeycloakClient) ConfigMapProviderUrlKey() string {
	key := "provider-url"
	envVarKeys := k.ConfigMapEnvVarKeys()
	if k.Spec.ConfigMap != nil {
		if k.Spec.ConfigMap.ProviderUrlKey != nil && *k.Spec.ConfigMap.ProviderUrlKey != "" {
			key = *k.Spec.ConfigMap.ProviderUrlKey
		}
	}
	if envVarKeys {
		key = strcase.UpperSnakeCase(key)
	}
	return key
}
