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


# Run the space manager. To do this we first keep track of which KUBECONFIG 
# we are going to use for the space manager. We then apply the space framework 
# CRDs on the space manager cluster. And finally we actually run the space manager.

export SM_KUBECONFIG=$PWD/sm.kubeconfig
cp $HOME/.kube/config $SM_KUBECONFIG
export KUBECONFIG=$SM_KUBECONFIG
kubectl apply -f kubestellar/space-framework/config/crds
kubestellar/space-framework/bin/space-manager --kubeconfig $SM_KUBECONFIG --context kind-sm-mgt &> /tmp/space-manager.log &
