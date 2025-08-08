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

set -e # exit on error
set -x # for debugging

platform=${1:-kind}

# Validate platform
if [[ "$platform" != "kind" && "$platform" != "k3d" ]]; then
    echo "ERROR: Unsupported platform '$platform'. Use 'kind' or 'k3d'"
    exit 1
fi

SRC_DIR="$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)"
scripts_dir="${SRC_DIR}/../../../scripts"
bash_dir="${SRC_DIR}/../bash"
common_srcs="${SRC_DIR}/../common"

source "${common_srcs}/setup-shell.sh"

echo "Testing demo environment setup with platform: $platform"

# Test the demo environment creation script
echo "Creating demo environment with $platform..."
if ! "${scripts_dir}/create-kubestellar-demo-env.sh" --platform $platform; then
    echo "ERROR: Demo environment creation script failed for $platform!"
    exit 1
fi

echo "Demo environment created successfully with $platform"

# Run E2E test only for kind since use-kubestellar.sh only supports kind,ocp
if [ "$platform" == "kind" ]; then
    echo "Running E2E bash test for $platform..."
    
    cd "${bash_dir}"
    if ! ./use-kubestellar.sh --env $platform; then
        echo "ERROR: E2E bash test failed for $platform!"
        exit 1
    fi
    echo "SUCCESS: E2E bash test validation completed for $platform!"
else
    echo "Skipping E2E bash test for $platform (only supports kind/ocp)"
fi

echo "SUCCESS: demo environment tests completed successfully!"
