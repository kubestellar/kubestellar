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

if base64 -w 0 <<<"" &>/dev/null
then nolbs='-w 0'
else nolbs=''
fi

function echoerr() {
   echo "ERROR: $1" >&2
}

function wait_kcp_ready() {
    echo "Waiting for kcp to be ready... this may take a while."
    (
        until [ "$(kubectl logs $(kubectl get pod --selector=app=kubestellar -o jsonpath='{.items[0].metadata.name}') -c kcp | grep '***READY***')" != "" ]; do
           sleep 10
        done
    )
}

function get_kcp_kubeconfig() {
    while ! kubectl get secret kubestellar; do
        echo "Waiting for kubestellar secret."
        sleep 5
    done
    kcp_kubeconfig_dir="/home/kubestellar/.kcp"
    kcp_kubeconfig="${kcp_kubeconfig_dir}/admin.kubeconfig"
    external_kcp_kubeconfig="${kcp_kubeconfig_dir}/external.kubeconfig"
    mkdir -p $kcp_kubeconfig_dir
    echo "Copying the kubeconfigs from kubestellar seret into ${kcp_kubeconfig} and ${external_kcp_kubeconfig}..."
    kubectl get secrets kubestellar -o jsonpath='{.data.admin\.kubeconfig}' | base64 --decode > $kcp_kubeconfig
    kubectl get secrets kubestellar -o jsonpath='{.data.external\.kubeconfig}' | base64 --decode > $external_kcp_kubeconfig
}

function create_kcp_provider_object() {
    get_kcp_kubeconfig     # kcp is created in a seperate container 
    kubectl apply -f - <<EOF
apiVersion: v1
kind: Secret
metadata:
  namespace: ${NAMESPACE}
  name: kcpsec
type: Opaque
data:
  kubeconfig: $(base64 $nolbs < $kcp_kubeconfig)
  external: $(base64 $nolbs < $external_kcp_kubeconfig)
---
apiVersion: space.kubestellar.io/v1alpha1
kind: SpaceProviderDesc
metadata:
  name: $PROVIDER_NAME
spec:
  ProviderType: "kcp"
  SpacePrefixForDiscovery: "ks-"
  secretRef:
    namespace: ${NAMESPACE}
    name: kcpsec
EOF
    echo "Waiting for spaceproviderdesc to reach the Ready phase."
    kubectl wait --for=jsonpath='{.status.Phase}'=Ready spaceproviderdesc $PROVIDER_NAME
}
 
function create_kubeflex_provider_object() {
    echo "Waiting for the kubeflex provider to be ready... this may take a while."
    (
        until [ "$(kubectl get pods -A | grep kubeflex-controller-manager | grep Running)" != "" ]; do
            sleep 10
        done
    )

    kubectl apply -f - <<EOF
apiVersion: space.kubestellar.io/v1alpha1
kind: SpaceProviderDesc
metadata:
  name: $PROVIDER_NAME
spec:
  ProviderType: "kubeflex"
  SpacePrefixForDiscovery: "ks-"
  secretRef:
    namespace: ${NAMESPACE}
    name: corecluster
EOF
    echo "Waiting for spaceproviderdesc to reach the Ready phase."
    kubectl wait --for=jsonpath='{.status.Phase}'=Ready spaceproviderdesc $PROVIDER_NAME
}


function create_spaceprovider_object() {
    echo "Waiting for space manager to be ready... this may take a while."
    (
        until [ "$(kubectl logs $(kubectl get pod --selector=app=kubestellar -o jsonpath='{.items[0].metadata.name}') -c space-manager | grep '***READY***')" != "" ]; do
            sleep 10
        done
    )
    if [ "$SPACE_PROVIDER_TYPE" == "kcp" ]; then
        create_kcp_provider_object
    elif [ "$SPACE_PROVIDER_TYPE" == "kubeflex" ]; then
        create_kubeflex_provider_object
    else
        echo "${SPACE_PROVIDER_TYPE} is not a valid space provider."
    fi
}

function wait-kubestellar-ready() {
    echo "Waiting for KubeStellar to be ready... this may take a while."
    (
        until [ "$(kubectl logs $(kubectl get pod --selector=app=kubestellar -o jsonpath='{.items[0].metadata.name}') -c init | grep '***READY***')" != "" ]; do
            sleep 10
        done
    )
    echo "Success!"
}

