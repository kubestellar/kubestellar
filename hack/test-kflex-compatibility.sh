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

# KubeFlex CLI compatibility test for KubeStellar
# Tests compatibility between KubeStellar core chart and various KubeFlex CLI versions

set -euo pipefail

# Default configuration
KUBEFLEX_FETCH_CMD="${KUBEFLEX_FETCH_CMD:-kflex ctx kubestellar-system --kubeconfig -}"
DRY_RUN=false
MAX_VERSIONS=5

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --max-versions)
            MAX_VERSIONS="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [--dry-run] [--max-versions N]"
            echo ""
            echo "Options:"
            echo "  --dry-run         Show what would be tested without actually running tests"
            echo "  --max-versions N  Maximum number of versions to test (default: 5)"
            echo ""
            echo "Environment variables:"
            echo "  KUBEFLEX_FETCH_CMD  Command to fetch kubeconfig (default: kflex ctx kubestellar-system --kubeconfig -)"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Source directory for KubeStellar scripts
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
KUBESTELLAR_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Extract minimum KubeFlex version from pre_req.sh
MIN_VERSION_LINE=$(grep -E "Kubeflex version: v[0-9]+\.[0-9]+\.[0-9]+" "${KUBESTELLAR_DIR}/scripts/check_pre_req.sh" || true)
if [[ -z "$MIN_VERSION_LINE" ]]; then
    echo "ERROR: Could not find minimum KubeFlex version in check_pre_req.sh"
    exit 1
fi

MIN_VERSION=$(echo "$MIN_VERSION_LINE" | grep -oE "v[0-9]+\.[0-9]+\.[0-9]+" | head -1)

