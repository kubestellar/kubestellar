# Using Argo CD with KubeStellar Core chart

This documents explains how to use the KubeStellar core Helm chart to:

- deploy Argo CD in KubeFlex hosting cluster;
- register every WDS as a target cluster in Argo CD; and
- create Argo CD applications as specified by the chart values.

## Pre-requisites

The prerequisites are the same as [installing KubeStellar using the Core chart](core-chart.md#pre-requisites).

The settings described in this document are an extesnion of the KubeStellar Core chart settings described [here](core-chart.md#kubestellar-core-chart-values).

## Installing Argo CD using KubeStellar Core chart

To enable the installation of Argo CD by the KubeStellar Core chart, use the flag `--set argocd.install=true`. Besides deploying an instance of Argo CD, KubeStellar Core chart will take care of registering all the WDSes installed by the chart as Argo CD target clusters.

When deploying in an **OpenShift** cluster, add the flag `--set argocd.openshift.enabled=true`.

When deploying in a **Kubernetes** cluster, use the flag `--set argocd.global.domain=<url>` to provide the URL for the **nginx** ingress, which defaults to `argocd.localtest.me`.

Note that when creating a local **Kubernetes** cluster using our scripts for **Kind** or **k3s**, the **nginx** ingress will be accessible on host port `9443`; therefore the Argo CD UI can be accessed at the address `https://argocd.localtest.me:9443`.

The initial password for the `admin` user can be retrieved using the command:

```shell
kubectl get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
```

## Deploying Argo CD applications

The KubeStellar Core chart can also be used to deploy Argo CD applications as specified by chart values. The example below shows the relevant fragment of the chart values that could be used for deploying an application corresponding to `scenario-6` in [KubeStellar docs](example-scenarios.md#scenario-6---multi-cluster-workload-deployment-of-app-with-serviceaccount-with-argocd).

```yaml
argocd:
  applications: # list of Argo CD applications to be create
  - name: scenario-6 # required, must be unique
    project: default # default: default
    repoURL: https://github.com/pdettori/sample-apps.git
    targetRevision: HEAD # default: HEAD
    path: nginx
    destinationWDS: wds1
    destinationNamespace: nginx-sa # default: default
    syncPolicy: auto # default: manual
```

Alternatively, the same result can be achieved from Helm CLI by using the followig minimal argument (note that the default values are not explicitely set):

```shell
--set-json='argocd.applications=[ { "name": "scenario-6", "repoURL": "https://github.com/pdettori/sample-apps.git", "path": "nginx", "destinationWDS": "wds1", "destinationNamespace": "nginx-sa" } ]'
```
