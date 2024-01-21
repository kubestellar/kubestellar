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

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

source "${CODE_GEN_DIR}/kube_codegen.sh"

rm -rf "${SCRIPT_ROOT}/pkg/generated"

kube::codegen::gen_helpers \
    --input-pkg-root github.com/kubestellar/kubestellar/api \
    --output-base "${SCRIPT_ROOT}/../../.." \
    --boilerplate "${SCRIPT_ROOT}/hack/boilerplate/boilerplate.generatego.txt"

kube::codegen::gen_client \
    --with-watch \
    --input-pkg-root github.com/kubestellar/kubestellar/api \
    --output-pkg-root github.com/kubestellar/kubestellar/pkg/generated \
    --output-base "${SCRIPT_ROOT}/../../.." \
    --boilerplate "${SCRIPT_ROOT}/hack/boilerplate/boilerplate.generatego.txt"
