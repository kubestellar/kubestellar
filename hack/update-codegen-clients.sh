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

SCRIPT_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)

if [ "$SCRIPT_ROOT" = "${SCRIPT_ROOT%/github.com/kubestellar/kubestellar}" ]; then
    cat >&2 <<EOF
Your local copy of the kubestellar repository needs to be elsewhere.
It is currently at '$SCRIPT_ROOT'.
Due to a restriction in k8s.io/code-generator (its Issue 165),
your local copy needs to be in a directory whose pathname
ends in '/github.com/kubestellar/kubestellar'.
EOF
    exit 1
fi

source "${CODE_GEN_DIR}/kube_codegen.sh"

rm -rf "${SCRIPT_ROOT}/pkg/generated"

kube::codegen::gen_client \
    --with-watch \
    --input-pkg-root github.com/kubestellar/kubestellar/api \
    --output-pkg-root github.com/kubestellar/kubestellar/pkg/generated \
    --output-base "${SCRIPT_ROOT}/../../.." \
    --boilerplate "${SCRIPT_ROOT}/hack/boilerplate/boilerplate.generatego.txt"
