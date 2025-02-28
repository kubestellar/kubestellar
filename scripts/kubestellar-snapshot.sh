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
TIMESTAMP="$(date +%F_%T)"
TMPFOLDER="$(mktemp -d -p . "kubestellar-snapshot-XXXX")"
OUTPUT_FOLDER="$TMPFOLDER/kubestellar-snapshot"


# Script info
SCRIPT_NAME="KubeStellar Snapshot"
SCRIPT_VERSION="0.3.0"


# Colors
COLOR_NONE="\033[0m"
COLOR_RED="\033[1;31m"
COLOR_GREEN="\033[1;32m"
COLOR_BLUE="\033[94m"
COLOR_YELLOW="\033[1;33m"
COLOR_PURPLE="\033[1;35m"
COLOR_WARNING="${COLOR_RED}"
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
Usage: $0 [options]

Options:
--kubeconfig|-K <filename> use the specified kubeconfig to find KubeStellar Helm chart
--context|-C <name>        use the specified context to find KubeStellar Helm chart
--logs|-L                  save the logs of the pods
--yaml|-Y                  save the YAML of the resources
--verbose|-V               output extra information
--version|-v               print out the script version
--help|-h                  show this information
-X                         enable verbose execution for debugging

Note 1: This script expects that KubeStellar was installed with its Helm chart and
        tries to find the latter in a context of the relevant kubeconfig.

