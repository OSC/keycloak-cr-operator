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
	"net/url"

	"github.com/Nerzal/gocloak/v13"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
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

func (m *MockGoCloak) GetClientSecret(ctx context.Context, token, realm, idOfClient string) (*gocloak.CredentialRepresentation, error) {
	args := m.Called(ctx, token, realm, idOfClient)
	return args.Get(0).(*gocloak.CredentialRepresentation), args.Error(1)
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
			clientID := "test"
			err := k8sClient.Get(ctx, typeNamespacedName, keycloakclient)
			if err != nil && errors.IsNotFound(err) {
				resource := &keycloakv1alpha1.KeycloakClient{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: keycloakv1alpha1.KeycloakClientSpec{
						ClientID: &clientID,
						Realm:    stringPtr("master"),
					},
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
					KeycloakURL: &url.URL{
						Scheme: "http",
						Host:   "keycloak.keycloak.svc",
					},
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

		It("should handle secret creation when ClientSecretRef is configured", func() {
			By("Creating a KeycloakClient with ClientSecretRef")

			// Create a KeycloakClient with ClientSecretRef configuration
			clientID := "test-client-with-secret"
			keycloakClientWithSecret := &keycloakv1alpha1.KeycloakClient{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-keycloak-client-with-secret",
					Namespace: "default",
				},
				Spec: keycloakv1alpha1.KeycloakClientSpec{
					ClientID:                &clientID,
					Realm:                   stringPtr("master"),
					ClientAuthenticatorType: stringPtr("client-secret"),
					PublicClient:            boolPtr(false),
					ClientSecretRef: &keycloakv1alpha1.KeycloakClientSecret{
						SecretKeySelector: corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "test-client-secret",
							},
							Key: "client-secret",
						},
						Create: boolPtr(true),
					},
				},
			}

			// Create the resource in the test cluster
			Expect(k8sClient.Create(ctx, keycloakClientWithSecret)).To(Succeed())

			// Clean up after test
			DeferCleanup(func() {
				Expect(k8sClient.Delete(ctx, keycloakClientWithSecret)).To(Succeed())
			})

			// Create a mock GoCloak client that returns a secret
			mockServer := new(MockGoCloak)
			mockServer.On("LoginAdmin", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&gocloak.JWT{
				AccessToken: "test-token",
			}, nil)

			mockServer.On("GetClients", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*gocloak.Client{}, nil).Once()

			mockServer.On("GetClients", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*gocloak.Client{
				{
					ID:       stringPtr("test"),
					ClientID: stringPtr("test"),
				},
			}, nil).Once()

			mockServer.On("CreateClient", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", nil)

			mockServer.On("GetClientSecret", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&gocloak.CredentialRepresentation{
				Value: stringPtr("secret"),
			}, nil)

			controllerReconciler := &KeycloakClientReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
				Server: mockServer,
				Config: &models.KeycloakConfig{
					KeycloakURL: &url.URL{
						Scheme: "http",
						Host:   "keycloak.keycloak.svc",
					},
					AdminUsername:  "admin",
					AdminPassword:  "password",
					AdminRealm:     "master",
					ClientIDPrefix: "kubernetes",
				},
			}

			// Reconcile the resource
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-keycloak-client-with-secret",
					Namespace: "default",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			// Verify that the secret was created
			secret := &corev1.Secret{}
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      "test-client-secret",
				Namespace: "default",
			}, secret)
			Expect(err).NotTo(HaveOccurred())
			Expect(secret.Data).To(HaveKey("client-secret"))
			Expect(secret.Data).To(HaveKey("CLIENT_SECRET"))
			Expect(secret.Data).To(HaveKey("cookie-secret"))
			Expect(secret.Data).To(HaveKey("COOKIE_SECRET"))
			cookieSecret, ok := secret.Data["cookie-secret"]
			Expect(ok).To(BeTrue())
			Expect(string(cookieSecret)).To(Not(BeEmpty()))
			cookieSecret, ok = secret.Data["COOKIE_SECRET"]
			Expect(ok).To(BeTrue())
			Expect(string(cookieSecret)).To(Not(BeEmpty()))

			// Verify the secret has the correct owner reference
			controllerRefs := secret.GetOwnerReferences()
			Expect(controllerRefs).To(HaveLen(1))
			Expect(controllerRefs[0].Name).To(Equal("test-keycloak-client-with-secret"))
			Expect(controllerRefs[0].Kind).To(Equal("KeycloakClient"))

			// Verify that all expectations were met
			mockServer.AssertExpectations(GinkgoT())
		})

		It("should handle config map creation and data population", func() {
			By("Creating a KeycloakClient with ConfigMapName specified")

			// Create a KeycloakClient with ConfigMapName
			clientID := "test-client-with-configmap"
			configMapName := "custom-configmap-name"
			keycloakClientWithConfigMap := &keycloakv1alpha1.KeycloakClient{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-keycloak-client-with-configmap",
					Namespace: "default",
				},
				Spec: keycloakv1alpha1.KeycloakClientSpec{
					ClientID:      &clientID,
					Realm:         stringPtr("master"),
					ConfigMapName: &configMapName,
				},
			}

			// Create the resource in the test cluster
			Expect(k8sClient.Create(ctx, keycloakClientWithConfigMap)).To(Succeed())

			// Clean up after test
			DeferCleanup(func() {
				Expect(k8sClient.Delete(ctx, keycloakClientWithConfigMap)).To(Succeed())
			})

			// Create a mock GoCloak client
			mockServer := new(MockGoCloak)
			mockServer.On("LoginAdmin", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&gocloak.JWT{
				AccessToken: "test-token",
			}, nil)

			mockServer.On("GetClients", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*gocloak.Client{}, nil).Once()

			mockServer.On("CreateClient", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", nil)

			controllerReconciler := &KeycloakClientReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
				Server: mockServer,
				Config: &models.KeycloakConfig{
					KeycloakURL: &url.URL{
						Scheme: "http",
						Host:   "keycloak.keycloak.svc",
					},
					AdminUsername:  "admin",
					AdminPassword:  "password",
					AdminRealm:     "master",
					DefaultRealm:   "master",
					ClientIDPrefix: "kubernetes",
				},
			}

			// Reconcile the resource
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-keycloak-client-with-configmap",
					Namespace: "default",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			// Verify that the config map was created
			configMap := &corev1.ConfigMap{}
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      "custom-configmap-name",
				Namespace: "default",
			}, configMap)
			Expect(err).NotTo(HaveOccurred())

			// Verify ConfigMap data
			Expect(configMap.Data).To(HaveKey("client-id"))
			Expect(configMap.Data).To(HaveKey("CLIENT_ID"))
			Expect(configMap.Data).To(HaveKey("keycloak-url"))
			Expect(configMap.Data).To(HaveKey("KEYCLOAK_URL"))
			Expect(configMap.Data).To(HaveKey("issuer-url"))
			Expect(configMap.Data).To(HaveKey("ISSUER_URL"))

			// Verify the values are correct
			Expect(configMap.Data["client-id"]).To(Equal("test-client-with-configmap"))
			Expect(configMap.Data["CLIENT_ID"]).To(Equal("test-client-with-configmap"))
			Expect(configMap.Data["keycloak-url"]).To(Equal("http://keycloak.keycloak.svc"))
			Expect(configMap.Data["KEYCLOAK_URL"]).To(Equal("http://keycloak.keycloak.svc"))
			Expect(configMap.Data["issuer-url"]).To(Equal("http://keycloak.keycloak.svc/realms/master"))
			Expect(configMap.Data["ISSUER_URL"]).To(Equal("http://keycloak.keycloak.svc/realms/master"))

			// Verify the ConfigMap has the correct owner reference
			controllerRefs := configMap.GetOwnerReferences()
			Expect(controllerRefs).To(HaveLen(1))
			Expect(controllerRefs[0].Name).To(Equal("test-keycloak-client-with-configmap"))
			Expect(controllerRefs[0].Kind).To(Equal("KeycloakClient"))

			// Verify that all expectations were met
			mockServer.AssertExpectations(GinkgoT())
		})

		It("should handle config map creation with default name", func() {
			By("Creating a KeycloakClient without ConfigMapName specified")

			// Create a KeycloakClient without specifying ConfigMapName
			clientID := "test-client-default-configmap"
			keycloakClientWithoutConfigMap := &keycloakv1alpha1.KeycloakClient{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-keycloak-client-default-configmap",
					Namespace: "default",
				},
				Spec: keycloakv1alpha1.KeycloakClientSpec{
					ClientID: &clientID,
					Realm:    stringPtr("master"),
				},
			}

			// Create the resource in the test cluster
			Expect(k8sClient.Create(ctx, keycloakClientWithoutConfigMap)).To(Succeed())

			// Clean up after test
			DeferCleanup(func() {
				Expect(k8sClient.Delete(ctx, keycloakClientWithoutConfigMap)).To(Succeed())
			})

			// Create a mock GoCloak client
			mockServer := new(MockGoCloak)
			mockServer.On("LoginAdmin", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&gocloak.JWT{
				AccessToken: "test-token",
			}, nil)

			mockServer.On("GetClients", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*gocloak.Client{}, nil).Once()

			mockServer.On("CreateClient", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", nil)

			controllerReconciler := &KeycloakClientReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
				Server: mockServer,
				Config: &models.KeycloakConfig{
					KeycloakURL: &url.URL{
						Scheme: "http",
						Host:   "keycloak.keycloak.svc",
					},
					AdminUsername:  "admin",
					AdminPassword:  "password",
					AdminRealm:     "master",
					DefaultRealm:   "master",
					ClientIDPrefix: "kubernetes",
				},
			}

			// Reconcile the resource
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-keycloak-client-default-configmap",
					Namespace: "default",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			// Verify that the config map was created with default name
			configMap := &corev1.ConfigMap{}
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      "test-keycloak-client-default-configmap-config",
				Namespace: "default",
			}, configMap)
			Expect(err).NotTo(HaveOccurred())

			// Verify ConfigMap data
			Expect(configMap.Data).To(HaveKey("client-id"))
			Expect(configMap.Data).To(HaveKey("keycloak-url"))
			Expect(configMap.Data).To(HaveKey("issuer-url"))

			// Verify the values are correct
			Expect(configMap.Data["client-id"]).To(Equal("test-client-default-configmap"))
			Expect(configMap.Data["keycloak-url"]).To(Equal("http://keycloak.keycloak.svc"))
			Expect(configMap.Data["issuer-url"]).To(Equal("http://keycloak.keycloak.svc/realms/master"))

			// Verify the ConfigMap has the correct owner reference
			controllerRefs := configMap.GetOwnerReferences()
			Expect(controllerRefs).To(HaveLen(1))
			Expect(controllerRefs[0].Name).To(Equal("test-keycloak-client-default-configmap"))
			Expect(controllerRefs[0].Kind).To(Equal("KeycloakClient"))

			// Verify that all expectations were met
			mockServer.AssertExpectations(GinkgoT())
		})
	})
})
