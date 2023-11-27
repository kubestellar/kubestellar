<!--teardown-the-environment-start-->
To remove the example usage, delete the IMW and WMW and kind clusters run the following commands:

``` {.bash}
rm florin-syncer.yaml guilder-syncer.yaml || true
kubectl ws root
kubectl delete workspace imw1
kubectl kubestellar remove wmw wmw1
kind delete cluster --name florin
kind delete cluster --name guilder
```

Teardown of KubeStellar depends on which style of deployment was used.

### Teardown bare processes

The following command will stop whatever KubeStellar controllers are running.

``` {.bash}
kubestellar stop
```

Stop and uninstall KubeStellar and kcp with the following command:

``` {.bash}
remove-kubestellar
```

### Teardown Kubernetes workload

With `kubectl` configured to manipulate the hosting cluster, the following command will remove the workload that is kcp and KubeStellar.

``` {.bash}
helm delete kubestellar
```

<!--teardown-the-environment-end-->
