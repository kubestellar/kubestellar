install brew:
    https://brew.sh

to install:
    HOMEBREW_NO_INSTALL_FROM_API=1 brew install -s --formula ~/projects/fork/kubestellar/brew/kubestellar_provider_kcp_kubectl.rb
    HOMEBREW_NO_INSTALL_FROM_API=1 brew install -s --formula ~/projects/fork/kubestellar/brew/kubestellar_provider_kcp.rb
    HOMEBREW_NO_INSTALL_FROM_API=1 brew install -s --formula ~/projects/fork/kubestellar/brew/kubestellar.rb

to remove: 
    brew remove kubestellar
    brew remove kubestellar_provider_kcp
    brew remove kubestellar_provider_kcp_kubectl
