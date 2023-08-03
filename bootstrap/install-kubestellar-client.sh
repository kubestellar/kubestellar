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

# This script installs kubestellar/kcp client binaries to a folder of choice
#
# Arguments:
# [--kcp-version release] set a specific kcp release version, default: latest
# [--kubestellar-version release] set a specific KubeStellar release version, default: latest
# [--os linux|darwin] set a specific OS type, default: autodetect
# [--arch amd64|arm64] set a specific architecture type, default: autodetect
# [--ensure-folder name] sets the installation folder, default: $PWD/kubestellar
# [-V|--verbose] verbose output
# [-X] `set -x`

set -e

kcp_download() {
    if [ $verbose == "true" ]; then
        curl -SL -o kcp-plugins.tar.gz "https://github.com/kcp-dev/kcp/releases/download/${kcp_version}/kubectl-kcp-plugin_${kcp_version//v}_${os_type}_${arch_type}.tar.gz"
    else
        curl -sSL -o kcp-plugins.tar.gz "https://github.com/kcp-dev/kcp/releases/download/${kcp_version}/kubectl-kcp-plugin_${kcp_version//v}_${os_type}_${arch_type}.tar.gz"
    fi
}

kcp_install() {
    tar -C $kubestellar_client_folder -zxf kcp-plugins.tar.gz
}

kcp_version() {
    echo "$(kubectl version --short 2> /dev/null | grep kcp | sed 's/.*kcp-//')"
}

kcp_get_latest_version() {
    curl -sL https://github.com/kcp-dev/kcp/releases/latest | grep "</h1>" | tail -n 1 | sed -e 's/<[^>]*>//g' | xargs
}

kubeconfig_valid() {
    if [ "$KUBECONFIG" == "" ]; then
        echo "false"
    elif [ -f "$KUBECONFIG" ]; then
        echo "true"
    else
        echo "false"
    fi
}

kubestellar_download() {
    if [ $verbose == "true" ]; then
        curl -SL -o kubestellar.tar.gz "https://github.com/kubestellar/kubestellar/releases/download/${kubestellar_version}/kubestellar_${kubestellar_version}_${os_type}_$arch_type.tar.gz"
    else
        curl -sSL -o kubestellar.tar.gz "https://github.com/kubestellar/kubestellar/releases/download/${kubestellar_version}/kubestellar_${kubestellar_version}_${os_type}_$arch_type.tar.gz"
    fi
}

kubestellar_install() {
    tar -C $kubestellar_client_folder -zxf kubestellar.tar.gz --exclude="config" --exclude="README.md"
}

kubestellar_get_latest_version() {
    curl -sL https://github.com/kubestellar/kubestellar/releases/latest | grep "</h1>" | tail -n 1 | sed -e 's/<[^>]*>//g' | xargs
}

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
      arm64*)   echo "arm64" ;;
      *)        echo "Unsupported architecture type: $HOSTTYPE" >&2 ; exit 1 ;;
  esac
}

get_full_path() {
    # echo "check folder '$1'"
    echo "$(cd "$1"; pwd)"
}

ensure_folder() {
    # echo "ensure folder '$1'"
    if [ -d "$1" ]; then :
    else
        mkdir -p "$1"
    fi
}

KCP_REQUIRED_VERSION="v0.11.0"
kcp_version=""
kubestellar_version=""
os_type=""
arch_type=""
folder=""
verbose="false"
flagx=""
user_exports=""

echo "< KubeStellar bootstrap started >----------------"

while (( $# > 0 )); do
    case "$1" in
    (--kcp-version)
        if (( $# > 1 ));
        then { kcp_version="$2"; shift; }
        else { echo "$0: missing release version" >&2; exit 1; }
        fi;;
    (--kubestellar-version)
        if (( $# > 1 ));
        then { kubestellar_version="$2"; shift; }
        else { echo "$0: missing release version" >&2; exit 1; }
        fi;;
    (--os)
        if (( $# > 1 ));
        then { os_type="$2"; shift; }
        else { echo "$0: missing OS type" >&2; exit 1; }
        fi;;
    (--arch)
        if (( $# > 1 ));
        then { arch_type="$2"; shift; }
        else { echo "$0: missing architecture type" >&2; exit 1; }
        fi;;
    (--ensure-folder)
        if (( $# > 1 ));
        then { folder="$2"; shift; }
        else { echo "$0: missing installation folder" >&2; exit 1; }
        fi;;
    (--verbose|-V)
        verbose="true";;
    (-X)
	set -x
	flagx="-X";;
    (-h|--help)
        echo "Usage: $0 [--kcp-version release_version] [--kubestellar-version release_version] [--os linux|darwin] [--arch amd64|arm64] [--ensure-folder installation_folder] [-V|--verbose] [-X]"
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
    kcp_version=$KCP_REQUIRED_VERSION
fi

if [ "$kubestellar_version" == "" ] || [ "$kubestellar_version" == latest ]; then
    kubestellar_version=$(kubestellar_get_latest_version)
fi

if [ "$os_type" == "" ]; then
    os_type=$(get_os_type)
fi

if [ "$arch_type" == "" ]; then
    arch_type=$(get_arch_type)
fi

if [ "$folder" == "" ]; then
    folder="$PWD"
fi

# Install kcp plugins
echo "< Ensure kcp plugins are installed >----------------------"

ensure_folder "$folder"
kubestellar_client_folder=$(get_full_path "$folder")
echo "Downloading kcp+plugins $kcp_version $os_type/$arch_type..."
kcp_download
echo "Installing kcp+plugins into '$kubestellar_client_folder'..."
kcp_install
rm kcp-plugins.tar.gz

# Ensure KubeStellar is installed
echo "< Ensure KubeStellar is installed >-------------"

echo "Downloading KubeStellar $kubestellar_version $os_type/$arch_type..."
kubestellar_download
echo "Installing KubeStellar into '$kubestellar_client_folder'..."
kubestellar_install
rm kubestellar.tar.gz

if [[ ! ":$PATH:" == *":$kubestellar_client_folder/bin:"* ]]; then
    export PATH=$kubestellar_client_folder/bin:$PATH
    echo "Add kubestellar client folder to the PATH: export PATH=\"$kubestellar_client_folder/bin:\$PATH\""
    user_exports="$user_exports"$'\n'"export PATH=\"$kubestellar_client_folder/bin:\$PATH\""
fi

echo "< KubeStellar bootstrap completed successfully >-"
if [ "$user_exports" != "" ]; then
    echo "Please create/update the following environment variables: $user_exports"
fi