function guess_kcp_dns() {
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

function run_kcp() {
    echo "--< Starting kcp >--"

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
         # required to fix the restart
        echo "Removing existing apiserver keys... "
        if ! rm /home/kubestellar/.kcp/apiserver.* &> /dev/null ; then
            echo "Nothing to remove... must be the first time."
        else
            echo "Existing keys removed... the container mast have restarted."
        fi
        echo -n "Running kcp with TLS keys... "
        kcp start --tls-sni-cert-key ${pieces_external[1]},${pieces_external[2]} --tls-sni-cert-key ${pieces_cluster[1]},${pieces_cluster[2]} & # &> kcp.log &
    else
        echo -n "Running kcp without TLS keys... "
        kcp start &
    fi
    PROVIDER_KUBECONFIG="${PWD}/.kcp/admin.kubeconfig"
    echo Started.

    # Waiting to be ready
    echo "Waiting for ${PROVIDER_KUBECONFIG}..."
    while [ ! -f "${PROVIDER_KUBECONFIG}" ]; do
        sleep 5;
    done
    echo 'Waiting for "root" workspace...'
    until [ "$(KUBECONFIG="$PROVIDER_KUBECONFIG" kubectl ws root 2> /dev/null)" != "" ]; do
        sleep 5;
    done
    echo $(date)': "root" workspace is ready'.
    echo "kcp version: $(KUBECONFIG="$PROVIDER_KUBECONFIG" kubectl version --short 2> /dev/null | grep kcp | sed 's/.*kcp-//')"

    # Generate the external.kubeconfig and cluster.kubeconfig
    if [ -n "$EXTERNAL_HOSTNAME" ] && [ ! -d "${PWD}/.kcp-${EXTERNAL_HOSTNAME}" ]; then
        echo Creating external.kubeconfig...
        switch-domain "${PROVIDER_KUBECONFIG}" .kcp/external.kubeconfig root ${EXTERNAL_HOSTNAME} ${EXTERNAL_PORT} ${pieces_external[0]}
        switch-domain "${PROVIDER_KUBECONFIG}" .kcp/cluster.kubeconfig root ${KUBESTELLAR_SERVICE} 6443 ${pieces_cluster[0]}
    fi

    # Ensure kubeconfig secret
    echo Creating the kubestellar secret...
    if [ -n "${EXTERNAL_HOSTNAME}" ]; then
	clucfg="  cluster.kubeconfig: $(base64 $nolbs < "${PWD}/.kcp/cluster.kubeconfig")"
    else
	clucfg=""
    fi
    sleep 10
    echo "$(date): Creating kubestellar Secret"
    kubectl apply -f - <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: kubestellar
type: Opaque
data:
  admin.kubeconfig: $(base64 $nolbs < "${PROVIDER_KUBECONFIG}")
  external.kubeconfig: $(base64 $nolbs < "${PWD}/.kcp/external.kubeconfig")
$clucfg
EOF

    touch ready
    while true ; do
        echo "***READY***"
        sleep 600
    done
}

function run_init() {
    echo "--< Starting init >--"
    create_spaceprovider_object
    kubestellar init -X --in-cluster --ensure-imw $ENSURE_IMW --ensure-wmw $ENSURE_WMW
    touch ready
    echo "***READY***"
    sleep infinity
}

function run_mailbox_controller() {
    echo "--< Starting mailbox-controller >--"
    wait-kubestellar-ready
    if ! mailbox-controller -v=${VERBOSITY} ; then
        echoerr "unable to start mailbox-controller!"
        exit 1
    fi
}

function run_where_resolver() {
    echo "--< Starting where-resolver >--"
    wait-kubestellar-ready
    if ! kubestellar-where-resolver -v ${VERBOSITY} ; then
        echoerr "unable to start kubestellar-where-resolver!"
        exit 1
    fi
}

function run_placement_translator() {
    echo "--< Starting placement-translator >--"
    wait-kubestellar-ready
    if ! placement-translator -v=${VERBOSITY} ; then
        echoerr "unable to start mailbox-controller!"
        exit 1
    fi
}

echo "--< Starting KubeStellar container >--"

export KUBECONFIG_DIR="${PWD}/space-config"
mkdir -p $KUBECONFIG_DIR

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
if [ "$NAMESPACE" == "" ]; then
    NAMESPACE="default"
fi
if [ "$PROVIDER_NAME" == "" ]; then
    PROVIDER_NAME="default"
    export PROVIDER_NAME
fi
if [ "$PROVIDER_NAMESPACE" == "" ]; then
    PROVIDER_NAMESPACE=spaceprovider-${PROVIDER_NAME}
    export PROVIDER_NAMESPACE
fi
if [ "$PROVIDER_SECRET_NAME" == "" ]; then
    PROVIDER_SECRET_NAME=psecret
fi
if [ "$PROVIDER_SECRET_NAMESPACE" == "" ]; then
    PROVIDER_SECRET_NAMESPACE=psecret_namespace
fi

echo "ESPW_NAME=${ESPW_NAME}"
echo "VERBOSITY=${VERBOSITY}"
echo "ENSURE_IMW=${ENSURE_IMW}"
echo "ENSURE_WMW=${ENSURE_WMW}"
echo "NAMESPACE=${NAMESPACE}"
echo "SPACE_PROVIDER_TYPE=${SPACE_PROVIDER_TYPE}"
echo "KUBECONFIG_DIR=${KUBECONFIG_DIR}"
echo "KUBECONFIG=${KUBECONFIG}"
echo "PROVIDER_NAME=${PROVIDER_NAME}"
echo "PROVIDER_NAMESPACE=${PROVIDER_NAMESPACE}"

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
