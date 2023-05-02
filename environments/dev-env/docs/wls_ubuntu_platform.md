# Instructions for Windows WSL/Ubuntu platform

## How to install pre-requisites for Windows - Windows Subsystem for Linux (WSL) envronment, using an Ubuntu 22.04.01 distribution

Tested on a Intel(R) Core(TM) i7-9850H CPU @ 2.60GHz 2.59 GHz with 32GB RAM, a 64-bit operating system, x64-based processor
Using Windows 11 Enterprise

### 1. NB: If you're using a VPN, turn it off

### 2. Install Ubuntu

#### 2.1. In a Windows command terminal run the following to list all the linux distributions that are available online
```
wsl -l -o
```
#### 2.2 Select a distribution and install it, like this:
```
wsl --install -d Ubuntu 22.04.01
```
You will see something like this:
Installing, this may take a few minutes...
Please create a default UNIX user account. The username does not need to match your Windows username.
For more information visit: https://aka.ms/wslusers
Enter new UNIX username:

#### 2.3 Enter your new username and password at the prompts, and you will eventually see something like this
```
Welcome to Ubuntu 22.04.1 LTS (GNU/Linux 5.10.102.1-microsoft-standard-WSL2 x86_64)
```

#### 2.4 Click on the Windows "Start" icon and type in the name of your distribution into the search box.
Your new linux distribution show appear as a local "App". You can pin it to the Windows task bar or to Start for your future convenience.
Start a VM using your distribution by clicking on the App.

### 3. Install pre-requisites into your new VM
#### 3.1 update apt-get packages
```
sudo apt-get update
```
#### 3.2 Install golang version 1.19 (currently the kcp build expects this version, although one could "make build IGNORE_GO_VERSION=1")
```
wget https://golang.org/dl/go1.19.linux-amd64.tar.gz
sudo tar -zxvf go1.19.linux-amd64.tar.gz -C /usr/local
echo export GOROOT=/usr/local/go | sudo tee -a /etc/profile
echo export PATH=$PATH:/usr/local/go/bin | sudo tee -a /etc/profile
source /etc/profile
go version
```
#### 3.3 Install ko (but don't do ko set action step)
```
go install github.com/google/ko@latest
```

#### 3.4 Install gcc
You should be able to "sudo apt install build-essential" theoretically, but I had an issue with it
```
sudo apt-get update
apt install gcc
gcc --version
```
#### 3.5 Install make
```
apt install make
```
#### 3.6 Install jq
```
DEBIAN_FRONTEND=noninteractive apt-get install -y jq
jq --version
```

#### 3.6 Install docker
The installation instructions from docker are not sufficient to get docker working with WSL

#### 3.6.1 
Follow instructions here to install docker https://docs.docker.com/engine/install/ubuntu/

Here some additonal steps you will need to take:

#### 3.6.2 Edit /etc/wsl.conf so that systemd will run on booting:
```
vi /etc/wsl.conf
```
Insert the following text into the file: (you can block copy using ctl-shift-v)
```
[boot]
systemd=true
```
#### 3.6.3 Edit /etc/sudoers 
```
vi /etc/sudoers
```
#### Insert the following text, using your own user account name to replace your user account
```
<your user account> ALL=(ALL) NOPASSWD: /usr/bin/dockerd
```  
#### 3.6.4 Add your user to the docker group
```
sudo usermod -aG docker $USER
```

#### 3.6.5 I encountered the iptables issue described here: https://github.com/microsoft/WSL/issues/6655 
The following commands fixed the issue: 
```
sudo update-alternatives --set iptables /usr/sbin/iptables-legacy
sudo update-alternatives --set ip6tables /usr/sbin/ip6tables-legacy
sudo dockerd & 
```   
### 4. You will now need to open a new Windows terminal (see step 2.4 above) to access the VM since dockerd is running in the foreground of this terminal 

In your new terminal, 

#### 4.1 Install kind
```
wget -nv https://github.com/kubernetes-sigs/kind/releases/download/v0.17.0/kind-linux-$(dpkg --print-architecture) -O kind 
sudo install -m 0755 kind /usr/local/bin/kind 
rm kind 
kind version
```

#### So far we have ubuntu 22.04.1, docker, kind, go 1.19, ko, make, gcc, jq, 

#### 4.2 install kubectl
```
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
curl -LO "https://dl.k8s.io/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl.sha256"
echo "$(cat kubectl.sha256)  kubectl" | sha256sum --check
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
```
### 5. Install kcp-edge 

Refer to the instructions [here](../README.md)


