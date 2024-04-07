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

set -x # echo so that users can understand what is happening
set -e # exit on error

:
: -------------------------------------------------------------------------
: "Cleaning up from previous run of an e2e test"

if [ -z "$USE_K3D" ]; then
   kind delete cluster --name cluster1
   kind delete cluster --name cluster2
   kind delete cluster --name kubeflex
else
   k3d cluster delete cluster1
   k3d cluster delete cluster2
   k3d cluster delete kubeflex
fi
kubectl config delete-context cluster1 || true
kubectl config delete-context cluster2 || true
kubectl config delete-context kubeflex || true
kubectl config delete-context imbs1 || true
kubectl config delete-context wds1 || true
