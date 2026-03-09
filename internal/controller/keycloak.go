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

func KeycloakLogin(ctx context.Context, client *gocloak.GoCloak, username, password, realm string, jwt *gocloak.JWT, tokenCreatedAt *time.Time) error {
	log := logf.FromContext(ctx)
	if jwt != nil && jwt.ExpiresIn != 0 && tokenCreatedAt != nil {
		expirationTime := tokenCreatedAt.Add(time.Duration(jwt.ExpiresIn) * time.Second)
		if time.Now().Before(expirationTime) {
			log.Info("admin token still valid", "expiresIn", jwt.ExpiresIn, "createdAt", tokenCreatedAt)
			return nil
		}
	}

	now := time.Now()
	token, err := client.LoginAdmin(context.Background(), username, password, realm)
	if err != nil {
		log.Error(err, "Failed to login to Keycloak", "username", username, "realm", realm)
		return err
	}
	log.Info("Successfully logged into Keycloak", "username", username, "realm", realm)
	jwt = token
	tokenCreatedAt = &now

	return nil
}
