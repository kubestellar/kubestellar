# Confusion due to hidden state in your kubeconfig

The `kflex` command maintains and works with a bit of state hidden in your kubeconfig file. This is where KubeFlex stashes the name of the kubeconfig context to use for accessing the KubeFlex hosting cluster. Following is an example of examining that state.

```console
mspreitz@mjs13 kubestellar % yq .preferences ${KUBECONFIG:-$HOME/.kube/config}
extensions:
  - extension:
      data:
        kflex-initial-ctx-name: kscore-stage
      metadata:
        creationTimestamp: null
        name: kflex-config-extension-name
    name: kflex-config-extension-name
```

The `kflex ctx` commands are normally hesitant to replace a bad value in there. Later releases of `kflex` are better than older ones, and the latest releases have ways on the command line to explicitly remove this hesitancy; the KubeStellar instructions and scripts use those.

Although it should no longer be necessary to use this, the following command shows a way to remove that bit of hidden state; after this, a `kflex ctx` command will succeed _if_ your current kubeconfig context is the one to use for accessing the KubeFlex hosting cluster.

```shell
yq -i 'del(.preferences)' ${KUBECONFIG:-$HOME/.kube/config}
```
