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

# Enhanced KubeStellar setup script with ArgoCD integration
set -o errexit          # Exit immediately if any command fails
set -o nounset          # Exit if any variable is unset
set -o pipefail         # Fail pipeline if any command in pipeline fails
set -o errtrace         # Enable error tracing
shopt -s inherit_errexit  # Ensure subshells inherit error handling

# Configuration defaults
DEFAULT_ENV="kind"
DEFAULT_ARGOCD_DOMAIN="argocd.localtest.me"
DEFAULT_KUBESTELLAR_VERBOSITY=5
DEFAULT_TRANSPORT_VERBOSITY=5
DEFAULT_CLUSTER_SOURCE="kind"
DEFAULT_HOSTING_CONTEXT="kind-kubeflex"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Global variables
declare -g CLUSTER_SOURCE="$DEFAULT_CLUSTER_SOURCE"
declare -g HOSTING_CONTEXT="$DEFAULT_HOSTING_CONTEXT"
declare -g ARGOCD_INSTALL=false
declare -g USE_RELEASE=false
declare -g KUBESTELLAR_VERBOSITY="$DEFAULT_KUBESTELLAR_VERBOSITY"
declare -g TRANSPORT_VERBOSITY="$DEFAULT_TRANSPORT_VERBOSITY"
declare -g ARGOCD_DOMAIN="$DEFAULT_ARGOCD_DOMAIN"
declare -g SRC_DIR=""
declare -g COMMON_SRCS=""

# Initialize logging
LOG_FILE="/tmp/kubestellar-setup-$(date +%Y%m%d%H%M%S).log"
exec > >(tee -a "$LOG_FILE") 2>&1

# Function to log messages with timestamp and color
log() {
    local level=$1
    local message=$2
    local timestamp
    timestamp=$(date +"%Y-%m-%d %H:%M:%S")
    
    case "$level" in
        "INFO") echo -e "${BLUE}[${timestamp} INFO]${NC} ${message}" ;;
        "SUCCESS") echo -e "${GREEN}[${timestamp} SUCCESS]${NC} ${message}" ;;
        "WARNING") echo -e "${YELLOW}[${timestamp} WARNING]${NC} ${message}" ;;
        "ERROR") echo -e "${RED}[${timestamp} ERROR]${NC} ${message}" ;;
        *) echo -e "[${timestamp} ${level}] ${message}" ;;
    esac
}

# Display help information
usage() {
    cat <<EOF
Usage: $0 [OPTIONS]

KubeStellar setup script with optional ArgoCD integration

Options:
  --env [kind|ocp]            Environment type (default: ${DEFAULT_ENV})
  --argocd                    Install ArgoCD (default: false)
  --argocd-domain DOMAIN      ArgoCD domain (default: ${DEFAULT_ARGOCD_DOMAIN})
  --released                  Use released version (default: false)
  --kubestellar-verbosity N   KubeStellar controller manager verbosity (default: ${DEFAULT_KUBESTELLAR_VERBOSITY})
  --transport-verbosity N     Transport controller verbosity (default: ${DEFAULT_TRANSPORT_VERBOSITY})
  --no-color                  Disable colored output
  --help                      Show this help message

EOF
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --env)
                validate_environment "$2"
                case "$2" in
                    kind) 
                        CLUSTER_SOURCE="kind"
                        HOSTING_CONTEXT="kind-kubeflex"
                        ;;
                    ocp)  
                        CLUSTER_SOURCE="existing"
                        HOSTING_CONTEXT="kscore"
                        ;;
                esac
                shift 2
                ;;
            --argocd)
                ARGOCD_INSTALL=true
                shift
                ;;
            --argocd-domain)
                ARGOCD_DOMAIN="$2"
                shift 2
                ;;
            --released)
                USE_RELEASE=true
                shift
                ;;
            --kubestellar-verbosity)
                validate_verbosity "$2"
                KUBESTELLAR_VERBOSITY="$2"
                shift 2
                ;;
            --transport-verbosity)
                validate_verbosity "$2"
                TRANSPORT_VERBOSITY="$2"
                shift 2
                ;;
            --no-color)
                RED=''; GREEN=''; YELLOW=''; BLUE=''; NC=''
                shift
                ;;
            --help|-h)
                usage
                exit 0
                ;;
            *)
                log "ERROR" "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done
}

