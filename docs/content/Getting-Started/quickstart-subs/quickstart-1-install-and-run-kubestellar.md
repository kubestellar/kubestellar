<!--quickstart-1-install-and-run-kubestellar-start-->

KubeStellar works in the context of kcp, so to use KubeStellar you also need kcp. Download the kcp and **KubeStellar** binaries and scripts into a `kubestellar` subfolder in your current working directory using the following command:

```shell
bash <(curl -s {{ config.repo_raw_url }}/{{ config.ks_branch }}/bootstrap/bootstrap-kubestellar.sh) --kubestellar-version {{ config.ks_tag }}
export PATH="$PATH:$(pwd)/kcp/bin:$(pwd)/kubestellar/bin"
export KUBECONFIG="$(pwd)/.kcp/admin.kubeconfig"
```

Check that `KubeStellar` is running:

First, check that controllers are running with the following command:

```shell
ps aux | grep -e mailbox-controller -e placement-translator -e kubestellar-scheduler
```

which should yield something like:

``` { .sh .no-copy }
user     1892  0.0  0.3 747644 29628 pts/1    Sl   10:51   0:00 mailbox-controller -v=2
user     1902  0.3  0.3 743652 27504 pts/1    Sl   10:51   0:02 kubestellar-scheduler -v 2
user     1912  0.3  0.5 760428 41660 pts/1    Sl   10:51   0:02 placement-translator -v=2
``` 

Second, check that the Edge Service Provider Workspace (`espw`) is created with the following command:

```shell
kubectl ws tree
```

which should yield:

``` { .sh .no-copy }
kubectl ws tree
.
└── root
    ├── compute
    └── espw
```
<!--quickstart-1-install-and-run-kubestellar-end-->
