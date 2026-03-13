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

	"github.com/Nerzal/gocloak/v13"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	keycloakv1alpha1 "github.com/OSC/keycloak-cr-operator/api/v1alpha1"
	"github.com/OSC/keycloak-cr-operator/internal/models"
)

// MockGoCloak is a mock implementation of the GoCloak interface for testing
type MockGoCloak struct {
	mock.Mock
}

func (m *MockGoCloak) LoginAdmin(ctx context.Context, realm, username, password string) (*gocloak.JWT, error) {
	args := m.Called(ctx, realm, username, password)
	return args.Get(0).(*gocloak.JWT), args.Error(1)
}

func (m *MockGoCloak) GetClients(ctx context.Context, token, realm string, params gocloak.GetClientsParams) ([]*gocloak.Client, error) {
	args := m.Called(ctx, token, realm, params)
	return args.Get(0).([]*gocloak.Client), args.Error(1)
}

func (m *MockGoCloak) CreateClient(ctx context.Context, token, realm string, client gocloak.Client) (string, error) {
	args := m.Called(ctx, token, realm, client)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockGoCloak) UpdateClient(ctx context.Context, token, realm string, client gocloak.Client) error {
	args := m.Called(ctx, token, realm, client)
	return args.Error(0)
}

func (m *MockGoCloak) DeleteClient(ctx context.Context, token, realm, idOfClient string) error {
	args := m.Called(ctx, token, realm, idOfClient)
	return args.Error(0)
}

var _ = Describe("KeycloakClient Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		keycloakclient := &keycloakv1alpha1.KeycloakClient{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind KeycloakClient")
			err := k8sClient.Get(ctx, typeNamespacedName, keycloakclient)
			if err != nil && errors.IsNotFound(err) {
				resource := &keycloakv1alpha1.KeycloakClient{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: keycloakv1alpha1.KeycloakClientSpec{},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &keycloakv1alpha1.KeycloakClient{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance KeycloakClient")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")

			// Create a mock GoCloak client
			mockServer := new(MockGoCloak)

			// Set up expectations for the mock
			mockServer.On("LoginAdmin", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&gocloak.JWT{
				AccessToken: "test-token",
			}, nil)

			mockServer.On("GetClients", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*gocloak.Client{}, nil)

			mockServer.On("CreateClient", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", nil)

			controllerReconciler := &KeycloakClientReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
				Server: mockServer,
				Config: &models.KeycloakConfig{
					AdminUsername:  "admin",
					AdminPassword:  "password",
					AdminRealm:     "master",
					ClientIDPrefix: "kubernetes",
				},
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.

			// Verify that all expectations were met
			mockServer.AssertExpectations(GinkgoT())
		})
	})
})
