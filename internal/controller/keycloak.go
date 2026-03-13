package controller

import (
	"context"
	"sync"
	"time"

	"github.com/Nerzal/gocloak/v13"
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
