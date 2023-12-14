#!/usr/bin/env bash

# Run the space manager. To do this we first keep track of which KUBECONFIG 
# we are going to use for the space manager. We then apply the space framework 
# CRDs on the space manager cluster. And finally we actually run the space manager.

export SM_KUBECONFIG=$PWD/sm.kubeconfig
cp $HOME/.kube/config $SM_KUBECONFIG
export KUBECONFIG=$SM_KUBECONFIG
kubectl apply -f kubestellar/space-framework/config/crds
kubestellar/space-framework/bin/space-manager --kubeconfig $SM_KUBECONFIG --context kind-sm-mgt &> /tmp/space-manager.log &
