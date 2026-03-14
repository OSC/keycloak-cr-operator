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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/OSC/keycloak-cr-operator/test/utils"
)

// namespace where the project is deployed in
const namespace = "keycloak-cr-operator-system"

// serviceAccountName created for the project
const serviceAccountName = "keycloak-cr-operator-controller-manager"

// metricsServiceName is the name of the metrics service of the project
const metricsServiceName = "keycloak-cr-operator-controller-manager-metrics-service"

// metricsRoleBindingName is the name of the RBAC that will be created to allow get the metrics data
const metricsRoleBindingName = "keycloak-cr-operator-metrics-binding"

const keycloakClientManifest = "config/samples/keycloak_v1alpha1_keycloakclient.yaml"
const keycloakClientManifestWithSecret = "config/samples/keycloak_v1alpha1_keycloakclient_with_secret.yaml"

var _ = Describe("Manager", Ordered, func() {
	var controllerPodName string

	// Before running the tests, set up the environment by creating the namespace,
	// enforce the restricted security policy to the namespace, installing CRDs,
	// and deploying the controller.
	BeforeAll(func() {
		By("creating manager namespace")
		cmd := exec.Command("kubectl", "create", "ns", namespace)
		_, err := utils.Run(cmd)
		Expect(err).NotTo(HaveOccurred(), "Failed to create namespace")

		By("labeling the namespace to enforce the restricted security policy")
		cmd = exec.Command("kubectl", "label", "--overwrite", "ns", namespace,
			"pod-security.kubernetes.io/enforce=restricted")
		_, err = utils.Run(cmd)
		Expect(err).NotTo(HaveOccurred(), "Failed to label namespace with restricted policy")

		By("installing CRDs")
		cmd = exec.Command("make", "install")
		_, err = utils.Run(cmd)
		Expect(err).NotTo(HaveOccurred(), "Failed to install CRDs")

		By("deploying the controller-manager")
		cmd = exec.Command("make", "deploy", fmt.Sprintf("IMG=%s", managerImage))
		_, err = utils.Run(cmd)
		Expect(err).NotTo(HaveOccurred(), "Failed to deploy the controller-manager")
	})

	// After all tests have been executed, clean up by undeploying the controller, uninstalling CRDs,
	// and deleting the namespace.
	AfterAll(func() {
		By("cleanup custom resources")
		cmd := exec.Command("kubectl", "delete", "keycloakclient", "--all", "--force")
		_, _ = utils.Run(cmd)

		By("cleaning up the curl pod for metrics")
		cmd = exec.Command("kubectl", "delete", "pod", "curl-metrics", "-n", namespace)
		_, _ = utils.Run(cmd)

		By("undeploying the controller-manager")
		cmd = exec.Command("make", "undeploy")
		_, _ = utils.Run(cmd)

		By("uninstalling CRDs")
		cmd = exec.Command("make", "uninstall")
		_, _ = utils.Run(cmd)

		By("removing manager namespace")
		cmd = exec.Command("kubectl", "delete", "ns", namespace)
		_, _ = utils.Run(cmd)

		By("removing metrics clusterrolebinding")
		cmd = exec.Command("kubectl", "delete", "clusterrolebinding", metricsRoleBindingName)
		_, _ = utils.Run(cmd)
	})

	// After each test, check for failures and collect logs, events,
	// and pod descriptions for debugging.
	AfterEach(func() {
		specReport := CurrentSpecReport()
		if specReport.Failed() {
			By("Fetching controller manager pod logs")
			cmd := exec.Command("kubectl", "logs", controllerPodName, "-n", namespace)
			controllerLogs, err := utils.Run(cmd)
			if err == nil {
				_, _ = fmt.Fprintf(GinkgoWriter, "Controller logs:\n %s", controllerLogs)
			} else {
				_, _ = fmt.Fprintf(GinkgoWriter, "Failed to get Controller logs: %s", err)
			}

			By("Fetching Kubernetes events")
			cmd = exec.Command("kubectl", "get", "events", "-n", namespace, "--sort-by=.lastTimestamp")
			eventsOutput, err := utils.Run(cmd)
			if err == nil {
				_, _ = fmt.Fprintf(GinkgoWriter, "Kubernetes events:\n%s", eventsOutput)
			} else {
				_, _ = fmt.Fprintf(GinkgoWriter, "Failed to get Kubernetes events: %s", err)
			}

			By("Fetching curl-metrics logs")
			cmd = exec.Command("kubectl", "logs", "curl-metrics", "-n", namespace)
			metricsOutput, err := utils.Run(cmd)
			if err == nil {
				_, _ = fmt.Fprintf(GinkgoWriter, "Metrics logs:\n %s", metricsOutput)
			} else {
				_, _ = fmt.Fprintf(GinkgoWriter, "Failed to get curl-metrics logs: %s", err)
			}

			By("Fetching controller manager pod description")
			cmd = exec.Command("kubectl", "describe", "pod", controllerPodName, "-n", namespace)
			podDescription, err := utils.Run(cmd)
			if err == nil {
				fmt.Println("Pod description:\n", podDescription)
			} else {
				fmt.Println("Failed to describe controller pod")
			}
		}
	})

	SetDefaultEventuallyTimeout(2 * time.Minute)
	SetDefaultEventuallyPollingInterval(time.Second)

	Context("Manager", func() {
		It("should run successfully", func() {
			By("validating that the controller-manager pod is running as expected")
			verifyControllerUp := func(g Gomega) {
				// Get the name of the controller-manager pod
				cmd := exec.Command("kubectl", "get",
					"pods", "-l", "control-plane=controller-manager",
					"-o", "go-template={{ range .items }}"+
						"{{ if not .metadata.deletionTimestamp }}"+
						"{{ .metadata.name }}"+
						"{{ \"\\n\" }}{{ end }}{{ end }}",
					"-n", namespace,
				)

				podOutput, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred(), "Failed to retrieve controller-manager pod information")
				podNames := utils.GetNonEmptyLines(podOutput)
				g.Expect(podNames).To(HaveLen(1), "expected 1 controller pod running")
				controllerPodName = podNames[0]
				g.Expect(controllerPodName).To(ContainSubstring("controller-manager"))

				// Validate the pod's status
				cmd = exec.Command("kubectl", "get",
					"pods", controllerPodName, "-o", "jsonpath={.status.phase}",
					"-n", namespace,
				)
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(Equal("Running"), "Incorrect controller-manager pod status")
			}
			Eventually(verifyControllerUp).Should(Succeed())
		})

		It("should ensure the metrics endpoint is serving metrics", func() {
			By("creating a ClusterRoleBinding for the service account to allow access to metrics")
			cmd := exec.Command("kubectl", "create", "clusterrolebinding", metricsRoleBindingName,
				"--clusterrole=keycloak-cr-operator-metrics-reader",
				fmt.Sprintf("--serviceaccount=%s:%s", namespace, serviceAccountName),
			)
			_, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred(), "Failed to create ClusterRoleBinding")

			By("validating that the metrics service is available")
			cmd = exec.Command("kubectl", "get", "service", metricsServiceName, "-n", namespace)
			_, err = utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred(), "Metrics service should exist")

			By("getting the service account token")
			token, err := serviceAccountToken()
			Expect(err).NotTo(HaveOccurred())
			Expect(token).NotTo(BeEmpty())

			By("ensuring the controller pod is ready")
			verifyControllerPodReady := func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "pod", controllerPodName, "-n", namespace,
					"-o", "jsonpath={.status.conditions[?(@.type=='Ready')].status}")
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(Equal("True"), "Controller pod not ready")
			}
			Eventually(verifyControllerPodReady, 3*time.Minute, time.Second).Should(Succeed())

			By("verifying that the controller manager is serving the metrics server")
			verifyMetricsServerStarted := func(g Gomega) {
				cmd := exec.Command("kubectl", "logs", controllerPodName, "-n", namespace)
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(ContainSubstring("Serving metrics server"),
					"Metrics server not yet started")
			}
			Eventually(verifyMetricsServerStarted, 3*time.Minute, time.Second).Should(Succeed())

			By("waiting for the webhook service endpoints to be ready")
			verifyWebhookEndpointsReady := func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "endpointslices.discovery.k8s.io", "-n", namespace,
					"-l", "kubernetes.io/service-name=keycloak-cr-operator-webhook-service",
					"-o", "jsonpath={range .items[*]}{range .endpoints[*]}{.addresses[*]}{end}{end}")
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred(), "Webhook endpoints should exist")
				g.Expect(output).ShouldNot(BeEmpty(), "Webhook endpoints not yet ready")
			}
			Eventually(verifyWebhookEndpointsReady, 3*time.Minute, time.Second).Should(Succeed())

			By("verifying the mutating webhook server is ready")
			verifyMutatingWebhookReady := func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "mutatingwebhookconfigurations.admissionregistration.k8s.io",
					"keycloak-cr-operator-mutating-webhook-configuration",
					"-o", "jsonpath={.webhooks[0].clientConfig.caBundle}")
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred(), "MutatingWebhookConfiguration should exist")
				g.Expect(output).ShouldNot(BeEmpty(), "Mutating webhook CA bundle not yet injected")
			}
			Eventually(verifyMutatingWebhookReady, 3*time.Minute, time.Second).Should(Succeed())

			By("verifying the validating webhook server is ready")
			verifyValidatingWebhookReady := func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "validatingwebhookconfigurations.admissionregistration.k8s.io",
					"keycloak-cr-operator-validating-webhook-configuration",
					"-o", "jsonpath={.webhooks[0].clientConfig.caBundle}")
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred(), "ValidatingWebhookConfiguration should exist")
				g.Expect(output).ShouldNot(BeEmpty(), "Validating webhook CA bundle not yet injected")
			}
			Eventually(verifyValidatingWebhookReady, 3*time.Minute, time.Second).Should(Succeed())

			By("waiting additional time for webhook server to stabilize")
			time.Sleep(5 * time.Second)

			// +kubebuilder:scaffold:e2e-metrics-webhooks-readiness

			By("creating the curl-metrics pod to access the metrics endpoint")
			cmd = exec.Command("kubectl", "run", "curl-metrics", "--restart=Never",
				"--namespace", namespace,
				"--image=curlimages/curl:latest",
				"--overrides",
				fmt.Sprintf(`{
					"spec": {
						"containers": [{
							"name": "curl",
							"image": "curlimages/curl:latest",
							"command": ["/bin/sh", "-c"],
							"args": [
								"sleep 600"
							],
							"env": [
								{
									"name": "TOKEN",
									"value": "%s"
								}
							],
							"securityContext": {
								"readOnlyRootFilesystem": true,
								"allowPrivilegeEscalation": false,
								"capabilities": {
									"drop": ["ALL"]
								},
								"runAsNonRoot": true,
								"runAsUser": 1000,
								"seccompProfile": {
									"type": "RuntimeDefault"
								}
							}
						}],
						"serviceAccountName": "%s"
					}
				}`, token, serviceAccountName))
			_, err = utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred(), "Failed to create curl-metrics pod")

			By("waiting for the curl-metrics pod to complete.")
			verifyCurlUp := func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "pods", "curl-metrics",
					"-o", "jsonpath={.status.phase}",
					"-n", namespace)
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(Or(Equal("Succeeded"), Equal("Running")), "curl pod in wrong status")
			}
			Eventually(verifyCurlUp, 5*time.Minute).Should(Succeed())

			By("getting the metrics by checking curl-metrics logs")
			verifyMetricsAvailable := func(g Gomega) {
				metricsOutput, err := getMetricsOutput()
				g.Expect(err).NotTo(HaveOccurred(), "Failed to retrieve logs from curl pod")
				g.Expect(metricsOutput).NotTo(BeEmpty())
				g.Expect(metricsOutput).To(ContainSubstring("< HTTP/1.1 200 OK"))
			}
			Eventually(verifyMetricsAvailable, 2*time.Minute).Should(Succeed())
		})

		It("should provisioned cert-manager", func() {
			By("validating that cert-manager has the certificate Secret")
			verifyCertManager := func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "secrets", "webhook-server-cert", "-n", namespace)
				_, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
			}
			Eventually(verifyCertManager).Should(Succeed())
		})

		It("should have CA injection for mutating webhooks", func() {
			By("checking CA injection for mutating webhooks")
			verifyCAInjection := func(g Gomega) {
				cmd := exec.Command("kubectl", "get",
					"mutatingwebhookconfigurations.admissionregistration.k8s.io",
					"keycloak-cr-operator-mutating-webhook-configuration",
					"-o", "go-template={{ range .webhooks }}{{ .clientConfig.caBundle }}{{ end }}")
				mwhOutput, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(len(mwhOutput)).To(BeNumerically(">", 10))
			}
			Eventually(verifyCAInjection).Should(Succeed())
		})

		It("should have CA injection for validating webhooks", func() {
			By("checking CA injection for validating webhooks")
			verifyCAInjection := func(g Gomega) {
				cmd := exec.Command("kubectl", "get",
					"validatingwebhookconfigurations.admissionregistration.k8s.io",
					"keycloak-cr-operator-validating-webhook-configuration",
					"-o", "go-template={{ range .webhooks }}{{ .clientConfig.caBundle }}{{ end }}")
				vwhOutput, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(len(vwhOutput)).To(BeNumerically(">", 10))
			}
			Eventually(verifyCAInjection).Should(Succeed())
		})

		// +kubebuilder:scaffold:e2e-webhooks-checks

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
				secret, err := getSecret("keycloak-test", "client-secret")
				g.Expect(err).NotTo(HaveOccurred(), "Failed to retrieve secret")
				g.Expect(*client.Secret).To(Equal(secret))
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

			By("Client deleted from Keycloak")
			verifyKeycloakClientDelete := func(g Gomega) {
				cmd := exec.Command("kubectl", "delete", "-f", keycloakClientManifestWithSecret)
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(Or(ContainSubstring("deleted")))
				client := getKeycloakClient("kubernetes-keycloakclient-sample", "master")
				g.Expect(client).To(BeNil(), "keycloak client still present")
			}
			Eventually(verifyKeycloakClientDelete, 2*time.Minute).Should(Succeed())
		})
	})
})

