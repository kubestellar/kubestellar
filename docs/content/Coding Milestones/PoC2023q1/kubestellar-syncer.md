---
short_name: kubestellar-syncer
manifest_name: 'docs/content/Coding Milestones/PoC2023q1/kubestellar-syncer.md'
pre_req_name: 'docs/content/common-subs/pre-req.md'
---
[![docs-ecutable - kubestellar-syncer]({{config.repo_url}}/actions/workflows/docs-ecutable-syncer.yml/badge.svg?branch={{config.ks_branch}})]({{config.repo_url}}/actions/workflows/docs-ecutable-syncer.yml)&nbsp;&nbsp;&nbsp;
{%
   include-markdown "../../common-subs/required-packages.md"
   start="<!--required-packages-start-->"
   end="<!--required-packages-end-->"
%}
{%
   include-markdown "../../common-subs/save-some-time.md"
   start="<!--save-some-time-start-->"
   end="<!--save-some-time-end-->"
%}

KubeStellar Syncer runs in the target cluster and sync kubernetes resource objects from the target cluster to a mailbox workspace and vice versa.

![kubestellar-syncer drawio](images/kubestellar-syncer-overview.png)


## Steps to try the syncer
The KubeStellar Syncer can be exercised after setting up KubeStellar mailbox workspaces. Firstly we'll follow to similar steps in [example1](../example1) until `The mailbox controller` in stage 2. 

{%
   include-markdown "example1-subs/example1-pre-kcp.md"
   start="<!--example1-pre-kcp-start-->"
   end="<!--example1-pre-kcp-end-->"
%}

{%
   include-markdown "example1-subs/example1-start-kcp.md"
   start="<!--example1-start-kcp-start-->"
   end="<!--example1-start-kcp-end-->"
%}

{%
   include-markdown "example1-subs/example1-post-kcp.md"
   start="<!--example1-post-kcp-start-->"
   end="<!--example1-post-kcp-end-->"
%}

{%
   include-markdown "example1-subs/example1-stage-1a.md"
   start="<!--example1-stage-1a-start-->"
   end="<!--example1-stage-1a-end-->"
%}

### Register KubeStellar Syncer on the target clusters

Once KubeStellar setup is done, KubeStellar Syncer can be deployed on the target cluster easily by the following steps.
#### For the target cluster of `guilder`,
{%
   include-markdown "kubestellar-syncer-subs/kubestellar-syncer-0-deploy-guilder.md"
   start="<!--kubestellar-syncer-0-deploy-guilder-start-->"
   end="<!--kubestellar-syncer-0-deploy-guilder-end-->"
%}

#### For the target cluster of `florin`,
{%
   include-markdown "kubestellar-syncer-subs/kubestellar-syncer-0-deploy-florin.md"
   start="<!--kubestellar-syncer-0-deploy-florin-start-->"
   end="<!--kubestellar-syncer-0-deploy-florin-end-->"
%}

### Teardown the environment

{%
   include-markdown "../../common-subs/teardown-the-environment.md"
   start="<!--teardown-the-environment-start-->"
   end="<!--teardown-the-environment-end-->"
%}

---
## Deep-dive

### The details about the registration of KubeStellar Syncer on an Edge cluster and a workspace

Edge-syncer is deployed on Edge cluster easily by the following steps.

1. Create SyncTarget and Location
    - Mailbox controller creates mailbox workspace automatically. 
2. Get mailbox workspace name
3. Use command to obtain yaml manifests to bootstrap KubeStellar Syncer
    ```console
    kubectl ws <mb-ws name>
    bin/kubectl-kubestellar-syncer_gen <Edge Sync Target name> --syncer-image <KubeStellar-Syncer image> -o kubestellar-syncer.yaml
    ```
    Here `bin/kubectl-kubestellar-syncer_gen` refers to a special variant of KubeStellar's
    kubectl plugin that includes the implementation of the functionality needed
    here.  This variant, under the special name shown here, is a normal part of
    the `bin` of edge-mc.
    For the KubeStellar Syncer image, please select an official image from https://quay.io/repository/kubestellar/syncer?tab=tags. For example, `--syncer-image quay.io/kubestellar/syncer:v0.2.2`. You can also create a syncer image from the source following [Build KubeStellar Syncer Image](#build-kubestellar-syncer-image).
