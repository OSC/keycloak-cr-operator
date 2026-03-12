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

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	keycloakv1alpha1 "github.com/OSC/keycloak-cr-operator/api/v1alpha1"

	"github.com/Nerzal/gocloak/v13"
)

const (
	clientFinalizerName         = "client.keycloak.osc.edu/finalizer"
	typeAvailableKeycloakClient = "Available"
)

type GoCloakServer interface {
	LoginAdmin(ctx context.Context, realm, username, password string) (*gocloak.JWT, error)
	GetClients(ctx context.Context, token, realm string, params gocloak.GetClientsParams) ([]*gocloak.Client, error)
	CreateClient(ctx context.Context, token, realm string, client gocloak.Client) (string, error)
	UpdateClient(ctx context.Context, token, realm string, client gocloak.Client) error
	DeleteClient(ctx context.Context, token, realm, idOfClient string) error
}

// KeycloakClientReconciler reconciles a KeycloakClient object
type KeycloakClientReconciler struct {
	runtimeclient.Client
	Scheme            *runtime.Scheme
	Server            GoCloakServer
	SecretWaitTimeout *time.Duration
	Config            *KeycloakConfig
}

// +kubebuilder:rbac:groups=keycloak.osc.edu,resources=keycloakclients,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=keycloak.osc.edu,resources=keycloakclients/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=keycloak.osc.edu,resources=keycloakclients/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.23.1/pkg/reconcile
func (r *KeycloakClientReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Fetch the KeycloakClient instance
	keycloakClient := &keycloakv1alpha1.KeycloakClient{}
	log.Info("Received reconcile for KeycloakClient", "namespace", req.Namespace, "name", req.Name)
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

	// Let's just set the status as Unknown when no status is available
	if len(keycloakClient.Status.Conditions) == 0 {
		err := r.setStatus(ctx, keycloakClient, metav1.Condition{
			Type:    typeAvailableKeycloakClient,
			Status:  metav1.ConditionUnknown,
			Reason:  "Reconciling",
			Message: "Starting reconciliation",
		})
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	var secret string
	if keycloakClient.Spec.ClientSecretRef == nil {
		log.V(1).Info("Secret not defined", "namespace", keycloakClient.Namespace, "name", keycloakClient.Name)
	} else {
		secret, err = r.getSecret(ctx, keycloakClient)
		if err != nil {
			_ = r.setStatus(ctx, keycloakClient, metav1.Condition{
				Type:    typeAvailableKeycloakClient,
				Status:  metav1.ConditionFalse,
				Reason:  "Failed",
				Message: fmt.Sprintf("Unable to get secret %s", keycloakClient.Spec.ClientSecretRef.Name),
			})
			log.Error(err, "Unable to get secret")
			return ctrl.Result{}, err
		}
	}
	log.V(1).Info("Get gocloak Client", "namespace", keycloakClient.Namespace, "name", keycloakClient.Name)
	gocloakClient := keycloakClient.GetClient(r.Config.ClientIDPrefix, secret)

	delete, err := r.handleFinalizer(ctx, keycloakClient, gocloakClient)
	if err != nil {
		log.Error(err, "failed to handle finalizer")
		_ = r.setStatus(ctx, keycloakClient, metav1.Condition{
			Type:    typeAvailableKeycloakClient,
			Status:  metav1.ConditionFalse,
			Reason:  "Failed",
			Message: "failed to handle finalizer",
		})
		return ctrl.Result{}, err
	}
	if delete {
		return ctrl.Result{}, nil
	}
	// Let's re-fetch the Custom Resource after updating the status
	// so that we have the latest state of the resource on the cluster and we will avoid
	// raising the error "the object has been modified, please apply
	// your changes to the latest version and try again" which would re-trigger the reconciliation
	// if we try to update it again in the following operations
	if err := r.Get(ctx, req.NamespacedName, keycloakClient); err != nil {
		log.Error(err, "Failed to re-fetch keycloakClient")
		return ctrl.Result{}, err
	}

	// Check if the client exists in Keycloak and create/update if needed
	err = r.ensureKeycloakClient(ctx, keycloakClient, gocloakClient)
	if err != nil {
		log.Error(err, "Failed to ensure Keycloak client")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *KeycloakClientReconciler) handleFinalizer(ctx context.Context, keycloakClient *keycloakv1alpha1.KeycloakClient, gocloakClient *gocloak.Client) (bool, error) {
	log := logf.FromContext(ctx)
	if keycloakClient.DeletionTimestamp.IsZero() {
		// add finalizer in case of create/update
		if !controllerutil.ContainsFinalizer(keycloakClient, clientFinalizerName) {
			ok := controllerutil.AddFinalizer(keycloakClient, clientFinalizerName)
			log.Info("Add Finalizer", "name", clientFinalizerName, "ok", ok)
			return false, retry.RetryOnConflict(retry.DefaultRetry, func() error {
				return r.Update(ctx, keycloakClient)
			})
		}
	} else {
		// remove finalizer in case of deletion
		if controllerutil.ContainsFinalizer(keycloakClient, clientFinalizerName) {
			if err := r.deleteKeycloakClient(ctx, keycloakClient, gocloakClient); err != nil {
				return true, err
			}
			ok := controllerutil.RemoveFinalizer(keycloakClient, clientFinalizerName)
			log.Info("Remove Finalizer", "name", clientFinalizerName, "ok", ok)
			return true, retry.RetryOnConflict(retry.DefaultRetry, func() error {
				return r.Update(ctx, keycloakClient)
			})
		}
	}
	return false, nil
}

func (r *KeycloakClientReconciler) setStatus(ctx context.Context, keycloakClient *keycloakv1alpha1.KeycloakClient, condition metav1.Condition) error {
	log := logf.FromContext(ctx)
	namespacedName := types.NamespacedName{Name: keycloakClient.Name, Namespace: keycloakClient.Namespace}
	// Re-fetch object to avoid "the object has been modified" errors
	if err := r.Get(ctx, namespacedName, keycloakClient); err != nil {
		log.Error(err, "Failed to re-fetch keycloakClient")
		return err
	}
	log.V(1).Info("Set KeycloakClient status", "namespace", keycloakClient.Namespace, "name", keycloakClient.Name,
		"status", condition.Status, "message", condition.Message)
	meta.SetStatusCondition(&keycloakClient.Status.Conditions, condition)
	log.V(1).Info("Updating KeycloakClient with status", "namespace", keycloakClient.Namespace, "name", keycloakClient.Name)
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return r.Status().Update(ctx, keycloakClient)
	})
	if err != nil {
		log.Error(err, "Failed to update KeycloakClient status")
		return err
	}
	return nil
}

