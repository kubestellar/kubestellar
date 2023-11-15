<!--pre-position-core-image-start-->
Install the core container image in the ks-core `kind` cluster.

```shell
kind load docker-image quay.io/kubestellar/kubestellar:$EXTRA_CORE_TAG --name ks-core
kind load docker-image quay.io/kubestellar/space-framework:$EXTRA_CORE_TAG --name ks-core
```
<!--pre-position-core-image-end-->
