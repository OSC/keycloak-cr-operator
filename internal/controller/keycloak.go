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
	"time"

	"github.com/Nerzal/gocloak/v13"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type KeycloakToken struct {
	gocloak.JWT
	CreatedAt *time.Time
}

func KeycloakLogin(ctx context.Context, client *gocloak.GoCloak, username, password, realm string, token *KeycloakToken) error {
	log := logf.FromContext(ctx)
	if token != nil && token.ExpiresIn != 0 && token.CreatedAt != nil {
		expirationTime := token.CreatedAt.Add(time.Duration(token.ExpiresIn) * time.Second)
		if time.Now().Before(expirationTime) {
			log.Info("admin token still valid", "expiresIn", token.ExpiresIn, "createdAt", token.CreatedAt)
			return nil
		}
	}

	now := time.Now()
	t, err := client.LoginAdmin(context.Background(), username, password, realm)
	if err != nil {
		log.Error(err, "Failed to login to Keycloak", "username", username, "realm", realm)
		return err
	}
	log.Info("Successfully logged into Keycloak", "username", username, "realm", realm)
	token.AccessToken = t.AccessToken
	token.IDToken = t.IDToken
	token.ExpiresIn = t.ExpiresIn
	token.CreatedAt = &now

	return nil
}
