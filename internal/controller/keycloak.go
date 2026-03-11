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
	"sync"
	"time"

	"github.com/Nerzal/gocloak/v13"
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

func KeycloakLogin(ctx context.Context, server GoCloakServer, username, password, realm string) error {
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
	t, err := server.LoginAdmin(ctx, username, password, realm)
	if err != nil {
		log.Error(err, "Failed to login to Keycloak", "username", username, "realm", realm)
		return err
	}
	log.Info("Successfully logged into Keycloak", "username", username, "realm", realm)
	Token.AccessToken = t.AccessToken
	Token.IDToken = t.IDToken
	Token.ExpiresIn = t.ExpiresIn
	Token.CreatedAt = &now

	return nil
}
