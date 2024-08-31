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

package perf

import (
	"context"
	"fmt"
	"time"

	"github.com/onsi/ginkgo/v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ksapi "github.com/kubestellar/kubestellar/api/control/v1alpha1"
	"github.com/kubestellar/kubestellar/test/util"
)

const (
	ns = "nginx"
)

var _ = ginkgo.Describe("performance tests", func() {
	ginkgo.BeforeEach(func(ctx context.Context) {
		util.CleanupWDS(ctx, wdsClient, ksWdsClient, ns)
	})

	ginkgo.It("propagates a new deployment from WDS to WEC", func(ctx context.Context) {
		const numDeployments = 10
		sumBinding, sumManifestwork, sumDeploymentWDS, sumDeploymentWEC, sumTotal := 0, 0, 0, 0, 0

		for i := 0; i < numDeployments; i++ {
			start := time.Now()
			testName := fmt.Sprint(i)
			util.CreateBindingPolicy(ctx, ksWdsClient, testName,
				[]metav1.LabelSelector{
					{MatchLabels: map[string]string{"name": "cluster1"}},
				},
				[]ksapi.DownsyncPolicyClause{
					{DownsyncObjectTest: ksapi.DownsyncObjectTest{
						ObjectSelectors: []metav1.LabelSelector{
							{MatchLabels: map[string]string{"app.kubernetes.io/name": "nginx"}},
						}}}})
			util.CreateDeployment(ctx, wdsClient, ns, testName,
				map[string]string{
					"app.kubernetes.io/name": "nginx",
				})
			util.ValidateNumDeployments(ctx, "wec1", wec1Client, ns, 1)

			timeDeploymentWDS := int(util.GetDeploymentTime(ctx, wdsClient, ns, testName).Sub(start).Seconds())
			timeBinding := int(util.GetBindingTime(ctx, ksWdsClient, "performance").Sub(start).Seconds())
			timeManifestwork := int(util.GetManifestworkTime(ctx, itsClient, "cluster1", "performance-wds1").Sub(start).Seconds())
			timeDeploymentWEC := int(util.GetDeploymentTime(ctx, wec1Client, ns, testName).Sub(start).Seconds())
			elapsed := int(time.Since(start).Seconds())

			sumDeploymentWDS += timeDeploymentWDS
			sumBinding += timeBinding
			sumManifestwork += timeManifestwork
			sumDeploymentWEC += timeDeploymentWEC
			sumTotal += elapsed

			if !justSummary {
				fmt.Fprintf(ginkgo.GinkgoWriter, "Run %d: wds deployment=%d, binding=%d, manifestwork=%d, wec deployment=%d, total=%d\n",
					i, timeDeploymentWDS, timeBinding, timeManifestwork, timeDeploymentWEC, elapsed)
			}
			util.DeleteDeployment(ctx, wdsClient, ns, testName)
			util.ValidateNumDeployments(ctx, "wec1", wec1Client, ns, 0)
			util.DeleteBindingPolicy(ctx, ksWdsClient, testName)
		}

		if !justSummary {
			fmt.Fprintf(ginkgo.GinkgoWriter, "----------------------------------------------------------------------------------\n")
		}
		fmt.Fprintf(ginkgo.GinkgoWriter, "Avg:   wds deployment=%d, binding=%d, manifestwork=%d, wec deployment=%d, total=%d\n",
			int(sumDeploymentWDS/numDeployments), int(sumBinding/numDeployments), int(sumManifestwork/numDeployments),
			int(sumDeploymentWEC/numDeployments), int(sumTotal/numDeployments))
	})
})
