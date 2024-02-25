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

set -x # echo so that users can understand what is happening
set -e # exit on error
SRC_DIR="$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)"

:
: -------------------------------------------------------------------------
: "Cleaning up from previous run of an e2e test"

kind delete cluster --name cluster1
kind delete cluster --name cluster2
kind delete cluster --name kubeflex
kubectl config delete-context cluster1 || true
kubectl config delete-context cluster2 || true

## as temp solution transport runs as executable. cleanup of the created artifacts.
cd ${SRC_DIR}/../../.. ## go up KubeStellar directory
rm -f transport.log
ocm_transport_plugin_process="$(pgrep ocm-transport-plugin || true)"
if [[ ! -z "$ocm_transport_plugin_process" ]]; 
then 
    kill ${ocm_transport_plugin_process}
fi
rm -f ocm-transport-plugin
