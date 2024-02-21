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

set -x # echo so users can understand what is happening
set -e # exit on error
SRC_DIR="$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)"

if [[ "$KFLEX_DISABLE_CHATTY" = true ]] ; then
   disable_chatty_status="--chatty-status=false"
   echo "disable_chatty_status = $disable_chatty_status"
fi

:
: -------------------------------------------------------------------------
: Create a Kind hosting cluster with nginx ingress controller and KubeFlex controller-manager
:
kflex init --create-kind $disable_chatty_status
: Kubeflex kind cluster created.

:
: -------------------------------------------------------------------------
: Install the post-create-hooks for ocm and kubstellar controller manager
:
kubectl apply -f ${SRC_DIR}/../../../config/postcreate-hooks/ocm.yaml
kubectl apply -f ${SRC_DIR}/../../../config/postcreate-hooks/kubestellar.yaml
: Kubestellar post-create-hooks applied.

:
: -------------------------------------------------------------------------
: 'Create an inventory & mailbox space of type vcluster running OCM (Open Cluster Management) directly in KubeFlex. Note that -p ocm runs a post-create hook on the vcluster control plane which installs OCM on it.'
:
kflex create imbs1 --type vcluster -p ocm $disable_chatty_status
: imbs1 created.

:
: -------------------------------------------------------------------------
: Install singleton status return addon in IMBS1
:
wait-for-cmd kubectl --context imbs1 api-resources "|" grep managedclusteraddons
helm --kube-context imbs1 upgrade --install status-addon -n open-cluster-management oci://ghcr.io/kubestellar/ocm-status-addon-chart --version v0.2.0-rc2

:
: -------------------------------------------------------------------------
: Create a Workload Description Space wds1 directly in KubeFlex.
:
kflex create wds1 $disable_chatty_status
kubectl --context kind-kubeflex label cp wds1 kflex.kubestellar.io/cptype=wds

cd "${SRC_DIR}/../../.."
pwd
make ko-build-local
make install-local-chart KUBE_CONTEXT=kind-kubeflex
cd -
echo "wds1 created."

:
: -------------------------------------------------------------------------
: Run OCM transport controller executable
:
cd "${SRC_DIR}/../../.." ## go up to KubeStellar directory
KUBESTELLAR_DIR="$(pwd)"
## this is a temp solution - run as executble process. this should be replaced to run as pod, ideally using helm
wget https://github.com/kubestellar/ocm-transport-plugin/archive/refs/tags/v0.1.0-rc1.tar.gz
tar -xf v0.1.0-rc1.tar.gz
rm -rf v0.1.0-rc1.tar.gz
cd ocm-transport-plugin-0.1.0-rc1
OCM_TRANSPORT_PLUGIN_DIR="$(pwd)"
pwd
echo "replace github.com/kubestellar/kubestellar => ${KUBESTELLAR_DIR}/" >> go.mod
make build
mv ./bin/ocm-transport-plugin ${KUBESTELLAR_DIR}/ocm-transport-plugin
cd "${KUBESTELLAR_DIR}"
pwd
rm -rf ${OCM_TRANSPORT_PLUGIN_DIR}
echo "running ocm transport plugin..."
./ocm-transport-plugin --transport-context imbs1 --wds-context wds1 --wds-name wds1 &> transport.log &

echo "transport controller is running as background process."

:
: -------------------------------------------------------------------------
: Create clusters and register with OCM
:
function create_cluster() {
  cluster=$1
  kind create cluster --name $cluster
  kubectl config rename-context kind-${cluster} $cluster
  clusteradm --context imbs1 get token | grep '^clusteradm join' | sed "s/<cluster_name>/${cluster}/" | awk '{print $0 " --context '${cluster}' --force-internal-endpoint-lookup"}' | sh
}

create_cluster cluster1
create_cluster cluster2

: Wait for csrs in imbs1
wait-for-cmd '(($(kubectl --context imbs1 get csr 2>/dev/null | grep -c Pending) >= 2))'

clusteradm --context imbs1 accept --clusters cluster1
clusteradm --context imbs1 accept --clusters cluster2

kubectl --context imbs1 get managedclusters
kubectl --context imbs1 label managedcluster cluster1 location-group=edge name=cluster1
kubectl --context imbs1 label managedcluster cluster2 location-group=edge name=cluster2

:
: -------------------------------------------------------------------------
: Get all deployments and statefulsets running in the hosting cluster.
: Expect to see the wds1 kubestellar-controller-manager created in the wds1-system 
: namespace and the imbs1 statefulset created in the imbs1-system namespace.
:
if ! expect-cmd-output 'kubectl --context kind-kubeflex get deployments,statefulsets --all-namespaces' 'grep -e wds1 -e imbs1 | wc -l | grep -wq 4'
then
    echo "Failed to see wds1 deployment and imbs1 statefulset."
    exit 1
fi

:
: -------------------------------------------------------------------------
: "Get available clusters with label location-group=edge and check there are two of them"
:
if ! expect-cmd-output 'kubectl --context imbs1 get managedclusters -l location-group=edge' 'wc -l | grep -wq 3'
then
    echo "Failed to see two clusters."
    exit 1
fi
