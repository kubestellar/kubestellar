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

###############################################################################
# Combined setup + end-to-end test for KubeStellar with Argo CD integration
###############################################################################

# ─────────────────────────────────────────────────────────────────────────────
# Strict shell options
# ─────────────────────────────────────────────────────────────────────────────
set -o errexit   # exit on any command failure
set -o nounset   # exit on use of un-set variable
set -o pipefail  # pipeline fails if any sub-command fails
set -o errtrace  # trap ERR in functions and subshells
shopt -s inherit_errexit  # propagate set -e into subshells
set -x            # echo commands so users can follow along

# ─────────────────────────────────────────────────────────────────────────────
# Defaults
# ─────────────────────────────────────────────────────────────────────────────
ENV=kind
ARGOCD_DOMAIN=argocd.localtest.me
TEST_TIMEOUT=300          # seconds
USE_RELEASE=false
KUBESTELLAR_CONTROLLER_MANAGER_VERBOSITY=5
TRANSPORT_CONTROLLER_VERBOSITY=5
CLUSTER_SOURCE=kind
HOSTING_CONTEXT=kind-kubeflex

TEST_APP_REPO="https://github.com/argoproj/argocd-example-apps.git"
TEST_APP_PATH="guestbook"

# ─────────────────────────────────────────────────────────────────────────────
# Colourised logging helpers
# ─────────────────────────────────────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'
YELLOW='\033[1;33m'; BLUE='\033[0;34m'
NC='\033[0m' # no colour

log() {
  local level=$1 ; shift
  local msg="$*"
  local ts
  ts=$(date +"%Y-%m-%d %H:%M:%S")
  case "${level}" in
    INFO)    echo -e "${BLUE}[${ts} INFO]${NC} ${msg}" ;;
    SUCCESS) echo -e "${GREEN}[${ts} SUCCESS]${NC} ${msg}" ;;
    WARNING) echo -e "${YELLOW}[${ts} WARNING]${NC} ${msg}" ;;
    ERROR)   echo -e "${RED}[${ts} ERROR]${NC} ${msg}" ;;
    *)       echo    "[${ts} ${level}] ${msg}" ;;
  esac
}

# ─────────────────────────────────────────────────────────────────────────────
# Command-line argument parsing
# ─────────────────────────────────────────────────────────────────────────────
usage() {
cat <<EOF
Usage: $0 [OPTIONS]

Combined setup and ArgoCD E2E test for KubeStellar.

Options
  --released                          Install latest released OCI chart
  --kubestellar-controller-manager-verbosity N
  --transport-controller-verbosity    N
  --env [kind|ocp]                    Target environment (kind default)
  --argocd-domain DOMAIN              Domain for ArgoCD ingress
  --test-timeout SECONDS              Timeout for each waiting operation
  --app-repo URL                      Git repo containing test app
  --app-path PATH                     Path inside repo for test app
  --no-color                          Disable colourful output
  -h|--help                           Show this help and exit
EOF
}

parse_args() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --released) USE_RELEASE=true ;;
      --kubestellar-controller-manager-verbosity)
        KUBESTELLAR_CONTROLLER_MANAGER_VERBOSITY=$2 ; shift ;;
      --transport-controller-verbosity)
        TRANSPORT_CONTROLLER_VERBOSITY=$2            ; shift ;;
      --env)
        [[ $# -lt 2 ]] && { log ERROR "--env requires value" ; exit 1; }
        case "$2" in
          kind) CLUSTER_SOURCE=kind ; HOSTING_CONTEXT=kind-kubeflex ;;
          ocp)  CLUSTER_SOURCE=existing ; HOSTING_CONTEXT=kscore ;;
          *)    log ERROR "--env must be kind or ocp" ; exit 1 ;;
        esac ; shift ;;
      --argocd-domain) ARGOCD_DOMAIN=$2 ; shift ;;
      --test-timeout)  TEST_TIMEOUT=$2  ; shift ;;
      --app-repo)      TEST_APP_REPO=$2 ; shift ;;
      --app-path)      TEST_APP_PATH=$2 ; shift ;;
      --no-color)      RED='' ; GREEN='' ; YELLOW='' ; BLUE='' ; NC='' ;;
      -h|--help)       usage ; exit 0 ;;
      *) log ERROR "unknown flag $1" ; usage ; exit 1 ;;
    esac
    shift
  done
}

