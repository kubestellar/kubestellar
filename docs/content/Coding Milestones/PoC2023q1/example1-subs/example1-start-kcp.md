<!--example1-start-kcp-start-->
### Start kcp

Download and build or install [kcp](https://github.com/kcp-dev/kcp/releases/tag/v0.11.0),
according to your preference.

In some shell that will be used only for this purpose, issue the `kcp
start` command.  If you have junk from previous runs laying around,
you should probably `rm -rf .kcp` first.

In the shell commands in all the following steps it is assumed that
`kcp` is running and `$KUBECONFIG` is set to the
`.kcp/admin.kubeconfig` that `kcp` produces, except where explicitly
noted that the florin or guilder cluster is being accessed.

It is also assumed that you have the usual kcp kubectl plugins on your
`$PATH`.

clone the v0.11.0 branch kcp source:
```shell
git clone -b v0.11.0 https://github.com/kcp-dev/kcp kcp
```
build the kubectl-ws binary and include it in `$PATH`
```shell
pushd kcp
make build
export PATH=$(pwd)/bin:$PATH
```

Run the kcp server in a forked shell.  Even though the subcommand is "start", it does not just launch the server, it continues with running the server.
```shell
export KUBECONFIG=$(pwd)/.kcp/admin.kubeconfig
kcp start &> /tmp/kcp.log &
popd
sleep 30 
```
<!--example1-start-kcp-end-->
<!-- > /dev/null & -->
