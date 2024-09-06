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

use_release=false
KUBESTELLAR_CONTROLLER_MANAGER_VERBOSITY=5
TRANSPORT_CONTROLLER_VERBOSITY=5
CLUSTER_SOURCE=kind
HOSTING_CONTEXT=kind-kubeflex

while [ $# != 0 ]; do
    case "$1" in
        (-h|--help)
            echo "$0 usage: (--released | --kubestellar-controller-manager-verbosity \$num | --transport-controller-verbosity \$num | --env \$kind_or_ocp)*"
            exit;;
        (--released)
            use_release=true;;
        (--kubestellar-controller-manager-verbosity)
          if (( $# > 1 )); then
            KUBESTELLAR_CONTROLLER_MANAGER_VERBOSITY="$2"
            shift
          else
            echo "Missing kubestellar-controller-manager-verbosity value" >&2
            exit 1;
          fi;;
        (--transport-controller-verbosity)
          if (( $# > 1 )); then
            TRANSPORT_CONTROLLER_VERBOSITY="$2"
            shift
          else
            echo "Missing transport-controller-verbosity value" >&2
            exit 1;
          fi;;
        (--env)
          if (( $# < 1 )); then
            echo "Missing --env value" >&2
            exit 1
          fi
          case "$2" in
            (kind) CLUSTER_SOURCE=kind;     HOSTING_CONTEXT=kind-kubeflex;;
            (ocp)  CLUSTER_SOURCE=existing; HOSTING_CONTEXT=kscore;;
            (*) echo "--env must be given 'kind' or 'ocp'" >&2
                exit 1;;
          esac
          shift;;
        (*) echo "$0: unrecognized argument/flag '$1'" >&2
            exit 1
    esac
    shift
done

if ! [[ "$KUBESTELLAR_CONTROLLER_MANAGER_VERBOSITY" =~ [0-9]+ ]]
then echo "$0: \$KUBESTELLAR_CONTROLLER_MANAGER_VERBOSITY must be an integer" >&2
     exit 1
fi

if ! [[ "$TRANSPORT_CONTROLLER_VERBOSITY" =~ [0-9]+ ]]
then echo "$0: \$TRANSPORT_CONTROLLER_VERBOSITY must be an integer" >&2
     exit 1
fi

if [ "$use_release" = true ]
then wds_extra="-p kubestellar --set ControllerManagerVerbosity=$KUBESTELLAR_CONTROLLER_MANAGER_VERBOSITY --set TransportControllerVerbosity=$TRANSPORT_CONTROLLER_VERBOSITY"
else wds_extra=""
fi

if [[ "$KFLEX_DISABLE_CHATTY" = true ]] ; then
   disable_chatty_status="--chatty-status=false"
   echo "disable_chatty_status = $disable_chatty_status"
fi

SRC_DIR="$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)"
COMMON_SRCS="${SRC_DIR}/../common"
source "$COMMON_SRCS/setup-shell.sh"

:
: -------------------------------------------------------------------------
: Initialize KubeFlex, creating or using an existing hosting cluster.
:
case "$CLUSTER_SOURCE" in
    (kind)
        kflex init --create-kind $disable_chatty_status
        : Kubeflex kind cluster created.
        ;;
    (existing)
        kubectl config use-context "$HOSTING_CONTEXT"
        kflex init $disable_chatty_status
        : KubeFlex initialized to use existing cluster in "$HOSTING_CONTEXT" context
        ;;
esac


:
: -------------------------------------------------------------------------
: Install the post-create-hooks for ocm and kubstellar controller manager
:
kubectl apply -f ${SRC_DIR}/../../../config/postcreate-hooks/ocm.yaml
if [ "$use_release" == true ]
then kubectl apply -f ${SRC_DIR}/../../../config/postcreate-hooks/kubestellar.yaml
fi
: 'Kubestellar post-create-hook(s) applied.'

:
: -------------------------------------------------------------------------
: 'Create an inventory & mailbox space of type vcluster running OCM (Open Cluster Management) directly in KubeFlex. Note that -p ocm runs a post-create hook on the vcluster control plane which installs OCM on it.'
:
kflex create its1 --type vcluster -p ocm $disable_chatty_status
: its1 created.

:
: -------------------------------------------------------------------------
: Create a Workload Description Space wds1 directly in KubeFlex.
:
kflex create wds1 $wds_extra $disable_chatty_status
kubectl --context "$HOSTING_CONTEXT" label cp wds1 kflex.kubestellar.io/cptype=wds

if [ "$use_release" != true ]; then
  cd "${SRC_DIR}/../../.."
  pwd
  make ko-build-controller-manager-local
  make install-local-chart KUBE_CONTEXT="$HOSTING_CONTEXT" "KUBESTELLAR_CONTROLLER_MANAGER_VERBOSITY=$KUBESTELLAR_CONTROLLER_MANAGER_VERBOSITY"
  cd -
fi
echo "wds1 created."

:
: -------------------------------------------------------------------------
: Run OCM transport controller in a pod
:
if [ "$use_release" != true ]; then
  pushd "${SRC_DIR}/../../.." ## go up to KubeStellar directory
  OCM_TRANSPORT_PLUGIN_RELEASE=local
  IMAGE_TAG=${OCM_TRANSPORT_PLUGIN_RELEASE} make ko-build-transport-local
  kind load --name kubeflex docker-image ko.local/ocm-transport-controller:${OCM_TRANSPORT_PLUGIN_RELEASE} # load local image to kubeflex
  echo "running ocm transport plugin..."
  IMAGE_PULL_POLICY=Never ./scripts/deploy-transport-controller.sh wds1 its1 ko.local/ocm-transport-controller:${OCM_TRANSPORT_PLUGIN_RELEASE} --controller-verbosity "$TRANSPORT_CONTROLLER_VERBOSITY" --context "$HOSTING_CONTEXT"
  popd
fi

wait-for-cmd "(kubectl --context '$HOSTING_CONTEXT' -n wds1-system wait --for=condition=Ready pod/\$(kubectl --context '$HOSTING_CONTEXT' -n wds1-system get pods -l name=transport-controller -o jsonpath='{.items[0].metadata.name}'))"

echo "transport controller is running."

wait-for-cmd 'kubectl --context its1 get ns customization-properties'

:
: -------------------------------------------------------------------------
: Create clusters and register with OCM
:
function add_wec() {
    cluster=$1
    case "$CLUSTER_SOURCE" in
        (kind)
            kind create cluster --name $cluster
            kubectl config rename-context kind-${cluster} $cluster;;
        (existing) ;;
    esac
  clusteradm --context its1 get token | grep '^clusteradm join' | sed "s/<cluster_name>/${cluster}/" | awk '{print $0 " --context '${cluster}' --singleton --force-internal-endpoint-lookup"}' | sh
}

"${SRC_DIR}/../../../hack/check_pre_req.sh" --assert --verbose ocm

add_wec cluster1
add_wec cluster2

: Wait for csrs in its1
wait-for-cmd '(($(kubectl --context its1 get csr 2>/dev/null | grep -c Pending) >= 2))'

clusteradm --context its1 accept --clusters cluster1
clusteradm --context its1 accept --clusters cluster2

kubectl --context its1 get managedclusters
kubectl --context its1 label managedcluster cluster1 location-group=edge name=cluster1 region=east
kubectl --context its1 create cm -n customization-properties cluster1 --from-literal clusterURL=https://my.clusters/1001-abcd
kubectl --context its1 label managedcluster cluster2 location-group=edge name=cluster2 region=west
kubectl --context its1 create cm -n customization-properties cluster2 --from-literal clusterURL=https://my.clusters/2002-cdef

:
: -------------------------------------------------------------------------
: Get all deployments and statefulsets running in the hosting cluster.
: Expect to see the wds1 kubestellar-controller-manager and transport-controller created in the wds1-system
: namespace and the its1 statefulset created in the its1-system namespace.
:
wait-for-cmd "((\$(kubectl --context '$HOSTING_CONTEXT' get deployments,statefulsets --all-namespaces | grep -e wds1 -e its1 | wc -l) == 5))"

:
: -------------------------------------------------------------------------
: "Get available clusters with label location-group=edge and check there are two of them"
:
if ! expect-cmd-output 'kubectl --context its1 get managedclusters -l location-group=edge' 'wc -l | grep -wq 3'
then
    echo "Failed to see two clusters."
    exit 1
fi
