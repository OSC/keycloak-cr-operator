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

func WebhookDefaulting() {
	var (
		obj       *keycloakv1alpha1.KeycloakClient
		defaulter KeycloakClientCustomDefaulter
	)
	BeforeEach(func() {
		obj = &keycloakv1alpha1.KeycloakClient{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-keycloak-client",
				Namespace: "test-namespace",
			},
		}
		defaulter = KeycloakClientCustomDefaulter{
			keycloakConfig: &models.KeycloakConfig{
				DefaultRealm:   "master",
				ClientIDPrefix: "kubernetes",
			},
		}
		Expect(defaulter).NotTo(BeNil(), "Expected defaulter to be initialized")
		Expect(obj).NotTo(BeNil(), "Expected obj to be initialized")
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

		It("Should apply default ConfigMapName when not set", func() {
			By("Calling the Default method to apply defaults")
			err := defaulter.Default(ctx, obj)
			Expect(err).ToNot(HaveOccurred())

			By("Checking that the default ConfigMapName is set")
			Expect(obj.Spec.ConfigMapName).NotTo(BeNil())
			Expect(*obj.Spec.ConfigMapName).To(Equal("test-keycloak-client-config"))
		})

		It("Should not override existing ConfigMapName", func() {
			By("Setting an explicit ConfigMapName")
			configMapName := "existing-configmap"
			obj.Spec.ConfigMapName = &configMapName

			By("Calling the Default method to apply defaults")
			err := defaulter.Default(ctx, obj)
			Expect(err).ToNot(HaveOccurred())

			By("Checking that the existing ConfigMapName is preserved")
			Expect(obj.Spec.ConfigMapName).NotTo(BeNil())
			Expect(*obj.Spec.ConfigMapName).To(Equal("existing-configmap"))
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

		Context("When ClientIDRequired is set", func() {
			It("Should apply ClientIDRequired template to generate ClientID", func() {
				By("Setting up config with ClientIDRequired template")
				tmpl, err := template.New("clientID").Parse("{{.Config.ClientIDPrefix}}-{{.Obj.Namespace}}-{{.Obj.Name}}")
				Expect(err).ToNot(HaveOccurred())
				defaulter.keycloakConfig.ClientIDRequired = tmpl

				By("Calling the Default method to apply defaults")
				err = defaulter.Default(ctx, obj)
				Expect(err).ToNot(HaveOccurred())

				By("Checking that ClientID is set according to template")
				Expect(obj.Spec.ClientID).NotTo(BeNil())
				Expect(*obj.Spec.ClientID).To(Equal("kubernetes-test-namespace-test-keycloak-client"))
			})

			It("Should override existing ClientID with template result", func() {
				By("Setting up client with existing ClientID")
				clientID := "existing-client-id"
				obj.Spec.ClientID = &clientID

				By("Setting up config with ClientIDRequired template")
				tmpl, err := template.New("clientID").Parse("{{.Config.ClientIDPrefix}}-{{.Obj.Namespace}}-{{.Obj.Name}}")
				Expect(err).ToNot(HaveOccurred())
				defaulter.keycloakConfig.ClientIDRequired = tmpl

				By("Calling the Default method to apply defaults")
				err = defaulter.Default(ctx, obj)
				Expect(err).ToNot(HaveOccurred())

				By("Checking that ClientID is overridden by template")
				Expect(obj.Spec.ClientID).NotTo(BeNil())
				Expect(*obj.Spec.ClientID).To(Equal("kubernetes-test-namespace-test-keycloak-client"))
			})

			It("Should work with complex templates", func() {
				By("Setting up config with complex ClientIDRequired template")
				tmpl, err := template.New("clientID").Parse("prefix-{{.Obj.Namespace}}-{{.Obj.Name}}-suffix")
				Expect(err).ToNot(HaveOccurred())
				defaulter.keycloakConfig.ClientIDRequired = tmpl

				By("Calling the Default method to apply defaults")
				err = defaulter.Default(ctx, obj)
				Expect(err).ToNot(HaveOccurred())

				By("Checking that ClientID is set according to complex template")
				Expect(obj.Spec.ClientID).NotTo(BeNil())
				Expect(*obj.Spec.ClientID).To(Equal("prefix-test-namespace-test-keycloak-client-suffix"))
			})
		})
	})
}
