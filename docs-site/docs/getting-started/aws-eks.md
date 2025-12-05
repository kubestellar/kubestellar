---
sidebar_position: 4
title: KubeStellar on AWS EKS
description: Production-ready steps to deploy KubeStellar on AWS EKS with optional WEC clusters and verification.
---

Last updated: 2025 • Author: DevRhylme Foundation / Rishi Mondal

## Overview

This guide installs KubeStellar on AWS EKS (Kubernetes 1.34) following the existing docs style. It covers a host EKS cluster running KubeStellar (ITS + WDS), and optional WECs (Workload Execution Clusters) registered to KubeStellar.

- Prefer a local/dev install? See Getting Started → Installation.
- Prefer a kubectl plugin workflow? See CLI Reference and plugin notes.

## Architecture

```
                    ┌──────────────────────────┐
                    │     Host EKS Cluster     │
                    │  (runs KubeStellar Core) │
                    └────────────┬─────────────┘
                                 │
                     KubeFlex / OCM Hub (ITS)
                                 │
              ┌──────────────────┴──────────────────┐
              │                                     │
      Workload Description Space (WDS1)      Workload Description Space (WDS2)
              │                                     │
              └──────────────────┬──────────────────┘
                                 │
                       Binding Policies / Sync
                                 │
          ┌──────────────────────┴────────────────────────┐
          │                                               │
  WEC Cluster 1 (EKS 1.34)                        WEC Cluster 2 (EKS 1.34)
```

## Prerequisites

### AWS

- Permissions: EC2, EKS, IAM, VPC, CloudFormation
- Region: `us-east-1` recommended
- Networking: IPv4 (public or private subnets)
- Internet egress for images & Helm charts

Minimum quotas:
- vCPU: 12
- Elastic IPs: 4
- Target Groups: 5
- NLBs: 2

### Local machine

- Linux or macOS
- kubectl (latest)
- eksctl (≥ 0.197 for Kubernetes 1.34)
- AWS CLI v2
- Helm v3
- kflex (latest)
- clusteradm (OCM) (latest)

Install tooling

```bash
# AWS CLI
curl -sSLO https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip
unzip -q awscli-exe-linux-x86_64.zip && sudo ./aws/install

# kubectl (latest)
curl -sSLO "https://dl.k8s.io/release/$(curl -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl && sudo mv kubectl /usr/local/bin/

# eksctl
curl -sSL "https://github.com/weaveworks/eksctl/releases/latest/download/eksctl_$(uname -s)_amd64.tar.gz" \
| tar xz -C /tmp && sudo mv /tmp/eksctl /usr/local/bin

# Helm
curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# KubeFlex CLI
curl -fsSL https://github.com/kubestellar/kubeflex/releases/download/v0.7.4/kflex_0.7.4_linux_amd64.tar.gz \
| tar xz && sudo mv kflex /usr/local/bin/

# clusteradm (OCM)
curl -fsSL https://raw.githubusercontent.com/open-cluster-management-io/clusteradm/main/install.sh | bash
```

Configure AWS

```bash
aws configure
# Region: us-east-1, Output: json
aws sts get-caller-identity
```

## Create Host EKS Cluster (Kubernetes 1.34)

```bash
cat > kubestellar-host-cluster.yaml <<'EOF'
apiVersion: eksctl.io/v1alpha5
kind: ClusterConfig

metadata:
  name: kubestellar-host
  region: us-east-1
  version: "1.34"

kubernetesNetworkConfig:
  ipFamily: IPv4

iam:
  withOIDC: true

managedNodeGroups:
  - name: ng-1
    instanceType: t3.large
    desiredCapacity: 3
    minSize: 2
    maxSize: 4
    volumeSize: 50
    amiFamily: AmazonLinux2023
    privateNetworking: false

addons:
  - name: vpc-cni
    version: latest
  - name: kube-proxy
    version: latest
  - name: coredns
    version: latest
EOF

eksctl create cluster -f kubestellar-host-cluster.yaml
aws eks update-kubeconfig --name kubestellar-host --region us-east-1
kubectl get nodes
```

## Install Ingress (NGINX)

