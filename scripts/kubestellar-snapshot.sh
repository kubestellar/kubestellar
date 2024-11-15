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


# Exit on error
set -e


# Global variables
timestamp="$(date +%F_%T)"
output_folder="/tmp/kubestellar-snapshot"


# Script info
SCRIPT_NAME="KubeStellar Snapshot"
SCRIPT_VERSION="0.10"


# Colors
COLOR_NONE="\033[0m"
COLOR_RED="\033[1;31m"
COLOR_GREEN="\033[1;32m"
COLOR_BLUE="\033[94m"
COLOR_YELLOW="\033[1;33m"
COLOR_PURPLE="\033[1;35m"
COLOR_ERROR="${COLOR_RED}"
COLOR_STATUS_TRUE="${COLOR_GREEN}"
COLOR_STATUS_FALSE="${COLOR_RED}"
COLOR_INFO="${COLOR_BLUE}"
COLOR_TITLE="${COLOR_YELLOW}"
COLOR_YAML="${COLOR_PURPLE}"


# Command line arguments
arg_kubeconfig=""
arg_context=""
arg_logs=false
arg_yaml=false
arg_verbose=false


# Display the command line help
display_help() {
  cat << EOF
Usage: $0 [--kubeconfig|-K <filename>] [--context|-C <name>] [--logs|-L] [--yaml|-Y] [--verbose|-V] [--version|-v] [--help|-h] [-X]

--kubeconfig|-K <filename> use the specified kubeconfig to find KubeStellar Helm chart
--context|-C <name>        use the specified context to find KubeStellar Helm chart
--logs|-L                  save the logs of the pods
--yaml|-Y                  save the YAML of the resources
--verbose|-V               output extra information
--version|-v               print out the script version
--help|-h                  show this information
-X                         enable verbose execution for debugging

Note: After piping the output of the script to file use "more"
      to inspect the file content in color. ALternatively install
      "ansi2txt" and pipe the script through it to remove escape
      characters and obtain a plan text report.
EOF
}


# Indent JSON
indent() {
    sed 's/^/  /'
}

# Echo in color
echocolor() {
    # $1 = color
    # $2 = message
    echo -e "$1$2${COLOR_NONE}"
}


# Echo to stderr
echoerr() {
    # $1 = error message
    >&2 echocolor ${COLOR_ERROR} "ERROR: $1"
}


# Echo colorized title
echotitle() {
    # $1 = message
    echocolor ${COLOR_TITLE} "\n$1"
}


# Echo the status in color
echostatus() {
    # $1 = status text
    status="$(echo $1 | tr '[:upper:]' '[:lower:]')" # lowercase
    if [[ "true succeeded running active 1" =~  "$status" ]] ; then
        echocolor ${COLOR_STATUS_TRUE} "$status"
    else
        echocolor ${COLOR_STATUS_FALSE} "$status"
    fi
}


# Check if a pre-requisite is installed
is_installed() {
    # $1 == name
    # $2 == command name to search
    # $3 == command to get the version, unstructured
    if which $2 > /dev/null ; then
        echov -e "${COLOR_GREEN}\xE2\x9C\x94${COLOR_NONE} ${COLOR_INFO}$1${COLOR_NONE} version ${COLOR_INFO}$(eval "$3" 2> /dev/null || true)${COLOR_NONE} at ${COLOR_INFO}$(which $2)${COLOR_NONE}"
    else
        echoerr "missing dependency:"
        echo -e "${COLOR_RED}X${COLOR_NONE} $1"
        exit 1
    fi
}


