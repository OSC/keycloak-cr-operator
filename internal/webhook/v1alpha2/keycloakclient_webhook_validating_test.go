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

package v1alpha2

import (
	"text/template"

	keycloakv1alpha2 "github.com/OSC/keycloak-cr-operator/api/v1alpha2"
	"github.com/OSC/keycloak-cr-operator/internal/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func WebhookValidating() {
	var (
		obj              *keycloakv1alpha2.KeycloakClient
		oldObj           *keycloakv1alpha2.KeycloakClient
		validator        KeycloakClientCustomValidator
		defaultConfigMap *keycloakv1alpha2.KeycloakClientConfigMap
		defaultSecret    *keycloakv1alpha2.KeycloakClientSecret
	)

	BeforeEach(func() {
		obj = &keycloakv1alpha2.KeycloakClient{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-keycloak-client",
				Namespace: "test-namespace",
			},
		}
		oldObj = &keycloakv1alpha2.KeycloakClient{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-keycloak-client",
				Namespace: "test-namespace",
			},
		}
		validator = KeycloakClientCustomValidator{
			keycloakConfig: &models.KeycloakConfig{
				DefaultRealm:   "master",
				ClientIDPrefix: "kubernetes",
				AllowedRealms:  []string{},
			},
		}
		defaultConfigMap = &keycloakv1alpha2.KeycloakClientConfigMap{
			Name:       &configMapName,
			EnvVarKeys: new(true),
		}
		defaultSecret = &keycloakv1alpha2.KeycloakClientSecret{
			Name:            new("test-secret"),
			ClientSecretKey: new("TEST_KEY"),
			ClientIdKey:     new("TEST_KEY"),
			IssuerUrlKey:    new("TEST_KEY"),
			CookieSecretKey: new("COOKIE_SECRET"),
			Create:          new(true),
			EnvVarKeys:      new(true),
		}
		Expect(validator).NotTo(BeNil(), "Expected validator to be initialized")
		Expect(oldObj).NotTo(BeNil(), "Expected oldObj to be initialized")
		Expect(obj).NotTo(BeNil(), "Expected obj to be initialized")
	})

	Context("When creating or updating KeycloakClient under Validating Webhook", func() {
		It("Should deny creation if ClientID is not set", func() {
			By("Setting empty ClientID")
			obj.Spec.ClientID = nil

			By("Validating creation should fail")
			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(warnings).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("clientID must be set"))
		})

		It("Should deny creation if ClientID has incorrect prefix", func() {
			By("Setting ClientID")
			clientID := "test-client"
			obj.Spec.ClientID = &clientID

			By("Validating creation should fail")
			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(warnings).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("clientID must begin with the prefix"))
		})

		It("Should validate ClientID against ClientIDRequired template", func() {
			By("Setting up config with ClientIDRequired template")
			tmpl, err := template.New("clientID").Parse("{{.Obj.Namespace}}-{{.Obj.Name}}-{{.Config.DefaultRealm}}")
			Expect(err).ToNot(HaveOccurred())
			validator.keycloakConfig.ClientIDRequired = tmpl

			By("Setting valid ClientID that matches the template")
			clientID := "test-namespace-test-keycloak-client-master"
			obj.Spec.ClientID = &clientID
			obj.Spec.Realm = &testRealm
			obj.Spec.ConfigMap = defaultConfigMap
			obj.Spec.Secret = defaultSecret

			By("Validating creation should succeed")
			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(warnings).To(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should deny creation if ClientID does not match ClientIDRequired template", func() {
			By("Setting up config with ClientIDRequired template")
			tmpl, err := template.New("clientID").Parse("{{.Obj.Namespace}}-{{.Obj.Name}}-{{.Config.DefaultRealm}}")
			Expect(err).ToNot(HaveOccurred())
			validator.keycloakConfig.ClientIDRequired = tmpl

			By("Setting invalid ClientID that does not match the template")
			clientID := "invalid-client-id"
			obj.Spec.ClientID = &clientID

			By("Validating creation should fail")
			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(warnings).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("clientID must match the required template"))
		})

		It("Should deny creation if Realm is not set", func() {
			By("Setting empty Realm")
			obj.Spec.Realm = nil

			By("Validating creation should fail")
			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(warnings).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("realm must be set"))
		})

		It("Should deny creation if Realm is not in allowed list", func() {
			By("Setting valid Realms")
			validator.keycloakConfig.AllowedRealms = []string{"test-realm"}

			By("Setting empty Realm")
			obj.Spec.Realm = new("master")

			By("Validating creation should fail")
			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(warnings).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("realm must be one of: test-realm"))
		})

		It("Should allow creation if both ClientID and Realm are set", func() {
			By("Setting valid ClientID and Realm")
			obj.Spec.ClientID = &clientIDWithPrefix
			obj.Spec.Realm = &testRealm
			obj.Spec.Secret = defaultSecret
			obj.Spec.ConfigMap = &keycloakv1alpha2.KeycloakClientConfigMap{
				Name:       new("test-keycloak-client-config"),
				EnvVarKeys: new(true),
			}

			By("Validating creation should succeed")
			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(warnings).To(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should allow creation of ClientID without ClientIDPrefix", func() {
			By("Setting empty ClientIDPrefix")
			validator.keycloakConfig.ClientIDPrefix = ""

			By("Setting valid ClientID and Realm")
			clientID := "test-client"
			obj.Spec.ClientID = &clientID
			obj.Spec.Realm = &testRealm
			obj.Spec.ConfigMap = defaultConfigMap
			obj.Spec.Secret = defaultSecret

			By("Validating creation should succeed")
			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(warnings).To(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should deny creation if Secret is missing", func() {
			By("Setting up client without Secret")
			obj.Spec.ClientID = &clientIDWithPrefix
			obj.Spec.Realm = &testRealm
			obj.Spec.ConfigMap = defaultConfigMap

			By("Validating creation should fail due to missing Secret")
			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(warnings).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("secret must be set"))
		})

		It("Should allow creation if Secret is present", func() {
			By("Setting up client Secret")
			obj.Spec.ClientID = &clientIDWithPrefix
			obj.Spec.Realm = &testRealm
			obj.Spec.ConfigMap = defaultConfigMap
			obj.Spec.Secret = defaultSecret

			By("Validating creation should succeed")
			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(warnings).To(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should validate ConfigMap structure properly", func() {
			By("Setting up a client with ConfigMap structure")
			obj.Spec.ClientID = &clientIDWithPrefix
			obj.Spec.Realm = &testRealm
			obj.Spec.ConfigMap = defaultConfigMap
			obj.Spec.Secret = defaultSecret

			By("Validating creation should succeed")
			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(warnings).To(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should deny creation if ConfigMap.Name is not set", func() {
			By("Setting up a client with empty ConfigMap.Name")
			obj.Spec.ClientID = &clientIDWithPrefix
			obj.Spec.Realm = &testRealm
			obj.Spec.Secret = defaultSecret
			obj.Spec.ConfigMap = &keycloakv1alpha2.KeycloakClientConfigMap{
				Name:       nil,
				EnvVarKeys: new(true),
			}

			By("Validating creation should fail")
			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(warnings).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("configMap.name must be set"))
		})

		It("Should deny creation if ConfigMap.EnvVarKeys is not set", func() {
			By("Setting up a client with empty ConfigMap.EnvVarKeys")
			obj.Spec.ClientID = &clientIDWithPrefix
			obj.Spec.Realm = &testRealm
			obj.Spec.Secret = defaultSecret
			obj.Spec.ConfigMap = &keycloakv1alpha2.KeycloakClientConfigMap{
				Name:       &configMapName,
				EnvVarKeys: nil,
			}

			By("Validating creation should fail")
			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(warnings).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("configMap.envVarKeys must be set"))
		})

		It("Should validate updates correctly", func() {
			By("Setting up a valid client")
			obj.Spec.ClientID = &clientIDWithPrefix
			obj.Spec.Realm = &testRealm
			obj.Spec.ConfigMap = defaultConfigMap
			obj.Spec.Secret = defaultSecret

			By("Validating update should succeed")
			warnings, err := validator.ValidateUpdate(ctx, oldObj, obj)
			Expect(warnings).To(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should deny update if ClientID is not set", func() {
			By("Setting empty ClientID in update")
			obj.Spec.ClientID = nil

			By("Validating update should fail")
			warnings, err := validator.ValidateUpdate(ctx, oldObj, obj)
			Expect(warnings).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("clientID must be set"))
		})

		Context("When creating or updating KeycloakClient under Validating Webhook - Secret Validation", func() {
			It("Should deny creation if Secret name is empty", func() {
				By("Setting up client with Secret empty name")
				obj.Spec.ClientID = &clientIDWithPrefix
				obj.Spec.Realm = &testRealm
				obj.Spec.ConfigMap = defaultConfigMap

				secretRef := keycloakv1alpha2.KeycloakClientSecret{
					Name:            new(""),
					ClientSecretKey: new("test-key"),
				}
				obj.Spec.Secret = &secretRef

				By("Validating creation should fail due to empty name")
				warnings, err := validator.ValidateCreate(ctx, obj)
				Expect(warnings).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("secret name must be set"))
				Expect(err.Error()).To(ContainSubstring("secret create must be set"))
				Expect(err.Error()).To(ContainSubstring("secret envVarKeys must be set"))
			})

			It("Should deny creation if Secret keys are empty", func() {
				By("Setting up client with Secret with empty keys")
				obj.Spec.ClientID = &clientIDWithPrefix
				obj.Spec.Realm = &testRealm
				obj.Spec.ConfigMap = defaultConfigMap

				secretRef := keycloakv1alpha2.KeycloakClientSecret{
					Name:            new("test-client"),
					ClientSecretKey: new(""),
					ClientIdKey:     new(""),
					IssuerUrlKey:    new(""),
					CookieSecretKey: new(""),
				}
				obj.Spec.Secret = &secretRef

				By("Validating creation should fail due to empty clientSecretKey")
				warnings, err := validator.ValidateCreate(ctx, obj)
				Expect(warnings).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("secret clientSecretKey must be set"))
				Expect(err.Error()).To(ContainSubstring("secret clientIdKey must be set"))
				Expect(err.Error()).To(ContainSubstring("secret issuerUrlKey must be set"))
				Expect(err.Error()).To(ContainSubstring("secret cookieSecretKey must be set"))
				Expect(err.Error()).To(ContainSubstring("secret create must be set"))
				Expect(err.Error()).To(ContainSubstring("secret envVarKeys must be set"))
			})

			It("Should deny creation if Secret keys are not upper snake case when EnvVarKeys is true", func() {
				By("Setting up client with Secret with non-upper-snake-case clientSecretKey")
				obj.Spec.ClientID = &clientIDWithPrefix
				obj.Spec.Realm = &testRealm
				obj.Spec.ConfigMap = defaultConfigMap

				secretRef := keycloakv1alpha2.KeycloakClientSecret{
					Name:            new(""),
					ClientSecretKey: new("testKey"),
					ClientIdKey:     new("testKey"),
					IssuerUrlKey:    new("testKey"),
					CookieSecretKey: new("testKey"),
					Create:          new(true),
					EnvVarKeys:      new(true),
				}
				obj.Spec.Secret = &secretRef

				By("Validating creation should fail due to non-upper-snake-case key")
				warnings, err := validator.ValidateCreate(ctx, obj)
				Expect(warnings).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("secret clientSecretKey must be upper snake case when envVarKeys is true"))
				Expect(err.Error()).To(ContainSubstring("secret clientIdKey must be upper snake case when envVarKeys is true"))
				Expect(err.Error()).To(ContainSubstring("secret issuerUrlKey must be upper snake case when envVarKeys is true"))
				Expect(err.Error()).To(ContainSubstring("secret cookieSecretKey must be upper snake case when envVarKeys is true"))
			})

			It("Should allow creation if Secret keys are upper snake case when EnvVarKeys is true", func() {
				By("Setting up client with Secret with upper snake case clientSecretKey")
				obj.Spec.ClientID = &clientIDWithPrefix
				obj.Spec.Realm = &testRealm
				obj.Spec.ConfigMap = defaultConfigMap

				secretRef := keycloakv1alpha2.KeycloakClientSecret{
					Name:            new("test-secret"),
					ClientSecretKey: new("TEST_KEY"),
					ClientIdKey:     new("TEST_KEY"),
					IssuerUrlKey:    new("TEST_KEY"),
					CookieSecretKey: new("COOKIE_SECRET"),
					Create:          new(true),
					EnvVarKeys:      new(true),
				}
				obj.Spec.Secret = &secretRef

				By("Validating creation should succeed")
				warnings, err := validator.ValidateCreate(ctx, obj)
				Expect(warnings).To(BeNil())
				Expect(err).ToNot(HaveOccurred())
			})

			It("Should allow creation if ClientSecretRef key is not upper snake case when EnvVarKeys is false", func() {
				By("Setting up client with Secret with non-upper-snake-case key")
				obj.Spec.ClientID = &clientIDWithPrefix
				obj.Spec.Realm = &testRealm
				obj.Spec.ConfigMap = defaultConfigMap

				secretRef := keycloakv1alpha2.KeycloakClientSecret{
					Name:            new("test-secret"),
					ClientSecretKey: new("testKey"),
					ClientIdKey:     new("testKey"),
					IssuerUrlKey:    new("testKey"),
					CookieSecretKey: new("testKey"),
					Create:          new(true),
					EnvVarKeys:      new(false),
				}
				obj.Spec.Secret = &secretRef

				By("Validating creation should succeed")
				warnings, err := validator.ValidateCreate(ctx, obj)
				Expect(warnings).To(BeNil())
				Expect(err).ToNot(HaveOccurred())
			})

		})
	})

	Context("When creating or updating KeycloakClient under Validating Webhook - ProtocolMappers Validation", func() {
		It("Should allow creation if ProtocolMappers is nil", func() {
			By("Setting up a client with nil ProtocolMappers")
			obj.Spec.ClientID = &clientIDWithPrefix
			obj.Spec.Realm = &testRealm
			obj.Spec.ConfigMap = defaultConfigMap
			obj.Spec.Secret = defaultSecret
			obj.Spec.ProtocolMappers = nil

			By("Validating creation should succeed")
			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(warnings).To(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should allow creation if ProtocolMappers is empty slice", func() {
			By("Setting up a client with empty ProtocolMappers")
			obj.Spec.ClientID = &clientIDWithPrefix
			obj.Spec.Realm = &testRealm
			obj.Spec.ConfigMap = defaultConfigMap
			obj.Spec.Secret = defaultSecret
			obj.Spec.ProtocolMappers = []*keycloakv1alpha2.KeycloakClientProtocolMapper{}

			By("Validating creation should succeed")
			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(warnings).To(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should deny creation if ProtocolMappers has a mapper with missing name", func() {
			By("Setting up a client with ProtocolMappers that has a mapper with missing name")
			obj.Spec.ClientID = &clientIDWithPrefix
			obj.Spec.Realm = &testRealm
			obj.Spec.ConfigMap = defaultConfigMap
			obj.Spec.Secret = defaultSecret
			obj.Spec.ProtocolMappers = []*keycloakv1alpha2.KeycloakClientProtocolMapper{
				{
					Name:     nil, // Missing name
					Type:     new("oidc-hardcoded-claim-mapper"),
					Protocol: new("openid-connect"),
				},
			}

			By("Validating creation should fail")
			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(warnings).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("name must be set"))
		})

		It("Should deny creation if ProtocolMappers has a mapper with empty name", func() {
			By("Setting up a client with ProtocolMappers that has a mapper with empty name")
			obj.Spec.ClientID = &clientIDWithPrefix
			obj.Spec.Realm = &testRealm
			obj.Spec.ConfigMap = defaultConfigMap
			obj.Spec.Secret = defaultSecret
			obj.Spec.ProtocolMappers = []*keycloakv1alpha2.KeycloakClientProtocolMapper{
				{
					Name:     new(""), // Empty name
					Type:     new("oidc-hardcoded-claim-mapper"),
					Protocol: new("openid-connect"),
				},
			}

			By("Validating creation should fail")
			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(warnings).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("name must be set"))
		})

		It("Should deny creation if ProtocolMappers has a mapper with missing type", func() {
			By("Setting up a client with ProtocolMappers that has a mapper with missing type")
			obj.Spec.ClientID = &clientIDWithPrefix
			obj.Spec.Realm = &testRealm
			obj.Spec.ConfigMap = defaultConfigMap
			obj.Spec.Secret = defaultSecret
			obj.Spec.ProtocolMappers = []*keycloakv1alpha2.KeycloakClientProtocolMapper{
				{
					Name:     new("test-mapper"),
					Type:     nil, // Missing type
					Protocol: new("openid-connect"),
				},
			}

			By("Validating creation should fail")
			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(warnings).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("type must be set"))
		})

		It("Should deny creation if ProtocolMappers has a mapper with empty type", func() {
			By("Setting up a client with ProtocolMappers that has a mapper with empty type")
			obj.Spec.ClientID = &clientIDWithPrefix
			obj.Spec.Realm = &testRealm
			obj.Spec.ConfigMap = defaultConfigMap
			obj.Spec.Secret = defaultSecret
			obj.Spec.ProtocolMappers = []*keycloakv1alpha2.KeycloakClientProtocolMapper{
				{
					Name:     new("test-mapper"),
					Type:     new(""), // Empty type
					Protocol: new("openid-connect"),
				},
			}

			By("Validating creation should fail")
			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(warnings).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("type must be set"))
		})

		It("Should deny creation if ProtocolMappers has a mapper with missing protocol", func() {
			By("Setting up a client with ProtocolMappers that has a mapper with missing protocol")
			obj.Spec.ClientID = &clientIDWithPrefix
			obj.Spec.Realm = &testRealm
			obj.Spec.ConfigMap = defaultConfigMap
			obj.Spec.Secret = defaultSecret
			obj.Spec.ProtocolMappers = []*keycloakv1alpha2.KeycloakClientProtocolMapper{
				{
					Name:     new("test-mapper"),
					Type:     new("oidc-hardcoded-claim-mapper"),
					Protocol: nil, // Missing protocol
				},
			}

			By("Validating creation should fail")
			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(warnings).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("protocol must be set"))
		})

		It("Should deny creation if ProtocolMappers has a mapper with empty protocol", func() {
			By("Setting up a client with ProtocolMappers that has a mapper with empty protocol")
			obj.Spec.ClientID = &clientIDWithPrefix
			obj.Spec.Realm = &testRealm
			obj.Spec.ConfigMap = defaultConfigMap
			obj.Spec.Secret = defaultSecret
			obj.Spec.ProtocolMappers = []*keycloakv1alpha2.KeycloakClientProtocolMapper{
				{
					Name:     new("test-mapper"),
					Type:     new("oidc-hardcoded-claim-mapper"),
					Protocol: new(""), // Empty protocol
				},
			}

			By("Validating creation should fail")
			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(warnings).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("protocol must be set"))
		})

		It("Should allow creation if ProtocolMappers has valid mapper with all required fields", func() {
			By("Setting up a client with valid ProtocolMappers")
			obj.Spec.ClientID = &clientIDWithPrefix
			obj.Spec.Realm = &testRealm
			obj.Spec.ConfigMap = defaultConfigMap
			obj.Spec.Secret = defaultSecret
			obj.Spec.ProtocolMappers = []*keycloakv1alpha2.KeycloakClientProtocolMapper{
				{
					Name:     new("test-mapper"),
					Type:     new("oidc-hardcoded-claim-mapper"),
					Protocol: new("openid-connect"),
				},
			}

			By("Validating creation should succeed")
			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(warnings).To(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should deny creation if ProtocolMappers has oidc-audience-mapper with missing includedClientAudience", func() {
			By("Setting up a client with ProtocolMappers that has oidc-audience-mapper missing includedClientAudience")
			obj.Spec.ClientID = &clientIDWithPrefix
			obj.Spec.Realm = &testRealm
			obj.Spec.ConfigMap = defaultConfigMap
			obj.Spec.Secret = defaultSecret
			obj.Spec.ProtocolMappers = []*keycloakv1alpha2.KeycloakClientProtocolMapper{
				{
					Name:     new("audience-mapper"),
					Type:     new("oidc-audience-mapper"),
					Protocol: new("openid-connect"),
					// Missing includedClientAudience
				},
			}

			By("Validating creation should fail")
			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(warnings).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("includedClientAudience must be set when type is oidc-audience-mapper"))
		})

		It("Should deny creation if ProtocolMappers has oidc-audience-mapper with empty includedClientAudience", func() {
			By("Setting up a client with ProtocolMappers that has oidc-audience-mapper with empty includedClientAudience")
			obj.Spec.ClientID = &clientIDWithPrefix
			obj.Spec.Realm = &testRealm
			obj.Spec.ConfigMap = defaultConfigMap
			obj.Spec.Secret = defaultSecret
			obj.Spec.ProtocolMappers = []*keycloakv1alpha2.KeycloakClientProtocolMapper{
				{
					Name:                   new("audience-mapper"),
					Type:                   new("oidc-audience-mapper"),
					Protocol:               new("openid-connect"),
					IncludedClientAudience: new(""), // Empty includedClientAudience
				},
			}

			By("Validating creation should fail")
			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(warnings).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("includedClientAudience must be set when type is oidc-audience-mapper"))
		})

		It("Should allow creation if ProtocolMappers has oidc-audience-mapper with valid includedClientAudience", func() {
			By("Setting up a client with ProtocolMappers that has oidc-audience-mapper with valid includedClientAudience")
			obj.Spec.ClientID = &clientIDWithPrefix
			obj.Spec.Realm = &testRealm
			obj.Spec.ConfigMap = defaultConfigMap
			obj.Spec.Secret = defaultSecret
			obj.Spec.ProtocolMappers = []*keycloakv1alpha2.KeycloakClientProtocolMapper{
				{
					Name:                   new("audience-mapper"),
					Type:                   new("oidc-audience-mapper"),
					Protocol:               new("openid-connect"),
					IncludedClientAudience: new("test-client"),
				},
			}

			By("Validating creation should succeed")
			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(warnings).To(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should allow creation if ProtocolMappers has multiple valid mappers", func() {
			By("Setting up a client with multiple valid ProtocolMappers")
			obj.Spec.ClientID = &clientIDWithPrefix
			obj.Spec.Realm = &testRealm
			obj.Spec.ConfigMap = defaultConfigMap
			obj.Spec.Secret = defaultSecret
			obj.Spec.ProtocolMappers = []*keycloakv1alpha2.KeycloakClientProtocolMapper{
				{
					Name:     new("mapper1"),
					Type:     new("oidc-hardcoded-claim-mapper"),
					Protocol: new("openid-connect"),
				},
				{
					Name:                   new("mapper2"),
					Type:                   new("oidc-audience-mapper"),
					Protocol:               new("openid-connect"),
					IncludedClientAudience: new("test-client"),
				},
			}

			By("Validating creation should succeed")
			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(warnings).To(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should validate updates correctly with ProtocolMappers", func() {
			By("Setting up a valid client with ProtocolMappers")
			obj.Spec.ClientID = &clientIDWithPrefix
			obj.Spec.Realm = &testRealm
			obj.Spec.ConfigMap = defaultConfigMap
			obj.Spec.Secret = defaultSecret
			obj.Spec.ProtocolMappers = []*keycloakv1alpha2.KeycloakClientProtocolMapper{
				{
					Name:     new("test-mapper"),
					Type:     new("oidc-hardcoded-claim-mapper"),
					Protocol: new("openid-connect"),
				},
			}

			By("Validating update should succeed")
			warnings, err := validator.ValidateUpdate(ctx, oldObj, obj)
			Expect(warnings).To(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should deny update if ProtocolMappers has a mapper with missing name", func() {
			By("Setting up a client with ProtocolMappers that has a mapper with missing name")
			obj.Spec.ClientID = &clientIDWithPrefix
			obj.Spec.Realm = &testRealm
			obj.Spec.ConfigMap = defaultConfigMap
			obj.Spec.Secret = defaultSecret
			obj.Spec.ProtocolMappers = []*keycloakv1alpha2.KeycloakClientProtocolMapper{
				{
					Name:     nil, // Missing name
					Type:     new("oidc-hardcoded-claim-mapper"),
					Protocol: new("openid-connect"),
				},
			}

			By("Validating update should fail")
			warnings, err := validator.ValidateUpdate(ctx, oldObj, obj)
			Expect(warnings).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("name must be set"))
		})
	})
}