# Validate environment parameter
validate_environment() {
    local env=$1
    case "$env" in
        kind|ocp) ;;
        *)
            log "ERROR" "Invalid environment: $env. Must be 'kind' or 'ocp'"
            exit 1
            ;;
    esac
}

# Validate verbosity level
validate_verbosity() {
    local verbosity=$1
    if ! [[ "$verbosity" =~ ^[0-9]+$ ]]; then
        log "ERROR" "Verbosity must be an integer"
        exit 1
    fi
}

# Check prerequisites with version validation
check_prerequisites() {
    log "INFO" "Checking system prerequisites..."
    
    local required_tools=("kind" "kubectl" "helm" "kubeflex" "clusteradm" "yq")
    local missing_tools=()
    
    for tool in "${required_tools[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            missing_tools+=("$tool")
        fi
    done
    
    if [[ ${#missing_tools[@]} -gt 0 ]]; then
        log "ERROR" "Missing required tools: ${missing_tools[*]}"
        exit 1
    fi
    
    log "SUCCESS" "All prerequisites are installed"
}

# Setup KubeFlex hosting cluster
setup_hosting_cluster() {
    log "INFO" "Setting up hosting cluster..."
    
    case "$CLUSTER_SOURCE" in
        (kind)
            log "INFO" "Creating kind cluster 'kubeflex'..."
            "${SRC_DIR}/../../../scripts/create-kind-cluster-with-SSL-passthrough.sh" --name kubeflex
            log "SUCCESS" "Kubeflex kind cluster created"
            ;;
        (existing)
            if ! kubectl config use-context "$HOSTING_CONTEXT" &> /dev/null; then
                log "ERROR" "Failed to switch to context '$HOSTING_CONTEXT'"
                exit 1
            fi
            log "INFO" "Using existing cluster in context '$HOSTING_CONTEXT'"
            ;;
    esac
}

# Install KubeStellar core chart
install_kubestellar() {
    log "INFO" "Installing KubeStellar core chart..."
    
    local argocd_params=""
    if [ "$ARGOCD_INSTALL" = true ]; then
        argocd_params="--set argocd.install=true"
        
        case "$CLUSTER_SOURCE" in
            (kind)
                argocd_params="$argocd_params --set argocd.global.domain=$ARGOCD_DOMAIN"
                ;;
            (existing)
                if [ "$HOSTING_CONTEXT" = "kscore" ]; then
                    argocd_params="$argocd_params --set argocd.openshift.enabled=true"
                else
                    argocd_params="$argocd_params --set argocd.global.domain=$ARGOCD_DOMAIN"
                fi
                ;;
        esac
        
        log "INFO" "ArgoCD installation enabled with parameters: $argocd_params"
    fi

    pushd "${SRC_DIR}/../../.." > /dev/null
    
    if [ "$USE_RELEASE" = true ]; then
        log "INFO" "Installing released version of KubeStellar"
        helm upgrade --install ks-core oci://ghcr.io/kubestellar/kubestellar/core-chart \
            --version $(yq .KUBESTELLAR_VERSION core-chart/values.yaml) \
            --kube-context "$HOSTING_CONTEXT" \
            --set-json='ITSes=[{"name":"its1"}]' \
            --set-json='WDSes=[{"name":"wds1"}]' \
            --set verbosity.kubestellar="$KUBESTELLAR_VERBOSITY" \
            --set verbosity.transport="$TRANSPORT_VERBOSITY" \
            $argocd_params
    else
        log "INFO" "Installing development version of KubeStellar"
        make kind-load-image
        helm dependency update core-chart/
        helm upgrade --install ks-core core-chart/ \
            --set KUBESTELLAR_VERSION=$(git rev-parse --short HEAD) \
            --kube-context "$HOSTING_CONTEXT" \
            --set-json='ITSes=[{"name":"its1"}]' \
            --set-json='WDSes=[{"name":"wds1"}]' \
            --set verbosity.kubestellar="$KUBESTELLAR_VERBOSITY" \
            --set verbosity.transport="$TRANSPORT_VERBOSITY" \
            $argocd_params
    fi
    
    popd > /dev/null
}

# Wait for OCM hub to be ready
wait_for_ocm_hub() {
    log "INFO" "Waiting for OCM hub to be ready..."
    
    if ! kubectl wait controlplane.tenancy.kflex.kubestellar.org/its1 \
        --for 'jsonpath={.status.postCreateHooks.its-with-clusteradm}=true' \
        --timeout 400s; then
        log "ERROR" "Timeout waiting for OCM hub to be ready"
        exit 1
    fi
    
    if ! kubectl wait -n its1-system job.batch/its-with-clusteradm \
        --for condition=Complete --timeout 400s; then
        log "ERROR" "Timeout waiting for clusteradm job to complete"
        exit 1
    fi
    
    if ! kubectl wait -n its1-system job.batch/update-cluster-info \
        --for condition=Complete --timeout 200s; then
        log "ERROR" "Timeout waiting for cluster-info update job to complete"
        exit 1
    fi
    
    log "SUCCESS" "OCM hub is ready"
}

# Wait for transport controller to be ready
wait_for_transport_controller() {
    log "INFO" "Waiting for transport controller to be ready..."
    
    if ! wait-for-cmd "(kubectl --context '$HOSTING_CONTEXT' -n wds1-system wait --for=condition=Ready pod/\$(kubectl --context '$HOSTING_CONTEXT' -n wds1-system get pods -l name=transport-controller -o jsonpath='{.items[0].metadata.name}'))"; then
        log "ERROR" "Timeout waiting for transport controller to be ready"
        exit 1
    fi
    
    log "SUCCESS" "Transport controller is running"
}

# Configure kubectl contexts
configure_contexts() {
    log "INFO" "Configuring kubectl contexts..."
    
    kubectl config use-context "$HOSTING_CONTEXT"
    kflex ctx --set-current-for-hosting
    kflex ctx --overwrite-existing-context wds1
    kflex ctx --overwrite-existing-context its1
    
    log "INFO" "Current contexts:"
    kflex ctx
}

# Wait for ArgoCD to be ready
wait_for_argocd() {
    if [ "$ARGOCD_INSTALL" != true ]; then
        return 0
    fi
    
    log "INFO" "Waiting for ArgoCD components to be ready..."
    
    local components=("argocd-server" "argocd-application-controller" "argocd-repo-server")
    local all_healthy=true
    
    for component in "${components[@]}"; do
        if ! wait-for-cmd "kubectl --context $HOSTING_CONTEXT get pods -A -l app.kubernetes.io/name=$component -o jsonpath='{.items[*].status.phase}' | grep -q Running"; then
            log "ERROR" "$component is not running"
            all_healthy=false
        fi
    done
    
    if [[ "$all_healthy" == false ]]; then
        log "ERROR" "Not all ArgoCD components are healthy"
        exit 1
    fi
    
    log "SUCCESS" "ArgoCD installation completed successfully"
    
    local ui_url="https://$ARGOCD_DOMAIN"
    if [ "$CLUSTER_SOURCE" = "kind" ]; then
        ui_url="${ui_url}:9443"
    fi
    
    cat <<EOF
${GREEN}ArgoCD Access Information:${NC}
  - UI: ${ui_url}
  - Username: admin
  - Password: $(kubectl -n default get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d)
EOF
}

# Add workload cluster
add_workload_cluster() {
    local cluster=$1
    local joinflags=""
    
    log "INFO" "Adding workload cluster: $cluster"
    
    case "$CLUSTER_SOURCE" in
        (kind)
            if ! kind create cluster --name "$cluster"; then
                log "ERROR" "Failed to create kind cluster: $cluster"
                exit 1
            fi
            kubectl config rename-context "kind-${cluster}" "$cluster"
            joinflags="--force-internal-endpoint-lookup"
            ;;
        (existing)
            log "INFO" "Using existing cluster: $cluster"
            ;;
    esac
    
    local join_command
    join_command=$(clusteradm --context its1 get token | grep '^clusteradm join' | sed "s/<cluster_name>/${cluster}/" | awk '{print $0 " --context '"${cluster}"' --singleton '"${joinflags}"'"}')
    
    if ! eval "$join_command"; then
        log "ERROR" "Failed to join cluster: $cluster"
        exit 1
    fi
    
    log "SUCCESS" "Cluster $cluster added successfully"
}

# Register clusters with OCM
register_clusters() {
    log "INFO" "Registering clusters with OCM..."
    
    "${SRC_DIR}/../../../scripts/check_pre_req.sh" --assert --verbose ocm
    
    if ! kubectl --context "$HOSTING_CONTEXT" wait controlplane.tenancy.kflex.kubestellar.org/its1 \
        --for 'jsonpath={.status.postCreateHooks.its-with-clusteradm}=true' --timeout 200s; then
        log "ERROR" "Timeout waiting for OCM control plane"
        exit 1
    fi
    
    if ! kubectl --context "$HOSTING_CONTEXT" wait -n its1-system job.batch/its-with-clusteradm \
        --for condition=Complete --timeout 400s; then
        log "ERROR" "Timeout waiting for clusteradm job"
        exit 1
    fi
    
    add_workload_cluster "cluster1"
    add_workload_cluster "cluster2"
    
    log "INFO" "Waiting for cluster registration..."
    if ! wait-for-cmd '(( $(kubectl --context its1 get csr 2>/dev/null | grep -c Pending) >= 2 ))'; then
        log "ERROR" "Timeout waiting for cluster registration CSRs"
        exit 1
    fi
    
    clusteradm --context its1 accept --clusters cluster1
    clusteradm --context its1 accept --clusters cluster2
    
    log "INFO" "Labeling and configuring clusters..."
    kubectl --context its1 label managedcluster cluster1 location-group=edge name=cluster1 region=east
    kubectl --context its1 create cm -n customization-properties cluster1 --from-literal clusterURL=https://my.clusters/1001-abcd
    kubectl --context its1 label managedcluster cluster2 location-group=edge name=cluster2 region=west
    kubectl --context its1 create cm -n customization-properties cluster2 --from-literal clusterURL=https://my.clusters/2002-cdef
    
    log "SUCCESS" "Cluster registration completed"
}

# Verify deployment
verify_deployment() {
    log "INFO" "Verifying deployments..."
    
    local expected_count=5
    if [ "$ARGOCD_INSTALL" = true ]; then
        expected_count=8  # Additional ArgoCD deployments
    fi
    
    if ! wait-for-cmd "(( \$(kubectl --context '$HOSTING_CONTEXT' get deployments,statefulsets --all-namespaces | grep -e wds1 -e its1 -e argocd | wc -l) >= $expected_count ))"; then
        log "ERROR" "Deployment verification failed"
        exit 1
    fi
    
    log "SUCCESS" "All expected deployments are running"
}

# Verify cluster registration
verify_cluster_registration() {
    log "INFO" "Verifying cluster registration..."
    
    if ! wait-for-cmd 'kubectl --context its1 get managedclusters -l location-group=edge --no-headers | wc -l | grep -wq 2'; then
        log "ERROR" "Cluster registration verification failed"
        exit 1
    fi
    
    log "SUCCESS" "Cluster registration verified"
}

# Main execution
main() {
    parse_args "$@"
    
    log "INFO" "Starting KubeStellar setup"
    log "INFO" "Environment: $CLUSTER_SOURCE"
    log "INFO" "Hosting context: $HOSTING_CONTEXT"
    log "INFO" "ArgoCD installation: $ARGOCD_INSTALL"
    log "INFO" "Use released version: $USE_RELEASE"
    
    SRC_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    COMMON_SRCS="${SRC_DIR}/../common"
    
    # Source common setup functions
    if [[ -f "$COMMON_SRCS/setup-shell.sh" ]]; then
        source "$COMMON_SRCS/setup-shell.sh"
    else
        log "ERROR" "Common setup files not found"
        exit 1
    fi
    
    check_prerequisites
    setup_hosting_cluster
    install_kubestellar
    wait_for_ocm_hub
    wait_for_transport_controller
    configure_contexts
    wait_for_argocd
    register_clusters
    verify_deployment
    verify_cluster_registration
    
    cat <<EOF

${GREEN}🚀 KubeStellar Setup Summary${NC}
-----------------------------------------
${GREEN}✅ All components installed successfully${NC}

${BLUE}Access Information:${NC}
  - Hosting Cluster: ${HOSTING_CONTEXT}
  - WDS Context: wds1
  - ITS Context: its1
  - Workload Clusters: cluster1, cluster2
  - Log File: ${LOG_FILE}

EOF
    
    if [ "$ARGOCD_INSTALL" = true ]; then
        cat <<EOF
${YELLOW}ArgoCD Integration:${NC}
  - UI: https://${ARGOCD_DOMAIN}$([ "$CLUSTER_SOURCE" = "kind" ] && echo ":9443")
  - Username: admin
  - Password: $(kubectl -n default get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d)

EOF
    fi
    
    log "SUCCESS" "KubeStellar setup completed successfully!"
}

main "$@"