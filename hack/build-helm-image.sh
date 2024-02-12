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

# Purpose: build a helm image and push it to the registry.

# Usage: $0 [--version|-v release] [--registry registry] [--platform platforms] [-X]
# Working directory does not matter.

set -e

helm_version="" # ==> latest
helm_folder=helm
registry=quay.io/kubestellar
platform=linux/amd64,linux/arm64,linux/ppc64le

get_latest_version() {
    curl -sL https://github.com/helm/helm/releases/latest | grep "</h1>" | tail -n 1 | sed -e 's/<[^>]*>//g' | xargs | awk  '{print $2}'
}

while (( $# > 0 )); do
    case "$1" in
    (--version|-v)
        if (( $# > 1 ));
        then { helm_version="$2"; shift; }
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

if [ "$helm_version" == "" ]; then
    helm_version=$(get_latest_version)
fi

# Remove the initial "v", if present
helm_version=${helm_version#v}

echo "Using helm v${helm_version}."

git clone -b "v$helm_version" --depth 1 https://github.com/helm/helm.git "$helm_folder"

cd "$helm_folder"

export KO_DOCKER_REPO=$registry

ko build -B ./cmd/helm -t $helm_version --sbom=none --platform $platform

cd ~
rm -rf $helm_folder
