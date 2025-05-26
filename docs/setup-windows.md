
##  KubeStellar Installation Guide for Windows (PowerShell/CMD)

## Work in Progress -- More steps coming soon!!!!!!

### Prerequisites

You need to install the following tools:

1. [Git for Windows](https://git-scm.com/)
2. [Docker Desktop](https://www.docker.com/products/docker-desktop/)
3. [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl-windows/)
4. [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
5. [Helm](https://helm.sh/docs/intro/install/)
6. [OCM CLI (`clusteradm`)](https://open-cluster-management.io/)
7. [kubeflex CLI (`kflex`)](https://github.com/kubestellar/kubeflex)



###  Step-by-Step Windows Installation

### Invoke-WebRequest -> This is a built-in PowerShell command used to download files or access web pages. It's similar to curl or wget in Linux/macOS

### -Uri -> This specifies the URL you want to request/download from

### -OutFile -> This specifies where to save the downloaded file on your computer

### "$env:USERPROFILE\kind.exe"  -> This uses an environment variable $env:USERPROFILE to refer to your Windows user folder (like C:\Users\YourName), and saves the file there as kind.exe



#### 1. Install kubectl

```powershell
Invoke-WebRequest -Uri "https://dl.k8s.io/release/v1.32.2/bin/windows/amd64/kubectl.exe" -OutFile "$env:USERPROFILE\kubectl.exe"
```

* Add to `PATH`:

  * Press `Win + R` → `SystemPropertiesAdvanced` → Environment Variables
  * Add: `C:\Users\<YourUsername>\` to `Path`

* Test:
```powershell
kubectl version --client
```

---

#### 2. Install kind

```powershell
Invoke-WebRequest -Uri https://kind.sigs.k8s.io/dl/v0.22.0/kind-windows-amd64 -OutFile "$env:USERPROFILE\kind.exe"
```

* Rename and move (optional):

```powershell
Rename-Item "$env:USERPROFILE\kind.exe" "kind.exe"
Move-Item "$env:USERPROFILE\kind.exe" "C:\Program Files\kind.exe"
```

* Add to PATH or run directly from folder

* Test:
```powershell
kind version
```

---

#### 3. Install Helm

```powershell
Invoke-WebRequest -Uri https://get.helm.sh/helm-v3.14.0-windows-amd64.zip -OutFile helm.zip
Expand-Archive .\helm.zip -DestinationPath .\
Move-Item .\windows-amd64\helm.exe "$env:USERPROFILE\helm.exe"
```

* Add to PATH and test:

```powershell
helm version
```

---

#### 4. Install `clusteradm` (OCM CLI)

```powershell
Invoke-WebRequest -Uri https://github.com/open-cluster-management-io/clusteradm/releases/download/v0.10.1/clusteradm-windows-amd64.exe -OutFile "$env:USERPROFILE\clusteradm.exe"
```

* Add to PATH and test:

```powershell
clusteradm version
```

---

#### 5. Install `kflex` (kubeflex CLI)

```powershell
Invoke-WebRequest -Uri https://github.com/kubestellar/kubeflex/releases/download/v0.8.9/kflex-windows-amd64.exe -OutFile "$env:USERPROFILE\kflex.exe"
```

* Rename to `kflex.exe`, add to PATH and test:

```powershell
kflex version
```

---

###  Final Sanity Check

You can verify all prerequisites using:

```powershell
curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/main/scripts/check_pre_req.sh | bash
```

 This one still needs **Git Bash or WSL** due to Unix-style shebangs. You can skip this if you’ve manually confirmed all binaries.



### Optional: Create a kind cluster

```powershell
kind create cluster --name kubestellar
kubectl cluster-info --context kind-kubestellar
```
---

