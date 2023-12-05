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

# This script deploys Kubestellar, either as bare processes
# or as Kubernetes workload, and installs executables to a folder of choice
#
# Arguments:
# [--deploy bool] indicate if kubestellar will be deployed
# [--external-endpoint domain-name:port]
# [--openshift bool]
# [--kubestellar-version release] set a specific KubeStellar release version, default: latest
# [--os linux|darwin] set a specific OS type, default: autodetect
# [--arch amd64|arm64] set a specific architecture type, default: autodetect
# [--ensure-folder name] sets the installation folder, default: $PWD/kubestellar
# [--bind-address address] bind the space provider to a specific ip address
# [--ensure-imw name] create a Inventory Management Workspace (IMW)
# [--ensure-wmw name] create a Workload Management Workspace (WMW)
# [--host-ns namespace] specify namespace in hosting cluster that gets chart
# [-V|--verbose] verbose output
# [-X] `set -x`

set -e

sp_installed() {
    if [[ "$(which $sp_name)" == "" || "$(which kubectl-ws)" == "" ]]; then
        echo "false"
    else
        echo "true"
    fi
}

sp_download() {
    if [ $verbose == "true" ]; then
        curl -SL -o "${sp_name}.tar.gz" "https://github.com/${sp_name}-dev/${sp_name}/releases/download/${sp_version}/${sp_name}_${sp_version//v}_${os_type}_${arch_type}.tar.gz"
        curl -SL -o "${sp_name}-plugins.tar.gz" "https://github.com/${sp_name}-dev/${sp_name}/releases/download/${sp_version}/kubectl-${sp_name}-plugin_${sp_version//v}_${os_type}_${arch_type}.tar.gz"
    else
        curl -sSL -o "${sp_name}.tar.gz" "https://github.com/${sp_name}-dev/${sp_name}/releases/download/${sp_version}/${sp_name}_${sp_version//v}_${os_type}_${arch_type}.tar.gz"
        curl -sSL -o "${sp_name}-plugins.tar.gz" "https://github.com/${sp_name}-dev/${sp_name}/releases/download/${sp_version}/kubectl-${sp_name}-plugin_${sp_version//v}_${os_type}_${arch_type}.tar.gz"
    fi
}

sp_install() {
    tar -C $sp_folder -zxf "${sp_name}-plugins.tar.gz"
    tar -C $sp_folder -zxf "${sp_name}.tar.gz"
}

sp_running() {
    if [ "$(pgrep -f "${sp_name} start")" == "" ]; then
        echo "false"
    else
        echo "true"
    fi
}

sp_ready() {
    if [ "$(kubectl ws root:compute 2> /dev/null)" == "" ]; then
        echo "false"
    else
        echo "true"
    fi
}

