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
	"k8s.io/client-go/tools/events"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	keycloakv1alpha1 "github.com/OSC/keycloak-cr-operator/api/v1alpha1"
	"github.com/OSC/keycloak-cr-operator/internal/models"

	"github.com/Nerzal/gocloak/v13"
)

const (
	clientFinalizerName         = "client.keycloak.osc.edu/finalizer"
	typeAvailableKeycloakClient = "Available"
	keycloakClientMatchLabel    = "keycloak.osc.edu/keycloakclient"
	clientSecretVal             = "client-secret"
)

type GoCloakServer interface {
	LoginAdmin(ctx context.Context, realm, username, password string) (*gocloak.JWT, error)
	GetClients(ctx context.Context, token, realm string, params gocloak.GetClientsParams) ([]*gocloak.Client, error)
	GetClientSecret(ctx context.Context, token, realm, idOfClient string) (*gocloak.CredentialRepresentation, error)
	CreateClient(ctx context.Context, token, realm string, client gocloak.Client) (string, error)
	UpdateClient(ctx context.Context, token, realm string, client gocloak.Client) error
	DeleteClient(ctx context.Context, token, realm, idOfClient string) error
}

// KeycloakClientReconciler reconciles a KeycloakClient object
type KeycloakClientReconciler struct {
	runtimeclient.Client
	Scheme            *runtime.Scheme
	Recorder          events.EventRecorder
	Server            GoCloakServer
	SecretWaitTimeout *time.Duration
	Config            *models.KeycloakConfig
}

