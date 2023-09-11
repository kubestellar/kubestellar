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

if [[ -z "${CONTROLLER_GEN:-}" ]]; then
    echo "$CONTROLLER_GEN must be defined to refer to the right binary, e.g. as is done in the Makefile"
    exit 1
fi
if [[ -z "${API_GEN:-}" ]]; then
    echo "$API_GEN must be defined to refer to the right binary, e.g. as is done in the Makefile"
    exit 1
fi

REPO_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)

# Update generated CRD YAML
cd "${REPO_ROOT}/pkg/apis"
"../../${CONTROLLER_GEN}" \
    crd \
    rbac:roleName=manager-role \
    webhook \
    paths="./..." \
    output:crd:artifacts:config="${REPO_ROOT}"/config/crds

cd "${REPO_ROOT}/space-framework/pkg/apis"
"../../../${CONTROLLER_GEN}" \
    crd \
    rbac:roleName=manager-role \
    webhook \
    paths="./..." \
    output:crd:artifacts:config="${REPO_ROOT}"/space-framework/config/crds

for CRD in "${REPO_ROOT}"/config/crds/*.yaml "${REPO_ROOT}"/space-framework/config/crds/*.yaml; do
    if [ -f "${CRD}-patch" ]; then
        echo "Applying ${CRD}"
        ${YAML_PATCH} -o "${CRD}-patch" < "${CRD}" > "${CRD}.patched"
        mv "${CRD}.patched" "${CRD}"
    fi
done

# ${CONTROLLER_GEN} \
#     crd \
#     rbac:roleName=manager-role \
#     webhook \
#     paths="${REPO_ROOT}/test/e2e/reconciler/cluster/..." \
#     output:crd:artifacts:config="${REPO_ROOT}"/test/e2e/reconciler/cluster/

"${REPO_ROOT}/${API_GEN}" --input-dir "${REPO_ROOT}"/config/crds --output-dir "${REPO_ROOT}"/config/exports
"${REPO_ROOT}/${API_GEN}" --input-dir "${REPO_ROOT}"/space-framework/config/crds --output-dir "${REPO_ROOT}"/space-framework/config/exports
