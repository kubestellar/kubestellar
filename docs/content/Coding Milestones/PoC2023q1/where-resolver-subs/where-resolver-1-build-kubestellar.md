<!--where-resolver-1-build-kubestellar-start-->
```shell
make imports
make codegen
make crds
make update-contextual-logging
make build
export PATH=$(pwd)/bin:$PATH
```
<!--where-resolver-1-build-kubestellar-end-->
