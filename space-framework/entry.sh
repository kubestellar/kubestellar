#!/usr/bin/env bash

# Copyright 2023 The KubeStellar Authors.
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


# This is only a place holder once the integration with helm is done this file will be replaced
set -e

function echoerr() {
   echo "ERROR: $1" >&2
}

function create_or_replace() { # usage: filename
    filename="$1"
    kind=$(grep kind: "$filename" | head -1 | awk '{ print $2 }')
    name=$(grep name: "$filename" | head -1 | awk '{ print $2 }')
    if kubectl get "$kind" "$name" &> /dev/null ; then
        kubectl --kubeconfig /home/spacecore/.kube/config replace -f "$filename"
    else
        kubectl --kubeconfig /home/spacecore/.kube/config create -f "$filename"
    fi
}

function set_config_and_secret_and_crds() {
    mkdir -p /home/spacecore/.kube
    KUBECONFIG=/home/spacecore/.kube/config
    kubectl config set-cluster space-mgt --server="https://${KUBERNETES_SERVICE_HOST}:${KUBERNETES_SERVICE_PORT}" --certificate-authority=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt
    kubectl config set-credentials space-mgt --token="$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)"
    kubectl config set-context space-mgt --cluster=space-mgt --user=space-mgt 
    kubectl config use-context space-mgt

    echo "Delete if it exists and then create a secret for the core cluster."
    if ! kubectl --kubeconfig /home/spacecore/.kube/config delete secret -n ${NAMESPACE} corecluster ; then
        echo "Nothing to delete."
    fi
    kubectl --kubeconfig /home/spacecore/.kube/config create secret generic -n ${NAMESPACE} corecluster --from-file=kubeconfig=/home/spacecore/.kube/config

    echo "Apply space manager CRDs."
    for crd in /home/spacecore/config/crds/*.yaml; do
    create_or_replace $crd
    done
}

function run_space_manager() {
    echo "--< Starting space-manager >--"
    bin/space-manager --v=${VERBOSITY} --context space-mgt --kubeconfig /home/spacecore/.kube/config &
    while true ; do
        echo "***READY***"
        sleep 600
    done
}

echo "--< Starting SpaceManager container >--"
echo "Environment variables:"
if [ $# -ne 0 ] ; then
    ACTION="$1"
else
    ACTION="sleep"
fi
if [ "$VERBOSITY" == "" ]; then
    VERBOSITY="2"
fi
if [ "$NAMESPACE" == "" ]; then
    NAMESPACE="default"
fi
echo "VERBOSITY=${VERBOSITY}"
echo "NAMESPACE=${NAMESPACE}"
echo "ACTION=${ACTION}"

case "${ACTION}" in
(space-manager)
    set_config_and_secret_and_crds
    run_space_manager;;
(sleep)
    echo "Nothing to do... sleeping forever."
    sleep infinity;;
(*)
    echoerr "unknown action '$1'!"
    exit 1;;
esac
