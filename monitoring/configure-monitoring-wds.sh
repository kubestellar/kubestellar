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
wds="wds1"
monitoring_ns="ks-monitoring"

while [ $# != 0 ]; do
    case "$1" in
        (-h|--help) echo "$0 usage: ( -h | --help | --kubeflex-hosting-cluster-context (e.g., --kubeflex-hosting-cluster-context core-cluster (default value: kind-kubeflex)) | --space-name (e.g., --space-name wds1 (default value: wds1)) | --monitoring-ns (e.g., --monitoring-ns ks-monitoring (default value: ks-monitoring)))*"
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
            wds="$2"
            shift
          else
            echo "Missing space-name value" >&2
            exit 1;
          fi;;
        (--monitoring-ns)
          if (( $# > 1 )); then
            wds="$2"
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
: Configure APF for prometheus traffic to the WDS API server
: --------------------------------------------------------------------
kubectl --context $wds apply -f ${SCRIPT_DIR}/configuration/prometheus-pod-exempt.yaml

: --------------------------------------------------------------------
: Configure kubestellar controller manager pod for prometheus scraping
: --------------------------------------------------------------------

: 1. Create service monitor for KS ctl manager
sed s/%WDS%/$wds/g ${SCRIPT_DIR}/configuration/ks-ctl-manager-sm.yaml | kubectl -n $monitoring_ns apply -f -


: --------------------------------------------------------------------
: Configure kubestellar controller manager pod for pyroscope scraping
: --------------------------------------------------------------------
: 1. Add labels for pyroscope scraping:
kubectl -n $wds-system get deploy kubestellar-controller-manager -o yaml | yq '.spec.template.metadata.annotations."profiles.grafana.com/cpu.port" |= "8082" |  .spec.template.metadata.annotations."profiles.grafana.com/cpu.scrape"|= "true" | .spec.template.metadata.annotations."profiles.grafana.com/goroutine.port" |= "8082" | .spec.template.metadata.annotations."profiles.grafana.com/goroutine.scrape" |= "true" |
.spec.template.metadata.annotations."profiles.grafana.com/memory.port" |= "8082" | .spec.template.metadata.annotations."profiles.grafana.com/memory.scrape" |= "true"' | kubectl -n $wds-system apply -f -


: --------------------------------------------------------------------
: Configure kubestellar transport controller pod for prometheus scraping
: --------------------------------------------------------------------
: 1. Create transport controller service:
kubectl -n $wds-system apply -f ${SCRIPT_DIR}/configuration/ks-transport-ctl-svc.yaml

: 2. Adding declarations of the metrics and pprof ports, so that transport controller service definition can refer to it by name
kubectl -n $wds-system get deploy transport-controller -o yaml | yq '(del(.status) |.spec.template.spec.containers.[0].ports[0].name |= "metrics")' | yq '.spec.template.spec.containers.[0].ports[0].protocol |= "TCP"' | yq '.spec.template.spec.containers.[0].ports[0].containerPort |= 8090' | yq '.spec.template.spec.containers.[0].ports[1].name |= "pprof"' | yq '.spec.template.spec.containers.[0].ports[1].protocol |= "TCP"' | yq '.spec.template.spec.containers.[0].ports[1].containerPort |= 8092' | kubectl --context $ctx apply --namespace=$wds-system -f -

: 3. Create the service monitor:
sed s/%WDS%/$wds/g ${SCRIPT_DIR}/configuration/ks-transport-ctl-sm.yaml | kubectl -n $monitoring_ns apply -f -


: --------------------------------------------------------------------
: Configure kubestellar transport controller pod for pyroscope scraping
: --------------------------------------------------------------------
kubectl -n $wds-system get deploy transport-controller -o yaml | yq '.spec.template.metadata.annotations."profiles.grafana.com/cpu.port" |= "8092" |  .spec.template.metadata.annotations."profiles.grafana.com/cpu.scrape"|= "true" | .spec.template.metadata.annotations."profiles.grafana.com/goroutine.port" |= "8092" | .spec.template.metadata.annotations."profiles.grafana.com/goroutine.scrape" |= "true" |
.spec.template.metadata.annotations."profiles.grafana.com/memory.port" |= "8092" | .spec.template.metadata.annotations."profiles.grafana.com/memory.scrape" |= "true"' | kubectl -n $wds-system apply -f -


: --------------------------------------------------------------------
: Configure WDS API server pod for prometheus scraping
: --------------------------------------------------------------------
: 1. Create a SA and give the right RBAC to talk to the wds API server
kubectl --context $wds get ns $monitoring_ns || kubectl --context $wds create ns $monitoring_ns
kubectl --context $wds -n $monitoring_ns apply -f ${SCRIPT_DIR}/prometheus/prometheus-rbac.yaml

: 2. Create token secret for prometheus in the target wds space
kubectl --context $wds -n $monitoring_ns apply -f ${SCRIPT_DIR}/configuration/prometheus-secret.yaml

: 3. Copy secret from wds space and re-create it in prometheus NS in core or host kubeflex cluster:
export secret_name="prometheus-$wds-secret"
kubectl --context $wds -n $monitoring_ns get secret prometheus-secret -o yaml | yq '.metadata |= .name |= strenv(secret_name) ' | yq '.metadata |= (del(.annotations) |.annotations."kubernetes.io/service-account.name" |= "prometheus-kube-prometheus-prometheus") |= with_entries(select(.key == "name" or .key == "annotations"))' | kubectl --context $ctx apply --namespace=$monitoring_ns -f -

: 4. Add label to the wds apiserver service
kubectl -n $wds-system label svc/$wds app=kube-apiserver

: 5. Create the service monitor for prometheus to talk with wds apiserver
sed s/%WDS%/$wds/g ${SCRIPT_DIR}/configuration/wds-apiserver-sm.yaml | kubectl -n $monitoring_ns apply -f -