# Get the kubeconfig of a particular Control Plane
get_kubeconfig() {
    context="$1" # context in the kubeconfig
    cp_name="$2" # name of the Control Plane
    cp_type="$3" # type of the Control Plane

    # wait for CP ready
    while [[ $(k --context $context get controlplane $cp_name -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; do
        sleep 5
    done

    # put into the shell variable "kubeconfig" the kubeconfig contents for use from outside of the hosting cluster
    if [[ "$cp_type" == "host" ]] ; then
        kubeconfig=$(k --context $context config view --flatten --minify)
    else
        # determine the secret name and namespace
                     key=$(k --context $context get controlplane $cp_name -o=jsonpath='{.status.secretRef.key}')
             secret_name=$(k --context $context get controlplane $cp_name -o=jsonpath='{.status.secretRef.name}')
        secret_namespace=$(k --context $context get controlplane $cp_name -o=jsonpath='{.status.secretRef.namespace}')
        # get the kubeconfig
        kubeconfig=$(k --context $context get secret $secret_name -n $secret_namespace -o=jsonpath="{.data.$key}" | base64 -d)
    fi

    # return the kubeconfig file contents in YAML format
    echo "$kubeconfig"
}


###############################################################################
# Parse command line arguments
###############################################################################
while (( $# > 0 )); do
    case "$1" in
    (--kubeconfig|-K)
        if (( $# > 1 ));
        then { arg_kubeconfig="$2"; shift; }
        else { echo "$0: missing kubeconfig filename" >&2; exit 1; }
        fi;;
    (--context|-C)
        if (( $# > 1 ));
        then { arg_context="$2"; shift; }
        else { echo "$0: missing context name" >&2; exit 1; }
        fi;;
    (--logs|-L)
        arg_logs=true;;
    (--yaml|-Y)
        arg_yaml=true;;
    (--verbose|-V)
        arg_verbose=true;;
    (--version|-v)
        echo "${SCRIPT_NAME} v${SCRIPT_VERSION}"
        exit 0;;
    (-X)
        set -x;;
    (-h|--help)
        display_help
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


###############################################################################
# Define alias definition
###############################################################################
# Define the echov function based on verbosity
if [ "$arg_verbose" == "true" ]; then
    echov() { echo "$@" ; }
else
    echov() { :; }
fi

# Define the special kubectl
k() {
    KUBECONFIG="$kubeconfig" kubectl $*
}


###############################################################################
# Script info
###############################################################################
echov -e "${COLOR_INFO}$($0 --version)${COLOR_NONE}\n"
echov -e "Script run on ${COLOR_INFO}$timestamp${COLOR_NONE}"


###############################################################################
# Check dependencies
###############################################################################
echov "Checking script dependencies:"

is_installed 'kubectl' \
    'kubectl' \
    'kubectl version --client | head -1 | cut -d" " -f3'

is_installed 'helm' \
    'helm' \
    'helm version | cut -d"\"" -f2'

is_installed 'jq' \
    'jq' \
    'jq --version'


###############################################################################
# Ensure output folder
###############################################################################
if [[ "$arg_logs" == "true" || "$arg_yaml" == "true" ]] ; then
    yes "yes" | rm -vRI "/tmp/kubestellar-snapshot" > /dev/null 2>&1 || true
    mkdir -p "$output_folder"
fi


###############################################################################
# Determine the list of kubeconfigs to search
# based on https://kubernetes.io/docs/reference/k/generated/k_config/
###############################################################################
if [[ "$arg_kubeconfig" != "" ]] ; then
    kubeconfig="$arg_kubeconfig"
elif [[ "$KUBECONFIG" != "" ]] ; then
    kubeconfig="$KUBECONFIG"
else
    kubeconfig="$HOME/.kube/config"
fi
echov -e "Using kubeconfig(s): ${COLOR_INFO}$kubeconfig${COLOR_NONE}"


###############################################################################
# Determine the list of contexts to search
###############################################################################
if [[ "$arg_context" == "" ]] ; then
    contexts=($(k config get-contexts --no-headers -o name))
else
    contexts=("$arg_context")
fi

echov "Validating contexts(s): "
vc_n=0
valid_context=()
for context in "${contexts[@]}" ; do # for all contexts
    [[ -z "$context" ]] && continue
    if k --context $context get secrets -A > /dev/null 2>&1 ; then
        echov -e "${COLOR_GREEN}\xE2\x9C\x94${COLOR_NONE} ${COLOR_INFO}${context}${COLOR_NONE}"
        valid_context[vc_n]="$context"
        vc_n=$((vc_n+1))
    else
        echov -e "${COLOR_RED}X${COLOR_NONE} ${COLOR_INFO}${context}${COLOR_NONE}"
    fi
done
contexts=("${valid_context[@]}")

if [[ -z "${contexts[@]}" ]] ; then
    echoerr "no context(s) found in the kubeconfig file $kubeconfig!"
    exit 1
fi


###############################################################################
# Look for KubeStellar Helm chart
###############################################################################
echotitle "KubeStellar:"
helm_context=""
for context in "${contexts[@]}" ; do # for all contexts
    chart="$(helm --kubeconfig $kubeconfig --kube-context $context list 2> /dev/null | grep 'core-chart' || true)"
    if [[ "$chart" != "" ]] ; then
        helm_context=$context
        name="$(echo "$chart" | awk '{print $1}')" || true
        version="$(echo "$chart" | awk '{print $NF}')" || true
        secret="$(k --context $helm_context get secret --all-namespaces -l "owner=helm"  --no-headers -o name 2> /dev/null | grep "$name" || true)"
        secret="${secret##*/}"
        echo -e "- Helm chart ${COLOR_INFO}${name}${COLOR_NONE} (${COLOR_INFO}v${version}${COLOR_NONE}) in context ${COLOR_INFO}${context}${COLOR_NONE}"
        echo -e "  - Secret=${COLOR_INFO}${secret}${COLOR_NONE}"
        if [[ "$arg_yaml" == "true" ]] ; then
            mkdir -p "$output_folder/kubestellar-core-chart"
            echo -e "$(k --context $helm_context get secret $secret -o jsonpath='{.data.release}' | base64 -d | base64 -d | gzip -d | jq -r '.info'     | sed -e 's/\"/"/g')" > "$output_folder/kubestellar-core-chart/info.json"
            echo -e "$(k --context $helm_context get secret $secret -o jsonpath='{.data.release}' | base64 -d | base64 -d | gzip -d | jq -r '.chart'    | sed -e 's/\"/"/g')" > "$output_folder/kubestellar-core-chart/chart.json"
            echo -e "$(k --context $helm_context get secret $secret -o jsonpath='{.data.release}' | base64 -d | base64 -d | gzip -d | jq -r '.manifest' | sed -e 's/\"/"/g')" > "$output_folder/kubestellar-core-chart/manifest.yaml"
        fi
        break
    fi
done
if [[ "$helm_context" == "" ]] ; then
    echoerr "Helm chart was not found in any of the context(s)!"
    exit 1
fi


###############################################################################
# Look for Kubeflex deployment
###############################################################################
echotitle "KubeFlex:"
if [ -z "$(k --context $helm_context get ns kubeflex-system --no-headers -o name 2> /dev/null || true)" ] ; then
    echoerr "KubeFlex namespace not found!"
else
    echo -e "- ${COLOR_INFO}kubeflex-system${COLOR_NONE} namespace in context ${COLOR_INFO}$helm_context${COLOR_NONE}"
    kubeflex_pod=$(k --context $helm_context -n kubeflex-system get pod -l "control-plane=controller-manager" -o name 2> /dev/null | cut -d'/' -f2 || true)
    if [ -z "$kubeflex_pod" ]; then
        echoerr "KubeFlex pod not found!"
    else
        kubeflex_version=$(k --context $helm_context -n kubeflex-system get pod $kubeflex_pod -o json 2> /dev/null | jq -r '.spec.containers[] | select(.image | contains("kubestellar/kubeflex/manager")) | .image' | cut -d':' -f2 || true)
        kubeflex_status=$(k --context $helm_context -n kubeflex-system get pod $kubeflex_pod -o jsonpath='{.status.phase}' 2> /dev/null || true)
        echo -n -e "- ${COLOR_INFO}controller-manager${COLOR_NONE}: version=${COLOR_INFO}$kubeflex_version${COLOR_NONE}, pod=${COLOR_INFO}$kubeflex_pod${COLOR_NONE}, status="
        echostatus "$kubeflex_status"
        postgresql_pod=$(k --context $helm_context -n kubeflex-system get pod postgres-postgresql-0 2> /dev/null || true)
        if [ -z "$postgresql_pod" ]; then
            echoerr "postgres-postgresql-0 pod not found!"
        else
            postgresql_status=$(k --context $helm_context -n kubeflex-system get pod postgres-postgresql-0 -o jsonpath='{.status.phase}' 2> /dev/null || true)
            echo -n -e "- ${COLOR_INFO}postgres-postgresql-0${COLOR_NONE}: pod=${COLOR_INFO}postgres-postgresql-0${COLOR_NONE}, status="
            echostatus "$postgresql_status"
        fi
        if [[ "$arg_logs" == "true" ]] ; then
            mkdir -p "$output_folder/kubeflex"
            k --context $helm_context -n kubeflex-system logs $kubeflex_pod > "$output_folder/kubeflex/kubeflex-controller.log"
            k --context $helm_context -n kubeflex-system logs postgres-postgresql-0 > "$output_folder/kubeflex/postgresql.log"
        fi
    fi
fi


###############################################################################
# Listing Control Planes
###############################################################################
echotitle "Control Planes:"
cp_n=0
cps=($(k --context $helm_context get controlplanes -no-headers -o name 2> /dev/null || true))
for i in "${!cps[@]}" ; do # for all control planes in context ${context}
    name=${cps[i]##*/}
    cp_context[cp_n]=$helm_context
    cp_name[cp_n]=$name
    cp_ns[cp_n]="${cp_name[cp_n]}-system"
    cp_type[cp_n]=$(k --context $helm_context get controlplane ${cp_name[cp_n]} -o jsonpath='{.spec.type}')
    cp_pch[cp_n]=$(k --context $helm_context get controlplane ${cp_name[cp_n]} -o jsonpath='{.spec.postCreateHook}')
    cp_kubeconfig[cp_n]=$(get_kubeconfig "${helm_context}" "${cp_name[cp_n]}" "${cp_type[cp_n]}")
    echo -e "- ${COLOR_INFO}${cp_name[cp_n]}${COLOR_NONE}: type=${COLOR_INFO}${cp_type[cp_n]}${COLOR_NONE}, pch=${COLOR_INFO}${cp_pch[cp_n]}${COLOR_NONE}, context=${COLOR_INFO}${cp_context[cp_n]}${COLOR_NONE}, namespace=${COLOR_INFO}${cp_name[cp_n]}-system${COLOR_NONE}"
    if [[ "${cp_pch[cp_n]}" == "its" ]] ; then
        its_pod=$(k --context $helm_context -n "${cp_ns[cp_n]}" get pod -l "job-name=its" -o name 2> /dev/null | cut -d'/' -f2 || true)
        its_status=$(k --context $helm_context -n "${cp_ns[cp_n]}" get pod $its_pod -o jsonpath='{.status.phase}' 2> /dev/null || true)
        if [[ "${cp_type[cp_n]}" == "host" ]] ; then
            status_ns="open-cluster-management"
        else
            status_ns="${cp_ns[cp_n]}"
        fi
        status_pod=$(k --context $helm_context -n "$status_ns" get pod -o name 2> /dev/null | grep addon-status-controller | cut -d'/' -f2 || true)
        status_status=$(k --context $helm_context -n "$status_ns" get pod $status_pod -o jsonpath='{.status.phase}' 2> /dev/null || true)
        echo -n -e "  - Post Create Hook: pod=${COLOR_INFO}$its_pod${COLOR_NONE}, ns=${COLOR_INFO}${cp_ns[cp_n]}${COLOR_NONE}, status="
        echostatus "$its_status"
        echo -n -e "  - Status addon: pod=${COLOR_INFO}$status_pod${COLOR_NONE}, ns=${COLOR_INFO}$status_ns${COLOR_NONE}, status="
        echostatus "$status_status"
    else
        kubestellar_pod=$(k --context $helm_context -n "${cp_ns[cp_n]}" get pod -l "control-plane=controller-manager" -o name 2> /dev/null | cut -d'/' -f2 || true)
        kubestellar_version=$(k --context $helm_context -n "${cp_ns[cp_n]}" get pod $kubestellar_pod -o json 2> /dev/null | jq -r '.spec.containers[] | select(.image | contains("kubestellar/controller-manager")) | .image' | cut -d':' -f2 || true)
        kubestellar_status=$(k --context $helm_context -n "${cp_ns[cp_n]}" get pod $kubestellar_pod -o jsonpath='{.status.phase}' 2> /dev/null || true)
        echo -e -n "  - KubeStellar controller: version=${COLOR_INFO}$kubestellar_version${COLOR_NONE}, pod=${COLOR_INFO}$kubestellar_pod${COLOR_NONE} namespace=${COLOR_INFO}"${cp_ns[cp_n]}"${COLOR_NONE}, status="
        echostatus "$kubestellar_status"
        trasport_pod=$(k --context $helm_context -n "${cp_ns[cp_n]}" get pod -l "name=transport-controller" -o name 2> /dev/null | cut -d'/' -f2 || true)
        trasport_version=$(k --context $helm_context -n "${cp_ns[cp_n]}" get pod $trasport_pod -o json 2> /dev/null | jq -r '.spec.containers[] | select(.image | contains("kubestellar/ocm-transport-controller")) | .image' | cut -d':' -f2 || true)
        trasport_status=$(k --context $helm_context -n "${cp_ns[cp_n]}" get pod $trasport_pod -o jsonpath='{.status.phase}' 2> /dev/null || true)
        echo -e -n "  - Transport controller: version=${COLOR_INFO}$trasport_version${COLOR_NONE}, pod=${COLOR_INFO}$trasport_pod${COLOR_NONE} namespace=${COLOR_INFO}${cp_ns[cp_n]}${COLOR_NONE}, status="
        echostatus "$trasport_status"
    fi
    if [[ "$arg_yaml" == "true" ]] ; then
        mkdir -p "$output_folder/$name"
        k --context $helm_context get controlplane $name -o yaml > "$output_folder/$name/cp.yaml"
        if [[ "${cp_pch[cp_n]}" == "its" ]] ; then
            k --context $helm_context -n "${cp_ns[cp_n]}" get pod $its_pod -o yaml > "$output_folder/$name/its-job.yaml"
            k --context $helm_context -n "$status_ns" get pod $status_pod -o yaml > "$output_folder/$name/status-addon.yaml"
        else
            k --context $helm_context -n "${cp_ns[cp_n]}" get pod $kubestellar_pod -o yaml > "$output_folder/$name/kubestellar-controller.yaml"
            k --context $helm_context -n "${cp_ns[cp_n]}" get pod $trasport_pod -o yaml > "$output_folder/$name/transport-controller.yaml"
        fi
    fi
    if [[ "$arg_logs" == "true" ]] ; then
        mkdir -p "$output_folder/$name"
        if [[ "${cp_pch[cp_n]}" == "its" ]] ; then
            k --context $helm_context -n "${cp_ns[cp_n]}" logs $its_pod -c its-clusteradm > "$output_folder/$name/its-job-clusteradm.log"
            k --context $helm_context -n "${cp_ns[cp_n]}" logs $its_pod -c its-statusaddon > "$output_folder/$name/its-job-status-addon.log"
            k --context $helm_context -n "$status_ns" logs $status_pod -c status-controller > "$output_folder/$name/status-addon.log"
        else
            k --context $helm_context -n "${cp_ns[cp_n]}" logs $kubestellar_pod > "$output_folder/$name/kubestellar-controller.log"
            k --context $helm_context -n "${cp_ns[cp_n]}" logs $trasport_pod -c transport-controller > "$output_folder/$name/transport-controller.log"
        fi
    fi
    cp_n=$((cp_n+1))
done


###############################################################################
# Listing managed clusters
###############################################################################
echotitle "Managed Clusters:"
mc_n=0
for j in "${!cp_pch[@]}" ; do
    if [[ "${cp_pch[$j]}" == "its" ]] ; then
        mcs=($(kubectl --kubeconfig <(echo "${cp_kubeconfig[$j]}") get managedcluster -no-headers -o name 2> /dev/null || true))
        for i in "${!mcs[@]}" ; do
            name=${mcs[i]##*/}
            mc_name[mc_n]=$name
            accepted="$(kubectl --kubeconfig <(echo "${cp_kubeconfig[$j]}") get managedcluster $name -o jsonpath='{.status.conditions}' | jq '.[] | select(.type == "HubAcceptedManagedCluster") | .status' | tr -d '"')"
            joined="$(kubectl --kubeconfig <(echo "${cp_kubeconfig[$j]}") get managedcluster $name -o jsonpath='{.status.conditions}' | jq '.[] | select(.type == "ManagedClusterJoined") | .status' | tr -d '"')"
            available="$(kubectl --kubeconfig <(echo "${cp_kubeconfig[$j]}") get managedcluster $name -o jsonpath='{.status.conditions}' | jq '.[] | select(.type == "ManagedClusterConditionAvailable") | .status' | tr -d '"')"
            synced="$(kubectl --kubeconfig <(echo "${cp_kubeconfig[$j]}") get managedcluster $name -o jsonpath='{.status.conditions}' | jq '.[] | select(.type == "ManagedClusterConditionClockSynced") | .status' | tr -d '"')"
            echo -e "- ${COLOR_INFO}${mc_name[mc_n]}${COLOR_NONE} in ${COLOR_INFO}${cp_name[j]}${COLOR_NONE}: accepted=$(echostatus $accepted), joined=$(echostatus $joined), available=$(echostatus $available), synced=$(echostatus $synced)"
            if [[ "$arg_verbose" == "true" ]] ; then
                echo -n -e "${COLOR_YAML}"
                kubectl --kubeconfig <(echo "${cp_kubeconfig[$j]}") get managedclusters $name -o jsonpath='{.metadata.labels}' | jq '. |= with_entries(select(.key|(contains("open-cluster-management")|not)))' | indent
                echo -n -e "${COLOR_NONE}"
            fi
            if [[ "$arg_yaml" == "true" ]] ; then
                mkdir -p "$output_folder/${cp_name[$j]}/managed-clusters"
                kubectl --kubeconfig <(echo "${cp_kubeconfig[$j]}") get managedcluster $name -o yaml > "$output_folder/${cp_name[$j]}/managed-clusters/$name.yaml"
            fi
            mc_n=$((mc_n+1))
        done
    fi
done


###############################################################################
# Listing binding policies
###############################################################################
echotitle "Binding Policies:"
bp_n=0
for j in "${!cp_pch[@]}" ; do
    if [[ "${cp_pch[$j]}" == "wds" ]] ; then
        bps=($(kubectl --kubeconfig <(echo "${cp_kubeconfig[$j]}") get bindingpolicy -no-headers -o name 2> /dev/null || true))
        for i in "${!bps[@]}" ; do
            name=${bps[i]##*/}
            bp_cp[bp_n]="${cp_name[$j]}"
            bp_name[bp_n]=$name
            echo -e "- ${COLOR_INFO}${bp_name[bp_n]}${COLOR_NONE} in control plane ${COLOR_INFO}${bp_cp[bp_n]}${COLOR_NONE}"
            if [[ "$arg_verbose" == "true" ]] ; then
                echo -n -e "${COLOR_YAML}"
                kubectl --kubeconfig <(echo "${cp_kubeconfig[$j]}") get bindingpolicy ${bp_name[bp_n]} -o jsonpath='{.spec}' | jq '.' | indent || true
                echo -n -e "${COLOR_NONE}"
            fi
            if [[ "$arg_yaml" == "true" ]] ; then
                mkdir -p "$output_folder/${bp_cp[bp_n]}/binding-politcies"
                kubectl --kubeconfig <(echo "${cp_kubeconfig[$j]}") get bindingpolicy ${bp_name[bp_n]} -o yaml > "$output_folder/${bp_cp[bp_n]}/binding-politcies/$name.yaml"
            fi
            bp_n=$((bp_n+1))
        done
    fi
done


###############################################################################
# Listing Manifest Works
###############################################################################
echotitle "Manifest Works:"
mw_n=0
for h in "${!cp_pch[@]}" ; do
    if [[ "${cp_pch[$h]}" == "its" ]] ; then
        ns=($(kubectl --kubeconfig <(echo "${cp_kubeconfig[$h]}") get ns -no-headers -o name 2> /dev/null || true))
        for j in "${!ns[@]}" ; do
            cluster=${ns[j]##*/}
            mws=($(kubectl --kubeconfig <(echo "${cp_kubeconfig[$h]}") --namespace $cluster get manifestwork --no-headers -o name 2> /dev/null || true))
            for i in "${!mws[@]}" ; do
                name="${mws[i]##*/}"
                origin="$(kubectl --kubeconfig <(echo "${cp_kubeconfig[$h]}") --namespace $cluster get manifestwork $name -o jsonpath='{.metadata.labels.transport\.kubestellar\.io\/originWdsName}' 2> /dev/null || true)"
                if [[ "$origin" == "" ]] ; then
                    continue
                fi
                mw_cp[mw_n]="${cp_name[$h]}"
                mw_name[mw_n]="$name"
                mw_cluster[mw_n]="$cluster"
                mw_origin[mw_n]="$origin"
                mw_binding[mw_n]="$(kubectl --kubeconfig <(echo "${cp_kubeconfig[$h]}") --namespace $cluster get manifestwork ${mw_name[mw_n]} -o jsonpath='{.metadata.labels.transport\.kubestellar\.io\/originOwnerReferenceBindingKey}')"
                echo -e "- ${COLOR_INFO}${mw_name[mw_n]}${COLOR_NONE} in cp=${COLOR_INFO}${mw_cp[mw_n]}${COLOR_NONE}, namespace=${COLOR_INFO}${mw_cluster[mw_n]}${COLOR_NONE}: ${COLOR_INFO}${mw_origin[mw_n]}${COLOR_NONE} --> ${COLOR_INFO}${mw_binding[mw_n]}${COLOR_NONE} --> ${COLOR_INFO}${mw_cluster[mw_n]}${COLOR_NONE}"
                if [[ "$arg_verbose" == "true" ]] ; then
                    echo -n -e "${COLOR_YAML}"
                    kubectl --kubeconfig <(echo "${cp_kubeconfig[$h]}") --namespace $cluster get manifestwork $name -o jsonpath='{.spec.workload.manifests}' | jq '.[] | {"apiVersion", "kind", "metadata"} | (.name = .metadata.name) | del(.metadata)' | indent || true
                    echo -n -e "${COLOR_NONE}"
                fi
                if [[ "$arg_yaml" == "true" ]] ; then
                    mkdir -p "$output_folder/${cp_name[$h]}/manifest-works/$cluster"
                    kubectl --kubeconfig <(echo "${cp_kubeconfig[$h]}") --namespace $cluster get manifestwork $name -o yaml > "$output_folder/${cp_name[$h]}/manifest-works/$cluster/$name.yaml"
                fi
                mw_n=$((mw_n+1))
            done
        done
    fi
done


###############################################################################
# Listing Work Statuses
###############################################################################
echotitle "Work Statuses:"
sw_n=0
for h in "${!cp_pch[@]}" ; do
    if [[ "${cp_pch[$h]}" == "its" ]] ; then
        ns=($(kubectl --kubeconfig <(echo "${cp_kubeconfig[$h]}") get ns -no-headers -o name 2> /dev/null || true))
        for j in "${!ns[@]}" ; do
            cluster=${ns[j]##*/}
            sws=($(kubectl --kubeconfig <(echo "${cp_kubeconfig[$h]}") --namespace $cluster get workstatuses --no-headers -o name 2> /dev/null || true))
            for i in "${!sws[@]}" ; do
                name="${sws[i]##*/}"
                origin="$(kubectl --kubeconfig <(echo "${cp_kubeconfig[$h]}") --namespace $cluster get workstatuses $name -o jsonpath='{.metadata.labels.transport\.kubestellar\.io\/originWdsName}' 2> /dev/null || true)"
                if [[ "$origin" == "" ]] ; then
                    continue
                fi
                sw_cp[sw_n]="${cp_name[$h]}"
                sw_name[sw_n]="$name"
                sw_cluster[sw_n]="$cluster"
                sw_origin[sw_n]="$origin"
                sw_binding[sw_n]="$(kubectl --kubeconfig <(echo "${cp_kubeconfig[$h]}") --namespace $cluster get workstatuses ${sw_name[sw_n]} -o jsonpath='{.metadata.labels.transport\.kubestellar\.io\/originOwnerReferenceBindingKey}')"
                echo -n -e "- ${COLOR_INFO}${sw_name[sw_n]}${COLOR_NONE} in cp=${COLOR_INFO}${sw_cp[sw_n]}${COLOR_NONE}, namespace=${COLOR_INFO}${sw_cluster[sw_n]}${COLOR_NONE}, status="
                echo -n $(echostatus $(kubectl --kubeconfig <(echo "${cp_kubeconfig[$h]}") --namespace $cluster get workstatuses $name -o jsonpath='{.status.phase}' || true))
                echo -e ": ${COLOR_INFO}${sw_cluster[sw_n]}${COLOR_NONE} --> ${COLOR_INFO}${sw_binding[sw_n]}${COLOR_NONE} --> ${COLOR_INFO}${sw_origin[sw_n]}${COLOR_NONE}"
                if [[ "$arg_verbose" == "true" ]] ; then
                    echo -n -e "${COLOR_YAML}"
                    kubectl --kubeconfig <(echo "${cp_kubeconfig[$h]}") --namespace $cluster get workstatuses $name -o jsonpath='{.spec.sourceRef}' | jq '.' | indent || true
                    echo -n -e "${COLOR_NONE}"
                fi
                if [[ "$arg_yaml" == "true" ]] ; then
                    mkdir -p "$output_folder/${cp_name[$h]}/work-statuses/$cluster"
                    kubectl --kubeconfig <(echo "${cp_kubeconfig[$h]}") --namespace $cluster get workstatuses $name -o yaml > "$output_folder/${cp_name[$h]}/work-statuses/$cluster/$name.yaml"
                fi
                sw_n=$((sw_n+1))
            done
        done
    fi
done


###############################################################################
# Create archive
###############################################################################
if [[ "$arg_logs" == "true" || "$arg_yaml" == "true" ]] ; then
    echov -e "\nSaving logs and/or YAML to ${COLOR_INFO}./kubestellar-snapshot.tar.gz${COLOR_NONE}"
    tar czf kubestellar-snapshot.tar.gz -C "$output_folder" .
fi
