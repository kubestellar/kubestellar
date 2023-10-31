<!--brew-install-start-->
``` {.bash .hide-me}
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
(echo; echo 'eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"') >> /home/runner/.bashrc
eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"
more /etc/hosts
# sudo echo $(curl https://api.ipify.org) kubestellar.core | sudo tee -a /etc/host
```
```shell
brew tap kubestellar/kubestellar
brew update
brew install kcp_cli
brew install kubestellar_cli
```
<!--brew-install-end-->
