#### Add edge-mc's workspaces to Argo CD as clusters
This includes the edge-mc Inventory Management Workspace (IMW) and Workload Management Workspace (WMW).
The procedure for the two workspaces are similar.
Let's take WMW as example:

First, create `kube-system` namespace in the workspace.

Second, make sure necessary apibindings exist in the workspace.
For WMW, we need one for Kubernetes and one for edge-mc's edge API.

Third, make sure the current context uses WMW, then identify the admin.kubeconfig.

The command and output should be similar to
```console
$ argocd cluster add --name wmw --kubeconfig ./admin.kubeconfig workspace.kcp.io/current
WARNING: This will create a service account `argocd-manager` on the cluster referenced by context `workspace.kcp.io/current` with full cluster level privileges. Do you want to continue [y/N]? y
INFO[0001] ServiceAccount "argocd-manager" already exists in namespace "kube-system"
INFO[0001] ClusterRole "argocd-manager-role" updated
INFO[0001] ClusterRoleBinding "argocd-manager-role-binding" updated
Cluster 'https://172.31.31.125:6443/clusters/root:my-org:wmw-turbo' added
```

#### Create Argo CD Applications against edge-mc's IWM
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

#### Create Argo CD Application against edge-mc's WMW
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

