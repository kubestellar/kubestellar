#!/usr/bin/env bash

# Copyright 2023 The KCP Authors.
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

# This script installs kcp and Kubestellar binaries to a folder of choice
#
# Arguments:
# [--kcp-version release] set a specific kcp release version, default: latest
# [--kubestellar-version release] set a specific KubeStellar release version, default: latest
# [--os linux|darwin] set a specific OS type, default: autodetect
# [--arch amd64|arm64] set a specific architecture type, default: autodetect
# [--ensure-folder name] sets the installation folder, default: $PWD/kcp-edge
# [--bind-address address] bind kcp to a specific ip address
# [--ensure-imw name] create a Inventory Management Workspace (IMW)
# [--ensure-wmw name] create a Workload Management Workspace (WMW)
# [-V|--verbose] verbose output
# [-X] `set -x`

set -e

kcp_installed() {
    if [ "$(which kcp)" == "" ]; then
        echo "false"
    else
        echo "true"
    fi
}

kcp_running() {
    if [ "$(pgrep -f 'kcp start')" == "" ]; then
        echo "false"
    else
        echo "true"
    fi
}

kcp_ready() {
    if [ "$(kubectl ws root:compute 2> /dev/null)" == "" ]; then
        echo "false"
    else
        echo "true"
    fi
}

kcp_get_latest_version() {
    curl -sL https://github.com/kcp-dev/kcp/releases/latest | grep "</h1>" | head -n 1 | sed -e 's/<[^>]*>//g' | xargs
}

kubestellar_installed() {
    if [[ "$(which mailbox-controller)" == "" || "$(which placement-translator)" == "" || "$(which kubestellar-scheduler)" == "" ]]; then
        echo "false"
    else
        echo "true"
    fi
}

kubestellar_running() {
    if [[ "$(pgrep -f mailbox-controller)" == "" ||  "$(pgrep -f placement-translator)" == "" || "$(pgrep -f 'kubestellar-scheduler')" == "" ]]; then
        echo "false"
    else
        echo "true"
    fi
}

kubestellar_get_latest_version() {
    curl -sL https://github.com/kcp-dev/edge-mc/releases/latest | grep "</h1>" | head -n 1 | sed -e 's/<[^>]*>//g' | xargs
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
      *)        echo "Unsupported architecture type: $HOSTTYPE" >&2 ; exit 1 ;;
  esac
}

get_full_path() {
    echo "$(cd "$1"; pwd)"
}

kcp_version=""
kubestellar_version=""
os_type=""
arch_type=""
folder=""
kcp_address=""
kubestellar_imw=""
kubestellar_wmw=""
verbose=""
flagx=""

echo "KubeStellar bootstrap started..."

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
    (--bind-address)
        if (( $# > 1 ));
        then { kcp_address="$2"; shift; }
        else { echo "$0: missing ip address" >&2; exit 1; }
        fi;;
    (--ensure-imw)
        if (( $# > 1 ));
        then { kubestellar_imw="$2"; shift; }
        else { echo "$0: missing workspace name" >&2; exit 1; }
        fi;;
    (--ensure-wmw)
        if (( $# > 1 ));
        then { kubestellar_wmw="$2"; shift; }
        else { echo "$0: missing workspace name" >&2; exit 1; }
        fi;;
    (--ensure-folder)
        if (( $# > 1 ));
        then { folder="$2"; shift; }
        else { echo "$0: missing installation folder" >&2; exit 1; }
        fi;;
    (--verbose|-V)
        verbose="-V";;
    (-X)
	set -x
	flagx="-X";;
    (-h|--help)
        echo "Usage: $0 [--kcp-version release_version] [--kubestellar-version release_version] [--os linux|darwin] [--arch amd64|arm64] [--ensure-folder installation_folder] [-V|--verbose] [-X]"
        exit 0;;
    (-*)
        echo "$0: unknown flag" >&2 ; exit 1;
        exit 1;;
    (*)
        echo "$0: unknown positional argument" >&2; exit 1;
        exit 1;;
    esac
    shift
done

if [ "$kcp_version" == "" ]; then
    kcp_version=$(kcp_get_latest_version)
fi

if [ "$kubestellar_version" == "" ]; then
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

# Ensure kcp is installed
if [ "$(kcp_installed)" == "false" ]; then
    if [ "$verbose" != "" ]; then
        echo "Installing kcp..."
    fi
    bash <(curl -s https://raw.githubusercontent.com/kcp-dev/edge-mc/main/bootstrap/install-kcp-with-plugins.sh) --version $kcp_version --os $os_type --arch $arch_type --ensure-folder  $folder/kcp $verbose $flagx
    if [[ ! ":$PATH:" == *":$(get_full_path $folder/kcp/bin):"* ]]; then
        export PATH=$PATH:$(get_full_path $folder/kcp/bin)
    fi
fi

# Ensure kcp is running
if [ "$(kcp_running)" == "false" ]; then
    if [ "$kcp_address" == "" ]; then
        if [ "$verbose" != "" ]; then
            echo "Starting kcp..."
        fi
        kcp start >& kcp_log.txt &
    else
        if [ "$verbose" != "" ]; then
            echo "Starting kcp bound to address $kcp_address..."
        fi
        kcp start --bind-address $kcp_address >& kcp_log.txt &
    fi
    export KUBECONFIG="$(pwd)/.kcp/admin.kubeconfig"
    sleep 10
    until $(kcp_ready)
    do
        sleep 1
    done
    sleep 10
    echo 'Export KUBECONFIG with the command: export KUBECONFIG="$(pwd)/.kcp/admin.kubeconfig"'
fi

# Ensure KubeStellar is installed
if [ "$(kubestellar_installed)" == "false" ]; then
    if [ "$verbose" != "" ]; then
        echo "Installing Kubestellar..."
    fi
    bash <(curl -s https://raw.githubusercontent.com/kcp-dev/edge-mc/main/bootstrap/install-kubestellar.sh) --version $kubestellar_version --os $os_type --arch $arch_type --ensure-folder  $folder/kubestellar $verbose $flagx
    if [[ ! ":$PATH:" == *":$(get_full_path $folder/kubestellar/bin):"* ]]; then
        export PATH=$PATH:$(get_full_path $folder/kubestellar/bin)
    fi
fi

# Ensure KubeStellar is running
if [ "$(kubestellar_running)" == "false" ]; then
    if [ "$verbose" != "" ]; then
        echo "Starting Kubestellar..."
    fi
    kubestellar start $verbose
fi

# Ensure imw
if [ "$kubestellar_imw" != "" ]; then
    if [ "$verbose" != "" ]; then
        echo "Ensuring IMW $kubestellar_imw..."
    fi
    if ! kubectl ws $kubestellar_imw &> /dev/null
    then
        if [ "$verbose" != "" ]; then
            kubectl ws create $kubestellar_imw
        else
            kubectl ws create $kubestellar_imw > /dev/null
        fi
    fi
fi

# Ensure wmw
if [ "$kubestellar_wmw" != "" ]; then
    if [ "$verbose" != "" ]; then
        echo "Ensuring WMW $kubestellar_wmw..."
    fi
    if ! kubectl ws $kubestellar_imw &> /dev/null
    then
        if [ "$verbose" != "" ]; then
            kubectl kubestellar ensure wmw $kubestellar_wmw
        else
            kubectl kubestellar ensure wmw $kubestellar_wmw > /dev/null
        fi
    fi
fi