sp_version() {
    if [ "$(kubectl version --short 2> /dev/null | grep ${sp_name} | sed "s/.*${sp_name}-//")" != "" ]; then
        echo "$(kubectl version --short 2> /dev/null | grep ${sp_name} | sed "s/.*${sp_name}-//")"
    else
        echo "$(kubectl version 2> /dev/null | grep ${sp_name} | sed "s/.*${sp_name}-//")"
    fi
}

sp_get_latest_version() {
    curl -sL "https://github.com/${sp_name}-dev/${sp_name}/releases/latest" | grep "</h1>" | tail -n 1 | sed -e 's/<[^>]*>//g' | xargs
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
    echo "$(cd "$1"; pwd)"
}

ensure_folder() {
    if [ -d "$1" ]; then :
    else
        mkdir -p "$1"
    fi
}

export sp_name="kcp"
sp_version="v0.11.0"
kubestellar_version=""
deploy="true"
os_type=""
arch_type=""
folder=""
sp_address=""
kubestellar_imw="imw1"
kubestellar_wmw="wmw1"
verbose="false"
flagx=""
user_exports=""
external_endpoint=""
openshift=false
host_ns="kubestellar"

echo "< KubeStellar bootstrap started >----------------"

while (( $# > 0 )); do
    case "$1" in
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
        then { sp_address="$2"; shift; }
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
    (--host-ns)
        if (( $# > 1 ));
        then { host_ns="$2"; shift; }
        else { echo "$0: missing host-ns setting" >&2; exit 1; }
        fi;;
    (--verbose|-V)
        verbose="true";;
    (-X)
	set -x
	flagx="-X";;
    (-h|--help)
        echo "Usage: $0 [--kubestellar-version release_version] [--deploy bool] [--os linux|darwin] [--arch amd64|arm64] [--ensure-folder installation_folder] [--ensure-imw imw-list] [--ensure-wmw wmw-list] [--host-ns namespace-in-hosting-cluster] [-V|--verbose] [-X]"
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

if [ $(wc -w <<<"$host_ns") != 1 ]; then
    echo "$0: --host-ns value must be one word, not '$host_ns'" >&2
    exit 1
fi

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

# Ensure space provider is installed
echo "< Ensure that the space provider is installed >----------------------"
if [ "$(sp_installed)" == "true" ]; then
    echo "Space provider found in the PATH at \"$(which ${sp_name})\" ... skip installation."
else
    echo "Space provider not found in the PATH."
    if [ "$(sp_running)" == "true" ]; then
        echo "Space provider process is running already: pid=$(pgrep ${sp_name}) ... add the space provider folder to the PATH or stop the space provider ... exiting."
        exit 2
    fi
    ensure_folder "$folder/${sp_name}"
    sp_folder=$(get_full_path "$folder/${sp_name}")
    sp_bin_folder="$sp_folder/bin"
    echo "Downloading the space provider and plugins $sp_version $os_type/$arch_type..."
    sp_download
    echo "Installing the space provider and plugins into '$sp_folder'..."
    sp_install
    rm "${sp_name}.tar.gz" "${sp_name}-plugins.tar.gz"
    if [[ ! ":$PATH:" == *":$sp_bin_folder:"* ]]; then
        export PATH=$sp_bin_folder:$PATH
        echo "Add the space provider folder to the PATH: export PATH=\"$sp_bin_folder:\$PATH\""
        user_exports="$user_exports"$'\n'"export PATH=\"$sp_bin_folder:\$PATH\""
    fi
fi

if [ "$deploy" == true ] && [ "$deploy_style" == bare ]; then
    # Ensure the space provider is running
    echo "< Ensure that the space provider is running >-----------------------"
    if [ "$(sp_running)" == "true" ]; then
	echo "The space provider process is running already: pid=$(pgrep ${sp_name}) ... skip running."
	if [ "$(kubeconfig_valid)" == "false" ]; then
            echo "KUBECONFIG environment variable is not set correctly: KUBECONFIG='$KUBECONFIG' ... exiting!"
            exit 3
	fi
	echo "Using 'KUBECONFIG=$KUBECONFIG'"
	echo "Waiting for the space provider to be ready... it may take a while"
	until $(sp_ready)
	do
            sleep 1
	done
	found_sp_version="$(sp_version)"
	if [ "$found_sp_version" != "$sp_version" ]; then
            echo "Space provider running version ${found_sp_version@Q} does not match the desired version $sp_version ... exiting!"
	    echo "FYI: \`kubectl version --short\` reports: $(kubectl version --short)"
            exit 4
	else
            echo "Space provider version $(sp_version) ... ok"
	fi
    else
	if [ "$sp_address" == "" ]; then
            echo "Running the space provider... logfile=$PWD/${sp_name}_log.txt"
            ${sp_name} start >& "${sp_name}_log.txt" &
	else
            echo "Running the space provider bound to address $sp_address... logfile=$PWD/${sp_name}_log.txt"
            ${sp_name} start --bind-address $sp_address >& "${sp_name}_log.txt" &
	fi
	export KUBECONFIG="$PWD/.${sp_name}/admin.kubeconfig"
	echo "Waiting for the space provider to be ready... it may take a while"
	sleep 10
	until $(sp_ready)
	do
            sleep 1
	done
	sleep 10
	found_sp_version="$(sp_version)"
	if [ "$found_sp_version" != "$sp_version" ]; then
            echo "Space provider version ${found_sp_version@Q} is not supported, KubeStellar requires ${sp_name} $sp_version ... exiting!"
	    echo "FYI: \`kubectl version --short\` reports: $(kubectl version --short)"
            exit 4
	else
            echo "Space provider version $(sp_version) ... ok"
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
            kubestellar start --ensure-imw $kubestellar_imw --ensure-wmw $kubestellar_wmw -V
	else
            kubestellar start --ensure-imw $kubestellar_imw --ensure-wmw $kubestellar_wmw
	fi
    fi
else
    echo "< Deploy KubeStellar into Kubernetes, namespace=$host_ns >---------------"
    kubectl get ns $host_ns || kubectl create ns $host_ns
    kubectl kubestellar deploy "${deploy_flags[@]}" -n $host_ns
    echo "< Waiting for startup and fetching $folder/kubestellar.kubeconfig >---------------"
    kubectl kubestellar get-external-kubeconfig -n $host_ns -o "$folder"/kubestellar.kubeconfig
    export KUBECONFIG="$folder/kubestellar.kubeconfig"
    echo "Export KUBECONFIG environment variable: export KUBECONFIG=\"$KUBECONFIG\""
    user_exports="$user_exports"$'\n'"export KUBECONFIG=\"$KUBECONFIG\""
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
