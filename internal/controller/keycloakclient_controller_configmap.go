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

package controller

import (
	"context"

	keycloakv1alpha1 "github.com/OSC/keycloak-cr-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// handleConfigMap creates or updates the corev1.ConfigMap resource for the KeycloakClient
func (r *KeycloakClientReconciler) handleConfigMap(ctx context.Context, keycloakClient *keycloakv1alpha1.KeycloakClient) error {
	log := logf.FromContext(ctx)
	log.V(1).Info("Handle Keycloak Client ConfigMap", "namespace", keycloakClient.Namespace, "name", keycloakClient.Name)

	// Get the config map using the GetConfigMap method from the KeycloakClient
	configMap := keycloakClient.GetConfigMap(r.Config)

	// Check if the config map already exists
	found := &corev1.ConfigMap{}
	err := r.Get(ctx, types.NamespacedName{Name: configMap.Name, Namespace: configMap.Namespace}, found)
	if err != nil && apierrors.IsNotFound(err) {
		// ConfigMap doesn't exist, create it
		log.Info("Creating a new ConfigMap", "configMap.Namespace", configMap.Namespace, "configMap.Name", configMap.Name)

		// Set controller reference to ensure ownership
		err = ctrl.SetControllerReference(keycloakClient, configMap, r.Scheme)
		if err != nil {
			log.Error(err, "Failed to set controller reference for configMap", "configMap.Namespace", configMap.Namespace, "configMap.Name", configMap.Name)
			return err
		}

		err = r.Create(ctx, configMap)
		if err != nil {
			log.Error(err, "Failed to create new ConfigMap", "configMap.Namespace", configMap.Namespace, "configMap.Name", configMap.Name)
			r.Recorder.Eventf(keycloakClient, nil, corev1.EventTypeWarning, "CreateConfigMapFailed", "Create",
				"Failed to create the ConfigMap %s for KeycloakClient %s in namespace %s: %s",
				configMap.Name, keycloakClient.Name, keycloakClient.Namespace, err)
			return err
		}
		log.Info("Created a new ConfigMap", "configMap.Namespace", configMap.Namespace, "configMap.Name", configMap.Name)
	} else if err != nil {
		// Some other error occurred
		log.Error(err, "Failed to get ConfigMap", "configMap.Namespace", configMap.Namespace, "configMap.Name", configMap.Name)
		return err
	} else {
		// ConfigMap exists, update it
		log.Info("Updating existing ConfigMap", "configMap.Namespace", configMap.Namespace, "configMap.Name", configMap.Name)

		// Set the controller reference to ensure ownership
		err = ctrl.SetControllerReference(keycloakClient, configMap, r.Scheme)
		if err != nil {
			log.Error(err, "Failed to set controller reference for configMap", "configMap.Namespace", configMap.Namespace, "configMap.Name", configMap.Name)
			return err
		}

		// Update the config map with new data
		err = r.Update(ctx, configMap)
		if err != nil {
			log.Error(err, "Failed to update ConfigMap", "configMap.Namespace", configMap.Namespace, "configMap.Name", configMap.Name)
			r.Recorder.Eventf(keycloakClient, nil, corev1.EventTypeWarning, "UpdateConfigMapFailed", "Update",
				"Failed to update the ConfigMap %s for KeycloakClient %s in namespace %s: %s",
				configMap.Name, keycloakClient.Name, keycloakClient.Namespace, err)
			return err
		}
		log.Info("Updated existing ConfigMap", "configMap.Namespace", configMap.Namespace, "configMap.Name", configMap.Name)
	}

	return nil
}
