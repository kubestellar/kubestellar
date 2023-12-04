<!--example1-cp-provider-start-->

#### Create a space provider description for KCP

Space provider for KCP will allow you to use it as a backend provider for spaces.
Use the following commands to create a provider secret for access and
a space provider definition.

```shell
KUBECONFIG=$SM_CONFIG kubectl --context sm-mgt create secret generic kcpsec --from-file=kubeconfig=$KUBECONFIG --from-file=incluster=$KUBECONFIG
KUBECONFIG=$SM_CONFIG kubectl --context sm-mgt apply -f - <<EOF
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
KUBECONFIG=$SM_CONFIG kubectl --context sm-mgt wait --for=jsonpath='{.status.Phase}'=Ready spaceproviderdesc/default --timeout=90s
```

<!--example1-cp-provider-end-->
