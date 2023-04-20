---
title: "KCP-Edge Syncer"
date: 2023-03-22
weight: 4
description: >
---

{{% pageinfo %}}
Edge syncer is a new syncer working with mailbox workspace in KCP-Edge. 
{{% /pageinfo %}}

![edge-syncer drawio](/docs/Coding%20Milestones/PoC2023q1/images/edge-syncer-overview.png)

#### Registering Edge Syncer on an Edge cluster

Edge-syncer can be deployed on Edge cluster easily by the following steps.
1. Create SyncTarget and Location
    - Mailbox controller creates mailbox workspace automatically. 
1. Get mailbox workspace name
1. Use command to register edge-syncer and obtain yaml manifests to bootstrap Edge Syncer
    ```console
    kubectl ws <mb-ws name>
    kubectl kcp workload edge-sync <EM Sync Target name> --syncer-image <EM Syncer Image> -o edge-syncer.yaml
    ```
1. Deploy edge-syncer on an Edge cluster
1. Syncer starts to run on the Edge cluster
    - Edge Syncer starts watching and consuming SyncerConfig

The overall diagram is as follows:

![edge-syncer boot](/docs/Coding%20Milestones/PoC2023q1/images/edge-syncer-boot.png)

#### What edge-sync plugin does

In order for Syncer to sync resources between upstream (workspace) and doenstream (physical cluster), both access information are required. For the upstream access, the registration command of Syncer (`kubectl kcp workload edge-sync`) creates a service account, clusterrole, and clusterrolebinding in the workspace, and then generates kubeconfig manifest from the service account token, KCP server URL, and the server certificates. The kubeconfig manifest is embedded in a secret manifest and the secret is mount to `/kcp/` in Syncer pod. The command generates such deployment manifest as Syncer reads `/kcp/` for the upstream Kubeconfig. On the other hand, for the downstream part, in addition to the deployment manifest, the command generates a service account, role/clusterrole, rolebinding/clusterrolebinding for Syncer to access resources on the physical cluster. These resources for the downstream part are the resources to be deployed to downstream cluster. The serviceaccount is set to `serviceAccountName` in the deployment manifest.

Note: In addtion to that, the command creates EdgeSyncConfig CRD if not exist, and creates a deffault EdgeSyncConfig resource with the name specified in the command (;`kubectl kcp workload edge-sync <name>`). The default EdgeSyncConfig is no longer needed since Syncer now consumes all EdgeSyncConfigs in the workspace. Furthermore, creation of EdgeSyncConfig CRD will also no longer be needed since we will switch to use SyncerConfig rather than EdgeSyncConfig in near future.

The source code of the command is https://github.com/yana1205/kcp/blob/emc/pkg/cliplugins/workload/plugin/edgesync.go.

The equivalent manual steps are as follows:
1. Generate UUID for Syncer identification
    ```
    uuidgen | tr '[:upper:]' '[:lower:]' | read syncer_id
    syncer_id="syncer-$syncer_id"
    ```
