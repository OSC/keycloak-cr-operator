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

	"github.com/OSC/keycloak-cr-operator/internal/models"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetConfigMap creates a corev1.ConfigMap object using data from KeycloakClient for name and namespace
// and the provided KeycloakConfig for host information. It sets data based on the client information
// and KeycloakConfig.
func (k *KeycloakClient) GetConfigMap(config *models.KeycloakConfig) *corev1.ConfigMap {
	var name string
	var configMap *KeycloakClientConfigMap
	defaultName := fmt.Sprintf("%s-config", k.Name)
	if k.Spec.ConfigMap == nil {
		envVarKeys := true
		configMap = &KeycloakClientConfigMap{
			Name:       &defaultName,
			EnvVarKeys: &envVarKeys,
		}
	} else {
		configMap = k.Spec.ConfigMap
	}
	if configMap.Name == nil || *configMap.Name == "" {
		name = defaultName
	} else {
		name = *configMap.Name
	}
	realm := config.DefaultRealm
	if k.Spec.Realm != nil && *k.Spec.Realm != "" {
		realm = *k.Spec.Realm
	}

	// Create data map for ConfigMap
	url := config.KeycloakURL.String()
	host := config.KeycloakURL.Host
	issuerUrl := config.KeycloakURL.JoinPath("realms", realm).String()
	data := make(map[string]string)
	if configMap.EnvVarKeys == nil || *configMap.EnvVarKeys {
		data["KEYCLOAK_URL"] = url
		data["KEYCLOAK_HOST"] = host
		data["ISSUER_URL"] = issuerUrl
	} else {
		data["keycloak-url"] = url
		data["keycloak-host"] = host
		data["issuer-url"] = issuerUrl
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: k.Namespace,
			Name:      name,
		},
		Data: data,
	}
}
