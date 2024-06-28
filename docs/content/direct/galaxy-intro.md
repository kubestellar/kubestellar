# galaxy

The KubeStellar **galaxy** is a secondary repository of _as-is_ KubeStellar-related tools and packages that are not part of the regular KubeStellar releases.
These integrations are beyond the scope of the core kubestellar repo, so are located here in a separate repository.
It's name is **galaxy** in line with our space theme and to indicate a broader constellation of projects for establishing integrations/collaborations.

It includes additional modules, tools and documentation to facilitate KubeStellar integration with other community projects, as well as some more experimental code we may be tinkering with for possible inclusion at some point.


Right now, **galaxy** includes some bash-based utility, and scripts to replicate demos and PoCs such as KFP + KubeStellar integration
and Argo Workflows + KubeStellar integration.

## Utility Scripts

- **suspend-webhook** - webhook used to suspend argo workflows (and in the future other types of workloads supporting the suspend flag)

- **shadow-pods** - controller used to support streaming logs in Argo Workflows and KFP.

- **clustermetrics** - a CRD and controller that provide basic cluster metrics info for each node in a cluster, designed to work together with KubeStellar sync/status sync mechanisms.

- **mc-scheduling** - A Multi-cluster scheduling framework supporting pluggable schedulers.


## KubeFlow Pipelines v2


## Argo Workflows 

# Learn More

To learn more [visit the repository at https://github.com/kubestellar/galaxy](https://github.com/kubestellar/gala)
**_Note that all the code in the galaxy repo is experimental and is available on an as-is basis **