// serviceAccountToken returns a token for the specified service account in the given namespace.
// It uses the Kubernetes TokenRequest API to generate a token by directly sending a request
// and parsing the resulting token from the API response.
func serviceAccountToken() (string, error) {
	const tokenRequestRawString = `{
		"apiVersion": "authentication.k8s.io/v1",
		"kind": "TokenRequest"
	}`

	// Temporary file to store the token request
	secretName := fmt.Sprintf("%s-token-request", serviceAccountName)
	tokenRequestFile := filepath.Join("/tmp", secretName)
	err := os.WriteFile(tokenRequestFile, []byte(tokenRequestRawString), os.FileMode(0o644))
	if err != nil {
		return "", err
	}

	var out string
	verifyTokenCreation := func(g Gomega) {
		// Execute kubectl command to create the token
		cmd := exec.Command("kubectl", "create", "--raw", fmt.Sprintf(
			"/api/v1/namespaces/%s/serviceaccounts/%s/token",
			namespace,
			serviceAccountName,
		), "-f", tokenRequestFile)

		output, err := cmd.CombinedOutput()
		g.Expect(err).NotTo(HaveOccurred())

		// Parse the JSON output to extract the token
		var token tokenRequest
		err = json.Unmarshal(output, &token)
		g.Expect(err).NotTo(HaveOccurred())

		out = token.Status.Token
	}
	Eventually(verifyTokenCreation).Should(Succeed())

	return out, err
}

