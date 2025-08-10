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
	"flag"
	"strings"
	"testing"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	ocmWorkClient "open-cluster-management.io/api/client/work/clientset/versioned"

	"k8s.io/client-go/kubernetes"

	ksClient "github.com/kubestellar/kubestellar/pkg/generated/clientset/versioned"
	"github.com/kubestellar/kubestellar/test/util"
)

func TestArgoCD(t *testing.T) {
	gomega.SetDefaultEventuallyTimeout(time.Minute * 3)
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubeStellar ArgoCD Integration Testing")
}

var (
	coreCluster        *kubernetes.Clientset
	wds                *kubernetes.Clientset
	ksWds              *ksClient.Clientset
	wec1               *kubernetes.Clientset
	wec2               *kubernetes.Clientset
	its                *ocmWorkClient.Clientset
	releasedFlag       bool
	ksSetupFlags       string
	skipSetupFlag      bool
	hostClusterCtxFlag string
	wds1CtxFlag        string
	its1CtxFlag        string
	wec1CtxFlag        string
	wec2CtxFlag        string
	argoCDNamespace    = "argocd"
)

func init() {
	flag.BoolVar(&releasedFlag, "released", false, "released controls whether we use a release image")
	flag.BoolVar(&skipSetupFlag, "skip-setup", false, "skip kubestellar cleanup and setup")
	flag.StringVar(&ksSetupFlags, "kubestellar-setup-flags", ksSetupFlags, "additional command line flags for setup-kubestellar.sh, space separated")
	flag.StringVar(&hostClusterCtxFlag, "host-cluster-context", "kind-kubeflex", "context for kubeflex hosting cluster")
	flag.StringVar(&wds1CtxFlag, "wds1-context", "wds1", "context for KS wds1 space")
	flag.StringVar(&its1CtxFlag, "its1-context", "its1", "context for KS its1 space")
	flag.StringVar(&wec1CtxFlag, "wec1-context", "cluster1", "context for wec1 cluster")
	flag.StringVar(&wec2CtxFlag, "wec2-context", "cluster2", "context for wec2 cluster")
}

var _ = ginkgo.BeforeSuite(func(ctx context.Context) {
	if !skipSetupFlag {
		separatedFlags := strings.Split(ksSetupFlags, " ")
		if len(ksSetupFlags) == 0 {
			separatedFlags = separatedFlags[:0]
		}
		// Add ArgoCD installation flag to setup
		separatedFlags = append(separatedFlags, "--argocd")

		util.Cleanup(ctx)
		util.SetupKubestellar(ctx, releasedFlag, separatedFlags...)
	}

	configCore := util.GetConfig(hostClusterCtxFlag)
	configWds := util.GetConfig(wds1CtxFlag)
	configITS := util.GetConfig(its1CtxFlag)
	configWec1 := util.GetConfig(wec1CtxFlag)
	configWec2 := util.GetConfig(wec2CtxFlag)

	coreCluster = util.CreateKubeClient(configCore)
	wds = util.CreateKubeClient(configWds)
	ksWds = util.CreateKSClient(configWds)
	its = util.CreateOcmWorkClient(configITS)
	wec1 = util.CreateKubeClient(configWec1)
	wec2 = util.CreateKubeClient(configWec2)

	ginkgo.GinkgoLogr.Info("ArgoCD Suite setup done")
	time.Sleep(10 * time.Second)
})