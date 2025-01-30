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


set -o errexit


TMPFOLDER="$(mktemp -d -p . "kubestellar-XXXX")"
trap 'rm -rf "$TMPFOLDER"' EXIT
BOOTSTRAP_KUBECONFIG="$TMPFOLDER/bootstrap-config"


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
arg_cp=""
arg_kubeconfig="$HOME/.kube/config"
arg_context=""
arg_ns="default"
arg_addr=""
arg_verbose=false


# Display help
display_help() {
  cat << EOF
Usage: $0 [options]

--controlplane|-c <name>   control plane name used to name the secret: <name>-bootstrap
--namespace|-n <name>      namespace name where to create the secret, default is "default"
--kubeconfig|-K <filename> use the specified kubeconfig
--context|-C <name>        use the specified context
--address|-A <addr>        specify a replacement internal address for the cluster
--verbose|-V               output extra information
--help|-h                  show this information
-X                         enable verbose execution for debugging
EOF
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


###############################################################################
# Parse command line arguments
###############################################################################
while (( $# > 0 )); do
    case "$1" in
    (--address|-A)
        if (( $# > 1 ));
        then { arg_addr="$2"; shift; }
        else { echo "$0: missing address value" >&2; exit 1; }
        fi;;
    (--namespace|-n)
        if (( $# > 1 ));
        then { arg_ns="$2"; shift; }
        else { echo "$0: missing namespace name" >&2; exit 1; }
        fi;;
    (--controlplane|-c)
        if (( $# > 1 ));
        then { arg_cp="$2"; shift; }
        else { echo "$0: missing control plane name" >&2; exit 1; }
        fi;;
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
    (--verbose|-V)
        arg_verbose=true;;
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
# Check arguments
###############################################################################
if [[ -z "$arg_cp" ]] ; then
    echoerr "a control plane name is required!"
    $0 --help
    exit 1
fi
if [[ -z "$arg_context" ]] ; then
    contexts=($(kubectl --kubeconfig "$arg_kubeconfig" config get-contexts --no-headers -o name 2> /dev/null))
    case ${#contexts[@]} in
    (0)
        echoerr "there are no contexts in the kubeconfig file!"
        $0 --help
        exit 1;;
    (1)
        arg_context="${contexts[0]}";;
    (*)
        for context in "${contexts[@]}" ; do # for all contexts
            if [[ "$context" =~ "$arg_cp" ]] ; then
                arg_context="$context"
                break
            fi
        done
        if [[ -z "$arg_context" ]] ; then
            echoerr "there are multiple contexts in the kubeconfig file, specify one with \"--context\"!"
            $0 --help
            exit 1
        fi;;
    esac
fi


###############################################################################
# Extract the external context
###############################################################################
[[ $arg_verbose ]] && echo -e "Extracting context ${COLOR_YELLOW}$arg_context${COLOR_NONE} from kubeconfig ${COLOR_YELLOW}$arg_kubeconfig${COLOR_NONE}..."
kubectl --kubeconfig=$arg_kubeconfig config view --minify --flatten --context=$arg_context > $BOOTSTRAP_KUBECONFIG


###############################################################################
# Replace address if necessary
###############################################################################
if [[ -n "$arg_addr" ]] ; then
    [[ $arg_verbose ]] && echo -e "Setting server internal address to ${COLOR_YELLOW}$arg_addr${COLOR_NONE}..."
    kubectl --kubeconfig=$BOOTSTRAP_KUBECONFIG config set-cluster $(kubectl --kubeconfig=$BOOTSTRAP_KUBECONFIG config current-context) --server=$arg_addr > /dev/null
fi


###############################################################################
# Create secret
###############################################################################
create_ns=true
for ns in $(kubectl get ns --no-headers -o name) ; do
    if [[ "namespace/${arg_ns}" == "$ns" ]] ; then
        create_ns=false
        break
    fi
done
if [[ $create_ns ]] ; then
    [[ $arg_verbose ]] && echo -e "Creating namespace ${COLOR_YELLOW}${arg_ns}${COLOR_NONE}..."
    kubectl create ns "$arg_ns"
fi
[[ $arg_verbose ]] && echo -e "Creating secret ${COLOR_YELLOW}${arg_cp}-bootstrap${COLOR_NONE} in namespace ${COLOR_YELLOW}${arg_ns}${COLOR_NONE}..."
kubectl create secret generic ${arg_cp}-bootstrap --from-file=kubeconfig-incluster=$BOOTSTRAP_KUBECONFIG --namespace $arg_ns
