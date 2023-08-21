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
        sudo apt-get install yq
        ```
        ``` title="kubectl - https://kubernetes.io/docs/tasks/tools/ (version range expected: 1.23-1.25)"
        curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/$(dpkg --print-architecture)/kubectl && chmod +x kubectl && sudo mv ./kubectl /usr/local/bin/kubectl
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
        ```
        ```
        # How to install pre-requisites for a Windows Subsystem for Linux (WSL) envronment using an Ubuntu 22.04.01 distribution

        Tested on a Intel(R) Core(TM) i7-9850H CPU @ 2.60GHz 2.59 GHz with 32GB RAM, a 64-bit operating system, x64-based processor
        Using Windows 11 Enterprise

        # 1. NB: If you're using a VPN, turn it off

        # 2. Install Ubuntu into WSL

        ## 2.1. In a Windows command terminal run the following to list all the linux distributions that are available online
        ```
        wsl -l -o
        ##2.2 Select a linux distribution and install it into WSL, like this:
        ```
        wsl --install -d Ubuntu 22.04.01
        ```
        You will see something like this:
        Installing, this may take a few minutes...
        Please create a default UNIX user account. The username does not need to match your Windows username.
        For more information visit: https://aka.ms/wslusers
        Enter new UNIX username:

        ## 2.3 Enter your new username and password at the prompts, and you will eventually see something like this
        ```
        Welcome to Ubuntu 22.04.1 LTS (GNU/Linux 5.10.102.1-microsoft-standard-WSL2 x86_64)
        ```

        # 2.4 Click on the Windows "Start" icon and type in the name of your distribution into the search box.
        Your new linux distribution should appear as a local "App". You can pin it to the Windows task bar or to Start for your future convenience.
        Start a VM using your distribution by clicking on the App.

        # 3. Install pre-requisites into your new VM
        ## 3.1 update apt-get packages
        ```
        sudo apt-get update
        ```
        ## 3.2 Install golang
        ```
        wget https://golang.org/dl/go1.19.linux-amd64.tar.gz
        sudo tar -zxvf go1.19.linux-amd64.tar.gz -C /usr/local
        echo export GOROOT=/usr/local/go | sudo tee -a /etc/profile
        echo export PATH=$PATH:/usr/local/go/bin | sudo tee -a /etc/profile
        source /etc/profile
        go version
        ```
        ######### the following didn't work, probably because the kcp wasn't installed yet - no instruction re kcp install
        go install ./cmd/kcp ./cmd/kubectl-*
        #######################################################
        ```
        ## 3.3 Install ko (but don't do ko set action step)
        ```
        go install github.com/google/ko@latest
        ```

        ## 3.4 Install gcc
        You should be able to "sudo apt install build-essential" theoretically, but I had an issue with it
        ```
        sudo apt-get update
        apt install gcc
        gcc --version
        ```
        ## 3.5 Install make
        ```
        apt install make
        ```
        ## 3.6 Install jq
        ```
        DEBIAN_FRONTEND=noninteractive apt-get install -y jq
        jq --version
        ```

        ## 3.6 Install docker
        The installation instructions from docker are not sufficient to get docker working with WSL

        ## 3.6.1 
        Follow instructions here to install docker https://docs.docker.com/engine/install/ubuntu/

        Here some additonal steps you will need to take:

        ## 3.6.2 Edit /etc/wsl.conf so that systemd will run on booting:
        ```
        vi /etc/wsl.conf
        ```
        Insert
        ```
        [boot]
        systemd=true
        ```
        ## 3.6.3 Edit /etc/sudoers 
        ```
        vi /etc/sudoers
        ``` 
        Insert
        ```
        # Docker daemon specification
        <your user account> ALL=(ALL) NOPASSWD: /usr/bin/dockerd
        ```

        ## 3.6.4 Add your user to the docker group
        ```
        sudo usermod -aG docker $USER
        ```

        ## 3.6.5 I encountered the an iptables issue described here: https://github.com/microsoft/WSL/issues/6655 
        The following commands fixed the issue: 
        ```
        sudo update-alternatives --set iptables /usr/sbin/iptables-legacy
        sudo update-alternatives --set ip6tables /usr/sbin/ip6tables-legacy
        sudo dockerd & 
        ```   
        ## 4. You will now need to open new terminals to access the VM since dockerd is running in the foreground of this terminal 
        ## In your new terminal, 

        ## 4.1 Install kind
        ```
        wget -nv https://github.com/kubernetes-sigs/kind/releases/download/v0.17.0/kind-linux-$(dpkg --print-architecture) -O kind 
        sudo install -m 0755 kind /usr/local/bin/kind 
        rm kind 
        kind version
        ```

        ## So far we have ubuntu 22.04.1, docker, kind, go 1.19, ko, make, gcc, jq, 

        ## 4.2 install kubectl
        ```
        curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
        curl -LO "https://dl.k8s.io/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl.sha256"
        echo "$(cat kubectl.sha256)  kubectl" | sha256sum --check
        sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
        ```
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
        ``` title="kind - https://kind.sigs.k8s.io/docs/user/quick-start/"
        curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.14.0/kind-linux-$(dpkg --print-architecture) && chmod +x ./kind && sudo mv ./kind /usr/local/bin
        ```

    === "Fedora/RHEL/CentOS"

        ``` title="docker - https://docs.docker.com/engine/install/"
        yum -y install epel-release && yum -y install docker && systemctl enable --now docker && systemctl status docker
        ```
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

<!--required-packages-end-->
<!-- 
## 
  - [docker](https://docs.docker.com/engine/install/)
  - [kind](https://kind.sigs.k8s.io/)
  - [kubectl](https://kubernetes.io/docs/tasks/tools/) (version range expected: 1.23-1.25)
  - [jq](https://stedolan.github.io/jq/download/) -->
  
