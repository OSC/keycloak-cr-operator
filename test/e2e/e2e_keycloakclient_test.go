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
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/OSC/keycloak-cr-operator/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	secretChecksum           string
	secretChecksumUpdated    string
	configmapChecksum        string
	configmapChecksumUpdated string
)

func KeycloakClientSpec() {
	Context("KeycloakClient", func() {
		It("should handle custom resources", func() {
			By("Apply custom KeycloakClient resource from samples")
			verifyKeycloakClientResource := func(g Gomega) {
				cmd := exec.Command("kubectl", "apply",
					"-f", keycloakClientManifest,
					"-f", keycloakClientManifestWithSecret,
					"-f", keycloakClientManifestPublic,
					"-f", keycloakClientManifestHeadlamp)
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(Or(ContainSubstring("created")))
				waitCmd := exec.Command("kubectl", "wait", "--for=condition=Available",
					"keycloakclient", "keycloakclient-test", "--timeout=20s")
				waitOut, waitErr := utils.Run(waitCmd)
				g.Expect(waitOut).To(ContainSubstring("condition met"))
				g.Expect(waitErr).NotTo(HaveOccurred())
				idCmd := exec.Command("kubectl", "get", "keycloakclient", "keycloakclient-test",
					"-o", "jsonpath={.status.id}")
				idOut, idErr := utils.Run(idCmd)
				g.Expect(idErr).NotTo(HaveOccurred())
				g.Expect(idOut).NotTo(BeEmpty())
				waitCmd = exec.Command("kubectl", "wait", "--for=condition=Available",
					"keycloakclient", "keycloakclient-sample", "--timeout=20s")
				waitOut, waitErr = utils.Run(waitCmd)
				g.Expect(waitOut).To(ContainSubstring("condition met"))
				g.Expect(waitErr).NotTo(HaveOccurred())
				idCmd = exec.Command("kubectl", "get", "keycloakclient", "keycloakclient-sample",
					"-o", "jsonpath={.status.id}")
				idOut, idErr = utils.Run(idCmd)
				g.Expect(idErr).NotTo(HaveOccurred())
				g.Expect(idOut).NotTo(BeEmpty())
			}
			Eventually(verifyKeycloakClientResource, 2*time.Minute).Should(Succeed())

			By("getting the metrics by checking for success")
			verifyMetricsSuccess := func(g Gomega) {
				metricsOutput, err := getMetricsOutput()
				g.Expect(err).NotTo(HaveOccurred(), "Failed to retrieve logs from curl pod")
				g.Expect(metricsOutput).To(MatchRegexp(`controller_runtime_reconcile_total\{controller="keycloakclient",result="success"\} [1-9]`))
			}
			Eventually(verifyMetricsSuccess, 2*time.Minute).Should(Succeed())
			By("Client exists in Keycloak")
			verifyClientExists := func(g Gomega) {
				client := getKeycloakClient("kubernetes-default-keycloakclient-test", "master")
				g.Expect(client).To(Not(BeNil()), "expected client found")
				g.Expect(*client.ClientID).To(Equal("kubernetes-default-keycloakclient-test"))
				g.Expect(*client.RedirectURIs).To(ConsistOf("https://example.com/*", "https://example.test.com/*"))
				g.Expect(*client.DefaultClientScopes).To(ConsistOf("web-origins", "profile", "email"))
				client = getKeycloakClient("kubernetes-keycloakclient-sample", "master")
				g.Expect(client).To(Not(BeNil()), "expected client found")
				g.Expect(*client.ClientID).To(Equal("kubernetes-keycloakclient-sample"))
				g.Expect(*client.Secret).To(Equal("sample-secret"))
				g.Expect(*client.RedirectURIs).To(ConsistOf("https://example.com/*", "https://example.test.com/*"))
				g.Expect(*client.DefaultClientScopes).To(ConsistOf("web-origins", "profile", "email"))
				client = getKeycloakClient("kubernetes-default-keycloakclient-headlamp", "master")
				g.Expect(*client.ProtocolMappers).To(HaveLen(1))
			}
			Eventually(verifyClientExists, 2*time.Minute).Should(Succeed())
			By("Keycloak client secrets handled")
			verifyClientSecrets := func(g Gomega) {
				client := getKeycloakClient("kubernetes-default-keycloakclient-test", "master")
				clientID, err := getSecret("keycloak-test", "CLIENT_ID")
				g.Expect(err).NotTo(HaveOccurred(), "Failed to retrieve secret")
				g.Expect(*client.ClientID).To(Equal(clientID))
				secret, err := getSecret("keycloak-test", "CLIENT_SECRET")
				g.Expect(err).NotTo(HaveOccurred(), "Failed to retrieve secret")
				g.Expect(*client.Secret).To(Equal(secret))
				secret, err = getSecret("keycloak-test", "COOKIE_SECRET")
				g.Expect(err).NotTo(HaveOccurred(), "Failed to retrieve cookie-secret")
				g.Expect(secret).NotTo(BeEmpty())

				client = getKeycloakClient("kubernetes-default-keycloakclient-test-public", "master")
				clientID, err = getSecret("keycloak-test-public", "CLIENT_ID")
				g.Expect(err).NotTo(HaveOccurred(), "Failed to retrieve secret")
				g.Expect(*client.ClientID).To(Equal(clientID))
				secret, err = getSecret("keycloak-test-public", "CLIENT_SECRET")
				g.Expect(err).NotTo(HaveOccurred(), "Failed to retrieve secret")
				g.Expect(secret).To(BeEmpty())

				client = getKeycloakClient("kubernetes-default-keycloakclient-headlamp", "master")
				clientID, err = getSecret("keycloak-headlamp", "OIDC_CLIENT_ID")
				g.Expect(err).NotTo(HaveOccurred(), "Failed to retrieve secret")
				g.Expect(*client.ClientID).To(Equal(clientID))
				secret, err = getSecret("keycloak-headlamp", "OIDC_CLIENT_SECRET")
				g.Expect(err).NotTo(HaveOccurred(), "Failed to retrieve secret")
				g.Expect(secret).NotTo(BeEmpty())
				issuerUrl, err := getSecret("keycloak-headlamp", "OIDC_ISSUER_URL")
				g.Expect(err).NotTo(HaveOccurred(), "Failed to retrieve issuer-url")
				g.Expect(issuerUrl).To(Equal("http://keycloak.keycloak.svc.cluster.local/realms/master"))

			}
			Eventually(verifyClientSecrets, 2*time.Minute).Should(Succeed())
			By("Keycloak client configmap handled")
			verifyClientConfigMap := func(g Gomega) {
				issuerUrl, err := getConfigMap("keycloak-config", "ISSUER_URL")
				g.Expect(err).NotTo(HaveOccurred(), "Failed to retrieve issuer-url")
				g.Expect(issuerUrl).To(Equal("http://keycloak.keycloak.svc.cluster.local/realms/master"))
			}
			Eventually(verifyClientConfigMap, 2*time.Minute).Should(Succeed())
			By("Deployment annotations handled")
			verifyDeploymentAnnotations := func(g Gomega) {
				var err error
				secretChecksum, err = getDeploymentAnnotation("app=nginx", "keycloak.osc.edu/secret-checksum")
				g.Expect(err).NotTo(HaveOccurred(), "Failed to get deployment annotation for the secret")
				g.Expect(secretChecksum).NotTo(BeEmpty())
				configmapChecksum, err = getDeploymentAnnotation("app=nginx", "keycloak.osc.edu/configmap-checksum")
				g.Expect(err).NotTo(HaveOccurred(), "Failed to get deployment annotation for the configmap")
				g.Expect(configmapChecksum).NotTo(BeEmpty())
			}
			Eventually(verifyDeploymentAnnotations, 2*time.Minute).Should(Succeed())
			By("Client updated in Keycloak")
			verifyClientUpdates := func(g Gomega) {
				cmd := exec.Command("kubectl", "patch", "keycloakclient", "keycloakclient-sample",
					"--type", "merge", "-p", "{\"spec\":{\"description\":\"sample\"}}")
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Patch keycloak client failed: %s", output))
				secretCmd := exec.Command("kubectl", "patch", "secret", "keycloak-sample",
					"-p", "{\"stringData\":{\"client-secret\":\"new-secret\"}}")
				secretOutput, secretErr := utils.Run(secretCmd)
				g.Expect(secretErr).NotTo(HaveOccurred(), fmt.Sprintf("Patch secret failed: %s", secretOutput))
				waitCmd := exec.Command("kubectl", "wait", "--for=condition=Available",
					"keycloakclient", "keycloakclient-sample", "--timeout=20s")
				waitOut, waitErr := utils.Run(waitCmd)
				g.Expect(waitOut).To(ContainSubstring("condition met"))
				g.Expect(waitErr).NotTo(HaveOccurred())
				client := getKeycloakClient("kubernetes-keycloakclient-sample", "master")
				g.Expect(client).To(Not(BeNil()), "expected client not found")
				g.Expect(*client.ClientID).To(Equal("kubernetes-keycloakclient-sample"))
				g.Expect(*client.Secret).To(Equal("new-secret"))
				g.Expect(*client.Description).To(Equal("sample"))
				g.Expect(*client.RedirectURIs).To(ConsistOf("https://example.com/*", "https://example.test.com/*"))
				g.Expect(*client.DefaultClientScopes).To(ConsistOf("web-origins", "profile", "email"))
			}
			Eventually(verifyClientUpdates, 2*time.Minute).Should(Succeed())
			By("Client updated in Keycloak with ConfigMap")
			verifyClientUpdatesConfigMap := func(g Gomega) {
				cmd := exec.Command("kubectl", "patch", "keycloakclient", "keycloakclient-test",
					"--type", "merge", "-p", "{\"spec\":{\"clientID\":\"kubernetes-foo\"}}")
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Patch keycloak client failed: %s", output))
				waitCmd := exec.Command("kubectl", "wait", "--for=condition=Available",
					"keycloakclient", "keycloakclient-test", "--timeout=20s")
				waitOut, waitErr := utils.Run(waitCmd)
				g.Expect(waitOut).To(ContainSubstring("condition met"))
				g.Expect(waitErr).NotTo(HaveOccurred())
				client := getKeycloakClient("kubernetes-foo", "master")
				g.Expect(client).To(Not(BeNil()), "expected client not found")
				g.Expect(*client.ClientID).To(Equal("kubernetes-foo"))
				clientID, err := getSecret("keycloak-test", "CLIENT_ID")
				g.Expect(err).NotTo(HaveOccurred(), "Failed to retrieve secret")
				g.Expect(clientID).To(Equal("kubernetes-foo"))
			}
			Eventually(verifyClientUpdatesConfigMap, 2*time.Minute).Should(Succeed())
			By("Deployment annotations updated")
			verifyDeploymentUpdatedAnnotations := func(g Gomega) {
				var err error
				secretChecksumUpdated, err = getDeploymentAnnotation("app=nginx", "keycloak.osc.edu/secret-checksum")
				g.Expect(err).NotTo(HaveOccurred(), "Failed to get deployment annotation for the secret")
				g.Expect(secretChecksumUpdated).NotTo(BeEmpty())
				g.Expect(secretChecksumUpdated).NotTo(Equal(secretChecksum))
				configmapChecksumUpdated, err = getDeploymentAnnotation("app=nginx", "keycloak.osc.edu/configmap-checksum")
				g.Expect(err).NotTo(HaveOccurred(), "Failed to get deployment annotation for the configmap")
				g.Expect(configmapChecksumUpdated).NotTo(BeEmpty())
				g.Expect(configmapChecksumUpdated).To(Equal(configmapChecksum))
			}
			Eventually(verifyDeploymentUpdatedAnnotations, 2*time.Minute).Should(Succeed())

			By("Testing webhook validation with invalid client creation")
			// Test creating an invalid KeycloakClient resource (missing ClientID)
			invalidClientTest := func(g Gomega) {
				// Create an invalid KeycloakClient resource (missing ClientID)
				invalidClientYAML := `
apiVersion: keycloak.osc.edu/v1alpha2
kind: KeycloakClient
metadata:
  name: invalid-client-test
  namespace: default
spec:
  realm: master
  clientID: invalid-client
`
				// Write the invalid client to a temporary file
				tmpFile := "/tmp/invalid-client.yaml"
				err := os.WriteFile(tmpFile, []byte(invalidClientYAML), 0644)
				g.Expect(err).NotTo(HaveOccurred())
				defer os.Remove(tmpFile)

				// Try to create the invalid client - this should fail
				cmd := exec.Command("kubectl", "create", "-f", tmpFile)
				output, err := utils.Run(cmd)
				g.Expect(err).To(HaveOccurred())
				g.Expect(output).To(ContainSubstring("clientID must begin with"))
			}
			Eventually(invalidClientTest, 2*time.Minute).Should(Succeed())

			By("Testing webhook validation with invalid patch")
			// Test patching with invalid data (empty ClientID)
			invalidPatchTest := func(g Gomega) {
				// Patch the existing client with invalid data (empty ClientID)
				patchCmd := exec.Command("kubectl", "patch", "keycloakclient", "keycloakclient-sample",
					"--type", "merge", "-p", "{\"spec\":{\"clientID\":\"keycloakclient-sample\"}}")
				patchOutput, patchErr := utils.Run(patchCmd)
				g.Expect(patchErr).To(HaveOccurred())
				g.Expect(patchOutput).To(ContainSubstring("clientID must begin with"))
			}
			Eventually(invalidPatchTest, 2*time.Minute).Should(Succeed())

			By("Delete clients")
			deleteClients := func(g Gomega) {
				cmd := exec.Command("kubectl", "delete",
					"-f", keycloakClientManifest,
					"-f", keycloakClientManifestWithSecret,
					"-f", keycloakClientManifestPublic,
					"-f", keycloakClientManifestHeadlamp)
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(Or(ContainSubstring("deleted"), ContainSubstring("not found")))
			}
			Eventually(deleteClients, 2*time.Minute).Should(Succeed())

			By("Client deleted from Keycloak")
			verifyKeycloakClientDelete := func(g Gomega) {
				client := getKeycloakClient("kubernetes-keycloakclient-sample", "master")
				g.Expect(client).To(BeNil(), "keycloak client still present")
				client = getKeycloakClient("kubernetes-foo", "master")
				g.Expect(client).To(BeNil(), "keycloak client still present")
				cmd := exec.Command("kubectl", "get", "secret", "keycloak-test")
				output, err := utils.Run(cmd)
				g.Expect(err).To(HaveOccurred())
				g.Expect(output).To(ContainSubstring("not found"))
				cmd = exec.Command("kubectl", "get", "configmap", "keycloak-config")
				output, err = utils.Run(cmd)
				g.Expect(err).To(HaveOccurred())
				g.Expect(output).To(ContainSubstring("not found"))
			}
			Eventually(verifyKeycloakClientDelete, 2*time.Minute).Should(Succeed())
		})

		It("should upgrade from v1alpha1 to v1alpha2 preserving managed resources", func() {
			By("Apply v1alpha1 KeycloakClient resource")
			verifyV1Alpha1Created := func(g Gomega) {
				cmd := exec.Command("kubectl", "apply", "-f", keycloakClientManifestUpgradeFrom)
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(ContainSubstring("created"))
				waitCmd := exec.Command("kubectl", "wait", "--for=condition=Available",
					"keycloakclient", "keycloakclient-test-upgrade", "--timeout=20s")
				waitOut, waitErr := utils.Run(waitCmd)
				g.Expect(waitOut).To(ContainSubstring("condition met"))
				g.Expect(waitErr).NotTo(HaveOccurred())
			}
			Eventually(verifyV1Alpha1Created, 2*time.Minute).Should(Succeed())

			By("Capturing v1alpha1 managed secret and configmap values")
			var v1Alpha1SecretValues map[string]string
			var v1Alpha1ConfigMapValues map[string]string
			captureV1Alpha1Values := func(g Gomega) {
				// Get all secret keys
				cmd := exec.Command("kubectl", "get", "secret", "keycloak-upgrade-secret",
					"-o", "jsonpath={.data}")
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())

				// Parse the secret data
				var secretData map[string]string
				err = json.Unmarshal([]byte(output), &secretData)
				g.Expect(err).NotTo(HaveOccurred())
				v1Alpha1SecretValues = secretData

				// Get all configmap keys
				cmd = exec.Command("kubectl", "get", "configmap", "keycloak-upgrade-config",
					"-o", "jsonpath={.data}")
				output, err = utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())

				// Parse the configmap data
				var configMapData map[string]string
				err = json.Unmarshal([]byte(output), &configMapData)
				g.Expect(err).NotTo(HaveOccurred())
				v1Alpha1ConfigMapValues = configMapData
			}
			Eventually(captureV1Alpha1Values).Should(Succeed())

			By("Upgrading to v1alpha2 by applying new manifest")
			verifyUpgrade := func(g Gomega) {
				// Apply v1alpha2 resource - this should update the CRD version
				cmd := exec.Command("kubectl", "apply", "-f", keycloakClientManifestUpgradeTo)
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(Or(ContainSubstring("created"), ContainSubstring("configured")))

				// Wait for the resource to be available after upgrade
				waitCmd := exec.Command("kubectl", "wait", "--for=condition=Available",
					"keycloakclient", "keycloakclient-test-upgrade", "--timeout=20s")
				waitOut, waitErr := utils.Run(waitCmd)
				g.Expect(waitOut).To(ContainSubstring("condition met"))
				g.Expect(waitErr).NotTo(HaveOccurred())
			}
			Eventually(verifyUpgrade, 2*time.Minute).Should(Succeed())

			By("Verifying v1alpha2 resource exists after upgrade")
			verifyV1Alpha2Exists := func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "keycloakclient", "keycloakclient-test-upgrade",
					"-o", "jsonpath={.apiVersion}")
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(ContainSubstring("v1alpha2"))
			}
			Eventually(verifyV1Alpha2Exists, 2*time.Minute).Should(Succeed())

			By("Verifying secret values are unchanged after upgrade")
			verifySecretValuesUnchanged := func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "secret", "keycloak-upgrade-secret",
					"-o", "jsonpath={.data}")
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())

				var secretData map[string]string
				err = json.Unmarshal([]byte(output), &secretData)
				g.Expect(err).NotTo(HaveOccurred())

				g.Expect(secretData).To(Equal(v1Alpha1SecretValues),
					"Secret values should remain the same after upgrade")
			}
			Eventually(verifySecretValuesUnchanged, 2*time.Minute).Should(Succeed())

			By("Verifying configmap values are unchanged after upgrade")
			verifyConfigMapValuesUnchanged := func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "configmap", "keycloak-upgrade-config",
					"-o", "jsonpath={.data}")
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())

				var configMapData map[string]string
				err = json.Unmarshal([]byte(output), &configMapData)
				g.Expect(err).NotTo(HaveOccurred())

				g.Expect(configMapData).To(Equal(v1Alpha1ConfigMapValues),
					"ConfigMap values should remain the same after upgrade")
			}
			Eventually(verifyConfigMapValuesUnchanged, 2*time.Minute).Should(Succeed())

			By("Verifying Keycloak client still exists with correct configuration")
			verifyKeycloakClientAfterUpgrade := func(g Gomega) {
				client := getKeycloakClient("kubernetes-default-keycloakclient-test-upgrade", "master")
				g.Expect(client).To(Not(BeNil()), "keycloak client should still exist after upgrade")
				g.Expect(*client.ClientID).To(Equal("kubernetes-default-keycloakclient-test-upgrade"))
				g.Expect(*client.RedirectURIs).To(ConsistOf("https://example.com/*", "https://example.test.com/*"))
				g.Expect(*client.DefaultClientScopes).To(ConsistOf("web-origins", "profile", "email"))
			}
			Eventually(verifyKeycloakClientAfterUpgrade, 2*time.Minute).Should(Succeed())

			By("Cleaning up upgraded resource")
			cleanupUpgrade := func(g Gomega) {
				cmd := exec.Command("kubectl", "delete", "-f", keycloakClientManifestUpgradeTo)
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(Or(ContainSubstring("deleted"), ContainSubstring("not found")))
			}
			Eventually(cleanupUpgrade, 2*time.Minute).Should(Succeed())

			By("Verifying resources are deleted after cleanup")
			verifyDeleted := func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "keycloakclient", "keycloakclient-test-upgrade")
				output, err := utils.Run(cmd)
				g.Expect(err).To(HaveOccurred())
				g.Expect(output).To(ContainSubstring("not found"))

				cmd = exec.Command("kubectl", "get", "secret", "keycloak-upgrade-secret")
				output, err = utils.Run(cmd)
				g.Expect(err).To(HaveOccurred())
				g.Expect(output).To(ContainSubstring("not found"))

				cmd = exec.Command("kubectl", "get", "configmap", "keycloak-upgrade-config")
				output, err = utils.Run(cmd)
				g.Expect(err).To(HaveOccurred())
				g.Expect(output).To(ContainSubstring("not found"))
			}
			Eventually(verifyDeleted, 2*time.Minute).Should(Succeed())
		})
	})
}
