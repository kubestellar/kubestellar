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
set -e

region=""
aws_key_name=""
num_workers=""
instance_type=""
ks_release="0.24.0"
archt='x86_64' # e.g., x86_64 and arm64

while (( $# > 0 )); do
    if [ "$1" == "--region" ]; then
        region=$2
        shift
    elif [ "$1" == "--aws_key_name" ]; then
        aws_key_name=$2
        shift
    elif [ "$1" == "--num_workers" ]; then
        num_workers=$2
        shift
    elif [ "$1" == "--instance_type" ]; then
        instance_type=$2
        shift
    elif [ "$1" == "--archt" ]; then
        archt=$2
        shift
    fi 
    shift
done

# 1. Deploy vpc:
ansible-playbook deploy_vpc_core.yaml -e "region=$region"

# 2. Deploy instances:
ansible-playbook create-ec2.yaml -e "cluster_name=core region=$region aws_key_name=$aws_key_name  num_workers=$num_workers instance_type=$instance_type archt=$archt" 

# 3. Install k8s:
ansible-playbook -i .data/hosts_core deploy-master.yaml --ssh-common-args='-o StrictHostKeyChecking=no'
ansible-playbook -i .data/hosts_core deploy-worker.yaml --ssh-common-args='-o StrictHostKeyChecking=no'

# 4. Deploy KubeStellar in the hosting cluster:
ansible-playbook -i .data/hosts_core deploy_ks_core.yaml --ssh-common-args='-o StrictHostKeyChecking=no' -e "ks_release=$ks_release"