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

KubeStellar-Syncer runs in the target cluster and sync kubernetes resource objects from the target cluster to a mailbox workspace and vice versa.

![kubestellar-syncer drawio](images/kubestellar-syncer-overview.png)


## Steps to try the Syncer
The KubeStellar-Syncer can be exercised after setting up KubeStellar mailbox workspaces. Firstly we'll follow to similar steps in [example1](../example1) until `The mailbox controller` in stage 2. 

{%
   include-markdown "example1-subs/example1-pre-kcp.md"
   start="<!--example1-pre-kcp-start-->"
   end="<!--example1-pre-kcp-end-->"
%}

{%
   include-markdown "example1-subs/example1-space-manager.md"
   start="<!--example1-space-manager-start-->"
   end="<!--example1-space-manager-end-->"
%}

{%
   include-markdown "example1-subs/example1-start-kcp.md"
   start="<!--example1-start-kcp-start-->"
   end="<!--example1-start-kcp-end-->"
%}

{%
   include-markdown "example1-subs/example1-kcp-provider.md"
   start="<!--example1-kcp-provider-start-->"
   end="<!--example1-kcp-provider-end-->"
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

### Register KubeStellar-Syncer on the target clusters

Once KubeStellar setup is done, KubeStellar-Syncer can be deployed on the target cluster easily by the following steps.
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

### The details about the registration of KubeStellar-Syncer on an Edge cluster and a workspace

KubeStellar-Syncer is deployed on an Edge cluster easily by the following steps.

1. Create SyncTarget and Location
    - Mailbox controller creates the mailbox workspace automatically 
2. Get the mailbox workspace name
3. Use the following command to obtain yaml manifests to bootstrap KubeStellar-Syncer
    ```console
    kubectl ws <mb-ws name>
    bin/kubectl-kubestellar-syncer_gen <Edge Sync Target name> --syncer-image <KubeStellar-Syncer image> -o kubestellar-syncer.yaml
    ```
    Here, `bin/kubectl-kubestellar-syncer_gen` refers to a special variant of KubeStellar's
    kubectl plugin that includes the implementation of the functionality needed
    here.  This variant, under the special name shown here, is a normal part of
    the `bin` of KubeStellar.
    For the KubeStellar-Syncer image, please select an official image from https://quay.io/repository/kubestellar/syncer?tab=tags. For example, `--syncer-image quay.io/kubestellar/syncer:git-08289ea05-clean`. You can also create a syncer image from the source following [Build KubeStellar-Syncer Image](#build-kubestellar-syncer-image).
4. Deploy KubeStellar-Syncer on an Edge cluster
5. Syncer starts to run on the Edge cluster
    - KubeStellar-Syncer starts watching and consuming SyncerConfig

The overall diagram is as follows:

![kubestellar-syncer boot](images/kubestellar-syncer-boot.png)

### What KubeStellar syncer-gen plugin does

In order for Syncer to sync resources between upstream (workspace) and downstream (workload execution cluster), access information for both is required. 

For the upstream access, Syncer's registration command (`kubectl kubestellar syncer-gen`) creates a ServiceAccount, ClusterRole, and ClusterRoleBinding in the workspace, and then generates a kubeconfig manifest from the ServiceAccount token, KCP server URL, and the server certificates. The kubeconfig manifest is embedded in a secret manifest and the secret is mounted to `/kcp/` in Syncer pod. The command generates such deployment manifest as Syncer reads `/kcp/` for the upstream Kubeconfig. 

On the other hand, for the downstream part, in addition to the deployment manifest, the command generates a ServiceAccount, Role/ClusterRole, RoleBinding/ClusterRoleBinding for Syncer to access resources on the WEC. These resources for the downstream part are the resources to be deployed to the downstream cluster. The ServiceAccount is set to `serviceAccountName` in the deployment manifest.

Note: In addition to that, the command creates an EdgeSyncConfig CRD if it does not exist, and creates a default EdgeSyncConfig resource with the name specified in the command (`kubectl kubestellar syncer-gen <name>`). The default EdgeSyncConfig is no longer needed since Syncer now consumes all EdgeSyncConfigs in the workspace. Furthermore, creation of the EdgeSyncConfig CRD will also no longer be needed since we will switch to using SyncerConfig rather than EdgeSyncConfig in near future.

The source code of the command is [{{ config.repo_url }}/blob/{{ config.ks_branch }}/pkg/cliplugins/kubestellar/syncer-gen/edgesync.go]({{ config.repo_url }}/blob/{{ config.ks_branch }}/pkg/cliplugins/kubestellar/syncer-gen/edgesync.go).

The equivalent manual steps are as follows: 

{%
   include-markdown "kubestellar-syncer-subs/kubestellar-syncer-1-syncer-gen-plugin.md"
   start="<!--kubestellar-syncer-1-syncer-gen-plugin-start-->"
   end="<!--kubestellar-syncer-1-syncer-gen-plugin-end-->"
%}

### Deploy workload objects from KubeStellar to Edge cluster

To deploy resources to Edge clusters, create the following in workload management workspace
- workload objects
  - Some objects are denatured if needed.
  - Other objects are as it is
- APIExport/API Schema corresponding to CRD such as Kubernetes [ClusterPolicyReport](https://github.com/kubernetes-sigs/wg-policy-prototypes/blob/master/policy-report/crd/v1beta1/wgpolicyk8s.io_clusterpolicyreports.yaml).
  - TBD: Conversion from CRD to APIExport/APISchema could be automated by using MutatingAdmissionWebhook on workload management workspace. This automation is already available (see the script [here]({{ config.repo_url }}/blob/{{ config.ks_branch }}/hack/update-codegen-crds.sh#L57)). 
- EdgePlacement

![kubestellar-syncer deploy](images/kubestellar-syncer-deploy.png)

After this, KubeStellar will put the following in the mailbox workspace.
- Workload objects (both denatured one and not-denatured one)
- SyncerConfig CR

**TODO**: This is something we should clarify..e.g. which existing controller(s) in KubeStellar will cover, or just create a new controller to handle uncovered one. @MikeSpreitzer gave the following suggestions.
  - The placement translator will put the workload objects and syncer config into the mailbox workspaces.
  - The placement translator will create syncer config based on the EdgePlacement objects and what they match.
  - The mailbox controller will put API Binding into the mailbox workspace.

### EdgeSyncConfig (will be replaced to SyncerConfig)
- The example of EdgeSyncConfig CR is [here]({{ config.repo_url }}/blob/{{ config.ks_branch }}/pkg/syncer/scripts/edge-sync-config-for-kyverno-helm.yaml). Its CRD is [here]({{ config.repo_url }}/blob/{{ config.ks_branch }}/config/crds/edge.kubestellar.io_edgesyncconfigs.yaml).
- The CR here is used from edge syncer. 
- The CR is placed at mb-ws to define
  - object selector
  - need of renaturing
  - need of returning reported states of downsynced objects
  - need of delete propagation for downsyncing
- The CR is managed by KubeStellar (placement transformer).
  - At the initial implementation before KubeStellar side controller become ready, we assume SyncerConfig is on workload management workspace (WDS), and then which will be copied into mb-ws like other workload objects.
  - This should be changed to be generated according to EdgePlacement spec. 
- This CR is a placeholder for defining how KubeStellar-Syncer behaves, and will be extended/split/merged according to further design discussion.
- One CR is initially created by the command for KubeStellar-Syncer enablement in mb-ws (`kubectl kubestellar syncer-gen <name>`)
  - The CR name is `<name>` and the contents are empty.
  - This name is registered in the bootstrap manifests for KubeStellar-Syncer install and KubeStellar-Syncer is told to watch the CR of this name.
- Currently KubeStellar-Syncer watches all CRs in the workspace
  - KubeStellar-Syncer merges them and decides which resources are down/up synced based on the merged information. 
  - This behavior may be changed to only watching the default CR once Placement Translator is to be the component that generates the CR from EdgePlacement: [related issue]({{ config.repo_url }}/pull/284#pullrequestreview-1375667129)

### SyncerConfig
- The spec is defined in {{ config.repo_url }}/blob/{{ config.ks_branch }}/pkg/apis/edge/v2alpha1/syncer-config.go
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
- The CR is managed by KubeStellar (placement translator).
  - At the initial implementation before KubeStellar side controller become ready, we assume SyncerConfig is on workload management workspace (WDS), and then which will be copied into mb-ws like other workload objects.
  - This should be changed to be generated according to EdgePlacement spec. 
- This CR is a placeholder for defining how KubeStellar-Syncer behaves, and will be extended/split/merged according to further design discussion.
- Currently KubeStellar-Syncer watches all CRs in the workspace
  - KubeStellar-Syncer merges them and decides which resources are down/up synced based on the merged information. 

### Downsyncing

- KubeStellar-Syncer does downsyncing, which copy workload objects on mailbox workspace to Edge cluster
- If workload objects are deleted on mailbox workspace, the corresponding objects on the Edge cluster will be also deleted according to SyncerConfig. 
- SyncerConfig specifies which objects should be downsynced.
  - object selector: group, version, kind, name, namespace (for namespaced objects), label, annotation
- Cover cluster-scope objects and CRD
  - CRD needs to be denatured if downsyncing is required. (May not scope in PoC2023q1 since no usage)
- Renaturing is applied if required (specified in SyncerConfig). (May not scope in PoC2023q1 since no usage)
- Current implementation is using polling to detect changes on mailbox workspace, but will be changed to use Informers. 

### Renaturing (May not scope in PoC2023q1 since no usage)
- KubeStellar-Syncer does renaturing, which converts workload objects to different forms of objects on a Edge cluster. 
- The conversion rules (downstream/upstream mapping) is specified in SyncerConfig.
- Some objects need to be denatured. 
  - CRD needs to be denatured when conflicting with APIBinding.

### Return of reported state
- KubeStellar-Syncer return the reported state of downsynced objects at Edge cluster to the status of objects on the mailbox workspace periodically. 
  - TODO: Failing to returning reported state of some resources (e.g. deployment and service). Need more investigation. 
- reported state returning on/off is configurable in SyncerConfig. (default is on)

### Resource Upsyncing
- KubeStellar-Syncer does upsyncing resources at Edge cluster to the corresponding mailbox workspace periodically. 
- SyncerConfig specifies which objects should be upsynced from Edge cluster.
  - object selector: group, version, kind, name, namespace (for namespaced objects), label, annotation (, and more such as ownership reference?)
- Upsyncing CRD is out of scope for now. This means when upsyncing a CR, corresponding APIBinding (not CRD) is available on the mailbox workspace. This limitation might be revisited later. 
- ~Upsynced objects can be accessed from APIExport set on the workload management workspace bound to the mailbox workspace (with APIBinding). This access pattern might be changed when other APIs such as summarization are provided in KubeStellar.~ => Upsynced objects are accessed through Mailbox informer.

### Feasibility study
We will verify if the design described here could cover the following 4 scenarios. 
- I can register a KubeStellar-Syncer on a Edge cluster to connect a mailbox workspace specified by name. (KubeStellar-Syncer registration)
- I can deploy Kyverno and its policy from mailbox workspace to Edge cluster just by using manifests (generated from Kyverno helm chart) rather than using OLM. (workload deployment by KubeStellar-Syncer's downsyncing)
- I can see the policy report generated at Edge cluster via API Export on workload management workspace. (resource upsyncing by KubeStellar-Syncer) 
- I can deploy the denatured objects on mailbox workspace to Edge cluster by renaturing them automatically in KubeStellar-Syncer. (workload deployment by renaturing)

## Build KubeStellar-Syncer image

Prerequisite
- Install ko (https://ko.build/install/)

### How to build the image in your local
1. `make build-kubestellar-syncer-image-local`
e.g.
```
$ make build-kubestellar-syncer-image-local
2023/04/24 11:50:37 Using base distroless.dev/static:latest@sha256:81018475098138883b80dcc9c1242eb02b53465297724b18e88591a752d2a49c for github.com/{{ config.repo_short_name }}/cmd/syncer
2023/04/24 11:50:38 Building github.com/{{ config.repo_short_name }}/cmd/syncer for linux/arm64
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

1. `make build-kubestellar-syncer-image`

The behavior can be modified with some make variables; their default values are what get used in a normal build.  The variables are as follows.

- `DOCKER_REPO`, `IMAGE_TAG`: the built multi-platform manifest will
  be pushed to `$DOCKER_REPO:$IMAGE_TAG`.  The default for
  `DOCKER_REPO` is `quay.io/kubestellar/syncer`.  The default for
  `IMAGE_TAG` is the concatenation of: "git-", a short ID of the
  current git commit, "-", and either "clean" or "dirty" depending on
  what `git status` has to say about it. It is **STRONGLY** recommended to **NOT**
  override `IMAGE_TAG` _unless_ the override _also_ identifies the git commit
  and cleanliness, as this tag is the ONLY place that this is recorded (for the
  sake of reproducable builds, the `go build` command is told to _not_ include
  git and time metadata). To _add_ tags, use your container runtime's and/or
  image registry's additional functionality.
- `SYNCER_PLATFORMS`: a
  comma-separated list of `docker build` "platforms".  The default is
  "linux/amd64,linux/arm64,linux/s390x".
- `ADDITIONAL_ARGS`: a word that will be added into the `ko build` command line.
  The default is the empty string.

For example,
```
$ make build-kubestellar-syncer-image DOCKER_REPO=quay.io/mspreitz/syncer SYNCER_PLATFORMS=linux/amd64,linux/arm64 
2023/11/06 13:46:15 Using base cgr.dev/chainguard/static:latest@sha256:d3465871ccaba3d4aefe51d6bb2222195850f6734cbbb6ef0dd7a3da49826159 for github.com/kubestellar/kubestellar/cmd/syncer
2023/11/06 13:46:16 Building github.com/kubestellar/kubestellar/cmd/syncer for linux/amd64
2023/11/06 13:46:16 Building github.com/kubestellar/kubestellar/cmd/syncer for linux/arm64
2023/11/06 13:46:43 Publishing quay.io/mspreitz/syncer:git-a4250b7ee-dirty
2023/11/06 13:46:44 pushed blob: sha256:250c06f7c38e52dc77e5c7586c3e40280dc7ff9bb9007c396e06d96736cf8542
2023/11/06 13:46:44 pushed blob: sha256:24e67d450bd33966f28c92760ffcb5eae57e75f86ce1c0e0266a5d3c159d1798
2023/11/06 13:46:44 pushed blob: sha256:cd93f0a485889b13c1b34307d8dde4b989b45b7ebdd1f13a2084c89c87cb2fbf
2023/11/06 13:46:44 pushed blob: sha256:aa2769d82ae2f06035ceb26ce127c604bc0797f3e9a09bfc0dc010afff25d5c6
2023/11/06 13:46:44 pushed blob: sha256:9ac3b3732a57658f71e51d440eba76d27be0fac6db083c3e227585d5d7b0be94
2023/11/06 13:46:44 pushed blob: sha256:512f2474620de277e19ecc783e8e2399f54cb2669873db4b54159ac3c47a1914
2023/11/06 13:46:44 pushed blob: sha256:f2ae5118c0fadc41f16d463484970c698e9640de5d574b0fd29d4065e6d92795
2023/11/06 13:46:44 pushed blob: sha256:836fc9b0d92a362f818a04d483219a5254b4819044506b26d2a78c27a49d8421
2023/11/06 13:46:44 pushed blob: sha256:74256082c076ec34b147fa439ebdafffb10043cb418abe7531c49964cc2e9376
2023/11/06 13:46:44 quay.io/mspreitz/syncer:sha256-905d0fda05d7f9312c0af44856e9af5004ed6e2369f38b71469761cb3f9da2d1.sbom: digest: sha256:d40e5035236f888f8a1a784c4c630998dd92ee66c1b375bf379f1c915c4f296d size: 374
2023/11/06 13:46:44 Published SBOM quay.io/mspreitz/syncer:sha256-905d0fda05d7f9312c0af44856e9af5004ed6e2369f38b71469761cb3f9da2d1.sbom
2023/11/06 13:46:44 quay.io/mspreitz/syncer:sha256-18045f17222f9d0ec4fa3f736eaba891041d2980d1fb8c9f7f0a7a562172c9e5.sbom: digest: sha256:cad30541f2af79b74f87a37d284fa508fefd35cb130ee35745cfe31d85318fe9 size: 373
2023/11/06 13:46:44 Published SBOM quay.io/mspreitz/syncer:sha256-18045f17222f9d0ec4fa3f736eaba891041d2980d1fb8c9f7f0a7a562172c9e5.sbom
2023/11/06 13:46:44 quay.io/mspreitz/syncer:sha256-25d71940766653861e3175feec34fd2a6faff4f4c4f7bd55784f035a860d3be2.sbom: digest: sha256:877daabfb8593ce25c377446f9ec07782eb89b1ff15afdf9a2dfe882b7f87b06 size: 374
2023/11/06 13:46:44 Published SBOM quay.io/mspreitz/syncer:sha256-25d71940766653861e3175feec34fd2a6faff4f4c4f7bd55784f035a860d3be2.sbom
2023/11/06 13:46:45 pushed blob: sha256:0757eb0b6bd5eb800545762141ea55fae14a3f421aa84ac0414bbf51ffd95509
2023/11/06 13:46:45 pushed blob: sha256:9b50f69553a78acc0412f1fba1e27553f47a0f1cc76acafaad983320fb4d2edd
2023/11/06 13:46:54 pushed blob: sha256:15df213e4830817c1a38d97fda67c3e8459c17bc955dc36ac7f2fbdea26a12d4
2023/11/06 13:46:54 quay.io/mspreitz/syncer@sha256:905d0fda05d7f9312c0af44856e9af5004ed6e2369f38b71469761cb3f9da2d1: digest: sha256:905d0fda05d7f9312c0af44856e9af5004ed6e2369f38b71469761cb3f9da2d1 size: 1211
2023/11/06 13:46:54 pushed blob: sha256:6144db4c37348e2bdba9e850652e46f260dbab377e4f62d29bcdb84fcceaca00
2023/11/06 13:46:55 quay.io/mspreitz/syncer@sha256:25d71940766653861e3175feec34fd2a6faff4f4c4f7bd55784f035a860d3be2: digest: sha256:25d71940766653861e3175feec34fd2a6faff4f4c4f7bd55784f035a860d3be2 size: 1211
2023/11/06 13:46:55 quay.io/mspreitz/syncer:git-a4250b7ee-dirty: digest: sha256:18045f17222f9d0ec4fa3f736eaba891041d2980d1fb8c9f7f0a7a562172c9e5 size: 986
2023/11/06 13:46:55 Published quay.io/mspreitz/syncer:git-a4250b7ee-dirty@sha256:18045f17222f9d0ec4fa3f736eaba891041d2980d1fb8c9f7f0a7a562172c9e5
echo KO_DOCKER_REPO=quay.io/mspreitz/syncer GOFLAGS=-buildvcs=false ko build --platform=linux/amd64,linux/arm64 --bare --tags git-a4250b7ee-dirty  ./cmd/syncer
KO_DOCKER_REPO=quay.io/mspreitz/syncer GOFLAGS=-buildvcs=false ko build --platform=linux/amd64,linux/arm64 --bare --tags git-a4250b7ee-dirty ./cmd/syncer
quay.io/mspreitz/syncer:git-a4250b7ee-dirty@sha256:18045f17222f9d0ec4fa3f736eaba891041d2980d1fb8c9f7f0a7a562172c9e5
```

The last line of the output shows the full image reference, including both tag and digest of the "image" (technically it is a multi-platform manifest).

## Teardown the environment

{%
   include-markdown "../../common-subs/teardown-the-environment.md"
   start="<!--teardown-the-environment-start-->"
   end="<!--teardown-the-environment-end-->"
%}
