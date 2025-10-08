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

# NOTE WELL: Two copies of this file exist, one in kubestellar/hack/
# and one in kubestellar/scripts/ . Keep them both up-to-date.
BASE_URL="https://docs.kubestellar.io"
VERSION="release-0.29.0-alpha.1"
INSTALLATION_ERROR_URL="${BASE_URL}/${VERSION}/direct/installation-errors#pod-errors-due-to-too-many-open-files"

set -e # exit on error

echoerr() {
    echo "$@" 1>&2;
}

is_installed() {
    # $1 == name
    # $2 == command name to search
    # $3 == command to get the version, unstructured
    # $4 == command to get the version for lexicographic comparison
    # $5 == help
    # $6 == min required version
    # (optional) $7 == another, numerically higher lexically lower, min version (covering case of roll from 9 to 10)
    # (optional) $8 == exclusive upper bound on version (required if $7 given)
    wantver="$6"
    addlver="$7"
    exclver="$8"
    #Checks if helm version installed is compatiable
    local INSTALLED_VERSION
    INSTALLED_VERSION=$(helm version --template={{.Version}} 2>/dev/null)
    local MIN_HELM_VERSION="v3.10.0" 
    if [ "$(printf '%s\n' "$MIN_HELM_VERSION" "$INSTALLED_VERSION" | sort -V | head -n1)" != "$MIN_HELM_VERSION" ]; then
        echo -e "\033[0;31mX\033[0m Helm version $INSTALLED_VERSION is less than required $MIN_HELM_VERSION"
        [ "$assert" == "true" ] && exit 2
    else
        echo -e "\033[0;32m✔\033[0m Helm version $INSTALLED_VERSION meets minimum requirement $MIN_HELM_VERSION"
    fi
    if which $2 > /dev/null ; then
        if [ $# -gt 2 ]; then
            gotver=$(eval "$4" 2> /dev/null)
            echo -e "\033[0;32m\xE2\x9C\x94\033[0m $1 ($gotver)"
            echov "  version (unstructured): $(eval "$3" 2> /dev/null)"
            if ! [[ "$gotver" < "$wantver" ]]; then : OK, easy case
            elif [ -n "$exclver" ] && [[ "$gotver" < "$exclver" ]] && ! [[ "$gotver" < "$addlver" ]]; then : OK, hard case
            else
                echo -e "\033[0;31mX\033[0m  structured version '$gotver' is less than required minimum '$wantver'" $([ -z "$addlver" ] || echo or "'$addlver' but less than '$exclver'") >&2
                exit 2
            fi
        else
            echo -e "\033[0;32m\xE2\x9C\x94\033[0m $1"
        fi
        echov "     path: $(which $2)"
    else
        echo -e "\033[0;31mX\033[0m $1"
        if [ $# -gt 3 ]; then
            echov "  how to install: $5; get at least version $wantver"
        fi
        if [ "$assert" == "true" ]; then
            exit 2
        fi
    fi
}

check_sysctl() {
    name="$1"
    minval="$2"
    # because the OS running this script might not be the Linux that runs the containers, use `docker run` to interrogate that Linux
    if report=$(docker run --rm busybox sysctl "$name"); then
        val=$(cut -d= -f2 <<<$report | awk '{ print $1 }')
        if [[ $val -ge "$minval" ]]; then
            echo -e "\033[0;32m\xE2\x9C\x94\033[0m $name is $val"
        else
            echo -e "\033[0;31mX\033[0m sysctl ${name} is only $val but must be at least $minval (see https://kind.sigs.k8s.io/docs/user/known-issues#pod-errors-due-to-too-many-open-files)" >&2
             echo -e "\033[0;31mX\033[0m sysctl ${name} is only $val but must be at least $minval (see $INSTALLATION_ERROR_URL#pod-errors-due-to-too-many-open-files)" >&2 
            [ "$assert" == "true" ] && exit 4
        fi
    else
        echo -e "\033[0;31mX\033[0m Failed to extract sysctl ${name}" >&2
        [ "$assert" == "true" ] &&  exit 3
    fi
}

is_installed_argo() {
    is_installed 'ArgoCD CLI' \
        'argocd' \
        'argocd version --short --client' \
        'argocd version --short --client -o json | jq -r .client.argocd' \
        'https://argo-cd.readthedocs.io/en/stable/cli_installation/' \
        v2
}

is_installed_brew() {
    is_installed 'Home Brew' \
        'brew' \
        'brew --version' \
        'brew --version | grep "^Homebrew "' \
        '/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"' \
        'Homebrew 4'
}

is_installed_docker() {
    is_installed 'Docker' \
        'docker' \
        'docker --version' \
        'docker --version' \
        'https://docs.docker.com/engine/install/' \
        'Docker version 20'
}

is_installed_go() {
    is_installed 'Go' \
        'go' \
        'go version' \
        'go version' \
        'https://go.dev/doc/install' \
        'go version 1.23'
}

is_installed_helm() {
    local log
    is_installed 'Helm' \
        'helm' \
        'helm version' \
        'helm version --template={{.Version}}' \
        'https://helm.sh/docs/intro/install/' \
        'v3'
    if log=$(helm show chart oci://ghcr.io/kubestellar/kubestellar/core-chart 2>&1)
    then echo -e "\033[0;32m\xE2\x9C\x94\033[0m helm can fetch public charts"
    else
        echo "$log" >&2
        echo >&2
        echo -e "\033[0;31mX\033[0m helm can not fetch public charts!" >&2
        [ "$assert" == "true" ] && exit 5
    fi
    
}

is_installed_jq() {
    is_installed 'jq' \
        'jq' \
        'jq --version' \
        'jq --version' \
        'https://jqlang.github.io/jq/download/' \
        'jq-1.5'
}

is_installed_kcp() {
    is_installed 'kcp' \
        'kcp' \
        'kcp version' \
        'echo dontcare' \
        'https://docs.kcp.io/kcp/main/' \
        'dontcare'
}

is_installed_kflex() {
    is_installed 'KubeFlex' \
        'kflex' \
        'kflex version | grep "^Kubeflex version:"' \
        'kflex version | grep "^Kubeflex version:"' \
        'https://github.com/kubestellar/kubeflex' \
        'Kubeflex version: v0.8.0'
}

is_installed_kind() {
    is_installed 'Kind' \
        'kind' \
        'kind version' \
        'kind version' \
        'https://kind.sigs.k8s.io/docs/user/quick-start/#installation' \
        'kind v0.20'
    check_sysctl fs.inotify.max_user_watches 524288
    check_sysctl fs.inotify.max_user_instances 512
}

is_installed_k3d() {
    is_installed 'k3d' \
        'k3d' \
        'k3d --version' \
        'k3d --version' \
        'https://k3d.io/#installation' \
        'k3d version v5.7.4'
}
is_installed_ko() {
    is_installed 'KO' \
        'ko' \
        'ko version' \
        'ko version' \
        'https://ko.build/install/' \
        '0.15'
}

is_installed_kubectl() {
    is_installed 'kubectl' \
        'kubectl' \
        'kubectl version --client | head -1' \
        "kubectl version --client -o json 2> /dev/null | grep gitVersion | cut '-d\"' -f4" \
        'https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/' \
        v1.29
}

is_installed_make() {
    is_installed 'GNU Make' \
        'make' \
        'make --version | head -1' \
        'make --version | grep "^GNU Make "' \
        'sudo apt-get install make or brew install make or install XCode Command Line Tools' \
        'GNU Make 3.5'
}

is_installed_ocm() {
    is_installed 'OCM CLI' \
        'clusteradm' \
        'clusteradm version 2> /dev/null | grep ^client' \
        "clusteradm version 2> /dev/null | grep ^client | awk '"'{ print $3 }'"'" \
        'bash <(curl -L https://raw.githubusercontent.com/open-cluster-management-io/clusteradm/main/install.sh) 0.10.1' \
        :v0.10 \
        :v0.10 \
        :v0.11
}

is_installed_yq() {
    is_installed 'yq' \
        'yq' \
        'yq --version' \
        "yq --version | sed 's/.*\(version .*\)$/\1/'" \
        'brew install yq or snap install yq' \
        'version v4'
}


# Global constants
PROGRAMS="@(argo|brew|docker|go|helm|jq|kflex|kind|ko|kubectl|make|ocm|yq|k3d)"

# Global parameters
assert="false"  # true => exit on missing program
list="false"    # true => display the list of programs and exit
verbose="false" # true => display verbose information about the programs: version, path, install info
programs=()

# Parse command line arguments
shopt -s extglob
while (( $# > 0 )); do
    case "$1" in
    (--assert|-A)
        assert="true";;
    (--list|-L)
        list="true";;
    (--verbose|-V)
        verbose="true";;
    (-X)
        set -x;;
    (--help|-help|-h)
        echo "Usage: $0 [-A|--assert] [-L|--list] [-V|--verbose] [-X] [program1] [program2] ..."
        exit 0;;
    ($PROGRAMS)
        programs+=("$1");;
    (-*)
        echoerr "$0: unknown flag \"$1\""
        exit 1;;
    (*)
        echoerr "$0: unknown positional argument \"$1\""
        exit 1;;
    esac
    shift
done

# Define the echov function based on verbosity
if [ "$verbose" == "true" ]; then
    echov() { echo "$@" ; }
else
    echov() { :; }
fi

# Dsiplay the list of programs, if requested
if [ "$list" == "true" ]; then
    IFS='@|()' read -r -a programs <<< "$(echo "$PROGRAMS" | sed -e "s/^@(//" -e "s/)//")"
    echo "${programs[@]}"
    exit 0
fi

if [ ${#programs[@]} -eq 0 ]; then
    echo "Checking pre-requisites for using KubeStellar:"
    is_installed_docker
    is_installed_kubectl
    is_installed_kflex
    is_installed_ocm
    is_installed_helm
    echo "Checking additional pre-requisites for running the examples:"
    is_installed_kind
    is_installed_k3d
    is_installed_argo
    echo "Checking pre-requisites for building KubeStellar:"
    is_installed_make
    is_installed_go
    is_installed_ko
else
    echov "Checking pre-requisites for KubeStellar:"
    for p in ${programs[@]} ; do
        eval "is_installed_$p"
    done
fi