4. Deploy edge-syncer on an Edge cluster
5. Syncer starts to run on the Edge cluster
    - KubeStellar Syncer starts watching and consuming SyncerConfig

The overall diagram is as follows:

![kubestellar-syncer boot](images/kubestellar-syncer-boot.png)

### What kubestellar syncer-gen plugin does

In order for Syncer to sync resources between upstream (workspace) and downstream (physical cluster), both access information are required. For the upstream access, the registration command of Syncer (`kubectl kubestellar syncer-gen`) creates a service account, clusterrole, and clusterrolebinding in the workspace, and then generates kubeconfig manifest from the service account token, KCP server URL, and the server certificates. The kubeconfig manifest is embedded in a secret manifest and the secret is mount to `/kcp/` in Syncer pod. The command generates such deployment manifest as Syncer reads `/kcp/` for the upstream Kubeconfig. On the other hand, for the downstream part, in addition to the deployment manifest, the command generates a service account, role/clusterrole, rolebinding/clusterrolebinding for Syncer to access resources on the physical cluster. These resources for the downstream part are the resources to be deployed to downstream cluster. The serviceaccount is set to `serviceAccountName` in the deployment manifest.

Note: In addition to that, the command creates EdgeSyncConfig CRD if not exist, and creates a default EdgeSyncConfig resource with the name specified in the command (;`kubectl kubestellar syncer-gen <name>`). The default EdgeSyncConfig is no longer needed since Syncer now consumes all EdgeSyncConfigs in the workspace. Furthermore, creation of EdgeSyncConfig CRD will also no longer be needed since we will switch to use SyncerConfig rather than EdgeSyncConfig in near future.

The source code of the command is [{{ config.repo_url }}/blob/{{ config.ks_branch }}/pkg/cliplugins/kcp-edge/syncer-gen/edgesync.go]({{ config.repo_url }}/blob/{{ config.ks_branch }}/pkg/cliplugins/kcp-edge/syncer-gen/edgesync.go).

The equivalent manual steps are as follows:

{%
   include-markdown "kubestellar-syncer-subs/kubestellar-syncer-1-syncer-gen-plugin.md"
   start="<!--kubestellar-syncer-1-syncer-gen-plugin-start-->"
   end="<!--kubestellar-syncer-1-syncer-gen-plugin-end-->"
%}

### Deploy workload objects from edge-mc to Edge cluster

To deploy resources to Edge clusters, create the following in workload management workspace
- workload objects
  - Some objects are denatured if needed.
  - Other objects are as it is
