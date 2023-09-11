#!/usr/bin/env bash

# Copyright 2021 The KubeStellar Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

export GOPATH=$(go env GOPATH)

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
pushd "${SCRIPT_ROOT}"
BOILERPLATE_HEADER="$( pwd )/hack/boilerplate/boilerplate.go.txt"
popd
CODEGEN_PKG=${CODEGEN_PKG:-$(cd "${SCRIPT_ROOT}"; go list -f '{{.Dir}}' -m k8s.io/code-generator)}
set
exit

bash "${CODEGEN_PKG}"/generate-groups.sh "deepcopy,client" \
  github.com/kubestellar/kubestellar/space-framework/pkg/client github.com/kubestellar/kubestellar/space-framework/pkg/apis \
  "space:v1alpha1" \
  --go-header-file "${SCRIPT_ROOT}"/hack/boilerplate/boilerplate.generatego.txt \
  --output-base "${SCRIPT_ROOT}" \
  --trim-path-prefix github.com/kubestellar/kubestellar


pushd ./pkg/apis
${CODE_GENERATOR} \
  "client:outputPackagePath=github.com/kubestellar/kubestellar/space-framework/pkg/client,apiPackagePath=github.com/kubestellar/kubestellar/space-framework/pkg/apis,singleClusterClientPackagePath=github.com/kubestellar/kubestellar/space-framework/pkg/client/clientset/versioned,headerFile=${BOILERPLATE_HEADER}" \
  "lister:apiPackagePath=github.com/kubestellar/kubestellar/space-framework/pkg/apis,headerFile=${BOILERPLATE_HEADER}" \
  "informer:outputPackagePath=github.com/kubestellar/kubestellar/space-framework/pkg/client,apiPackagePath=github.com/kubestellar/kubestellar/space-framework/pkg/apis,singleClusterClientPackagePath=github.com/kubestellar/kubestellar/space-framework/pkg/client/clientset/versioned,headerFile=${BOILERPLATE_HEADER}" \
  "paths=./..." \
  "output:dir=./../client"
popd


go install "${CODEGEN_PKG}"/cmd/openapi-gen