<!--build-core-image-start-->
Build the core images - both kubestellar and space-framework.

```shell
pwd
export EXTRA_CORE_TAG=$(date +test%m%d-%H%M%S)
make kubestellar-image-local
cd space-framework
make spacecore-image-local
cd ..
```
<!--build-core-image-end-->
