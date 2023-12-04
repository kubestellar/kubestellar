<!--where-resolver-0-pull-kubestellar-source-and-start-start-->
Clone the v0.11.0 branch kcp source:
```shell
git clone -b v0.11.0 https://github.com/kcp-dev/kcp kcp
```
Build the kubectl-ws binary and include it in `$PATH`
```shell
pushd kcp
make build
export PATH=$(pwd)/bin:$PATH
```

Run kcp (this will spit out tons of information and stay running in this terminal window).
Set your `KUBECONFIG` environment variable to name the kubernetes client config file that is generated.
```shell
kcp start &> /dev/null &
export KUBECONFIG=$(pwd)/.kcp/admin.kubeconfig
popd
sleep 30
```
<!--where-resolver-0-pull-kubestellar-source-and-start-end-->
