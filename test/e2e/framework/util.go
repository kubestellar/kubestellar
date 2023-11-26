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
Copyright Copyright 2022 The KCP Authors
Original code: https://github.com/kcp-dev/kcp/blob/release-0.11/test/e2e/framework/util.go,
https://github.com/kcp-dev/kcp/blob/release-0.11/test/e2e/framework/kubectl.go
*/

package framework

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"k8s.io/client-go/rest"
)

// Persistent mapping of test name to base temp dir used to ensure
// artifact paths have a common root across servers for a given test.
var (
	baseTempDirs     = map[string]string{}
	baseTempDirsLock = sync.Mutex{}
)

// CreateTempDirForTest creates the named directory with a unique base
// path derived from the name of the current test.
func CreateTempDirForTest(t *testing.T, dirName string) (string, error) {
	t.Helper()
	baseTempDir, err := ensureBaseTempDir(t)
	if err != nil {
		return "", err
	}
	dir := filepath.Join(baseTempDir, dirName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("could not create subdir: %w", err)
	}
	return dir, nil
}

// Eventually asserts that given condition will be met in waitFor time, periodically checking target function
// each tick. In addition to require.Eventually, this function t.Logs the raason string value returned by the condition
// function (eventually after 20% of the wait time) to aid in debugging.
func Eventually(t *testing.T, condition func() (success bool, reason string), waitFor time.Duration, tick time.Duration, msgAndArgs ...interface{}) {
	t.Helper()

	var last string
	start := time.Now()
	require.Eventually(t, func() bool {
		t.Helper()

		ok, msg := condition()
		if time.Since(start) > waitFor/5 {
			if !ok && msg != "" && msg != last {
				last = msg
				t.Logf("Waiting for condition, but got: %s", msg)
			} else if ok && msg != "" && last != "" {
				t.Logf("Condition became true: %s", msg)
			}
		}
		return ok
	}, waitFor, tick, msgAndArgs...)
}

// KubectlApply runs kubectl apply -f with the supplied input piped to stdin and returns
// the combined stderr and stdout output.
func KubectlApply(t *testing.T, kubeconfigPath string, input []byte) []byte {
	t.Helper()

	ctx, cancelFunc := context.WithCancel(context.Background())
	t.Cleanup(cancelFunc)

	cmdParts := []string{"kubectl", "--kubeconfig", kubeconfigPath, "apply", "-f", "-"}
	cmd := exec.CommandContext(ctx, cmdParts[0], cmdParts[1:]...)
	stdin, err := cmd.StdinPipe()
	require.NoError(t, err)
	_, err = stdin.Write(input)
	require.NoError(t, err)
	// Close to ensure kubectl doesn't keep waiting for input
	err = stdin.Close()
	require.NoError(t, err)

	t.Logf("running: %s", strings.Join(cmdParts, " "))

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("kubectl apply output:\n%s", output)
	}
	require.NoError(t, err)

	return output
}

// Kubectl runs kubectl with the given arguments and returns the combined stderr and stdout.
func Kubectl(t *testing.T, kubeconfigPath string, args ...string) []byte {
	t.Helper()

	ctx, cancelFunc := context.WithCancel(context.Background())
	t.Cleanup(cancelFunc)

	cmdParts := append([]string{"kubectl", "--kubeconfig", kubeconfigPath}, args...)
	cmd := exec.CommandContext(ctx, cmdParts[0], cmdParts[1:]...)
	t.Logf("running: %s", strings.Join(cmdParts, " "))

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("kubectl output:\n%s", output)
	}
	require.NoError(t, err)

	return output
}

func ConfigWithToken(token string, cfg *rest.Config) *rest.Config {
	cfgCopy := rest.CopyConfig(cfg)
	cfgCopy.CertData = nil
	cfgCopy.KeyData = nil
	cfgCopy.BearerToken = token
	return cfgCopy
}

// ensureBaseTempDir returns the name of a base temp dir for the
// current test, creating it if needed.
func ensureBaseTempDir(t *testing.T) (string, error) {
	t.Helper()

	baseTempDirsLock.Lock()
	defer baseTempDirsLock.Unlock()
	name := t.Name()
	if _, ok := baseTempDirs[name]; !ok {
		var baseDir string
		if dir, set := os.LookupEnv("ARTIFACT_DIR"); set {
			baseDir = dir
		} else {
			baseDir = t.TempDir()
		}
		baseDir = filepath.Join(baseDir, strings.NewReplacer("\\", "_", ":", "_").Replace(t.Name()))
		if err := os.MkdirAll(baseDir, 0755); err != nil {
			return "", fmt.Errorf("could not create base dir: %w", err)
		}
		baseTempDir, err := os.MkdirTemp(baseDir, "")
		if err != nil {
			return "", fmt.Errorf("could not create base temp dir: %w", err)
		}
		baseTempDirs[name] = baseTempDir
		t.Logf("Saving test artifacts for test %q under %q.", name, baseTempDir)

		// Remove the path from the cache after test completion to
		// ensure subsequent invocations of the test (e.g. due to
		// -count=<val> for val > 1) don't reuse the same path.
		t.Cleanup(func() {
			baseTempDirsLock.Lock()
			defer baseTempDirsLock.Unlock()
			delete(baseTempDirs, name)
		})
	}
	return baseTempDirs[name], nil
}

func repositoryDir() string {
	// Caller(0) returns the path to the calling test file rather than the path to this framework file. That
	// precludes assuming how many directories are between the file and the repo root. It's therefore necessary
	// to search in the hierarchy for an indication of a path that looks like the repo root.
	_, sourceFile, _, _ := runtime.Caller(0)
	currentDir := filepath.Dir(sourceFile)
	for {
		// go.mod should always exist in the repo root
		if _, err := os.Stat(filepath.Join(currentDir, "go.mod")); err == nil {
			break
		} else if errors.Is(err, os.ErrNotExist) {
			currentDir, err = filepath.Abs(filepath.Join(currentDir, ".."))
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}
	return currentDir
}
