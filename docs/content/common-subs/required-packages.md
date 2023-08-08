<!--required-packages-start-->
!!! tip "Required Packages:"
    === "General"
        You will need the following tools to run this KubeStellar quickstart example. 
        Select the tab for your environment for suggested commands to install them

        + make (only needed if you do more advanced builds; omitted from OS-specific instructions)

        + __curl__ (omitted from most OS-specific instructions)

        + [__jq__](https://stedolan.github.io/jq/download/)       

        + [__yq__](https://github.com/mikefarah/yq#install)

        + [__docker__](https://docs.docker.com/engine/install/)         

        + [__kind__](https://kind.sigs.k8s.io/docs/user/quick-start/)        

        + [__kubectl__](https://kubernetes.io/docs/tasks/tools/) (version range expected: 1.23-1.25) 

        + [__ko__](https://github.com/ko-build/ko) (required for compiling KubeStellar Syncer)

        + [__GO v1.19__](https://gist.github.com/jniltinho/8758e15a9ef80a189fce)         You will need GO to compile and run kcp and the Kubestellar processes         

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
        ``` title="slsa-verifier - https://github.com/slsa-framework (verify ko signer)
        SLSA_VERIFIER=slsa-verifier-linux-$(dpkg --print-architecture)
        curl -sSfLO https://github.com/slsa-framework/slsa-verifier/releases/latest/download/${SLSA_VERIFIER}
        chmod +x ${SLSA_VERIFIER}
        curl -sSfLO https://github.com/slsa-framework/slsa-verifier/releases/latest/download/${SLSA_VERIFIER}.intoto.jsonl
        ./${SLSA_VERIFIER} verify-artifact ./${SLSA_VERIFIER} --provenance-path ${SLSA_VERIFIER}.intoto.jsonl --source-uri github.com/slsa-framework/slsa-verifier
        sudo install ./${SLSA_VERIFIER} /usr/local/bin/slsa-verifier
        ```
        ``` title="ko - https://github.com/ko-build/ko (required for compiling KubeStellar Syncer)"
        KO_VERSION=$(curl -sSf https://api.github.com/repos/ko-build/ko/releases/latest | jq -r '.tag_name') # latest version (with v prefix)
        OS=$(uname -s)     # Linux, Darwin
        ARCH=$(uname -m)  # x86_64, arm64, i386, s390x
        curl -sSfL "https://github.com/ko-build/ko/releases/download/${KO_VERSION}/ko_${KO_VERSION#v}_${OS}_${ARCH}.tar.gz" > ko.tar.gz
        curl -sSfL https://github.com/ko-build/ko/releases/download/${KO_VERSION}/multiple.intoto.jsonl > multiple.intoto.jsonl
        slsa-verifier verify-artifact ko.tar.gz --provenance-path multiple.intoto.jsonl --source-uri github.com/ko-build/ko --source-tag "${KO_VERSION}"
        tar xzf ko.tar.gz ko
        chmod +x ./ko
        ./ko version
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
        ``` title="curl"
        choco install curl -y
        ```
        ``` title="jq - https://stedolan.github.io/jq/download/"
        choco install jq -y
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
<!--required-packages-end-->
<!-- 
## 
  - [docker](https://docs.docker.com/engine/install/)
  - [kind](https://kind.sigs.k8s.io/)
  - [kubectl](https://kubernetes.io/docs/tasks/tools/) (version range expected: 1.23-1.25)
  - [jq](https://stedolan.github.io/jq/download/) -->
  
