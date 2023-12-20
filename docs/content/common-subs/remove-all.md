```shell
ps -ef | grep mailbox-controller | grep -v grep | awk '{print $2}' | xargs kill >/dev/null 2>&1 || true
ps -ef | grep kubestellar-where-resolver | grep -v grep | awk '{print $2}' | xargs kill >/dev/null 2>&1 || true
ps -ef | grep placement-translator | grep -v grep | awk '{print $2}' | xargs kill >/dev/null 2>&1 || true
ps -ef | grep space-manager | grep -v grep | awk '{print $2}' | xargs kill >/dev/null 2>&1 || true
ps -ef | grep example-backend | grep -v grep | awk '{print $2}' | xargs kill >/dev/null 2>&1 || true
ps -ef | grep kcp | grep -v grep | awk '{print $2}' | xargs kill >/dev/null 2>&1 || true
ps -ef | grep 'exe/main -v 2' | grep -v grep | awk '{print $2}' | xargs kill >/dev/null 2>&1 || true
kind delete cluster --name florin 2>&1
kind delete cluster --name guilder 2>&1
kind delete cluster --name sm-mgt 2>&1
kubectl config delete-context sm-mgt 2>&1
rm florin-syncer.yaml florin-config.yaml guilder-syncer.yaml guilder-config.yaml || true
rm -rf dex/ kcp/ kube-bind/ my-kubeconfigs/ konnector-imw1* konnector-wmw* || true
```
