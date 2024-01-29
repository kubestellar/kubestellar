# KubeStellar Usage Examples

**NOTE**: This is unmaintained material that has only been observed to work for the commit tagged as
`v0.20.0-alpha.1`. CI regularly tests variants of scenarios 1 and 4 that exercise the copy of the repo that they are embedded in (rather than the copy tagged `v0.20.0-alpha.1`), and contributors can run these tests too; see [the e2e tests](../../../test/e2e).

## Prereqs

See [pre-reqs](pre-reqs.md).

## Common Setup

The following steps establish an initial state used in the examples below.

1. Create a Kind hosting cluster with nginx ingress controller and KubeFlex operator installed:

   ```shell
   kflex init --create-kind
   ```

2. Update the post-create-hooks in KubeFlex to install kubestellar with the v0.20.0-alpha.1 images:

   ```shell
   kubectl apply -f https://raw.githubusercontent.com/kubestellar/kubestellar/v0.20.0-alpha.1/config/postcreate-hooks/kubestellar.yaml
   kubectl apply -f https://raw.githubusercontent.com/kubestellar/kubestellar/v0.20.0-alpha.1/config/postcreate-hooks/ocm.yaml
   ```

3. Create an inventory & mailbox space of type `vcluster` running *OCM* (Open Cluster Management)
in KubeFlex. Note that `-p ocm` runs a post-create hook on the *vcluster* control plane
which installs OCM on it.

   ```shell
   kflex create imbs1 --type vcluster -p ocm
   ```