# Semantic version comparison using sort -V
version_le() {
    local v1_clean=${1#v}
    local v2_clean=${2#v}
    [[ "$v1_clean" < "$v2_clean" || "$v1_clean" == "$v2_clean" ]]
}

# Check prerequisites once before testing any versions
check_prerequisites() {
    local missing_tools=()
    
    if ! command -v kubectl &> /dev/null; then
        missing_tools+=("kubectl")
    fi
    
    if ! command -v helm &> /dev/null; then
        missing_tools+=("helm")
    fi
    
    if ! command -v kind &> /dev/null; then
        missing_tools+=("kind")
    fi
    
    if [[ ${#missing_tools[@]} -gt 0 ]]; then
        echo "ERROR: Required tools not found: ${missing_tools[*]}"
        echo "Please install these tools before running the test."
        return 1
    fi
    
    return 0
}

# Set up KubeStellar core once for all tests
setup_kubestellar_core() {
    local cluster_name="ks-compat-test-$$"
    echo "Setting up KubeStellar core for compatibility testing..."
    
    # Create kind cluster
    if ! kind create cluster --name "$cluster_name" --wait 5m; then
        echo "ERROR: Failed to create kind cluster"
        return 1
    fi
    
    # Export kubeconfig for use by kflex
    export KUBECONFIG="$(kind get kubeconfig --name "$cluster_name")"
    
    # Install KubeStellar core chart
    if ! helm repo add ks-core "https://kubestellar.github.io/kubestellar/core-chart" &>/dev/null; then
        echo "ERROR: Failed to add KubeStellar helm repository"
        kind delete cluster --name "$cluster_name" || true
        return 1
    fi
    
    if ! helm repo update ks-core &>/dev/null; then
        echo "ERROR: Failed to update helm repository"
        kind delete cluster --name "$cluster_name" || true
        return 1
    fi
    
    if ! helm install ks-core ks-core/kubestellar-core --create-namespace --namespace kubestellar-system --wait --timeout=5m &>/dev/null; then
        echo "ERROR: Failed to install KubeStellar core"
        kind delete cluster --name "$cluster_name" || true
        return 1
    fi
    
    # Wait for core components to be ready
    if ! kubectl wait --for=condition=available --timeout=300s deployment -n kubestellar-system --all &>/dev/null; then
        echo "ERROR: KubeStellar core components did not become ready"
        kind delete cluster --name "$cluster_name" || true
        return 1
    fi
    
    echo "KubeStellar core setup complete"
    echo "$cluster_name"  # Return cluster name for cleanup
}

# Clean up KubeStellar core
cleanup_kubestellar_core() {
    local cluster_name="$1"
    if [[ -n "$cluster_name" ]]; then
        kind delete cluster --name "$cluster_name" &>/dev/null || true
    fi
}

# Function to test a specific KubeFlex CLI version
test_kflex_version() {
    local version=$1
    local install_dir=$(mktemp -d)
    local kflex_binary="$install_dir/kflex"
    
    echo "Testing KubeFlex $version..."
    
    # Determine architecture
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)
    case $arch in
        x86_64) arch="amd64" ;;
        aarch64) arch="arm64" ;;
    esac
    
    local download_url="https://github.com/kubestellar/kubeflex/releases/download/${version}/kubeflex_${version#v}_${os}_${arch}.tar.gz"
    
    # Download and extract kflex binary directly to final location
    if ! curl -sL "$download_url" | tar -xz -C "$install_dir" --strip-components=1 bin/kflex 2>/dev/null; then
        echo "  WARNING: Failed to download KubeFlex $version, skipping..."
        rm -rf "$install_dir"
        return 1
    fi
    
    chmod +x "$kflex_binary"
    
    # Test basic CLI functionality
    local version_output
    if ! version_output=$("$kflex_binary" version 2>&1); then
        echo "  ERROR: kflex version command failed for $version"
        echo "  Output: $version_output"
        rm -rf "$install_dir"
        return 1
    fi
    
    # Test KubeStellar integration with this kflex version
    local integration_stderr
    if ! integration_stderr=$(test_kubestellar_integration "$kflex_binary" 2>&1); then
        echo "  ERROR: KubeStellar integration test failed for kflex $version"
        echo "  Details: $integration_stderr"
        rm -rf "$install_dir"
        return 1
    fi
    
    echo "  SUCCESS: KubeFlex $version is compatible"
    rm -rf "$install_dir"
    return 0
}

# Test actual KubeStellar integration with kflex
test_kubestellar_integration() {
    local kflex_binary=$1
    
    # Test 1: Verify kubestellar-system namespace exists and is accessible
    if ! kubectl get namespace kubestellar-system &>/dev/null; then
        echo "kubestellar-system namespace not found"
        return 1
    fi
    
    # Test 2: Verify core pods are running
    local not_running
    if not_running=$(kubectl get pods -n kubestellar-system --field-selector=status.phase!=Running --no-headers 2>/dev/null); then
        if [[ -n "$not_running" ]]; then
            echo "Some KubeStellar pods are not running: $not_running"
            return 1
        fi
    fi
    
    # Test 3: Test kflex context management with KubeStellar
    local kubeconfig_output
    if ! kubeconfig_output=$(eval "$KUBEFLEX_FETCH_CMD" 2>&1); then
        echo "Failed to fetch kubeconfig using: $KUBEFLEX_FETCH_CMD"
        echo "Error: $kubeconfig_output"
        return 1
    fi
    
    # Test 4: Verify the fetched kubeconfig is valid
    if ! echo "$kubeconfig_output" | kubectl --kubeconfig=/dev/stdin cluster-info &>/dev/null; then
        echo "Fetched kubeconfig is not valid"
        return 1
    fi
    
    # Test 5: Create and delete a test resource to verify write access
    local test_pod_name="ks-compat-test-$$-$(date +%s)"
    local test_manifest="apiVersion: v1
kind: Pod
metadata:
  name: $test_pod_name
  namespace: kubestellar-system
spec:
  containers:
  - name: test
    image: busybox:1.35
    command: ['sleep', '1']
  restartPolicy: Never"
    
    if ! echo "$test_manifest" | kubectl apply -f - &>/dev/null; then
        echo "Failed to create test pod"
        return 1
    fi
    
    # Wait briefly and clean up test pod
    sleep 2
    kubectl delete pod "$test_pod_name" -n kubestellar-system --ignore-not-found &>/dev/null
    
    return 0
}

# Main execution
main() {
    if [[ "$DRY_RUN" == "true" ]]; then
        echo "DRY RUN MODE - showing what would be tested"
        echo "KUBEFLEX_FETCH_CMD: $KUBEFLEX_FETCH_CMD"
        echo "MAX_VERSIONS: $MAX_VERSIONS"
        echo "MIN_VERSION: $MIN_VERSION"
        echo ""
    fi
    
    # Check prerequisites
    if ! check_prerequisites; then
        exit 1
    fi
    
    # Get available KubeFlex releases from GitHub API
    local releases_json
    if ! releases_json=$(curl -s "https://api.github.com/repos/kubestellar/kubeflex/releases?per_page=20" 2>/dev/null); then
        echo "ERROR: Could not fetch KubeFlex releases from GitHub API"
        exit 1
    fi
    
    if [[ "$releases_json" == "[]" ]]; then
        echo "ERROR: No releases found"
        exit 1
    fi
    
    # Extract version tags and filter by minimum version
    local versions
    if ! versions=$(echo "$releases_json" | jq -r '.[].tag_name' | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | sort -V | tail -n "$MAX_VERSIONS"); then
        echo "ERROR: No valid version tags found"
        exit 1
    fi
    
    if [[ -z "$versions" ]]; then
        echo "ERROR: No valid versions found after filtering"
        exit 1
    fi
    
    # Filter versions that meet minimum requirement
    local eligible_versions=()
    while IFS= read -r version; do
        if version_le "$MIN_VERSION" "$version"; then
            eligible_versions+=("$version")
        fi
    done <<< "$versions"
    
    if [[ ${#eligible_versions[@]} -eq 0 ]]; then
        echo "ERROR: No versions meet minimum requirement $MIN_VERSION"
        exit 1
    fi
    
    echo "Testing KubeFlex versions: ${eligible_versions[*]}"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        echo ""
        echo "Would test ${#eligible_versions[@]} versions:"
        printf '  %s\n' "${eligible_versions[@]}"
        exit 0
    fi
    
    # Set up KubeStellar core once for all tests
    local cluster_name
    if ! cluster_name=$(setup_kubestellar_core); then
        echo "ERROR: Failed to set up KubeStellar core"
        exit 1
    fi
    
    # Ensure cleanup happens on exit
    trap 'cleanup_kubestellar_core "$cluster_name"' EXIT
    
    # Test each eligible version
    local tested_count=0
    local passed_count=0
    local failed_count=0
    
    for version in "${eligible_versions[@]}"; do
        ((tested_count++))
        if test_kflex_version "$version"; then
            ((passed_count++))
        else
            ((failed_count++))
        fi
    done
    
    echo ""
    echo "=== Test Summary ==="
    echo "Tested versions: $tested_count"
    echo "Passed: $passed_count"
    echo "Failed: $failed_count"
    
    if [[ $failed_count -gt 0 ]]; then
        echo "ERROR: Some KubeFlex versions failed compatibility test"
        exit 1
    fi
    
    if [[ $tested_count -eq 0 ]]; then
        echo "WARNING: No versions were tested"
        exit 1
    fi
    
    echo "SUCCESS: All tested KubeFlex versions are compatible with KubeStellar"
}

# Only run main if script is executed directly (not sourced)
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
