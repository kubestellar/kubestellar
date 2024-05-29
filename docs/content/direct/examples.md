# KubeStellar Usage Examples

This document shows some simple examples of using the release that contains this version of this document. See also [the doc about using an existing hosting cluster and multi-machine scenarios](hosting-cluster.md) for considerations for different scenarios. For historical examples, see [our blog posts](https://medium.com/@kubestellar/list/predefined:e785a0675051:READING_LIST).

Each example scenario concludes with instructions on how to undo its effects, restoring the state established by [the common setup](#common-setup) so that it is easy to run through several scenarios without having to repeat the common setup.

There are also end-to-end (E2E) tests that are based on scenario 4 and an extended variant of scenario 1. These tests normally exercise the copy of the repo containing them (rather than a release). They can alternatively test a release. See the e2e tests (in `test/e2e`). Contributors can run these tests, and CI includes checking that these E2E tests pass. These tests are written in `bash`, so that contributors can easily follow them.

## Prereqs

See [pre-reqs](pre-reqs.md).

## Common Setup

The following steps establish an initial state used in the examples below.

There are to main way to proceed:

1. [Step-by-step manual setup](common-setup-step-by-step.md)
2. [Using KubeStellar Core chart](common-setup-core-chart.md)

## Create and register two clusters with OCM

Follow the steps to [create and register two clusters with OCM](example-wecs.md).

## Scenario 1 - multi-cluster workload deployment with kubectl

This scenario proceeds from the state established by the [common setup](#common-setup).

Check for available clusters with label `location-group=edge`

```shell
kubectl --context its1 get managedclusters -l location-group=edge
```

Create a BindingPolicy to deliver an app to all clusters in wds1:

```shell
kubectl --context wds1 apply -f - <<EOF
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: nginx-bpolicy
spec:
  clusterSelectors:
  - matchLabels: {"location-group":"edge"}
  downsync:
  - objectSelectors:
    - matchLabels: {"app.kubernetes.io/name":"nginx"}
EOF
```

This BindingPolicy configuration determines **where** to deploy the workload by using
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
kubectl --context its1 get manifestworks -n cluster1
kubectl --context its1 get manifestworks -n cluster2
```

Verify that the deployment has been created in both clusters

```shell
kubectl --context cluster1 get deployments -n nginx
kubectl --context cluster2 get deployments -n nginx
```

Please note, in line with Kubernetes’ best practices, the order in which you apply
a BindingPolicy and the objects doesn’t affect the outcome. You can apply the BindingPolicy
first followed by the objects, or vice versa. The result remains consistent because
the binding controller identifies any changes in either the BindingPolicy or the objects,
triggering the start of the reconciliation loop.

### [Optional] Teardown Scenario 1

```shell
kubectl --context wds1 delete ns nginx
kubectl --context wds1 delete bindingpolicies nginx-bpolicy
```

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

To create a second WDS based on the hosting cluster, run the commands:

```shell
kflex create wds2 -t host

helm --kube-context kind-kubeflex upgrade --install ocm-transport-plugin oci://ghcr.io/kubestellar/ocm-transport-plugin/chart/ocm-transport-plugin --version ${OCM_TRANSPORT_PLUGIN} \
    --set transport_cp_name=its1 \
    --set wds_cp_name=wds2 \
    -n wds2-system
```

where the `-t host` option specifies a control plane of type `host`.
You can only create one control plane of type `host`.

In this example, we use the helm chart method to install the kubestellat controller manager for the
hosting cluster so that we can pass additional startup options.

Label the `wds2` control plane as type `wds`:

```shell
kubectl label cp wds2 kflex.kubestellar.io/cptype=wds
```

For this example, we use the `AppWrapper` custom resource defined in the
[multi cluster app dispatcher](https://github.com/project-codeflare/multi-cluster-app-dispatcher)
project.

Install the AppWrapper CRD in the WDS and the WECs.

```shell
clusters=(wds2 cluster1 cluster2);
  for cluster in "${clusters[@]}"; do
  kubectl --context ${cluster} apply -f https://raw.githubusercontent.com/project-codeflare/multi-cluster-app-dispatcher/v1.39.0/config/crd/bases/workload.codeflare.dev_appwrappers.yaml
done
```

Apply the kubestellar controller-manager helm chart with the option to allow only delivery of objects with api group `workload.codeflare.dev`

```shell
helm --kube-context kind-kubeflex upgrade --install -n wds2-system kubestellar oci://ghcr.io/kubestellar/kubestellar/controller-manager-chart --version ${KUBESTELLAR_VERSION} --set ControlPlaneName=wds2 --set APIGroups=workload.codeflare.dev
```

Check that the kubestellar controller for wds2 is started:

```shell
kubectl get deployments.apps -n wds2-system kubestellar-controller-manager
```

If desired, you may remove the `kubeflex-manager-cluster-admin-rolebinding` after
the kubestellar-controller-manager is started, with the command
`kubectl --context kind-kubeflex delete clusterrolebinding kubeflex-manager-cluster-admin-rolebinding`

Run the following comamand to give permission for the Klusterlet to
operate on the appwrapper cluster resource.

```shell
clusters=(cluster1 cluster2);
for cluster in "${clusters[@]}"; do
kubectl --context ${cluster} apply -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: appwrappers-access
rules:
- apiGroups: ["workload.codeflare.dev"]
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

Next, apply an appwrapper object to wds2:

```shell
kubectl --context wds2 apply -f  https://raw.githubusercontent.com/project-codeflare/multi-cluster-app-dispatcher/v1.39.0/test/yaml/0008-aw-default.yaml
```

Label the appwrapper to match the binding policy:

```shell
kubectl --context wds2 label appwrappers.workload.codeflare.dev defaultaw-schd-spec-with-timeout-1 app.kubernetes.io/part-of=my-appwrapper-app
```

Finally, apply the BindingPolicy:

```shell
kubectl --context wds2 apply -f - <<EOF
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: aw-bpolicy
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

### [Optional] Teardown Scenario 2

```shell
kubectl --context wds2 delete bindingpolicies aw-bpolicy
kubectl --context wds2 delete appwrappers --all
```

Wait until the following commands show no appwrappers in cluster1 and cluster2.

```shell
kubectl --context cluster1 get appwrappers -A
kubectl --context cluster2 get appwrappers -A
```

Then continue.

```shell
for cluster in cluster1 cluster2; do
  kubectl --context $cluster delete clusterroles appwrappers-access
  kubectl --context $cluster delete clusterrolebindings klusterlet-appwrappers-access
done
```

If you have not already done so, then do the following command.

```shell
kubectl --context kind-kubeflex delete clusterrolebinding kubeflex-manager-cluster-admin-rolebinding
```

Continue as follows.

```shell
helm --kube-context kind-kubeflex uninstall -n wds2-system kubestellar

clusters=(wds2 cluster1 cluster2);
  for cluster in "${clusters[@]}"; do
  kubectl --context ${cluster} delete -f https://raw.githubusercontent.com/project-codeflare/multi-cluster-app-dispatcher/v1.39.0/config/crd/bases/workload.codeflare.dev_appwrappers.yaml
done

kflex delete wds2
```

## Scenario 3 - multi-cluster workload deployment with helm

This scenario proceeds from the state established by the [common setup](#common-setup).

Create a BindingPolicy for the helm chart app:

```shell
kubectl --context wds1 apply -f - <<EOF
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: postgres-bpolicy
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

This is because Helm creates a `Secret` object to hold its metadata about a "release" (chart instance) but Helm does not apply the usual labels to that object, so it is not selected by the `BindingPolicy` above and thus does not get delivered. The workload is functioning in the WECs, but `helm list` does not recognize its handiwork there. That labeling could be done for example with:

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

### [Optional] Teardown Scenario 3

```shell
helm --kube-context wds1 uninstall -n postgres-system postgres
kubectl --context wds1 delete ns postgres-system
kubectl --context wds1 delete bindingpolicies postgres-bpolicy
```

## Scenario 4 - Singleton status

This scenario proceeds from the state established by the [common setup](#common-setup).

This scenario shows how to get the full status updated when setting `wantSingletonReportedState`
in the BindingPolicy. This still an experimental feature.

Apply a BindingPolicy with the `wantSingletonReportedState` flag set:

```shell
kubectl --context wds1 apply -f - <<EOF
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: nginx-singleton-bpolicy
spec:
  wantSingletonReportedState: true
  clusterSelectors:
  - matchLabels: {"name":"cluster1"}
  downsync:
  - objectSelectors:
    - matchLabels: {"app.kubernetes.io/name":"nginx-singleton"}
EOF
```

Apply a new deployment for the singleton BindingPolicy:

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

### [Optional] Teardown Scenario 4

```shell
kubectl --context wds1 delete bindingpolicies nginx-singleton-bpolicy
kubectl --context wds1 delete deployments nginx-singleton-deployment
```

## Scenario 5 - Resiliency testing

This is a test that you can do after finishing Scenario 1.

TODO: rewrite this so that it makes sense after Scenario 4.

Bring down the control plane: stop and restart wds1 and its1 API servers,
KubeFlex and KubeStellar controllers:

First stop all:

```shell
kubectl --context kind-kubeflex scale deployment -n wds1-system kube-apiserver --replicas=0
kubectl --context kind-kubeflex scale statefulset -n its1-system vcluster --replicas=0
kubectl --context kind-kubeflex scale deployment -n kubeflex-system kubeflex-controller-manager --replicas=0
kubectl --context kind-kubeflex scale deployment -n wds1-system kubestellar-controller-manager --replicas=0
kubectl --context kind-kubeflex scale deployment -n wds1-system transport-controller --replicas=0
```

Then restart all:

```shell
kubectl --context kind-kubeflex scale deployment -n wds1-system kube-apiserver --replicas=1
kubectl --context kind-kubeflex scale statefulset -n its1-system vcluster --replicas=1
kubectl --context kind-kubeflex scale deployment -n kubeflex-system kubeflex-controller-manager --replicas=1
kubectl --context kind-kubeflex scale deployment -n wds1-system kubestellar-controller-manager --replicas=1
kubectl --context kind-kubeflex scale deployment -n wds1-system transport-controller --replicas=1
```

Wait for about a minute for all pods to restart, then apply a new BindingPolicy:

```shell
kubectl --context wds1 apply -f - <<EOF
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: nginx-res-bpolicy
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

## [Optional] Teardown Scenario 5

```shell
kubectl --context wds1 delete ns nginx-res
kubectl --context wds1 delete bindingpolicies nginx-res-bpolicy
```

Then continue with [teardown of scenario 1](#teardown-scenario-1).

## Scenario 6 - multi-cluster workload deployment of app with ServiceAccount with ArgoCD

This scenario is something you can do after the [common setup](#common-setup).

Before running this scenario, install ArgoCD on the hosting cluster and configure it
work with the WDS as outlined [here](./thirdparties.md#install-and-configure-argocd).

Including a ServiceAccount tests whether there will be a controller fight over a token Secret for that ServiceAccount, which was observed in some situations with older code.

Apply the following BindingPolicy to wds1:

```shell
kubectl --context wds1 apply -f - <<EOF
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: argocd-sa-bpolicy
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

### [Optional] Teardown Scenario 6

(Assuming that kubectl is still using context `kind-kubeflex` and namespace `argocd`.)

```shell
argocd app delete nginx-sa --cascade
kubectl --context wds1 delete bindingpolicies argocd-sa-bpolicy
```