4. Install status add-on on imbs1:

   Wait until the `managedclusteraddons` resource shows up on `imbs1`. You can check on that with the command:

   ```shell
   kubectl --context imbs1 api-resources | grep managedclusteraddons
   ```

   and then install the status add-on:

   ```shell
   helm --kube-context imbs1 upgrade --install status-addon -n open-cluster-management oci://quay.io/pdettori/status-addon-chart --version 0.1.0
   ```

   see [here](https://github.ibm.com/dettori/status-addon) for more details on the add-on.

5. Create a Workload Description Space `wds1` in KubeFlex. Similarly to before, `-p kubestellar`
runs a post-create hook on the *k8s* control plane that starts an instance of a KubeStellar controller
manager which connects to the `wds1` front-end and the `imbs1` OCM control plane back-end.

   ```shell
   kflex create wds1 -p kubestellar
   ```

6. Follow the steps to [create and register two clusters with OCM](example-wecs.md).

7. (optional) Check all deployments and statefulsets running in the hosting cluster. Expect to
see the wds1 kubestellar-controller-manager created in the wds1-system namespace and the imbs1
statefulset created in the imbs1-system namespace.

   ```shell
   kubectl --context kind-kubeflex get deployments,statefulsets --all-namespaces
   ```

## Scenario 1 - multi-cluster workload deployment with kubectl

This scenario proceeds from the state established by the [common setup](#common-setup).

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
to deploy through the downsync.labelSelectors expressions.
Each matchLabels expression is a criterion for selecting a set of objects based on
their labels. Other criteria can be added to filter objects based on their namespace,
api group, resource, and name. If these criteria are not specified, all objects with
the matching labels are selected. If an object has multiple labels, it is selected
only if it matches all the labels in the matchLabels expression.
If there are multiple objectSelectors, an object is selected if it matches any of them.

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

## Scenario 2 - using the hosting cluster as WDS to deploy a custom resource

This scenario follows on from the state established by scenario 1.

The hosting cluster can act as a Workload Description Space (WDS) to
distribute your workloads to multiple clusters. This feature works
well for Custom Resources, but not for standard Kubernetes resources
(deployments, pods, replicasets, etc.). The reason is that the hosting
cluster’s controller manager creates pods for those resources on the hosting
cluster itself, while the Kubestellar controller copies them to the Workload
Execution Clusters (WECs). You can use any Custom Resource to wrap any
Kubernetes object you want to send to the WECs. But if you have operators
or controllers on the hosting cluster that work on the Custom Resource
you want to send, make sure they don’t create workloads on the hosting
cluster that you did not intend to create there.

In order to run this scenario using the post-create-hook method you need
the raise the permissions for the kubeflex controller manager:

```shell
kubectl --context kind-kubeflex apply -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kubeflex-manager-cluster-admin-rolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: kubeflex-controller-manager
  namespace: kubeflex-system
EOF
```

To create a second WDS based on the hosting cluster, run the command:

```shell
kflex create wds2 -t host -p kubestellar
```

where the `-t host` option specifies a control plane of type `host`.
You can only create on control plane of type `host`.

Check that the kubestellar controller for wds2 is started:

```shell
kubectl get deployments.apps -n wds2-system kubestellar-controller-manager
```

If desired, you may remove the `kubeflex-manager-cluster-admin-rolebinding` after
the kubestellar-controller-manager is started, with the command
`kubectl --context kind-kubeflex delete clusterrolebinding kubeflex-manager-cluster-admin-rolebinding`

For this example, we use the `AppWrapper` custom resource defined in the
[multi cluster app dispatcher](https://github.com/project-codeflare/multi-cluster-app-dispatcher)
project.

Run the following comamand to give permission for the Klusterlet to
operate on your cluster resource.

```shell
clusters=(cluster1 cluster2);
for cluster in "${clusters[@]}"; do
kubectl --context ${cluster} apply -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: appwrappers-access
rules:
- apiGroups: ["mcad.ibm.com"]
  resources: ["appwrappers"]
  verbs: ["get", "list", "watch", "create", "update", "patch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: klusterlet-appwrappers-access
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: appwrappers-access
subjects:
- kind: ServiceAccount
  name: klusterlet-work-sa
  namespace: open-cluster-management-agent
EOF
done
```

This step will be eventually automated, see [this issue](https://github.com/kubestellar/kubestellar/issues/1542)
for more details.

Apply the appwrapper CRD to wds2:

```shell
kubectl --context wds2 apply -f https://raw.githubusercontent.com/project-codeflare/multi-cluster-app-dispatcher/v1.33.0/config/crd/bases/mcad.ibm.com_appwrappers.yaml
```

Now apply an appwrapper CR to wds2:

```shell
kubectl --context wds2 apply -f  https://raw.githubusercontent.com/project-codeflare/multi-cluster-app-dispatcher/v1.33.0/test/yaml/0001-aw-generic-deployment-3.yaml
```

Label the CRD and the CR:

```shell
kubectl --context wds2 label crd appwrappers.mcad.ibm.com app.kubernetes.io/part-of=my-appwrapper-app
kubectl --context wds2 label appwrappers 0001-aw-generic-deployment-3 app.kubernetes.io/part-of=my-appwrapper-app
```

Finally, apply the placement:

```shell
kubectl --context wds2 apply -f - <<EOF
apiVersion: edge.kubestellar.io/v1alpha1
kind: Placement
metadata:
  name: aw-placement
spec:
  clusterSelectors:
  - matchLabels: {"location-group":"edge"}
  downsync:
  - objectSelectors:
    - matchLabels: {"app.kubernetes.io/part-of":"my-appwrapper-app"}
EOF
```

Check that the app wrapper has been delivered to both clusters:

```shell
kubectl --context cluster1 get appwrappers
kubectl --context cluster2 get appwrappers
```

## Scenario 3 - multi-cluster workload deployment with helm

This scenario proceeds from the state established by the [common setup](#common-setup).

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

### [Optional] Propagate helm metadata Secret to managed clusters

Run "helm list" on the wds1:

```shell
$ helm --kube-context wds1 list -n postgres-system
NAME            NAMESPACE       REVISION        UPDATED                                 STATUS       CHART                    APP VERSION
postgres        postgres-system 1               2023-10-31 13:39:52.550071 -0400 EDT    deployed     postgresql-13.2.0        16.0.0
```

And try that on the managed clusters

```shell
$ helm list --kube-context cluster1 -n postgres-system
: returns empty
$ helm list --kube-context cluster2 -n postgres-system
: returns empty
```

This is because Helm creates a `Secret` object to hold its metadata about a "release" (chart instance) but Helm does not apply the usual labels to that object, so it is not selected by the `Placement` above and thus does not get delivered. The workload is functioning in the WECs, but `helm list` does not recognize its handiwork there. That labeling could be done for example with:

```shell
kubectl --context wds1 label secret -n postgres-system $(kubectl --context wds1 get secrets -n postgres-system -l name=postgres -l owner=helm  -o jsonpath='{.items[0].metadata.name}') app.kubernetes.io/managed-by=Helm app.kubernetes.io/instance=postgres
```

Verify that the chart shows up on the managed clusters:

```shell
helm list --kube-context cluster1 -n postgres-system
helm list --kube-context cluster2 -n postgres-system
```

Implementing this in a controller for automated propagation of
helm metadata is tracked in this [issue](https://github.com/kubestellar/kubestellar/issues/1543).

## Scenario 4 - Singleton status

This scenario proceeds from the state established by the [common setup](#common-setup).

This scenario shows how to get the full status updated when setting `wantSingletonReportedState`
in the placement. This still an experimental feature.

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
  - matchLabels: {"name":"cluster1"}
  downsync:
  - objectSelectors:
    - matchLabels: {"app.kubernetes.io/name":"nginx-singleton"}
EOF
```

Apply a new deployment for the singleton placement:

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

Verify that the status is available in wds1 for the deployment by
running the command:

```shell
kubectl --context wds1 get deployments nginx-singleton-deployment -o yaml
```

Finally, scale the deployment from 1 to 2 replicas in wds1:

```shell
kubectl --context wds1 scale deployment nginx-singleton-deployment --replicas=2
```

and verify that replicas has been updated in cluster1 and wds1:

```shell
kubectl --context cluster1 get deployment nginx-singleton-deployment
kubectl --context wds1 get deployment nginx-singleton-deployment
```

## Scenario 5 - Resiliency testing

This is a test that you can do after finishing Scenario 1.

TODO: rewrite this so that it makes sense after Scenario 4.

Bring down the control plane: stop and restart wds1 and imbs1 API servers,
KubeFlex and KubeStellar controllers:

First stop all:

```shell
kubectl --context kind-kubeflex scale deployment -n wds1-system kube-apiserver --replicas=0
kubectl --context kind-kubeflex scale statefulset -n imbs1-system vcluster --replicas=0
kubectl --context kind-kubeflex scale deployment -n kubeflex-system kubeflex-controller-manager --replicas=0
kubectl --context kind-kubeflex scale deployment -n wds1-system kubestellar-controller-manager --replicas=0
```

Then restart all:

```shell
kubectl --context kind-kubeflex scale deployment -n wds1-system kube-apiserver --replicas=1
kubectl --context kind-kubeflex scale statefulset -n imbs1-system vcluster --replicas=1
kubectl --context kind-kubeflex scale deployment -n kubeflex-system kubeflex-controller-manager --replicas=1
kubectl --context kind-kubeflex scale deployment -n wds1-system kubestellar-controller-manager --replicas=1
```

Wait for about a minute for all pods to restart, then apply a new placement:

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

## Scenario 6 - multi-cluster workload deployment of app with ServiceAccount with ArgoCD

This scenario is something you can do after the [common setup](#common-setup).

Before running this scenario, install ArgoCD on the hosting cluster and configure it
work with the WDS as outlined [here](./thirdparties.md#install-and-configure-argocd).

Including a ServiceAccount tests whether there will be a controller fight over a token Secret for that ServiceAccount, which was observed in some situations with older code.

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

Open the app `nginx-sa` and sync it by clicking the "sync" button and then "synchronize".

Alternatively, use the CLI to sync the app:

```shell
argocd app sync nginx-sa
```

Finally, check if the app has been deployed to the two clusters.

```shell
kubectl --context cluster1 -n nginx-sa get deployments,sa,secrets
kubectl --context cluster2 -n nginx-sa get deployments,sa,secrets
```

Repeat multiple syncing on Argo and verify that extra secrets for the service acccount
are not created both wds1 and clusters:

```shell
kubectl --context wds1 -n nginx-sa get secrets
kubectl --context cluster1 -n nginx-sa get secrets
kubectl --context cluster2 -n nginx-sa get secrets
```
