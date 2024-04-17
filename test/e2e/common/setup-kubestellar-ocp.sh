#!/usr/bin/env bash
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



set -x -e # echo so users can understand what is happening
set -e # exit on error

if [[ "$KFLEX_DISABLE_CHATTY" = true ]] ; then
   disable_chatty_status="--chatty-status=false"
   echo "disable_chatty_status = $disable_chatty_status"
fi



source <(curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/release-$KUBESTELLAR_VERSION/test/e2e/common/setup-shell.sh)

:
: -------------------------------------------------------------------------
: Init KubeFlex controller-manager
:
kubectl config use-context kscore
kflex init $disable_chatty_status

: Kubeflex controller-manager started.
:
: -------------------------------------------------------------------------
: Install the post-create-hooks for ocm and kubestellar controller manager
:
kubectl apply -f https://raw.githubusercontent.com/kubestellar/kubestellar/v${KUBESTELLAR_VERSION}/config/postcreate-hooks/kubestellar.yaml
kubectl apply -f https://raw.githubusercontent.com/kubestellar/kubestellar/v${KUBESTELLAR_VERSION}/config/postcreate-hooks/ocm.yaml
: 'Kubestellar post-create-hook(s) applied.'
:
: -------------------------------------------------------------------------
: 'Create an inventory & mailbox space of type vcluster running OCM (Open Cluster Management) directly in KubeFlex. Note that -p ocm runs a post-create hook on the vcluster control plane which installs OCM on it.'
:
kflex create its1 --type vcluster -p ocm $disable_chatty_status
: its1 created.

:
: -------------------------------------------------------------------------
: Install singleton status return addon in IMBS1
:
wait-for-cmd kubectl --context its1 api-resources "|" grep managedclusteraddons
helm --kube-context its1 upgrade --install status-addon -n open-cluster-management oci://ghcr.io/kubestellar/ocm-status-addon-chart --version v${OCM_STATUS_ADDON_VERSION}
:
: -------------------------------------------------------------------------
: Create a Workload Description Space wds1 directly in KubeFlex.
:
kflex create wds1 -p kubestellar
echo "wds1 created."

:
: -------------------------------------------------------------------------
: Run OCM transport controller in a pod
:
kflex ctx
bash <(curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/v${KUBESTELLAR_VERSION}/scripts/deploy-transport-controller.sh) wds1 its1
wait-for-cmd '(kubectl -n wds1-system wait --for=condition=Ready pod/$(kubectl -n wds1-system get pods -l name=transport-controller -o jsonpath='{.items[0].metadata.name}'))'

echo "transport controller is running."

:
: -------------------------------------------------------------------------

: Create clusters and register with OCM
:
function add_wec_cluster() {
  cluster=$1
  clusteradm --context its1 get token | grep '^clusteradm join' | sed "s/<cluster_name>/${cluster}/" | awk '{print $0 " --context '${cluster}' --singleton --force-internal-endpoint-lookup"}' | sh
}

bash <(curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/release-$KUBESTELLAR_VERSION/hack/check_pre_req.sh) --assert --verbose ocm

add_wec_cluster cluster1
add_wec_cluster cluster2

: Wait for csrs in its1
wait-for-cmd '(($(kubectl --context its1 get csr 2>/dev/null | grep -c Pending) >= 2))'

clusteradm --context its1 accept --clusters cluster1
clusteradm --context its1 accept --clusters cluster2

kubectl --context its1 get managedclusters
kubectl --context its1 label managedcluster cluster1 location-group=edge name=cluster1
kubectl --context its1 label managedcluster cluster2 location-group=edge name=cluster2

:
: -------------------------------------------------------------------------
: Get all deployments and statefulsets running in the hosting cluster.
: Expect to see the wds1 kubestellar-controller-manager and transport-controller created in the wds1-system
: namespace and the its1 statefulset created in the its1-system namespace.
:
if ! kubectl --context kscore get deployments,statefulsets --all-namespaces $disable_chatty_status | grep -e wds1 -e its1 | wc -l | grep -wq 5
then
    echo "Failed to see wds1 deployment and its1 statefulset."
    exit 1
fi

:
: -------------------------------------------------------------------------
: "Get available clusters with label location-group=edge and check there are two of them"
:
if ! expect-cmd-output 'kubectl --context its1 get managedclusters -l location-group=edge' 'wc -l | grep -wq 3'
then
    echo "Failed to see two clusters."
    exit 1
fi
