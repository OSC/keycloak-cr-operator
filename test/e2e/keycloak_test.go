//go:build e2e
// +build e2e

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

package e2e

import (
	"context"
	"fmt"

	"github.com/Nerzal/gocloak/v13"
	"github.com/OSC/keycloak-cr-operator/internal/controller"
	"github.com/OSC/keycloak-cr-operator/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	keycloakURL   = fmt.Sprintf("http://localhost:%s", utils.KeycloakPort)
	gocloakClient = gocloak.NewClient(keycloakURL)
)

func keycloakLogin() {
	By("Login to Keycloak")

	ctx := context.Background()
	logger := zap.New(zap.UseDevMode(true))
	ctrl.SetLogger(logger)

	// Create KeycloakConfig struct based on parameters
	config := &controller.KeycloakConfig{
		AdminUsername:  utils.KeycloakAdminUsername,
		AdminPassword:  utils.KeycloakAdminPassword,
		AdminRealm:     "master",
		DefaultRealm:   "",
		ClientIDPrefix: "kubernetes",
	}

	err := controller.KeycloakLogin(ctx, gocloakClient, config)
	Expect(err).NotTo(HaveOccurred(), "Failed to login to Keycloak")
}

func getKeycloakClient(clientID, realm string) *gocloak.Client {
	keycloakLogin()
	By("Getting Keycloak client")

	ctx := context.Background()
	getClientParams := gocloak.GetClientsParams{
		ClientID: &clientID,
	}

	clients, err := gocloakClient.GetClients(ctx, controller.Token.AccessToken, realm, getClientParams)
	Expect(err).NotTo(HaveOccurred(), "Failed to get Keycloak Clients")

	if len(clients) < 1 {
		return nil
	}
	client := clients[0]

	secret, err := gocloakClient.GetClientSecret(ctx, controller.Token.AccessToken, realm, *client.ID)
	Expect(err).NotTo(HaveOccurred(), "Failed to get Keycloak client secret")

	client.Secret = secret.Value

	return client
}
