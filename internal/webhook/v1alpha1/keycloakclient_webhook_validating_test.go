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

package v1alpha1

import (
	"text/template"

	keycloakv1alpha1 "github.com/OSC/keycloak-cr-operator/api/v1alpha1"
	"github.com/OSC/keycloak-cr-operator/internal/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	testSecretNameForValidating = "test-secret"
	testKeyForValidating        = "TEST_KEY"
)

func WebhookValidating() {
	var (
		obj              *keycloakv1alpha1.KeycloakClient
		oldObj           *keycloakv1alpha1.KeycloakClient
		validator        KeycloakClientCustomValidator
		defaultConfigMap *keycloakv1alpha1.KeycloakClientConfigMap
	)

	BeforeEach(func() {
		obj = &keycloakv1alpha1.KeycloakClient{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testKeycloakClient,
				Namespace: testNamespace,
			},
		}
		oldObj = &keycloakv1alpha1.KeycloakClient{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testKeycloakClient,
				Namespace: testNamespace,
			},
		}
		validator = KeycloakClientCustomValidator{
			keycloakConfig: &models.KeycloakConfig{
				DefaultRealm:   masterRealm,
				ClientIDPrefix: kubernetesPrefix,
				AllowedRealms:  []string{},
			},
		}
		defaultConfigMap = &keycloakv1alpha1.KeycloakClientConfigMap{
			Name:       &configMapName,
			EnvVarKeys: boolPtr(true),
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
			obj.Spec.Realm = stringPtr("master")

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
			obj.Spec.ConfigMap = &keycloakv1alpha1.KeycloakClientConfigMap{
				Name:       stringPtr("test-keycloak-client-config"),
				EnvVarKeys: boolPtr(true),
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

			By("Validating creation should succeed")
			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(warnings).To(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should deny creation if ClientAuthenticatorType is client-secret and Public is false but ClientSecretRef is missing", func() {
			By("Setting up client with client-secret auth type and public=false")
			obj.Spec.ClientID = &clientIDWithPrefix
			obj.Spec.Realm = &testRealm
			obj.Spec.ClientAuthenticatorType = &clientSecretType
			public := false
			obj.Spec.PublicClient = &public

			By("Validating creation should fail due to missing ClientSecretRef")
			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(warnings).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("clientSecretRef must be set when clientAuthenticatorType is client-secret and public is false"))
		})

		It("Should allow creation if ClientAuthenticatorType is client-secret, Public is false, and ClientSecretRef is present", func() {
			By("Setting up client with client-secret auth type, public=false, and ClientSecretRef")
			obj.Spec.ClientID = &clientIDWithPrefix
			obj.Spec.Realm = &testRealm
			obj.Spec.ClientAuthenticatorType = &clientSecretType
			public := false
			obj.Spec.PublicClient = &public
			create := true
			obj.Spec.ConfigMap = defaultConfigMap

			// Create a fake secret reference
			secretRef := keycloakv1alpha1.KeycloakClientSecret{
				SecretKeySelector: corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: testSecretNameForValidating,
					},
					Key: testKeyForValidating,
				},
				Create:     &create,
				EnvVarKeys: boolPtr(true),
			}
			obj.Spec.ClientSecretRef = &secretRef

			By("Validating creation should succeed")
			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(warnings).To(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should allow creation if ClientAuthenticatorType is client-secret and Public is true", func() {
			By("Setting up client with client-secret auth type and public=true")
			obj.Spec.ClientID = &clientIDWithPrefix
			obj.Spec.Realm = &testRealm
			obj.Spec.ClientAuthenticatorType = &clientSecretType
			public := true
			obj.Spec.PublicClient = &public
			obj.Spec.ConfigMap = defaultConfigMap

			By("Validating creation should succeed even without ClientSecretRef")
			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(warnings).To(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should validate ConfigMap structure properly", func() {
			By("Setting up a client with ConfigMap structure")
			obj.Spec.ClientID = &clientIDWithPrefix
			obj.Spec.Realm = &testRealm
			obj.Spec.ConfigMap = defaultConfigMap

			By("Validating creation should succeed")
			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(warnings).To(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should deny creation if ConfigMap.Name is not set", func() {
			By("Setting up a client with empty ConfigMap.Name")
			obj.Spec.ClientID = &clientIDWithPrefix
			obj.Spec.Realm = &testRealm
			obj.Spec.ConfigMap = &keycloakv1alpha1.KeycloakClientConfigMap{
				Name:       nil,
				EnvVarKeys: boolPtr(true),
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
			obj.Spec.ConfigMap = &keycloakv1alpha1.KeycloakClientConfigMap{
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

		Context("When creating or updating KeycloakClient under Validating Webhook - ClientSecretRef Validation", func() {
			It("Should allow creation if ClientSecretRef key is valid", func() {
				By("Setting up client with client-secret auth type, public=false, and ClientSecretRef with key")
				obj.Spec.ClientID = &clientIDWithPrefix
				obj.Spec.Realm = &testRealm
				obj.Spec.ClientAuthenticatorType = &clientSecretType
				public := false
				obj.Spec.PublicClient = &public
				create := true
				obj.Spec.ConfigMap = defaultConfigMap

				// Set a ClientSecretRef with key that is valid (upper snake case)
				secretRef := keycloakv1alpha1.KeycloakClientSecret{
					SecretKeySelector: corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: testSecretNameForValidating,
						},
						Key: testKeyForValidating,
					},
					Create:     &create,
					EnvVarKeys: boolPtr(true),
				}
				obj.Spec.ClientSecretRef = &secretRef

				By("Validating creation should succeed")
				warnings, err := validator.ValidateCreate(ctx, obj)
				Expect(warnings).To(BeNil())
				Expect(err).ToNot(HaveOccurred())
			})

			It("Should deny creation if ClientSecretRef name is empty", func() {
				By("Setting up client with client-secret auth type, public=false, and ClientSecretRef with empty name")
				obj.Spec.ClientID = &clientIDWithPrefix
				obj.Spec.Realm = &testRealm
				obj.Spec.ClientAuthenticatorType = &clientSecretType
				public := false
				obj.Spec.PublicClient = &public

				// Set a ClientSecretRef with empty name
				secretRef := keycloakv1alpha1.KeycloakClientSecret{
					SecretKeySelector: corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "",
						},
						Key: "test-key",
					},
				}
				obj.Spec.ClientSecretRef = &secretRef

				By("Validating creation should fail due to empty name")
				warnings, err := validator.ValidateCreate(ctx, obj)
				Expect(warnings).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("clientSecretRef name must be set when clientAuthenticatorType is client-secret and public is false"))
				Expect(err.Error()).To(ContainSubstring("clientSecretRef create must be set when clientAuthenticatorType is client-secret and public is false"))
				Expect(err.Error()).To(ContainSubstring("clientSecretRef envVarKeys must be set when clientAuthenticatorType is client-secret and public is false"))
			})

			It("Should deny creation if ClientSecretRef key is empty", func() {
				By("Setting up client with client-secret auth type, public=false, and ClientSecretRef with empty key")
				obj.Spec.ClientID = &clientIDWithPrefix
				obj.Spec.Realm = &testRealm
				obj.Spec.ClientAuthenticatorType = &clientSecretType
				public := false
				obj.Spec.PublicClient = &public

				// Set a ClientSecretRef with empty key
				secretRef := keycloakv1alpha1.KeycloakClientSecret{
					SecretKeySelector: corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: testSecretNameForValidating,
						},
						Key: "",
					},
				}
				obj.Spec.ClientSecretRef = &secretRef

				By("Validating creation should fail due to empty key")
				warnings, err := validator.ValidateCreate(ctx, obj)
				Expect(warnings).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("clientSecretRef key must be set when clientAuthenticatorType is client-secret and public is false"))
				Expect(err.Error()).To(ContainSubstring("clientSecretRef create must be set when clientAuthenticatorType is client-secret and public is false"))
				Expect(err.Error()).To(ContainSubstring("clientSecretRef envVarKeys must be set when clientAuthenticatorType is client-secret and public is false"))
			})

			It("Should deny creation if ClientSecretRef key is not upper snake case when EnvVarKeys is true", func() {
				By("Setting up client with client-secret auth type, public=false, and ClientSecretRef with non-upper-snake-case key")
				obj.Spec.ClientID = &clientIDWithPrefix
				obj.Spec.Realm = &testRealm
				obj.Spec.ClientAuthenticatorType = &clientSecretType
				public := false
				obj.Spec.PublicClient = &public
				create := true
				obj.Spec.ConfigMap = defaultConfigMap

				// Set a ClientSecretRef with key that is not upper snake case
				secretRef := keycloakv1alpha1.KeycloakClientSecret{
					SecretKeySelector: corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: testSecretNameForValidating,
						},
						Key: "testKey", // Not upper snake case
					},
					Create:     &create,
					EnvVarKeys: boolPtr(true), // EnvVarKeys is true
				}
				obj.Spec.ClientSecretRef = &secretRef

				By("Validating creation should fail due to non-upper-snake-case key")
				warnings, err := validator.ValidateCreate(ctx, obj)
				Expect(warnings).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("clientSecretRef key must be upper snake case when envVarKeys is true"))
			})

			It("Should allow creation if ClientSecretRef key is upper snake case when EnvVarKeys is true", func() {
				By("Setting up client with client-secret auth type, public=false, and ClientSecretRef with upper snake case key")
				obj.Spec.ClientID = &clientIDWithPrefix
				obj.Spec.Realm = &testRealm
				obj.Spec.ClientAuthenticatorType = &clientSecretType
				public := false
				obj.Spec.PublicClient = &public
				create := true
				obj.Spec.ConfigMap = defaultConfigMap

				// Set a ClientSecretRef with key that is upper snake case
				secretRef := keycloakv1alpha1.KeycloakClientSecret{
					SecretKeySelector: corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: testSecretNameForValidating,
						},
						Key: testKeyForValidating, // Upper snake case
					},
					Create:     &create,
					EnvVarKeys: boolPtr(true), // EnvVarKeys is true
				}
				obj.Spec.ClientSecretRef = &secretRef

				By("Validating creation should succeed")
				warnings, err := validator.ValidateCreate(ctx, obj)
				Expect(warnings).To(BeNil())
				Expect(err).ToNot(HaveOccurred())
			})

			It("Should allow creation if ClientSecretRef key is not upper snake case when EnvVarKeys is false", func() {
				By("Setting up client with client-secret auth type, public=false, and ClientSecretRef with non-upper-snake-case key")
				obj.Spec.ClientID = &clientIDWithPrefix
				obj.Spec.Realm = &testRealm
				obj.Spec.ClientAuthenticatorType = &clientSecretType
				public := false
				obj.Spec.PublicClient = &public
				create := true
				obj.Spec.ConfigMap = defaultConfigMap

				// Set a ClientSecretRef with key that is not upper snake case but EnvVarKeys is false
				secretRef := keycloakv1alpha1.KeycloakClientSecret{
					SecretKeySelector: corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: testSecretNameForValidating,
						},
						Key: "testKey", // Not upper snake case
					},
					Create:     &create,
					EnvVarKeys: boolPtr(false), // EnvVarKeys is false
				}
				obj.Spec.ClientSecretRef = &secretRef

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
			obj.Spec.ProtocolMappers = []*keycloakv1alpha1.KeycloakClientProtocolMapper{}

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
			obj.Spec.ProtocolMappers = []*keycloakv1alpha1.KeycloakClientProtocolMapper{
				{
					Name:     nil, // Missing name
					Type:     stringPtr("oidc-hardcoded-claim-mapper"),
					Protocol: stringPtr("openid-connect"),
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
			obj.Spec.ProtocolMappers = []*keycloakv1alpha1.KeycloakClientProtocolMapper{
				{
					Name:     stringPtr(""), // Empty name
					Type:     stringPtr("oidc-hardcoded-claim-mapper"),
					Protocol: stringPtr("openid-connect"),
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
			obj.Spec.ProtocolMappers = []*keycloakv1alpha1.KeycloakClientProtocolMapper{
				{
					Name:     stringPtr("test-mapper"),
					Type:     nil, // Missing type
					Protocol: stringPtr("openid-connect"),
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
			obj.Spec.ProtocolMappers = []*keycloakv1alpha1.KeycloakClientProtocolMapper{
				{
					Name:     stringPtr("test-mapper"),
					Type:     stringPtr(""), // Empty type
					Protocol: stringPtr("openid-connect"),
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
			obj.Spec.ProtocolMappers = []*keycloakv1alpha1.KeycloakClientProtocolMapper{
				{
					Name:     stringPtr("test-mapper"),
					Type:     stringPtr("oidc-hardcoded-claim-mapper"),
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
			obj.Spec.ProtocolMappers = []*keycloakv1alpha1.KeycloakClientProtocolMapper{
				{
					Name:     stringPtr("test-mapper"),
					Type:     stringPtr("oidc-hardcoded-claim-mapper"),
					Protocol: stringPtr(""), // Empty protocol
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
			obj.Spec.ProtocolMappers = []*keycloakv1alpha1.KeycloakClientProtocolMapper{
				{
					Name:     stringPtr("test-mapper"),
					Type:     stringPtr("oidc-hardcoded-claim-mapper"),
					Protocol: stringPtr("openid-connect"),
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
			obj.Spec.ProtocolMappers = []*keycloakv1alpha1.KeycloakClientProtocolMapper{
				{
					Name:     stringPtr("audience-mapper"),
					Type:     stringPtr("oidc-audience-mapper"),
					Protocol: stringPtr("openid-connect"),
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
			obj.Spec.ProtocolMappers = []*keycloakv1alpha1.KeycloakClientProtocolMapper{
				{
					Name:                   stringPtr("audience-mapper"),
					Type:                   stringPtr("oidc-audience-mapper"),
					Protocol:               stringPtr("openid-connect"),
					IncludedClientAudience: stringPtr(""), // Empty includedClientAudience
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
			obj.Spec.ProtocolMappers = []*keycloakv1alpha1.KeycloakClientProtocolMapper{
				{
					Name:                   stringPtr("audience-mapper"),
					Type:                   stringPtr("oidc-audience-mapper"),
					Protocol:               stringPtr("openid-connect"),
					IncludedClientAudience: stringPtr("test-client"),
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
			obj.Spec.ProtocolMappers = []*keycloakv1alpha1.KeycloakClientProtocolMapper{
				{
					Name:     stringPtr("mapper1"),
					Type:     stringPtr("oidc-hardcoded-claim-mapper"),
					Protocol: stringPtr("openid-connect"),
				},
				{
					Name:                   stringPtr("mapper2"),
					Type:                   stringPtr("oidc-audience-mapper"),
					Protocol:               stringPtr("openid-connect"),
					IncludedClientAudience: stringPtr("test-client"),
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
			obj.Spec.ProtocolMappers = []*keycloakv1alpha1.KeycloakClientProtocolMapper{
				{
					Name:     stringPtr("test-mapper"),
					Type:     stringPtr("oidc-hardcoded-claim-mapper"),
					Protocol: stringPtr("openid-connect"),
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
			obj.Spec.ProtocolMappers = []*keycloakv1alpha1.KeycloakClientProtocolMapper{
				{
					Name:     nil, // Missing name
					Type:     stringPtr("oidc-hardcoded-claim-mapper"),
					Protocol: stringPtr("openid-connect"),
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
