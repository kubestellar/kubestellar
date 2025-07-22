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

# Enhanced End-to-End Test for ArgoCD Integration with KubeStellar
set -o errexit          # Exit immediately if any command fails
set -o nounset          # Exit if any variable is unset
set -o pipefail         # Fail pipeline if any command in pipeline fails
set -o errtrace         # Enable error tracing
shopt -s inherit_errexit  # Ensure subshells inherit error handling

# Configuration defaults
DEFAULT_ENV="kind"
DEFAULT_ARGOCD_DOMAIN="argocd.localtest.me"
DEFAULT_TEST_TIMEOUT=300  # 5 minutes
DEFAULT_TEST_APP_REPO="https://github.com/argoproj/argocd-example-apps.git"
DEFAULT_TEST_APP_PATH="guestbook"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Global variables
declare -g HOSTING_CONTEXT=""
declare -g ARGOCD_NS=""
declare -g ARGOCD_POD=""
declare -g ARGOCD_PASSWORD=""
declare -g TEST_APP_NAME=""
declare -g CLEANUP_ON_EXIT=false
declare -g TEST_START_TIME=0

# Initialize logging
LOG_FILE="/tmp/kubestellar-argocd-test-$(date +%Y%m%d%H%M%S).log"
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

End-to-end test for ArgoCD integration with KubeStellar

Options:
  --env [kind|ocp]            Environment type (default: ${DEFAULT_ENV})
  --argocd-domain DOMAIN      ArgoCD domain (default: ${DEFAULT_ARGOCD_DOMAIN})
  --test-timeout SECONDS      Maximum test duration (default: ${DEFAULT_TEST_TIMEOUT})
  --app-repo URL              Test application repository (default: ${DEFAULT_TEST_APP_REPO})
  --app-path PATH             Path in repository for test app (default: ${DEFAULT_TEST_APP_PATH})
  --cleanup                   Clean up resources on exit (default: false)
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
                ENV="$2"
                shift 2
                ;;
            --argocd-domain)
                ARGOCD_DOMAIN="$2"
                shift 2
                ;;
            --test-timeout)
                [[ "$2" =~ ^[0-9]+$ ]] || {
                    log "ERROR" "Invalid timeout value: $2"
                    usage
                    exit 1
                }
                TEST_TIMEOUT="$2"
                shift 2
                ;;
            --app-repo)
                TEST_APP_REPO="$2"
                shift 2
                ;;
            --app-path)
                TEST_APP_PATH="$2"
                shift 2
                ;;
            --cleanup)
                CLEANUP_ON_EXIT=true
                shift
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

# Check prerequisites with version validation
check_prerequisites() {
    log "INFO" "Checking system prerequisites..."
    
    local required_tools=("kind" "kubectl" "helm" "kubeflex")
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
    
    # Verify minimum versions
    local kubectl_version
    kubectl_version=$(kubectl version --client -o json | jq -r '.clientVersion.gitVersion')
    if [[ $(echo "$kubectl_version" | cut -d. -f2) -lt 26 ]]; then
        log "WARNING" "kubectl version $kubectl_version may not be fully compatible (recommended v1.26+)"
    fi
    
    log "SUCCESS" "All prerequisites are installed"
}

# Setup test environment
setup_environment() {
    log "INFO" "Setting up test environment ($ENV)..."
    
    SRC_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    COMMON_SRCS="${SRC_DIR}/../common"
    
    # Source common setup functions
    if [[ -f "$COMMON_SRCS/setup-shell.sh" ]]; then
        source "$COMMON_SRCS/setup-shell.sh"
    else
        log "ERROR" "Common setup files not found"
        exit 1
    fi
    
    # Setup KubeStellar with ArgoCD
    if ! source "$COMMON_SRCS/setup-kubestellar.sh" --env "$ENV" --argocd --argocd-domain "$ARGOCD_DOMAIN"; then
        log "ERROR" "Failed to setup KubeStellar"
        exit 1
    fi
    
    # Set hosting context based on environment
    case "$ENV" in
        kind) HOSTING_CONTEXT="kind-kubeflex" ;;
        ocp)  HOSTING_CONTEXT="kscore" ;;
    esac
    
    log "SUCCESS" "Environment setup completed"
}

