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

// This is the entrypoint for KubeStellar sub-commands for kubectl.
// This should be compiled as an executable named "kubectl-kubestellar",
// allowing it to be used as a plugin to kubectl. This entry-level command
// is executed as "kubectl kubestellar", with sub-commands following.

package main

import (
	"github.com/kubestellar/kubestellar/cmd/kubectl-kubestellar/cmd"
)

// Run the root command. Error handling is done within Execute().
func main() {
	cmd.Execute()
}
