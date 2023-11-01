<!--build-core-image-start-->
Build the core image.

```shell
pwd
export EXTRA_CORE_TAG=$(date +test%m%d-%H%M%S)
make kubestellar-image-local
```
<!--build-core-image-end-->
