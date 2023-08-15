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

# This script deploys kcp and Kubestellar, either as bare processes
# or as Kubernetes workload, and installs executables to a folder of choice
#
# Arguments:
# [--deploy bool] indicate if kubestellar (and kcp) will be deployed
# [--external-endpoint domain-name:port]
# [--openshift bool]
# [--kcp-version release] set a specific kcp release version, default: latest
# [--kubestellar-version release] set a specific KubeStellar release version, default: latest
# [--os linux|darwin] set a specific OS type, default: autodetect
# [--arch amd64|arm64] set a specific architecture type, default: autodetect
# [--ensure-folder name] sets the installation folder, default: $PWD/kubestellar
# [--bind-address address] bind kcp to a specific ip address
# [--ensure-imw name] create a Inventory Management Workspace (IMW)
# [--ensure-wmw name] create a Workload Management Workspace (WMW)
# [-V|--verbose] verbose output
# [-X] `set -x`

set -e

kcp_installed() {
    if [[ "$(which kcp)" == "" || "$(which kubectl-ws)" == "" ]]; then
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
    kubectl version --short
    RESULT=$?
    if [ $RESULT -eq 0 ]; then
        echo "$(kubectl version --short 2> /dev/null | grep kcp | sed 's/.*kcp-//')"
    else
        echo "$(kubectl version 2> /dev/null | grep kcp | sed 's/.*kcp-//')"
    fi
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

kubestellar_installed() {
    if [[ "$(which mailbox-controller)" == "" || "$(which placement-translator)" == "" || "$(which kubestellar-where-resolver)" == "" ]]; then
        echo "false"
    else
        echo "true"
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
    tar -C $kubestellar_folder -zxf kubestellar.tar.gz
}

kubestellar_running() {
    if [[ "$(pgrep -f mailbox-controller)" == "" ||  "$(pgrep -f placement-translator)" == "" || "$(pgrep -f 'kubestellar-where-resolver')" == "" ]]; then
        echo "false"
    else
        echo "true"
    fi
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
deploy="true"
os_type=""
arch_type=""
folder=""
kcp_address=""
kubestellar_imw=""
kubestellar_wmw=""
verbose="false"
flagx=""
user_exports=""
external_endpoint=""
openshift=false

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
    (--deploy)
        if (( $# > 1 ));
        then { deploy="$2"; shift; }
        else { echo "$0: missing value for --deploy" >&2; exit 1; }
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
    (--external-endpoint)
        if (( $# > 1 ));
        then { external_endpoint="$2"; shift; }
        else { echo "$0: missing external-endpdoint" >&2; exit 1; }
        fi;;
    (--openshift)
        if (( $# > 1 ));
        then { openshift="$2"; shift; }
        else { echo "$0: missing openshift setting" >&2; exit 1; }
        fi;;
    (--verbose|-V)
        verbose="true";;
    (-X)
	set -x
	flagx="-X";;
    (-h|--help)
        echo "Usage: $0 [--kcp-version release_version] [--kubestellar-version release_version] [--deploy bool] [--os linux|darwin] [--arch amd64|arm64] [--ensure-folder installation_folder] [-V|--verbose] [-X]"
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

case "$deploy" in
    (true|false) ;;
    (*) echo "$0: --deploy value must be 'true' or 'false'" >&2
	exit 1;;
esac

case "$openshift" in
    ("") ;;
    (true|false) ;;
    (*) echo "$0: --openshift value must be 'true' or 'false'" >&2
	exit 1;;
esac

deploy_style=bare
deploy_flags=()

case "$external_endpoint" in
    ("")    ;;
    (*:*:*) echo "$0: --external-endpoint must have the form domain-name:port" >& 2
	    exit 1;;
    (*:*)   deploy_style=kube
	    deploy_flags=("--external-endpoint" $external_endpoint);;
    (*)     echo "$0: --external-endpoint must have the form domain-name:port" >& 2
	    exit 1;;
esac

if [ "$openshift" == true ]; then
    deploy_style=kube
    deploy_flags[${#deploy_flags[*]}]="--openshift"
    deploy_flags[${#deploy_flags[*]}]=true
fi

# Ensure kcp is installed
echo "< Ensure kcp is installed >----------------------"
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
    rm kcp.tar.gz kcp-plugins.tar.gz
    if [[ ! ":$PATH:" == *":$kcp_bin_folder:"* ]]; then
        export PATH=$kcp_bin_folder:$PATH
        echo "Add kcp folder to the PATH: export PATH=\"$kcp_bin_folder:\$PATH\""
        user_exports="$user_exports"$'\n'"export PATH=\"$kcp_bin_folder:\$PATH\""
    fi
fi

if [ "$deploy" == true ] && [ "$deploy_style" == bare ]; then
    # Ensure kcp is running
    echo "< Ensure kcp is running >-----------------------"
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
	found_kcp_version="$(kcp_version)"
	if [ "$found_kcp_version" != "$kcp_version" ]; then
            echo "kcp running version ${found_kcp_version@Q} does not match the desired version $kcp_version ... exiting!"
	    echo "FYI: \`kubectl version --short\` reports: $(kubectl version --short)"
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
	found_kcp_version="$(kcp_version)"
	if [ "$found_kcp_version" != "$KCP_REQUIRED_VERSION" ]; then
            echo "kcp version ${found_kcp_version@Q} is not supported, KubeStellar requires kcp $KCP_REQUIRED_VERSION ... exiting!"
	    echo "FYI: \`kubectl version --short\` reports: $(kubectl version --short)"
            exit 4
	else
            echo "kcp version $(kcp_version) ... ok"
	fi
	echo "Export KUBECONFIG environment variable: export KUBECONFIG=\"$KUBECONFIG\""
	user_exports="$user_exports"$'\n'"export KUBECONFIG=\"$KUBECONFIG\""
    fi
fi


# Ensure KubeStellar is installed
echo "< Ensure KubeStellar is installed >-------------"
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
    rm kubestellar.tar.gz
    if [[ ! ":$PATH:" == *":$kubestellar_bin_folder:"* ]]; then
        export PATH=$kubestellar_bin_folder:$PATH
        echo "Add KubeStellar folder to the PATH: export PATH=\"$kubestellar_bin_folder:\$PATH\""
        user_exports="$user_exports"$'\n'"export PATH=\"$kubestellar_bin_folder:\$PATH\""
    fi
fi


if [ "$deploy" == false ]; then :
elif [ "$deploy_style" == bare ]; then
    # Ensure KubeStellar is running
    echo "< Ensure KubeStellar is running >---------------"
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
else
    echo "< Deploy kcp and KubeStellar into Kubernetes >---------------"
    kubectl kubestellar deploy "${deploy_flags[@]}"
    echo "< Waiting for startup and fetching $folder/kubestellar.kubeconfig >---------------"
    kubectl kubestellar get-external-kubeconfig -o "$folder"/kubestellar.kubeconfig
    export KUBECONFIG="$folder/kubestellar.kubeconfig"
    echo "Export KUBECONFIG environment variable: export KUBECONFIG=\"$KUBECONFIG\""
    user_exports="$user_exports"$'\n'"export KUBECONFIG=\"$KUBECONFIG\""
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


if [ "$deploy" == true ]; then
    if [ "$verbose" == "true" ]; then
	kubectl ws root
	kubectl ws tree
    else
	kubectl ws root > /dev/null
    fi
fi

echo "< KubeStellar bootstrap completed successfully >-"
if [ "$user_exports" != "" ]; then
    echo "Please create/update the following environment variables: $user_exports"
fi
