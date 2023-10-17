## Instructions for installing the KubeStellar user commands and kubectl plugins using Brew package manager

install brew:
    https://brew.sh

to install:
    HOMEBREW_NO_INSTALL_FROM_API=1 brew install -s --formula kcp_cli.rb
    HOMEBREW_NO_INSTALL_FROM_API=1 brew install -s --formula kubestellar_cli.rb

to remove: 
    brew remove kubestellar_cli
    brew remove kcp_cli


after PR is accepted:

brew tap kubestellar/kubestellar https://github.com/kubestellar/kubestellar/brew
brew install kcp_cli
brew install kubestellar_cli
