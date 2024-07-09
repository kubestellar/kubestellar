# KubeStellar Example Scenarios

This document shows some simple examples of using the release that contains this version of this document. These scenarios can be used to test a KubeStellar installation for proper functionality. General setup instructions are outlined in [the Overview](user-guide-intro.md); a simple example setup is in [the quickstart](get-started.md).

Each scenario supposes that one ITS and one WDS have been created and two WECs have been created and registered. These scenarios are written as shell commands (bash or zsh). These commands assume that you have defined the following shell variables to convey the needed information about that ITS and WDS and those WECs.

- `host_context`: the name of the kubeconfig context to use when accessing the KubeFlex hosting cluster.
- `its_cp`: the name of the KubeFlex control plane that is playing the role of ITS.
- `its_context`: the name of the kubeconfig context to use when accessing the ITS.
- `wds_cp`: the name of the KubeFlex control plane that is playing the role of WDS.
- `wds_context`: the name of the kubeconfig context to use when accessing the WDS.
- `wec1_name`, `wec2_name`: the names of the `ManagedCluster` objects in the ITS representing the two WECs.
- `wec1_context`, `wec2_context`: the names of the kubeconfig contexts to use when accessing the two WECs.
- `label_query_both`: a restricted `kubectl` label query over `ManagedCluster` objects in the ITS that matches both WECs. The general form of label query usable here is a comma-separated series of `key=value` requirements.
- `label_query_one`: a restricted `kubectl` label query over `ManagedCluster` objects that picks out just one of the WECs.

Each example scenario concludes with instructions on how to undo its effects.

There are also end-to-end (E2E) tests that are based on scenario 4 and an extended variant of scenario 1. These tests normally exercise the copy of the repo containing them (rather than a release). They can alternatively test a release. See the e2e tests (in `test/e2e`). Contributors can run these tests, and CI includes checking that these E2E tests pass. Some of these tests, and the setup for all of them, are written in `bash` so that contributors can easily follow them.

## Scenario 0 - look around

The following command will list all the `ManagedCluster` objects that will be relevant to these scenarios.

```shell
kubectl --context "$its_context" get managedclusters -l "$label_query_both"
```

Expect to get a listing of your two `ManagedCluster` objects.

## Scenario 1 - multi-cluster workload deployment with kubectl

Create a BindingPolicy to deliver an app to all clusters in the WDS:

```shell
kubectl --context "$wec1_context" apply -f - <<EOF
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: nginx-bpolicy
spec:
  clusterSelectors:
  - matchLabels: {$(echo "$label_query_both" | tr , $'\n' | while IFS="=" read key val; do echo -n ", \"$key\": \"$val\""; done | tail +3c)}
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
kubectl --context "$wds_context" apply -f - <<EOF
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
kubectl --context "$its_context" get manifestworks -n "$wec1_name"
kubectl --context "$its_context" get manifestworks -n "$wec2_name"
```

Verify that the deployment has been created in both clusters

```shell
kubectl --context "$wec1_context" get deployments -n nginx
kubectl --context "$wec2_context" get deployments -n nginx
```

Please note, in line with Kubernetes’ best practices, the order in which you apply
a BindingPolicy and the objects doesn’t affect the outcome. You can apply the BindingPolicy
first followed by the objects, or vice versa. The result remains consistent because
the binding controller identifies any changes in either the BindingPolicy or the objects,
triggering the start of the reconciliation loop.

### [Optional] Teardown Scenario 1

```shell
kubectl --context "$wds_context" delete ns nginx
kubectl --context "$wds_context" delete bindingpolicies nginx-bpolicy
```

## Scenario 2 - Out-of-tree workload

This scenario is like the previous one but involves a workload whose
kind of objects is not built into Kubernetes. Instead, the workload
object kind is defined by a `CustomResourceDefinition` object. While
KubeStellar can handle the case where the CRD is part of the workload,
this example concerns the case where the CRD is established in the
WECs by some other means.

