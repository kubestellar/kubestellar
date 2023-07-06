<!--required-packages-start-->
!!! tip "Required Packages:"
    === "General"
        You will need the following tools to run the KubeStellar quickstart. Check the appropriate tab for your environment for suggested commands to install them
        ``` title="jq  - https://stedolan.github.io/jq/download/"         
        ```
        ``` title="yq  - https://github.com/mikefarah/yq#install"         
        ```
        ``` title="docker  - https://docs.docker.com/engine/install/"         
        ```
        ``` title="kind  - https://kind.sigs.k8s.io/docs/user/quick-start/"         
        ```
        ``` title="kubectl  - https://kubernetes.io/docs/tasks/tools/" (version range expected: 1.23-1.25)         ```
        ``` title="ko - https://github.com/ko-build/ko" (required for compiling KubeStellar Syncer)
        ```
        ``` title="GO v1.19 -(https://gist.github.com/jniltinho/8758e15a9ef80a189fce)"         You will need GO to compile and run kcp and the Kubestellar processes         
        ```
    === "Mac"
        ``` title="jq - https://stedolan.github.io/jq/download/"
        brew install jq
        ```
        ``` title="yq - https://github.com/mikefarah/yq#install"
        brew install yq
        ```
        ``` title="docker - https://docs.docker.com/engine/install/"
        brew install docker
        open -a Docker
        ```
        ``` title="kind - https://kind.sigs.k8s.io/docs/user/quick-start/"
        brew install kind
        ```
        ``` title="kubectl - https://kubernetes.io/docs/tasks/tools/ (version range expected: 1.23-1.25)"
        brew install kubectl
        ```
        ``` title="ko - https://github.com/ko-build/ko (required for compiling KubeStellar Syncer)"
        brew install ko
        ```
        [GO v1.19](https://gist.github.com/jniltinho/8758e15a9ef80a189fce) - You will need GO to compile and run kcp and the KubeStellar processes.  Currently kcp requires go version 1.19.
    === "Ubuntu"
        ``` title="jq - https://stedolan.github.io/jq/download/"
        sudo apt-get install jq
        ```
        ``` title="yq - https://github.com/mikefarah/yq#install"
        sudo apt-get install yq
        ```
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
        ``` title="kubectl - https://kubernetes.io/docs/tasks/tools/ (version range expected: 1.23-1.25)"
        curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/$(dpkg --print-architecture)/kubectl && chmod +x kubectl && sudo mv ./kubectl /usr/local/bin/kubectl
        ```
        ``` title="ko - https://github.com/ko-build/ko (required for compiling KubeStellar Syncer)"
        VERSION=TODO # choose the latest version (without v prefix)
        OS=Linux     # or Darwin
        ARCH=x86_64  # or arm64, i386, s390x
        curl -sSfL "https://github.com/ko-build/ko/releases/download/v${VERSION}/ko_${VERSION}_${OS}_${ARCH}.tar.gz" > ko.tar.gz
        curl -sSfL https://github.com/ko-build/ko/releases/download/v${VERSION}/attestation.intoto.jsonl > provenance.intoto.jsonl
        slsa-verifier -artifact-path ko.tar.gz -provenance provenance.intoto.jsonl -source github.com/google/ko -tag "v${VERSION}"
        tar xzf ko.tar.gz ko
        chmod +x ./ko
        ```
        ``` title="GO - You will need GO to compile and run kcp and the KubeStellar components.  Currently kcp requires go version 1.19"
        curl -L "https://go.dev/dl/go1.19.5.linux-$(dpkg --print-architecture).tar.gz" -o go.tar.gz
        tar -C /usr/local -xzf go.tar.gz
        rm go.tar.gz
        echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
        source /etc/profile
        go version
        ```
    === "Fedora/RHEL/CentOS"
        ``` title="jq - https://stedolan.github.io/jq/download/"
        yum -y install jq
        ```
        ``` title="yq - https://github.com/mikefarah/yq#install"
        # easiest to install with snap
        snap install yq
        ```
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
        ``` title="kubectl - https://kubernetes.io/docs/tasks/tools/ (version range expected: 1.23-1.25)"
        # For AMD64 / x86_64
        [ $(uname -m) = x86_64 ] && curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && chmod +x kubectl && mv ./kubectl /usr/local/bin/kubectl
        # for ARM64 / aarch64
        [ $(uname -m) = aarch64 ] && curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/arm64/kubectl" && chmod +x kubectl && mv ./kubectl /usr/local/bin/kubectl
        ```
        ``` title="ko - https://github.com/ko-build/ko (required for compiling KubeStellar Syncer)"
        VERSION="0.14.1" # choose the latest version (without v prefix)
        OS="Linux"    # or Darwin
        # set proper architecture ( ARCH=x86_64  # or arm64, i386, s390x)
        # for AMD64 / x86_64
        [ $(uname -m) = x86_64 ] && ARCH="x86_64"
        # for ARM64 / aarch64
        [ $(uname -m) = aarch64 ] && ARCH="arm64"
        # (simplified install without running slsa-verifier)
        curl -sSfL "https://github.com/ko-build/ko/releases/download/v${VERSION}/ko_${VERSION}_${OS}_${ARCH}.tar.gz" > ko.tar.gz
        tar xzf ko.tar.gz ko
        chmod +x ./ko && sudo mv ./ko /usr/local/bin/ko
        ```
        [GO v1.19](https://gist.github.com/jniltinho/8758e15a9ef80a189fce) - You will need GO to compile and run kcp and the KubeStellar processes.  Currently kcp requires go version 1.19.
    === "Windows"
        ``` title="Chocolatey - https://chocolatey.org/install#individual"
        Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))
        ```
        ``` title="jq - https://stedolan.github.io/jq/download/"
        choco install jq -y
        choco install curl -y
        ```
        ``` title="yq - https://github.com/mikefarah/yq#install"
        choco install yq -y
        ```
        ``` title="docker - https://docs.docker.com/engine/install/"
        choco install docker -y
        ```
        ``` title="kind - https://kind.sigs.k8s.io/docs/user/quick-start/"
        curl.exe -Lo kind-windows-amd64.exe https://kind.sigs.k8s.io/dl/v0.14.0/kind-windows-amd64
        ```
        ``` title="kubectl - https://kubernetes.io/docs/tasks/tools/install-kubectl-windows/ (version range expected: 1.23-1.25)"
        curl.exe -LO "https://dl.k8s.io/release/v1.27.2/bin/windows/amd64/kubectl.exe"
        ```
        ``` title="ko - https://github.com/ko-build/ko (required for compiling KubeStellar Syncer)"
        VERSION=TODO # choose the latest version (without v prefix)
        OS=Linux     # or Darwin
        ARCH=x86_64  # or arm64, i386, s390x
        curl -sSfL "https://github.com/ko-build/ko/releases/download/v${VERSION}/ko_${VERSION}_${OS}_${ARCH}.tar.gz" > ko.tar.gz
        curl -sSfL https://github.com/ko-build/ko/releases/download/v${VERSION}/attestation.intoto.jsonl > provenance.intoto.jsonl
        slsa-verifier -artifact-path ko.tar.gz -provenance provenance.intoto.jsonl -source github.com/google/ko -tag "v${VERSION}"
        tar xzf ko.tar.gz ko
        chmod +x ./ko
        ```
        [GO v1.19](https://gist.github.com/jniltinho/8758e15a9ef80a189fce) - You will need GO to compile and run kcp and the KubeStellar processes.  Currently kcp requires go version 1.19.
<!--required-packages-end-->
<!-- 
## 
  - [docker](https://docs.docker.com/engine/install/)
  - [kind](https://kind.sigs.k8s.io/)
  - [kubectl](https://kubernetes.io/docs/tasks/tools/) (version range expected: 1.23-1.25)
  - [jq](https://stedolan.github.io/jq/download/) -->
  