// getMetricsOutput retrieves and returns the logs from the curl pod used to access the metrics endpoint.
func getMetricsOutput() (string, error) {
	By("getting the curl-metrics logs")
	token, err := serviceAccountToken()
	Expect(err).NotTo(HaveOccurred())
	Expect(token).NotTo(BeEmpty())
	cmd := exec.Command("kubectl", "exec", "pod/curl-metrics", "-n", namespace, "--",
		"sh", "-c",
		fmt.Sprintf("curl -v -k -H \"Authorization: Bearer $TOKEN\" https://%s.%s.svc.cluster.local:8443/metrics", metricsServiceName, namespace),
	)
	return utils.Run(cmd)
}

func getSecret(name, key string) (string, error) {
	By("getting secret value")
	var secret string
	var err error
	getSecretValue := func(g Gomega) {
		cmd := exec.Command("kubectl", "get", "secret", name,
			fmt.Sprintf("--template={{ index .data \"%s\"}}", key))
		output, err := utils.Run(cmd)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(output).NotTo(BeEmpty())
		fmt.Printf("Returned secret: %s", output)
		decodedBytes, err := base64.StdEncoding.DecodeString(strings.TrimSpace(output))
		g.Expect(err).NotTo(HaveOccurred())
		secret = string(decodedBytes)
	}
	Eventually(getSecretValue).Should(Succeed())
	return secret, err
}

// tokenRequest is a simplified representation of the Kubernetes TokenRequest API response,
// containing only the token field that we need to extract.
type tokenRequest struct {
	Status struct {
		Token string `json:"token"`
	} `json:"status"`
}
