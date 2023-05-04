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

# Usage: $0 --kcp-version v0.11.0 --kcp-edge-version v0.1.0 --imw imw-1 --wmw wmw-1

# This script installs KCP-Edge binaries to a folder of choice
#
# Arguments:


Usage: $0
# [--kcp-version release_version] set a specific KCP release version, default: latest
# [--kcp-edge-version release_version] set a specific KCP-Edge release version, default: latest
# [--os linux|darwin] set a specific OS type, default: autodetect
# [--arch amd64|arm64] set a specific architecture type, default: autodetect
# [--folder installation_folder] sets the installation folder, default: $PWD/kcp-edge
# [--create-folder] create the instllation folder, if it does not exist
# [--bind address] bind to KCP public address for access from remote pcluster
# [--imw name] create an inverntory management workspace under root
# [--wmw name] create an worklowd management workspace under root
# [-V|--verbose] verbose output

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

kcp_get_latest_version() {
    curl -sL https://github.com/kcp-dev/kcp/releases/latest | grep "</h1>" | head -n 1 | sed -e 's/<[^>]*>//g' | xargs
}

kcp_edge_installed() {
    if [[ "$(which mailbox-controller)" == "" || "$(which placement-translator)" == "" || "$(which scheduler)" == "" ]]; then
        echo "false"
    else
        echo "true"
    fi
}

kcp_edge_running() {
    if [[ "$(pgrep -f mailbox-controller)" == "" ||  "$(pgrep -f placement-translator)" == "" || "$(pgrep -f 'scheduler -v')" == "" ]]; then
        echo "false"
    else
        echo "true"
    fi
}

kcp_edge_get_latest_version() {
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
      x86_64*)    echo "amd64" ;;
      aarch64*) echo "arm64" ;;
      *)        echo "Unsupported architecture type: $HOSTTYPE" >&2 ; exit 1 ;;
  esac
}

get_full_path() {
    echo "$(cd "$1"; pwd)"
}

kcp_version=""
kcp_edge_version=""
os_type=""
arch_type=""
folder=""
create_folder=""
verbose=""
kcp_address=""
kcp_edge_imw=""
kcp_edge_wmw=""

while (( $# > 0 )); do
    case "$1" in
    (--kcp-version)
        if (( $# > 1 ));
        then { kcp_version="$2"; shift; }
        else { echo "$0: missing release version" >&2; exit 1; }
        fi;;
    (--kcp-edge-version)
        if (( $# > 1 ));
        then { kcp_edge_version="$2"; shift; }
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
    (--bind)
        if (( $# > 1 ));
        then { kcp_address="$2"; shift; }
        else { echo "$0: missing ip address" >&2; exit 1; }
        fi;;
    (--imw)
        if (( $# > 1 ));
        then { kcp_edge_imw="$2"; shift; }
        else { echo "$0: missing workspace name" >&2; exit 1; }
        fi;;
    (--wmw)
        if (( $# > 1 ));
        then { kcp_edge_wmw="$2"; shift; }
        else { echo "$0: missing workspace name" >&2; exit 1; }
        fi;;
    (--folder)
        if (( $# > 1 ));
        then { folder="$2"; shift; }
        else { echo "$0: missing installation folder" >&2; exit 1; }
        fi;;
    (--create-folder)
        create_folder="true";;
    (--verbose|-V)
        verbose="-V";;
    (-h|--help)
        echo "Usage: $0 [--kcp-version release_version] [--kcp-edge-version release_version] [--os linux|darwin] [--arch amd64|arm64] [--folder installation_folder] [--create-folder] [--bind address] [--imw name] [--wmw name] [-V|--verbose]"
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

if [ "$kcp_edge_version" == "" ]; then
    kcp_edge_version=$(kcp_edge_get_latest_version)
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
    bash <(curl -s https://raw.githubusercontent.com/kcp-dev/edge-mc/main/hack/install-kcp-with-plugins.sh) --version $kcp_version --os $os_type --arch $arch_type --folder  $folder/kcp --create-folder $verbose
    if [[ ! ":$PATH:" == *":$(get_full_path $folder/kcp/bin):"* ]]; then
        export PATH=$PATH:$(get_full_path $folder/kcp/bin)
    fi
fi

# Ensure kcp is running
if [ "$(kcp_running)" == "false" ]; then
    if [ "$verbose" != "" ]; then
        echo "Starting kcp..."
    fi
    if [ "$kcp_address" == "" ]; then
        kcp start >& kcp_log.txt &
    else
        kcp start --bind-address $kcp_address >& kcp_log.txt &
    fi
    export KUBECONFIG="$(pwd)/.kcp/admin.kubeconfig"
    sleep 5
    until kubectl ws . &> /dev/null
    do
        sleep 1
    done
    sleep 10
    echo 'Export KUBECONFIG: export KUBECONFIG="$(pwd)/.kcp/admin.kubeconfig"'
fi

# Ensure kcp-edge is installed
if [ "$(kcp_edge_installed)" == "false" ]; then
    if [ "$verbose" != "" ]; then
        echo "Installing kcp-edge..."
    fi
    bash <(curl -s https://raw.githubusercontent.com/kcp-dev/edge-mc/main/hack/install-kcp-edge.sh) --version $kcp_edge_version --os $os_type --arch $arch_type --folder  $folder/kcp-edge --create-folder $verbose
    if [[ ! ":$PATH:" == *":$(get_full_path $folder/kcp-edge/bin):"* ]]; then
        export PATH=$PATH:$(get_full_path $folder/kcp-edge/bin)
    fi
fi

# Ensure kcp-edge is running
if [ "$(kcp_edge_running)" == "false" ]; then
    if [ "$verbose" != "" ]; then
        echo "Starting kcp-edge..."
    fi
    bash <(curl -s https://raw.githubusercontent.com/dumb0002/edge-mc/user-dev-kit/environments/dev-env/kcp-edge.sh) start --user kit $verbose
fi

# Ensure imw
if [ "$kcp_edge_imw" != "" ]; then
    if [ "$verbose" != "" ]; then
        echo "Ensuring IMW root:$kcp_edge_imw..."
    fi
    if ! kubectl ws root:$kcp_edge_imw &> /dev/null
    then
        if [ "$verbose" != "" ]; then
            kubectl ws root
            kubectl ws create $kcp_edge_imw
        else
            kubectl ws root > /dev/null
            kubectl ws create $kcp_edge_imw > /dev/null
        fi
    fi
fi

# Ensure wmw
if [ "$kcp_edge_wmw" != "" ]; then
    if [ "$verbose" != "" ]; then
        echo "Ensuring WMW root:$kcp_edge_wmw..."
    fi
    if ! kubectl ws root:$kcp_edge_wmw &> /dev/null
    then
        if [ "$verbose" != "" ]; then
            kubectl ws root
            ensure-wmw.sh wmw-1
            kubectl ws root
        else
            kubectl ws root > /dev/null
            ensure-wmw.sh wmw-1 > /dev/null
            kubectl ws root > /dev/null
        fi
    fi
fi
