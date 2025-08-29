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
opt="core" # (value: core or wec)
core_context="kind-kubeflex"
env="plain"


while [ $# != 0 ]; do
    case "$1" in
        (-h|--help) echo "$0 usage: (--cluster-context (e.g., --cluster-context core-cluster (default value: kind-kubeflex)) | --monitoring-ns (e.g., --monitoring-ns monitoring (default value: ks-monitoring)) | --set (e.g., --set wec (default value: core)) | --env (e.g., --env ocp (allowed values: ocp or plain)) | --kubeflex-hosting-cluster-context (e.g., --kubeflex-hosting-cluster-context kscore (default value: kind-kubeflex)) *"
                    exit;;
        (--cluster-context)
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
        (--set)
          if (( $# > 1 )); then
            opt="$2"
            shift
          else
            echo "Missing set value" >&2
            exit 1;
          fi;;
        (--env)
          if (( $# > 1 )); then
            env="$2"
            shift
          else
            echo "Missing set value" >&2
            exit 1;
          fi;;
        (--kubeflex-hosting-cluster-context)
          if (( $# > 1 )); then
            core_context="$2"
            shift
          else
            echo "Missing set value" >&2
            exit 1;
          fi;;
    esac
    shift
done

case "$opt" in
    (core|wec) ;;
    (*) echo "$0: --set must be 'core' or 'wec'" >&2
        exit 1;;
esac

case "$env" in
    (plain|ocp) ;;
    (*) echo "$0: --env must be 'plain' or 'ocp'" >&2
        exit 1;;
esac

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)
source "${SCRIPT_DIR}/setup-shell.sh"

"${SCRIPT_DIR}"/cleanup.sh --cluster-context $ctx --ns $ns
: --------------------------------------------------------------
: Create KS monitoring namespace

kubectl config use-context $ctx
kubectl create namespace $ns

if [ $opt == "core" ];then
    : --------------------------------------------------------------
    : Install Thanos with MinIO storage

    : Install MinIO MinIO object storage for Thanos to store metrics
    helm repo add minio https://charts.min.io/

    if [ $env == "plain" ];then
       helm install minio -n $ns -f ${SCRIPT_DIR}/prometheus/minio-custom-values.yaml minio/minio  --version 5.2.0 --set ingress.enabled=true

    elif [ $env == "ocp" ];then
       helm install minio -n $ns -f ${SCRIPT_DIR}/prometheus/minio-custom-values.yaml minio/minio  --version 5.2.0 --set ingress.enabled=false
       kubectl -n $ns apply -f ${SCRIPT_DIR}/prometheus/minio-route.yaml
    fi 

    wait-for-object $ctx $ns statefulset minio

    # Retrieve the MinIO credentials
    export ROOT_USER=$(kubectl get secret -n $ns minio -o jsonpath="{.data.rootUser}" | base64 -d)
    export ROOT_PASSWORD=$(kubectl get secret -n $ns minio -o jsonpath="{.data.rootPassword}" | base64 -d)

    # Create secret to allow Thanos access to object storage
    sed -e s/%USERNAME%/$ROOT_USER/g -e s/%PASSWORD%/$ROOT_PASSWORD/g ${SCRIPT_DIR}/prometheus/objstore.yml > /tmp/objstore.yml

    kubectl -n $ns create secret generic thanos-objstore --from-file=/tmp/objstore.yml
    # Delete temporary file
    rm /tmp/objstore.yml

    : Install Thanos
    helm repo add bitnami https://charts.bitnami.com/bitnami

    if [ $env == "plain" ];then
       helm -n $ns upgrade --install thanos bitnami/thanos --values ${SCRIPT_DIR}/prometheus/thanos-custom-values.yaml  --version 15.7.25  \
         --set queryFrontend.ingress.enabled=true  \
         --set receive.ingress.enabled=true

    elif [ $env == "ocp" ];then
       helm -n $ns upgrade --install thanos bitnami/thanos --values ${SCRIPT_DIR}/prometheus/thanos-custom-values.yaml  --version 15.7.25 \
         --set queryFrontend.ingress.enabled=false  \
         --set receive.ingress.enabled=false
       kubectl -n $ns apply -f ${SCRIPT_DIR}/prometheus/thanos-route.yaml
    fi 

    wait-for-object $ctx $ns deployment thanos-compactor
    wait-for-object $ctx $ns deployment thanos-query
    wait-for-object $ctx $ns deployment thanos-query-frontend 
    wait-for-object $ctx $ns statefulsets thanos-storegateway

    : --------------------------------------------------------------
    : Install Grafana
    helm repo add grafana https://grafana.github.io/helm-charts
    helm repo update

    helm upgrade -n $ns --install grafana grafana/grafana \
      --set image.repository=grafana/grafana \
      --set image.tag=main \
      --set env.GF_FEATURE_TOGGLES_ENABLE=flameGraph \
      --set env.GF_AUTH_ANONYMOUS_ENABLED=true \
      --set env.GF_AUTH_ANONYMOUS_ORG_ROLE=Admin \
      --set env.GF_DIAGNOSTICS_PROFILING_ENABLED=true \
      --set env.GF_SECURITY_ALLOW_EMBEDDING=true \
      --set env.GF_SECURITY_X_FRAME_OPTIONS='' \
      --set env.GF_DIAGNOSTICS_PROFILING_ADDR=0.0.0.0 \
      --set env.GF_DIAGNOSTICS_PROFILING_PORT=6060 \
      --set-string 'podAnnotations.pyroscope\.grafana\.com/scrape=true' \
      --set-string 'podAnnotations.pyroscope\.grafana\.com/port=6060' \
      --values ${SCRIPT_DIR}/grafana/datasources.yaml \
      --version 8.4.1

    wait-for-object $ctx $ns deployment grafana

    if [ $env == "ocp" ];then
         kubectl -n $ns apply -f ${SCRIPT_DIR}/grafana/grafana-route.yaml
    fi 

    # Set the Thanos URL for Prometheus remote write
    export thanos_host="thanos-receive.ks-monitoring.svc.cluster.local:19291"
    endpoint="http://$thanos_host"
    sed -e s,%THANOS_HOST%,$endpoint,g -e s,%ENABLE_KUBEAPISERVER%,true,g ${SCRIPT_DIR}/prometheus/prometheus-custom-values.yaml > /tmp/prometheus-custom-values-set.yaml

    # Configure Pyroscope to connect to MinIO 
    export ENDPOINT="minio:9000"
    export BUCKET="pyroscope"
    export INSECURE="true"
    sed -e s/%ENDPOINT%/$ENDPOINT/g -e s/%BUCKET%/$BUCKET/g -e s/%USERNAME%/$ROOT_USER/g -e s/%PASSWORD%/$ROOT_PASSWORD/g -e s/%INSECURE%/$INSECURE/g ${SCRIPT_DIR}/grafana/pyroscope-config.yaml > /tmp/pyroscope-config-values-set.yaml
    
elif [ $opt == "wec" ];then
  # Set the Thanos URL for Prometheus remote write
  if [ $env == "plain" ];then
       export THANOS_ENDPOINT="http://kubeflex-control-plane:80"
       export MinIO_ENDPOINT="kubeflex-control-plane:32000"
       export INSECURE="true"

  elif [ $env == "ocp" ];then
       thanos_host=$(kubectl --context $core_context -n $ns get route thanos-receive -o jsonpath="{.status.ingress[0].host}")
       export THANOS_ENDPOINT="https://$thanos_host"
       export MinIO_ENDPOINT=$(kubectl --context $core_context -n $ns get route minio-api -o jsonpath="{.status.ingress[0].host}")
       export INSECURE="false"
  fi

  sed -e s,%THANOS_HOST%,$THANOS_ENDPOINT,g -e s,%ENABLE_KUBEAPISERVER%,false,g ${SCRIPT_DIR}/prometheus/prometheus-custom-values.yaml > /tmp/prometheus-custom-values-set.yaml

  # Configure Pyroscope to connect to MinIO remote storage on the hosting cluster (hub)
  # (1) Retrieve the MinIO credentials
  export ROOT_USER=$(kubectl --context $core_context get secret -n $ns minio -o jsonpath="{.data.rootUser}" | base64 -d)
  export ROOT_PASSWORD=$(kubectl --context $core_context get secret -n $ns minio -o jsonpath="{.data.rootPassword}" | base64 -d)
  export BUCKET="pyroscope"
  
  # (2) Configure Pyroscope
  sed -e s/%ENDPOINT%/$MinIO_ENDPOINT/g -e s/%BUCKET%/$BUCKET/g -e s/%USERNAME%/$ROOT_USER/g -e s/%PASSWORD%/$ROOT_PASSWORD/g -e s/%INSECURE%/$INSECURE/g ${SCRIPT_DIR}/grafana/pyroscope-config.yaml > /tmp/pyroscope-config-values-set.yaml
else
   echo "Unknown value set for parameter --set" >&2
   exit 1;
fi

: --------------------------------------------------------------
: Install Prometheus with remote write Thanos
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

helm -n $ns install -f /tmp/prometheus-custom-values-set.yaml prometheus prometheus-community/kube-prometheus-stack \
   --set kubeStateMetrics.enabled=false \
   --set nodeExporter.enabled=false \
   --set grafana.enabled=false \
   --set alertmanager.enabled=false \
   --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false \
   --version 61.7.0

rm /tmp/prometheus-custom-values-set.yaml
wait-for-object $ctx $ns deployment prometheus-kube-prometheus-operator
wait-for-object $ctx $ns statefulsets prometheus-prometheus-kube-prometheus-prometheus

: --------------------------------------------------------------
: Install Pyroscope
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update
helm -n $ns -f /tmp/pyroscope-config-values-set.yaml install pyroscope grafana/pyroscope --version 1.7.1

wait-for-object $ctx $ns statefulset pyroscope
wait-for-object $ctx $ns statefulset pyroscope-alloy

# Delete temporary file
rm /tmp/pyroscope-config-values-set.yaml
