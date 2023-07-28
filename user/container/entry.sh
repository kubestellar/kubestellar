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

echo "< Starting Kubestellar container >-------------------------"

# Try to guess the route/ingress
if [ -z "$EXTERNAL_HOSTNAME" ]; then
    if kubectl get route kubestellar-route &> /dev/null; then
        EXTERNAL_HOSTNAME=$(kubectl get route kubestellar-route -o yaml -o jsonpath={.spec.host} 2> /dev/null)
    fi
fi
if [ -z "$EXTERNAL_HOSTNAME" ]; then
    if kubectl get ingress kubestellar-ingress &> /dev/null; then
        EXTERNAL_HOSTNAME=$(kubectl get ingress kubestellar-ingress -o yaml -o jsonpath={.spec.rules[0].host} 2> /dev/null)
    fi
fi

# Create the certificates
if [ -n "$EXTERNAL_HOSTNAME" ]; then
    echo "< Ensuring the TLS certificate >---------------------------"
    eval pieces=($(kubestellar-ensure-kcp-server-creds ${EXTERNAL_HOSTNAME}))
    echo "TLS certificates for server ${EXTERNAL_HOSTNAME}:"
    echo ${pieces[0]}
    echo ${pieces[1]}
    echo ${pieces[2]}
fi

# Start kcp
echo "< Starting kcp >-------------------------------------------"

mkdir -p ${PWD}/kubestellar-logs # required in Open Shift clusters

if [ -n "$EXTERNAL_HOSTNAME" ]; then
    echo -n "Running kcp with TLS keys... "
    kcp start --tls-sni-cert-key ${pieces[1]},${pieces[2]} &>> "${PWD}/kubestellar-logs/kcp.log" &
else
    echo -n "Running kcp without TLS keys... "
    kcp start &>> "${PWD}/kubestellar-logs/kcp.log" &
fi
echo "logfile=./kubestellar-logs/kcp.log"

echo "Waiting for kcp to be ready... it may take a while"
until [ "$(kubectl ws root:compute 2> /dev/null)" != "" ]; do
    sleep 5
done

echo "kcp version: $(kubectl version --short 2> /dev/null | grep kcp | sed 's/.*kcp-//')"

kubectl ws root

if [ -n "$EXTERNAL_HOSTNAME" ] && [ ! -d "${PWD}/.kcp-${EXTERNAL_HOSTNAME}" ]; then
    echo "Switching the admin.kubeconfig domain to ${EXTERNAL_HOSTNAME} and port ${EXTERNAL_PORT}..."
    switch-domain .kcp/admin.kubeconfig .kcp/external.kubeconfig root ${EXTERNAL_HOSTNAME} ${EXTERNAL_PORT} ${pieces[0]}
fi

# Starting KubeStellar
echo "< Starting KubeStellar >-----------------------------------"

kubestellar start

# Create secrets in Kuberntes cluster
echo "< Create secrets >-----------------------------------------"

# set KUBECONFIG to empty string so that `kubectl` falls back on its in-cluster technique for reading credentials from the filesystem.
KUBECONFIG=

echo "Ensure secret in the current namespace..."
if kubectl delete secret kubestellar 2> /dev/null; then
    echo " Deleted secret in the current namespace."
fi
if [ -n "${EXTERNAL_HOSTNAME}" ]; then
    if kubectl create secret generic kubestellar --from-file="${PWD}/.kcp/admin.kubeconfig" --from-file="${PWD}/.kcp/external.kubeconfig" 2> /dev/null; then
        echo " Created secret in the current namespace."
    else
        echo " Not running in a cluster!"
    fi
else
    if kubectl create secret generic kubestellar --from-file="${PWD}/.kcp/admin.kubeconfig" 2> /dev/null; then
        echo " Created secret in the current namespace."
    else
        echo " Not running in a cluster!"
    fi
fi

# Done, sleep forerver...
touch ready
echo "Ready!"
sleep infinity
