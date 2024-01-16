# KubeStellar 0.20


# TL;DR

**NOTE**: This is unmaintained material that was observed to work around Jan 10, 2024.

KubeStellar 0.20  supports multi-cluster deployment of Kubernetes objects, controlled by a 
simple placement policy and deploying Kubernetes objects in their native format.  It uses OCM as 
transport, with standard OCM agents (Klusterlet). We show examples of deploying workloads to 
multi-cluster with kubectl, helm and ArgoCD using a simple label-selectors-based placement policy.


## Supported Features:

1. *Multi-cluster Deployment:* Kubernetes objects are deployed across multiple clusters, controlled by a 
straightforward placement policy.
2. *Pure-Kube User Experience:* Deployment of non-wrapped objects is handled in a pure Kubernetes manner.
3. *Object Management via WDS:* Creation, update, and deletion of objects in managed clusters are performed from WDS.
4. *OCM as Transport:* The Open Cluster Management (OCM) is used as transport, with standard OCM agents (Klusterlet).
5. *Multi-WDS and single OCM Shard:* Multiple WDSs and a single OCM shard are supported.
6. *Resiliency:* All components are running in Kubernetes, ensuring continued operation even after restarts of any component.
7. *Re-evaluation of Objects:* Existing objects are re-evaluated when a new placement is added or updated.
8. *Object Removal:* Objects are removed from clusters when the placement that led to their deployment on
 those clusters is deleted or updated and the what or where no longer match.
9. *Dynamic Handling of APIs:* Dynamically start/stop informers when adding/removing CRDs.
10. *Simplified setup:* Just 3 commands to get a fully functional setup (`kflex init`, `kflex create imbs`, `kflex create wds`)
11. *OpenShift Support:* Same commands to set it up. All components have been tested in OpenShift, 
including OCM Klusterlet for the WECs.
12. *Singleton Status* Addressed by the status controller in KubeStellar 0.20 and the [Status Add-On for OCM](link to be added)

## To be supported

1. Status summarization
2. Customization
3. OCM sharding
4. Upsync
5. "Pluggable Transport" 


## Prereqs
- go version 1.20 and up
- make
- kubectl
- docker (or compatible docker engine that works with kind)
- kind
- helm

## Setup

