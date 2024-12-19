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
archt='x86_64' # e.g., x86_64 and arm64
num_wecs=5
wec_name_prefix=location

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
    elif [ "$1" == "--num_wecs" ]; then
        num_wecs=$2
        shift
    elif [ "$1" == "--wec_name_prefix" ]; then
        wec_name_prefix=$2
        shift
    fi 
    shift
done


## 1. Create EC2 instances:
ansible-playbook create-ec2.yaml -e "cluster_name=wec region=$region aws_key_name=$aws_key_name  num_masters=0  num_workers=$num_workers instance_type=$instance_type archt=$archt image_source=ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server-20240423" 

## 2. Create Kind cluster WECs and connect to KS Core cluster
#sudo ansible-playbook -i .data/hosts_wec deploy_ks_wec.yaml --ssh-common-args='-o StrictHostKeyChecking=no' -e "num_wecs=$num_wecs wec_name_prefix=$wec_name_prefix"

