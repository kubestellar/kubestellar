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

set -e # exit on error

echo "Create a placement to deliver an app to all clusters in wds1."
echo "This placement configuration determines where to deploy the workload by using the label selector expressions found in clusterSelectors. It also specifies what to deploy through the downsync.objectSelectors expressions. When there are multiple matchLabels expressions, they are combined using a logical AND operation. Conversely, when there are multiple objectSelectors, they are combined using a logical OR operation."
echo "-------------------------------------------------------------------------"
kubectl --context wds1 apply -f - <<EOF
apiVersion: edge.kubestellar.io/v1alpha1
kind: Placement
metadata:
  name: nginx-placement
spec:
  clusterSelectors:
  - matchLabels: {"location-group":"edge"}
  downsync:
  - objectSelectors:
    - matchLabels: {"app.kubernetes.io/name":"nginx"}
EOF

echo "Deploy the app"
echo "-------------------------------------------------------------------------"
kubectl --context wds1 apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  labels:
    app.kubernetes.io/name: nginx
  name: nginx
---
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

echo "Verify that manifestworks wrapping the objects have been created in the mailbox namespaces"
echo "-------------------------------------------------------------------------"
kubectl --context imbs1 get manifestworks -n cluster1 | tee out 
kubectl --context imbs1 get manifestworks -n cluster2 | tee -a out
if (("$(wc -l < out)" != "6")); then
  echo "Failed to see expected manifestworks."
  exit 1
fi

echo "Verify that the deployment has been created in both clusters"
echo "-------------------------------------------------------------------------"
function wait_for_deployment() {
  # kubectl wait can't be used on resources that haven't been created, so we need to spin
  cluster=$1
  echo "Waiting for deployment on $cluster"
  waitCounter=0
  while (($(kubectl --context $cluster wait deployment nginx-deployment -n nginx --for=condition=available --timeout=180s 2>/dev/null | grep -c "condition met") < 1)); do
    if (($waitCounter > 180)); then
      echo "Failed to observe deployment on ${cluster}."
      exit 1 
    fi
    ((waitCounter += 1))
    sleep 1
  done
}
wait_for_deployment cluster1
echo "Waiting for deployment on cluster2"
wait_for_deployment cluster2
echo "SUCCESS: confirmed deployments on both cluster1 and cluster2."
