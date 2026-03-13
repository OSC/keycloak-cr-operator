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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	keycloakv1alpha1 "github.com/OSC/keycloak-cr-operator/api/v1alpha1"
	"github.com/OSC/keycloak-cr-operator/internal/models"
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
		validator = KeycloakClientCustomValidator{}
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
	})

	Context("When creating or updating KeycloakClient under Validating Webhook", func() {
		// TODO (user): Add logic for validating webhooks
		// Example:
		// It("Should deny creation if a required field is missing", func() {
		//     By("simulating an invalid creation scenario")
		//     obj.SomeRequiredField = ""
		//     Expect(validator.ValidateCreate(ctx, obj)).Error().To(HaveOccurred())
		// })
		//
		// It("Should admit creation if all required fields are present", func() {
		//     By("simulating an invalid creation scenario")
		//     obj.SomeRequiredField = "valid_value"
		//     Expect(validator.ValidateCreate(ctx, obj)).To(BeNil())
		// })
		//
		// It("Should validate updates correctly", func() {
		//     By("simulating a valid update scenario")
		//     oldObj.SomeRequiredField = "updated_value"
		//     obj.SomeRequiredField = "updated_value"
		//     Expect(validator.ValidateUpdate(ctx, oldObj, obj)).To(BeNil())
		// })
	})

})
