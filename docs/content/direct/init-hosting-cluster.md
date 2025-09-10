# Initializing the KubeFlex hosting cluster

The [KubeFlex](https://github.com/kubestellar/kubeflex) implementation
has to be installed in the cluster chosen to play the role of KubeFlex
hosting cluster. This can be done in any of the following ways.

## Bundled with cluster creation

As mentioned earlier, there are a couple of ways to both create the hosting cluster and initialize it for KubeFlex in one operation.

- [Using kflex init --create-kind](acquire-hosting-cluster.md#create-and-init-a-kind-cluster-as-hosting-cluster-with-kflex).
- [curl-to-bash script](acquire-hosting-cluster.md#create-and-init-a-kind-cluster-as-hosting-cluster-with-curl-to-bash-script).

## kflex init

The following command will install the KubeFlex implementation in the cluster that `kubectl` is configured to access, if you have sufficient privileges.

```shell
kflex init
```

### Using an existing OpenShift cluster as the hosting cluster

When the hosting cluster is an OpenShift cluster, the recipe for registering a WEC with the ITS ([to be written](wec.md)) needs to be modified. In the `clusteradm` command, omit the `--force-internal-endpoint-lookup` flag. If following [Getting Started](get-started.md#create-and-register-two-workload-execution-clusters) literally, this means to define `flags=""` rather than `flags="--force-internal-endpoint-lookup"`.

## KubeStellar core Helm chart

The [KubeStellar core Helm chart](core-chart.md) will install the KubeFlex implementation in the cluster that `kubectl` is configured to access, as well as create ITSes and WDSes.
