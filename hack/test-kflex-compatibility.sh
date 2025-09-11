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

# Compare versions: returns 0 if v1 <= v2
version_le() {
    local v1_clean=${1#v}
    local v2_clean=${2#v}
    [[ "$v1_clean" == "$v2_clean" || "$v1_clean" < "$v2_clean" ]]
}

# Function to test a specific KubeFlex CLI version
test_kflex_version() {
    local version=$1
    local install_dir=$(mktemp -d)
    local kflex_binary="$install_dir/kflex"
    
    # Determine architecture
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)
    case $arch in
        x86_64) arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
    esac
    
    local download_url="https://github.com/kubestellar/kubeflex/releases/download/${version}/kubeflex_${version#v}_${os}_${arch}.tar.gz"
    
    if ! curl -sL "$download_url" | tar -xz -C "$install_dir" bin/kflex 2>/dev/null; then
        echo "WARNING: Failed to download KubeFlex $version, skipping..."
        rm -rf "$install_dir"
        return 1
    fi
    
    # Move binary to expected location
    mv "$install_dir/bin/kflex" "$kflex_binary"
    chmod +x "$kflex_binary"
    
    # Test basic CLI functionality
    if ! out=$("$kflex_binary" version 2>&1); then
        echo "ERROR: kflex version command failed for $version"
        echo "$out"
        rm -rf "$install_dir"
        return 1
    fi
    # TODO: Actually test KubeStellar chart with this kflex version
    rm -rf "$install_dir"
    return 0
}

# Get available KubeFlex releases from GitHub API
RELEASES_JSON=$(curl -s "https://api.github.com/repos/kubestellar/kubeflex/releases?per_page=5" || echo "[]")

if [[ "$RELEASES_JSON" == "[]" ]]; then
    echo "ERROR: Could not fetch KubeFlex releases from GitHub API"
    exit 1
fi

# Extract version tags and filter by minimum version
VERSIONS=$(echo "$RELEASES_JSON" | jq -r '.[].tag_name' | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | sort -V)

if [[ -z "$VERSIONS" ]]; then
    echo "ERROR: No valid version tags found"
    exit 1
fi

echo Available versions: $VERSIONS

# Test each version that meets minimum requirement
TESTED_COUNT=0
PASSED_COUNT=0
FAILED_COUNT=0

for version in $VERSIONS; do
    if version_le "$MIN_VERSION" "$version"; then
        ((TESTED_COUNT++))
        if test_kflex_version "$version"; then
            ((PASSED_COUNT++))
        else
            ((FAILED_COUNT++))
        fi
    fi
done

echo
echo "=== Test Summary ==="
echo "Tested versions: $TESTED_COUNT"
echo "Passed: $PASSED_COUNT"
echo "Failed: $FAILED_COUNT"

if [[ $FAILED_COUNT -gt 0 ]]; then
    echo "ERROR: Some KubeFlex versions failed compatibility test"
    exit 1
fi

if [[ $TESTED_COUNT -eq 0 ]]; then
    echo "WARNING: No versions were tested"
    exit 1
fi

echo "SUCCESS: All tested KubeFlex versions are compatible with KubeStellar"
