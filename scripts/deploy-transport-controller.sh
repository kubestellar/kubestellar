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

# Transport deployment script assumes it runs within a kubeflex context.

export WDS_NAME="$1" ## first argument is WDS name
export IMBS_NAME="$2" ## second argument is IMBS name
export TRANSPORT_CONTROLLER_IMAGE="${3:=ghcr.io/kubestellar/ocm-transport-plugin/transport-controller:0.1.0-rc4}" ## third argument is transport controller image
export HOST_IP=$(ifconfig | grep -w inet | awk '{ print $2 }' | grep -v 127.0.0.1 | head -1)
export IMAGE_PULL_POLICY="${IMAGE_PULL_POLICY:=Always}"

SRC_DIR="$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)"
envsubst '$WDS_NAME $IMBS_NAME $TRANSPORT_CONTROLLER_IMAGE $HOST_IP $IMAGE_PULL_POLICY' < "${SRC_DIR}/../deploy/transport-controller-deployment.yaml.template"  | kubectl apply -f -
