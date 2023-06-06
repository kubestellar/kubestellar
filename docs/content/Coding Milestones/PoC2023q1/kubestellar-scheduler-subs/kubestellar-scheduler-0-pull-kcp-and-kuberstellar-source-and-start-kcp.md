<!--kubestellar-scheduler-0-pull-kcp-and-kuberstellar-source-and-start-kcp-start-->
```shell
git clone {{ config.repo_url }} KubeStellar
```

clone the v0.11.0 branch kcp source:
```shell
git clone -b v0.11.0 https://github.com/kcp-dev/kcp kcp
```
build the kubectl-ws binary and include it in `$PATH`
```shell
cd kcp
make build
export PATH=$(pwd)/bin:$PATH
```

run kcp (kcp will spit out tons of information and stay running in this terminal window)
```shell
kcp start &> /dev/null &
sleep 30
```
<!--kubestellar-scheduler-0-pull-kcp-and-kuberstellar-source-and-start-kcp-end-->