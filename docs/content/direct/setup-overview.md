# Setting up KubeStellar

"Setup" is a porous grouping of some of the steps in [the full outline](user-guide-intro.md#the-full-story), and comprises the following. Also, bear in mind the [Setup limitations](setup-limitations.md).

- Install software prerequisites. See [prerequisites](pre-reqs.md).
- KubeFlex Hosting cluster
  - Acquire the ability to use a Kubernetes cluster to serve as the [KubeFlex](https://github.com/kubestellar/kubeflex/) hosting cluster. See [Acquire cluster for KubeFlex hosting](acquire-hosting-cluster.md).
  - [Initialize that cluster as a KubeFlex hosting cluster](init-hosting-cluster.md).
- Core Spaces
  - Create an [Inventory and Transport Space](its.md) (ITS).
  - Create a [Workload Description Space](wds.md) (WDS).
- [Core Helm Chart](core-chart.md) (covering three of the above topics).
- Workload Execution Clusters
  - Create a [Workload Execution Cluster](wec.md) (WEC).
  - [Register the WEC in the ITS](wec-registration.md).