In order to run this scenario using the post-create-hook method you need
the raise the permissions for the kubeflex controller manager (TODO 1: move this material and its undo to [the doc on WDS](wds.md); TODO 2: why is this needed? Is it needed for the core chart too? can we remove the need for this?):

```shell
kubectl --context "$host_context" apply -f - <<EOF
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

For this example, we use the `AppWrapper` custom resource defined in the
[multi cluster app dispatcher](https://github.com/project-codeflare/multi-cluster-app-dispatcher)
project.

Install the AppWrapper CRD in the WDS and the WECs.

```shell
clusters=("$wds_context" "$wec1_context" "$wec2_context");
  for cluster in "${clusters[@]}"; do
  kubectl --context ${cluster} apply -f https://raw.githubusercontent.com/project-codeflare/multi-cluster-app-dispatcher/v1.39.0/config/crd/bases/workload.codeflare.dev_appwrappers.yaml
done
```

If desired, you may remove the `kubeflex-manager-cluster-admin-rolebinding` after
the kubestellar-controller-manager is started, with the command
`kubectl --context "$host_context" delete clusterrolebinding kubeflex-manager-cluster-admin-rolebinding`

Run the following command to give permission for the klusterlet to
operate on the appwrapper cluster resource.

```shell
clusters=("$wec1_context" "$wec2_context");
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

Next, apply an appwrapper object to the WDS:

```shell
kubectl --context "$wds_context" apply -f  https://raw.githubusercontent.com/project-codeflare/multi-cluster-app-dispatcher/v1.39.0/test/yaml/0008-aw-default.yaml
```

Label the appwrapper to match the binding policy:

```shell
kubectl --context "$wds_context" label appwrappers.workload.codeflare.dev defaultaw-schd-spec-with-timeout-1 app.kubernetes.io/part-of=my-appwrapper-app
```

Finally, apply the BindingPolicy:

```shell
kubectl --context "$wds_context" apply -f - <<EOF
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: aw-bpolicy
spec:
  clusterSelectors:
  - matchLabels: {$(echo "$label_query_both" | tr , $'\n' | while IFS="=" read key val; do echo -n ", \"$key\": \"$val\""; done | tail +3c)}
  downsync:
  - objectSelectors:
    - matchLabels: {"app.kubernetes.io/part-of":"my-appwrapper-app"}
EOF
```

Check that the app wrapper has been delivered to both clusters:

```shell
kubectl --context "$wec1_context" get appwrappers
kubectl --context "$wec2_context" get appwrappers
```

### [Optional] Teardown Scenario 2

```shell
kubectl --context "$wds_context" delete bindingpolicies aw-bpolicy
kubectl --context "$wds_context" delete appwrappers --all
```

Wait until the following commands show no appwrappers in the two WECs.

```shell
kubectl --context "$wec1_context" get appwrappers -A
kubectl --context "$wec2_context" get appwrappers -A
```

Then continue.

```shell
for cluster in "$wec1_context" "$wec2_context"; do
  kubectl --context $cluster delete clusterroles appwrappers-access
  kubectl --context $cluster delete clusterrolebindings klusterlet-appwrappers-access
done
```

If you have not already done so, then do the following command.

```shell
kubectl --context "$host_context" delete clusterrolebinding kubeflex-manager-cluster-admin-rolebinding
```

Delete the CRD from the WDS and the WECs.

```shell
clusters=("$wds_context" "$wec1_context" "$wec2_context");
  for cluster in "${clusters[@]}"; do
  kubectl --context ${cluster} delete -f https://raw.githubusercontent.com/project-codeflare/multi-cluster-app-dispatcher/v1.39.0/config/crd/bases/workload.codeflare.dev_appwrappers.yaml
done
```

## Scenario 3 - multi-cluster workload deployment with helm

Create a BindingPolicy for the helm chart app:

