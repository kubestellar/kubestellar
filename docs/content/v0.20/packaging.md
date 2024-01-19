# Packaging and Delivery

## status-addon

The [status-addon](https://github.ibm.com/dettori/status-addon) repo is the source of a RedHat-style operator. 

The operator is delivered by a Helm chart at `quay.io/pdettori/status-addon-chart`. Chart versioning? How is the chart built, pushed?

There are probably one or more container images involved.

## KubeStellar

### Central container image

Appears to be at `ghcr.io/kubestellar/kubestellar/kubestellar-operator:0.20.0-alpha.1`. `make ko-build-push` will create that image. `make ko-build-local` will make a local image for just the local platform.

### KubeStellar Operator Helm Chart

The source for a Helm chart is in [core-helm-chart](../../../core-helm-chart). `make chart` (re)derives it from loal sources.

This chart creates (among other things) a `Deployment` object that runs a container from the image `ghcr.io/kubestellar/kubestellar/kubestellar-operator:0.20.0-alpha.1`.

### KubeFlex PostCreateHooks

There are two `PostCreateHook` objects defined in [config/postcreate-hooks](../../../config/postcreate-hooks).

- `ocm.yaml` adds `clusteradm` by running a container using the image `quay.io/kubestellar/clusteradm:0.7.1` which is built from ... what?
- `kubestellar.yaml` runs container image `quay.io/kubestellar/helm:v3.13.2` (which is built by what? From what?) to instantiate the chart from `oci://ghcr.io/kubestellar/kubestellar/kubestellar-operator-chart` (tag?), which is built by (what?) from (what? probably answered above).
