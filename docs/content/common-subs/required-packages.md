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
  
