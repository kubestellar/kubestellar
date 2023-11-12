<!--example1-kcp-provider-start-->

#### Create a space provider description for KCP

Space provider for KCP will allow to use KCP as backend provider for spaces.
Use the following commands to create a provider secret for KCP access and
a space provider definition.

```shell
KUBECONFIG=~/.kube/config kubectl --context space-mgt create secret generic kcpsec --from-file=kubeconfig=$KUBECONFIG
KUBECONFIG=~/.kube/config kubectl --context space-mgt apply -f - <<EOF
apiVersion: space.kubestellar.io/v1alpha1
kind: SpaceProviderDesc
metadata:
  name: default
spec:
  ProviderType: "kcp"
  SpacePrefixForDiscovery: "ks-"
  secretRef:
    namespace: default
    name: kcpsec
EOF
```

Next, use the following command to wait for the space-manger to process the provider.

```shell
KUBECONFIG=~/.kube/config kubectl --context space-mgt wait --for=jsonpath='{.status.Phase}'=Ready spaceproviderdesc/default --timeout=10s
```

<!--example1-kcp-provider-end-->