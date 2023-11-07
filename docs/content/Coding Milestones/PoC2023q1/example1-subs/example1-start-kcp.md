<!--example1-start-kcp-start-->
### Deploy kcp and KubeStellar

You need kcp and KubeStellar and can deploy them in either of two
ways: as bare processes on whatever host you are using to run this
example, or as workload in a Kubernetes cluster (an OpenShift cluster
qualifies).  Do one or the other, not both.

KubeStellar only works with release `v0.11.0` of kcp. To downsync ServiceAccount objects you will need a patched version of that in order to get the denaturing of them as discussed [in the design outline](../../outline/#needs-to-be-denatured-in-center-natured-in-edge).

### Deploy kcp and KubeStellar as bare processes

#### Start kcp

The following commands fetch the appropriate kcp server and plugins for your OS and ISA and download them and put them on your `$PATH`.

```shell
rm -rf kcp
mkdir kcp
pushd kcp
(
  set -x
  case "$OSTYPE" in
      linux*)   os_type="linux" ;;
      darwin*)  os_type="darwin" ;;
      *)        echo "Unsupported operating system type: $OSTYPE" >&2
                false ;;
  esac
  case "$HOSTTYPE" in
      x86_64*)  arch_type="amd64" ;;
      aarch64*) arch_type="arm64" ;;
      arm64*)   arch_type="arm64" ;;
      *)        echo "Unsupported architecture type: $HOSTTYPE" >&2
                false ;;
  esac
  kcp_version=v0.11.0
  trap "rm kcp.tar.gz kcp-plugins.tar.gz" EXIT
  curl -SL -o kcp.tar.gz "https://github.com/kubestellar/kubestellar/releases/download/v0.12.0/kcp_0.11.0_${os_type}_${arch_type}.tar.gz"
  curl -SL -o kcp-plugins.tar.gz "https://github.com/kcp-dev/kcp/releases/download/${kcp_version}/kubectl-kcp-plugin_${kcp_version//v}_${os_type}_${arch_type}.tar.gz"
  tar -xzf kcp-plugins.tar.gz
  tar -xzf kcp.tar.gz
)
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

<!--example1-start-kcp-end-->
<!-- > /dev/null & -->