```shell
kubectl --context "$wds_context" apply -f - <<EOF
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: postgres-bpolicy
spec:
  clusterSelectors:
  - matchLabels: {$(echo "$label_query_both" | tr , $'\n' | while IFS="=" read key val; do echo -n ", \"$key\": \"$val\""; done | tail +3c)}
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
kubectl --context "$wds_context" create ns postgres-system
kubectl --context "$wds_context" label ns postgres-system app.kubernetes.io/managed-by=Helm app.kubernetes.io/instance=postgres
helm --kube-context "$wds_context" install -n postgres-system postgres oci://registry-1.docker.io/bitnamicharts/postgresql
```

Verify that `StatefulSet` has been created in both clusters

```shell
kubectl --context "$wec1_context" get statefulsets -n postgres-system
kubectl --context "$wec2_context" get statefulsets -n postgres-system
```

### [Optional] Propagate helm metadata Secret to managed clusters

Run "helm list" on the WDS:

```shell
$ helm --kube-context "$wds_context" list -n postgres-system
NAME            NAMESPACE       REVISION        UPDATED                                 STATUS       CHART                    APP VERSION
postgres        postgres-system 1               2023-10-31 13:39:52.550071 -0400 EDT    deployed     postgresql-13.2.0        16.0.0
```

And try that on the managed clusters

```shell
$ helm list --kube-context "$wec1_context" -n postgres-system
: returns empty
$ helm list --kube-context "$wec2_context" -n postgres-system
: returns empty
```

This is because Helm creates a `Secret` object to hold its metadata about a "release" (chart instance) but Helm does not apply the usual labels to that object, so it is not selected by the `BindingPolicy` above and thus does not get delivered. The workload is functioning in the WECs, but `helm list` does not recognize its handiwork there. That labeling could be done for example with:

```shell
kubectl --context "$wds_context" label secret -n postgres-system $(kubectl --context "$wds_context" get secrets -n postgres-system -l name=postgres -l owner=helm  -o jsonpath='{.items[0].metadata.name}') app.kubernetes.io/managed-by=Helm app.kubernetes.io/instance=postgres
```

Verify that the chart shows up on the managed clusters:

```shell
helm list --kube-context "$wec1_context" -n postgres-system
helm list --kube-context "$wec2_context" -n postgres-system
```