// +kubebuilder:rbac:groups=keycloak.osc.edu,resources=keycloakclients,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=keycloak.osc.edu,resources=keycloakclients/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=keycloak.osc.edu,resources=keycloakclients/finalizers,verbs=update
// +kubebuilder:rbac:groups=events.k8s.io,resources=events,verbs=create;patch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;update;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;update;patch

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

	delete, err := r.handleFinalizer(ctx, keycloakClient)
	if err != nil {
		log.Error(err, "failed to handle finalizer")
		_ = r.setStatus(ctx, keycloakClient, metav1.Condition{
			Type:    typeAvailableKeycloakClient,
			Status:  metav1.ConditionFalse,
			Reason:  "Failed",
			Message: "failed to handle finalizer",
		})
		r.Recorder.Eventf(keycloakClient, nil, corev1.EventTypeWarning, "HandleFinalizerFailed", "Handle",
			"Failed to handle the finalizer for KeycloakClient %s in namespace %s: %s",
			keycloakClient.Name, keycloakClient.Namespace, err)
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

	// Get the gocloak Client struct based on KeycloakClient spec
	log.V(1).Info("Get gocloak Client", "namespace", keycloakClient.Namespace, "name", keycloakClient.Name)
	gocloakClient, err := keycloakClient.GetClient(r.Config)
	if err != nil {
		log.Error(err, "Failed to get gocloak client")
		_ = r.setStatus(ctx, keycloakClient, metav1.Condition{
			Type:    typeAvailableKeycloakClient,
			Status:  metav1.ConditionFalse,
			Reason:  "Failed",
			Message: fmt.Sprintf("Failed to gocloak client: %s", err),
		})
		return ctrl.Result{}, err
	}

	// Get the ClientSecret from clientSecretRef, if set
	if keycloakClient.Spec.ClientSecretRef == nil {
		log.V(1).Info("Secret not defined", "namespace", keycloakClient.Namespace, "name", keycloakClient.Name)
	} else if shouldLookupSecret(keycloakClient) {
		secret, err := r.getSecret(ctx, keycloakClient)
		if err != nil {
			_ = r.setStatus(ctx, keycloakClient, metav1.Condition{
				Type:    typeAvailableKeycloakClient,
				Status:  metav1.ConditionFalse,
				Reason:  "Failed",
				Message: fmt.Sprintf("Unable to get secret %s", keycloakClient.Spec.ClientSecretRef.Name),
			})
			log.Error(err, "Unable to get secret")
			r.Recorder.Eventf(keycloakClient, nil, corev1.EventTypeWarning, "GetSecretFailed", "Get",
				"Failed to get the Secret for KeycloakClient %s in namespace %s: %s",
				keycloakClient.Name, keycloakClient.Namespace, err)
			return ctrl.Result{}, err
		}
		gocloakClient.Secret = &secret
	}

	// Check if the client exists in Keycloak and create/update if needed
	err = r.ensureKeycloakClient(ctx, keycloakClient, gocloakClient)
	if err != nil {
		log.Error(err, "Failed to ensure Keycloak client")
		return ctrl.Result{}, err
	}

	// Handle the secret creation/update if needed
	if shouldCreateSecret(keycloakClient) {
		err = r.handleSecret(ctx, keycloakClient, gocloakClient)
		if err != nil {
			log.Error(err, "Failed to handle secret")
			_ = r.setStatus(ctx, keycloakClient, metav1.Condition{
				Type:    typeAvailableKeycloakClient,
				Status:  metav1.ConditionFalse,
				Reason:  "Failed",
				Message: fmt.Sprintf("Failed to create Keycloak client secret: %s", err),
			})
			return ctrl.Result{}, err
		}
	}

	// Handle the config map creation/update
	err = r.handleConfigMap(ctx, keycloakClient)
	if err != nil {
		log.Error(err, "Failed to handle config map")
		_ = r.setStatus(ctx, keycloakClient, metav1.Condition{
			Type:    typeAvailableKeycloakClient,
			Status:  metav1.ConditionFalse,
			Reason:  "Failed",
			Message: fmt.Sprintf("Failed to create Keycloak client config map: %s", err),
		})
		return ctrl.Result{}, err
	}

	if err = r.setStatus(ctx, keycloakClient, metav1.Condition{
		Type:    typeAvailableKeycloakClient,
		Status:  metav1.ConditionTrue,
		Reason:  "Successful",
		Message: "Keycloak client processed successfully",
	}); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *KeycloakClientReconciler) handleFinalizer(ctx context.Context, keycloakClient *keycloakv1alpha1.KeycloakClient) (bool, error) {
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
			if err := r.deleteKeycloakClient(ctx, keycloakClient); err != nil {
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

	log.V(1).Info("Ensure Keycloak Client", "namespace", keycloakClient.Namespace, "name", keycloakClient.Name, "clientID", *gocloakClient.ClientID, "realm", *keycloakClient.Spec.Realm)

	if *keycloakClient.Spec.Realm == "" {
		return fmt.Errorf("realm is not defined for KeycloakClient %s", keycloakClient.Name)
	}

	// Get an access token first
	log.V(1).Info("Keycloak Login", "namespace", keycloakClient.Namespace, "name", keycloakClient.Name,
		"clientID", *gocloakClient.ClientID, "realm", *keycloakClient.Spec.Realm,
		"admin-realm", r.Config.AdminRealm, "admin-username", r.Config.AdminUsername)
	err := KeycloakLogin(ctx, r.Server, r.Config)
	if err != nil {
		return err
	}

	client, err := GetKeycloakClient(ctx, r.Server, keycloakClient)
	if err != nil {
		return err
	}
	var id string
	if client == nil {
		log.Info("Keycloak client not found, creating new one", "clientID", *keycloakClient.Spec.ClientID, "realm", *keycloakClient.Spec.Realm)
		id, err = r.Server.CreateClient(ctx, Token.AccessToken, *keycloakClient.Spec.Realm, *gocloakClient)
		if err != nil {
			log.Error(err, "Failed to create Keycloak client", "clientID", *keycloakClient.Spec.ClientID, "realm", *keycloakClient.Spec.Realm)
			_ = r.setStatus(ctx, keycloakClient, metav1.Condition{
				Type:    typeAvailableKeycloakClient,
				Status:  metav1.ConditionFalse,
				Reason:  "Failed",
				Message: "Failed to create Keycloak client",
			})
			r.Recorder.Eventf(keycloakClient, nil, corev1.EventTypeWarning, "CreateKeycloakClientFailed", "Create",
				"Failed to create KeycloakClient %s in namespace %s: %s",
				keycloakClient.Name, keycloakClient.Namespace, err)
			return err
		}
		log.Info("Successfully created Keycloak client", "clientID", *keycloakClient.Spec.ClientID, "realm", *keycloakClient.Spec.Realm)
	} else {
		log.Info("Keycloak client already exists, updating", "clientID", *keycloakClient.Spec.ClientID, "realm", *keycloakClient.Spec.Realm)
		gocloakClient.ID = client.ID
		err = r.Server.UpdateClient(ctx, Token.AccessToken, *keycloakClient.Spec.Realm, *gocloakClient)
		if err != nil {
			log.Error(err, "Failed to update Keycloak client", "clientID", *keycloakClient.Spec.ClientID, "realm", *keycloakClient.Spec.Realm)
			_ = r.setStatus(ctx, keycloakClient, metav1.Condition{
				Type:    typeAvailableKeycloakClient,
				Status:  metav1.ConditionFalse,
				Reason:  "Failed",
				Message: "Failed to update Keycloak client",
			})
			r.Recorder.Eventf(keycloakClient, nil, corev1.EventTypeWarning, "UpdateKeycloakClientFailed", "Update",
				"Failed to update KeycloakClient %s in namespace %s: %s",
				keycloakClient.Name, keycloakClient.Namespace, err)
			return err
		}
		id = *client.ID
		log.Info("Successfully updated Keycloak client", "clientID", *keycloakClient.Spec.ClientID, "realm", *keycloakClient.Spec.Realm)
	}

	// Re-fetch object to avoid "the object has been modified" errors
	if err := r.Get(ctx, types.NamespacedName{Name: keycloakClient.Name, Namespace: keycloakClient.Namespace}, keycloakClient); err != nil {
		log.Error(err, "Failed to re-fetch keycloakClient")
		return err
	}
	keycloakClient.Status.ID = &id
	log.V(1).Info("Updating KeycloakClient with ID status", "namespace", keycloakClient.Namespace, "name", keycloakClient.Name)
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return r.Status().Update(ctx, keycloakClient)
	})
	if err != nil {
		log.Error(err, "Failed to update KeycloakClient status")
		return err
	}
	return nil
}

