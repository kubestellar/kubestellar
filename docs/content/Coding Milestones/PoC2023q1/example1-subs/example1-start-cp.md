<!--example1-start-cp-start-->
### Deploy kcp and KubeStellar

You need kcp and KubeStellar and can deploy them in either of two
ways: as bare processes on whatever host you are using to run this
example, or as workload in a Kubernetes cluster (an OpenShift cluster
qualifies).  Do one or the other, not both.

KubeStellar only works with release `v0.11.0` of kcp.

### Deploy kcp and KubeStellar as bare processes

#### Start kcp

Download and build or install [kcp](https://github.com/kcp-dev/kcp/releases/tag/v0.11.0),
according to your preference.  See the start of [the kcp quickstart](https://docs.kcp.io/kcp/v0.11/#quickstart) for instructions on that, but get [release v0.11.0](https://github.com/kcp-dev/kcp/releases/tag/v0.11.0) rather than the latest (the downloadable assets appear after the long list of changes/features/etc).

Clone the v0.11.0 branch of the kcp source:
```shell
git clone -b v0.11.0 https://github.com/kcp-dev/kcp kcp
```
and build the kubectl-ws binary and include it in `$PATH`
```shell
pushd kcp
make build
export PATH=$(pwd)/bin:$PATH
```

Running the kcp server creates a hidden subdirectory named `.kcp` to
hold all sorts of state related to the server. If you have run it
before and want to start over from scratch then you should `rm -rf
.kcp` first.

Use the following commands to: (a) run the kcp server in a forked
command, (b) update your `KUBECONFIG` environment variable to
configure `kubectl` to use the kubeconfig produced by the kcp server,
and (c) wait for the kcp server to get through some
initialization. The choice of `-v=3` for the kcp server makes it log a
line for every HTTP request (among other things).

```shell
kcp start -v=3 &> /tmp/kcp.log &
export KUBECONFIG=$(pwd)/.kcp/admin.kubeconfig
popd
# wait until KCP is ready checking availability of ws resource
while ! kubectl ws tree &> /dev/null; do
  sleep 10
done
```

Note that you now care about two different kubeconfig files: the one
that you were using earlier, which holds the contexts for your `kind`
clusters, and the one that the kcp server creates. The remainder of
this document assumes that your `kind` cluster contexts are in
`~/.kube/config`.

<!--example1-start-cp-end-->
<!-- > /dev/null & -->
