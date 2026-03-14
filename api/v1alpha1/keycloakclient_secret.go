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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetSecret creates a corev1.Secret object using data from ClientSecretRef for name and key
// and KeycloakClient for namespace. It sets data based on the secret argument to StringData.
func (k *KeycloakClient) GetSecret(clientSecret string) *corev1.Secret {
	var key, name string
	if k.Spec.ClientSecretRef != nil {
		name = k.Spec.ClientSecretRef.Name
		key = k.Spec.ClientSecretRef.Key
	} else {
		name = fmt.Sprintf("%s-secret", k.Name)
		key = "client-secret"
	}
	data := make(map[string]string)
	data[key] = clientSecret

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: k.Namespace,
			Name:      name,
		},
		StringData: data,
	}
}
