# This work in progress document explains how to deploy KubeStellar Core to a MicroShift on a Raspebrry PI

This documents explains how to use KubeStellar Core chart in MicroShift.

The information provided is specific for the following release:

```shell
export KUBESTELLAR_VERSION=0.23.0
```

## Pre-requisites

In this exxample we assume to have two Raspberry Pi running a MicroShift cluster. Detailed informations on how to setup MicroShift on the Raspberry Pi can be found [here](https://community.ibm.com/community/user/cloud/blogs/alexei-karve/2021/11/28/microshift-4). The first Raspberry Pi (_e.g._, core.local) will be used to deploy the KubeStellar Core, while the second Raspberry Pi (_e.g._, wec.local) will be used as a Workload Execution Cluster.

The WEC Raspberry Pi must be able to reach the Core Raspberry Pi, _e.g._, `ping core.local` should succeed.

The Core Raspberry Pi must open port `6443` on its firewall.

The Core Rasperry Pi will need `kubectl`, `helm`, and `clusteradm`.

The WEC Rasperry Pi will need `kubectl` and `clusteradm`.

## Setting up KubeStellar Core

On the Core Raspberry Pi install KubeStellar chart with the command:

```shell
helm upgrade --install ks-core oci://ghcr.io/kubestellar/kubestellar/core-chart --version $KUBESTELLAR_VERSION \
  --set "kubeflex-operator.isOpenShift=true" \
  --set-json='ITSes=[{"name":"its0","type":"host"}]' \
  --set-json='WDSes=[{"name":"wds0","type":"host"}]'
```

Note that:
-  the ITS `its0` must be set to type `host` so that it can be reached from other machines on the network using the doman or IP address of the Core Raspberry Pi
-  the WDS `wds0` must be set to type `host` becase of an issue in KubeFlex that has not been resolved yet

## Joining the WEC raspberry Pi

After all the pods are ready retrieve the cluster adm join command using:

```shell
clusteradm get token
```

Which should return something like:

```shell
clusteradm join --core-token ... --hub-apiserver https://127.0.0.1:6443 --cluster-name <cluster_name>
```

On the WEC Raspberry Pi, execute the above command after replacing:
- the ip address `127.0.0.1` with the ip address or domanin of the Core Raspberry Pi, _e.g._, `core.local`
- the `<cluster_name>` with the name to be assigned to the WEC Raspberry Pi, _e.g._, `wec`

In summary, on the WEC Raspberry Pi, the join command may look like:

```shell
clusteradm join --core-token ... --hub-apiserver https://core.local:6443 --cluster-name wec
```

Assuming that the command succeeds, then on the Core Raspberry Pi we should see a pending CSR:

```shell
$ kubectl get csr
NAME        AGE   SIGNERNAME                            REQUESTOR                                                         CONDITION
wec-qh8zq   13s   kubernetes.io/kube-apiserver-client   system:serviceaccount:open-cluster-management:cluster-bootstrap   Pending
```

Then, we can join the WEC Raspberry Pi by using the command:

```shell
clusteradm accept --clusters wec
```

Assuming that the command succeeds, then on the Core Raspberry Pi we should see something like:

```shell
$ kubectl get csr
NAME                           AGE    SIGNERNAME                            REQUESTOR                                                         CONDITION
addon-wec-addon-status-kvr6j   3s     kubernetes.io/kube-apiserver-client   system:open-cluster-management:wec:8wjxl                          Approved,Issued
wec-qh8zq                      2m6s   kubernetes.io/kube-apiserver-client   system:serviceaccount:open-cluster-management:cluster-bootstrap   Approved,Issued
```

## Next steps

See the [examples](./examples.md).