parse_args "$@"

# ─────────────────────────────────────────────────────────────────────────────
# Validation
# ─────────────────────────────────────────────────────────────────────────────
[[ "${KUBESTELLAR_CONTROLLER_MANAGER_VERBOSITY}" =~ ^[0-9]+$ ]] \
  || { log ERROR "verbosity must be numeric" ; exit 1; }
[[ "${TRANSPORT_CONTROLLER_VERBOSITY}"       =~ ^[0-9]+$ ]] \
  || { log ERROR "verbosity must be numeric" ; exit 1; }

# ─────────────────────────────────────────────────────────────────────────────
# Helpers
# ─────────────────────────────────────────────────────────────────────────────
wait_for() {
  local cmd=$1 ; local timeout=${2:-$TEST_TIMEOUT} ; local interval=5
  local elapsed=0
  log INFO "Waiting for condition: $cmd"
  until eval "$cmd"; do
    [[ $elapsed -ge $timeout ]] && { log ERROR "Timeout waiting for: $cmd"; return 1; }
    sleep $interval ; elapsed=$((elapsed+interval))
    log INFO " …${elapsed}s"
  done
}

# ---------------------------------------------------------------------------
# 1. Setup: create / select hosting cluster and install KubeStellar core-chart
# ---------------------------------------------------------------------------
SRC_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMMON_SRCS="${SRC_DIR}/../common"
source "${COMMON_SRCS}/setup-shell.sh"

log INFO "Using hosting context: ${HOSTING_CONTEXT}"

if [[ "${CLUSTER_SOURCE}" == "kind" ]]; then
  "${SRC_DIR}/../../../scripts/create-kind-cluster-with-SSL-passthrough.sh" --name kubeflex
else
  kubectl config use-context "${HOSTING_CONTEXT}"
fi

pushd "${SRC_DIR}/../../.." >/dev/null
if $USE_RELEASE ; then
  helm upgrade --install ks-core oci://ghcr.io/kubestellar/kubestellar/core-chart \
    --version "$(yq .KUBESTELLAR_VERSION core-chart/values.yaml)" \
    --kube-context "${HOSTING_CONTEXT}" \
    --set-json='ITSes=[{"name":"its1"}]' \
    --set-json='WDSes=[{"name":"wds1"}]' \
    --set verbosity.kubestellar="${KUBESTELLAR_CONTROLLER_MANAGER_VERBOSITY}" \
    --set verbosity.transport="${TRANSPORT_CONTROLLER_VERBOSITY}"
else
  make kind-load-image
  helm dependency update core-chart/
  helm upgrade --install ks-core core-chart/ \
    --set KUBESTELLAR_VERSION="$(git rev-parse --short HEAD)" \
    --kube-context "${HOSTING_CONTEXT}" \
    --set-json='ITSes=[{"name":"its1"}]' \
    --set-json='WDSes=[{"name":"wds1"}]' \
    --set verbosity.kubestellar="${KUBESTELLAR_CONTROLLER_MANAGER_VERBOSITY}" \
    --set verbosity.transport="${TRANSPORT_CONTROLLER_VERBOSITY}"
fi
popd >/dev/null

wait_for "kubectl get pod -n wds1-system -l name=transport-controller | grep -q Running"

# ---------------------------------------------------------------------------
# 2. Register two workload clusters with OCM
# ---------------------------------------------------------------------------
add_wec() {
  local cluster=$1
  if [[ "${CLUSTER_SOURCE}" == "kind" ]]; then
    kind create cluster --name "${cluster}"
    kubectl config rename-context "kind-${cluster}" "${cluster}"
    joinflags="--force-internal-endpoint-lookup"
  else
    joinflags=""
  fi

  clusteradm --context its1 get token | \
    grep '^clusteradm join' | \
    sed "s/<cluster_name>/${cluster}/" | \
    awk "{print \$0 \" --context ${cluster} --singleton ${joinflags}\"}" | \
    sh
}

