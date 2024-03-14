/*
Copyright 2023 The KubeStellar Authors.

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
	"testing"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	ocmWorkClient "open-cluster-management.io/api/client/work/clientset/versioned"

	"k8s.io/client-go/kubernetes"

	ksClient "github.com/kubestellar/kubestellar/pkg/generated/clientset/versioned"
	"github.com/kubestellar/kubestellar/test/util"
)

func TestGinkgo(t *testing.T) {
	gomega.SetDefaultEventuallyTimeout(time.Minute)
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Kubestellar testing")
}

var (
	ctx           context.Context
	coreCluster   *kubernetes.Clientset
	wds           *kubernetes.Clientset
	ksWds         *ksClient.Clientset
	wec1          *kubernetes.Clientset
	wec2          *kubernetes.Clientset
	its           *ocmWorkClient.Clientset
	releasedFlag  bool
	skipSetupFlag bool
)

func init() {
	flag.BoolVar(&releasedFlag, "released", false, "released controls whether we use a release image")
	flag.BoolVar(&skipSetupFlag, "skip-setup", false, "skip kubestellar cleanup and setup")
}

var _ = ginkgo.BeforeSuite(func() {
	if !skipSetupFlag {
		util.Cleanup()
		util.SetupKubestellar(releasedFlag)
	}

	ctx = context.Background()
	configCore := util.GetConfig("kind-kubeflex")
	configWds := util.GetConfig("wds1")
	configITS := util.GetConfig("imbs1")
	configWec1 := util.GetConfig("cluster1")
	configWec2 := util.GetConfig("cluster2")
	coreCluster = util.CreateKubeClient(configCore)
	wds = util.CreateKubeClient(configWds)
	ksWds = util.CreateKSClient(configWds)
	its = util.CreateOcmWorkClient(configITS)
	wec1 = util.CreateKubeClient(configWec1)
	wec2 = util.CreateKubeClient(configWec2)

	util.CreateNS(ctx, wds, ns)
})
