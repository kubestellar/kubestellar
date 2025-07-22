#!/usr/bin/env bash
# Copyright 2024 The KubeStellar Authors.
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

# Simple test to verify ArgoCD is working with KubeStellar
set -x
set -e

env="kind"
if [ "$1" == "--env" ]; then
    env="$2"
fi

SRC_DIR="$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)"
COMMON_SRCS="${SRC_DIR}/../common"
source "$COMMON_SRCS/setup-shell.sh"

"${SRC_DIR}/../../../scripts/check_pre_req.sh" --assert --verbose kind kubectl helm ko

source "$COMMON_SRCS/setup-kubestellar.sh" --env "$env" --argocd

case "$env" in
    (kind) HOSTING_CONTEXT=kind-kubeflex;;
    (ocp)  HOSTING_CONTEXT=kscore;;
esac

:
: -------------------------------------------------------------------------
: "Test ArgoCD is working"
:

echo "Checking ArgoCD pods..."
wait-for-cmd 'kubectl --context '"$HOSTING_CONTEXT"' get pods -A | grep argocd | grep Running | wc -l | grep -v ^0$'

ARGOCD_POD=$(kubectl --context "$HOSTING_CONTEXT" get pods -A -l app.kubernetes.io/name=argocd-server -o 'jsonpath={.items[0].metadata.name}')
ARGOCD_NS=$(kubectl --context "$HOSTING_CONTEXT" get pods -A -l app.kubernetes.io/name=argocd-server -o 'jsonpath={.items[0].metadata.namespace}')

ARGOCD_PASSWORD=$(kubectl --context "$HOSTING_CONTEXT" -n "$ARGOCD_NS" get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d)
kubectl --context "$HOSTING_CONTEXT" -n "$ARGOCD_NS" exec "$ARGOCD_POD" -- argocd login ks-core-argocd-server."$ARGOCD_NS" --username admin --password "$ARGOCD_PASSWORD" --insecure

kubectl --context "$HOSTING_CONTEXT" -n "$ARGOCD_NS" exec "$ARGOCD_POD" -- argocd cluster list | grep wds1

echo "SUCCESS: ArgoCD is working!"
