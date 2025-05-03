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

# Purpose: build a kubectl image and push it to the registry.

# Usage: $0 [--version|-v release] [--registry registry] [--platform platforms] [-X]
# Working directory does not matter.

set -e

version="" # ==> latest
registry=quay.io/kubestellar
platform=linux/amd64,linux/arm64,linux/ppc64le

get_latest_version() {
    curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt
}

while (( $# > 0 )); do
    case "$1" in
    (--version|-v)
        if (( $# > 1 ));
        then { version="$2"; shift; }
        else { echo "$0: missing release version" >&2; exit 1; }
        fi;;
    (--registry)
        if (( $# > 1 ));
        then { registry="$2"; shift; }
        else { echo "$0: missing registry url" >&2; exit 1; }
        fi;;
    (--platform)
        if (( $# > 1 ));
        then { platform="$2"; shift; }
        else { echo "$0: missing comma separated list of platforms" >&2; exit 1; }
        fi;;
    (-X)
    	set -x;;
    (-h|--help)
        echo "Usage: $0 [--version|-v release] [--registry registry] [--platform platforms] [-X]"
        exit 0;;
    (-*)
        echo "$0: unknown flag" >&2
        exit 1;;
    (*)
        echo "$0: unknown positional argument" >&2
        exit 1;;
    esac
    shift
done

if [ "$version" == "" ]; then
    version=$(get_latest_version)
fi

# Remove the initial "v", if present
version=${version#v}

echo "Using kubectl v${version}."

docker buildx create --name kubestellar --platform ${platform} --use

docker buildx build . \
    --push \
    --platform ${platform} \
    --tag quay.io/kubestellar/kubectl:${version#v} \
    --build-arg version=${version#v} \
    --tag quay.io/kubestellar/kubectl:${version#v} \
    -f "$(dirname "$0")/Dockerfile.kubectl"

docker buildx rm kubestellar
