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

# Transport deployment script assumes it runs using the kubeflex hosting cluster context.

export WDS_NAME="$1" ## first argument is WDS name
export IMBS_NAME="$2" ## second argument is IMBS name
export TRANSPORT_CONTROLLER_IMAGE="${3:-ghcr.io/kubestellar/ocm-transport-plugin/transport-controller:0.1.0}" ## third argument is transport controller image
export IMAGE_PULL_POLICY="${IMAGE_PULL_POLICY:=Always}"

# generate from template and env vars and then apply a configmap and a deployment for transport-controller
kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: transport-controller-config
  namespace: ${WDS_NAME}-system
data:
  get-kubeconfig.sh: |
    #!/bin/bash
    # this script receives a ControlPlane name and a parameter
    # that determines whether to extract the ControlPlane's in-cluster kubeconfig
    # or the external kubeconfig (if set to "true", the first will be retrieved).
    # The function returns the requested kubeconfig's content in base64.
    # it assumes the kubeconfig context is set to access the hosting cluster.

    cpname="\$1"
    get_incluster_key="\$2"

    key=""
    if [[ "\$get_incluster_key" == "true" ]]; then
      key=\$(kubectl get controlplane \$cpname -o=jsonpath='{.status.secretRef.inClusterKey}');
    else
      key=\$(kubectl get controlplane \$cpname -o=jsonpath='{.status.secretRef.key}');
    fi

    # get secret details
    secret_name=\$(kubectl get controlplane \$cpname -o=jsonpath='{.status.secretRef.name}')
    secret_namespace=\$(kubectl get controlplane \$cpname -o=jsonpath='{.status.secretRef.namespace}')

    # get the kubeconfig in base64
    kubectl get secret \$secret_name -n \$secret_namespace -o=jsonpath="{.data.\$key}"
---

apiVersion: apps/v1
kind: Deployment
metadata:
  name: transport-controller
  namespace: ${WDS_NAME}-system
spec:
  replicas: 1
  selector:
    matchLabels:
      name: transport-controller
  template:
    metadata:
      labels:
        name: transport-controller
    spec:
      serviceAccountName: kubestellar-controller-manager
      initContainers:
      - name: setup-wds-kubeconfig
        image: quay.io/kubestellar/kubectl:1.27.8
        imagePullPolicy: Always
        command: [ "bin/sh", "-c", "sh /mnt/config/get-kubeconfig.sh ${WDS_NAME} true | base64 -d > /mnt/shared/wds-kubeconfig"]
        volumeMounts:
        - name: config-volume
          mountPath: /mnt/config
        - name: shared-volume
          mountPath: /mnt/shared
      - name: setup-imbs-kubeconfig
        image: quay.io/kubestellar/kubectl:1.27.8
        imagePullPolicy: Always
        command: [ "bin/sh", "-c", "sh /mnt/config/get-kubeconfig.sh ${IMBS_NAME} true | base64 -d > /mnt/shared/transport-kubeconfig"]
        volumeMounts:
        - name: config-volume
          mountPath: /mnt/config
        - name: shared-volume
          mountPath: /mnt/shared
      containers:
        - name: transport-controller
          image: ${TRANSPORT_CONTROLLER_IMAGE}
          imagePullPolicy: ${IMAGE_PULL_POLICY}
          args:
          - --transport-kubeconfig=/mnt/shared/transport-kubeconfig
          - --wds-kubeconfig=/mnt/shared/wds-kubeconfig
          - --wds-name=${WDS_NAME}
          - -v=4
          volumeMounts:
          - name: shared-volume
            mountPath: /mnt/shared
            readOnly: true
      volumes:
      - name: shared-volume
        emptyDir: {}
      - name: config-volume
        configMap:
          name: transport-controller-config
          defaultMode: 0744
EOF
