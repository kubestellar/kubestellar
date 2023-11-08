#!/usr/bin/env bash

# Copyright 2021 The KubeStellar Authors.
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

#################################################################################

# Get a kubeflex release, untar it, and use kflex to install and initialize kflex.
#################################################################################


while (( $# > 0 )); do
    case "$1" in
    (--kubeconfig)
        if (( $# > 1 ));
        then { config="$2"; shift; }
        else { echo "$0: kubeconfig file" >&2; exit 1; }
        fi;;
    (-h|--help)
        echo "Usage: $0 [--kubeconfig]"
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

wget https://github.com/kubestellar/kubeflex/releases/download/v0.2.5/kubeflex_0.2.5_linux_amd64.tar.gz
mkdir kubeflex
tar xf kubeflex_0.2.5_linux_amd64.tar.gz -C kubeflex
kubeflex/bin/kflex --kubeconfig $config init
rm kubeflex_0.2.5_linux_amd64.tar.gz