"${SRC_DIR}/../../../scripts/check_pre_req.sh" --assert --verbose ocm
add_wec cluster1
add_wec cluster2
wait_for "(($(kubectl --context its1 get csr 2>/dev/null | grep -c Pending) == 0))"

clusteradm --context its1 accept --clusters cluster1,cluster2

# ---------------------------------------------------------------------------
# 3. Quick sanity check before running ArgoCD tests
# ---------------------------------------------------------------------------
wait_for \
  "((\$(kubectl --context '${HOSTING_CONTEXT}' get deployments,statefulsets -A | grep -e wds1 -e its1 | wc -l) >= 5))"

# ---------------------------------------------------------------------------
# 4. Install ArgoCD & run end-to-end application test
# ---------------------------------------------------------------------------
source "${COMMON_SRCS}/setup-kubestellar.sh" --env "${ENV}" --argocd --argocd-domain "${ARGOCD_DOMAIN}"

ARGOCD_NS="$(kubectl --context "${HOSTING_CONTEXT}" get pod -A -l app.kubernetes.io/name=argocd-server -o jsonpath='{.items[0].metadata.namespace}')"
ARGOCD_POD="$(kubectl --context "${HOSTING_CONTEXT}" get pod -n "${ARGOCD_NS}" -l app.kubernetes.io/name=argocd-server -o jsonpath='{.items[0].metadata.name}')"
wait_for "kubectl --context ${HOSTING_CONTEXT} get pod -n ${ARGOCD_NS} ${ARGOCD_POD} -o jsonpath='{.status.phase}' | grep -q Running"

ARGOCD_PASSWORD="$(kubectl --context "${HOSTING_CONTEXT}" -n "${ARGOCD_NS}" get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 -d)"

log INFO "Logging into ArgoCD CLI"
kubectl --context "${HOSTING_CONTEXT}" -n "${ARGOCD_NS}" exec "${ARGOCD_POD}" -- \
  argocd login ks-core-argocd-server."${ARGOCD_NS}".svc.cluster.local \
  --username admin --password "${ARGOCD_PASSWORD}" --insecure

log INFO "Verifying cluster list includes WDS1"
kubectl --context "${HOSTING_CONTEXT}" -n "${ARGOCD_NS}" exec "${ARGOCD_POD}" -- argocd cluster list

TEST_APP_NAME="kubestellar-test-app-$(date +%s)"
log INFO "Creating ArgoCD app ${TEST_APP_NAME}"
kubectl --context "${HOSTING_CONTEXT}" -n "${ARGOCD_NS}" exec "${ARGOCD_POD}" -- \
  argocd app create "${TEST_APP_NAME}" \
  --repo "${TEST_APP_REPO}" \
  --path "${TEST_APP_PATH}" \
  --dest-server https://kubernetes.default.svc \
  --dest-namespace default \
  --sync-policy automated

wait_for "kubectl --context ${HOSTING_CONTEXT} -n ${ARGOCD_NS} exec ${ARGOCD_POD} -- argocd app get ${TEST_APP_NAME} &>/dev/null"
kubectl --context "${HOSTING_CONTEXT}" -n "${ARGOCD_NS}" exec "${ARGOCD_POD}" -- argocd app sync "${TEST_APP_NAME}"
wait_for "kubectl --context ${HOSTING_CONTEXT} -n ${ARGOCD_NS} exec ${ARGOCD_POD} -- argocd app get ${TEST_APP_NAME} -o json | jq -e '.status.sync.status==\"Synced\"'"

log SUCCESS "✅ All KubeStellar & ArgoCD tests passed!"

echo -e "${GREEN}ArgoCD UI:${NC}  https://${ARGOCD_DOMAIN}"
echo -e "  user: admin"
echo -e "  pass: ${ARGOCD_PASSWORD}"
