/*
Copyright 2022 The KubeStellar Authors.

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

/*
Portions of this code are based on or inspired by the KCP Author's work
The original code is https://github.com/kcp-dev/kcp/blob/release-0.11/test/e2e/framework/suites.go,
https://github.com/kcp-dev/kcp/blob/release-0.11/test/e2e/framework/base.go
The original copyright is as follows:
*/

/*
Copyright 2022 The KCP Authors.

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

package framework

import (
	"flag"
	"path/filepath"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
)

type testConfig struct {
	kcpKubeconfig string
	suites        string
}

var TestConfig *testConfig

func (c *testConfig) KCPKubeconfig() string {
	if c.kcpKubeconfig != "" {
		return c.kcpKubeconfig
	}
	return filepath.Join(repositoryDir(), ".kcp", "admin.kubeconfig")
}

func (c *testConfig) Suites() []string {
	return strings.Split(c.suites, ",")
}

// Suite should be called at the very beginning of a test case, to ensure that a test is only
// run when the suite containing it is selected by the user running tests.
func Suite(t *testing.T, suite string) {
	t.Helper()
	if !sets.NewString(TestConfig.Suites()...).Has(suite) {
		t.Skipf("suite %s disabled", suite)
	}
}

func init() {
	klog.InitFlags(flag.CommandLine)
	if err := flag.Lookup("v").Value.Set("4"); err != nil {
		panic(err)
	}
	TestConfig = &testConfig{}
	registerFlags(TestConfig)
	// The testing package will call flags.Parse()
}

func registerFlags(c *testConfig) {
	flag.StringVar(&c.kcpKubeconfig, "kcp-kubeconfig", "", "Path to the kubeconfig for a kcp server.")
	flag.StringVar(&c.suites, "suites", "kubestellar-syncer", "A comma-delimited list of suites to run.")
}
