<!--required-packages-start-->
!!! tip "Required Packages for running and using KubeStellar:"
    === "General"
        You will need the following tools to deploy and use KubeStellar. 
        Select the tab for your environment for suggested commands to install them

        + __curl__ (omitted from most OS-specific instructions)

        + [__jq__](https://stedolan.github.io/jq/download/)

        + [__yq__](https://github.com/mikefarah/yq#install)

        + [__kubectl__](https://kubernetes.io/docs/tasks/tools/) (version range expected: 1.23-1.25)

    === "Mac"
        ``` title="jq - https://stedolan.github.io/jq/download/"
        brew install jq
        ```
        ``` title="yq - https://github.com/mikefarah/yq#install"
        brew install yq
        ```
        ``` title="kubectl - https://kubernetes.io/docs/tasks/tools/ (version range expected: 1.23-1.25)"
        brew install kubectl
        ```
    === "Ubuntu"
        ``` title="jq - https://stedolan.github.io/jq/download/"
        sudo apt-get install jq
        ```
        ``` title="yq - https://github.com/mikefarah/yq#install"
        sudo snap install yq
        ```
        ``` title="kubectl - https://kubernetes.io/docs/tasks/tools/ (version range expected: 1.23-1.25)"
        curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/$(dpkg --print-architecture)/kubectl" && chmod +x kubectl && sudo mv ./kubectl /usr/local/bin/kubectl
        ```
    === "Debian"
        ``` title="jq - https://stedolan.github.io/jq/download/"
        sudo apt-get install jq
        ```
        ``` title="yq - https://github.com/mikefarah/yq#install"
        sudo apt-get install yq
        ```
        ``` title="kubectl - https://kubernetes.io/docs/tasks/tools/ (version range expected: 1.23-1.25)"
        curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/$(dpkg --print-architecture)/kubectl" && chmod +x kubectl && sudo mv ./kubectl /usr/local/bin/kubectl
        ```
    === "Fedora/RHEL/CentOS"
        ``` title="jq - https://stedolan.github.io/jq/download/"
        yum -y install jq
        ```
        ``` title="yq - https://github.com/mikefarah/yq#install"
        # easiest to install with snap
        snap install yq
        ```
        ``` title="kubectl - https://kubernetes.io/docs/tasks/tools/ (version range expected: 1.23-1.25)"
        # For AMD64 / x86_64
        [ $(uname -m) = x86_64 ] && curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && chmod +x kubectl && mv ./kubectl /usr/local/bin/kubectl
        # for ARM64 / aarch64
        [ $(uname -m) = aarch64 ] && curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/arm64/kubectl" && chmod +x kubectl && mv ./kubectl /usr/local/bin/kubectl
        ```
    === "Windows"
        ``` title="Chocolatey - https://chocolatey.org/install#individual"
        Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))
        ```
        ``` title="curl"
        choco install curl -y
        ```
        ``` title="jq - https://stedolan.github.io/jq/download/"
        choco install jq -y
        ```
        ``` title="yq - https://github.com/mikefarah/yq#install"
        choco install yq -y
        ```
        ``` title="kubectl - https://kubernetes.io/docs/tasks/tools/install-kubectl-windows/ (version range expected: 1.23-1.25)"
        curl.exe -LO "https://dl.k8s.io/release/v1.27.2/bin/windows/amd64/kubectl.exe"    
        ```  
    === "WSL with Ubuntu"  
        ### How to install pre-requisites for a Windows Subsystem for Linux (WSL) envronment using an Ubuntu 22.04.01 distribution

        (Tested on a Intel(R) Core(TM) i7-9850H CPU @ 2.60GHz 2.59 GHz with 32GB RAM, a 64-bit operating system, x64-based processor
        Using Windows 11 Enterprise)

        ###### 1. If you're using a VPN, turn it off

        ###### 2. Install Ubuntu into WSL

        ###### 2.0 If wsl is not yet installed, open a powershell administrator window and run the following
        ```
        wsl --install
        ```
        ###### 2.1 reboot your system

        ###### 2.2 In a Windows command terminal run the following to list all the linux distributions that are available online
        ```
        wsl -l -o
        ```
        ###### 2.3 Select a linux distribution and install it into WSL
        ```
        wsl --install -d Ubuntu 22.04.01
        ```
        You will see something like:
        ```
        Installing, this may take a few minutes...
        Please create a default UNIX user account. The username does not need to match your Windows username.
        For more information visit: https://aka.ms/wslusers
        Enter new UNIX username:
        ```

        ###### 2.4 Enter your new username and password at the prompts, and you will eventually see something like:
        ```
        Welcome to Ubuntu 22.04.1 LTS (GNU/Linux 5.10.102.1-microsoft-standard-WSL2 x86_64)
        ```

        ###### 2.5 Click on the Windows "Start" icon and type in the name of your distribution into the search box.
        Your new linux distribution should appear as a local "App". You can pin it to the Windows task bar or to Start for your future convenience.
        Start a VM using your distribution by clicking on the App.

        ###### 3. Install pre-requisites into your new VM
        ###### 3.1 update and apply apt-get packages
        ```
        sudo apt-get update
        sudo apt-get upgrade
        ```
        
        ###### 3.2 Install golang
        ```
        wget https://golang.org/dl/go1.19.linux-amd64.tar.gz
        sudo tar -zxvf go1.19.linux-amd64.tar.gz -C /usr/local
        echo export GOROOT=/usr/local/go | sudo tee -a /etc/profile
        echo export PATH="$PATH:/usr/local/go/bin" | sudo tee -a /etc/profile
        source /etc/profile
        go version
        ```
        
        ###### 3.3 Install ko (but don't do ko set action step)
        ```
        go install github.com/google/ko@latest
        ```

        ###### 3.4 Install gcc
        Either run this:
        ```
        sudo apt install build-essential
        ```
        or this:
        ```
        sudo apt-get update
        apt install gcc
        gcc --version
        ```

        ###### 3.5 Install make (if you installed build-essential this may already be installed)
        ```
        apt install make
        ```

        ###### 3.6 Install jq
        ```
        DEBIAN_FRONTEND=noninteractive apt-get install -y jq
        jq --version
        ```

        ###### 3.7 install kubectl
        ```
        curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
        curl -LO "https://dl.k8s.io/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl.sha256"
        echo "$(cat kubectl.sha256)  kubectl" | sha256sum --check
        sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
        ```

!!! tip "Required Packages for the example usage:"
    === "General"
        You will need the following tools for the example usage of KubeStellar in this quickstart example.
        Select the tab for your environment for suggested commands to install them

        + [__docker__](https://docs.docker.com/engine/install/) or [__podman__](https://podman.io/), to support `kind`

        + [__kind__](https://kind.sigs.k8s.io/docs/user/quick-start/)

    === "Mac"

        ``` title="docker - https://docs.docker.com/engine/install/"
        brew install docker
        open -a Docker
        ```
        ``` title="kind - https://kind.sigs.k8s.io/docs/user/quick-start/"
        brew install kind
        ```

    === "Ubuntu"

        ``` title="docker - https://docs.docker.com/engine/install/"
        sudo mkdir -p /etc/apt/keyrings
        curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
        echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
        sudo apt update
        sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
        ```
        ``` title="Enable rootless usage of Docker (requires relogin) - https://docs.docker.com/engine/security/rootless/"
        sudo apt-get install -y dbus-user-session # *** Relogin after this
        sudo apt-get install -y uidmap
        dockerd-rootless-setuptool.sh install
        systemctl --user restart docker.service
        ```
        ``` title="kind - https://kind.sigs.k8s.io/docs/user/quick-start/"
        curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.14.0/kind-linux-$(dpkg --print-architecture) && chmod +x ./kind && sudo mv ./kind /usr/local/bin
        ```

    === "Debian"

        ``` title="docker - https://docs.docker.com/engine/install/"
        # Add Docker's official GPG key:
        sudo apt-get update
        sudo apt-get install ca-certificates curl gnupg
        sudo install -m 0755 -d /etc/apt/keyrings
        curl -fsSL https://download.docker.com/linux/debian/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
        sudo chmod a+r /etc/apt/keyrings/docker.gpg
        
        # Add the repository to Apt sources:
        echo \
          "deb [arch="$(dpkg --print-architecture)" signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian \
          "$(. /etc/os-release && echo "$VERSION_CODENAME")" stable" | \
          sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
        sudo apt-get update
        
        # Install packages
        sudo apt-get install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
        ```
        ``` title="Enable rootless usage of Docker (requires relogin) - https://docs.docker.com/engine/security/rootless/"
        sudo apt-get install -y dbus-user-session # *** Relogin after this
        sudo apt-get install -y fuse-overlayfs
        sudo apt-get install -y slirp4netns
        dockerd-rootless-setuptool.sh install
        ```
        ``` title="kind - https://kind.sigs.k8s.io/docs/user/quick-start/"
        curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.14.0/kind-linux-$(dpkg --print-architecture) && chmod +x ./kind && sudo mv ./kind /usr/local/bin
        ```

    === "Fedora/RHEL/CentOS"

        ``` title="docker - https://docs.docker.com/engine/install/"
        yum -y install epel-release && yum -y install docker && systemctl enable --now docker && systemctl status docker
        ```
        Enable rootless usage of Docker by following the instructions at [https://docs.docker.com/engine/security/rootless/](https://docs.docker.com/engine/security/rootless/)
        ``` title="kind - https://kind.sigs.k8s.io/docs/user/quick-start/"
        # For AMD64 / x86_64
        [ $(uname -m) = x86_64 ] && curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.14.0/kind-linux-amd64
        # For ARM64
        [ $(uname -m) = aarch64 ] && curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.14.0/kind-linux-arm64 
        chmod +x ./kind && sudo mv ./kind /usr/local/bin/kind
        ```

    === "Windows"

        ``` title="docker - https://docs.docker.com/engine/install/"
        choco install docker -y
        ```
        ``` title="kind - https://kind.sigs.k8s.io/docs/user/quick-start/"
        curl.exe -Lo kind-windows-amd64.exe https://kind.sigs.k8s.io/dl/v0.14.0/kind-windows-amd64
        ```
    === "WSL with Ubuntu"  
        ## How to install docker and kind into a Windows Subsystem for Linux (WSL) environment using an Ubuntu 22.04.01 distribution

        ###### 1.0 Start a VM terminal by clicking on the App you configured using the instructions in the General pre-requisites described above.
        
        ###### 2.0 Install docker
        The installation instructions from docker are not sufficient to get docker working with WSL

        ###### 2.1 Follow instructions here to install docker https://docs.docker.com/engine/install/ubuntu/

        Here some additonal steps you will need to take:

        ###### 2.2 Ensure that /etc/wsl.conf is configured so that systemd will run on booting.
        If /etc/wsl.conf does not contain [boot] systemd=true, then edit /etc/wsl.com as follows:
        ```
        sudo vi /etc/wsl.conf
        ```
        Insert
        ```
        [boot]
        systemd=true
        ```

        ###### 2.3 Edit /etc/sudoers: it is strongly recommended to not add directives directly to /etc/sudoers, but instead to put them in files in /etc/sudoers.d which are auto-included. So make/modify a new file via
        ```
        sudo vi /etc/sudoers.d/docker
        ``` 
        Insert
        ```
        # Docker daemon specification
        <your user account> ALL=(ALL) NOPASSWD: /usr/bin/dockerd
        ```

        ###### 2.4 Add your user to the docker group
        ```
        sudo usermod -aG docker $USER
        ```

        ###### 2.5 If dockerd is already running, then stop it and restart it as follows (note: the new dockerd instance will be running in the foreground):
        ```
        sudo systemctl stop docker
        sudo dockerd &
        ```

        ###### 2.5.1 If you encounter an iptables issue, which is described here: https://github.com/microsoft/WSL/issues/6655 
        The following commands will fix the issue: 
        ```
        sudo update-alternatives --set iptables /usr/sbin/iptables-legacy
        sudo update-alternatives --set ip6tables /usr/sbin/ip6tables-legacy
        sudo dockerd & 
        ``` 

        ###### 3. You will now need to open new terminals to access the VM since dockerd is running in the foreground of this terminal   

        ###### 3.1 In your new terminal, install kind
        ```
        wget -nv https://github.com/kubernetes-sigs/kind/releases/download/v0.17.0/kind-linux-$(dpkg --print-architecture) -O kind 
        sudo install -m 0755 kind /usr/local/bin/kind 
        rm kind 
        kind version
        ```

<!--required-packages-end-->
<!-- 
## 
  - [docker](https://docs.docker.com/engine/install/)
  - [kind](https://kind.sigs.k8s.io/)
  - [kubectl](https://kubernetes.io/docs/tasks/tools/) (version range expected: 1.23-1.25)
  - [jq](https://stedolan.github.io/jq/download/) -->
  
