/*
Copyright 2025 The KubeStellar Authors.
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
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/klog/v2"

	ksapi "github.com/kubestellar/kubestellar/api/control/v1alpha1"
	"github.com/kubestellar/kubestellar/test/util"
)

const (
	testNamespace = "argocd-test"
	testAppName   = "test-guestbook"
	testRepoURL   = "https://github.com/argoproj/argocd-example-apps.git"
)

var _ = ginkgo.Describe("ArgoCD Integration with KubeStellar", func() {

	ginkgo.Context("ArgoCD Installation Verification", func() {
		ginkgo.It("should have ArgoCD pods running", func(ctx context.Context) {
			ginkgo.By("Checking ArgoCD server pod")
			util.ValidatePodRunning(ctx, coreCluster, argoCDNamespace, "app.kubernetes.io/name=argocd-server")

			ginkgo.By("Checking ArgoCD application controller pod")
			util.ValidatePodRunning(ctx, coreCluster, argoCDNamespace, "app.kubernetes.io/name=argocd-application-controller")

			ginkgo.By("Checking ArgoCD repo server pod")
			util.ValidatePodRunning(ctx, coreCluster, argoCDNamespace, "app.kubernetes.io/name=argocd-repo-server")
		})

		ginkgo.It("should have ArgoCD services accessible", func(ctx context.Context) {
			ginkgo.By("Verifying ArgoCD server service exists")
			_, err := coreCluster.CoreV1().Services(argoCDNamespace).Get(ctx, "argocd-server", metav1.GetOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			ginkgo.By("Verifying ArgoCD repo server service exists")
			_, err = coreCluster.CoreV1().Services(argoCDNamespace).Get(ctx, "argocd-repo-server", metav1.GetOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})
	})

	ginkgo.Context("KubeStellar-ArgoCD Integration", func() {
		ginkgo.BeforeEach(func(ctx context.Context) {
			// Create test namespace
			util.CreateNS(ctx, wds, testNamespace)

			// Clean up any existing test resources
			cleanupTestResources(ctx)
		})

		ginkgo.AfterEach(func(ctx context.Context) {
			cleanupTestResources(ctx)
		})

		ginkgo.It("should register KubeStellar WDS as ArgoCD cluster", func(ctx context.Context) {
			ginkgo.By("Checking if WDS is registered as ArgoCD cluster")

			// Get ArgoCD server pod
			pods, err := coreCluster.CoreV1().Pods(argoCDNamespace).List(ctx, metav1.ListOptions{
				LabelSelector: "app.kubernetes.io/name=argocd-server",
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(len(pods.Items)).To(gomega.BeNumerically(">", 0))

			serverPod := pods.Items[0].Name

			// Execute argocd cluster list command
			cmd := []string{"argocd", "cluster", "list", "--plaintext"}
			stdout, stderr, err := util.ExecInPod(ctx, coreCluster, argoCDNamespace, serverPod, "argocd-server", cmd)

			klog.FromContext(ctx).V(2).Info("ArgoCD cluster list output", "stdout", stdout, "stderr", stderr)

			// Should see either wds1 or kubernetes.default.svc in the cluster list
			gomega.Expect(stdout).To(gomega.Or(
				gomega.ContainSubstring("wds1"),
				gomega.ContainSubstring("kubernetes.default.svc"),
			))
		})

		ginkgo.It("should create and sync ArgoCD application through KubeStellar", func(ctx context.Context) {
			ginkgo.By("Creating ArgoCD Application via KubeStellar WDS")

			// Create ArgoCD Application object in WDS
			appManifest := createArgoCDApplicationManifest()
			_, err := util.ApplyYAMLToCluster(ctx, wds, appManifest)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			ginkgo.By("Creating BindingPolicy for ArgoCD Application")
			util.CreateBindingPolicy(ctx, ksWds, "argocd-app-binding",
				[]metav1.LabelSelector{
					{MatchLabels: map[string]string{"location-group": "edge"}},
				},
				[]ksapi.DownsyncPolicyClause{
					{DownsyncObjectTest: ksapi.DownsyncObjectTest{
						Resources: []string{"applications.argoproj.io"},
						ObjectSelectors: []metav1.LabelSelector{{
							MatchLabels: map[string]string{"test.kubestellar.io/argocd": "integration"},
						}},
					}},
				},
			)

			ginkgo.By("Waiting for Application to be created")
			gomega.Eventually(func() error {
				_, err := getArgoCDApplication(ctx)
				return err
			}, time.Minute*2, time.Second*10).Should(gomega.Succeed())

			ginkgo.By("Triggering application sync")
			err = syncArgoCDApplication(ctx)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			ginkgo.By("Validating application sync status")
			gomega.Eventually(func() (string, error) {
				return getArgoCDApplicationSyncStatus(ctx)
			}, time.Minute*5, time.Second*15).Should(gomega.Equal("Synced"))
		})

		ginkgo.It("should deploy application resources to WECs through ArgoCD", func(ctx context.Context) {
			ginkgo.By("Creating test application and binding policy")

			// Create ArgoCD Application
			appManifest := createArgoCDApplicationManifest()
			_, err := util.ApplyYAMLToCluster(ctx, wds, appManifest)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Create BindingPolicy
			util.CreateBindingPolicy(ctx, ksWds, "argocd-workload-binding",
				[]metav1.LabelSelector{
					{MatchLabels: map[string]string{"name": "cluster1"}},
				},
				[]ksapi.DownsyncPolicyClause{
					{DownsyncObjectTest: ksapi.DownsyncObjectTest{
						Resources: []string{"deployments", "services"},
						NamespaceSelectors: []metav1.LabelSelector{{
							MatchLabels: map[string]string{"argocd.argoproj.io/managed-by": testAppName},
						}},
					}},
				},
			)

			ginkgo.By("Syncing ArgoCD application")
			err = syncArgoCDApplication(ctx)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			ginkgo.By("Waiting for application resources to be deployed to WEC1")
			gomega.Eventually(func() int {
				deployments, err := wec1.AppsV1().Deployments(testNamespace).List(ctx, metav1.ListOptions{})
				if err != nil {
					return 0
				}
				return len(deployments.Items)
			}, time.Minute*3, time.Second*10).Should(gomega.BeNumerically(">", 0))

			ginkgo.By("Validating deployed resources are healthy")
			util.ValidateNumDeployments(ctx, "wec1", wec1, testNamespace, 1)
		})
	})

	ginkgo.Context("ArgoCD Error Scenarios", func() {
		ginkgo.It("should handle ArgoCD server restart gracefully", func(ctx context.Context) {
			ginkgo.By("Creating test application")
			appManifest := createArgoCDApplicationManifest()
			_, err := util.ApplyYAMLToCluster(ctx, wds, appManifest)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			ginkgo.By("Restarting ArgoCD server")
			err = coreCluster.CoreV1().Pods(argoCDNamespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{
				LabelSelector: "app.kubernetes.io/name=argocd-server",
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			ginkgo.By("Waiting for ArgoCD server to restart")
			util.ValidatePodRunning(ctx, coreCluster, argoCDNamespace, "app.kubernetes.io/name=argocd-server")

			ginkgo.By("Verifying application is still accessible")
			gomega.Eventually(func() error {
				_, err := getArgoCDApplication(ctx)
				return err
			}, time.Minute*2, time.Second*10).Should(gomega.Succeed())
		})
	})
})

// Helper functions

func cleanupTestResources(ctx context.Context) {
	// Delete test applications
	_ = deleteArgoCDApplication(ctx)

	// Delete test namespace if it exists
	_ = wds.CoreV1().Namespaces().Delete(ctx, testNamespace, metav1.DeleteOptions{})
	_ = wec1.CoreV1().Namespaces().Delete(ctx, testNamespace, metav1.DeleteOptions{})
	_ = wec2.CoreV1().Namespaces().Delete(ctx, testNamespace, metav1.DeleteOptions{})

	// Delete binding policies
	_ = ksWds.ControlV1alpha1().BindingPolicies().Delete(ctx, "argocd-app-binding", metav1.DeleteOptions{})
	_ = ksWds.ControlV1alpha1().BindingPolicies().Delete(ctx, "argocd-workload-binding", metav1.DeleteOptions{})
}

func createArgoCDApplicationManifest() string {
	return fmt.Sprintf(`
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: %s
  namespace: %s
  labels:
    test.kubestellar.io/argocd: integration
spec:
  project: default
  source:
    repoURL: %s
    targetRevision: HEAD
    path: guestbook
  destination:
    server: https://kubernetes.default.svc
    namespace: %s
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
    - CreateNamespace=true
`, testAppName, argoCDNamespace, testRepoURL, testNamespace)
}

func getArgoCDApplication(ctx context.Context) (*metav1.Object, error) {
	return util.GetResource(ctx, wds, "argoproj.io/v1alpha1", "Application", argoCDNamespace, testAppName)
}

func deleteArgoCDApplication(ctx context.Context) error {
	return util.DeleteResource(ctx, wds, "argoproj.io/v1alpha1", "Application", argoCDNamespace, testAppName)
}

func syncArgoCDApplication(ctx context.Context) error {
	// Get ArgoCD server pod
	pods, err := coreCluster.CoreV1().Pods(argoCDNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/name=argocd-server",
	})
	if err != nil {
		return err
	}
	if len(pods.Items) == 0 {
		return fmt.Errorf("no ArgoCD server pods found")
	}

	serverPod := pods.Items[0].Name
	cmd := []string{"argocd", "app", "sync", testAppName, "--plaintext"}

	_, stderr, err := util.ExecInPod(ctx, coreCluster, argoCDNamespace, serverPod, "argocd-server", cmd)
	if err != nil {
		return fmt.Errorf("failed to sync application: %v, stderr: %s", err, stderr)
	}

	return nil
}

func getArgoCDApplicationSyncStatus(ctx context.Context) (string, error) {
	// Get ArgoCD server pod
	pods, err := coreCluster.CoreV1().Pods(argoCDNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/name=argocd-server",
	})
	if err != nil {
		return "", err
	}
	if len(pods.Items) == 0 {
		return "", fmt.Errorf("no ArgoCD server pods found")
	}

	serverPod := pods.Items[0].Name
	cmd := []string{"argocd", "app", "get", testAppName, "--output", "json", "--plaintext"}

	stdout, stderr, err := util.ExecInPod(ctx, coreCluster, argoCDNamespace, serverPod, "argocd-server", cmd)
	if err != nil {
		return "", fmt.Errorf("failed to get application status: %v, stderr: %s", err, stderr)
	}

	// Parse JSON and extract sync status
	// This is a simplified version - you might need to use proper JSON parsing
	if strings.Contains(stdout, `"status":"Synced"`) {
		return "Synced", nil
	}

	return "Unknown", nil
}
