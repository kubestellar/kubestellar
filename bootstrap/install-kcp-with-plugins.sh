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

# Usage: $0

# This script installs KCP binaries and plugins to a folder of choice
#
# Arguments:
# [--version release] set a specific KCP release version, default: latest
# [--os linux|darwin] set a specific OS type, default: autodetect
# [--arch amd64|arm64] set a specific architecture type, default: autodetect
# [--ensure-folder name] sets the installation folder, default: $PWD/KCP
# [--strip-bin] remove the bin sub-folder
# [-V|--verbose] verbose output
# [-X] `set -x`

set -e

kcp_version=""
kcp_os=""
kcp_arch=""
folder=""
strip_bin="false"
verbose="false"

get_os_type() {
  case "$OSTYPE" in
      linux*)   echo "linux" ;;
      darwin*)  echo "darwin" ;;
      *)        echo "Unsupported operating system type: $OSTYPE" >&2 ; exit 1 ;;
  esac
}

get_arch_type() {
  case "$HOSTTYPE" in
      x86_64*)  echo "amd64" ;;
      aarch64*) echo "arm64" ;;
      *)        echo "Unsupported architecture type: $HOSTTYPE" >&2 ; exit 1 ;;
  esac
}

get_latest_version() {
    curl -sL https://github.com/kcp-dev/kcp/releases/latest | grep "</h1>" | head -n 1 | sed -e 's/<[^>]*>//g' | xargs
}

get_full_path() {
    echo "$(cd "$1"; pwd)"
}

while (( $# > 0 )); do
    case "$1" in
    (--version)
        if (( $# > 1 ));
        then { kcp_version="$2"; shift; }
        else { echo "$0: missing release version" >&2; exit 1; }
        fi;;
    (--os)
        if (( $# > 1 ));
        then { kcp_os="$2"; shift; }
        else { echo "$0: missing OS type" >&2; exit 1; }
        fi;;
    (--arch)
        if (( $# > 1 ));
        then { kcp_arch="$2"; shift; }
        else { echo "$0: missing architecture type" >&2; exit 1; }
        fi;;
    (--ensure-folder)
        if (( $# > 1 ));
        then { folder="$2"; shift; }
        else { echo "$0: missing installation folder" >&2; exit 1; }
        fi;;
    (--strip-bin)
        strip_bin="true";;
    (--verbose|-V)
        verbose="true";;
    (-X)
	set -x;;
    (-h|--help)
        echo "Usage: $0 [--version release] [--os linux|darwin] [--arch amd64|arm64] [--ensure-folder name] [--strip-bin] [-V|--verbose] [-X]"
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

if [ "$kcp_version" == "" ]; then
    kcp_version=$(get_latest_version)
fi

if [ "$kcp_os" == "" ]; then
    kcp_os=$(get_os_type)
fi

if [ "$kcp_arch" == "" ]; then
    kcp_arch=$(get_arch_type)
fi

if [ "$folder" == "" ]; then
    folder="$PWD/kcp"
fi

if [ -d "$folder" ]; then :
else
    if [ $verbose == "true" ]; then
        echo "Creating folder: $folder"
    fi
    mkdir -p "$folder"
fi

if [ $verbose == "true" ]; then
    echo "Downloading KCP $kcp_version $kcp_os/$kcp_arch..."
    curl -SL -o kcp.tar.gz "https://github.com/kcp-dev/kcp/releases/download/${kcp_version}/kcp_${kcp_version//v}_${kcp_os}_${kcp_arch}.tar.gz"
    echo "Downloading KCP plugins $kcp_version $kcp_os/$kcp_arch..."
    curl -SL -o kcp-plugins.tar.gz "https://github.com/kcp-dev/kcp/releases/download/${kcp_version}/kubectl-kcp-plugin_${kcp_version//v}_${kcp_os}_${kcp_arch}.tar.gz"
else
    curl -sSL -o kcp.tar.gz "https://github.com/kcp-dev/kcp/releases/download/${kcp_version}/kcp_${kcp_version//v}_${kcp_os}_${kcp_arch}.tar.gz"
    curl -sSL -o kcp-plugins.tar.gz "https://github.com/kcp-dev/kcp/releases/download/${kcp_version}/kubectl-kcp-plugin_${kcp_version//v}_${kcp_os}_${kcp_arch}.tar.gz"
fi

if [ $verbose == "true" ]; then
    echo "Extracting archive to: $folder"
fi

if [ $strip_bin == "true" ]; then
    tar -C $folder -zxf kcp-plugins.tar.gz --wildcards --strip-components=1 bin/*
    tar -C $folder -zxf kcp.tar.gz --wildcards --strip-components=1 bin/*
    bin_folder=$(get_full_path "$folder")
else
    tar -C $folder -zxf kcp-plugins.tar.gz
    tar -C $folder -zxf kcp.tar.gz
    bin_folder=$(get_full_path "$folder/bin")
fi

if [ $verbose == "true" ]; then
    echo "Removing downloaded archives..."
fi

rm kcp.tar.gz
rm kcp-plugins.tar.gz

if [[ ! ":$PATH:" == *":$bin_folder:"* ]]; then
    echo "Add KCP folder to the path: export PATH="\$PATH:$bin_folder""
fi
