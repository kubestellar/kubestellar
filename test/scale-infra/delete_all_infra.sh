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
vpc_name=""

while [ $# != 0 ]; do
    case "$1" in
        (-h|--help) echo "$0 usage: (--region (e.g., us-east-1) | --vpc-name (e.g., kscore))*"
                    exit;;
        (--region)
          if (( $# > 1 )); then
            region="$2"
            shift
          else
            echo "Missing region value" >&2
            exit 1;
          fi;;
        (--vpc-name)
          if (( $# > 1 )); then
            vpc_name="$2"
            shift
          else
            echo "Missing vpc name value" >&2
            exit 1;
          fi;;
    esac
    shift
done

hosts_wec_exist=false
hosts_core_exist=false
vpc_config_exist=false

if [ $region == "" ];then
   echo "AWS region name is empty."
   exit 1;
fi

if [ $vpc_name == "" ];then
   echo "VPC name is empty."
   exit 1;
fi


if [ -f ".data/${region}_${vpc_name}/hosts_wec" ]; then
   hosts_wec_exist=true
fi

if [ -f ".data/${region}_${vpc_name}/hosts_core" ]; then
   hosts_core_exist=true
fi

if [ -f ".data/${region}_${vpc_name}/vpc_helper_vars.yaml" ]; then
   vpc_config_exist=true
fi


if $hosts_wec_exist; then
    ## 1. Delete WECs:
    ansible-playbook -i .data/${region}_${vpc_name}/hosts_wec delete-ec2.yaml -e "cluster_name=wec region=$region group=$vpc_name"
else
    echo "File 'hosts_wec' does not exist - skipping deletion of the WEC infrastructure"
fi

if $hosts_core_exist; then
    ## 2. Delete CORE:
    ansible-playbook -i .data/${region}_${vpc_name}/hosts_core delete-ec2.yaml -e "cluster_name=core region=$region group=$vpc_name"
else
    echo "File 'hosts_core' does not exist - skipping deletion of the KS control plane hosting infrastructure"
fi

if $vpc_config_exist; then
    ## 3. Delete VPC:
    ansible-playbook delete_vpc_infra.yaml -e "region=$region vpc_name=$vpc_name"
else
    echo "File with deployed AWS VPC configuration does not exist - skipping deletion of the VPC infrastructure"
fi