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

var (
	ctx               context.Context
	coreClusterClient *kubernetes.Clientset
	wdsClient         *kubernetes.Clientset
	ksWdsClient       *ksClient.Clientset
	wec1Client        *kubernetes.Clientset
	wec2Client        *kubernetes.Clientset
	itsClient         *ocmWorkClient.Clientset
	releasedFlag      bool
	skipSetupFlag     bool
	justSummary       bool
)

func TestGinkgo(t *testing.T) {
	gomega.SetDefaultEventuallyTimeout(time.Minute)
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Kubestellar testing")
}

func init() {
	flag.BoolVar(&releasedFlag, "released", false, "released controls whether we use a release image")
	flag.BoolVar(&skipSetupFlag, "skip-setup", false, "skip kubestellar cleanup and setup")
	flag.BoolVar(&justSummary, "just-summary", false, "just print the summary info")
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
	coreClusterClient = util.CreateKubeClient(configCore)
	wdsClient = util.CreateKubeClient(configWds)
	ksWdsClient = util.CreateKSClient(configWds)
	itsClient = util.CreateOcmWorkClient(configITS)
	wec1Client = util.CreateKubeClient(configWec1)
	wec2Client = util.CreateKubeClient(configWec2)

	util.CreateNS(ctx, wdsClient, ns)
})
