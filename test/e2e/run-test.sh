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

# This is an end to end test of multi cluster deployement.
# For readable instructions, please visit ../../../docs/content/direct

set -x # so users can see what is going on

env="kind"
test="bash"
fail_flag=""


while [ $# != 0 ]; do
    case "$1" in
        (-h|--help) echo "$0 usage: (--released | --env | --test-type | --kubestellar-controller-manager-verbosity \$num | --transport-controller-verbosity \$num | --fail-fast)*"
                    exit;;
        (--released) setup_flags="$setup_flags $1";;
        (--kubestellar-controller-manager-verbosity|--transport-controller-verbosity)
          if (( $# > 1 )); then
            setup_flags="$setup_flags $1 $2"
            shift
          else
            echo "Missing $1 value" >&2
            exit 1;
          fi;;
        (--env)
          if (( $# > 1 )); then
            env="$2"
            shift
          else
            echo "Missing environment value" >&2
            exit 1;
          fi;;
        (--test-type)
          if (( $# > 1 )); then
            test="$2"
            shift
          else
            echo "Missing test type value" >&2
            exit 1;
          fi;;
        (--fail-fast) fail_flag="--fail-fast";;
        (*) echo "$0: unrecognized argument '$1'" >&2
            exit 1
    esac
    shift
done

case "$env" in
    (kind|ocp) ;;
    (*) echo "$0: --env must be 'kind' or 'ocp'" >&2
        exit 1;;
esac

case "$test" in
    (bash|ginkgo|argocd-ginkpo) ;;
    (*) echo "$0: --test-type must be 'bash' , 'ginkgo' , 'argocd-ginkpo'" >&2
        exit 1;;
esac

set -e # exit on error

SRC_DIR="$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)"
COMMON_SRCS="${SRC_DIR}/common"
scripts_dir="${SRC_DIR}/../../scripts"

"${scripts_dir}/check_pre_req.sh" --assert --verbose kubectl docker kind make go ko yq helm kflex ocm

"${COMMON_SRCS}/cleanup.sh" --env "$env"
source "${COMMON_SRCS}/setup-shell.sh"
if [ "$test" == "argocd-ginkgo" ]; then
    setup_flags="$setup_flags --argocd"
fi
"${COMMON_SRCS}/setup-kubestellar.sh" $setup_flags --env "$env"

if [ $test == "bash" ];then
    "${SRC_DIR}/bash/use-kubestellar.sh" --env "$env"
elif [ $test == "ginkgo" ];then
    GINKGO_DIR="${SRC_DIR}/ginkgo"
    KFLEX_DISABLE_CHATTY=true ginkgo --vv --trace --no-color $fail_flag $GINKGO_DIR -- -skip-setup
     else
    echo "$0: unknown test type '$test'"
    exit 1
fi
