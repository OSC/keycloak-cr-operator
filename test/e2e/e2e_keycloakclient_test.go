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
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/OSC/keycloak-cr-operator/test/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func KeycloakClientSpec() {
	Context("KeycloakClient", func() {
		It("should handle custom resources", func() {
			By("Apply custom KeycloakClient resource from samples")
			verifyKeycloakClientResource := func(g Gomega) {
				cmd := exec.Command("kubectl", "apply",
					"-f", keycloakClientManifest,
					"-f", keycloakClientManifestWithSecret)
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(Or(ContainSubstring("created")))
				waitCmd := exec.Command("kubectl", "wait", "--for=condition=Available",
					"keycloakclient", "keycloakclient-test", "--timeout=20s")
				waitOut, waitErr := utils.Run(waitCmd)
				g.Expect(waitOut).To(ContainSubstring("condition met"))
				g.Expect(waitErr).NotTo(HaveOccurred())
				waitCmd = exec.Command("kubectl", "wait", "--for=condition=Available",
					"keycloakclient", "keycloakclient-sample", "--timeout=20s")
				waitOut, waitErr = utils.Run(waitCmd)
				g.Expect(waitOut).To(ContainSubstring("condition met"))
				g.Expect(waitErr).NotTo(HaveOccurred())
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
				client := getKeycloakClient("kubernetes-keycloakclient-test", "master")
				g.Expect(client).To(Not(BeNil()), "expected client not found")
				g.Expect(*client.ID).To(Equal("kubernetes-keycloakclient-test"))
				g.Expect(*client.ClientID).To(Equal("kubernetes-keycloakclient-test"))
				g.Expect(*client.RedirectURIs).To(ConsistOf("https://example.com/*", "https://example.test.com/*"))
				g.Expect(*client.DefaultClientScopes).To(ConsistOf("web-origins", "profile", "email"))
				client = getKeycloakClient("kubernetes-keycloakclient-sample", "master")
				g.Expect(client).To(Not(BeNil()), "expected client not found")
				g.Expect(*client.ID).To(Equal("kubernetes-keycloakclient-sample"))
				g.Expect(*client.ClientID).To(Equal("kubernetes-keycloakclient-sample"))
				g.Expect(*client.Secret).To(Equal("sample-secret"))
				g.Expect(*client.RedirectURIs).To(ConsistOf("https://example.com/*", "https://example.test.com/*"))
				g.Expect(*client.DefaultClientScopes).To(ConsistOf("web-origins", "profile", "email"))
			}
			Eventually(verifyClientExists, 2*time.Minute).Should(Succeed())
			By("Keycloak client secrets handled")
			verifyClientSecrets := func(g Gomega) {
				client := getKeycloakClient("kubernetes-keycloakclient-test", "master")
				secret, err := getSecret("keycloak-test", "client-secret")
				g.Expect(err).NotTo(HaveOccurred(), "Failed to retrieve secret")
				g.Expect(*client.Secret).To(Equal(secret))
				secret, err = getSecret("keycloak-test", "cookie-secret")
				g.Expect(err).NotTo(HaveOccurred(), "Failed to retrieve cookie-secret")
				g.Expect(secret).NotTo(BeEmpty())
			}
			Eventually(verifyClientSecrets, 2*time.Minute).Should(Succeed())
			By("Keycloak client configmap handled")
			verifyClientConfigMap := func(g Gomega) {
				client := getKeycloakClient("kubernetes-keycloakclient-test", "master")
				clientID, err := getConfigMap("keycloak-config", "client-id")
				g.Expect(err).NotTo(HaveOccurred(), "Failed to retrieve configmap")
				g.Expect(*client.ClientID).To(Equal(clientID))
				issuerUrl, err := getConfigMap("keycloak-config", "issuer-url")
				g.Expect(err).NotTo(HaveOccurred(), "Failed to retrieve cookie-secret")
				g.Expect(issuerUrl).To(Equal("http://keycloak.keycloak.svc.cluster.local/realms/master"))
			}
			Eventually(verifyClientConfigMap, 2*time.Minute).Should(Succeed())
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
				g.Expect(*client.ID).To(Equal("kubernetes-keycloakclient-sample"))
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
					"--type", "merge", "-p", "{\"spec\":{\"id\":\"kubernetes-foo\",\"clientID\":\"kubernetes-foo\"}}")
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Patch keycloak client failed: %s", output))
				waitCmd := exec.Command("kubectl", "wait", "--for=condition=Available",
					"keycloakclient", "keycloakclient-test", "--timeout=20s")
				waitOut, waitErr := utils.Run(waitCmd)
				g.Expect(waitOut).To(ContainSubstring("condition met"))
				g.Expect(waitErr).NotTo(HaveOccurred())
				client := getKeycloakClient("kubernetes-foo", "master")
				g.Expect(client).To(Not(BeNil()), "expected client not found")
				g.Expect(*client.ID).To(Equal("kubernetes-foo"))
				g.Expect(*client.ClientID).To(Equal("kubernetes-foo"))
				clientID, err := getConfigMap("keycloak-config", "client-id")
				g.Expect(err).NotTo(HaveOccurred(), "Failed to retrieve configmap")
				g.Expect(clientID).To(Equal("kubernetes-foo"))
			}
			Eventually(verifyClientUpdatesConfigMap, 2*time.Minute).Should(Succeed())

			By("Testing webhook validation with invalid client creation")
			// Test creating an invalid KeycloakClient resource (missing ClientID)
			invalidClientTest := func(g Gomega) {
				// Create an invalid KeycloakClient resource (missing ClientID)
				invalidClientYAML := `
apiVersion: keycloak.osc.edu/v1alpha1
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
					"-f", keycloakClientManifestWithSecret)
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
	})
}
