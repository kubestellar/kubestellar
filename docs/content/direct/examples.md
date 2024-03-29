# KubeStellar Usage Examples

This document shows some simple examples of using the release that contains this version of this document. See also [the doc about using an existing hosting cluster and multi-machine scenarios](hosting-cluster.md) for considerations for different scenarios. For historical examples, see [our blog posts](https://medium.com/@kubestellar/list/predefined:e785a0675051:READING_LIST).

There are also end-to-end (E2E) tests that are based on scenario 4 and an extended variant of scenario 1. These tests normally exercise the copy of the repo containing them (rather than a release). They can alternatively test a release. See the e2e tests (in `test/e2e`). Contributors can run these tests, and CI includes checking that these E2E tests pass. These tests are written in `bash`, so that contributors can easily follow them.

## Prereqs

See [pre-reqs](pre-reqs.md).

## Common Setup

The following steps establish an initial state used in the examples below.

1. You may want to `set -e` in your shell so that any failures in the setup or usage scenarios are not lost.

1. If you ran through these scenarios previously then you will need to do a bit of cleanup first. See how it is done in the cleanup script for our E2E tests (in `test/e2e/common/cleanup.sh`).

1. Set environment variables to hold KubeStellar and OCM-status-addon desired versions:

    ```shell
    export KUBESTELLAR_VERSION=0.21.2-rc1
    export OCM_STATUS_ADDON_VERSION=0.2.0-rc6
    ```

1. Create a Kind hosting cluster with nginx ingress controller and KubeFlex controller-manager installed:

    ```shell
    kflex init --create-kind
    ```
   If you are installing KubeStellar on an existing Kubernetes or OpenShift cluster, just use the command `kflex init`.

1. Update the post-create-hooks in KubeFlex to install kubestellar with the desired images:

    ```shell
    kubectl apply -f https://raw.githubusercontent.com/kubestellar/kubestellar/v${KUBESTELLAR_VERSION}/config/postcreate-hooks/kubestellar.yaml
    kubectl apply -f https://raw.githubusercontent.com/kubestellar/kubestellar/v${KUBESTELLAR_VERSION}/config/postcreate-hooks/ocm.yaml
    ```

1. Create an inventory & mailbox space of type `vcluster` running *OCM* (Open Cluster Management)
in KubeFlex. Note that `-p ocm` runs a post-create hook on the *vcluster* control plane
which installs OCM on it.

    ```shell
    kflex create imbs1 --type vcluster -p ocm
    ```

1. Install status add-on on imbs1:

    Wait until the `managedclusteraddons` resource shows up on `imbs1`. You can check on that with the command:

    ```shell
    kubectl --context imbs1 api-resources | grep managedclusteraddons
    ```

    and then install the status add-on:

    ```shell
    helm --kube-context imbs1 upgrade --install status-addon -n open-cluster-management oci://ghcr.io/kubestellar/ocm-status-addon-chart --version v${OCM_STATUS_ADDON_VERSION}
    ```

    see [here](./architecture.md#ocm-status-add-on-agent) for more details on the add-on.

1. Create a Workload Description Space `wds1` in KubeFlex. Similarly to before, `-p kubestellar`
runs a post-create hook on the *k8s* control plane that starts an instance of a KubeStellar controller
manager which connects to the `wds1` front-end and the `imbs1` OCM control plane back-end.

    ```shell
    kflex create wds1 -p kubestellar
    ```

1. Run the OCM based transport controller in a pod.  
**NOTE**: This is work in progress, in the future the controller will be deployed through a Helm chart.

    Run the transport deployment script (in `scripts/deploy-transport-controller.sh`), as follows.
    This script requires that the user's current kubeconfig context be for the kubeflex hosting cluster.
    This script expects to get two or three arguments - (1) wds name; (2) imbs name; and (3) transport controller image.  
    While the first and second arguments are mandatory, the third one is optional.
    The transport controller image argument can be specified to a specific image, or, if omitted, it defaults to the OCM transport plugin release that preceded the KubeStellar release being used.
    For example, one can deploy transport controller using the following commands:
    ```shell
    kflex ctx
    bash <(curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/v${KUBESTELLAR_VERSION}/scripts/deploy-transport-controller.sh) wds1 imbs1
    ```

1. Follow the steps to [create and register two clusters with OCM](example-wecs.md).

1. (optional) Check relevant deployments and statefulsets running in the hosting cluster. Expect to
see the `kubestellar-controller-manager` in the `wds1-system` namespace and the 
statefulset `vcluster` in the `imbs1-system` namespace, both fully ready.

    ```shell
    kubectl --context kind-kubeflex get deployments,statefulsets --all-namespaces
    ```
   The output should look something like the following:

    ```shell
    NAMESPACE            NAME                                             READY   UP-TO-DATE   AVAILABLE   AGE
    ingress-nginx        deployment.apps/ingress-nginx-controller         1/1     1            1           22h
    kube-system          deployment.apps/coredns                          2/2     2            2           22h
    kubeflex-system      deployment.apps/kubeflex-controller-manager      1/1     1            1           22h
    local-path-storage   deployment.apps/local-path-provisioner           1/1     1            1           22h
    wds1-system          deployment.apps/kube-apiserver                   1/1     1            1           22m
    wds1-system          deployment.apps/kube-controller-manager          1/1     1            1           22m
    wds1-system          deployment.apps/kubestellar-controller-manager   1/1     1            1           21m
    wds1-system          deployment.apps/transport-controller             1/1     1            1           21m

    NAMESPACE         NAME                                   READY   AGE
    imbs1-system      statefulset.apps/vcluster              1/1     11h
    kubeflex-system   statefulset.apps/postgres-postgresql   1/1     22h
    ```

## Scenario 1 - multi-cluster workload deployment with kubectl

This scenario proceeds from the state established by the [common setup](#common-setup).

Check for available clusters with label `location-group=edge`

```shell
kubectl --context imbs1 get managedclusters -l location-group=edge
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
kubectl --context imbs1 get manifestworks -n cluster1
kubectl --context imbs1 get manifestworks -n cluster2
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
bash <(curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/v${KUBESTELLAR_VERSION}/scripts/deploy-transport-controller.sh) wds2 imbs1
```

where the `-t host` option specifies a control plane of type `host`.
You can only create on control plane of type `host`.

In this example, we use the helm chart method to install the kubestellat controller manager for the
hosting cluster so that we can pass additional startup options.

Label the `wds2` control plane as type `wds`:

```shell 
kubectl label cp wds2 kflex.kubestellar.io/cptype=wds
```

For this example, we use the `AppWrapper` custom resource defined in the
[multi cluster app dispatcher](https://github.com/project-codeflare/multi-cluster-app-dispatcher)
project.

Install the AppWrapper CRD in the WDS and the WECs. Note that due to 
[this issue](https://github.com/kubestellar/kubestellar/issues/1705) CRDs must be pre-installed 
on the WDS and on the WECs when using API group filtering. For this release of KubeStellar, that pre-installation is required 
when a WDS has a large number of API resources (such as would be found in a hosting cluster that is OpenShift).

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
kubectl --context kind-kubeflex scale deployment -n wds1-system transport-controller --replicas=0
```

Then restart all:

```shell
kubectl --context kind-kubeflex scale deployment -n wds1-system kube-apiserver --replicas=1
kubectl --context kind-kubeflex scale statefulset -n imbs1-system vcluster --replicas=1
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
