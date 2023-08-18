<!--user-example1-install-all-start-->
### Install KubeStellar

```shell
HOMEBREW_NO_INSTALL_FROM_API=1 brew install -s --formula ~/projects/fork/kubestellar/brew/kcp.rb
HOMEBREW_NO_INSTALL_FROM_API=1 brew install -s --formula ~/projects/fork/kubestellar/brew/kcp_kubectl.rb
HOMEBREW_NO_INSTALL_FROM_API=1 brew install -s --formula ~/projects/fork/kubestellar/brew/kubestellar.rb
```

Test that your installation is working, so far

```
export KUBECONFIG=$(pwd)/.kcp/admin.kubeconfig
kubectl ws tree
```

 you should see
``` { .bash .no-copy }
 .
└── root
    └── compute
```
<!--user-example1-install-all-start-->

<!-- once merged we can use... -->
<!-- brew install kubestellar/kubestellar/brew/kcp.rb -->
<!-- brew install kubestellar/kubestellar/brew/kubestellar.rb -->
