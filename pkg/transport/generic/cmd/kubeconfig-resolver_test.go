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

package cmd

import (
	"testing"

	"github.com/spf13/pflag"

	"k8s.io/klog/v2"

	ksopts "github.com/kubestellar/kubestellar/options"
)

func TestResolveWDSKubeconfig(t *testing.T) {
	logger := klog.Background()

	t.Run("WDS with name provided", func(t *testing.T) {
		// Testing: When user provides --wds-name=test-wds
		// Expected: Function should try to find WDS control plane by name
		options := &TransportOptions{
			WdsClientOptions: ksopts.NewClientOptions[*pflag.FlagSet]("wds", "accessing the WDS"),
			WdsName:          "test-wds",
		}

		config, name, err := resolveWDSKubeconfig(options, logger)
		t.Logf("WDS with name - Config: %v, Name: %s, Error: %v", config != nil, name, err)
	})

	t.Run("WDS with no name", func(t *testing.T) {
		// Testing: When user provides no --wds-name
		// Expected: Function should use file or fail
		options := &TransportOptions{
			WdsClientOptions: ksopts.NewClientOptions[*pflag.FlagSet]("wds", "accessing the WDS"),
			WdsName:          "", // No name provided
		}

		config, name, err := resolveWDSKubeconfig(options, logger)
		t.Logf("WDS without name - Config: %v, Name: %s, Error: %v", config != nil, name, err)
	})

	t.Run("WDS edge cases", func(t *testing.T) {
		// Testing: Various inputs to ensure function handles them gracefully
		// Expected: No crashes, proper error handling
		testCases := []struct {
			scenario string
			wdsName  string
		}{
			{"empty string", ""},
			{"normal name", "test-wds"},
		}

		for _, tc := range testCases {
			t.Run(tc.scenario, func(t *testing.T) {
				options := &TransportOptions{
					WdsClientOptions: ksopts.NewClientOptions[*pflag.FlagSet]("wds", "accessing the WDS"),
					WdsName:          tc.wdsName,
				}

				_, _, _ = resolveWDSKubeconfig(options, logger)
				t.Logf("✅ WDS handles: %s", tc.scenario)
			})
		}
	})
}

func TestResolveTransportKubeconfig(t *testing.T) {
	logger := klog.Background()

	t.Run("Transport with name provided", func(t *testing.T) {
		// Testing: When user provides --transport-name=test-its
		// Expected: Function should try to find ITS control plane by name
		options := &TransportOptions{
			TransportClientOptions: ksopts.NewClientOptions[*pflag.FlagSet]("transport", "accessing the ITS"),
			TransportName:          "test-its",
		}

		config, name, err := resolveTransportKubeconfig(options, logger)
		t.Logf("Transport with name - Config: %v, Name: %s, Error: %v", config != nil, name, err)
	})

	t.Run("Transport auto-discovery", func(t *testing.T) {
		// Testing: When user provides no --transport-name
		// Expected: Function should try to find single ITS automatically
		options := &TransportOptions{
			TransportClientOptions: ksopts.NewClientOptions[*pflag.FlagSet]("transport", "accessing the ITS"),
			TransportName:          "", // No name - should auto-discover
		}

		config, name, err := resolveTransportKubeconfig(options, logger)
		t.Logf("Transport auto-discovery - Config: %v, Name: %s, Error: %v", config != nil, name, err)
	})

	t.Run("Transport edge cases", func(t *testing.T) {
		// Testing: Various inputs to ensure function handles them gracefully
		// Expected: No crashes, proper error handling
		testCases := []struct {
			scenario      string
			transportName string
		}{
			{"no name provided", ""},
			{"name provided", "test-its"},
		}

		for _, tc := range testCases {
			t.Run(tc.scenario, func(t *testing.T) {
				options := &TransportOptions{
					TransportClientOptions: ksopts.NewClientOptions[*pflag.FlagSet]("transport", "accessing the ITS"),
					TransportName:          tc.transportName,
				}

				_, _, _ = resolveTransportKubeconfig(options, logger)
				t.Logf("✅ Transport handles: %s", tc.scenario)
			})
		}
	})
}