- APIExport/API Schema corresponding to CRD such as Kubernetes [ClusterPolicyReport](https://github.com/kubernetes-sigs/wg-policy-prototypes/blob/master/policy-report/crd/v1beta1/wgpolicyk8s.io_clusterpolicyreports.yaml).
  - TBD: Conversion from CRD to APIExport/APISchema could be automated by using MutatingAdmissionWebhook on workload management workspace. This automation is already available (see the script [here]({{ config.repo_url }}/blob/{{ config.ks_branch }}/hack/update-codegen-crds.sh#L57)). 
- EdgePlacement

![kubestellar-syncer deploy](images/kubestellar-syncer-deploy.png)

After this, Edge-mc will put the following in the mailbox workspace.
- Workload objects (both denatured one and not-denatured one)
- SyncerConfig CR

**TODO**: This is something we should clarify..e.g. which existing controller(s) in edge-mc will cover, or just create a new controller to handle uncovered one. @MikeSpreitzer gave the following suggestions.
  - The placement translator will put the workload objects and syncer config into the mailbox workspaces.
  - The placement translator will create syncer config based on the EdgePlacement objects and what they match.
  - The mailbox controller will put API Binding into the mailbox workspace.

### EdgeSyncConfig (will be replaced to SyncerConfig)
- The example of EdgeSyncConfig CR is [here]({{ config.repo_url }}/blob/{{ config.ks_branch }}/pkg/syncer/scripts/edge-sync-config-for-kyverno-helm.yaml). Its CRD is [here]({{ config.repo_url }}/blob/{{ config.ks_branch }}/config/crds/edge.kcp.io_edgesyncconfigs.yaml).
- The CR here is used from edge syncer. 
- The CR is placed at mb-ws to define
  - object selector
  - need of renaturing
  - need of returning reported states of downsynced objects
  - need of delete propagation for downsyncing
- The CR is managed by edge-mc (placement transformer).
  - At the initial implementation before edge-mc side controller become ready, we assume SyncerConfig is on workload management workspace (wm-ws), and then which will be copied into mb-ws like other workload objects.
  - This should be changed to be generated according to EdgePlacement spec. 
- This CR is a placeholder for defining how edge-syncer behaves, and will be extended/split/merged according to further design discussion.
- One CR is initially created by the command for KubeStellar Syncer enablement in mb-ws (`kubectl kubestellar syncer-gen <name>`)
  - The CR name is `<name>` and the contents are empty.
  - This name is registered in the bootstrap manifests for KubeStellar Syncer install and KubeStellar Syncer is told to watch the CR of this name.
- Currently KubeStellar Syncer watches all CRs in the workspace
  - KubeStellar Syncer merges them and decides which resources are down/up synced based on the merged information. 
  - This behavior may be changed to only watching the default CR once Placement Translator is to be the component that generates the CR from EdgePlacement: [related issue]({{ config.repo_url }}/pull/284#pullrequestreview-1375667129)

### SyncerConfig
- The spec is defined in {{ config.repo_url }}/blob/{{ config.ks_branch }}/pkg/apis/edge/v1alpha1/syncer-config.go
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
- The example CR is {{ config.repo_url }}/blob/{{ config.ks_branch }}/test/e2e/kubestellar-syncer/testdata/kyverno/syncer-config.yaml
- The CR is used from KubeStellar-Syncer
- The CR is placed in mb-ws to define
  - object selector
  - need of renaturing (May not scope in PoC2023q1)
  - need of returning reported states of downsynced objects (May not scope in PoC2023q1)
  - need of delete propagation for downsyncing (May not scope in PoC2023q1)
- The CR is managed by edge-mc (placement translator).
  - At the initial implementation before edge-mc side controller become ready, we assume SyncerConfig is on workload management workspace (wm-ws), and then which will be copied into mb-ws like other workload objects.
  - This should be changed to be generated according to EdgePlacement spec. 
- This CR is a placeholder for defining how edge-syncer behaves, and will be extended/split/merged according to further design discussion.
- Currently KubeStellar Syncer watches all CRs in the workspace
  - KubeStellar Syncer merges them and decides which resources are down/up synced based on the merged information. 

### Downsyncing

- Edge syncer does downsyncing, which copy workload objects on mailbox workspace to Edge cluster
- If workload objects are deleted on mailbox workspace, the corresponding objects on the Edge cluster will be also deleted according to SyncerConfig. 
- SyncerConfig specifies which objects should be downsynced.
  - object selector: group, version, kind, name, namespace (for namespaced objects), label, annotation
- Cover cluster-scope objects and CRD
  - CRD needs to be denatured if downsyncing is required. (May not scope in PoC2023q1 since no usage)
- Renaturing is applied if required (specified in SyncerConfig). (May not scope in PoC2023q1 since no usage)
- Current implementation is using polling to detect changes on mailbox workspace, but will be changed to use Informers. 

### Renaturing (May not scope in PoC2023q1 since no usage)
- Edge syncer does renaturing, which converts workload objects to different forms of objects on a Edge cluster. 
- The conversion rules (downstream/upstream mapping) is specified in SyncerConfig.
- Some objects need to be denatured. 
  - CRD needs to be denatured when conflicting with APIBinding.

### Return of reported state
- Edge syncer return the reported state of downsynced objects at Edge cluster to the status of objects on the mailbox workspace periodically. 
  - TODO: Failing to returning reported state of some resources (e.g. deployment and service). Need more investigation. 
- reported state returning on/off is configurable in SyncerConfig. (default is on)

### Resource Upsyncing
- KubeStellar-Syncer does upsyncing resources at Edge cluster to the corresponding mailbox workspace periodically. 
- SyncerConfig specifies which objects should be upsynced from Edge cluster.
  - object selector: group, version, kind, name, namespace (for namespaced objects), label, annotation (, and more such as ownership reference?)
- Upsyncing CRD is out of scope for now. This means when upsyncing a CR, corresponding APIBinding (not CRD) is available on the mailbox workspace. This limitation might be revisited later. 
- ~Upsynced objects can be accessed from APIExport set on the workload management workspace bound to the mailbox workspace (with APIBinding). This access pattern might be changed when other APIs such as summarization are provided in edge-mc.~ => Upsynced objects are accessed through Mailbox informer.

### Feasibility study
We will verify if the design described here could cover the following 4 scenarios. 
- I can register an edge-syncer on a Edge cluster to connect a mailbox workspace specified by name. (edge-syncer registration)
- I can deploy Kyverno and its policy from mailbox workspace to Edge cluster just by using manifests (generated from Kyverno helm chart) rather than using OLM. (workload deployment by edge-syncer's downsyncing)
- I can see the policy report generated at Edge cluster via API Export on workload management workspace. (resource upsyncing by edge-syncer) 
- I can deploy the denatured objects on mailbox workspace to Edge cluster by renaturing them automatically in edge-syncer. (workload deployment by renaturing)

## Build KubeStellar Syncer image

Prerequisite
- Install ko (https://ko.build/install/)

### How to build the image in your local
1. `make build-kubestellar-syncer-image-local`
e.g.
```
$ make build-kubestellar-syncer-image-local
2023/04/24 11:50:37 Using base distroless.dev/static:latest@sha256:81018475098138883b80dcc9c1242eb02b53465297724b18e88591a752d2a49c for github.com/kcp-dev/edge-mc/cmd/syncer
2023/04/24 11:50:38 Building github.com/kcp-dev/edge-mc/cmd/syncer for linux/arm64
2023/04/24 11:50:39 Loading ko.local/syncer-273dfcc28dbb16dfcde62c61d54e1ca9:c4759f6f841075649a22ff08bdf4afe32600f8bb31743d1aa553454e07375c96
2023/04/24 11:50:40 Loaded ko.local/syncer-273dfcc28dbb16dfcde62c61d54e1ca9:c4759f6f841075649a22ff08bdf4afe32600f8bb31743d1aa553454e07375c96
2023/04/24 11:50:40 Adding tag latest
2023/04/24 11:50:40 Added tag latest
kubestellar-syncer image:
ko.local/syncer-273dfcc28dbb16dfcde62c61d54e1ca9:c4759f6f841075649a22ff08bdf4afe32600f8bb31743d1aa553454e07375c96
```
`ko.local/syncer-273dfcc28dbb16dfcde62c61d54e1ca9:c4759f6f841075649a22ff08bdf4afe32600f8bb31743d1aa553454e07375c96` is the image stored in your local Docker registry.

You can also set a shell variable to the output of this Make task.

For example
```
image=`make build-kubestellar-syncer-image-local`
```

### How to build the image with multiple architectures and push it to Docker registry
1. `make build-kubestellar-syncer-image DOCKER_REPO=ghcr.io/yana1205/edge-mc/syncer IMAGE_TAG=dev-2023-04-24-x ARCHS=linux/amd64,linux/arm64`

For example
```
$ make build-kubestellar-syncer-image DOCKER_REPO=ghcr.io/yana1205/edge-mc/syncer IMAGE_TAG=dev-2023-04-24-x ARCHS=linux/amd64,linux/arm64
2023/04/24 11:50:16 Using base distroless.dev/static:latest@sha256:81018475098138883b80dcc9c1242eb02b53465297724b18e88591a752d2a49c for github.com/kcp-dev/edge-mc/cmd/syncer
2023/04/24 11:50:17 Building github.com/kcp-dev/edge-mc/cmd/syncer for linux/arm64
2023/04/24 11:50:17 Building github.com/kcp-dev/edge-mc/cmd/syncer for linux/amd64
2023/04/24 11:50:18 Publishing ghcr.io/yana1205/edge-mc/syncer:dev-2023-04-24-x
2023/04/24 11:50:19 existing blob: sha256:85a5162a65b9641711623fa747dab446265400043a75c7dfa42c34b740dfdaba
2023/04/24 11:50:20 pushed blob: sha256:00b7b3ca30fa5ee9336a9bc962efef2001c076a3149c936b436f409df710b06f
2023/04/24 11:50:21 ghcr.io/yana1205/edge-mc/syncer:sha256-a52fb1cf432d321b278ac83600d3b83be3b8e6985f30e5a0f6f30c594bc42510.sbom: digest: sha256:4b1407327a486c0506188b67ad24222ed7924ba57576e47b59a4c1ac73dacd40 size: 368
2023/04/24 11:50:21 Published SBOM ghcr.io/yana1205/edge-mc/syncer:sha256-a52fb1cf432d321b278ac83600d3b83be3b8e6985f30e5a0f6f30c594bc42510.sbom
2023/04/24 11:50:21 existing blob: sha256:930413008565fd110e7ab2d37aab538449f058e7d83e7091d1aa0930a0086f58
2023/04/24 11:50:22 pushed blob: sha256:bd830efcc6c0a934a273202ffab27b1a8927368a7b99c4ae0cf850fadb865ead
2023/04/24 11:50:23 ghcr.io/yana1205/edge-mc/syncer:sha256-02db9874546b79ee765611474eb647128292e8cda92f86ca1b7342012eb79abe.sbom: digest: sha256:5c79e632396b893c3ecabf6b9ba43d8f20bb3990b0c6259f975bf81c63f0e41e size: 369
2023/04/24 11:50:23 Published SBOM ghcr.io/yana1205/edge-mc/syncer:sha256-02db9874546b79ee765611474eb647128292e8cda92f86ca1b7342012eb79abe.sbom
2023/04/24 11:50:24 existing blob: sha256:bb5ef9628a98afa48a9133f5890c43ed1499eb82a33fe173dd9067d7a9cdfb0a
2023/04/24 11:50:25 pushed blob: sha256:61f19080792ae91e8b37ecf003376497b790a411d7a8fa4435c7457b0e15874c
2023/04/24 11:50:25 ghcr.io/yana1205/edge-mc/syncer:sha256-c4759f6f841075649a22ff08bdf4afe32600f8bb31743d1aa553454e07375c96.sbom: digest: sha256:8d82388bb534933d7193c661743fca8378cc561a2ad8583c0107f687acb37c1b size: 369
2023/04/24 11:50:25 Published SBOM ghcr.io/yana1205/edge-mc/syncer:sha256-c4759f6f841075649a22ff08bdf4afe32600f8bb31743d1aa553454e07375c96.sbom
2023/04/24 11:50:26 existing manifest: sha256:02db9874546b79ee765611474eb647128292e8cda92f86ca1b7342012eb79abe
2023/04/24 11:50:26 existing manifest: sha256:c4759f6f841075649a22ff08bdf4afe32600f8bb31743d1aa553454e07375c96
2023/04/24 11:50:27 ghcr.io/yana1205/edge-mc/syncer:dev-2023-04-24-x: digest: sha256:a52fb1cf432d321b278ac83600d3b83be3b8e6985f30e5a0f6f30c594bc42510 size: 690
2023/04/24 11:50:27 Published ghcr.io/yana1205/edge-mc/syncer:dev-2023-04-24-x@sha256:a52fb1cf432d321b278ac83600d3b83be3b8e6985f30e5a0f6f30c594bc42510
echo KO_DOCKER_REPO=ghcr.io/yana1205/edge-mc/syncer ko build --platform=linux/amd64,linux/arm64 --bare --tags ./cmd/syncer
KO_DOCKER_REPO=ghcr.io/yana1205/edge-mc/syncer ko build --platform=linux/amd64,linux/arm64 --bare --tags ./cmd/syncer
kubestellar-syncer image
ghcr.io/yana1205/edge-mc/syncer:dev-2023-04-24-x@sha256:a52fb1cf432d321b278ac83600d3b83be3b8e6985f30e5a0f6f30c594bc42510
```
`ghcr.io/yana1205/edge-mc/syncer:dev-2023-04-24-x@sha256:a52fb1cf432d321b278ac83600d3b83be3b8e6985f30e5a0f6f30c594bc42510` is the image pushed to the registry.

## Teardown the environment

{%
   include-markdown "../../common-subs/teardown-the-environment.md"
   start="<!--teardown-the-environment-start-->"
   end="<!--teardown-the-environment-end-->"
%}
