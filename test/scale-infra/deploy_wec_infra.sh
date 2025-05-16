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

clustername="wec"
region="us-east-1"
aws_key_name=""
num_hosts=1
instance_type="t2.micro"
archt='x86_64' # e.g., x86_64 and arm64
num_wecs=1
ec2_image_id="ami-00eb69d236edcfaf8"
vpc_name="kscore"


while [ $# != 0 ]; do
    case "$1" in
        (-h|--help) echo "$0 usage: (--cluster-name | --region | --vpc-name | --aws-key-name | --wecs-hosting-instances | --instances-type | --ec2-image-id | --arch )*"
                    exit;;
        (--cluster-name)
          if (( $# > 1 )); then
            clustername="$2"
            shift
          else
            echo "Missing cluster name value" >&2
            exit 1;
          fi;;
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
        (--wecs-hosting-instances)
          if (( $# > 1 )); then
            num_hosts="$2"
            shift
          else
            echo "Missing number of k8s masters node value" >&2
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
    esac
    shift
done


if [ $aws_key_name == "" ];then
   echo "AWS ssh public key name is empty."
   exit 1;
fi

## Create EC2 instances:
ansible-playbook create-ec2.yaml -e "cluster_name=$clustername region=$region vpc_name=$vpc_name aws_key_name=$aws_key_name  wecs_hosting_instances=$num_hosts instance_type=$instance_type ec2_arch=$arch ec2_image=$ec2_image_id"