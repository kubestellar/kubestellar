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

# Usage: $0 $kubectl_context $local_image_ref $dest_image_name:tag

# This script uploads a given local container image to an OCP cluster.
# The user must already be logged into that cluster with `oc`.

# This script is given a kubectl context that refers to an OCP cluster
# as a user with admin privilege. A value of "." means the current
# context.

# https://access.redhat.com/documentation/en-us/openshift_container_platform/4.13/html/registry/registry-overview-1

# https://access.redhat.com/documentation/en-us/openshift_container_platform/4.13/html/images/managing-images

# This script first ensures that the OpenShift image registry has an
# external hostname, configuring it to use its default route if
# necessary.  Then this script adds a new reference to a locally-held
# image: the pre-existing local image reference is $local_image_ref
# (which can be either registry/na/mespace/name:tag or hash) and the
# new reference is $regext/openshift/$dest_image_name:tag. Here
# $regext is the external hostname of the registry. Then this script
# does a `docker login` to that registry.  Then this script does a
# `docker push` of the new image reference.

# This script also prints to stdout a line of the form
# internalRegistryHostname=host:port

# This script also prints to stdout a line of the form
# in-cluster-image-ref=host:port/openshift/$dest_image_name:tag

# So a cluster-internal reference to the image can be obtained by
# piping the output of this script through
#
# grep in-cluster-image-ref= | cut -f 2 -d=

if [ $# != 3 ]; then
    echo "$0 usage: \$kubeconfig_context \$local_image_ref \$dest_image_name:tag" >&2
    exit 1
fi

context="$1"
shift
local_image="$1"
shift
remote_name_tag="$1"
shift

set -x
set -e

if [ "$context" == "." ]; then
    context=$(kubectl config current-context)
fi

regext=$(kubectl --context "$context" get images.config cluster -o 'jsonpath={.status.externalRegistryHostnames[0]}')

if [ -z "$regext" ]; then
    oc --context "$context" patch configs.imageregistry.operator.openshift.io/cluster --patch '{"spec":{"defaultRoute":true}}' --type=merge
    sleep 5
    regext=$(kubectl --context "$context" get images.config cluster -o 'jsonpath={.status.externalRegistryHostnames[0]}')
    if [ -z "$regext" ]; then
	echo "$0: Failed to get external hostname for OpenShift image registry" >&2
	exit 10
    fi
fi

regint=$(kubectl --context "$context" get images.config cluster -o 'jsonpath={.status.internalRegistryHostname}')

if [ -z "$regint" ]; then
    echo "$0: Failed to get internal hostname for OpenShift image registry" >&2
    exit 20
fi
echo "internalRegistryHostname=$regint"
echo "in-cluster-image-ref=${regint}/openshift/$remote_name_tag"

oc --context "$context" whoami -t | docker login -u kubeadmin --password-stdin "$regext"

docker tag "$local_image" "$regext/openshift/$remote_name_tag"
docker push "$regext/openshift/$remote_name_tag"
