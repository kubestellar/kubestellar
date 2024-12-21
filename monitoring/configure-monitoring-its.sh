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

ctx="kind-kubeflex"
its="its1"
monitoring_ns="ks-monitoring"

while [ $# != 0 ]; do
    case "$1" in
        (-h|--help) echo "$0 usage: ( -h | --help | --kubeflex-hosting-cluster-context (e.g., --kubeflex-hosting-cluster-context core-cluster (default value: kind-kubeflex)) | --space-name (e.g., --space-name its1 (default value: its1)) | --monitoring-ns (e.g., --monitoring-ns ks-monitoring (default value: ks-monitoring)))*"
                    exit;;
        (--kubeflex-hosting-cluster-context)
          if (( $# > 1 )); then
            ctx="$2"
            shift
          else
            echo "Missing host-cluster-context value" >&2
            exit 1;
          fi;;
        (--space-name)
          if (( $# > 1 )); then
            its="$2"
            shift
          else
            echo "Missing space-name value" >&2
            exit 1;
          fi;;
        (--monitoring-ns)
          if (( $# > 1 )); then
            its="$2"
            shift
          else
            echo "Missing monitoring-ns value" >&2
            exit 1;
          fi;;
    esac
    shift
done

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)

# set core context
kubectl config use-context $ctx

: --------------------------------------------------------------------
: Configure addon-status controller pod for prometheus scraping
: --------------------------------------------------------------------
pod_name=$(kubectl -n $its-system get pods -o=jsonpath='{range .items..metadata}{.name}{"\n"}{end}' | grep addon-status-controller-*)
pod_label=$(kubectl -n $its-system get pod $pod_name -o jsonpath='{.metadata.labels}' | jq | grep "status-controller" | tr "," " ")

: 1. Create addon-status controller service:
sed "s^%STATUS_CTL_LABEL%^$pod_label^g" ${SCRIPT_DIR}/configuration/status-addon-ctl-svc.yaml | kubectl -n $its-system apply -f -

: 2. Create the service monitor:
sed s/%ITS%/$its/g ${SCRIPT_DIR}/configuration/status-addon-ctl-sm.yaml | kubectl -n $monitoring_ns apply -f -

: --------------------------------------------------------------------
: Configure ITS API server pod for prometheus scraping
: --------------------------------------------------------------------
: 1. Create a SA and give the right RBAC to talk to the ITS API server
kubectl --context $its get ns $monitoring_ns || kubectl --context $its create ns $monitoring_ns
kubectl --context $its -n $monitoring_ns apply -f ${SCRIPT_DIR}/prometheus/prometheus-rbac.yaml

: 2. Create token secret for prometheus in the target ITS space
kubectl --context $its -n $monitoring_ns apply -f ${SCRIPT_DIR}/configuration/prometheus-secret.yaml

: 3. Copy secret from ITS space and re-create it in prometheus NS in core or host kubeflex cluster:
export secret_name="prometheus-$its-secret"
kubectl --context $its -n $monitoring_ns get secret prometheus-secret -o yaml | yq '.metadata |= .name |= strenv(secret_name)' | yq '.metadata |= (del(.annotations) |.annotations."kubernetes.io/service-account.name" |= "prometheus-kube-prometheus-prometheus") |= with_entries(select(.key == "name" or .key == "annotations"))' | kubectl --context $ctx apply --namespace=$monitoring_ns -f -

: 4. Add label to the ITS apiserver service
kubectl -n $its-system label svc/vcluster its-app=kube-apiserver

: 5. Create the service monitor for prometheus to talk with ITS apiserver
sed s/%ITS%/$its/g ${SCRIPT_DIR}/configuration/its-apiserver-sm.yaml | kubectl -n $monitoring_ns apply -f -