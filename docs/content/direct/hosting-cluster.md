# Using existing Kubernetes cluster to host KubeStellar

Status of this document: it is the barest of a start. Much more needs to be written.

## Using an existing Kind cluster as the hosting cluster

This requires a pre-existing Kind cluster that has an Ingress controller that is listening on host port 9443 and configured with TLS passthrough.

The examples say to create a Kind cluster for hosting using the following command.

```shell
kflex init --create-kind
```

To use a pre-existing Kind cluster instead, make sure that your current kubeconfig context is for accessing that cluster and issue the following command.

```shell
kflex init
```

All the subsequent kubectl and helm commands that say to use the kubeconfig context named `kind-kubeflex` need to be modified to use the appropriate kubeconfig context for accessing the hosting cluster.

## Using an existing OpenShift cluster as the hosting cluster

This is similar to using an existing Kind cluster but requires an additional modification. Modify the `kflex` init command and subsequent kubeconfig context references as in the existing-kind-cluster scenario.

Additionally, the recipe for registering a WEC with the ITS needs to be modified. In the `clusteradm` command, omit the `--force-internal-endpoint-lookup` flag. If following the example commands literally, this means to define `flags=""` rather than `flags="--force-internal-endpoint-lookup"`.