Note 2: After piping the output of the script to file use "more"
        to inspect the file content in color. Alternatively install
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
    context="$1" # context in the relevant kubeconfig
    cp_name="$2" # name of the Control Plane
    cp_type="$3" # type of the Control Plane

    # check if the CP is ready
    if [[ $(kubectl --context $context get controlplane $cp_name -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; then
        echo ""
        return
    fi

    # put into the shell variable "kubeconfig" the kubeconfig contents for use from outside of the hosting cluster
    if [[ "$cp_type" == "host" ]] ; then
        kubeconfig=$(kubectl --context $context config view --flatten --minify)
    else
        # determine the secret name and namespace
        if [[ "$cp_type" == "external" ]] ; then
            key=$(kubectl --context $context get controlplane $cp_name -o=jsonpath='{.status.secretRef.inClusterKey}')
        else
            key=$(kubectl --context $context get controlplane $cp_name -o=jsonpath='{.status.secretRef.key}')
        fi
        secret_name=$(kubectl --context $context get controlplane $cp_name -o=jsonpath='{.status.secretRef.name}')
        secret_namespace=$(kubectl --context $context get controlplane $cp_name -o=jsonpath='{.status.secretRef.namespace}')
        # get the kubeconfig
        kubeconfig=$(kubectl --context $context get secret $secret_name -n $secret_namespace -o=jsonpath="{.data.$key}" | base64 -d)
    fi

    # return the kubeconfig file contents in YAML format
    echo "$kubeconfig"
}


# List all CRDs of interest
list_crds() {
    modifier="$1" # kubeconfig modifier
    pattern="$2" # matching pattern
    indent="$3" # indentation

    echo "$indent- CRDs:"

    kubectl $modifier get crds --no-headers -o name | grep "$pattern" | while read -r crd; do
        echo -n -e "$indent  - ${COLOR_INFO}${crd##*/}${COLOR_NONE}: established="
        echostatus $(kubectl $modifier get $crd -o jsonpath='{.status.conditions[?(@.type=="Established")].status}')
    done
}


# This function is called when the script exists normally or on error
on_exit() {
    echov ""

    # Create archive
    if [[ "$arg_logs" == "true" || "$arg_yaml" == "true" ]] ; then
        echov -e "Saving logs and/or YAML from ${COLOR_INFO}$OUTPUT_FOLDER${COLOR_NONE} to ${COLOR_INFO}./kubestellar-snapshot.tar.gz${COLOR_NONE}"
        tar czf kubestellar-snapshot.tar.gz -C "$OUTPUT_FOLDER" .
    fi

    # Cleaning up
    echov -e "Removing temporary folder: ${COLOR_INFO}$TMPFOLDER${COLOR_NONE}"
    rm -rf "$TMPFOLDER"
}


trap on_exit EXIT


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
        echo "$0: unknown flag \"$1\"" >&2
        exit 1;;
    (*)
        echo "$0: unknown positional argument \"$1\"" >&2
        exit 1;;
    esac
    shift
done


###############################################################################
# Alias definitions
###############################################################################
# Define the echov function based on verbosity
if [ "$arg_verbose" == "true" ]; then
    echov() { echo "$@" ; }
else
    echov() { :; }
fi


###############################################################################
# Script info
###############################################################################
echov -e "${COLOR_INFO}${SCRIPT_NAME} v${SCRIPT_VERSION}${COLOR_NONE}\n"
echov -e "Script run on ${COLOR_INFO}$TIMESTAMP${COLOR_NONE}"


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

is_installed 'yq' \
    'yq' \
    'yq --version'


###############################################################################
# Ensure output folder
###############################################################################
if [[ "$arg_logs" == "true" || "$arg_yaml" == "true" ]] ; then
    echov -e "Creating temporary folder: ${COLOR_INFO}$TMPFOLDER${COLOR_NONE}"
    mkdir -p "$OUTPUT_FOLDER"
fi


###############################################################################
# Determine the list of kubeconfigs to search
# based on https://kubernetes.io/docs/reference/k/generated/k_config/
###############################################################################
[[ -n "$arg_kubeconfig" ]] && export KUBECONFIG="$arg_kubeconfig"
[[ -z "$KUBECONFIG" ]] && export KUBECONFIG="$HOME/.kube/config"
echov -e "Using kubeconfig(s): ${COLOR_INFO}$KUBECONFIG${COLOR_NONE}"


###############################################################################
# Determine the list of contexts to search
###############################################################################
if [[ "$arg_context" == "" ]] ; then
    contexts=($(kubectl config get-contexts --no-headers -o name))
else
    contexts=("$arg_context")
fi

if [[ -z "${contexts[@]}" ]] ; then
    echoerr "No context(s) found in the kubeconfig file $KUBECONFIG!"
    exit 1
fi

echov "Validating contexts(s): "
CURRENT_CONTEXT="$(kubectl config current-context || true)"
valid_contexts=()
for context in "${contexts[@]}" ; do # for all contexts
    [[ -z "$context" ]] && continue
    [[ "$context" == "$CURRENT_CONTEXT" ]] && cc="*" || cc=""
    if kubectl --context $context get secrets -A > /dev/null 2>&1 ; then
        echov -e "${COLOR_GREEN}\xE2\x9C\x94${COLOR_NONE} ${COLOR_INFO}${context}${COLOR_NONE} ${COLOR_GREEN}$cc${COLOR_NONE}"
        [[ -z "$cc" ]] && valid_contexts+=("$context") || valid_contexts=("$context" "${valid_contexts[@]}")
    else
        echov -e "${COLOR_RED}X${COLOR_NONE} ${COLOR_INFO}${context}${COLOR_NONE} ${COLOR_GREEN}$cc${COLOR_NONE}"
    fi
done

if [[ -z "${valid_contexts[@]}" ]] ; then
    echoerr "No valid context(s) found in the kubeconfig file $KUBECONFIG!"
    exit 1
fi


###############################################################################
# Look for KubeStellar Helm chart
###############################################################################
echotitle "KubeStellar:"
helm_context=""
more_than_one="false"
for context in "${valid_contexts[@]}" ; do
    while IFS= read -r line; do
        [[ -z "$line" ]] && continue
        name="$(echo "$line" | awk '{print $1}')"
        namespace="$(echo "$line" | awk '{print $2}')"
        version="$(echo "$line" | awk '{print $NF}')"
        secret="$(kubectl --context $context get secret -n "$namespace" -l "owner=helm,name=$name" --no-headers -o name 2> /dev/null || true)"
        secret="${secret##*/}"
        if [[ ! -z "$(kubectl --context $context get secret $secret -n "$namespace" -o jsonpath='{.data.release}' | base64 -d | base64 -d | gzip -d | jq -r '.chart.metadata.description' | grep -i "KubeStellar" 2> /dev/null || true)" ]] ; then
            echo -e "- Helm chart ${COLOR_INFO}${name}${COLOR_NONE} (${COLOR_INFO}v${version}${COLOR_NONE}) in namespace ${COLOR_INFO}${namespace}${COLOR_NONE} in context ${COLOR_INFO}${context}${COLOR_NONE}"
            echo -e "  - Secret=${COLOR_INFO}${secret}${COLOR_NONE} in namespace ${COLOR_INFO}${namespace}${COLOR_NONE}"
            if [[ -z "$helm_context" ]] ; then
                helm_context="$context"
                if [[ "$arg_yaml" == "true" ]] ; then
                    mkdir -p "$OUTPUT_FOLDER/kubestellar-core-chart"
                    kubectl --context $helm_context get secret $secret -o jsonpath='{.data.release}' | base64 -d | base64 -d | gzip -d | jq -r '.info'      > "$OUTPUT_FOLDER/kubestellar-core-chart/info.json"
                    kubectl --context $helm_context get secret $secret -o jsonpath='{.data.release}' | base64 -d | base64 -d | gzip -d | jq -r '.chart'     > "$OUTPUT_FOLDER/kubestellar-core-chart/chart.json"
                    kubectl --context $helm_context get secret $secret -o jsonpath='{.data.release}' | base64 -d | base64 -d | gzip -d | jq -r '.manifest'  > "$OUTPUT_FOLDER/kubestellar-core-chart/manifest.yaml"
                fi
            else
                more_than_one="true"
            fi
        fi
    done <<< "$(helm --kube-context $context list --no-headers -A 2> /dev/null || true)"
done
if [[ -z "$helm_context" ]] ; then
    echoerr "KubeStellar Helm chart was not found in any of the context(s)!"
    exit 1
elif [[ "$more_than_one" == "true" ]] ; then
    echo -e "${COLOR_WARNING}WARNING: found more than one Helm chart for KubeStellar... using the first one!${COLOR_NONE}"

fi


###############################################################################
# Look for Kubeflex deployment
###############################################################################
echotitle "KubeFlex:"
if [[ -z "$(kubectl --context $helm_context get ns kubeflex-system --no-headers 2> /dev/null || true)" ]] ; then
    echoerr "KubeFlex namespace not found!"
else
    echo -e "- ${COLOR_INFO}kubeflex-system${COLOR_NONE} namespace in context ${COLOR_INFO}$helm_context${COLOR_NONE}"
    kubeflex_pod=$(kubectl --context $helm_context -n kubeflex-system get pod -l "control-plane=controller-manager" -o name 2> /dev/null | cut -d'/' -f2 || true)
    if [[ -z "$kubeflex_pod" ]]; then
        echoerr "KubeFlex pod not found!"
    else
        kubeflex_version=$(kubectl --context $helm_context -n kubeflex-system get pod $kubeflex_pod -o json 2> /dev/null | jq -r '.spec.containers[] | select(.image | contains("kubestellar/kubeflex/manager")) | .image' | cut -d':' -f2 || true)
        kubeflex_status=$(kubectl --context $helm_context -n kubeflex-system get pod $kubeflex_pod -o jsonpath='{.status.phase}' 2> /dev/null || true)
        echo -n -e "- ${COLOR_INFO}controller-manager${COLOR_NONE}: version=${COLOR_INFO}$kubeflex_version${COLOR_NONE}, pod=${COLOR_INFO}$kubeflex_pod${COLOR_NONE}, status="
        echostatus "$kubeflex_status"
        postgresql_pod=$(kubectl --context $helm_context -n kubeflex-system get pod postgres-postgresql-0 2> /dev/null || true)
        if [ -z "$postgresql_pod" ]; then
            echoerr "postgres-postgresql-0 pod not found!"
            postgresql_install_pod=$(kubectl --context $helm_context -n kubeflex-system get pod --no-headers -o name | grep "install-postgres" | head -1 | cut -d'/' -f2 2> /dev/null || true)
            if [ -n "$postgresql_install_pod" ]; then
                echo -e "Found at least one ${COLOR_INFO}postgresql${COLOR_NONE} installation pod that did not complete: ${COLOR_INFO}$postgresql_install_pod${COLOR_NONE}"
                if kubectl --context $helm_context -n kubeflex-system logs $postgresql_install_pod | grep "toomanyrequests" ; then
                    echoerr "there may be an issue pulling the postgresql image from Docker Hub."
                fi
            fi
        else
            postgresql_status=$(kubectl --context $helm_context -n kubeflex-system get pod postgres-postgresql-0 -o jsonpath='{.status.phase}' 2> /dev/null || true)
            echo -n -e "- ${COLOR_INFO}postgres-postgresql-0${COLOR_NONE}: pod=${COLOR_INFO}postgres-postgresql-0${COLOR_NONE}, status="
            echostatus "$postgresql_status"
        fi
        list_crds "--context $helm_context" "kflex" ""
        if [[ "$arg_logs" == "true" ]] ; then
            mkdir -p "$OUTPUT_FOLDER/kubeflex"
            [ -n "$kubeflex_pod" ] && kubectl --context $helm_context -n kubeflex-system logs $kubeflex_pod &> "$OUTPUT_FOLDER/kubeflex/kubeflex-controller.log" || true
            [ -n "$postgresql_pod" ] && kubectl --context $helm_context -n kubeflex-system logs postgres-postgresql-0 &> "$OUTPUT_FOLDER/kubeflex/postgresql.log" || true
        fi
    fi
fi


###############################################################################
# Listing Control Planes
###############################################################################
echotitle "Control Planes:"
cp_n=0
cps=($(kubectl --context $helm_context get controlplanes -no-headers -o name 2> /dev/null || true))
for i in "${!cps[@]}" ; do # for all control planes in context ${context}
    name=${cps[i]##*/}
    cp_context[cp_n]=$helm_context
    cp_name[cp_n]=$name
    cp_ns[cp_n]="${cp_name[cp_n]}-system"
    cp_type[cp_n]=$(kubectl --context $helm_context get controlplane ${cp_name[cp_n]} -o jsonpath='{.spec.type}')
    cp_pch[cp_n]=$(kubectl --context $helm_context get controlplane ${cp_name[cp_n]} -o jsonpath='{.spec.postCreateHook}')
    cp_kubeconfig_content=$(get_kubeconfig "${helm_context}" "${cp_name[cp_n]}" "${cp_type[cp_n]}")
    echo -e "- ${COLOR_INFO}${cp_name[cp_n]}${COLOR_NONE}: type=${COLOR_INFO}${cp_type[cp_n]}${COLOR_NONE}, pch=${COLOR_INFO}${cp_pch[cp_n]}${COLOR_NONE}, context=${COLOR_INFO}${cp_context[cp_n]}${COLOR_NONE}, namespace=${COLOR_INFO}${cp_name[cp_n]}-system${COLOR_NONE}"
    if [[ -z "$cp_kubeconfig_content" ]] ; then
        cp_kubeconfig[cp_n]=""
        echo -e "  ${COLOR_WARNING}WARNING: the Control Plane is not ready, the kubeconfig is not available!${COLOR_NONE}"
    else
        cp_kubeconfig[cp_n]="$TMPFOLDER/$name-kubeconfig"
        echo "$cp_kubeconfig_content" > "${cp_kubeconfig[cp_n]}"
    fi
    if [[ "${cp_pch[cp_n]}" =~ ^its ]] ; then
        its_pod=$(kubectl --context $helm_context -n "${cp_ns[cp_n]}" get pod -l "job-name=${cp_pch[cp_n]}" -o name 2> /dev/null | cut -d'/' -f2 || true)
        its_status=$(kubectl --context $helm_context -n "${cp_ns[cp_n]}" get pod $its_pod -o jsonpath='{.status.phase}' 2> /dev/null || true)
        if [[ "${cp_type[cp_n]}" != "vcluster" ]] ; then
            status_ns="open-cluster-management"
        else
            status_ns="${cp_ns[cp_n]}"
        fi
        if [[ "${cp_type[cp_n]}" == "external" ]] ; then
            if ! KUBECONFIG=${cp_kubeconfig[cp_n]} kubectl get secrets -A > /dev/null 2>&1 ; then
                cp_cluster="$(KUBECONFIG=${cp_kubeconfig[cp_n]} kubectl config view --minify | yq '.contexts[0].context.cluster')"
                for context in "${valid_contexts[@]}" ; do
                    if [[ "$(kubectl --context $context config view --minify | yq '.contexts[0].context.cluster')" == "$cp_cluster" ]] ; then
                        echo "$(kubectl --context $context config view --minify --flatten)" > "${cp_kubeconfig[cp_n]}"
                        break
                    fi
                done
            fi
            status_pod=$(KUBECONFIG=${cp_kubeconfig[cp_n]} kubectl -n "$status_ns" get pod -o name 2> /dev/null | grep addon-status-controller | cut -d'/' -f2 || true)
            status_version=$(KUBECONFIG=${cp_kubeconfig[cp_n]} kubectl -n "$status_ns" get pod $status_pod -o jsonpath='{.spec.containers}' | jq -r '.[].image | select(contains("status-addon"))' | cut -d: -f2 || true)
            status_status=$(KUBECONFIG=${cp_kubeconfig[cp_n]} kubectl -n "$status_ns" get pod $status_pod -o jsonpath='{.status.phase}' 2> /dev/null || true)
        else
            status_pod=$(kubectl --context $helm_context -n "$status_ns" get pod -o name 2> /dev/null | grep addon-status-controller | cut -d'/' -f2 || true)
            status_version=$(kubectl --context $helm_context -n "$status_ns" get pod $status_pod -o jsonpath='{.spec.containers}' | jq -r '.[].image | select(contains("status-addon"))' | cut -d: -f2 || true)
            status_status=$(kubectl --context $helm_context -n "$status_ns" get pod $status_pod -o jsonpath='{.status.phase}' 2> /dev/null || true)
        fi
        echo -n -e "  - Post Create Hook: pod=${COLOR_INFO}$its_pod${COLOR_NONE}, ns=${COLOR_INFO}${cp_ns[cp_n]}${COLOR_NONE}, status="
        echostatus "$its_status"
        echo -n -e "  - Status addon controller: pod=${COLOR_INFO}$status_pod${COLOR_NONE}, ns=${COLOR_INFO}$status_ns${COLOR_NONE}, version=${COLOR_INFO}$status_version${COLOR_NONE}, status="
        echostatus "$status_status"
        if [ -n "${cp_kubeconfig[cp_n]}" ]; then
            ocm_version=$(KUBECONFIG=${cp_kubeconfig[cp_n]} kubectl get deploy -n open-cluster-management cluster-manager -o jsonpath={.spec.template.spec.containers[0].image} | cut -d: -f2)
            if [ -n "$ocm_version" ]; then
                echo -e "  - Open-cluster-manager: version=${COLOR_INFO}$ocm_version${COLOR_NONE}"
            else
                echo -e "  - Open-cluster-manager: ${COLOR_ERROR}not found${COLOR_NONE}"
            fi
            list_crds "--kubeconfig ${cp_kubeconfig[cp_n]}" "kubestellar\|open-cluster-management" "  "
        fi
    else
        kubestellar_pod=$(kubectl --context $helm_context -n "${cp_ns[cp_n]}" get pod -l "control-plane=controller-manager" -o name 2> /dev/null | cut -d'/' -f2 || true)
        kubestellar_version=$(kubectl --context $helm_context -n "${cp_ns[cp_n]}" get pod $kubestellar_pod -o json 2> /dev/null | jq -r '.spec.containers[] | select(.image | contains("kubestellar/controller-manager")) | .image' | cut -d':' -f2 || true)
        kubestellar_status=$(kubectl --context $helm_context -n "${cp_ns[cp_n]}" get pod $kubestellar_pod -o jsonpath='{.status.phase}' 2> /dev/null || true)
        echo -e -n "  - KubeStellar controller manager: version=${COLOR_INFO}$kubestellar_version${COLOR_NONE}, pod=${COLOR_INFO}$kubestellar_pod${COLOR_NONE} namespace=${COLOR_INFO}"${cp_ns[cp_n]}"${COLOR_NONE}, status="
        echostatus "$kubestellar_status"
        trasport_pod=$(kubectl --context $helm_context -n "${cp_ns[cp_n]}" get pod -l "name=transport-controller" -o name 2> /dev/null | cut -d'/' -f2 || true)
        trasport_version=$(kubectl --context $helm_context -n "${cp_ns[cp_n]}" get pod $trasport_pod -o json 2> /dev/null | jq -r '.spec.containers[] | select(.image | contains("kubestellar/ocm-transport-controller")) | .image' | cut -d':' -f2 || true)
        trasport_status=$(kubectl --context $helm_context -n "${cp_ns[cp_n]}" get pod $trasport_pod -o jsonpath='{.status.phase}' 2> /dev/null || true)
        echo -e -n "  - Transport controller: version=${COLOR_INFO}$trasport_version${COLOR_NONE}, pod=${COLOR_INFO}$trasport_pod${COLOR_NONE} namespace=${COLOR_INFO}${cp_ns[cp_n]}${COLOR_NONE}, status="
        echostatus "$trasport_status"
        if [ -n "${cp_kubeconfig[cp_n]}" ]; then
            list_crds "--kubeconfig ${cp_kubeconfig[cp_n]}" "kubestellar" "  "
        fi
    fi
    if [[ "$arg_yaml" == "true" ]] ; then
        mkdir -p "$OUTPUT_FOLDER/$name"
        kubectl --context $helm_context get controlplane $name -o yaml > "$OUTPUT_FOLDER/$name/cp.yaml"
        if [[ "${cp_pch[cp_n]}" =~ ^its ]] ; then
            kubectl --context $helm_context -n "${cp_ns[cp_n]}" get pod $its_pod -o yaml > "$OUTPUT_FOLDER/$name/its-job.yaml"
            if [[ "${cp_type[cp_n]}" == "external" ]] ; then
                KUBECONFIG=${cp_kubeconfig[cp_n]} kubectl -n "$status_ns" get pod $status_pod -o yaml > "$OUTPUT_FOLDER/$name/status-addon.yaml"
            else
                kubectl --context $helm_context -n "$status_ns" get pod $status_pod -o yaml > "$OUTPUT_FOLDER/$name/status-addon.yaml"
            fi
        else
            kubectl --context $helm_context -n "${cp_ns[cp_n]}" get pod $kubestellar_pod -o yaml > "$OUTPUT_FOLDER/$name/kubestellar-controller.yaml"
            kubectl --context $helm_context -n "${cp_ns[cp_n]}" get pod $trasport_pod -o yaml > "$OUTPUT_FOLDER/$name/transport-controller.yaml"
        fi
    fi
    if [[ "$arg_logs" == "true" ]] ; then
        mkdir -p "$OUTPUT_FOLDER/$name"
        if [[ "${cp_pch[cp_n]}" =~ ^its ]] ; then
            if [ -n "$its_pod" ] ; then
                containers=$(kubectl --context $helm_context -n "${cp_ns[cp_n]}" get pod $its_pod -o jsonpath='{.spec.containers[*].name}')
                for ctr in $containers; do
                    { kubectl --context $helm_context -n "${cp_ns[cp_n]}" logs $its_pod -c "$ctr" || true; } &> "$OUTPUT_FOLDER/$name/its-job-${ctr}.log"
                done
            fi
            if [[ -n "$status_pod" ]] ; then
                if [[ "${cp_type[cp_n]}" == "external" ]] ; then
                    KUBECONFIG=${cp_kubeconfig[cp_n]} kubectl -n "$status_ns" logs $status_pod -c status-controller &> "$OUTPUT_FOLDER/$name/status-addon.log" || true
                else
                    kubectl --context $helm_context -n "$status_ns" logs $status_pod -c status-controller &> "$OUTPUT_FOLDER/$name/status-addon.log" || true
                fi
            fi
        else
            [ -n "$kubestellar_pod" ] && kubectl --context $helm_context -n "${cp_ns[cp_n]}" logs $kubestellar_pod &> "$OUTPUT_FOLDER/$name/kubestellar-controller.log" || true
            [ -n "$trasport_pod" ] && kubectl --context $helm_context -n "${cp_ns[cp_n]}" logs $trasport_pod -c transport-controller &> "$OUTPUT_FOLDER/$name/transport-controller.log" || true
        fi
    fi
    cp_n=$((cp_n+1))
done


###############################################################################
# Listing Managed Clusters
###############################################################################
echotitle "Managed Clusters:"
mc_n=0
for j in "${!cp_pch[@]}" ; do
    if [[ "${cp_pch[$j]}" =~ ^its ]] ; then
        mcs=($(KUBECONFIG="${cp_kubeconfig[$j]}" kubectl get managedcluster -no-headers -o name 2> /dev/null || true))
        for i in "${!mcs[@]}" ; do
            name=${mcs[i]##*/}
            mc_name[mc_n]=$name
            accepted="$(KUBECONFIG="${cp_kubeconfig[$j]}" kubectl get managedcluster $name -o jsonpath='{.status.conditions}' | jq '.[] | select(.type == "HubAcceptedManagedCluster") | .status' | tr -d '"')"
            joined="$(KUBECONFIG="${cp_kubeconfig[$j]}" kubectl get managedcluster $name -o jsonpath='{.status.conditions}' | jq '.[] | select(.type == "ManagedClusterJoined") | .status' | tr -d '"')"
            available="$(KUBECONFIG="${cp_kubeconfig[$j]}" kubectl get managedcluster $name -o jsonpath='{.status.conditions}' | jq '.[] | select(.type == "ManagedClusterConditionAvailable") | .status' | tr -d '"')"
            synced="$(KUBECONFIG="${cp_kubeconfig[$j]}" kubectl get managedcluster $name -o jsonpath='{.status.conditions}' | jq '.[] | select(.type == "ManagedClusterConditionClockSynced") | .status' | tr -d '"')"
            echo -e "- ${COLOR_INFO}${mc_name[mc_n]}${COLOR_NONE} in ${COLOR_INFO}${cp_name[j]}${COLOR_NONE}: accepted=$(echostatus $accepted), joined=$(echostatus $joined), available=$(echostatus $available), synced=$(echostatus $synced)"
            if [[ "$arg_verbose" == "true" ]] ; then
                echo -n -e "${COLOR_YAML}"
                KUBECONFIG="${cp_kubeconfig[$j]}" kubectl get managedclusters $name -o jsonpath='{.metadata.labels}' | jq '. |= with_entries(select(.key|(contains("open-cluster-management")|not)))' | indent
                echo -n -e "${COLOR_NONE}"
            fi
            if [[ "$arg_yaml" == "true" ]] ; then
                mkdir -p "$OUTPUT_FOLDER/${cp_name[$j]}/managed-clusters"
                KUBECONFIG="${cp_kubeconfig[$j]}" kubectl get managedcluster $name -o yaml > "$OUTPUT_FOLDER/${cp_name[$j]}/managed-clusters/$name.yaml"
            fi
            mc_n=$((mc_n+1))
        done
    fi
done


###############################################################################
# Listing Binding Policies
###############################################################################
echotitle "Binding Policies:"
bp_n=0
for j in "${!cp_pch[@]}" ; do
    if [[ "${cp_pch[$j]}" == "wds" ]] ; then
        bps=($(KUBECONFIG="${cp_kubeconfig[$j]}" kubectl get bindingpolicy -no-headers -o name 2> /dev/null || true))
        for i in "${!bps[@]}" ; do
            name=${bps[i]##*/}
            bp_cp[bp_n]="${cp_name[$j]}"
            bp_name[bp_n]=$name
            echo -e "- ${COLOR_INFO}${bp_name[bp_n]}${COLOR_NONE} in control plane ${COLOR_INFO}${bp_cp[bp_n]}${COLOR_NONE}"
            if [[ "$arg_verbose" == "true" ]] ; then
                echo -n -e "${COLOR_YAML}"
                KUBECONFIG="${cp_kubeconfig[$j]}" kubectl get bindingpolicy ${bp_name[bp_n]} -o jsonpath='{.spec}' | jq '.' | indent || true
                echo -n -e "${COLOR_NONE}"
            fi
            if [[ "$arg_yaml" == "true" ]] ; then
                mkdir -p "$OUTPUT_FOLDER/${bp_cp[bp_n]}/binding-policies"
                KUBECONFIG="${cp_kubeconfig[$j]}" kubectl get bindingpolicy ${bp_name[bp_n]} -o yaml > "$OUTPUT_FOLDER/${bp_cp[bp_n]}/binding-policies/$name.yaml"
            fi
            bp_n=$((bp_n+1))
        done
    fi
done


###############################################################################
# Listing Bindings
###############################################################################
echotitle "Bindings:"
for j in "${!cp_pch[@]}" ; do
    if ! [[ "${cp_pch[$j]}" == "wds" ]] ; then continue; fi
    bindings=($(KUBECONFIG="${cp_kubeconfig[$j]}" kubectl get bindings.control.kubestellar.io -no-headers -o name 2> /dev/null || true))
    for i in "${!bindings[@]}" ; do
        binding_name=${bindings[i]##*/}
        binding_cp="${cp_name[$j]}"
        spec=$(KUBECONFIG="${cp_kubeconfig[$j]}" kubectl get bindings.control.kubestellar.io "$binding_name" -o jsonpath='{.spec}')
        dests=$(jq '.destinations//[]' <<<$spec)
        cobjs=$(jq '.workload.clusterScope//[]' <<<$spec)
        nobjs=$(jq '.workload.namespaceScope//[]' <<<$spec)
        clisting=$(jq -r --argjson dests "$dests" --argjson objs "$cobjs" '[$dests, $objs] | combinations | .[0].clusterId+" "+.[1].resource+"."+.[1].group+"/"+.[1].version+" "+.[1].name' <<<0)
        [ -n "$clisting" ] && { echo "$clisting" | while read wec rsc name; do
            echo -e "- wds=${COLOR_INFO}${binding_cp}${COLOR_NONE} binding=${COLOR_INFO}${binding_name}${COLOR_NONE} WEC=${COLOR_INFO}${wec}${COLOR_NONE} rsc=${COLOR_INFO}${rsc}${COLOR_NONE} name=${COLOR_INFO}${name}${COLOR_NONE}"
        done; }
        nlisting=$(jq -r --argjson dests "$dests" --argjson objs "$nobjs" '[$dests, $objs] | combinations | .[0].clusterId+" "+.[1].resource+"."+.[1].group+"/"+.[1].version+" "+.[1].namespace+" "+.[1].name' <<<0)
        [ -n "$nlisting" ] && { echo "$nlisting" | while read wec rsc ns name; do
            echo -e "- wds=${COLOR_INFO}${binding_cp}${COLOR_NONE} binding=${COLOR_INFO}${binding_name}${COLOR_NONE} WEC=${COLOR_INFO}${wec}${COLOR_NONE} rsc=${COLOR_INFO}${rsc}${COLOR_NONE} ns=${COLOR_INFO}${ns}${COLOR_NONE} name=${COLOR_INFO}${name}${COLOR_NONE}"
        done; }
        if [[ -z "$clisting" ]] && [[ -z "$nlisting" ]]; then
            ndests=$(jq length <<<$dests <<<0)
	    ncobjs=$(jq length <<<$cobjs <<<0)
            nnobjs=$(jq length <<<$nobjs <<<0)
            [ "$ndests" = 0 ] && colord=$COLOR_WARNING || colord=$COLOR_INFO
            [ "$ncobjs" = 0 ] && colorc=$COLOR_WARNING || colorc=$COLOR_INFO
            [ "$nnobjs" = 0 ] && colorn=$COLOR_WARNING || colorn=$COLOR_INFO
            echo -e "- wds=${COLOR_INFO}${binding_cp}${COLOR_NONE} binding=${COLOR_INFO}${binding_name}${COLOR_NONE} num_wecs=${colord}${ndests}${COLOR_NONE} num_clusterScoped_objs=${colorc}${ncobjs}${COLOR_NONE} num_namespaced_objs=${colorn}${nnobjs}${COLOR_NONE}"
        fi
        if [[ "$arg_yaml" == "true" ]] ; then
            mkdir -p "$OUTPUT_FOLDER/${binding_cp}/bindings"
            KUBECONFIG="${cp_kubeconfig[$j]}" kubectl get binding.control.kubestellar.io ${binding_name} -o yaml > "$OUTPUT_FOLDER/${binding_cp}/bindings/$binding_name.yaml"
        fi
    done
done


###############################################################################
# Listing Manifest Works
###############################################################################
echotitle "Manifest Works:"
mw_n=0
for h in "${!cp_pch[@]}" ; do
    if [[ "${cp_pch[$h]}" =~ ^its ]] ; then
        ns=($(KUBECONFIG="${cp_kubeconfig[$h]}" kubectl get ns -no-headers -o name 2> /dev/null || true))
        for j in "${!ns[@]}" ; do
            cluster=${ns[j]##*/}
            mws=($(KUBECONFIG="${cp_kubeconfig[$h]}" kubectl --namespace $cluster get manifestwork --no-headers -o name 2> /dev/null || true))
            for i in "${!mws[@]}" ; do
                name="${mws[i]##*/}"
                origin="$(KUBECONFIG="${cp_kubeconfig[$h]}" kubectl --namespace $cluster get manifestwork $name -o jsonpath='{.metadata.labels.transport\.kubestellar\.io\/originWdsName}' 2> /dev/null || true)"
                if [[ "$origin" == "" ]] ; then
                    continue
                fi
                mw_cp[mw_n]="${cp_name[$h]}"
                mw_name[mw_n]="$name"
                mw_cluster[mw_n]="$cluster"
                mw_origin[mw_n]="$origin"
                mw_binding[mw_n]="$(KUBECONFIG="${cp_kubeconfig[$h]}" kubectl --namespace $cluster get manifestwork ${mw_name[mw_n]} -o jsonpath='{.metadata.labels.transport\.kubestellar\.io\/originOwnerReferenceBindingKey}')"
                echo -e "- ${COLOR_INFO}${mw_name[mw_n]}${COLOR_NONE} in cp=${COLOR_INFO}${mw_cp[mw_n]}${COLOR_NONE}, namespace=${COLOR_INFO}${mw_cluster[mw_n]}${COLOR_NONE}: ${COLOR_INFO}${mw_origin[mw_n]}${COLOR_NONE} --> ${COLOR_INFO}${mw_binding[mw_n]}${COLOR_NONE} --> ${COLOR_INFO}${mw_cluster[mw_n]}${COLOR_NONE}"
                if [[ "$arg_verbose" == "true" ]] ; then
                    echo -n -e "${COLOR_YAML}"
                    KUBECONFIG="${cp_kubeconfig[$h]}" kubectl --namespace $cluster get manifestwork $name -o jsonpath='{.spec.workload.manifests}' | jq '.[] | {"apiVersion", "kind", "metadata"} | (.name = .metadata.name) | del(.metadata)' | indent || true
                    echo -n -e "${COLOR_NONE}"
                fi
                if [[ "$arg_yaml" == "true" ]] ; then
                    mkdir -p "$OUTPUT_FOLDER/${cp_name[$h]}/manifest-works/$cluster"
                    KUBECONFIG="${cp_kubeconfig[$h]}" kubectl --namespace $cluster get manifestwork $name -o yaml > "$OUTPUT_FOLDER/${cp_name[$h]}/manifest-works/$cluster/$name.yaml"
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
    if [[ "${cp_pch[$h]}" =~ ^its ]] ; then
        ns=($(KUBECONFIG="${cp_kubeconfig[$h]}" kubectl get ns -no-headers -o name 2> /dev/null || true))
        for j in "${!ns[@]}" ; do
            cluster=${ns[j]##*/}
            sws=($(KUBECONFIG="${cp_kubeconfig[$h]}" kubectl --namespace $cluster get workstatuses --no-headers -o name 2> /dev/null || true))
            for i in "${!sws[@]}" ; do
                name="${sws[i]##*/}"
                origin="$(KUBECONFIG="${cp_kubeconfig[$h]}" kubectl --namespace $cluster get workstatuses $name -o jsonpath='{.metadata.labels.transport\.kubestellar\.io\/originWdsName}' 2> /dev/null || true)"
                if [[ "$origin" == "" ]] ; then
                    continue
                fi
                sw_cp[sw_n]="${cp_name[$h]}"
                sw_name[sw_n]="$name"
                sw_cluster[sw_n]="$cluster"
                sw_origin[sw_n]="$origin"
                sw_binding[sw_n]="$(KUBECONFIG="${cp_kubeconfig[$h]}" kubectl --namespace $cluster get workstatuses ${sw_name[sw_n]} -o jsonpath='{.metadata.labels.transport\.kubestellar\.io\/originOwnerReferenceBindingKey}')"
                echo -n -e "- ${COLOR_INFO}${sw_name[sw_n]}${COLOR_NONE} in cp=${COLOR_INFO}${sw_cp[sw_n]}${COLOR_NONE}, namespace=${COLOR_INFO}${sw_cluster[sw_n]}${COLOR_NONE}, status="
                echo -n $(echostatus $(KUBECONFIG="${cp_kubeconfig[$h]}" kubectl --namespace $cluster get workstatuses $name -o jsonpath='{.status.phase}' || true))
                echo -e ": ${COLOR_INFO}${sw_cluster[sw_n]}${COLOR_NONE} --> ${COLOR_INFO}${sw_binding[sw_n]}${COLOR_NONE} --> ${COLOR_INFO}${sw_origin[sw_n]}${COLOR_NONE}"
                if [[ "$arg_verbose" == "true" ]] ; then
                    echo -n -e "${COLOR_YAML}"
                    KUBECONFIG="${cp_kubeconfig[$h]}" kubectl --namespace $cluster get workstatuses $name -o jsonpath='{.spec.sourceRef}' | jq '.' | indent || true
                    echo -n -e "${COLOR_NONE}"
                fi
                if [[ "$arg_yaml" == "true" ]] ; then
                    mkdir -p "$OUTPUT_FOLDER/${cp_name[$h]}/work-statuses/$cluster"
                    KUBECONFIG="${cp_kubeconfig[$h]}" kubectl --namespace $cluster get workstatuses $name -o yaml > "$OUTPUT_FOLDER/${cp_name[$h]}/work-statuses/$cluster/$name.yaml"
                fi
                sw_n=$((sw_n+1))
            done
        done
    fi
done
