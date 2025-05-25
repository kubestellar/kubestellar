#!/bin/bash

# Copyright 2023 The KubeStellar Authors.
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

# Purpose: update several docs that reference a release number explicitly 

# Usage: $0 <version number in the format of X.Y.Z>
# Should run from the kubestellar root directory


# Input version should be X.Y.Z
input_version="$1"

if ! [[ "$1" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-rc[0-9]+)?$ ]]; then
    echo "Error: Invalid version format. Please provide a version in the format X.Y.Z[-RCn] (X, Y, Z, n are all numbers). " >&2
    exit 1
fi

# Replace version string using sed

sed -i "/- --version/{n;s/- \"[0-9]\+\.[0-9]\+\.[0-9]\+\"/- \"$input_version\"/}" config/postcreate-hooks/kubestellar.yaml

sed -i "s/export KUBESTELLAR_VERSION=[0-9]\+\.[0-9]\+\.[0-9]\+/export KUBESTELLAR_VERSION=$input_version/" docs/content/direct/examples.md

sed -i "/The latest release is/s/[0-9]\+\.[0-9]\+\.[0-9]\+/$input_version/g" docs/content/direct/README.md
