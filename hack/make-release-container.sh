#!/usr/bin/env bash

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

# Usage: $0 KCPE_VERSION

if [ $# != 1 ]; then
    echo "$0 usage: KCPE_VERSION target_os target_arch" >&2
    exit 1
fi

kcpe_version="$1"

set -e

srcdir=$(dirname "$0")
cd "$srcdir/.."
#docker login quay.io
GIT_COMMIT=$(git rev-parse --short HEAD || echo 'local')
GIT_WARN=$( [ $(git status --porcelain=v2 | wc -l) == 0 ] || echo "-dirty")
# KO_DOCKER_REPO=quay.io/kubestellar/syncer ko build --platform=linux/amd64,linux/arm64,... --bare --tags="${kcpe_version},git${GIT_COMMIT}${GIT_WARN}" ./cmd/syncer
cd "$srcdir"