1. Go to a workspace (It's exactly a mailbox workspace in the case of Edge MC)
    ```
    kubectl ws create ws1 --enter
    ```
1. Create a serviceaccount in the workspace
    ```
    cat << EOL | kubectl apply -f -
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: $syncer_id
    EOL
    ```
1. Create clusterrole and clusterrolebinding to bind the serviceaccount to the role
    ```
    cat << EOL | kubectl apply -f -
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRole
    metadata:
      name: $syncer_id
    rules:
    - apiGroups: ["*"]
      resources: ["*"]
      verbs: ["*"]
    - nonResourceURLs: ["/"]
      verbs: ["access"]
    ---
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRoleBinding
    metadata:
      name: $syncer_id
    roleRef:
      apiGroup: rbac.authorization.k8s.io
      kind: ClusterRole
      name: $syncer_id
    subjects:
    - apiGroup: ""
      kind: ServiceAccount
      name: $syncer_id
      namespace: default
    EOL
    ```
1. Get the serviceaccount token that will be set in the upstream kubeconfig manifest
    ```
    kubectl get secret -o custom-columns=":.metadata.name"| grep $syncer_id | read secret_name
    kubectl get secret $secret_name -o jsonpath='{.data.token}' | base64 -d | read token
    ```
1. Get the certificates that will be set in the upstream kubeconfig manifest
    ```
    kubectl config view --minify --raw | yq ".clusters[0].cluster.certificate-authority-data" | read cacrt
    echo $cacrt
    ```
1. Get the server host and port that will be set in the upstream kubeconfig manifest
    ```
    kubectl config view --minify --raw | yq ".clusters[0].cluster.server" | sed -e 's|https://\(.*\):\([^/]*\)/.*|\1 \2|g' | read host port
    echo $host, $port
    ```
1. Set some other parameters
    1. server_url of KCP from host and port
        ```
        server_url="https://$host:$port"
        ```
    1. downstream_namespace where Syncer Pod runs
        ```
        downstream_namespace="kcp-edge-$syncer_id"
        ```
    1. Syncer image
        ```
        image="quay.io/kcpedge/syncer:dev-2023-03-30"
        ```
1. Generate manifests to bootstrap Edge Syncer
    ```
    syncer_id=$syncer_id cacrt=$cacrt token=$token server_url=$server_url downstream_namespace=$downstream_namespace image=$image envsubst < ./pkg/syncer/scripts/edge-syncer-bootstrap.template.yaml
    ```
    For debug purpose, you can extract kubeconfig.yaml of the upstream
    ```
    syncer_id=$syncer_id cacrt=$cacrt token=$token server_url=$server_url downstream_namespace=$downstream_namespace image=$image envsubst < ./pkg/syncer/scripts/edge-syncer-bootstrap.template.yaml | yq e "select(.kind == \"Secret\" and .metadata.name == \"$syncer_id\")" | yq .stringData.kubeconfig 
    ```
1. For now, EdgeSyncConfig API is required. Please create EdgeSyncConfig CRD in the workspace, if you run Syncer from the generated bootstrap manifest.
    ```
    kubectl create ./pkg/syncer/config/crds/edge.kcp.io_edgesyncconfigs.yaml
    ```

#### Deploy workload objects from edge-mc to Edge cluster

To deploy resources to Edge clusters, create the following in workload management workspace
- workload objects
  - Some objects are denatured if needed.
  - Other objects are as it is
- APIExport/API Schema corresponding to CRD such as Kubernetes [ClusterPolicyReport](https://github.com/kubernetes-sigs/wg-policy-prototypes/blob/master/policy-report/crd/v1beta1/wgpolicyk8s.io_clusterpolicyreports.yaml).
  - TBD: Conversion from CRD to APIExport/APISchema could be automated by using MutatingAdmissionWebhook on workload management workspace. This automation is already available (see the sciprt [here](https://github.com/kcp-dev/edge-mc/blob/main/hack/update-codegen-crds.sh#L57)). 
- EdgePlacement

![edge-syncer deploy](/docs/Coding%20Milestones/PoC2023q1/images/edge-syncer-deploy.png)

After this, Edge-mc will put the following in the mailbox workspace.
- Workload objects (both denatured one and not-denatured one)
- SyncerConfig CR

**TODO**: This is something we should clarify..e.g. which existing controller(s) in edge-mc will cover, or just create a new controller to handle uncovered one. @MikeSpreitzer gave the following suggestions.
  - The placement translator will put the workload objects and syncer config into the mailbox workspaces.
  - The placement translator will create syncer config based on the EdgePlacement objects and what they match.
  - The mailbox controller will put API Binding into the mailbox workspace.

#### EdgeSyncConfig (will be replaced to SyncerConfig)
- The example of EdgeSycnerConfig CR is [here](https://github.com/yana1205/edge-mc/blob/edge-syncer/pkg/syncer/scripts/edge-sync-config-for-kyverno-helm.yaml). Its CRD is [here](https://github.com/yana1205/edge-mc/blob/edge-syncer/pkg/syncer/config/crds/edge.kcp.io_edgesyncconfigs.yaml).
- The CR here is used from edge syncer. 
- The CR is placed at mb-ws to define
  - object selector
  - need of renaturing
  - need of returning reported states of downsynced objects
  - need of delete propagation for downsyncing
- The CR is managed by edge-mc (placement transformer).
  - At the initial implementation before edge-mc side controller become ready, we assume SyncerConfig is on workload management workspace (wm-ws), and then which will be copied into mb-ws like other workload objects.
  - This should be changed to be generated according to EdgePlacement spec. 
- This CR is a placeholder for defining how edge-syncer behaves, and will be extended/splitted/merged according to further design discussion.
- One CR is initially created by the command for Edge Syncer enablement in mb-ws (`kubectl kcp workload edge-syncer <name>`)
  - The CR name is `<name>` and the contents are empty.
  - This name is registered in the bootstrap manifests for Edge Syncer install and Edge Syncer is told to watch the CR of this name.
- Currently Edge Syncer watches all CRs in the workspace
  - Edge Syncer merges them and decides which resources are down/up synced based on the merged information. 
  - This behavior may be changed to only watching the default CR once Placement Translater is to be the component that generates the CR from EdgePlacement: [related issue](https://github.com/kcp-dev/edge-mc/pull/284#pullrequestreview-1375667129)

#### SyncerConfig
- The spec is defined in https://github.com/kcp-dev/edge-mc/blob/main/pkg/apis/edge/v1alpha1/syncer-config.go
  - `namespaceScope` field is for namespace scoped objects.
    - `namespaces` is field for which namespaces to be downsynced.
    - `resources` is field for what resource's objects in the above namespaces are downsynced. All objects in the selected resource are downsynced.
  - `clusterScope` field is for cluster scoped objects
    - It's an array of `apiVersion`, `group`, `resource`, and `objects`.
    - `objects` can be specified by wildcard (`*`) meaning all objects.
  - `upsync` field is for upsynced objects including both namespace and cluster scoped objects.
    - It's an array of `apiGroup`, `resources`, `namespaces`, and `names`.
    - `apiGroup` is group.
    - `resources` is an array of upsynced resource.
    - `namespaces` is an array of namespace for namespace objects.
    - `names` is an array of upsynced object name. Wildcard (`*`) is available.
- The example CR is https://github.com/yana1205/edge-mc/blob/support-syncer-config/test/e2e/edgesyncer/testdata/kyverno/syncer-config.yaml
- The CR is used from edge syncer
- The CR is placed in mb-ws to define
  - object selector
  - need of renaturing (May not scope in PoC2023q1)
  - need of returning reported states of downsynced objects (May not scope in PoC2023q1)
  - need of delete propagation for downsyncing (May not scope in PoC2023q1)
- The CR is managed by edge-mc (placement translator).
  - At the initial implementation before edge-mc side controller become ready, we assume SyncerConfig is on workload management workspace (wm-ws), and then which will be copied into mb-ws like other workload objects.
  - This should be changed to be generated according to EdgePlacement spec. 
- This CR is a placeholder for defining how edge-syncer behaves, and will be extended/splitted/merged according to further design discussion.
- Currently Edge Syncer watches all CRs in the workspace
  - Edge Syncer merges them and decides which resources are down/up synced based on the merged information. 

#### Downsyncing

- Edge syncer does downsyncing, which copy workload objects on mailbox workspace to Edge cluster
- If workload objects are deleted on mailbox workspace, the corresponding objects on the Edge cluster will be also deleted according to SyncerConfig. 
- SyncerConfig specifies which objects should be downsynced.
  - object selector: group, version, kind, name, namespace (for namespaced objects), label, annotation
- Cover cluster-scope objects and CRD
  - CRD needs to be denatured if downsyncing is required. (May not scope in PoC2023q1 since no usage)
- Renaturing is applied if required (specified in SyncerConfig). (May not scope in PoC2023q1 since no usage)
- Current implementation is using polling to detect changes on mailbox workspace, but will be changed to use Informers. 

#### Renaturing (May not scope in PoC2023q1 since no usage)
- Edge syncer does renaturing, which converts workload objects to different forms of objects on a Edge cluster. 
- The conversion rules (downstream/upstream mapping) is specified in SyncerConfig.
- Some objects need to be denatured. 
  - CRD needs to be denatured when conflicting with APIBinding.

#### Return of reported state
- Edge syncer return the reported state of downsynced objects at Edge cluster to the status of objects on the mailbox workspace periodically. 
  - TODO: Failing to returning reported state of some resources (e.g. deployment and service). Need more investigation. 
- reported state returning on/off is configurable in SyncerConfig. (default is on)

#### Resource Upsyncing
- Edge syncer does upsyncing resources at Edge cluster to the corresponding mailbox workspace periodically. 
- SyncerConfig specifies which objects should be upsynced from Edge cluster.
  - object selector: group, version, kind, name, namespace (for namespaced objects), label, annotation (, and more such as ownership reference?)
- Upsyncing CRD is out of scope for now. This means when upsyncing a CR, corresponding APIBinding (not CRD) is available on the mailbox workspace. This limitation might be revisited later. 
- ~Upsynced objects can be accessed from APIExport set on the workload management workspace bound to the mailbox workspace (with APIBinding). This access pattern might be changed when other APIs such as summarization are provided in edge-mc.~ => Upsynced objects are accessed through Mailbox informer.

#### Feasibility study
We will verify if the design describled here could cover the following 4 scenarios. 
- I can register an edge-syncer on a Edge cluster to connect a mailbox workspace specified by name. (edge-syncer registration)
- I can deploy Kyverno and its policy from mailbox workspace to Edge cluster just by using manifests (generated from Kyverno helm chart) rather than using OLM. (workload deployment by edge-syncer's downsyncing)
- I can see the policy report generated at Edge cluster via API Export on workload management workspace. (resource upsyncing by edge-syncer) 
- I can deploy the denatured objects on mailbox workspace to Edge cluster by renaturing them automatically in edge-syncer. (workload deployment by renaturing)
