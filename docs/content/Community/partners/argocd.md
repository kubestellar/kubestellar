This document explains how to add KubeStellar's 'workspaces' as Argo CD's 'clusters'.

### Add KubeStellar's workspaces to Argo CD as clusters
As of today, the 'workspaces', aka 'logical clusters' used by KubeStellar are not identical with ordinary Kubernetes clusters.
Thus, in order to add them as Argo CD's 'clusters', there are a few more steps to take.

For KubeStellar's Inventory Management Workspace (IMW) and Workload Management Workspace (WMW).
The steps are similar.
Let's take WMW as an example:

1. Create `kube-system` namespace in the workspace.

2. Make sure necessary apibindings exist in the workspace.
For WMW, we need one for Kubernetes and one for KubeStellar's edge API.

3. Exclude `ClusterWorkspace` from discovery and sync.

```shell
kubectl -n argocd edit cm argocd-cm
```

Make sure `resource.exclusions` exists in the `data` field of the `argocd-cm` configmap as follows:
```yaml
data:
  resource.exclusions: |
    - apiGroups:
      - "tenancy.kcp.io"
      kinds:
      - "ClusterWorkspace"
      clusters:
      - "*"
```

Restart the Argo CD server.
```shell
kubectl -n argocd rollout restart deployment argocd-server
```

Argo CD's documentation mentions this feature as [Resource Exclusion/Inclusion](https://argo-cd.readthedocs.io/en/stable/operator-manual/declarative-setup/#resource-exclusioninclusion).

4. Make sure the current context uses WMW, then identify the admin.kubeconfig.

The command and output should be similar to
```console
$ argocd cluster add --name wmw --kubeconfig ./admin.kubeconfig workspace.kcp.io/current
WARNING: This will create a service account `argocd-manager` on the cluster referenced by context `workspace.kcp.io/current` with full cluster level privileges. Do you want to continue [y/N]? y
INFO[0001] ServiceAccount "argocd-manager" already exists in namespace "kube-system"
INFO[0001] ClusterRole "argocd-manager-role" updated
INFO[0001] ClusterRoleBinding "argocd-manager-role-binding" updated
Cluster 'https://172.31.31.125:6443/clusters/root:my-org:wmw-turbo' added
```

### Create Argo CD Applications
Once KubeStellar's workspaces are added, Argo CD Applications can be created as normal.
There are a few examples listed [here](https://github.com/edge-experiments/gitops-source/tree/main/edge-mc),
and the commands to use the examples are listed as follows.

#### Create Argo CD Applications against KubeStellar's IMW
Create two Locations. The command and output should be similar to
```console
$ argocd app create locations \
--repo https://github.com/edge-experiments/gitops-source.git \
--path edge-mc/locations/ \
--dest-server https://172.31.31.125:6443/clusters/root:imw-turbo \
--sync-policy automated
application 'locations' created
```

Create two SyncTargets. The command and output should be similar to
```console
$ argocd app create synctargets \
--repo https://github.com/edge-experiments/gitops-source.git \
--path edge-mc/synctargets/ \
--dest-server https://172.31.31.125:6443/clusters/root:imw-turbo \
--sync-policy automated
application 'synctargets' created
```

#### Create Argo CD Application against KubeStellar's WMW
Create a Namespace. The command and output should be similar to
```console
$ argocd app create namespace \
--repo https://github.com/edge-experiments/gitops-source.git \
--path edge-mc/namespaces/ \
--dest-server https://172.31.31.125:6443/clusters/root:my-org:wmw-turbo \
--sync-policy automated
application 'namespace' created
```

Create a Deployment for 'cpumemload'. The command and output should be similar to
```console
$ argocd app create cpumemload \
--repo https://github.com/edge-experiments/gitops-source.git \
--path edge-mc/workloads/cpumemload/ \
--dest-server https://172.31.31.125:6443/clusters/root:my-org:wmw-turbo \
--sync-policy automated
application 'cpumemload' created
```

Create an EdgePlacement. The command and output should be similar to
```console
$ argocd app create edgeplacement \
--repo https://github.com/edge-experiments/gitops-source.git \
--path edge-mc/placements/ \
--dest-server https://172.31.31.125:6443/clusters/root:my-org:wmw-turbo \
--sync-policy automated
application 'edgeplacement' created
```

# Other Resources
Medium - [Sync 10,000 ArgoCD Applications in One Shot](https://medium.com/itnext/sync-10-000-argo-cd-applications-in-one-shot-bfcda04abe5b)<br/>
Medium - [Sync 10,000 ArgoCD Applications in One Shot, by Yourself](https://medium.com/@filepp/how-to-sync-10-000-argo-cd-applications-in-one-shot-by-yourself-9e389ab9e8ad)<br/>
Medium - [GitOpsCon - here we come](https://medium.com/@clubanderson/gitopscon-here-we-come-9a8b8ffe2a33)<br/>
### ArgoCD Scale Experiment - KubeStellar Community Demo Day
![type:video](https://www.youtube.com/embed/7XuEJF7--Sc)
### GitOpsCon 2023 - A Quantitative Study on Argo Scalability - Andrew Anderson & Jun Duan, IBM
![type:video](https://www.youtube.com/embed/PB3OTXDjFjg)
