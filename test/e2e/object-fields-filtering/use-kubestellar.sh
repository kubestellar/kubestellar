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
: "Create a bindingpolicy in wds1 to deliver a service to one WEC."
:
kubectl --context wds1 apply -f - <<EOF
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: object-filtering-bindingpolicy
spec:
  clusterSelectors:
  - matchLabels: {"name":"cluster1"}
  downsync:
  - objectSelectors:
    - matchLabels: {"app.kubernetes.io/name":"object-fields-filtering-test"}
EOF

:
: -------------------------------------------------------------------------
: "Deploy the service"
:
kubectl --context wds1 apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  labels:
    app.kubernetes.io/name: object-fields-filtering-test
  name: object-filtering
---
apiVersion: v1
kind: Service
metadata:
  name: hello-world
  namespace: object-filtering
  labels:
    app.kubernetes.io/name: object-fields-filtering-test
spec:
  selector:
    app.kubernetes.io/name: hello-world
  ports:
    - protocol: TCP
      port: 80
      targetPort: 80
  type: NodePort
EOF

:
: -------------------------------------------------------------------------
: "Verify that the service has been created in cluster1"
:
wait-for-cmd 'kubectl --context cluster1 get services -n object-filtering hello-world'
: "SUCCESS: Confirmed service created on cluster1."
