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

set -x # echo so users can understand what is happening
set -e # exit on error

SRC_DIR="$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)"
COMMON_SRCS="${SRC_DIR}/../common"
source "${COMMON_SRCS}/setup-shell.sh"

# Test cases for the update/delete of the bindingpolicy and the workload objects.
# This test script should be executed after a successful execution of the use-kubestellar.sh script, located in the current directory.

#
#  Update of the managedcluster object in the OCM hub should update the bindings
#
: -------------------------------------------------------------------------
: Test update of the managedcluster:
: Change the location-group label on one of the managed cluster and verify the downsynced
: objects are properly updated 
:
kubectl --context imbs1 label managedcluster cluster2 location-group=blah name=cluster2 --overwrite
wait-for-cmd '(($(kubectl --context cluster1 get ns nginx | wc -l) == 2))'
wait-for-cmd '(($(kubectl --context cluster2 get ns nginx | wc -l) == 0))'
kubectl --context imbs1 label managedcluster cluster2 location-group=edge name=cluster2 --overwrite
wait-for-cmd '(($(kubectl --context cluster1 get ns nginx | wc -l) == 2))'
wait-for-cmd '(($(kubectl --context cluster2 get ns nginx | wc -l) == 2))'
:
: Test passed


#
#  Update of the workload object on WDS should update the object on the WECs
#
: -------------------------------------------------------------------------
: Test update of the workload object
: Increase the number of replicas from 1 to 2, and verify the update on the WECs
:
kubectl --context wds1 patch deployment -n nginx nginx-deployment -p '{"spec":{"replicas": 2}}'
wait-for-cmd '[ "$(kubectl --context cluster1 get deployment -n nginx nginx-deployment -o jsonpath="{.status.availableReplicas}")" == 2 ]'
wait-for-cmd '[ "$(kubectl --context cluster2 get deployment -n nginx nginx-deployment -o jsonpath="{.status.availableReplicas}")" == 2 ]'
:
: Test passed


#
#  Changing the bindingpolicy objectSelector to no longer match should delete the object from the WECs
#
: -------------------------------------------------------------------------
: Change the bindingpolicy objectSelector to no longer match the labels
: Verify that the object is deleted from the WECs
:
kubectl --context wds1 patch bindingpolicy nginx-bindingpolicy --type=merge -p '{"spec": {"downsync": [{"objectSelectors": [{"matchLabels": {"app.kubernetes.io/name": "invalid"}}]}]}}'
wait-for-cmd '(($(kubectl --context cluster1 get ns nginx | wc -l) == 0))'
wait-for-cmd '(($(kubectl --context cluster2 get ns nginx | wc -l) == 0))'
:
: Test passed


#
#  Changing the bindingpolicy objectSelector to match should create the object on the WECs
#
: -------------------------------------------------------------------------
: Change the bindingpolicy objectSelector to match the labels
: Verify that the object is created on the WECs
:
kubectl --context wds1 patch bindingpolicy nginx-bindingpolicy --type=merge -p '{"spec": {"downsync": [{"objectSelectors": [{"matchLabels": {"app.kubernetes.io/name": "nginx"}}]}]}}'
wait-for-cmd kubectl --context cluster1 get deployment -n nginx nginx-deployment
wait-for-cmd kubectl --context cluster2 get deployment -n nginx nginx-deployment
:
: Test passed


#
#  Delete of the bindingpolicy object should delete the object on the WECs
#
: -------------------------------------------------------------------------
: Delete the bindingpolicy
: Verify that the object is deleted from the WECs
:
kubectl --context wds1 delete bindingpolicy nginx-bindingpolicy
wait-for-cmd '(($(kubectl --context cluster1 get ns nginx | wc -l) == 0))'
wait-for-cmd '(($(kubectl --context cluster2 get ns nginx | wc -l) == 0))'
:
: Test passed


#
#  Delete of the overlapping bindingpolicy object should not delete the object on the WECs
#
: -------------------------------------------------------------------------
: Create an object and two bindingpolicies that match the object '(overlapping bindingpolicies)'
: Verify that by deleting one of the bindingpolicies the object stays in the WEC
:
kubectl --context wds1 apply -f - <<EOF
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: nginx-bindingpolicy
spec:
  clusterSelectors:
  - matchLabels: {"location-group":"edge"}
  downsync:
  - objectSelectors:
    - matchLabels: {"app.kubernetes.io/name":"nginx"}
EOF

kubectl --context wds1 apply -f - <<EOF
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: nginx-bindingpolicy-2
spec:
  clusterSelectors:
  - matchLabels: {"location-group":"edge"}
  downsync:
  - objectSelectors:
    - matchLabels: {"app.kubernetes.io/name":"nginx"}
EOF

wait-for-cmd kubectl --context cluster1 get deployment -n nginx nginx-deployment
wait-for-cmd kubectl --context cluster2 get deployment -n nginx nginx-deployment

kubectl --context wds1 delete bindingpolicy nginx-bindingpolicy-2
sleep 5 #give it a chance to fail
wait-for-cmd kubectl --context cluster1 get deployment -n nginx nginx-deployment
wait-for-cmd kubectl --context cluster2 get deployment -n nginx nginx-deployment
:
: Test passed


#
#  Delete of the workload object on WDS should delete the object on the WECs
#
: -------------------------------------------------------------------------
: Delete the workload object
: Verify that the object is deleted from the WECs
:
kubectl --context wds1 delete deployment -n nginx nginx-deployment
wait-for-cmd '(($(kubectl --context cluster1 get deployment -n nginx nginx-deployment | wc -l) == 0))'
wait-for-cmd '(($(kubectl --context cluster2 get deployment -n nginx nginx-deployment | wc -l) == 0))'
:
: Test passed


#
#  Re-create of the workload object on WDS should re-create the object on the WECs
#
: -------------------------------------------------------------------------
: Re-create the workload object
: Verify that the object is created on the WECs
:
kubectl --context wds1 apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  namespace: nginx
  labels:
    app.kubernetes.io/name: nginx
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: public.ecr.aws/nginx/nginx:latest 
        ports:
        - containerPort: 80
EOF

wait-for-cmd kubectl --context cluster1 get deployment -n nginx nginx-deployment
wait-for-cmd kubectl --context cluster2 get deployment -n nginx nginx-deployment
:
: Test passed
: -------------------------------------------------------------------------
