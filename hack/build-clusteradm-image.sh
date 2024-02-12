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

# Purpose: build a clusteradm image and push it to the registry.

# Usage: $0 [--version|-v release] [--registry registry] [--platform platforms] [-X]
# Working directory does not matter.

set -e

clusteradm_version="" # ==> latest
clusteradm_folder=clusteradm
registry=quay.io/kubestellar
platform=linux/amd64,linux/arm64,linux/ppc64le

get_latest_version() {
    curl -sL https://github.com/open-cluster-management-io/clusteradm/releases/latest | grep "</h1>" | tail -n 1 | sed -e 's/<[^>]*>//g' | xargs
}

while (( $# > 0 )); do
    case "$1" in
    (--version|-v)
        if (( $# > 1 ));
        then { clusteradm_version="$2"; shift; }
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

if [ "$clusteradm_version" == "" ]; then
    clusteradm_version=$(get_latest_version)
fi

# Remove the initial "v", if present
clusteradm_version=${clusteradm_version#v}

echo "Using clusteradm v${clusteradm_version}."

git clone -b "v$clusteradm_version" --depth 1 https://github.com/open-cluster-management-io/clusteradm.git "$clusteradm_folder"

cd "$clusteradm_folder"

export KO_DOCKER_REPO=$registry

ko build -B ./cmd/clusteradm -t $clusteradm_version --sbom=none --platform $platform

cd ~
rm -rf $clusteradm_folder
