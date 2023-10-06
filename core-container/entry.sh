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

set -e

KUBESTELLAR_SERVICE="kubestellar"


function echoerr() {
   echo "ERROR: $1" >&2
}


function ensure_inv() {
    while IFS=':' read -ra items; do
        for i in "${items[@]}"; do
            if [ "$i" == "" ] ; then
                break
            fi
            if ! kubectl ws "$i" &> /dev/null ; then #
                if ! kubectl ws create "$i" --enter &> /dev/null ; then
                    echoerr 'unable to create "$i"!'
                    break
                fi
            fi
        done
    done <<< "$1"
}

function ensure_invs() {
    initial_ws=$(kubectl ws . --short)

    while IFS=',' read -ra items; do
        for i in "${items[@]}"; do
            if [ "$i" == "" ] ; then
                continue
            fi
            echo "Ensuring INV: $i"
            ensure_inv "$i"
            kubectl ws use $initial_ws &> /dev/null # go back to the initial state
        done
    done <<< "$1"

    kubectl ws use $initial_ws &> /dev/null
}


function ensure_wmw() {
    while IFS=':' read -ra items; do
        len=${#items[@]}
        if [[ ${len} -eq 0 ]]; then
            break # nothing to do
        fi
        # separate the last item
        last_item=${items[$len-1]}
        unset items[$len-1]
        for i in "${items[@]}"; do
            if [ "$i" == "" ] ; then
                break
            fi
            if ! kubectl ws "$i" &> /dev/null ; then #
                if ! kubectl ws create "$i" --enter &> /dev/null ; then
                    echoerr 'unable to create "$i"!'
                    break
                fi
            fi
        done
        kubectl kubestellar ensure wmw $last_item --with-kube true &> /dev/null
    done <<< "$1"
}


function ensure_wmws() {
    initial_ws=$(kubectl ws . --short)

    while IFS=',' read -ra items; do
        for i in "${items[@]}"; do
            if [ "$i" == "" ] ; then
                continue
            fi
            echo "Ensuring WMW: $i"
            ensure_wmw "$i"
            kubectl ws use $initial_ws &> /dev/null # go back to the initial state
        done
    done <<< "$1"

    kubectl ws use $initial_ws &> /dev/null
}


function wait_kcp_ready() {
    echo "Waiting for kcp to be ready... this may take a while."
    (
        KUBECONFIG=
        until [ "$(kubectl logs $(kubectl get pod --selector=app=kubestellar -o jsonpath='{.items[0].metadata.name}') -c kcp | grep '***READY***')" != "" ]; do
            sleep 5
        done
    )
    echo "Succes!"
    echo "Copying the admin.kubeconfig from kubestellar seret..."
    mkdir -p /home/kubestellar/.kcp
    (
        KUBECONFIG=
        kubectl get secrets kubestellar -o 'go-template={{index .data "admin.kubeconfig"}}' | base64 --decode > /home/kubestellar/.kcp/admin.kubeconfig
    )
}


function wait-kubestellar-ready() {
    wait_kcp_ready
    echo "Waiting for KubeStellar to be ready... this may take a while."
    (
        KUBECONFIG=
        until [ "$(kubectl logs $(kubectl get pod --selector=app=kubestellar -o jsonpath='{.items[0].metadata.name}') -c init | grep '***READY***')" != "" ]; do
            sleep 5
        done
    )
    echo "Succes!"
}


function guess_kcp_dns() {
    KUBECONFIG=
    if [ -z "$EXTERNAL_HOSTNAME" ]; then
        # Try to guess the route
        if kubectl get route kubestellar-route &> /dev/null; then
            EXTERNAL_HOSTNAME=$(kubectl get route kubestellar-route -o yaml -o jsonpath={.spec.host} 2> /dev/null)
        fi
    fi
    if [ -z "$EXTERNAL_HOSTNAME" ]; then
        # Try to guess the ingress
        if kubectl get ingress kubestellar-ingress &> /dev/null; then
            EXTERNAL_HOSTNAME=$(kubectl get ingress kubestellar-ingress -o yaml -o jsonpath={.spec.rules[0].host} 2> /dev/null)
        fi
    fi
    echo "${EXTERNAL_HOSTNAME}"
}


function create_or_replace() { # usage: filename
    filename="$1"
    kind=$(grep kind: "$filename" | head -1 | awk '{ print $2 }')
    name=$(grep name: "$filename" | head -1 | awk '{ print $2 }')
    if kubectl get "$kind" "$name" &> /dev/null ; then
        kubectl replace -f "$filename"
    else
        kubectl create -f "$filename"
    fi
}


function run_kcp() {
    echo "--< Starting kcp >--"

    echo Attempting to delete kubestellar secret...
    (
        KUBECONFIG=
        ! kubectl delete secret kubestellar
    )
    echo "EXTERNAL_HOSTNAME=${EXTERNAL_HOSTNAME}"

    # Check EXTERNAL_HOSTNAME
    if [ -z "$EXTERNAL_HOSTNAME" ]; then
        echo "Trying to guess the DNS from route/ingress...."
        export EXTERNAL_HOSTNAME=$(guess_kcp_dns)
    fi
    echo "EXTERNAL_HOSTNAME=${EXTERNAL_HOSTNAME}"

    # Create the certificates
    if [ -n "$EXTERNAL_HOSTNAME" ]; then
        echo "Creating the TLS certificates"
        # mkdir -p .kcp
        cd .kcp
        eval pieces_external=($(kubestellar-ensure-kcp-server-creds ${EXTERNAL_HOSTNAME}))
        eval pieces_cluster=($(kubestellar-ensure-kcp-server-creds ${KUBESTELLAR_SERVICE})) #############
        cd ..
    fi

    # Running kcp
    if [ -n "$EXTERNAL_HOSTNAME" ]; then
        echo "Removing existing apiserver keys... "
        ! rm /home/kubestellar/.kcp/apiserver.* &> /dev/null # required to fix the restart
        echo -n "Running kcp with TLS keys... "
        kcp start --tls-sni-cert-key ${pieces_external[1]},${pieces_external[2]} --tls-sni-cert-key ${pieces_cluster[1]},${pieces_cluster[2]} & # &> kcp.log &
    else
        echo -n "Running kcp without TLS keys... "
        kcp start &
    fi
    echo Started.

    # Waiting to be ready
    echo "Waiting for ${KUBECONFIG}..."
    while [ ! -f "${KUBECONFIG}" ]; do sleep 5; done
    echo 'Waiting for "root:compute" workspace...'
    until [ "$(kubectl ws root:compute 2> /dev/null)" != "" ]; do sleep 5;done
    echo '"root:compute" workspace is ready'.
    echo "kcp version: $(kubectl version --short 2> /dev/null | grep kcp | sed 's/.*kcp-//')"
    kubectl ws root

    # Generate the external.kubeconfig
    if [ -n "$EXTERNAL_HOSTNAME" ] && [ ! -d "${PWD}/.kcp-${EXTERNAL_HOSTNAME}" ]; then
        echo Creating external.kubeconfig...
        switch-domain .kcp/admin.kubeconfig .kcp/external.kubeconfig root ${EXTERNAL_HOSTNAME} ${EXTERNAL_PORT} ${pieces_external[0]}
        switch-domain .kcp/admin.kubeconfig .kcp/cluster.kubeconfig root ${KUBESTELLAR_SERVICE} 6443 ${pieces_cluster[0]}
    fi

    # Ensure kubeconfig secret
    echo Creating the kubestellar secret...
    (
        KUBECONFIG=
        if [ -n "${EXTERNAL_HOSTNAME}" ]; then
            kubectl create secret generic kubestellar --from-file="${PWD}/.kcp/admin.kubeconfig" --from-file="${PWD}/.kcp/cluster.kubeconfig" --from-file="${PWD}/.kcp/external.kubeconfig"
        else
            kubectl create secret generic kubestellar --from-file="${PWD}/.kcp/admin.kubeconfig"
        fi
    )
    echo "***READY***"
    sleep infinity
}


function run_init() {
    echo "--< Starting init >--"
    wait_kcp_ready
    kubestellar init --local-kcp false
    # ensure_root_compute_configd
    # ensure_espw
    kubectl ws root
    ensure_invs $ENSURE_IMW
    ensure_wmws $ENSURE_WMW
    touch ready
    echo "***READY***"
    sleep infinity
}


function run_mailbox_controller() {
    echo "--< Starting mailbox-controller >--"
    wait-kubestellar-ready
    kubectl ws root:espw
    if ! mailbox-controller -v=${VERBOSITY} ; then
        echoerr "unable to start mailbox-controller!"
        exit 1
    fi
}


function run_where_resolver() {
    echo "--< Starting where-resolver >--"
    wait-kubestellar-ready
    kubectl ws root:espw
    if ! kubestellar-where-resolver -v ${VERBOSITY} ; then
        echoerr "unable to start kubestellar-where-resolver!"
        exit 1
    fi
}


function run_placement_translator() {
    echo "--< Starting placement-translator >--"
    wait-kubestellar-ready
    kubectl ws root:espw
    if ! placement-translator --allclusters-context  "system:admin" -v=${VERBOSITY} ; then
        echoerr "unable to start mailbox-controller!"
        exit 1
    fi
}


echo "--< Starting KubeStellar container >--"

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
if [ "$ESPW_NAME" == "" ]; then
    ESPW_NAME="espw"
fi
echo "ESPW_NAME=${ESPW_NAME}"
echo "VERBOSITY=${VERBOSITY}"
echo "ENSURE_IMW=${ENSURE_IMW}"
echo "ENSURE_WMW=${ENSURE_WMW}"


case "${ACTION}" in
(kcp)
    run_kcp;;
(init)
    run_init;;
(mailbox-controller)
    run_mailbox_controller;;
(where-resolver)
    run_where_resolver;;
(placement-translator)
    run_placement_translator;;
(sleep)
    echo "Nothing to do... sleeping forever."
    sleep infinity;;
(*)
    echoerr "unknown action '$1'!"
    exit 1;;
esac
