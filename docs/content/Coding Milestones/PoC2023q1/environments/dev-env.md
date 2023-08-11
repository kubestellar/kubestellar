---
title: "Development-Environment (dev-env)"
linkTitle: "Development-Environment (dev-env)"
weight: 100
description: >-
 
---

**Mostly under construction - coming soon**

## Hosting KubeStellar in a Kind cluster

### Create a Kind cluster with a port mapping

Create a **Kind** cluster with the `extraPortMappings` for the Ingress
controller, which will listen at port 443 on the one kind node.  We
pick a port number here that does not run afoul of the usual
prohibition of ordinary user processes listening at low port numbers.

```shell
kind create cluster --name ks-host --config=- <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 443
    hostPort: 6443
    protocol: TCP
EOF
```

### Create an nginx Ingress controller with SSL passthrough

Create an `nginx-ingress` with SSL passthrough. Following [Kind NGINX ingress instructions](https://kind.sigs.k8s.io/docs/user/ingress/), we have modified the YAML at https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml to include the `--enable-ssl-passthrough=true` argument. [This](https://raw.githubusercontent.com/kubestellar/kubestellar/main/user/yaml/kind-nginx-ingress-with-SSL-passthrough.yaml) is the link to our raw modified nginx controller deployment YAML.

```shell
kubectl apply -f https://raw.githubusercontent.com/kubestellar/kubestellar/main/user/yaml/kind-nginx-ingress-with-SSL-passthrough.yaml
```

Wait for the ingress to be ready:

```shell
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=90s
```

### Load a locally-built container image into the kind cluster

Remember that [you can do
this](https://kind.sigs.k8s.io/docs/user/quick-start#loading-an-image-into-your-cluster).
