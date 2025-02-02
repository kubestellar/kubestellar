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
num_masters=""
num_workers=""
instance_type=""
ks_release=""
archt='x86_64' # e.g., x86_64 and arm64
ec2_image_id=""
vpc_name=""


while [ $# != 0 ]; do
    case "$1" in
        (-h|--help) echo "$0 usage: (--region | --vpc-name | --aws-key-name | --k8s-num-masters |  --k8s-num-workers |  --instances-type | --ec2-image-id | --arch |  --ks-release)*"
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
        (--aws-key-name)
          if (( $# > 1 )); then
            aws_key_name="$2"
            shift
          else
            echo "Missing aws key name value" >&2
            exit 1;
          fi;;
        (--k8s-num-masters)
          if (( $# > 1 )); then
            num_masters="$2"
            shift
          else
            echo "Missing number of k8s masters node value" >&2
            exit 1;
          fi;;
        (--k8s-num-workers)
          if (( $# > 1 )); then
            num_workers="$2"
            shift
          else
            echo "Missing number of k8s workers node value" >&2
            exit 1;
          fi;;
        (--instances-type)
          if (( $# > 1 )); then
            instance_type="$2"
            shift
          else
            echo "Missing instance type value" >&2
            exit 1;
          fi;;
        (--ec2-image-id)
          if (( $# > 1 )); then
            ec2_image_id="$2"
            shift
          else
            echo "Missing ec2 image id value" >&2
            exit 1;
          fi;;
        (--arch)
          if (( $# > 1 )); then
            archt="$2"
            shift
          else
            echo "Missing arch value" >&2
            exit 1;
          fi;;
        (--ks-release)
          if (( $# > 1 )); then
            ks_releas="$2"
            shift
          else
            echo "Missing ks release value" >&2
            exit 1;
          fi;;
    esac
    shift
done

# 1. Deploy vpc:
ansible-playbook deploy_vpc_core.yaml -e "name=$vpc_name region=$region"

# 2. Deploy instances:
ansible-playbook create-ec2.yaml -e "cluster_name=core region=$region vpc_name=$vpc_name aws_key_name=$aws_key_name  num_masters=$num_masters num_workers=$num_workers instance_type=$instance_type arch=$arch ec2_image=$ec2_image_id" 

# 3. Install k8s:
ansible-playbook -i .data/hosts_core deploy-masters.yaml --ssh-common-args='-o StrictHostKeyChecking=no'
ansible-playbook -i .data/hosts_core deploy-workers.yaml --ssh-common-args='-o StrictHostKeyChecking=no'

# 4. Deploy KubeStellar in the hosting cluster:
ansible-playbook -i .data/hosts_core deploy_ks_core.yaml --ssh-common-args='-o StrictHostKeyChecking=no' -e "ks_release=$ks_release"