// ensureKeycloakClient checks if the client exists in Keycloak and creates/updates it if needed
func (r *KeycloakClientReconciler) ensureKeycloakClient(ctx context.Context, keycloakClient *keycloakv1alpha1.KeycloakClient, gocloakClient *gocloak.Client) error {
	log := logf.FromContext(ctx)
	realm := r.Config.DefaultRealm
	if keycloakClient.Spec.Realm != nil && *keycloakClient.Spec.Realm != "" {
		realm = *keycloakClient.Spec.Realm
	}
	log.V(1).Info("Ensure Keycloak Client", "namespace", keycloakClient.Namespace, "name", keycloakClient.Name, "clientID", gocloakClient.ClientID, "realm", realm)

	// Get an access token first
	log.V(1).Info("Keycloak Login", "namespace", keycloakClient.Namespace, "name", keycloakClient.Name,
		"clientID", gocloakClient.ClientID, "realm", realm,
		"admin-realm", r.Config.AdminRealm, "admin-username", r.Config.AdminUsername)
	err := KeycloakLogin(ctx, r.Server, r.Config)
	if err != nil {
		return err
	}

	// Check if the client already exists in Keycloak by using GetClients with ClientID param
	getClientParams := gocloak.GetClientsParams{
		ClientID: gocloakClient.ClientID,
	}
	log.V(1).Info("Check if client exists", "namespace", keycloakClient.Namespace, "name", keycloakClient.Name, "clientID", gocloakClient.ClientID, "realm", realm)
	Token.lock.RLock()
	defer Token.lock.RUnlock()
	clients, err := r.Server.GetClients(ctx, Token.AccessToken, realm, getClientParams)
	if err != nil {
		log.Error(err, "Failed to get Keycloak Clients", "clientID", *gocloakClient.ClientID, "realm", realm)
		return err
	}
	log.V(1).Info(fmt.Sprintf("Number of clients returned: %d", len(clients)), "namespace", keycloakClient.Namespace, "name", keycloakClient.Name, "clientID", gocloakClient.ClientID, "realm", realm)
	if len(clients) < 1 {
		log.Info("Keycloak client not found, creating new one", "clientID", *gocloakClient.ClientID, "realm", realm)
		_, err = r.Server.CreateClient(ctx, Token.AccessToken, realm, *gocloakClient)
		if err != nil {
			log.Error(err, "Failed to create Keycloak client", "clientID", *gocloakClient.ClientID, "realm", realm)
			_ = r.setStatus(ctx, keycloakClient, metav1.Condition{
				Type:    typeAvailableKeycloakClient,
				Status:  metav1.ConditionFalse,
				Reason:  "Failed",
				Message: "Failed to create Keycloak client",
			})
			return err
		}
		log.Info("Successfully created Keycloak client", "clientID", *gocloakClient.ClientID, "realm", realm)
	} else {
		log.Info("Keycloak client already exists, updating", "clientID", *gocloakClient.ClientID, "realm", realm)
		err = r.Server.UpdateClient(ctx, Token.AccessToken, realm, *gocloakClient)
		if err != nil {
			log.Error(err, "Failed to update Keycloak client", "clientID", *gocloakClient.ClientID, "realm", realm)
			_ = r.setStatus(ctx, keycloakClient, metav1.Condition{
				Type:    typeAvailableKeycloakClient,
				Status:  metav1.ConditionFalse,
				Reason:  "Failed",
				Message: "Failed to update Keycloak client",
			})
			return err
		}
		log.Info("Successfully updated Keycloak client", "clientID", *gocloakClient.ClientID, "realm", realm)
	}

	err = r.setStatus(ctx, keycloakClient, metav1.Condition{
		Type:    typeAvailableKeycloakClient,
		Status:  metav1.ConditionTrue,
		Reason:  "Successful",
		Message: "Keycloak client processed successfully",
	})
	if err != nil {
		return err
	}

	return nil
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

func (r *KeycloakClientReconciler) deleteKeycloakClient(ctx context.Context, keycloakClient *keycloakv1alpha1.KeycloakClient, gocloakClient *gocloak.Client) error {
	log := logf.FromContext(ctx)
	realm := r.Config.DefaultRealm
	if keycloakClient.Spec.Realm != nil && *keycloakClient.Spec.Realm != "" {
		realm = *keycloakClient.Spec.Realm
	}

	log.V(1).Info("Keycloak Login", "namespace", keycloakClient.Namespace, "name", keycloakClient.Name,
		"clientID", gocloakClient.ClientID, "realm", realm,
		"admin-realm", r.Config.AdminRealm, "admin-username", r.Config.AdminUsername)
	err := KeycloakLogin(ctx, r.Server, r.Config)
	if err != nil {
		log.Error(err, "Failed to get Keycloak admin token for deletion")
		return err
	}

	log.Info("Deleting Keycloak client", "clientID", *gocloakClient.ClientID, "realm", realm)
	Token.lock.RLock()
	defer Token.lock.RUnlock()
	err = r.Server.DeleteClient(ctx, Token.AccessToken, realm, *gocloakClient.ID)
	if err != nil {
		log.Error(err, "Failed to delete Keycloak client", "clientID", *gocloakClient.ClientID, "realm", realm)
		return err
	}
	log.Info("Successfully deleted Keycloak client", "clientID", *gocloakClient.ClientID, "realm", realm)
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KeycloakClientReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&keycloakv1alpha1.KeycloakClient{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}
