<!--quickstart-1-install-and-run-kubestellar-start-->

KubeStellar works in the context of kcp, so to use KubeStellar you also need kcp. The following commands will download the kcp and **KubeStellar** executables into subdirectories of your current working directory, deploy (i.e., start and configure) kcp and KubeStellar, and configure your shell to use kcp and KubeStellar.  If you want to suppress the deployment part then add `--deploy false` to the first command's flags (e.g., after the specification of the KubeStellar version); for the deployment-only part, once the executable have been fecthed, see [kcp control](../../../Coding Milestones/PoC2023q1/commands/#kcp-control) and [KubeStellar platform control](../../../Coding Milestones/PoC2023q1/commands/#kubestellar-platform-control).

```shell
bash <(curl -s {{ config.repo_raw_url }}/{{ config.ks_branch }}/bootstrap/bootstrap-kubestellar.sh) --kubestellar-version {{ config.ks_tag }}
export PATH="$PATH:$(pwd)/kcp/bin:$(pwd)/kubestellar/bin"
export KUBECONFIG="$(pwd)/.kcp/admin.kubeconfig"
```

Check that `KubeStellar` is running:

First, check that controllers are running with the following command:

```shell
ps aux | grep -e mailbox-controller -e placement-translator -e kubestellar-where-resolver
```

which should yield something like:

``` { .sh .no-copy }
user     1892  0.0  0.3 747644 29628 pts/1    Sl   10:51   0:00 mailbox-controller -v=2
user     1902  0.3  0.3 743652 27504 pts/1    Sl   10:51   0:02 kubestellar-where-resolver -v 2
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