func (r *KeycloakClientReconciler) deleteKeycloakClient(ctx context.Context, keycloakClient *keycloakv1alpha1.KeycloakClient) error {
	log := logf.FromContext(ctx)

	if *keycloakClient.Spec.ClientID == "" {
		return fmt.Errorf("clientID is not defined for KeycloakClient %s", keycloakClient.Name)
	}
	if *keycloakClient.Spec.Realm == "" {
		return fmt.Errorf("realm is not defined for KeycloakClient %s", keycloakClient.Name)
	}

	log.V(1).Info("Keycloak Login", "namespace", keycloakClient.Namespace, "name", keycloakClient.Name,
		"clientID", *keycloakClient.Spec.ClientID, "realm", *keycloakClient.Spec.Realm,
		"admin-realm", r.Config.AdminRealm, "admin-username", r.Config.AdminUsername)
	err := KeycloakLogin(ctx, r.Server, r.Config)
	if err != nil {
		log.Error(err, "Failed to get Keycloak admin token for deletion")
		return err
	}

	gocloakClient, err := GetKeycloakClient(ctx, r.Server, keycloakClient)
	if err != nil {
		return err
	}
	if gocloakClient == nil {
		log.Info("Keycloak Client already deleted", "clientID", *keycloakClient.Spec.ClientID, "realm", *keycloakClient.Spec.Realm)
		return nil
	}

	log.Info("Deleting Keycloak client", "id", *gocloakClient.ID, "clientID", *keycloakClient.Spec.ClientID, "realm", *keycloakClient.Spec.Realm)
	Token.lock.RLock()
	defer Token.lock.RUnlock()
	err = r.Server.DeleteClient(ctx, Token.AccessToken, *keycloakClient.Spec.Realm, *gocloakClient.ID)
	if err != nil {
		log.Error(err, "Failed to delete Keycloak client", "id", *gocloakClient.ID, "clientID", *keycloakClient.Spec.ClientID, "realm", *keycloakClient.Spec.Realm)
		r.Recorder.Eventf(keycloakClient, nil, corev1.EventTypeWarning, "DeleteKeycloakClientFailed", "Delete",
			"Failed to delete KeycloakClient %s in namespace %s: %s",
			keycloakClient.Name, keycloakClient.Namespace, err)
		return err
	}
	log.Info("Successfully deleted Keycloak client", "id", *gocloakClient.ID, "clientID", *keycloakClient.Spec.ClientID, "realm", *keycloakClient.Spec.Realm)
	return nil
}

func mapSecretToKeycloakClient(ctx context.Context, obj runtimeclient.Object) []reconcile.Request {
	log := logf.FromContext(ctx)
	log.V(1).Info("Entered manager check for secret")
	secret, ok := obj.(*corev1.Secret)
	if !ok {
		log.V(1).Info("Return manager secret check, not a secret")
		return []reconcile.Request{}
	}
	labels := secret.GetLabels()
	if labels == nil {
		log.V(1).Info("Return manager secret check, no labels")
		return []reconcile.Request{}
	}
	keycloakClientName, exists := labels[keycloakClientMatchLabel]
	if !exists {
		log.V(1).Info("Return manager secret check, keycloakclient secret label missing")
		return []reconcile.Request{}
	}
	log.V(1).Info("Trigger KeycloakClient reconcile from secret",
		"keycloakclient", keycloakClientName, "secret", secret.Name, "namespace", secret.Namespace)
	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      keycloakClientName,
				Namespace: secret.Namespace,
			},
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *KeycloakClientReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&keycloakv1alpha1.KeycloakClient{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Owns(&corev1.Secret{}).
		Owns(&corev1.ConfigMap{}).
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(mapSecretToKeycloakClient),
		).
		Complete(r)
}
