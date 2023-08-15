<!--quickstart-1-install-and-run-kubestellar-start-->

KubeStellar works in the context of kcp, so to use KubeStellar you also need kcp.

We support two ways to deploy kcp and KubeStellar. The older way is to run them as bare processes. The newer way is to deploy them as workload in a Kubernetes (possibly OpenShift) cluster.

### Deploy kcp and KubeStellar as bare processes

The following commands will download the kcp and **KubeStellar** executables into subdirectories of your current working directory, deploy (i.e., start and configure) kcp and KubeStellar as bare processes, and configure your shell to use kcp and KubeStellar.  If you want to suppress the deployment part then add `--deploy false` to the first command's flags (e.g., after the specification of the KubeStellar version); for the deployment-only part, once the executable have been fetched, see the documentation about [the commands for bare process deployment](../../../Coding Milestones/PoC2023q1/commands/#bare-process-deployment).

```shell
bash <(curl -s {{ config.repo_raw_url }}/{{ config.ks_branch }}/bootstrap/bootstrap-kubestellar.sh) --kubestellar-version {{ config.ks_tag }}
export PATH="$PATH:$(pwd)/kcp/bin:$(pwd)/kubestellar/bin"
export KUBECONFIG="$(pwd)/.kcp/admin.kubeconfig"
```

Check that `KubeStellar` is running.

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

Second, check that TMC compute service provider workspace and the KubeStellar Edge Service Provider Workspace (`espw`) have been created with the following command:

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

### Deploy kcp and KubeStellar as Kubernetes workload

This requires a KubeStellar release GREATER THAN v0.5.0.

This example uses a total of three `kind` clusters, which tends to run
into [a known issue with a known
work-around](https://kind.sigs.k8s.io/docs/user/known-issues/#pod-errors-due-to-too-many-open-files),
so take care of that.

Before you can deploy kcp and KubeStellar as workload in a Kubernetes
cluster, you need a Kubernetes cluster and it needs to have an Ingress
controller installed.  We use the term "hosting cluster" for the
cluster that plays this role.  In this quickstart, we make such a
cluster with [kind](https://kind.sigs.k8s.io/).  Follow the [developer
directions for making a hosting cluster with kind](../../../Coding%20Milestones/PoC2023q1/environments/dev-env/#hosting-kubestellar-in-a-kind-cluster);
you need not worry about loading a locally built container image into
that cluster.

This example uses the domain name "hostname.favorite.my" for the machine where you invoked `kind create cluster`. If you have not already done so then issue the following command, replacing `a_good_IP_address_for_this_machine` with an IPv4 address for your machine that can be reached from inside a container or VM (i.e., not 127.0.0.1).

``` {.bash}
sudo sh -c "echo a_good_IP_address_for_this_machine hostname.favorite.my >> /etc/hosts"
```

The next command relies on `kubectl` already being configured to manipulate the hosting cluster, which is indeed the state that `kind create cluster` leaves it in.

The following commands will (a) download the kcp and **KubeStellar** executables into subdirectories of your current working directory and (b) deploy (i.e., start and configure) kcp and KubeStellar as workload in the hosting cluster. If you want to suppress the deployment part then add `--deploy false` to the first command's flags (e.g., after the specification of the KubeStellar version); for the deployment-only part, once the executable have been fecthed, see the documentation for [the commands about deployment into a Kubernetes cluster](../../../Coding Milestones/PoC2023q1/commands/#deployment-into-a-kubernetes-cluster).

``` {.bash}
bash <(curl -s {{ config.repo_raw_url }}/{{ config.ks_branch }}/bootstrap/bootstrap-kubestellar.sh) --kubestellar-version {{ config.ks_tag }} --external-endpoint hostname.favorite.my:6443
export PATH="$PATH:$(pwd)/kcp/bin:$(pwd)/kubestellar/bin"
```

Using your original `kubectl` configuration that manipulates the hosting cluster, check that the KubeStellar Deployment has its intended one running Pod.

``` {.bash}
kubectl get deployments -n kubestellar
```

which should yield something like:

``` { .sh .no-copy }
NAME                 READY   UP-TO-DATE   AVAILABLE   AGE
kubestellar-server   1/1     1            1           2m42s
``` 

It may take some time for that Pod to reach Running state.

The bootstrap command above will print out instructions to set your KUBECONFIG environment variable to the pathname of a kubeconfig file that you can use as a user of kcp and KubeStellar.  Do that now, for the benefit of the remaining commands in this example.  It will look something like the following command.

``` {.bash}
export KUBECONFIG="$(pwd)/kubestellar.kubeconfig"
```

Check that the TMC compute service provider workspace and the KubeStellar Edge Service Provider Workspace (`espw`) have been created with the following command:

``` {.bash}
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
