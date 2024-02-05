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

:
: -------------------------------------------------------------------------
: "Create a placement in wds1 to deliver an app to one WEC."
:
kubectl --context wds1 apply -f - <<EOF
apiVersion: control.kubestellar.io/v1alpha1
kind: Placement
metadata:
  name: nginx-singleton-placement
spec:
  wantSingletonReportedState: true
  clusterSelectors:
  - matchLabels: {"name":"cluster1"}
  downsync:
  - objectSelectors:
    - matchLabels: {"app.kubernetes.io/name":"nginx-singleton"}
EOF

:
: -------------------------------------------------------------------------
: "Deploy the app"
:
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

:
: -------------------------------------------------------------------------
: "Verify that manifestworks wrapping the objects have been created in the mailbox namespaces"
: Expect: one header line, one for the nginx namespace, one for the nginx deployment, one for the status agent Deployment
:
if ! wait-for-cmd "expect-cmd-output 'kubectl --context transport1 get manifestworks -n cluster1' 'wc -l | grep -wq 4'"
then
    echo "Failed to see expected manifestworks."
    exit 1
fi

:
: -------------------------------------------------------------------------
: "Verify that the deployment has been created in cluster1"
:
wait-for-cmd 'kubectl --context cluster1 get deployments -n nginx nginx-deployment'
: "Confirmed deployment on cluster1."

:
: -------------------------------------------------------------------------
: "Verify that the status has been returned to wds1"
:
wait-for-cmd '[ "$(kubectl --context wds1 get deployments -n nginx nginx-deployment -o jsonpath="{.status.availableReplicas}")" == 1 ]'
: "SUCCESS: status has been returned to wds1"
