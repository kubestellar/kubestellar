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

kcp_download() {
    if [ $verbose == "true" ]; then
        curl -SL -o kcp.tar.gz "https://github.com/kcp-dev/kcp/releases/download/${kcp_version}/kcp_${kcp_version//v}_${os_type}_${arch_type}.tar.gz"
        curl -SL -o kcp-plugins.tar.gz "https://github.com/kcp-dev/kcp/releases/download/${kcp_version}/kubectl-kcp-plugin_${kcp_version//v}_${os_type}_${arch_type}.tar.gz"
    else
        curl -sSL -o kcp.tar.gz "https://github.com/kcp-dev/kcp/releases/download/${kcp_version}/kcp_${kcp_version//v}_${os_type}_${arch_type}.tar.gz"
        curl -sSL -o kcp-plugins.tar.gz "https://github.com/kcp-dev/kcp/releases/download/${kcp_version}/kubectl-kcp-plugin_${kcp_version//v}_${os_type}_${arch_type}.tar.gz"
    fi
}

kcp_install() {
    tar -C $kcp_folder -zxf kcp-plugins.tar.gz
    tar -C $kcp_folder -zxf kcp.tar.gz
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

kcp_version() {
    echo "$(kubectl version --short 2> /dev/null | grep kcp | sed 's/.*kcp-//')"
}

kcp_get_latest_version() {
    curl -sL https://github.com/kcp-dev/kcp/releases/latest | grep "</h1>" | head -n 1 | sed -e 's/<[^>]*>//g' | xargs
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

kubestellar_installed() {
    if [[ "$(which mailbox-controller)" == "" || "$(which placement-translator)" == "" || "$(which kubestellar-scheduler)" == "" ]]; then
        echo "false"
    else
        echo "true"
    fi
}

kubestellar_download() {
    if [ $verbose == "true" ]; then
        curl -SL -o kubestellar.tar.gz "https://github.com/kcp-dev/edge-mc/releases/download/${kubestellar_version}/kubestellar_${kubestellar_version}_${os_type}_$arch_type.tar.gz"
    else
        curl -sSL -o kubestellar.tar.gz "https://github.com/kcp-dev/edge-mc/releases/download/${kubestellar_version}/kubestellar_${kubestellar_version}_${os_type}_$arch_type.tar.gz"
    fi
}

kubestellar_install() {
    tar -C $kubestellar_folder -zxf kubestellar.tar.gz
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
kcp_address=""
kubestellar_imw=""
kubestellar_wmw=""
verbose="false"
flagx=""
user_exports=""

echo "************************************************"
echo "KubeStellar bootstrap started"
echo "************************************************"

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

# Ensure kcp is installed
echo "************************************************"
echo "Ensure kcp is installed"
echo "************************************************"
if [ "$(kcp_installed)" == "true" ]; then
    echo "kcp found in the PATH at '$(which kcp)' ... skip installation."
else
    echo "kcp not found in the PATH."
    if [ "$(kcp_running)" == "true" ]; then
        echo "kcp process is running already: pid=$(pgrep kcp) ... add kcp folder to the PATH or stop kcp ... exiting."
        exit 2
    fi
    ensure_folder "$folder/kcp"
    kcp_folder=$(get_full_path "$folder/kcp")
    kcp_bin_folder="$kcp_folder/bin"
    echo "Downloading kcp+plugins $kcp_version $os_type/$arch_type..."
    kcp_download
    echo "Installing kcp+plugins into '$kcp_folder'..."
    kcp_install
    if [[ ! ":$PATH:" == *":$kcp_bin_folder:"* ]]; then
        export PATH=$kcp_bin_folder:$PATH
        echo "Add kcp folder to the PATH: export PATH=\"$kcp_bin_folder:\$PATH\""
        user_exports="$user_exports"$'\n'"export PATH=\"$kcp_bin_folder:\$PATH\""
    fi
fi

# Ensure kcp is running
echo "************************************************"
echo "Ensure kcp is running"
echo "************************************************"
if [ "$(kcp_running)" == "true" ]; then
    echo "kcp process is running already: pid=$(pgrep kcp) ... skip running."
    if [ "$(kubeconfig_valid)" == "false" ]; then
        echo "KUBECONFIG environment variable is not set correctly: KUBECONFIG='$KUBECONFIG' ... exiting!"
        exit 3
    fi
    echo "Using 'KUBECONFIG=$KUBECONFIG'"
    echo "Waiting for kcp to be ready... it may take a while"
    until $(kcp_ready)
    do
        sleep 1
    done
    if [ "$(kcp_version)" != "$kcp_version" ]; then
        echo "kcp running version $(kcp_version) does not match the desired version $kcp_version ... exiting!"
        exit 4
    else
        echo "kcp version $(kcp_version) ... ok"
    fi
else
    if [ "$kcp_address" == "" ]; then
        echo "Running kcp... logfile=$PWD/kcp_log.txt"
        kcp start >& kcp_log.txt &
    else
        echo "Running kcp bound to address $kcp_address... logfile=$PWD/kcp_log.txt"
        kcp start --bind-address $kcp_address >& kcp_log.txt &
    fi
    export KUBECONFIG="$PWD/.kcp/admin.kubeconfig"
    echo "Waiting for kcp to be ready... it may take a while"
    sleep 10
    until $(kcp_ready)
    do
        sleep 1
    done
    sleep 10
    if [ "$(kcp_version)" != "$KCP_REQUIRED_VERSION" ]; then
        echo "kcp version $(kcp_version) is not supported, KubeStellar requires kcp $KCP_REQUIRED_VERSION ... exiting!"
        exit 4
    else
        echo "kcp version $(kcp_version) ... ok"
    fi
    echo "Export KUBECONFIG environment variable: export KUBECONFIG=\"$KUBECONFIG\""
    user_exports="$user_exports"$'\n'"export KUBECONFIG=\"$KUBECONFIG\""
fi

# Ensure KubeStellar is installed
echo "************************************************"
echo "Ensure KubeStellar is installed"
echo "************************************************"
if [ "$(kubestellar_installed)" == "true" ]; then
    echo "KubeStellar found in the PATH at '$(which kubestellar)' ... skip installation."
else
    echo "KubeStellar not found in the PATH."
    if [ "$(kubestellar_running)" == "true" ]; then
        echo "KubeStellar processes are running already ... add KubeStellar folder to the PATH or stop KubeStellar with \"kubestellar stop\" ... exiting."
        exit 5
    fi
    ensure_folder "$folder/kubestellar"
    kubestellar_folder=$(get_full_path "$folder/kubestellar")
    kubestellar_bin_folder="$kubestellar_folder/bin"
    echo "Downloading KubeStellar $kubestellar_version $os_type/$arch_type..."
    kubestellar_download
    echo "Installing KubeStellar into '$kubestellar_folder'..."
    kubestellar_install
    if [[ ! ":$PATH:" == *":$kubestellar_bin_folder:"* ]]; then
        export PATH=$kubestellar_bin_folder:$PATH
        echo "Add KubeStellar folder to the PATH: export PATH=\"$kubestellar_bin_folder:\$PATH\""
        user_exports="$user_exports"$'\n'"export PATH=\"$kubestellar_bin_folder:\$PATH\""
    fi
fi

# Ensure KubeStellar is running
echo "************************************************"
echo "Ensure KubeStellar is running"
echo "************************************************"
kubectl ws root > /dev/null
if [ "$(kubestellar_running)" == "true" ]; then
    echo "KubeStellar processes are running ... ok"
    if ! kubectl get workspaces.tenancy.kcp.io espw &> /dev/null ; then
        echo "KubeStellar ESPW does not exists ... run 'kubestellar stop' first ... exiting!"
        exit 6
    else
        echo "KubeStellar ESPW found ... ok"
    fi
else
    echo "Starting or restarting KubeStellar..."
    if [ $verbose == "true" ]; then
        kubestellar start -V
    else
        kubestellar start
    fi
fi

# Ensure imw
if [ "$kubestellar_imw" != "" ]; then
    echo "Ensuring IMW \"$kubestellar_imw\"..."
    if ! kubectl ws $kubestellar_imw &> /dev/null ; then
        if [ "$verbose" != "" ]; then
            kubectl ws create $kubestellar_imw
        else
            kubectl ws create $kubestellar_imw > /dev/null
        fi
    fi
fi

# Ensure wmw
if [ "$kubestellar_wmw" != "" ]; then
    echo "Ensuring WMW \"$kubestellar_wmw\"..."
    if ! kubectl ws $kubestellar_imw &> /dev/null ; then
        if [ "$verbose" != "" ]; then
            kubectl kubestellar ensure wmw $kubestellar_wmw
        else
            kubectl kubestellar ensure wmw $kubestellar_wmw > /dev/null
        fi
    fi
fi


if [ "$verbose" == "true" ]; then
    kubectl ws root
    kubectl ws tree
else
    kubectl ws root > /dev/null
fi

echo "************************************************"
echo "KubeStellar bootstrap completed succesfully"
echo "************************************************"
if [ "$user_exports" != "" ]; then
    echo "Please create/update the following enviroment variables: $user_exports"
fi
