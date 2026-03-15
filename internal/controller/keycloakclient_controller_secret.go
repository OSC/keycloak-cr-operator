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
	"fmt"
	"time"

	keycloakv1alpha1 "github.com/OSC/keycloak-cr-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/Nerzal/gocloak/v13"
)

func usesClientSecret(keycloakClient *keycloakv1alpha1.KeycloakClient) bool {
	if keycloakClient.Spec.ClientAuthenticatorType != nil && *keycloakClient.Spec.ClientAuthenticatorType == clientSecretVal &&
		keycloakClient.Spec.PublicClient != nil && !*keycloakClient.Spec.PublicClient {
		return true
	} else {
		return false
	}
}

func (r *KeycloakClientReconciler) lookupSecret(keycloakClient *keycloakv1alpha1.KeycloakClient) bool {
	if keycloakClient.Spec.ClientSecretRef != nil &&
		keycloakClient.Spec.ClientSecretRef.Create != nil && !*keycloakClient.Spec.ClientSecretRef.Create &&
		usesClientSecret(keycloakClient) {
		return true
	} else {
		return false
	}
}

func (r *KeycloakClientReconciler) createSecret(keycloakClient *keycloakv1alpha1.KeycloakClient) bool {
	if keycloakClient.Spec.ClientSecretRef != nil &&
		keycloakClient.Spec.ClientSecretRef.Create != nil && *keycloakClient.Spec.ClientSecretRef.Create &&
		usesClientSecret(keycloakClient) {
		return true
	} else {
		return false
	}
}

func (r *KeycloakClientReconciler) getSecret(ctx context.Context, keycloakClient *keycloakv1alpha1.KeycloakClient) (string, error) {
	log := logf.FromContext(ctx)
	log.V(1).Info("Begin get secret", "namespace", keycloakClient.Namespace, "name", keycloakClient.Name)
	secret := &corev1.Secret{}
	log.V(1).Info("Get secret", "namespace", keycloakClient.Namespace, "name", keycloakClient.Name,
		"secret", keycloakClient.Spec.ClientSecretRef.Name, "key", keycloakClient.Spec.ClientSecretRef.Key)

	// Set up retry logic with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, *r.SecretWaitTimeout)
	defer cancel()

	// Retry until the secret is found or timeout occurs
	for {
		err := r.Get(ctxWithTimeout, types.NamespacedName{Name: keycloakClient.Spec.ClientSecretRef.Name, Namespace: keycloakClient.Namespace}, secret)
		if err == nil {
			// Secret found successfully
			clientSecret, found := secret.Data[keycloakClient.Spec.ClientSecretRef.Key]
			if !found {
				return "", fmt.Errorf("unable to find secret key %s in secret %s", keycloakClient.Spec.ClientSecretRef.Key, keycloakClient.Spec.ClientSecretRef.Name)
			}
			return string(clientSecret), nil
		}

		// Check if the error is "not found" - if so, retry
		if apierrors.IsNotFound(err) {
			// Check if we've timed out
			select {
			case <-ctxWithTimeout.Done():
				return "", fmt.Errorf("timed out waiting for secret %s to become available", keycloakClient.Spec.ClientSecretRef.Name)
			default:
				// Not timed out yet, continue with retry
				log.V(1).Info("Secret not found, retrying", "namespace", keycloakClient.Namespace, "secret", keycloakClient.Spec.ClientSecretRef.Name, "timeout", r.SecretWaitTimeout)
				time.Sleep(1 * time.Second) // Wait 1 second before retrying
				continue
			}
		} else {
			// Some other error occurred, return it
			return "", err
		}
	}
}

// handleSecret creates or updates the corev1.Secret resource for the KeycloakClient
func (r *KeycloakClientReconciler) handleSecret(ctx context.Context, keycloakClient *keycloakv1alpha1.KeycloakClient, gocloakClient *gocloak.Client) error {
	log := logf.FromContext(ctx)
	log.V(1).Info("Get Keycloak Client", "clientID", *keycloakClient.Spec.ClientID, "realm", *keycloakClient.Spec.Realm)
	client, err := GetKeycloakClient(ctx, r.Server, keycloakClient)
	if err != nil {
		log.Error(err, "Failed to get Keycloak client", "clientID", *keycloakClient.Spec.ClientID, "realm", *keycloakClient.Spec.Realm)
		return err
	}
	if client == nil {
		err = fmt.Errorf("no Keycloak client returned")
		log.Error(err, "Unable to get Keycloak Client", "clientID", *keycloakClient.Spec.ClientID, "realm", *keycloakClient.Spec.Realm)
		return err
	}
	gocloakClient.Secret = client.Secret

	// Get the secret using the GetSecret method from the KeycloakClient
	secret := keycloakClient.GetSecret(*gocloakClient.Secret)

	// Check if the secret already exists
	found := &corev1.Secret{}
	err = r.Get(ctx, types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}, found)
	if err != nil && apierrors.IsNotFound(err) {
		// Secret doesn't exist, create it
		log.Info("Creating a new Secret", "secret.Namespace", secret.Namespace, "secret.Name", secret.Name)
		err = ctrl.SetControllerReference(keycloakClient, secret, r.Scheme)
		if err != nil {
			log.Error(err, "Failed to set controller reference for secret", "secret.Namespace", secret.Namespace, "secret.Name", secret.Name)
			return err
		}
		err = r.Create(ctx, secret)
		if err != nil {
			log.Error(err, "Failed to create new Secret", "secret.Namespace", secret.Namespace, "secret.Name", secret.Name)
			return err
		}
		log.Info("Created a new Secret", "secret.Namespace", secret.Namespace, "secret.Name", secret.Name)
	} else if err != nil {
		// Some other error occurred
		log.Error(err, "Failed to get Secret", "secret.Namespace", secret.Namespace, "secret.Name", secret.Name)
		return err
	} else {
		// Secret exists, update it
		log.Info("Updating existing Secret", "secret.Namespace", secret.Namespace, "secret.Name", secret.Name)
		// Set the controller reference to ensure ownership
		err = ctrl.SetControllerReference(keycloakClient, secret, r.Scheme)
		if err != nil {
			log.Error(err, "Failed to set controller reference for secret", "secret.Namespace", secret.Namespace, "secret.Name", secret.Name)
			return err
		}
		// Use retry to handle potential conflicts
		err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return r.Update(ctx, secret)
		})
		if err != nil {
			log.Error(err, "Failed to update Secret", "secret.Namespace", secret.Namespace, "secret.Name", secret.Name)
			return err
		}
		log.Info("Updated existing Secret", "secret.Namespace", secret.Namespace, "secret.Name", secret.Name)
	}

	return nil
}
