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

	keycloakv1alpha2 "github.com/OSC/keycloak-cr-operator/api/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

const (
	testNamespace   = "default"
	testSecretName  = "test-secret"
	clientSecretKey = "client-secret"
)

var _ = Describe("KeycloakClient Conversion", func() {
	Context("ConvertTo", func() {
		It("should convert v1alpha1 KeycloakClient to v1alpha2 correctly", func() {
			// Create a v1alpha1 KeycloakClient with all fields set
			src := &KeycloakClient{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-client",
					Namespace: testNamespace,
				},
				Spec: KeycloakClientSpec{
					ClientID:                     stringPtr("test-client-id"),
					Realm:                        stringPtr("master"),
					Description:                  stringPtr("Test client"),
					Enabled:                      boolPtr(true),
					PublicClient:                 boolPtr(false),
					Protocol:                     stringPtr("openid-connect"),
					StandardFlowEnabled:          boolPtr(true),
					DirectAccessGrantsEnabled:    boolPtr(true),
					WebOrigins:                   &[]string{"https://example.com"},
					RedirectURIs:                 &[]string{"https://example.com/*"},
					DefaultClientScopes:          &[]string{"profile", "email"},
					LoginTheme:                   stringPtr("keycloak"),
					AdminURL:                     stringPtr("https://admin.example.com"),
					BaseURL:                      stringPtr("https://example.com"),
					RootURL:                      stringPtr("https://root.example.com"),
					ConsentRequired:              boolPtr(false),
					FullScopeAllowed:             boolPtr(true),
					AuthorizationServicesEnabled: boolPtr(false),
					FrontChannelLogout:           boolPtr(true),
					ImplicitFlowEnabled:          boolPtr(false),
					Name:                         stringPtr("Test Client"),
					NodeReRegistrationTimeout:    int32Ptr(300),
					NotBefore:                    int32Ptr(0),
					OptionalClientScopes:         &[]string{"phone"},
					Origin:                       stringPtr("https://example.com"),
					RegistrationAccessToken:      stringPtr("reg-token"),
					ServiceAccountsEnabled:       boolPtr(false),
					SurrogateAuthRequired:        boolPtr(false),
				},
			}

			// Set up ClientSecretRef with envVarKeys=true
			create := true
			envVarKeys := true
			keyPrefix := ""
			src.Spec.ClientSecretRef = &KeycloakClientSecret{
				SecretKeySelector: corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: testSecretName,
					},
					Key: clientSecretKey,
				},
				Create:     &create,
				EnvVarKeys: &envVarKeys,
				KeyPrefix:  &keyPrefix,
			}

			// Set up ConfigMap
			configMapName := "test-configmap"
			src.Spec.ConfigMap = &KeycloakClientConfigMap{
				Name:       &configMapName,
				EnvVarKeys: &envVarKeys,
			}

			// Set up ProtocolMappers
			idTokenClaim := true
			accessTokenClaim := true
			consentRequired := false
			src.Spec.ProtocolMappers = []*KeycloakClientProtocolMapper{
				{
					Name:             stringPtr("client roles"),
					Protocol:         stringPtr("openid-connect"),
					Type:             stringPtr("client-roles"),
					IDTokenClaim:     &idTokenClaim,
					AccessTokenClaim: &accessTokenClaim,
					ConsentRequired:  &consentRequired,
					Config:           &map[string]string{"claim.name": "roles", "jsonType.label": "String", "user.attribute": "foo", "multivalued": "true"},
				},
			}

			// Convert to v1alpha2
			dst := &keycloakv1alpha2.KeycloakClient{}
			err := src.ConvertTo(dstRaw(dst))
			Expect(err).NotTo(HaveOccurred())

			// Verify ObjectMeta
			Expect(dst.Name).To(Equal("test-client"))
			Expect(dst.Namespace).To(Equal("default"))

			// Verify all spec fields are converted
			Expect(dst.Spec.ClientID).To(Equal(src.Spec.ClientID))
			Expect(dst.Spec.Realm).To(Equal(src.Spec.Realm))
			Expect(dst.Spec.Description).To(Equal(src.Spec.Description))
			Expect(dst.Spec.Enabled).To(Equal(src.Spec.Enabled))
			Expect(dst.Spec.PublicClient).To(Equal(src.Spec.PublicClient))
			Expect(dst.Spec.Protocol).To(Equal(src.Spec.Protocol))
			Expect(dst.Spec.StandardFlowEnabled).To(Equal(src.Spec.StandardFlowEnabled))
			Expect(dst.Spec.DirectAccessGrantsEnabled).To(Equal(src.Spec.DirectAccessGrantsEnabled))
			Expect(dst.Spec.WebOrigins).To(Equal(src.Spec.WebOrigins))
			Expect(dst.Spec.RedirectURIs).To(Equal(src.Spec.RedirectURIs))
			Expect(dst.Spec.DefaultClientScopes).To(Equal(src.Spec.DefaultClientScopes))
			Expect(dst.Spec.LoginTheme).To(Equal(src.Spec.LoginTheme))
			Expect(dst.Spec.AdminURL).To(Equal(src.Spec.AdminURL))
			Expect(dst.Spec.BaseURL).To(Equal(src.Spec.BaseURL))
			Expect(dst.Spec.RootURL).To(Equal(src.Spec.RootURL))
			Expect(dst.Spec.ConsentRequired).To(Equal(src.Spec.ConsentRequired))
			Expect(dst.Spec.FullScopeAllowed).To(Equal(src.Spec.FullScopeAllowed))
			Expect(dst.Spec.AuthorizationServicesEnabled).To(Equal(src.Spec.AuthorizationServicesEnabled))
			Expect(dst.Spec.FrontChannelLogout).To(Equal(src.Spec.FrontChannelLogout))
			Expect(dst.Spec.ImplicitFlowEnabled).To(Equal(src.Spec.ImplicitFlowEnabled))
			Expect(dst.Spec.Name).To(Equal(src.Spec.Name))
			Expect(dst.Spec.NodeReRegistrationTimeout).To(Equal(src.Spec.NodeReRegistrationTimeout))
			Expect(dst.Spec.NotBefore).To(Equal(src.Spec.NotBefore))
			Expect(dst.Spec.OptionalClientScopes).To(Equal(src.Spec.OptionalClientScopes))
			Expect(dst.Spec.Origin).To(Equal(src.Spec.Origin))
			Expect(dst.Spec.RegistrationAccessToken).To(Equal(src.Spec.RegistrationAccessToken))
			Expect(dst.Spec.ServiceAccountsEnabled).To(Equal(src.Spec.ServiceAccountsEnabled))
			Expect(dst.Spec.SurrogateAuthRequired).To(Equal(src.Spec.SurrogateAuthRequired))

			// Verify Secret conversion
			Expect(dst.Spec.Secret).NotTo(BeNil())
			Expect(*dst.Spec.Secret.Name).To(Equal("test-secret"))
			Expect(*dst.Spec.Secret.ClientSecretKey).To(Equal("client-secret"))
			// clientIdKey and issuerUrlKey should be upper snake case when envVarKeys=true
			Expect(*dst.Spec.Secret.ClientIdKey).To(Equal("CLIENT_ID"))
			Expect(*dst.Spec.Secret.IssuerUrlKey).To(Equal("ISSUER_URL"))
			Expect(*dst.Spec.Secret.Create).To(BeTrue())
			Expect(*dst.Spec.Secret.EnvVarKeys).To(BeTrue())
			Expect(*dst.Spec.Secret.KeyPrefix).To(Equal(""))

			// Verify ConfigMap conversion
			Expect(dst.Spec.ConfigMap).NotTo(BeNil())
			Expect(*dst.Spec.ConfigMap.Name).To(Equal("test-configmap"))
			Expect(*dst.Spec.ConfigMap.EnvVarKeys).To(BeTrue())

			// Verify ProtocolMappers conversion
			Expect(dst.Spec.ProtocolMappers).NotTo(BeNil())
			Expect(dst.Spec.ProtocolMappers).To(HaveLen(1))
			Expect(*dst.Spec.ProtocolMappers[0].Name).To(Equal("client roles"))
			Expect(*dst.Spec.ProtocolMappers[0].Protocol).To(Equal("openid-connect"))
			Expect(*dst.Spec.ProtocolMappers[0].Type).To(Equal("client-roles"))
			Expect(*dst.Spec.ProtocolMappers[0].IDTokenClaim).To(BeTrue())
			Expect(*dst.Spec.ProtocolMappers[0].AccessTokenClaim).To(BeTrue())
			Expect(*dst.Spec.ProtocolMappers[0].ConsentRequired).To(BeFalse())
			Expect(*dst.Spec.ProtocolMappers[0].Config).To(HaveKey("claim.name"))
			Expect(*dst.Spec.ProtocolMappers[0].Config).To(HaveKey("jsonType.label"))
		})

		It("should handle nil ClientSecretRef", func() {
			src := &KeycloakClient{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-client-no-secret",
					Namespace: testNamespace,
				},
				Spec: KeycloakClientSpec{
					ClientID: stringPtr("test-client-id"),
					Realm:    stringPtr("master"),
				},
			}

			dst := &keycloakv1alpha2.KeycloakClient{}
			err := src.ConvertTo(dstRaw(dst))
			Expect(err).NotTo(HaveOccurred())
			Expect(dst.Spec.Secret).To(BeNil())
		})

		It("should handle nil ConfigMap", func() {
			src := &KeycloakClient{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-client-no-configmap",
					Namespace: testNamespace,
				},
				Spec: KeycloakClientSpec{
					ClientID: stringPtr("test-client-id"),
					Realm:    stringPtr("master"),
				},
			}

			dst := &keycloakv1alpha2.KeycloakClient{}
			err := src.ConvertTo(dstRaw(dst))
			Expect(err).NotTo(HaveOccurred())
			Expect(dst.Spec.ConfigMap).To(BeNil())
		})

		It("should handle nil ProtocolMappers", func() {
			src := &KeycloakClient{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-client-no-mappers",
					Namespace: testNamespace,
				},
				Spec: KeycloakClientSpec{
					ClientID: stringPtr("test-client-id"),
					Realm:    stringPtr("master"),
				},
			}

			dst := &keycloakv1alpha2.KeycloakClient{}
			err := src.ConvertTo(dstRaw(dst))
			Expect(err).NotTo(HaveOccurred())
			Expect(dst.Spec.ProtocolMappers).To(BeNil())
		})

		It("should derive correct clientIdKey and issuerUrlKey based on EnvVarKeys=false", func() {
			create := true
			envVarKeysFalse := false
			src := &KeycloakClient{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-client-no-envvars",
					Namespace: testNamespace,
				},
				Spec: KeycloakClientSpec{
					ClientID: stringPtr("test-client-id"),
					Realm:    stringPtr("master"),
				},
			}
			src.Spec.ClientSecretRef = &KeycloakClientSecret{
				SecretKeySelector: corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: testSecretName,
					},
					Key: "client-secret",
				},
				Create:     &create,
				EnvVarKeys: &envVarKeysFalse,
			}

			dst := &keycloakv1alpha2.KeycloakClient{}
			err := src.ConvertTo(dstRaw(dst))
			Expect(err).NotTo(HaveOccurred())

			Expect(dst.Spec.Secret).NotTo(BeNil())
			Expect(*dst.Spec.Secret.ClientIdKey).To(Equal("client-id"))
			Expect(*dst.Spec.Secret.IssuerUrlKey).To(Equal("issuer-url"))
		})

		It("should handle nil values in ClientSecretRef", func() {
			src := &KeycloakClient{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-client-nil-values",
					Namespace: testNamespace,
				},
				Spec: KeycloakClientSpec{
					ClientID: stringPtr("test-client-id"),
					Realm:    stringPtr("master"),
				},
			}
			// Create ClientSecretRef with all nil fields
			src.Spec.ClientSecretRef = &KeycloakClientSecret{}

			dst := &keycloakv1alpha2.KeycloakClient{}
			err := src.ConvertTo(dstRaw(dst))
			Expect(err).NotTo(HaveOccurred())

			Expect(dst.Spec.Secret).NotTo(BeNil())
			// Name should be empty string since source was empty struct
			Expect(*dst.Spec.Secret.Name).To(Equal(""))
			Expect(*dst.Spec.Secret.ClientSecretKey).To(Equal(""))
			// clientIdKey and issuerUrlKey should have default values (upper snake case)
			Expect(*dst.Spec.Secret.ClientIdKey).To(Equal("CLIENT_ID"))
			Expect(*dst.Spec.Secret.IssuerUrlKey).To(Equal("ISSUER_URL"))
		})
	})

	Context("ConvertFrom", func() {
		It("should convert v1alpha2 KeycloakClient to v1alpha1 correctly", func() {
			// Create a v1alpha2 KeycloakClient with all fields set
			src := &keycloakv1alpha2.KeycloakClient{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-client",
					Namespace: testNamespace,
				},
				Spec: keycloakv1alpha2.KeycloakClientSpec{
					ClientID:                     stringPtr("test-client-id"),
					Realm:                        stringPtr("master"),
					Description:                  stringPtr("Test client"),
					Enabled:                      boolPtr(true),
					PublicClient:                 boolPtr(false),
					Protocol:                     stringPtr("openid-connect"),
					StandardFlowEnabled:          boolPtr(true),
					DirectAccessGrantsEnabled:    boolPtr(true),
					WebOrigins:                   &[]string{"https://example.com"},
					RedirectURIs:                 &[]string{"https://example.com/*"},
					DefaultClientScopes:          &[]string{"profile", "email"},
					LoginTheme:                   stringPtr("keycloak"),
					AdminURL:                     stringPtr("https://admin.example.com"),
					BaseURL:                      stringPtr("https://example.com"),
					RootURL:                      stringPtr("https://root.example.com"),
					ConsentRequired:              boolPtr(false),
					FullScopeAllowed:             boolPtr(true),
					AuthorizationServicesEnabled: boolPtr(false),
					FrontChannelLogout:           boolPtr(true),
					ImplicitFlowEnabled:          boolPtr(false),
					Name:                         stringPtr("Test Client"),
					NodeReRegistrationTimeout:    int32Ptr(300),
					NotBefore:                    int32Ptr(0),
					OptionalClientScopes:         &[]string{"phone"},
					Origin:                       stringPtr("https://example.com"),
					RegistrationAccessToken:      stringPtr("reg-token"),
					ServiceAccountsEnabled:       boolPtr(false),
					SurrogateAuthRequired:        boolPtr(false),
				},
			}

			// Set up Secret
			clientSecretKey := "CLIENT_SECRET"
			clientIdKey := "CLIENT_ID"
			issuerUrlKey := "ISSUER_URL"
			create := true
			envVarKeys := true
			keyPrefix := ""
			src.Spec.Secret = &keycloakv1alpha2.KeycloakClientSecret{
				Name:            stringPtr("test-secret"),
				ClientSecretKey: stringPtr(clientSecretKey),
				ClientIdKey:     stringPtr(clientIdKey),
				IssuerUrlKey:    stringPtr(issuerUrlKey),
				Create:          &create,
				EnvVarKeys:      &envVarKeys,
				KeyPrefix:       stringPtr(keyPrefix),
			}

			// Set up ConfigMap
			configMapName := "test-configmap"
			src.Spec.ConfigMap = &keycloakv1alpha2.KeycloakClientConfigMap{
				Name:       stringPtr(configMapName),
				EnvVarKeys: &envVarKeys,
			}

			// Set up ProtocolMappers
			idTokenClaim := true
			accessTokenClaim := true
			consentRequired := false
			src.Spec.ProtocolMappers = []*keycloakv1alpha2.KeycloakClientProtocolMapper{
				{
					Name:             stringPtr("client roles"),
					Protocol:         stringPtr("openid-connect"),
					Type:             stringPtr("client-roles"),
					IDTokenClaim:     &idTokenClaim,
					AccessTokenClaim: &accessTokenClaim,
					ConsentRequired:  &consentRequired,
					Config:           &map[string]string{"claim.name": "roles", "jsonType.label": "String", "user.attribute": "foo", "multivalued": "true"},
				},
			}

			// Convert to v1alpha1
			dst := &KeycloakClient{}
			err := dst.ConvertFrom(srcRaw(src))
			Expect(err).NotTo(HaveOccurred())

			// Verify ObjectMeta
			Expect(dst.Name).To(Equal("test-client"))
			Expect(dst.Namespace).To(Equal("default"))

			// Verify all spec fields are converted
			Expect(dst.Spec.ClientID).To(Equal(src.Spec.ClientID))
			Expect(dst.Spec.Realm).To(Equal(src.Spec.Realm))
			Expect(dst.Spec.Description).To(Equal(src.Spec.Description))
			Expect(dst.Spec.Enabled).To(Equal(src.Spec.Enabled))
			Expect(dst.Spec.PublicClient).To(Equal(src.Spec.PublicClient))
			Expect(dst.Spec.Protocol).To(Equal(src.Spec.Protocol))
			Expect(dst.Spec.StandardFlowEnabled).To(Equal(src.Spec.StandardFlowEnabled))
			Expect(dst.Spec.DirectAccessGrantsEnabled).To(Equal(src.Spec.DirectAccessGrantsEnabled))
			Expect(dst.Spec.WebOrigins).To(Equal(src.Spec.WebOrigins))
			Expect(dst.Spec.RedirectURIs).To(Equal(src.Spec.RedirectURIs))
			Expect(dst.Spec.DefaultClientScopes).To(Equal(src.Spec.DefaultClientScopes))
			Expect(dst.Spec.LoginTheme).To(Equal(src.Spec.LoginTheme))
			Expect(dst.Spec.AdminURL).To(Equal(src.Spec.AdminURL))
			Expect(dst.Spec.BaseURL).To(Equal(src.Spec.BaseURL))
			Expect(dst.Spec.RootURL).To(Equal(src.Spec.RootURL))
			Expect(dst.Spec.ConsentRequired).To(Equal(src.Spec.ConsentRequired))
			Expect(dst.Spec.FullScopeAllowed).To(Equal(src.Spec.FullScopeAllowed))
			Expect(dst.Spec.AuthorizationServicesEnabled).To(Equal(src.Spec.AuthorizationServicesEnabled))
			Expect(dst.Spec.FrontChannelLogout).To(Equal(src.Spec.FrontChannelLogout))
			Expect(dst.Spec.ImplicitFlowEnabled).To(Equal(src.Spec.ImplicitFlowEnabled))
			Expect(dst.Spec.Name).To(Equal(src.Spec.Name))
			Expect(dst.Spec.NodeReRegistrationTimeout).To(Equal(src.Spec.NodeReRegistrationTimeout))
			Expect(dst.Spec.NotBefore).To(Equal(src.Spec.NotBefore))
			Expect(dst.Spec.OptionalClientScopes).To(Equal(src.Spec.OptionalClientScopes))
			Expect(dst.Spec.Origin).To(Equal(src.Spec.Origin))
			Expect(dst.Spec.RegistrationAccessToken).To(Equal(src.Spec.RegistrationAccessToken))
			Expect(dst.Spec.ServiceAccountsEnabled).To(Equal(src.Spec.ServiceAccountsEnabled))
			Expect(dst.Spec.SurrogateAuthRequired).To(Equal(src.Spec.SurrogateAuthRequired))

			// Verify ClientSecretRef conversion
			Expect(dst.Spec.ClientSecretRef).NotTo(BeNil())
			Expect(dst.Spec.ClientSecretRef.Name).To(Equal("test-secret"))
			Expect(dst.Spec.ClientSecretRef.Key).To(Equal("CLIENT_SECRET"))
			Expect(*dst.Spec.ClientSecretRef.Create).To(BeTrue())
			Expect(*dst.Spec.ClientSecretRef.EnvVarKeys).To(BeTrue())
			Expect(*dst.Spec.ClientSecretRef.KeyPrefix).To(Equal(""))

			// Verify ConfigMap conversion
			Expect(dst.Spec.ConfigMap).NotTo(BeNil())
			Expect(*dst.Spec.ConfigMap.Name).To(Equal("test-configmap"))
			Expect(*dst.Spec.ConfigMap.EnvVarKeys).To(BeTrue())

			// Verify ProtocolMappers conversion
			Expect(dst.Spec.ProtocolMappers).NotTo(BeNil())
			Expect(dst.Spec.ProtocolMappers).To(HaveLen(1))
			Expect(*dst.Spec.ProtocolMappers[0].Name).To(Equal("client roles"))
			Expect(*dst.Spec.ProtocolMappers[0].Protocol).To(Equal("openid-connect"))
			Expect(*dst.Spec.ProtocolMappers[0].Type).To(Equal("client-roles"))
			Expect(*dst.Spec.ProtocolMappers[0].IDTokenClaim).To(BeTrue())
			Expect(*dst.Spec.ProtocolMappers[0].AccessTokenClaim).To(BeTrue())
			Expect(*dst.Spec.ProtocolMappers[0].ConsentRequired).To(BeFalse())
		})

		It("should handle nil Secret", func() {
			src := &keycloakv1alpha2.KeycloakClient{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-client-no-secret",
					Namespace: testNamespace,
				},
				Spec: keycloakv1alpha2.KeycloakClientSpec{
					ClientID: stringPtr("test-client-id"),
					Realm:    stringPtr("master"),
				},
			}

			dst := &KeycloakClient{}
			err := dst.ConvertFrom(srcRaw(src))
			Expect(err).NotTo(HaveOccurred())
			Expect(dst.Spec.ClientSecretRef).To(BeNil())
		})

		It("should handle nil ConfigMap", func() {
			src := &keycloakv1alpha2.KeycloakClient{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-client-no-configmap",
					Namespace: testNamespace,
				},
				Spec: keycloakv1alpha2.KeycloakClientSpec{
					ClientID: stringPtr("test-client-id"),
					Realm:    stringPtr("master"),
				},
			}

			dst := &KeycloakClient{}
			err := dst.ConvertFrom(srcRaw(src))
			Expect(err).NotTo(HaveOccurred())
			Expect(dst.Spec.ConfigMap).To(BeNil())
		})

		It("should handle nil ProtocolMappers", func() {
			src := &keycloakv1alpha2.KeycloakClient{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-client-no-mappers",
					Namespace: testNamespace,
				},
				Spec: keycloakv1alpha2.KeycloakClientSpec{
					ClientID: stringPtr("test-client-id"),
					Realm:    stringPtr("master"),
				},
			}

			dst := &KeycloakClient{}
			err := dst.ConvertFrom(srcRaw(src))
			Expect(err).NotTo(HaveOccurred())
			Expect(dst.Spec.ProtocolMappers).To(BeNil())
		})

		It("should handle empty Secret fields", func() {
			src := &keycloakv1alpha2.KeycloakClient{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-client-empty-secret",
					Namespace: testNamespace,
				},
				Spec: keycloakv1alpha2.KeycloakClientSpec{
					ClientID: stringPtr("test-client-id"),
					Realm:    stringPtr("master"),
				},
			}
			// Create Secret with nil Name and Key
			src.Spec.Secret = &keycloakv1alpha2.KeycloakClientSecret{}

			dst := &KeycloakClient{}
			err := dst.ConvertFrom(srcRaw(src))
			Expect(err).NotTo(HaveOccurred())

			Expect(dst.Spec.ClientSecretRef).NotTo(BeNil())
			// Name should be empty string since source was nil
			Expect(dst.Spec.ClientSecretRef.Name).To(Equal(""))
			Expect(dst.Spec.ClientSecretRef.Key).To(Equal(""))
			// Create, EnvVarKeys, KeyPrefix should have their zero values
			Expect(dst.Spec.ClientSecretRef.Create).To(BeNil())
			Expect(dst.Spec.ClientSecretRef.EnvVarKeys).To(BeNil())
			Expect(dst.Spec.ClientSecretRef.KeyPrefix).To(BeNil())
		})
	})

	Context("Bidirectional conversion", func() {
		It("should preserve data through ConvertTo and ConvertFrom", func() {
			// Create a v1alpha1 KeycloakClient
			create := true
			envVarKeys := true
			keyPrefix := "OIDC_"
			src := &KeycloakClient{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-bidirectional",
					Namespace: testNamespace,
				},
				Spec: KeycloakClientSpec{
					ClientID: stringPtr("test-client-id"),
					Realm:    stringPtr("master"),
					ConfigMap: &KeycloakClientConfigMap{
						Name:       stringPtr("test-configmap"),
						EnvVarKeys: &envVarKeys,
					},
				},
			}
			src.Spec.ClientSecretRef = &KeycloakClientSecret{
				SecretKeySelector: corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: testSecretName,
					},
					Key: "client-secret",
				},
				Create:     &create,
				EnvVarKeys: &envVarKeys,
				KeyPrefix:  stringPtr(keyPrefix),
			}

			// Convert to v1alpha2
			hub := &keycloakv1alpha2.KeycloakClient{}
			err := src.ConvertTo(dstRaw(hub))
			Expect(err).NotTo(HaveOccurred())

			// Convert back to v1alpha1
			dst := &KeycloakClient{}
			err = dst.ConvertFrom(srcRaw(hub))
			Expect(err).NotTo(HaveOccurred())

			// Verify all fields are preserved
			Expect(dst.Spec.ClientID).To(Equal(src.Spec.ClientID))
			Expect(dst.Spec.Realm).To(Equal(src.Spec.Realm))
			Expect(*dst.Spec.ConfigMap.Name).To(Equal(*src.Spec.ConfigMap.Name))
			Expect(*dst.Spec.ConfigMap.EnvVarKeys).To(Equal(*src.Spec.ConfigMap.EnvVarKeys))
			Expect(dst.Spec.ClientSecretRef.Name).To(Equal(src.Spec.ClientSecretRef.Name))
			Expect(dst.Spec.ClientSecretRef.Key).To(Equal(src.Spec.ClientSecretRef.Key))
			Expect(*dst.Spec.ClientSecretRef.Create).To(Equal(*src.Spec.ClientSecretRef.Create))
			Expect(*dst.Spec.ClientSecretRef.EnvVarKeys).To(Equal(*src.Spec.ClientSecretRef.EnvVarKeys))
			Expect(*dst.Spec.ClientSecretRef.KeyPrefix).To(Equal(*src.Spec.ClientSecretRef.KeyPrefix))
		})
	})
})

// Helper functions for conversion tests
func dstRaw(dst conversion.Hub) *keycloakv1alpha2.KeycloakClient {
	return dst.(*keycloakv1alpha2.KeycloakClient)
}

func srcRaw(src conversion.Hub) *keycloakv1alpha2.KeycloakClient {
	return src.(*keycloakv1alpha2.KeycloakClient)
}

func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func int32Ptr(i int32) *int32 {
	return &i
}
