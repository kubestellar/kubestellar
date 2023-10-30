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

function run_space_manager() {
    echo "--< Starting space-manager 1 >--"
    KUBECONFIG=
    kubectl config set-cluster space-mgt --server="https://${KUBERNETES_SERVICE_HOST}:${KUBERNETES_SERVICE_PORT}" --certificate-authority=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt
    kubectl config set-credentials space-mgt --token="$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)"
    kubectl config set-context space-mgt --cluster=space-mgt --user=space-mgt 
    kubectl config use-context space-mgt
    KUBECONFIG=/home/spacecore/.kube/config

    echo "Apply space manager CRDs."
    kubectl apply -f /home/spacecore/config/crds

    echo "Waiting for kcp to be ready... this may take a while."
    (
        until [ "$(kubectl logs $(kubectl get pod --selector=app=kubestellar -o jsonpath='{.items[0].metadata.name}') -c kcp | grep '***READY***')" != "" ]; do
            sleep 10
        done
    )
 
    echo "Create the kcp provider secret."
    kubectl --kubeconfig /home/spacecore/.kube/config get secrets kubestellar -o 'go-template={{index .data "admin.kubeconfig"}}' | base64 --decode > kcpsecret
    kubectl --kubeconfig /home/spacecore/.kube/config create secret generic kcpsec --from-file=kubeconfig="kcpsecret"    

    echo "Create a secret for the core cluster which is also where the kubeflex core cluster."
    kubectl --kubeconfig /home/spacecore/.kube/config create secret generic kflex --from-file=kubeconfig=/home/spacecore/.kube/config

    echo "Apply kubeflex and kcp providers."
    kubectl --kubeconfig /home/spacecore/.kube/config apply -f config/spaceproviderdesc-kflex.yaml
    kubectl --kubeconfig /home/spacecore/.kube/config apply -f config/spaceproviderdesc-kcp.yaml

    # Running the space-manager 
    if ! bin/space-manager --v=${VERBOSITY} --context space-mgt --kubeconfig /home/spacecore/.kube/config; then
        echoerr "unable to start space-manager!"
        exit 1
    fi
}

echo "--< Starting SpaceManager container >--"

echo "Environment variables:"
if [ $# -ne 0 ] ; then
    ACTION="$1"
else
    ACTION="sleep"
fi
echo "ACTION=${ACTION}"
if [ "$VERBOSITY" == "" ]; then
    VERBOSITY="2"
fi

echo "VERBOSITY=${VERBOSITY}"

case "${ACTION}" in
(space-manager)
    run_space_manager;;
(sleep)
    echo "Nothing to do... sleeping forever."
    sleep infinity;;
(*)
    echoerr "unknown action '$1'!"
    exit 1;;
esac