# Cleanup resources
cleanup() {
    if [[ "$CLEANUP_ON_EXIT" == true ]]; then
        log "INFO" "Cleaning up test resources..."
        
        if [[ -n "$TEST_APP_NAME" && -n "$ARGOCD_POD" && -n "$ARGOCD_NS" && -n "$HOSTING_CONTEXT" ]]; then
            if kubectl --context "$HOSTING_CONTEXT" -n "$ARGOCD_NS" exec "$ARGOCD_POD" -- \
                argocd app get "$TEST_APP_NAME" &>/dev/null; then
                log "INFO" "Deleting test application: $TEST_APP_NAME"
                kubectl --context "$HOSTING_CONTEXT" -n "$ARGOCD_NS" exec "$ARGOCD_POD" -- \
                    argocd app delete "$TEST_APP_NAME" --yes || true
            fi
        fi
        
        log "INFO" "Cleanup complete"
    fi
    
    local test_duration=$(( $(date +%s) - TEST_START_TIME ))
    log "INFO" "Test completed in ${test_duration} seconds"
    log "INFO" "Detailed logs available at: $LOG_FILE"
}

# Wait for command with timeout
wait_for() {
    local cmd="$1"
    local timeout=${2:-$DEFAULT_TEST_TIMEOUT}
    local interval=5
    local elapsed=0
    
    log "INFO" "Waiting for: $cmd"
    
    while [[ $elapsed -lt $timeout ]]; do
        if eval "$cmd"; then
            return 0
        fi
        sleep $interval
        elapsed=$((elapsed + interval))
        log "INFO" "Waiting... (${elapsed}s elapsed)"
    done
    
    log "ERROR" "Timeout after ${elapsed} seconds waiting for: $cmd"
    return 1
}

# Get ArgoCD credentials
get_argocd_credentials() {
    log "INFO" "Retrieving ArgoCD credentials..."
    
    ARGOCD_NS=$(kubectl --context "$HOSTING_CONTEXT" get pods -A -l app.kubernetes.io/name=argocd-server -o jsonpath='{.items[0].metadata.namespace}')
    ARGOCD_POD=$(kubectl --context "$HOSTING_CONTEXT" get pods -A -l app.kubernetes.io/name=argocd-server -o jsonpath='{.items[0].metadata.name}')
    
    if [[ -z "$ARGOCD_NS" || -z "$ARGOCD_POD" ]]; then
        log "ERROR" "Failed to locate ArgoCD components"
        exit 1
    fi
    
    log "INFO" "Found ArgoCD in namespace: $ARGOCD_NS"
    
    # Wait for ArgoCD to be ready
    if ! wait_for "kubectl --context $HOSTING_CONTEXT -n $ARGOCD_NS get pod $ARGOCD_POD -o jsonpath='{.status.phase}' | grep -q Running"; then
        log "ERROR" "ArgoCD pod did not reach Running state"
        exit 1
    fi
    
    # Get admin password
    ARGOCD_PASSWORD=$(kubectl --context "$HOSTING_CONTEXT" -n "$ARGOCD_NS" get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 -d)
    if [[ -z "$ARGOCD_PASSWORD" ]]; then
        log "ERROR" "Failed to retrieve ArgoCD admin password"
        exit 1
    fi
    
    log "SUCCESS" "ArgoCD credentials retrieved"
}

# Test ArgoCD server components
test_argocd_components() {
    log "INFO" "Testing ArgoCD components..."
    
    local components=("argocd-server" "argocd-application-controller" "argocd-repo-server")
    local all_healthy=true
    
    for component in "${components[@]}"; do
        if ! wait_for "kubectl --context $HOSTING_CONTEXT -n $ARGOCD_NS get pods -l app.kubernetes.io/name=$component -o jsonpath='{.items[*].status.phase}' | grep -q Running"; then
            log "ERROR" "$component is not running"
            all_healthy=false
        fi
    done
    
    if [[ "$all_healthy" == false ]]; then
        log "ERROR" "Not all ArgoCD components are healthy"
        exit 1
    fi
    
    log "SUCCESS" "All ArgoCD components are running"
}

# Test ArgoCD CLI authentication
test_argocd_cli_auth() {
    log "INFO" "Testing ArgoCD CLI authentication..."
    
    if ! kubectl --context "$HOSTING_CONTEXT" -n "$ARGOCD_NS" exec "$ARGOCD_POD" -- \
        argocd login ks-core-argocd-server."$ARGOCD_NS".svc.cluster.local --username admin --password "$ARGOCD_PASSWORD" --insecure; then
        log "ERROR" "ArgoCD CLI login failed"
        exit 1
    fi
    
    log "SUCCESS" "ArgoCD CLI authentication successful"
}

