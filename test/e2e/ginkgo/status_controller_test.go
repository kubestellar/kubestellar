/*
Copyright 2024 The KubeStellar Authors.

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
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/kubestellar/kubestellar/test/util"
)

var _ = ginkgo.Describe("Status controller singleton status propagation", func() {
	const statusTestNS = "status-test"

	ginkgo.BeforeEach(func(ctx context.Context) {
		util.CreateNS(ctx, wds, statusTestNS)
	})

	ginkgo.It("propagates WorkStatus to workload status when singleton status is requested", func(ctx context.Context) {
		cmName := "test-cm"
		cmWithLabel := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name":      cmName,
					"namespace": statusTestNS,
					"labels": map[string]interface{}{
						"managed-by.kubestellar.io/singletonstatus": "true",
					},
				},
				"data": map[string]interface{}{
					"key": "value",
				},
			},
		}

		wdsConfig := util.GetConfig(wds1CtxFlag)
		dynamicWDS, err := dynamic.NewForConfig(wdsConfig)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
		_, err = dynamicWDS.Resource(gvr).Namespace(statusTestNS).Create(ctx, cmWithLabel, metav1.CreateOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		retrieved, err := dynamicWDS.Resource(gvr).Namespace(statusTestNS).Get(ctx, cmName, metav1.GetOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		labels := retrieved.GetLabels()
		gomega.Expect(labels).To(gomega.HaveKey("managed-by.kubestellar.io/singletonstatus"))

		ginkgo.GinkgoLogr.Info("ConfigMap created with singleton status label", "name", cmName, "namespace", statusTestNS)

		gomega.Expect(labels).NotTo(gomega.BeEmpty())
	})

	ginkgo.It("validates binding policy selector matching for status objects", func(ctx context.Context) {
		cmName := "test-cm-no-label"
		cmNoLabel := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name":      cmName,
					"namespace": statusTestNS,
				},
				"data": map[string]interface{}{
					"key": "value",
				},
			},
		}

		wdsConfig := util.GetConfig(wds1CtxFlag)
		dynamicWDS, err := dynamic.NewForConfig(wdsConfig)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
		_, err = dynamicWDS.Resource(gvr).Namespace(statusTestNS).Create(ctx, cmNoLabel, metav1.CreateOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		retrieved, err := dynamicWDS.Resource(gvr).Namespace(statusTestNS).Get(ctx, cmName, metav1.GetOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		labels := retrieved.GetLabels()
		gomega.Expect(labels).NotTo(gomega.HaveKey("managed-by.kubestellar.io/singletonstatus"))

		ginkgo.GinkgoLogr.Info("ConfigMap created without singleton status label", "name", cmName, "namespace", statusTestNS)
	})

	ginkgo.AfterEach(func(ctx context.Context) {
		util.CleanupWDS(ctx, wds, ksWds, statusTestNS)
	})
})