```bash
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update

helm install ingress-nginx ingress-nginx/ingress-nginx \
  --namespace ingress-nginx \
  --create-namespace \
  --version 4.12.1 \
  --set controller.extraArgs.enable-ssl-passthrough="" \
  --set controller.service.type=LoadBalancer \
  --set controller.service.annotations."service\.beta\.kubernetes\.io/aws-load-balancer-type"="nlb" \
  --set controller.service.annotations."service\.beta\.kubernetes\.io/aws-load-balancer-nlb-target-type"="instance" \
  --set controller.service.annotations."service\.beta\.kubernetes\.io/aws-load-balancer-scheme"="internet-facing"

kubectl get svc -n ingress-nginx ingress-nginx-controller
```

## Install KubeStellar Core

```bash
export KUBESTELLAR_VERSION=0.27.2

helm upgrade --install ks-core \
  oci://ghcr.io/kubestellar/kubestellar/core-chart \
  --version $KUBESTELLAR_VERSION \
  --set-json='ITSes=[{"name":"its1"}]' \
  --set-json='WDSes=[{"name":"wds1"},{"name":"wds2","type":"host"}]' \
  --timeout 24h
```

## Create Workload Execution Clusters (WECs)

WEC 1 (cluster1)

```bash
cat > cluster1.yaml <<'EOF'
apiVersion: eksctl.io/v1alpha5
kind: ClusterConfig

metadata:
  name: cluster1
  region: us-east-1
  version: "1.34"

managedNodeGroups:
  - name: ng-1
    instanceType: t3.medium
    desiredCapacity: 2
EOF

eksctl create cluster -f cluster1.yaml
```

WEC 2 (cluster2)

```bash
cat > cluster2.yaml <<'EOF'
apiVersion: eksctl.io/v1alpha5
kind: ClusterConfig

metadata:
  name: cluster2
  region: us-east-1
  version: "1.34"

managedNodeGroups:
  - name: ng-1
    instanceType: t3.medium
    desiredCapacity: 2
EOF

eksctl create cluster -f cluster2.yaml
```

## Register WECs with KubeStellar

Get join token from ITS

```bash
joincmd=$(clusteradm --context its1 get token | awk '/clusteradm join/ {print}')
```

Register cluster1

```bash
${joincmd/<cluster_name>/cluster1} \
  --context cluster1 \
  --singleton \
  --force-internal-endpoint-lookup \
  --wait-timeout 240s
```

Register cluster2

```bash
${joincmd/<cluster_name>/cluster2} \
  --context cluster2 \
  --singleton \
  --force-internal-endpoint-lookup \
  --wait-timeout 240s
```

Accept and label

```bash
clusteradm --context its1 accept --clusters cluster1
clusteradm --context its1 accept --clusters cluster2

kubectl --context its1 label managedcluster cluster1 location-group=edge --overwrite
kubectl --context its1 label managedcluster cluster2 location-group=edge --overwrite
```

## Deploy a Test App via KubeStellar

Namespace and deployment

```bash
kubectl apply -f - <<'EOF'
apiVersion: v1
kind: Namespace
metadata:
  name: test-app
EOF

kubectl apply -f - <<'EOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-test
  namespace: test-app
spec:
  replicas: 2
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
        image: nginx:latest
EOF
```

BindingPolicy to target WECs

```bash
kubectl apply -f - <<'EOF'
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: nginx-test-policy
  namespace: test-app
spec:
  clusterSelectors:
  - matchLabels:
      location-group: edge
  downsync:
  - objectSelectors:
    - matchLabels:
        app: nginx
EOF
```

Verify

```bash
kubectl --context cluster1 get deploy -n test-app
kubectl --context cluster2 get deploy -n test-app
```

## kubectl Plugin (optional)

- Krew (using release manifest `kubestellar.yaml`):

```bash
kubectl krew install --manifest=kubestellar.yaml
kubectl kubestellar --help
```

- Python-based alias (local dev):

```bash
uv tool install .   # or pipx install .
kubectl a2a --help
```

Note: kubectl discovers executables named `kubectl-<name>` on PATH.

## Troubleshooting

```bash
# Registration
kubectl --context its1 get managedclusters

# Agent issues
kubectl --context cluster1 -n open-cluster-management-agent get pods
kubectl --context cluster1 get csr

# KubeStellar components
kubectl get controlplanes -A
kubectl logs -n kubeflex-system -l app=kubeflex-controller-manager
```

## Cleanup

```bash
eksctl delete cluster --name cluster1 --region us-east-1
eksctl delete cluster --name cluster2 --region us-east-1
eksctl delete cluster --name kubestellar-host --region us-east-1
```