Implementing this in a controller for automated propagation of
helm metadata is tracked in this [issue](https://github.com/kubestellar/kubestellar/issues/1543).

### [Optional] Teardown Scenario 3

```shell
helm --kube-context "$wds_context" uninstall -n postgres-system postgres
kubectl --context "$wds_context" delete ns postgres-system
kubectl --context "$wds_context" delete bindingpolicies postgres-bpolicy
```

## Scenario 4 - Singleton status

This scenario shows how to get the full status updated when setting `wantSingletonReportedState`
in the BindingPolicy. This still an experimental feature.

Apply a BindingPolicy with the `wantSingletonReportedState` flag set:

```shell
kubectl --context "$wds_context" apply -f - <<EOF
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
kubectl --context "$wds_context" apply -f - <<EOF
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

Verify that the status is available in the WDS for the deployment by
running the command:

```shell
kubectl --context "$wds_context" get deployments nginx-singleton-deployment -o yaml
```

Finally, scale the deployment from 1 to 2 replicas in the WDS:

```shell
kubectl --context "$wds_context" scale deployment nginx-singleton-deployment --replicas=2
```

and verify that replicas has been updated in the WEC and the WDS:

```shell
kubectl --context "$wec1_context" get deployment nginx-singleton-deployment
kubectl --context "$wds_context" get deployment nginx-singleton-deployment
```

### [Optional] Teardown Scenario 4

```shell
kubectl --context "$wds_context" delete bindingpolicies nginx-singleton-bpolicy
kubectl --context "$wds_context" delete deployments nginx-singleton-deployment
```

## Scenario 5 - Resiliency testing

This is a test that you can do after finishing Scenario 1.

TODO: rewrite this so that it makes sense after Scenario 4.

Bring down the control plane: stop and restart the ITS and WDS API servers,
KubeFlex and KubeStellar controllers:

First stop all:

```shell
kubectl --context "$host_context" scale deployment -n "$wds_cp"-system kube-apiserver --replicas=0
kubectl --context "$host_context" scale statefulset -n "$its_cp"-system vcluster --replicas=0
kubectl --context "$host_context" scale deployment -n kubeflex-system kubeflex-controller-manager --replicas=0
kubectl --context "$host_context" scale deployment -n "$wds_cp"-system kubestellar-controller-manager --replicas=0
kubectl --context "$host_context" scale deployment -n "$wds_cp"-system transport-controller --replicas=0
```

Then restart all:

```shell
kubectl --context "$host_context" scale deployment -n "$wds_cp"-system kube-apiserver --replicas=1
kubectl --context "$host_context" scale statefulset -n "$its_cp"-system vcluster --replicas=1
kubectl --context "$host_context" scale deployment -n kubeflex-system kubeflex-controller-manager --replicas=1
kubectl --context "$host_context" scale deployment -n "$wds_cp"-system kubestellar-controller-manager --replicas=1
kubectl --context "$host_context" scale deployment -n "$wds_cp"-system transport-controller --replicas=1
```

Wait for about a minute for all pods to restart, then apply a new BindingPolicy:

```shell
kubectl --context "$wds_context" apply -f - <<EOF
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: nginx-res-bpolicy
spec:
  clusterSelectors:
  - matchLabels: {$(echo "$label_query_both" | tr , $'\n' | while IFS="=" read key val; do echo -n ", \"$key\": \"$val\""; done | tail +3c)}
  downsync:
  - objectSelectors:
    - matchLabels: {"app.kubernetes.io/name":"nginx-res"}
EOF
```

and a new workload:

```shell
kubectl --context "$wds_context" apply -f - <<EOF
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
kubectl --context "$wec1_context" get deployments -n nginx-res
kubectl --context "$wec2_context" get deployments -n nginx-res
```

### [Optional] Teardown Scenario 5

```shell
kubectl --context "$wds_context" delete ns nginx-res
kubectl --context "$wds_context" delete bindingpolicies nginx-res-bpolicy
```

## Scenario 6 - multi-cluster workload deployment of app with ServiceAccount with ArgoCD

Before running this scenario, install ArgoCD on the hosting cluster and configure it
work with the WDS as outlined [here](argo-to-wds1.md).

Including a ServiceAccount tests whether there will be a controller fight over a token Secret for that ServiceAccount, which was observed in some situations with older code.

Apply the following BindingPolicy to the WDS:

```shell
kubectl --context "$wds_context" apply -f - <<EOF
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: argocd-sa-bpolicy
spec:
  clusterSelectors:
  - matchLabels: {$(echo "$label_query_both" | tr , $'\n' | while IFS="=" read key val; do echo -n ", \"$key\": \"$val\""; done | tail +3c)}
  downsync:
  - objectSelectors:
    - matchLabels: {"argocd.argoproj.io/instance":"nginx-sa"}
EOF
```

Switch context to hosting cluster and argocd namespace (this is required by argo to
create an app with the CLI)

```shell
kubectl config use-context "$host_context"
kubectl config set-context --current --namespace=argocd
```

Create a new application in ArgoCD:

```shell
argocd app create nginx-sa --repo https://github.com/pdettori/sample-apps.git --path nginx --dest-server https://"${wds_cp}.{wds_cp}-system" --dest-namespace nginx-sa
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
kubectl --context "$wec1_context" -n nginx-sa get deployments,sa,secrets
kubectl --context "$wec2_context" -n nginx-sa get deployments,sa,secrets
```

Repeat multiple syncing on Argo and verify that extra secrets for the service account
are not created in the WDS and both clusters:

```shell
kubectl --context "$wds_context" -n nginx-sa get secrets
kubectl --context "$wec1_context" -n nginx-sa get secrets
kubectl --context "$wec2_context" -n nginx-sa get secrets
```

### [Optional] Teardown Scenario 6

(Assuming that kubectl is still using the context for the hosting cluster and namespace `argocd`.)

```shell
argocd app delete nginx-sa --cascade
kubectl --context "$wds_context" delete bindingpolicies argocd-sa-bpolicy
```
