<!--teardown-the-environment-start-->
To remove the example usage, delete the IMW and WMW and kind clusters run the following commands:

``` {.bash}
rm florin-syncer.yaml guilder-syncer.yaml || true
kubectl ws root
kubectl delete workspace example-imw
kubectl ws root:my-org
kubectl kubestellar remove wmw example-wmw
kubectl ws root
kubectl delete workspace my-org
kind delete cluster --name florin
kind delete cluster --name guilder
```

Stop and uninstall KubeStellar use the following command:

``` {.bash}
kubestellar stop
```

Stop and uninstall KubeStellar and kcp with the following command:

``` {.bash}
remove-kubestellar
```
<!--teardown-the-environment-end-->