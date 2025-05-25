#!/bin/bash

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

ctx="cluster1"
monitoring_ns="ks-monitoring"

while [ $# != 0 ]; do
    case "$1" in
        (-h|--help) echo "$0 usage: ( -h | --help | --wec-cluster-context (e.g., --wec-cluster-context cluster1 (default value: cluster1)) | --monitoring-ns (e.g., --monitoring-ns ks-monitoring (default value: ks-monitoring)))*"
                    exit;;
        (--wec-cluster-context)
          if (( $# > 1 )); then
            ctx="$2"
            shift
          else
            echo "Missing wec-cluster-context value" >&2
            exit 1;
          fi;;
        (--monitoring-ns)
          if (( $# > 1 )); then
            monitoring_ns="$2"
            shift
          else
            echo "Missing monitoring-ns value" >&2
            exit 1;
          fi;;
    esac
    shift
done

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)

# set wec context
kubectl config use-context $ctx

: --------------------------------------------------------------------
: Configure Status Agent-Addon pod for Prometheus scraping
: --------------------------------------------------------------------
: 1. Create status agent-addon service:
kubectl -n open-cluster-management-agent-addon apply -f ${SCRIPT_DIR}/configuration/status-agent-svc.yaml

: 2. Create service monitor for status agent-addon
sed s/%WEC%/$ctx/g ${SCRIPT_DIR}/configuration/status-agent-sm.yaml | kubectl -n $monitoring_ns apply -f -