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
ns="ks-monitoring"


while [ $# != 0 ]; do
    case "$1" in
        (-h|--help) echo "$0 usage: (--host-cluster-context (e.g., --host-cluster-context core-cluster (default value: kind-kubeflex)) | --monitoring-ns (e.g., --monitoring-ns monitoring (default value: ks-monitoring))*"
                    exit;;
        (--host-cluster-context)
          if (( $# > 1 )); then
            ctx="$2"
            shift
          else
            echo "Missing host-cluster-context value" >&2
            exit 1;
          fi;;
        (--monitoring-ns)
          if (( $# > 1 )); then
            ns="$2"
            shift
          else
            echo "Missing monitoring-ns value" >&2
            exit 1;
          fi;;
    esac
    shift
done

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)
source "${SCRIPT_DIR}/setup-shell.sh"

"${SCRIPT_DIR}"/cleanup.sh
: --------------------------------------------------------------
: Create KS monitoring namespace

kubectl config use-context $ctx
kubectl create namespace $ns

: --------------------------------------------------------------
: Install Pyroscope

helm repo add grafana https://grafana.github.io/helm-charts
helm repo update

helm -n $ns install pyroscope grafana/pyroscope --version 1.7.1

wait-for-object $ctx $ns statefulset pyroscope
wait-for-object $ctx $ns statefulset pyroscope-alloy

: --------------------------------------------------------------
: Install Grafana

helm upgrade -n $ns --install grafana grafana/grafana \
  --set image.repository=grafana/grafana \
  --set image.tag=main \
  --set env.GF_FEATURE_TOGGLES_ENABLE=flameGraph \
  --set env.GF_AUTH_ANONYMOUS_ENABLED=true \
  --set env.GF_AUTH_ANONYMOUS_ORG_ROLE=Admin \
  --set env.GF_DIAGNOSTICS_PROFILING_ENABLED=true \
  --set env.GF_DIAGNOSTICS_PROFILING_ADDR=0.0.0.0 \
  --set env.GF_DIAGNOSTICS_PROFILING_PORT=6060 \
  --set-string 'podAnnotations.pyroscope\.grafana\.com/scrape=true' \
  --set-string 'podAnnotations.pyroscope\.grafana\.com/port=6060' \
  --values ${SCRIPT_DIR}/grafana/datasources.yaml \
  --version 8.4.1

wait-for-object $ctx $ns deployment grafana

: --------------------------------------------------------------
: Install Prometheus

helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

helm -n $ns install -f ${SCRIPT_DIR}/prometheus/prometheus-custom-values.yaml prometheus prometheus-community/kube-prometheus-stack \
   --set kubeStateMetrics.enabled=false \
   --set nodeExporter.enabled=false \
   --set grafana.enabled=false \
   --set alertmanager.enabled=false \
   --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false \
   --version 61.7.0

wait-for-object $ctx $ns deployment prometheus-kube-prometheus-operator
wait-for-object $ctx $ns statefulsets prometheus-prometheus-kube-prometheus-prometheus