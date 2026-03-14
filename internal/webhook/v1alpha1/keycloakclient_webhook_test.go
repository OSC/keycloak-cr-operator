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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	keycloakv1alpha1 "github.com/OSC/keycloak-cr-operator/api/v1alpha1"
	"github.com/OSC/keycloak-cr-operator/internal/models"
)

var (
	clientIDWithPrefix = "kubernetes-test-client"
	testRealm          = "test-realm"
)

var _ = Describe("KeycloakClient Webhook", func() {
	var (
		obj       *keycloakv1alpha1.KeycloakClient
		oldObj    *keycloakv1alpha1.KeycloakClient
		validator KeycloakClientCustomValidator
		defaulter KeycloakClientCustomDefaulter
	)

	BeforeEach(func() {
		obj = &keycloakv1alpha1.KeycloakClient{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-keycloak-client",
				Namespace: "test-namespace",
			},
		}
		oldObj = &keycloakv1alpha1.KeycloakClient{
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
		Expect(validator).NotTo(BeNil(), "Expected validator to be initialized")
		defaulter = KeycloakClientCustomDefaulter{
			keycloakConfig: &models.KeycloakConfig{
				DefaultRealm:   "master",
				ClientIDPrefix: "kubernetes",
			},
		}
		Expect(defaulter).NotTo(BeNil(), "Expected defaulter to be initialized")
		Expect(oldObj).NotTo(BeNil(), "Expected oldObj to be initialized")
		Expect(obj).NotTo(BeNil(), "Expected obj to be initialized")
	})

	AfterEach(func() {
		// TODO (user): Add any teardown logic common to all tests
	})

	Context("When creating KeycloakClient under Defaulting Webhook", func() {
		It("Should apply defaults for ClientID, Realm, and ID when not set", func() {
			By("Calling the Default method to apply defaults")
			err := defaulter.Default(ctx, obj)
			Expect(err).ToNot(HaveOccurred())

			By("Checking that the default ClientID is set")
			Expect(obj.Spec.ClientID).NotTo(BeNil())
			Expect(*obj.Spec.ClientID).To(Equal("kubernetes-test-namespace-test-keycloak-client"))

			By("Checking that the default Realm is set")
			Expect(obj.Spec.Realm).NotTo(BeNil())
			Expect(*obj.Spec.Realm).To(Equal("master"))

			By("Checking that the default ID is set to ClientID")
			Expect(obj.Spec.ID).NotTo(BeNil())
			Expect(*obj.Spec.ID).To(Equal(*obj.Spec.ClientID))
		})

		It("Should not override existing ClientID", func() {
			By("Setting an explicit ClientID")
			clientID := "existing-client-id"
			obj.Spec.ClientID = &clientID

			By("Calling the Default method to apply defaults")
			err := defaulter.Default(ctx, obj)
			Expect(err).ToNot(HaveOccurred())

			By("Checking that the existing ClientID is preserved")
			Expect(obj.Spec.ClientID).NotTo(BeNil())
			Expect(*obj.Spec.ClientID).To(Equal("existing-client-id"))
		})

		It("Should not override existing Realm", func() {
			By("Setting an explicit Realm")
			realm := "custom-realm"
			obj.Spec.Realm = &realm

			By("Calling the Default method to apply defaults")
			err := defaulter.Default(ctx, obj)
			Expect(err).ToNot(HaveOccurred())

			By("Checking that the existing Realm is preserved")
			Expect(obj.Spec.Realm).NotTo(BeNil())
			Expect(*obj.Spec.Realm).To(Equal("custom-realm"))
		})

		It("Should not override existing ID", func() {
			By("Setting an explicit ID")
			id := "existing-id"
			obj.Spec.ID = &id

			By("Calling the Default method to apply defaults")
			err := defaulter.Default(ctx, obj)
			Expect(err).ToNot(HaveOccurred())

			By("Checking that the existing ID is preserved")
			Expect(obj.Spec.ID).NotTo(BeNil())
			Expect(*obj.Spec.ID).To(Equal("existing-id"))
		})

		It("Should handle empty ClientIDPrefix", func() {
			By("Setting empty ClientIDPrefix")
			defaulter.keycloakConfig.ClientIDPrefix = ""

			By("Calling the Default method to apply defaults")
			err := defaulter.Default(ctx, obj)
			Expect(err).ToNot(HaveOccurred())

			By("Checking that the default ClientID is set without prefix")
			Expect(obj.Spec.ClientID).NotTo(BeNil())
			Expect(*obj.Spec.ClientID).To(Equal("test-namespace-test-keycloak-client"))
		})

		Context("When creating KeycloakClient under Defaulting Webhook - ClientSecretRef Defaulting", func() {
			It("Should set default ClientSecretRef when ClientAuthenticatorType is client-secret and PublicClient is false", func() {
				By("Setting up client with client-secret auth type and public=false")
				obj.Spec.ClientID = &clientIDWithPrefix
				obj.Spec.Realm = &testRealm
				obj.Spec.ClientAuthenticatorType = &clientSecretType
				public := false
				obj.Spec.PublicClient = &public

				By("Calling the Default method to apply defaults")
				err := defaulter.Default(ctx, obj)
				Expect(err).ToNot(HaveOccurred())

				By("Checking that ClientSecretRef is set")
				Expect(obj.Spec.ClientSecretRef).NotTo(BeNil())
				Expect(obj.Spec.ClientSecretRef.Name).To(Equal("test-keycloak-client-secret"))
				Expect(obj.Spec.ClientSecretRef.Key).To(Equal("client-secret"))
			})

			It("Should not set default ClientSecretRef when ClientAuthenticatorType is not client-secret", func() {
				By("Setting up client with different auth type and public=false")
				obj.Spec.ClientID = &clientIDWithPrefix
				obj.Spec.Realm = &testRealm
				authType := "other-auth-type"
				obj.Spec.ClientAuthenticatorType = &authType
				public := false
				obj.Spec.PublicClient = &public

				By("Calling the Default method to apply defaults")
				err := defaulter.Default(ctx, obj)
				Expect(err).ToNot(HaveOccurred())

				By("Checking that ClientSecretRef is not set")
				Expect(obj.Spec.ClientSecretRef).To(BeNil())
			})

			It("Should not set default ClientSecretRef when PublicClient is true", func() {
				By("Setting up client with client-secret auth type and public=true")
				obj.Spec.ClientID = &clientIDWithPrefix
				obj.Spec.Realm = &testRealm
				obj.Spec.ClientAuthenticatorType = &clientSecretType
				public := true
				obj.Spec.PublicClient = &public

				By("Calling the Default method to apply defaults")
				err := defaulter.Default(ctx, obj)
				Expect(err).ToNot(HaveOccurred())

				By("Checking that ClientSecretRef is not set")
				Expect(obj.Spec.ClientSecretRef).To(BeNil())
			})

			It("Should not override existing ClientSecretRef", func() {
				By("Setting up client with existing ClientSecretRef")
				obj.Spec.ClientID = &clientIDWithPrefix
				obj.Spec.Realm = &testRealm
				obj.Spec.ClientAuthenticatorType = &clientSecretType
				public := false
				obj.Spec.PublicClient = &public
				create := false

				// Set an existing ClientSecretRef
				existingRef := keycloakv1alpha1.KeycloakClientSecret{
					SecretKeySelector: corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "existing-secret",
						},
						Key: "existing-key",
					},
					Create: &create,
				}
				obj.Spec.ClientSecretRef = &existingRef

				By("Calling the Default method to apply defaults")
				err := defaulter.Default(ctx, obj)
				Expect(err).ToNot(HaveOccurred())

				By("Checking that existing ClientSecretRef is preserved")
				Expect(obj.Spec.ClientSecretRef).NotTo(BeNil())
				Expect(obj.Spec.ClientSecretRef.Name).To(Equal("existing-secret"))
				Expect(obj.Spec.ClientSecretRef.Key).To(Equal("existing-key"))
				Expect(*obj.Spec.ClientSecretRef.Create).To(BeFalse())
			})

			It("Should set default name when ClientSecretRef.Name is empty", func() {
				By("Setting up client with empty ClientSecretRef.Name")
				obj.Spec.ClientID = &clientIDWithPrefix
				obj.Spec.Realm = &testRealm
				obj.Spec.ClientAuthenticatorType = &clientSecretType
				public := false
				obj.Spec.PublicClient = &public

				// Set a ClientSecretRef with empty name
				obj.Spec.ClientSecretRef = &keycloakv1alpha1.KeycloakClientSecret{
					SecretKeySelector: corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "",
						},
						Key: "some-key",
					},
				}

				By("Calling the Default method to apply defaults")
				err := defaulter.Default(ctx, obj)
				Expect(err).ToNot(HaveOccurred())

				By("Checking that default name is set")
				Expect(obj.Spec.ClientSecretRef).NotTo(BeNil())
				Expect(obj.Spec.ClientSecretRef.Name).To(Equal("test-keycloak-client-secret"))
				Expect(*obj.Spec.ClientSecretRef.Create).To(BeTrue())
			})

			It("Should set default key when ClientSecretRef.Key is empty", func() {
				By("Setting up client with empty ClientSecretRef.Key")
				obj.Spec.ClientID = &clientIDWithPrefix
				obj.Spec.Realm = &testRealm
				obj.Spec.ClientAuthenticatorType = &clientSecretType
				public := false
				obj.Spec.PublicClient = &public

				// Set a ClientSecretRef with empty key
				obj.Spec.ClientSecretRef = &keycloakv1alpha1.KeycloakClientSecret{
					SecretKeySelector: corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "some-secret",
						},
						Key: "",
					},
				}

				By("Calling the Default method to apply defaults")
				err := defaulter.Default(ctx, obj)
				Expect(err).ToNot(HaveOccurred())

				By("Checking that default key is set")
				Expect(obj.Spec.ClientSecretRef).NotTo(BeNil())
				Expect(obj.Spec.ClientSecretRef.Key).To(Equal("client-secret"))
				Expect(*obj.Spec.ClientSecretRef.Create).To(BeTrue())
			})
		})
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
			realm := "master"
			obj.Spec.Realm = &realm

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

			// Create a fake secret reference
			secretRef := keycloakv1alpha1.KeycloakClientSecret{
				SecretKeySelector: corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "test-secret",
					},
					Key: "test-key",
				},
				Create: &create,
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

			By("Validating creation should succeed even without ClientSecretRef")
			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(warnings).To(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should validate updates correctly", func() {
			By("Setting up a valid client")
			obj.Spec.ClientID = &clientIDWithPrefix
			obj.Spec.Realm = &testRealm

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

				// Set a ClientSecretRef with key
				secretRef := keycloakv1alpha1.KeycloakClientSecret{
					SecretKeySelector: corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "test-secret",
						},
						Key: "test-key",
					},
					Create: &create,
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
							Name: "test-secret",
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
			})
		})
	})

})
