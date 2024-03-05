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

# get-kubeconfig-for-controlplane receives a ControlPlane name and a parameter
# that determines whether to extract the ControlPlane's in-cluster kubeconfig
# or the external kubeconfig (if set to "true", the first will be retrieved).
# The function returns the requested kubeconfig's content in base64.
get-kubeconfig-for-controlplane() {
  cpname="$1"
  get_incluster_key="$2"

  key=""
  if [[ $get_incluster_key == "true" ]];
  then
    key=$(kubectl get controlplane $cpname -o=jsonpath='{.status.secretRef.inClusterKey}');
  else
    key=$(kubectl get controlplane $cpname -o=jsonpath='{.status.secretRef.key}');
  fi

  # get secret details
  secret_name=$(kubectl get controlplane $cpname -o=jsonpath='{.status.secretRef.name}')
  secret_namespace=$(kubectl get controlplane $cpname -o=jsonpath='{.status.secretRef.namespace}')

  # get the kubeconfig (already base64)
  local kubeconfig_base64=$(kubectl get secret $secret_name -n $secret_namespace -o=jsonpath="{.data.$key}")
  echo "$kubeconfig_base64"
}

export -f get-kubeconfig-for-controlplane