# Test cluster connectivity
test_cluster_connectivity() {
    log "INFO" "Testing cluster connectivity..."
    
    local clusters
    clusters=$(kubectl --context "$HOSTING_CONTEXT" -n "$ARGOCD_NS" exec "$ARGOCD_POD" -- argocd cluster list)
    
    if [[ -z "$clusters" ]]; then
        log "ERROR" "No clusters found in ArgoCD"
        exit 1
    fi
    
    echo "$clusters"
    
    if ! grep -q "wds1" <<< "$clusters"; then
        log "WARNING" "WDS1 cluster not found in ArgoCD cluster list"
    else
        log "SUCCESS" "ArgoCD can see WDS1 cluster"
    fi
}

# Test application lifecycle
test_application_lifecycle() {
    log "INFO" "Testing application lifecycle..."
    
    TEST_APP_NAME="kubestellar-test-app-$(date +%s)"
    
    log "INFO" "Creating test application: $TEST_APP_NAME"
    
    if ! kubectl --context "$HOSTING_CONTEXT" -n "$ARGOCD_NS" exec "$ARGOCD_POD" -- \
        argocd app create "$TEST_APP_NAME" \
        --repo "$TEST_APP_REPO" \
        --path "$TEST_APP_PATH" \
        --dest-server https://kubernetes.default.svc \
        --dest-namespace default \
        --sync-policy automated; then
        log "ERROR" "Failed to create ArgoCD application"
        exit 1
    fi
    
    # Wait for application to be created
    if ! wait_for "kubectl --context $HOSTING_CONTEXT -n $ARGOCD_NS exec $ARGOCD_POD -- argocd app get $TEST_APP_NAME &>/dev/null"; then
        log "ERROR" "Application creation verification failed"
        exit 1
    fi
    
    log "INFO" "Application created successfully. Current status:"
    kubectl --context "$HOSTING_CONTEXT" -n "$ARGOCD_NS" exec "$ARGOCD_POD" -- argocd app get "$TEST_APP_NAME"
    
    # Sync the application
    log "INFO" "Syncing application..."
    if ! kubectl --context "$HOSTING_CONTEXT" -n "$ARGOCD_NS" exec "$ARGOCD_POD" -- argocd app sync "$TEST_APP_NAME"; then
        log "ERROR" "Application sync failed"
        exit 1
    fi
    
    # Wait for sync to complete
    if ! wait_for "kubectl --context $HOSTING_CONTEXT -n $ARGOCD_NS exec $ARGOCD_POD -- argocd app get $TEST_APP_NAME --output json | jq -e '.status.sync.status == \"Synced\"' &>/dev/null"; then
        log "ERROR" "Application did not reach Synced state"
        exit 1
    fi
    
    log "SUCCESS" "Application lifecycle test completed"
}

# Main test execution
main() {
    TEST_START_TIME=$(date +%s)
    trap cleanup EXIT
    
    # Set defaults
    local ENV="$DEFAULT_ENV"
    local ARGOCD_DOMAIN="$DEFAULT_ARGOCD_DOMAIN"
    local TEST_TIMEOUT="$DEFAULT_TEST_TIMEOUT"
    local TEST_APP_REPO="$DEFAULT_TEST_APP_REPO"
    local TEST_APP_PATH="$DEFAULT_TEST_APP_PATH"
    
    parse_args "$@"
    
    log "INFO" "Starting ArgoCD-KubeStellar Integration Test"
    log "INFO" "Environment: $ENV"
    log "INFO" "ArgoCD Domain: $ARGOCD_DOMAIN"
    log "INFO" "Test Timeout: $TEST_TIMEOUT seconds"
    
    check_prerequisites
    setup_environment
    get_argocd_credentials
    test_argocd_components
    test_argocd_cli_auth
    test_cluster_connectivity
    test_application_lifecycle
    
    log "SUCCESS" "All ArgoCD integration tests passed successfully!"
    
    # Display access information
    cat <<EOF

${GREEN}🚀 ArgoCD Integration Test Summary${NC}
-----------------------------------------
${GREEN}✅ All tests completed successfully${NC}

${BLUE}Access Information:${NC}
  - Environment: ${ENV}
  - ArgoCD Namespace: ${ARGOCD_NS}
  - Admin Username: admin
  - Admin Password: ${ARGOCD_PASSWORD}

${YELLOW}Next Steps:${NC}
  1. Access the ArgoCD UI using the credentials above
  2. Explore the test application: ${TEST_APP_NAME}
  3. Check cluster connectivity in ArgoCD
  4. Review detailed logs at: ${LOG_FILE}

EOF
}

main "$@"