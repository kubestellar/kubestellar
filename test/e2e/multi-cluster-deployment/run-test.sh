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

while [ $# != 0 ]; do
    case "$1" in
        (-h|--help) echo "$0 usage: (--released | --kubestellar-controller-manager-verbosity \$num | --transport-controller-verbosity \$num)*"
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
        (*) echo "$0: unrecognized argument '$1'" >&2
            exit 1
    esac
    shift
done

set -e # exit on error

SRC_DIR="$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)"
COMMON_SRCS="${SRC_DIR}/../common"
HACK_DIR="${SRC_DIR}/../../../hack"

"${HACK_DIR}/check_pre_req.sh" --assert --verbose kubectl docker kind make go ko yq helm kflex ocm

"${COMMON_SRCS}/cleanup.sh"
source "${COMMON_SRCS}/setup-shell.sh"
"${COMMON_SRCS}/setup-kubestellar.sh" $setup_flags
"${SRC_DIR}/use-kubestellar.sh"
