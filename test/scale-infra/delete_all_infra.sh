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

while (( $# > 0 )); do
    if [ "$1" == "--region" ]; then
        region=$2
        shift
    fi 
    shift
done


## 1. Delete WECs:
ansible-playbook -i .data/hosts_wec delete-ec2.yaml -e "cluster_name=wec region=$region"

## 2. Delete CORE:
ansible-playbook -i .data/hosts_core delete-ec2.yaml -e "cluster_name=core region=$region"

## 3. Delete VPC:
ansible-playbook delete_vpc_infra.yaml -e "region=$region"
