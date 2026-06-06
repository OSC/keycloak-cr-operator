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
	"bytes"
	"fmt"

	"github.com/OSC/keycloak-cr-operator/internal/models"
	"github.com/stoewer/go-strcase"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetSecret creates a corev1.Secret object using data from ClientSecretRef for name and key
// and KeycloakClient for namespace. It sets data based on the secret argument to StringData.
func (k *KeycloakClient) GetSecret(config *models.KeycloakConfig, clientSecret string) (*corev1.Secret, error) {
	var defKey, key, name, clientID string
	var envVarKeys bool
	var keyPrefix string
	if k.Spec.ClientSecretRef != nil {
		name = k.Spec.ClientSecretRef.Name
		defKey = k.Spec.ClientSecretRef.Key
		if k.Spec.ClientSecretRef.EnvVarKeys == nil {
			envVarKeys = true
		} else {
			envVarKeys = *k.Spec.ClientSecretRef.EnvVarKeys
		}
		if k.Spec.ClientSecretRef.KeyPrefix != nil {
			keyPrefix = *k.Spec.ClientSecretRef.KeyPrefix
		}
	} else {
		name = fmt.Sprintf("%s-secret", k.Name)
		envVarKeys = true
		defKey = "CLIENT_SECRET"
	}
	if config.ClientIDRequired != nil && (k.Spec.ClientID == nil || *k.Spec.ClientID == "") {
		requiredClientID, err := RequiredClientID(config, k)
		if err != nil {
			return nil, err
		}
		clientID = requiredClientID
	} else if k.Spec.ClientID == nil || *k.Spec.ClientID == "" {
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
	issuerUrl := config.KeycloakURL.JoinPath("realms", realm)

	data := make(map[string]string)
	if envVarKeys {
		data[fmt.Sprintf("%sCLIENT_ID", keyPrefix)] = clientID
		key = strcase.UpperSnakeCase(defKey)
		data[fmt.Sprintf("%sISSUER_URL", keyPrefix)] = issuerUrl.String()
	} else {
		data[fmt.Sprintf("%sclient-id", keyPrefix)] = clientID
		key = defKey
		data[fmt.Sprintf("%sissuer-url", keyPrefix)] = issuerUrl.String()
	}
	data[fmt.Sprintf("%s%s", keyPrefix, key)] = clientSecret

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: k.Namespace,
			Name:      name,
		},
		StringData: data,
	}, nil
}

func RequiredClientID(config *models.KeycloakConfig, obj *KeycloakClient) (string, error) {
	// Apply ClientIDRequired template if set
	if config.ClientIDRequired != nil {
		data := KeycloakClientData{
			Obj:    *obj,
			Config: *config,
		}

		var buf bytes.Buffer
		err := config.ClientIDRequired.Execute(&buf, data)
		if err != nil {
			return "", err
		}
		return buf.String(), nil
	}
	return "", nil
}