1. Install [KubeFlex](https://github.com/kubestellar/kubeflex/tree/main#installation). 
Make sure you have at least version 0.3.3 or higher. To upgrade from an existing installation, 
follow [these instructions](https://github.com/kubestellar/kubeflex/blob/main/docs/users.md#upgrading-kubeflex).

2. Create a Kind hosting cluster with nginx ingress controller and KubeFlex operator installed:

```shell
kflex init --create-kind
```

3. Create an inventory & mailbox space of type `vcluster` running *OCM* (Open Cluster Management)
directly in KubeFlex. Note that `-p ocm` runs a post-create hook on the *vcluster* control plane 
which installs OCM on it.

```shell
kflex create imbs1 --type vcluster -p ocm
```

3.a (Optional) Install status add-on on imbs1:

```shell
helm --kube-context imbs1 upgrade --install status-addon -n open-cluster-management oci://quay.io/pdettori/status-addon-chart --version 0.1.0
```

see [here](https://github.ibm.com/dettori/status-addon) for more details on the add-on.

4. Create a Workload Description Space `wds1` directly in KubeFlex. Similarly to before, `-p kslight` 
runs a post-create hook on the *k8s* control plane that starts an instance of a KubeStellar 0.20 controller
manager which connects to the `wds1` front-end and the `imbs1` OCM control plane back-end.

```shell
kflex create wds1 -p kslight
```

5. Follow the steps to [register clusters with OCM](./thirdparties.md#register-clusters-with-ocm).


6. (optional) Check all deployments and statefulsets running in the hosting cluster:

```shell
kubectl --context kind-kubeflex get deployments,statefulsets --all-namespaces
```

## Scenario 1 - multi-cluster workload deployment with kubectl

Check for available clusters with label `location-group=edge`

```shell
kubectl --context imbs1 get managedclusters -l location-group=edge
```

Create a placement to deliver an app to all clusters in wds1:

```shell
kubectl --context wds1 apply -f - <<EOF
apiVersion: edge.kubestellar.io/v1alpha1
kind: Placement
metadata:
  name: nginx-placement
spec:
  clusterSelectors:
  - matchLabels: {"location-group":"edge"}
  downsync:
  - objectSelectors:
    - matchLabels: {"app.kubernetes.io/name":"nginx"}
EOF
```

This placement configuration determines **where** to deploy the workload by using 
the label selector expressions found in *clusterSelectors*. It also specifies **what** 
to deploy through the downsync.labelSelectors expressions. When there are multiple 
*matchLabels* expressions, they are combined using a logical **AND** operation. Conversely, 
when there are multiple *labelSelectors*, they are combined using a logical **OR** operation.

Now deploy the app:

```shell
kubectl --context wds1 apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  labels:
    app.kubernetes.io/name: nginx
  name: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  namespace: nginx
  labels:
    app.kubernetes.io/name: nginx
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
```

Verify that *manifestworks* wrapping the objects have been created in the mailbox 
namespaces:

```shell
kubectl --context imbs1 get manifestworks -n cluster1
kubectl --context imbs1 get manifestworks -n cluster2
```

Verify that the deployment has been created in both clusters

```shell
kubectl --context cluster1 get deployments -n nginx
kubectl --context cluster2 get deployments -n nginx
```

Please note, in line with Kubernetes’ best practices, the order in which you apply 
a placement and the objects doesn’t affect the outcome. You can apply the placement 
first followed by the objects, or vice versa. The result remains consistent because 
the placement controller identifies any changes in either the placement or the objects, 
triggering the start of the reconciliation loop.

## Scenario 2 - multi-cluster workload deployment with helm

Create a placement for the helm chart app:

```shell
kubectl --context wds1 apply -f - <<EOF
apiVersion: edge.kubestellar.io/v1alpha1
kind: Placement
metadata:
  name: postgres-placement
spec:
  clusterSelectors:
  - matchLabels: {"location-group":"edge"}
  downsync:
  - objectSelectors:
    - matchLabels: {
      "app.kubernetes.io/managed-by": Helm,
      "app.kubernetes.io/instance": postgres}
EOF
```

Note that helm sets `app.kubernetes.io/instance` to the *name* of the installed *release*.

Create and label the namespace and install the chart:

```shell
kubectl --context wds1 create ns postgres-system
kubectl --context wds1 label ns postgres-system app.kubernetes.io/managed-by=Helm app.kubernetes.io/instance=postgres
helm --kube-context wds1 install -n postgres-system postgres oci://registry-1.docker.io/bitnamicharts/postgresql
```

Verify that statefulset has been created in both clusters

```shell
kubectl --context cluster1 get statefulsets -n postgres-system
kubectl --context cluster2 get statefulsets -n postgres-system
```

### [Optional] Propagate helm metadata to managed clusters

Run "helm list" on the wds1:

```shell
$ helm --kube-context wds1 list -n postgres-system 
NAME            NAMESPACE       REVISION        UPDATED                                 STATUS       CHART                    APP VERSION
postgres        postgres-system 1               2023-10-31 13:39:52.550071 -0400 EDT    deployed     postgresql-13.2.0        16.0.0    
```

And try that on the managed clusters 

```shell
$ helm list --kube-context cluster1 -n postgres-system
# returns empty
$ helm list --kube-context cluster2 -n postgres-system
# returns empty
```

This is because the helm secret does not get delivered. That could be automated for example with:

```shell
kubectl --context wds1 get secrets -n postgres-system -l name=postgres -l owner=helm  -o jsonpath='{.items[0].metadata.name}'  | awk '{print "kubectl --context wds1 label secret -n postgres-system "$1" app.kubernetes.io/managed-by=Helm app.kubernetes.io/instance=postgres"}' | sh
```

Verify that the chart shows up on the managed clusters:

```shell
helm list --kube-context cluster1 -n postgres-system
helm list --kube-context cluster2 -n postgres-system
```
Implementing this in a controller for automated propagation of
helm metadata is tracked in this [issue](https://github.ibm.com/dettori/kslight/issues/1).

## Scenario 3 - multi-cluster workload deployment with ArgoCD

Before running this scenario, install ArgoCD on the hosting cluster and configure it 
work with the WDS as outlined [here](./docs/thirdparties.md#install-and-configure-argocd).

Apply the following placement to wds1:

```shell
kubectl --context wds1 apply -f - <<EOF
apiVersion: edge.kubestellar.io/v1alpha1
kind: Placement
metadata:
  name: argocd-placement
spec:
  clusterSelectors:
  - matchLabels: {"location-group":"edge"}
  downsync:
  - objectSelectors:
    - matchLabels: {"argocd.argoproj.io/instance":"guestbook"}
EOF
```

Switch context to hosting cluster and argocd namespace (this is required by argo to 
create an app with the CLI)

```shell
kubectl config use-context kind-kubeflex
kubectl config set-context --current --namespace=argocd
```

Create a new application in ArgoCD:

```shell
argocd app create guestbook --repo https://github.com/argoproj/argocd-example-apps.git --path guestbook --dest-server https://wds1.wds1-system --dest-namespace default
```

Open browser to Argo UI:

```shell
open https://argocd.localtest.me:9443
```

open the app `guestbook` and sync it clicking the "sync" button and then "synchronize", 
and finally check if the app has been deployed to the two clusters.

```shell
kubectl --context cluster1 get deployments,svc
kubectl --context cluster2 get deployments,svc
```

## Scenario 4 - Singleton status 

This scenario shows how to get the full status updated when setting `wantSingletonReportedState`
in the placement. This still an experimental feature, so some additional manual steps are required
at this time, such as the optional setup step 3.a.

Apply a placement with the `wantSingletonReportedState` flag set:

```shell
kubectl --context wds1 apply -f - <<EOF
apiVersion: edge.kubestellar.io/v1alpha1
kind: Placement
metadata:
  name: nginx-singleton-placement
spec:
  wantSingletonReportedState: true
  clusterSelectors:
  - matchLabels: {"location-group":"edge"}
  downsync:
  - objectSelectors:
    - matchLabels: {"app.kubernetes.io/name":"nginx-singleton"}
EOF
```
3. Apply a new deployment for the singleton placement:
```shell
kubectl --context wds1 apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-singleton-deployment
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
```

## Scenario 4 - Resiliency testing

Bring down the control plane: stop and restart wds1 and imbs1 API servers, 
KubeFlex and KubeStellar 0.20 controllers:

First stop all:

```shell
kubectl --context kind-kubeflex scale deployment -n wds1-system kube-apiserver --replicas=0
kubectl --context kind-kubeflex scale statefulset -n imbs1-system vcluster --replicas=0
kubectl --context kind-kubeflex scale deployment -n kubeflex-system kubeflex-controller-manager --replicas=0
kubectl --context kind-kubeflex scale deployment -n wds1-system kslight-controller-manager --replicas=0
```

Then restart all:

```shell
kubectl --context kind-kubeflex scale deployment -n wds1-system kube-apiserver --replicas=1
kubectl --context kind-kubeflex scale statefulset -n imbs1-system vcluster --replicas=1
kubectl --context kind-kubeflex scale deployment -n kubeflex-system kubeflex-controller-manager --replicas=1
kubectl --context kind-kubeflex scale deployment -n wds1-system kslight-controller-manager --replicas=1
```

Wait about a minute for all to restart, then apply a new placement:

```shell
kubectl --context wds1 apply -f - <<EOF
apiVersion: edge.kubestellar.io/v1alpha1
kind: Placement
metadata:
  name: nginx-res-placement
spec:
  clusterSelectors:
  - matchLabels: {"location-group":"edge"}
  downsync:
  - objectSelectors:
    - matchLabels: {"app.kubernetes.io/name":"nginx-res"}
EOF
```

and a new workload:

```shell
kubectl --context wds1 apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  labels:
    app.kubernetes.io/name: nginx-res
  name: nginx-res
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-res-deployment
  namespace: nginx-res
  labels:
    app.kubernetes.io/name: nginx-res
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx-res
  template:
    metadata:
      labels:
        app: nginx-res
    spec:
      containers:
      - name: nginx-res
        image: public.ecr.aws/nginx/nginx:latest 
        ports:
        - containerPort: 80
EOF
```

Verify that deployment has been created in both clusters

```shell
kubectl --context cluster1 get deployments -n nginx-res
kubectl --context cluster2 get deployments -n nginx-res
```

## Scenario 5 - multi-cluster workload deployment of app with SA with ArgoCD 

Apply the following placement to wds1:

```shell
kubectl --context wds1 apply -f - <<EOF
apiVersion: edge.kubestellar.io/v1alpha1
kind: Placement
metadata:
  name: argocd-sa-placement
spec:
  clusterSelectors:
  - matchLabels: {"location-group":"edge"}
  downsync:
  - objectSelectors:
    - matchLabels: {"argocd.argoproj.io/instance":"nginx-sa"}
EOF
```
Switch context to hosting cluster and argocd namespace (this is required by argo to 
create an app with the CLI)

```shell
kubectl config use-context kind-kubeflex
kubectl config set-context --current --namespace=argocd
```

Create a new application in ArgoCD:

```shell
argocd app create nginx-sa --repo https://github.com/pdettori/sample-apps.git --path nginx --dest-server https://wds1.wds1-system --dest-namespace nginx-sa
```

Open browser to Argo UI:

```shell
open https://argocd.localtest.me:9443
```

open the app `nginx-sa` and sync it clicking the "sync" button and then "synchronize", 
and finally check if the app has been deployed to the two clusters.

```shell
kubectl --context cluster1 -n nginx-sa get deployments,sa,secrets
kubectl --context cluster2 -n nginx-sa get deployments,sa,secrets
```

Repeat multiple syncing on Argo and check for potential proliferation of secrets on
both wds1 and clusters:

```shell
kubectl --context wds1 -n nginx-sa get secrets
kubectl --context cluster1 -n nginx-sa get secrets
kubectl --context cluster2 -n nginx-sa get secrets
```