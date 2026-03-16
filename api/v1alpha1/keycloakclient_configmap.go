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
	var name, clientID string
	if k.Spec.ConfigMapName == nil || *k.Spec.ConfigMapName == "" {
		name = fmt.Sprintf("%s-config", k.Name)
	} else {
		name = *k.Spec.ConfigMapName
	}
	if k.Spec.ClientID == nil || *k.Spec.ClientID == "" {
		if config.ClientIDPrefix != "" {
			clientID = fmt.Sprintf("%s-%s-%s", config.ClientIDPrefix, k.Namespace, k.Name)
		} else {
			clientID = fmt.Sprintf("%s-%s", k.Namespace, k.Name)
		}
	} else {
		clientID = *k.Spec.ClientID
	}
	realm := config.DefaultRealm
	if k.Spec.Realm != nil && *k.Spec.Realm != "" {
		realm = *k.Spec.Realm
	}

	// Create data map for ConfigMap
	data := make(map[string]string)
	data["client-id"] = clientID
	data["keycloak-url"] = config.KeycloakURL.String()
	data["issuer-url"] = config.KeycloakURL.JoinPath("realms", realm).String()

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: k.Namespace,
			Name:      name,
		},
		Data: data,
	}
}
