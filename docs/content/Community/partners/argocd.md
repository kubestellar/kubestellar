This document explains how to add KubeStellar's 'workspaces' as Argo CD's 'clusters'.

### Add KubeStellar's workspaces to Argo CD as clusters
As of today, the 'workspaces', aka 'logical clusters' used by KubeStellar are not identical with ordinary Kubernetes clusters.
Thus, in order to add them as Argo CD's 'clusters', there are a few more steps to take.

For KubeStellar's Inventory Management Workspace (IMW) and Workload Management Workspace (WMW).
The steps are similar.
Let's take WMW as an example:

<ol>
<li>Create `kube-system` namespace in the workspace.</li>
<li>Make sure necessary apibindings exist in the workspace. 
For WMW, we need one for Kubernetes and one for KubeStellar's edge API.</li>
<li>Exclude `ClusterWorkspace` from discovery and sync.

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
</li>
<li>Make sure the current context uses WMW, then identify the admin.kubeconfig.
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
There are a few examples listed [here](https://github.com/edge-experiments/gitops-source/tree/main/kubestellar),
and the commands to use the examples are listed as follows.

#### Create Argo CD Applications against KubeStellar's IMW
Create two Locations. The command and output should be similar to
```console
$ argocd app create locations \
--repo https://github.com/edge-experiments/gitops-source.git \
--path kubestellar/locations/ \
--dest-server https://172.31.31.125:6443/clusters/root:imw-turbo \
--sync-policy automated
application 'locations' created
```

Create two SyncTargets. The command and output should be similar to
```console
$ argocd app create synctargets \
--repo https://github.com/edge-experiments/gitops-source.git \
--path kubestellar/synctargets/ \
--dest-server https://172.31.31.125:6443/clusters/root:imw-turbo \
--sync-policy automated
application 'synctargets' created
```

#### Create Argo CD Application against KubeStellar's WMW
Create a Namespace. The command and output should be similar to
```console
$ argocd app create namespace \
--repo https://github.com/edge-experiments/gitops-source.git \
--path kubestellar/namespaces/ \
--dest-server https://172.31.31.125:6443/clusters/root:my-org:wmw-turbo \
--sync-policy automated
application 'namespace' created
```

Create a Deployment for 'cpumemload'. The command and output should be similar to
```console
$ argocd app create cpumemload \
--repo https://github.com/edge-experiments/gitops-source.git \
--path kubestellar/workloads/cpumemload/ \
--dest-server https://172.31.31.125:6443/clusters/root:my-org:wmw-turbo \
--sync-policy automated
application 'cpumemload' created
```

Create an EdgePlacement. The command and output should be similar to
```console
$ argocd app create edgeplacement \
--repo https://github.com/edge-experiments/gitops-source.git \
--path kubestellar/placements/ \
--dest-server https://172.31.31.125:6443/clusters/root:my-org:wmw-turbo \
--sync-policy automated
application 'edgeplacement' created
```
</li>
</ol>

# Other Resources
Medium - [Sync 10,000 ArgoCD Applications in One Shot](https://medium.com/itnext/sync-10-000-argo-cd-applications-in-one-shot-bfcda04abe5b)<br/>
Medium - [Sync 10,000 ArgoCD Applications in One Shot, by Yourself](https://medium.com/@filepp/how-to-sync-10-000-argo-cd-applications-in-one-shot-by-yourself-9e389ab9e8ad)<br/>
Medium - [GitOpsCon - here we come](https://medium.com/@clubanderson/gitopscon-here-we-come-9a8b8ffe2a33)<br/>
### ArgoCD Scale Experiment - KubeStellar Community Demo Day
<p align=center>
<div id="spinner1">
  <img width="140" height="140" src="../../../images/spinner.gif" class="centerImage">
</div>
<iframe class="centerImage" id="embed1" width="720" height="400" src="https://www.youtube.com/embed/7XuEJF7--Sc?start=90" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen style="visibility:hidden;" onload= "document.getElementById('spinner1').style.display='none';document.getElementById('embed1').style.visibility='visible';document.getElementById('embed1').width='720';document.getElementById('embed1').height='400';"></iframe>
</p>

### GitOpsCon 2023 - A Quantitative Study on Argo Scalability - Andrew Anderson & Jun Duan, IBM
<p align=center>
<div id="spinner2">
  <img width="140" height="140" src="../../../images/spinner.gif" class="centerImage">
</div>
<iframe class="centerImage" id="embed2" width="0" height="0" src="https://www.youtube.com/embed/PB3OTXDjFjg" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen style="visibility:hidden;" onload= "document.getElementById('spinner2').style.display='none';document.getElementById('embed2').style.visibility='visible';document.getElementById('embed2').width='720';document.getElementById('embed2').height='400';"></iframe>
</p>

### ArgoCD and KubeStellar in the news
<p align=center>
<div id="spinner3">
  <img width="140" height="140" src="../../../images/spinner.gif" class="centerImage">
</div>
<iframe class="centerImage" id="embed3" src="https://www.linkedin.com/embed/feed/update/urn:li:share:7031032280722632704" scrolling=no height="0" width="0" frameborder="0" allowfullscreen="" title="Embedded post" style="visibility:hidden;" onload= "document.getElementById('spinner3').style.display='none';document.getElementById('embed3').style.visibility='visible';document.getElementById('embed3').width='740';document.getElementById('embed3').height='400';"></iframe>
</p>
</br>
<p align=center>
<div id="spinner4">
  <img width="140" height="140" src="../../../images/spinner.gif" class="centerImage">
</div>
<iframe class="centerImage" id="embed4" src="https://www.linkedin.com/embed/feed/update/urn:li:share:7046166635367268352" scrolling=no height="0" width="0" frameborder="0" allowfullscreen="" title="Embedded post" style="visibility:hidden;" onload="document.getElementById('spinner4').style.display='none';document.getElementById('embed4').style.visibility='visible';document.getElementById('embed4').width='740';document.getElementById('embed4').height='400';"></iframe>
</p>
</br>
<p align=center>
<div id="spinner5">
  <img width="140" height="140" src="../../../images/spinner.gif" class="centerImage">
</div>
<iframe class="centerImage" id="embed5" src="https://www.linkedin.com/embed/feed/update/urn:li:share:7060337925300838400" scrolling=no height="0" width="0" frameborder="0" allowfullscreen="" title="Embedded post" style="visibility:hidden;" onload="document.getElementById('spinner5').style.display='none';document.getElementById('embed5').style.visibility='visible';document.getElementById('embed5').width='740';document.getElementById('embed5').height='400';"></iframe>
</p>
</br>
<p align=center>
<div id="spinner6">
  <img width="140" height="140" src="../../../images/spinner.gif" class="centerImage">
</div>
<iframe class="centerImage" id="embed6" src="https://www.linkedin.com/embed/feed/update/urn:li:ugcPost:7074718212461899776" scrolling=no height="0" width="0" frameborder="0" allowfullscreen="" title="Embedded post" style="visibility:hidden;" onload="document.getElementById('spinner6').style.display='none';document.getElementById('embed6').style.visibility='visible';document.getElementById('embed6').width='740';document.getElementById('embed6').height='400';"></iframe>
</p>

<style type="text/css">
.centerImage
{
 display: block;
 margin: auto;
}
</style>