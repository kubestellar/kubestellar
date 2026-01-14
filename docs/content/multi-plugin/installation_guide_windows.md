# Kubectl-multi binary installation guide for windows

### Downloading step for windows (commands for powershell)
```bash
# Step 1: Download kubectl-multi binary for windows

# TAG="v0.0.3"

curl.exe -LO "https://github.com/kubestellar/kubectl-plugin/releases/download/v0.0.3/kubectl-multi_0.0.3_windows_amd64.zip"

# Step 2: Extract and install
Expand-Archive .\kubectl-multi_0.0.3_windows_amd64.zip

# Step 3: making a new directory for plugin 
New-Item -ItemType Directory -Force -Path C:\kubectl-plugins

# Step 4 : navigate to the directory where kubectl-plugin installed - Downloads
Move-Item .\kubectl-multi.exe C:\kubectl-plugins\kubectl-multi.exe

# Step 5:  Add the Folder to Your System PATH
Go to Control Panel → System and Security → System → Advanced system settings → Environment Variables.  

In “System variables”, select Path, click “Edit”, then “New” and enter C:\kubectl-plugins

#to test (if this command don't work at first then try restarting the powershell terminal )
kubectl-multi version

```

### Downloading steps for windows (git bash)
```bash
# Step 1: Download kubectl-multi binary for Windows
# TAG="v0.0.3"
curl -LO "https://github.com/kubestellar/kubectl-plugin/releases/download/v0.0.3/kubectl-multi_0.0.3_windows_amd64.zip"

# Step 2: Extract the ZIP file
unzip kubectl-multi_0.0.3_windows_amd64.zip

# Step 3: Create a new directory for plugins
mkdir -p /c/kubectl-plugins

# Step 4: Move the extracted binary to the plugins directory
mv kubectl-multi.exe /c/kubectl-plugins/

# Step 5: Add the plugins directory to your PATH (session only)
export PATH=$PATH:/c/kubectl-plugins

# To make PATH permanent, add the above export line to your ~/.bashrc and restart Git Bash

# Step 6: Test the plugin
kubectl-multi version

```