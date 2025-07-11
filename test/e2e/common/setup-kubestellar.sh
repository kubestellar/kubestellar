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


if [[ "$KFLEX_DISABLE_CHATTY" = true ]] ; then
   disable_chatty_status="--chatty-status=false"
   echo "disable_chatty_status = $disable_chatty_status"
fi

SRC_DIR="$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)"
COMMON_SRCS="${SRC_DIR}/../common"
source "$COMMON_SRCS/setup-shell.sh"

:
: -------------------------------------------------------------------------
: Create the KubeFlex hosting cluster, if necessary.
:
case "$CLUSTER_SOURCE" in
    (kind)
        ${SRC_DIR}/../../../scripts/create-kind-cluster-with-SSL-passthrough.sh --name kubeflex
        : Kubeflex kind cluster created.
        ;;
    (existing)
        kubectl config use-context "$HOSTING_CONTEXT"
        : kubectl configured to use existing cluster in "$HOSTING_CONTEXT" context
        ;;
esac


:
: -------------------------------------------------------------------------
: Install the core-chart
:

pushd "${SRC_DIR}/../../.."
if [ "$use_release" = true ] ; then
  helm upgrade --install ks-core oci://ghcr.io/kubestellar/kubestellar/core-chart \
    --version $(yq .KUBESTELLAR_VERSION core-chart/values.yaml) \
    --kube-context $HOSTING_CONTEXT \
    --set-json='ITSes=[{"name":"its1"}]' \
    --set-json='WDSes=[{"name":"wds1"}]' \
    --set verbosity.kubestellar=${KUBESTELLAR_CONTROLLER_MANAGER_VERBOSITY} \
    --set verbosity.transport=${TRANSPORT_CONTROLLER_VERBOSITY}
else
  make kind-load-image
  helm dependency update core-chart/
  helm upgrade --install ks-core core-chart/ \
    --set KUBESTELLAR_VERSION=$(git rev-parse --short HEAD) \
    --kube-context $HOSTING_CONTEXT \
    --set-json='ITSes=[{"name":"its1"}]' \
    --set-json='WDSes=[{"name":"wds1"}]' \
    --set verbosity.kubestellar=${KUBESTELLAR_CONTROLLER_MANAGER_VERBOSITY} \
    --set verbosity.transport=${TRANSPORT_CONTROLLER_VERBOSITY}
  fi
popd

: Waiting for OCM hub to be ready...
kubectl wait controlplane.tenancy.kflex.kubestellar.org/its1 --for 'jsonpath={.status.postCreateHooks.its-with-clusteradm}=true' --timeout 400s
kubectl wait -n its1-system job.batch/its-with-clusteradm --for condition=Complete --timeout 400s
kubectl wait -n its1-system job.batch/update-cluster-info --for condition=Complete --timeout 200s

wait-for-cmd "(kubectl --context '$HOSTING_CONTEXT' -n wds1-system wait --for=condition=Ready pod/\$(kubectl --context '$HOSTING_CONTEXT' -n wds1-system get pods -l name=transport-controller -o jsonpath='{.items[0].metadata.name}'))"

echo "transport controller is running."

kubectl config use-context "$HOSTING_CONTEXT"
kflex ctx --set-current-for-hosting
kflex ctx --overwrite-existing-context wds1
kflex ctx --overwrite-existing-context its1

kflex ctx

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
            kubectl config rename-context kind-${cluster} $cluster
            joinflags="--force-internal-endpoint-lookup";;
        (existing)
            joinflags="";;
    esac
    clusteradm --context its1 get token | grep '^clusteradm join' | sed "s/<cluster_name>/${cluster}/" | awk '{print $0 " --context '${cluster}' --singleton '${joinflags}'"}' | sh
}

"${SRC_DIR}/../../../scripts/check_pre_req.sh" --assert --verbose ocm

kubectl --context $HOSTING_CONTEXT wait controlplane.tenancy.kflex.kubestellar.org/its1 --for 'jsonpath={.status.postCreateHooks.its-with-clusteradm}=true' --timeout 200s
kubectl --context $HOSTING_CONTEXT wait -n its1-system job.batch/its-with-clusteradm --for condition=Complete --timeout 400s

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
