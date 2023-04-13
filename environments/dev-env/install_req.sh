#!/usr/bin/env bash

get_os_type() {
  case "$OSTYPE" in
      darwin*)  echo "darwin" ;;
      linux*)   echo "linux" ;;
      *)        echo "unknown: $OSTYPE" && exit 1 ;;
  esac
}

os_type=$(get_os_type)

# Install golang for x86_64
if [ $os_type == "linux" ]; then
    if !(command -v go) then
        sudo apt update
        sudo apt upgrade -y
        curl -O https://dl.google.com/go/go1.19.2.linux-amd64.tar.gz
        tar -xvf go1.19.2.linux-amd64.tar.gz
        sudo mv go  /usr/bin/
        sudo echo 'export GOROOT=/usr/bin/go' >> ~/.bashrc 
        sudo echo 'export PATH=$PATH:$GOROOT/bin' >> ~/.bashrc 
        sleep 5
        exec bash
    else
        echo "Go is already installed"
        echo "***********************"
    fi
elif [ $os_type == "darwin" ]; then
    if !(command -v go) then
        brew update
        curl -O https://dl.google.com/go/go1.19.2.darwin-amd64.pkg
        tar -xvf go1.19.2.darwin-amd64.pkg
        sudo mv go  /usr/bin/
        sudo echo 'export GOROOT=/usr/bin/go' >> ~/.bashrc 
        sudo echo 'export PATH=$PATH:$GOROOT/bin' >> ~/.bashrc 
        sleep 5
        exec bash
    else
        echo "Go is already installed"
        echo "***********************"
    fi
fi


# Install make
if [ $os_type == "linux" ]; then
    if !(dpkg -l make) then
        apt update
        apt install -y make
    else
    echo "Make is already installed"
    echo "*************************"
    fi
elif [ $os_type == "darwin" ]; then
    if !(command -v make) then
        brew update
        brew install make
    else
        echo "Make is already installed"
        echo "*************************"
    fi
fi


# Install jq
if [ $os_type == "linux" ]; then
    if !(dpkg -l jq) then
        apt update
        apt install -y jq
    else
        echo "jq is already installed"
        echo "****************************"
    fi
elif [ $os_type == "darwin" ]; then
    if !(command -v jq) then
        brew update
        brew install jq
    else
        echo "jq is already installed"
        echo "****************************"
    fi
fi


# Install ko
if [ $os_type == "linux" ]; then
    if !(dpkg -l ko) then
        apt update
        VERSION=0.13.0 
        OS=Linux
        ARCH=x86_64
        curl -sSfL "https://github.com/ko-build/ko/releases/download/v${VERSION}/ko_${VERSION}_${OS}_${ARCH}.tar.gz" > ko.tar.gz 
        curl -sSfL https://github.com/ko-build/ko/releases/download/v${VERSION}/attestation.intoto.jsonl > provenance.intoto.jsonl
        slsa-verifier -artifact-path ko.tar.gz -provenance provenance.intoto.jsonl -source github.com/google/ko -tag "v${VERSION}"
        tar xzf ko.tar.gz ko
        chmod +x ./ko
        mv ./ko  /usr/local/bin/kind
        rm ko.tar.gz
    else
        echo "ko is already installed"
        echo "****************************"
    fi
elif [ $os_type == "darwin" ]; then
    if !(command -v ko) then
        brew update
        brew install ko
    else
        echo "ko is already installed"
        echo "****************************"
    fi
fi

# Install gcc
if [ $os_type == "linux" ]; then
    if !(dpkg -l gcc) then
        apt update
        apt install -y build-essential
    else
        echo "gcc is already installed"
        echo "****************************"
    fi 
elif [ $os_type == "darwin" ]; then
    if !(command -v gcc) then
        brew update
        brew install gcc
    else
        echo "gcc is already installed"
        echo "****************************"
    fi 
fi


# Install Docker
if [ $os_type == "linux" ]; then
    if !(dpkg -l docker) then
        sudo apt update -y
        sudo apt install docker.io -y
        sudo systemctl enable docker
        sudo systemctl start docker
    else
        echo "docker is already installed"
        echo "****************************"
    fi
elif [ $os_type == "darwin" ]; then
     if !(command -v docker) then
        echo "WARNING: PLEASE install Docker ....."
     else
        echo "docker is already installed"
        echo "****************************"
     fi
fi


# Install kubectl
if [ $os_type == "linux" ]; then
    if !(dpkg -l kubectl) then
        sudo apt-get update
        sudo apt-get install -y ca-certificates curl
        sudo curl -fsSLo /usr/share/keyrings/kubernetes-archive-keyring.gpg https://packages.cloud.google.com/apt/doc/apt-key.gpg
        echo "deb [signed-by=/usr/share/keyrings/kubernetes-archive-keyring.gpg] https://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee /etc/apt/sources.list.d/kubernetes.list
        sudo apt-get update
        sudo apt-get install -y kubectl
    else
        echo "kubectl is already installed"
        echo "****************************"
    fi
elif [ $os_type == "darwin" ]; then
     if !(command -v kubectl) then
        brew update
        brew install kubectl
     else
        echo "kubectl is already installed"
        echo "****************************"
     fi
fi


# Install kind
if [ $os_type == "linux" ]; then
    if !(command -v kind) then
        curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.16.0/kind-linux-amd64
        chmod +x ./kind
        sudo mv ./kind /usr/local/bin/kind
    else
        echo "kind is already installed"
        echo "*************************"
    fi
elif [ $os_type == "darwin" ]; then
     if !(command -v kind) then
        brew update
        brew install kind
     else
        echo "kind is already installed"
        echo "*************************"
     fi
fi