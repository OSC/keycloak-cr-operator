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

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	keycloakv1alpha1 "github.com/OSC/keycloak-cr-operator/api/v1alpha1"

	"github.com/Nerzal/gocloak/v13"
)

// KeycloakClientReconciler reconciles a KeycloakClient object
type KeycloakClientReconciler struct {
	client.Client
	Scheme                *runtime.Scheme
	Server                *gocloak.GoCloak
	KeycloakAdminUsername string
	KeycloakAdminPassword string
	KeycloakAdminRealm    string
	DefaultRealm          string
	ClientIDPrefix        string
	Token                 *KeycloakToken
}

// +kubebuilder:rbac:groups=keycloak.osc.edu,resources=keycloakclients,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=keycloak.osc.edu,resources=keycloakclients/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=keycloak.osc.edu,resources=keycloakclients/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the KeycloakClient object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.23.1/pkg/reconcile
func (r *KeycloakClientReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Fetch the KeycloakClient instance
	keycloakClient := &keycloakv1alpha1.KeycloakClient{}
	err := r.Get(ctx, req.NamespacedName, keycloakClient)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("KeycloakClient resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get KeycloakClient")
		return ctrl.Result{}, err
	}

	// Get the client object from the KeycloakClient resource
	secret, err := r.getSecret(ctx, keycloakClient)
	if err != nil {
		log.Error(err, "Unable to get secret")
	}
	gocloakClient := keycloakClient.GetClient(r.ClientIDPrefix, secret)

	// Check if the client exists in Keycloak and create/update if needed
	err = r.ensureKeycloakClient(ctx, keycloakClient, gocloakClient)
	if err != nil {
		log.Error(err, "Failed to ensure Keycloak client")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// ensureKeycloakClient checks if the client exists in Keycloak and creates/updates it if needed
func (r *KeycloakClientReconciler) ensureKeycloakClient(ctx context.Context, keycloakClient *keycloakv1alpha1.KeycloakClient, gocloakClient *gocloak.Client) error {
	log := logf.FromContext(ctx)

	// Get realm name - use the default realm if not specified
	realm := r.DefaultRealm

	// Get an access token first
	token, err := r.Server.LoginAdmin(ctx, r.KeycloakAdminRealm, r.KeycloakAdminUsername, r.KeycloakAdminPassword)
	if err != nil {
		return err
	}

	// Check if the client already exists in Keycloak by using GetClients with ClientID param
	getClientParams := gocloak.GetClientsParams{
		ClientID: gocloakClient.ClientID,
	}
	clients, err := r.Server.GetClients(ctx, token.AccessToken, realm, getClientParams)
	if err != nil {
		log.Error(err, "Failed to get Keycloak Clients", "clientID", *gocloakClient.ClientID, "realm", realm)
		return err
	}

	if len(clients) < 1 {
		log.Info("Keycloak client not found, creating new one", "clientID", *gocloakClient.ClientID, "realm", realm)
		_, err = r.Server.CreateClient(ctx, token.AccessToken, realm, *gocloakClient)
		if err != nil {
			log.Error(err, "Failed to create Keycloak client", "clientID", *gocloakClient.ClientID, "realm", realm)
			return err
		}
		log.Info("Successfully created Keycloak client", "clientID", *gocloakClient.ClientID, "realm", realm)
	} else {
		log.Info("Keycloak client already exists, updating", "clientID", *gocloakClient.ClientID, "realm", realm)
		err = r.Server.UpdateClient(ctx, token.AccessToken, realm, *gocloakClient)
		if err != nil {
			log.Error(err, "Failed to update Keycloak client", "clientID", *gocloakClient.ClientID, "realm", realm)
			return err
		}
		log.Info("Successfully updated Keycloak client", "clientID", *gocloakClient.ClientID, "realm", realm)
	}

	meta.SetStatusCondition(&keycloakClient.Status.Conditions, metav1.Condition{
		Type:    "Available",
		Status:  metav1.ConditionTrue,
		Reason:  "Successful",
		Message: "Keycloak client processed successfully",
	})

	err = r.Status().Update(ctx, keycloakClient)
	if err != nil {
		log.Error(err, "Failed to update KeycloakClient status")
		return err
	}

	return nil
}

func (r *KeycloakClientReconciler) getSecret(ctx context.Context, keycloakClient *keycloakv1alpha1.KeycloakClient) (string, error) {
	if keycloakClient.Spec.ClientSecretRef == nil {
		return "", nil
	}
	secretName := keycloakClient.Spec.ClientSecretRef.Name
	secretKey := keycloakClient.Spec.ClientSecretRef.Key
	secret := &corev1.Secret{}
	err := r.Client.Get(ctx, types.NamespacedName{Name: secretName, Namespace: keycloakClient.Namespace}, secret)
	if err != nil {
		return "", err
	}
	clientSecret, found := secret.Data[secretKey]
	if !found {
		return "", fmt.Errorf("Unable to find secret key %s in secret %s", secretKey, secretName)
	}
	return string(clientSecret), nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KeycloakClientReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&keycloakv1alpha1.KeycloakClient{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
