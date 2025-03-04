# Using Argo CD with KubeStellar Core chart

This documents explains how to use KubeStellar Core chart to:

- Deploy an instance of Argo CD using KubeStellar Core chart that automatically lists the WDSes as desitnation clusters
- Automatically crete Argo CD applications, including scenarios examples

## Pre-requisites

The prerequisites are the same of [installing KubeStellar using the Core chart](./core-chart.md#pre-requisites).

The settings described in this document are an extesnion of the KubeStellar Core chart settings described [here](./core-chart.md#kubestellar-core-chart-values).

## Installing Argo CD using KubeStellar Core chart

To enable the installation of Argo CD by the KubeStellar Core chart, use the flag `--set "argo-cd.install=true"`. Besides deploying an instance of Argo CD, KubeStellar Core chart will take care of registering all the WDSes installed by the chart as Argo CD target clusters.

When deploying in an **OpenShift** cluster, add the flag `--set "argo-cd.openshift.enabled=true"`.

When deploying in a **Kubernetes** cluster, use the flag `--set "argo-cd.global.domain=<url>"` to provide the URL for the **nginx** ingress, default `argocd.localtest.me`.

Note that when creating a local **Kubernetes** cluster using our scripts for **Kind** or **k3s**, the **nginx** ingress will be accessible on port `9443`; therefore the Argo CD UI can be accessed at the address `https://argocd.localtest.me:9443`.

The initial password for the `admin` user can be retrieved using the command:

```shell
kubectl get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
```

## Deploying Argo CD applications

KubeStellar Core chart can also be used to automatically deploy Argo CD applications. The example below show an example deploying an application to corresponding to `scenario 1` in [KubeStellar docs](./example-scenarios.md#scenario-1---multi-cluster-workload-deployment-with-kubectl).

```yaml
argo-cd:
  applications: # list of Argo CD applications to be create
  - name: scenario-1 # required, must be unique
    project: default # default: default
    repoURL: https://github.com/kubestellar/kubestellar
    targetRevision: HEAD # default: HEAD
    path: config/argocd/scenario-1
    destinationWDS: wds1
    destinationNamespace: default # default: default
```

Alternatively, the same result can be achieved from Helm CLI by using the followig minimal argument (note that the default values are not explicitely set):

```shell
--set-json='argo-cd.applications=[ \
    { \
      "name": "scenario-1", \
      "repoURL": "https://github.com/kubestellar/kubestellar", \
      "path": "config/argocd/scenario-1", \
      "destinationWDS": wds1" \
    }\
  ]'
```
