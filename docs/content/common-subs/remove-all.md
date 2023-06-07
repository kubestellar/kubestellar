```shell
ps -ef | grep mailbox-controller | grep -v grep | awk '{print $2}' | xargs kill || true
ps -ef | grep kubestellar-scheduler | grep -v grep | awk '{print $2}' | xargs kill || true
ps -ef | grep placement-translator | grep -v grep | awk '{print $2}' | xargs kill || true
ps -ef | grep kcp | grep -v grep | awk '{print $2}' | xargs kill || true
kind delete cluster --name florin
kind delete cluster --name guilder
```