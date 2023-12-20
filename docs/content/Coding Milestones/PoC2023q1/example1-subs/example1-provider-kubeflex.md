<!--example1-provider-kubeflex-start-->
### Deploy KubeFlex and KubeStellar as bare processes

#### Initialize KubeFlex

```shell
wget https://github.com/kubestellar/kubeflex/releases/download/v0.3.3/kubeflex_0.3.3_linux_amd64.tar.gz
mkdir kubeflex
tar xf kubeflex_0.3.3_linux_amd64.tar.gz -C kubeflex
kubeflex/bin/kflex --kubeconfig $SM_CONFIG init
rm kubeflex_0.3.3_linux_amd64.tar.gz
```

#### Create a space provider description for KubeFlex

Space provider for KubeFlex will allow you to use KubeFlex as backend provider for spaces.
Use the following commands to create a provider secret for KubeFlex access and
a space provider definition.

```shell
KUBECONFIG=$SM_CONFIG kubectl --context sm-mgt create secret generic kfsec --from-file=kubeconfig=$SM_CONFIG --from-file=incluster=$SM_CONFIG
KUBECONFIG=$SM_CONFIG kubectl --context sm-mgt apply -f - <<EOF
apiVersion: space.kubestellar.io/v1alpha1
kind: SpaceProviderDesc
metadata:
  name: default
spec:
  ProviderType: "kubeflex"
  SpacePrefixForDiscovery: "ks-"
  secretRef:
    namespace: default
    name: kfsec
EOF
```

Next, use the following command to wait for the space-manger to process the provider.

```shell
KUBECONFIG=$SM_CONFIG kubectl --context sm-mgt wait --for=jsonpath='{.status.Phase}'=Ready spaceproviderdesc/default --timeout=90s
```

The following variable will be used in later commands to indicate that
they are being invoked close enough to the provider's apiserver to
use the more efficient networking (see [doc on
"in-cluster"](../../commands/#in-cluster)).

```shell
in_cluster="--in-cluster"
kube_needed=false
```
<!--example1-provider-kubeflex-end-->
