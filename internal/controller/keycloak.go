package controller

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/OSC/keycloak-cr-operator/api/v1alpha1"
	"github.com/OSC/keycloak-cr-operator/internal/models"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	Token = &KeycloakToken{}
)

type KeycloakToken struct {
	gocloak.JWT
	lock      sync.RWMutex
	CreatedAt *time.Time
}

func KeycloakLogin(ctx context.Context, server GoCloakServer, config *models.KeycloakConfig) error {
	log := logf.FromContext(ctx)
	Token.lock.Lock()
	defer Token.lock.Unlock()
	if Token != nil && Token.ExpiresIn != 0 && Token.CreatedAt != nil {
		expirationTime := Token.CreatedAt.Add(time.Duration(Token.ExpiresIn) * time.Second)
		if time.Now().Before(expirationTime) {
			log.V(1).Info("admin token still valid", "expiresIn", Token.ExpiresIn, "createdAt", Token.CreatedAt)
			return nil
		}
	}

	now := time.Now()
	t, err := server.LoginAdmin(ctx, config.AdminUsername, config.AdminPassword, config.AdminRealm)
	if err != nil {
		log.Error(err, "Failed to login to Keycloak", "username", config.AdminUsername, "realm", config.AdminRealm)
		return err
	}
	log.Info("Successfully logged into Keycloak", "username", config.AdminUsername, "realm", config.AdminRealm)
	Token.AccessToken = t.AccessToken
	Token.IDToken = t.IDToken
	Token.ExpiresIn = t.ExpiresIn
	Token.CreatedAt = &now

	return nil
}

func GetKeycloakClient(ctx context.Context, server GoCloakServer, keycloakClient *v1alpha1.KeycloakClient) (*gocloak.Client, error) {
	log := logf.FromContext(ctx)
	getClientParams := gocloak.GetClientsParams{
		ClientID: keycloakClient.Spec.ClientID,
	}
	log.V(1).Info("Check if client exists", "namespace", keycloakClient.Namespace, "name", keycloakClient.Name, "clientID", keycloakClient.Spec.ClientID, "realm", keycloakClient.Spec.Realm)
	Token.lock.RLock()
	defer Token.lock.RUnlock()
	clients, err := server.GetClients(ctx, Token.AccessToken, *keycloakClient.Spec.Realm, getClientParams)
	log.V(1).Info(fmt.Sprintf("Number of clients returned: %d", len(clients)), "namespace", keycloakClient.Namespace, "name", keycloakClient.Name,
		"clientID", *keycloakClient.Spec.ClientID, "realm", *keycloakClient.Spec.Realm)
	if err != nil {
		log.Error(err, "Failed to get Keycloak Clients", "clientID", *keycloakClient.Spec.ClientID, "realm", *keycloakClient.Spec.Realm)
		return nil, err
	}
	if len(clients) < 1 {
		return nil, nil
	}
	client := clients[0]
	if keycloakClient.Spec.ClientAuthenticatorType != nil && *keycloakClient.Spec.ClientAuthenticatorType == "client-secret" &&
		keycloakClient.Spec.PublicClient != nil && !*keycloakClient.Spec.PublicClient {
		log.V(1).Info("Get client secret", "namespace", keycloakClient.Namespace, "name", keycloakClient.Name, "clientID", keycloakClient.Spec.ClientID, "realm", keycloakClient.Spec.Realm)
		secret, err := server.GetClientSecret(ctx, Token.AccessToken, *keycloakClient.Spec.Realm, *client.ID)
		if err != nil {
			return nil, err
		}
		if secret == nil {
			err = fmt.Errorf("unable to get client secret")
			log.Error(err, "Unable to get Client Secret", "clientID", keycloakClient.Spec.ClientID, "realm", keycloakClient.Spec.Realm)
			return nil, err
		}
		client.Secret = secret.Value
	}
	return client, nil
}
