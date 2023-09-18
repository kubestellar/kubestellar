#!/bin/bash

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

# ------------------------------------------------------------------------
## Before running this script, make sure KCP is running (./bin/kcp start)
## and KCPCONFIG in line 32 is set correctly
# ------------------------------------------------------------------------

. demo-magic.sh
clear

TYPE_SPEED=20
#DEMO_PROMPT="\[\033[01;32m\]\u@\h\[\033[00m\]:\[\033[01;34m\]\w\[\033[00m\]\$ "
DEMO_PROMPT="\[\033[01;32m\]\u@WSL2\[\033[00m\]:\[\033[01;34m\]\[\033[00m\]\$"

MGT_NAME="mgt"
MGT_CTX="kind-$MGT_NAME"
# change KCPCONFIG to point to KCP admin.kubeconfig file
KCPCONFIG="/home/user/go/src/github.com/kcp-dev/kcp/.kcp/admin.kubeconfig"

# For live demo use 'pe'  for recording or auto demo use 'pei'
#PCMD="pe"
PCMD="pei"

# add wait to be able to start a demo
p ""
#wait

#1 Create mgt cluster
$PCMD "kind create cluster --name $MGT_NAME"
echo "  "

#2 list the kind clusters
$PCMD "kind get clusters"
echo "  "

#3 Apply the space and provider CRDs on the mgt cluster
$PCMD "kubectl --context $MGT_CTX create -f config/crds/space.kubestellar.io_spaces.yaml"
$PCMD "kubectl --context $MGT_CTX create -f config/crds/space.kubestellar.io_spaceproviderdescs.yaml"
echo "  "
echo "Start the space manager in a second window "
p "press <enter> to continue."
#4 Start the manager in a second window
echo "  "

#5 Create a provider (set config, show the provider yaml,  filter prefix, etc..)
$PCMD "./cmd/ms-client-test/set-provider-config.sh $KCPCONFIG config/samples/spaceproviderdesc_kcp.yaml"
echo " "
$PCMD "cat config/samples/spaceproviderdesc_kcp.yaml"
echo " "
echo " "
$PCMD "kubectl --context $MGT_CTX create -f config/samples/spaceproviderdesc_kcp.yaml"
echo " "
$PCMD  "kubectl --context $MGT_CTX get spaceproviderdesc"
echo " "

#6 Show the created NS
$PCMD "kubectl --context $MGT_CTX get namespaces"
echo " "

#7 Create a space1
$PCMD "cat config/samples/space1.yaml"
echo " "
echo " "
$PCMD "kubectl --context $MGT_CTX create -f config/samples/space1.yaml"
echo " "

#8 Show status of space1 changes from initializing to READY
i=0; while ((i<3)); do kubectl --context $MGT_CTX describe space space1 -n spaceprovider-default  2>&1 | grep 'Phase: '; sleep 2; ((i=i+1)); done
kubectl --context $MGT_CTX wait spaces -n spaceprovider-default space1 --for=jsonpath='{.status.Phase}'=Ready  &>/dev/null
kubectl --context $MGT_CTX describe space space1 -n spaceprovider-default  2>&1 | grep 'Phase: '
echo " "

#9 Create a space2
echo " "
$PCMD "kubectl --context $MGT_CTX create -f config/samples/space2.yaml"
echo " "

#10 Show status of space2 changes from initializing to READY
i=0; while ((i<3)); do kubectl --context $MGT_CTX describe space space2 -n spaceprovider-default  2>&1 | grep 'Phase: ' ; sleep 2; ((i=i+1)); done
kubectl --context $MGT_CTX wait spaces -n spaceprovider-default space2 --for=jsonpath='{.status.Phase}'=Ready  &>/dev/null
kubectl --context $MGT_CTX describe space space2 -n spaceprovider-default  2>&1 | grep 'Phase: '
echo " "

#11 list KCP workspaces - show the created workspaces X1 & X2
$PCMD "kubectl ws tree --kubeconfig $KCPCONFIG"
echo "  "

$PCMD "kubectl --context $MGT_CTX get space -A"
echo "  "


#12 delete space1
$PCMD "kubectl --context $MGT_CTX delete space -n spaceprovider-default space1"
echo "  "

$PCMD "kubectl --context $MGT_CTX get space -A"
echo "  "

#13 List workspaces
$PCMD "kubectl ws tree --kubeconfig $KCPCONFIG"
echo "  "

#14 Delete space2 on the KCP
$PCMD "kubectl delete ws space2 --kubeconfig $KCPCONFIG"
echo "  "

$PCMD "kubectl ws tree --kubeconfig $KCPCONFIG"
echo "  "

$PCMD "kubectl --context $MGT_CTX get space -A"
echo "  "

#15 Show space2 is in "not-ready" state
$PCMD "kubectl --context $MGT_CTX describe space space2 -n spaceprovider-default | grep 'Phase: '"
echo "  "

#16 Delete space2
$PCMD "kubectl --context $MGT_CTX delete space -n spaceprovider-default space2"
echo "  "

#17 Create Y1 on KCP
$PCMD "kubectl ws create lc3 --kubeconfig $KCPCONFIG"
echo "  "

$PCMD "kubectl ws tree --kubeconfig $KCPCONFIG"
echo "  "

$PCMD "kubectl --context $MGT_CTX get space -A"
echo "  "

#18 Create prefix-Y2 on KCP
$PCMD "kubectl ws create ks-lc4 --kubeconfig $KCPCONFIG"
echo "  "

#19 List LCs - show that we discover only prefix-Y2
$PCMD "kubectl --context $MGT_CTX get space -A"
echo "  "
