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

set -x # echo so that users can understand what is happening
set -e # exit on error

function wait-for-cmd() {
    local wait_counter
    cmd="$1"
    wait_counter=0
    while ! (eval "$cmd") ; do
        if (($wait_counter > 36)); then
            echo "Failed to ${cmd}."
            exit 1 
        fi
        ((wait_counter += 1))
        sleep 5
    done
}

echo "Create a placement to deliver an app to the WEC in wds1."
echo "-------------------------------------------------------------------------"
kubectl --context wds1 apply -f - <<EOF
apiVersion: edge.kubestellar.io/v1alpha1
kind: Placement
metadata:
  name: nginx-singleton-placement
spec:
  wantSingletonReportedState: true
  clusterSelectors:
  - matchLabels: {"location-group":"edge"}
  downsync:
  - objectSelectors:
    - matchLabels: {"app.kubernetes.io/name":"nginx-singleton"}
EOF

echo "Deploy the app"
echo "-------------------------------------------------------------------------"
kubectl --context wds1 apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  labels:
    app.kubernetes.io/name: nginx-singleton
  name: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  namespace: nginx
  labels:
    app.kubernetes.io/name: nginx-singleton
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

echo "Verify that manifestworks wrapping the objects have been created in the mailbox namespaces"
echo "-------------------------------------------------------------------------"
wait-for-cmd "kubectl --context imbs1 get manifestworks -n cluster1 appsv1-deployment-nginx-nginx-deployment"

echo "Verify that the deployment has been created in cluster1"
echo "-------------------------------------------------------------------------"
wait-for-cmd 'kubectl --context cluster1 get deployments -n nginx nginx-deployment'
echo "Confirmed deployment on cluster1."

echo "Verify that the status has been returned to wds1"
echo "-------------------------------------------------------------------------"
wait-for-cmd '[ "$(kubectl --context wds1 get deployments -n nginx nginx-deployment -o jsonpath="{.status.availableReplicas}")" == 1 ]'
echo "SUCCESS: status has been returned to wds1"

rm out
