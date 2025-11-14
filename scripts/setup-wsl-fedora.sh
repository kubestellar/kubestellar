#!/usr/bin/env bash
# Copyright 2024 The KubeStellar Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


set -euo pipefail


# Global constants
BLUE_COLOR="94"
GREEN_COLOR="32"
RED_COLOR="91"
YELLOW_COLOR="93"
TITLE_COLOR=$BLUE_COLOR
PROGRESS_COLOR=$YELLOW_COLOR
SUCCESS_COLOR=$GREEN_COLOR


# Script arguments
PRETTY=false


# Function to print text in color
echo_color() {
  local color_code="$1"
  local text="$2"
  # Print the text with the given color code, then reset the color
  echo -e "\033[${color_code}m${text}\033[0m"
}


# Function to print a title in green, enclosed by ###
title() {
  local text="$1"
  local prefix="---< "
  local suffix=" >---"
  local base="${prefix}${text}${suffix}"
  local total_width=80
  local base_len=${#base}

  # Calculate how many '-' are needed to pad to 40 chars
  local pad=$(( total_width - base_len ))
  if (( pad < 0 )); then
    pad=0
  fi

  local line="${base}$(printf '%*s' "$pad" '' | tr ' ' '-')"

  echo_color "$TITLE_COLOR" "$line"
}


progress() {
  local text="$1"
  echo_color "$PROGRESS_COLOR" "$text"
}


success() {
  local text="$1"
  echo_color "$SUCCESS_COLOR" "$text"
}


# Parse arguments
while [[ $# -gt 0 ]]; do
  case "$1" in
    -X)
      # Enable command tracing
      set -x
      shift
      ;;
    -H|--help)
      echo "Available options:"
      echo "  -X              Enable command tracing (set -x)"
      echo "  -H, --help      Show this help message"
      echo "  -P, --pretty    Enable pretty mode (sets PRETTY=true)"
      exit 0
      ;;
    -P|--pretty)
      PRETTY=true
      shift
      ;;
    *)
      echo "Unknown option: $1"
      echo "Use -H or --help for usage information."
      exit 1
      ;;
  esac
done



title "🔄 Updating system packages"
sudo dnf -y upgrade


# -------------------------------
# Essential baseline programs
# -------------------------------
title "📦 Installing base developer tools"
sudo dnf -y install openssl git curl wget tar unzip 7zip make jq yq gcc nano micro vim fastfetch gawk golang
success "✅ Baseline programs installed."


# -------------------------------
# Docker & Docker Compose
# -------------------------------
title "📦 Installing Docker"

progress "🐳 Installing Docker and Docker Compose..."
sudo dnf -y install dnf-plugins-core
sudo dnf config-manager addrepo --from-repofile=https://download.docker.com/linux/fedora/docker-ce.repo --overwrite
sudo dnf -y install docker-ce docker-ce-cli containerd.io docker-compose-plugin

progress "⚙️ Enabling and starting Docker service..."
sudo systemctl enable docker
sudo systemctl start docker

progress "👤 Adding current user to docker group (no sudo needed)..."
sudo groupadd -f docker
sudo usermod -aG docker $USER
success "✅ Docker setup complete."


# -------------------------------
# kind
# -------------------------------
title "📦 Installing kind (Kubernetes in Docker)"
curl -SfLo ./kind https://kind.sigs.k8s.io/dl/latest/kind-linux-amd64
sudo install -m 555 kind /usr/local/bin/kind
rm kind
progress "⚙️ Icreasing setting for more than 2 kind clusters..."
sudo tee -a /etc/sysctl.conf << EOF
# Enable more than 2 Kind clusters
# https://kind.sigs.k8s.io/docs/user/known-issues/#pod-errors-due-to-too-many-open-files
fs.inotify.max_user_watches = 524288
fs.inotify.max_user_instances = 512
EOF
success "✅ kind installed."


# -------------------------------
# Helm
# -------------------------------
title "⛵ Installing Helm"
curl -Sf https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
success "✅ Helm installed."


# -------------------------------
# kubectl
# -------------------------------
title "📡 Installing kubectl"
curl -SfLO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -m 555 kubectl /usr/local/bin/kubectl
rm kubectl
success "✅ kubectl installed."


# -------------------------------
# argo cli
# -------------------------------
title "📦 Installing Argo CLI"
ARGO_VERSION=$(curl -s "https://api.github.com/repos/argoproj/argo-cd/releases/latest" | jq -r '.tag_name')
curl -SfL -o argocd https://github.com/argoproj/argo-cd/releases/download/$ARGO_VERSION/argocd-linux-amd64
sudo install -m 555 argocd /usr/local/bin/argocd
rm argocd
argocd version --client
success "✅ Argo CLI installed."


# -------------------------------
# ko
# -------------------------------
title "📦 Installing KO"
KO_VERSION=$(curl -s https://api.github.com/repos/ko-build/ko/releases/latest | jq -r '.tag_name')
curl -SfL "https://github.com/ko-build/ko/releases/download/${KO_VERSION}/ko_${KO_VERSION#v}_$(uname -s)_$(uname -m).tar.gz" -o ko.tar.gz
tar -xzf ko.tar.gz --touch ko
sudo install -m 555 ko /usr/local/bin/ko
rm ko ko.tar.gz
ko version
success "✅ KO installed."


# -------------------------------
# oc
# -------------------------------
title "📦 Installing oc cli"
OC_VERSION=$(curl -s https://api.github.com/repos/okd-project/okd/releases/latest | jq -r '.tag_name')
curl -SfL "https://github.com/okd-project/okd/releases/latest/download/openshift-client-linux-${OC_VERSION}.tar.gz" -o oc.tar.gz
tar -xzf oc.tar.gz --touch oc
sudo install -m 555 oc /usr/local/bin/oc
rm oc oc.tar.gz
oc version
success "✅ oc installed."


# -------------------------------
# OCM clusteradmin
# -------------------------------
title "📦 Installing OCM clusteradmin"
bash <(curl -s -L https://raw.githubusercontent.com/open-cluster-management-io/clusteradm/main/install.sh) v0.10.1
clusteradm version || true
success "✅ OCM clusteradmin installed."


# -------------------------------
# kflex
# -------------------------------
title "📦 Installing kflex..."
sudo su <<EOF
bash <(curl -s https://raw.githubusercontent.com/kubestellar/kubeflex/main/scripts/install-kubeflex.sh) --ensure-folder /usr/local/bin --strip-bin --verbose
EOF
success "✅ kflex installed."


if [ "$PRETTY" = true ]; then

	# -------------------------------
	# On-My-Posh
	# -------------------------------
	title "✨ Installing On-My-Posh"
	curl -Sf https://ohmyposh.dev/install.sh | bash
	echo 'eval "$(oh-my-posh init bash --config ~/.cache/oh-my-posh/themes/catppuccin_frappe.omp.json)"' >> ~/.bashrc
	success "✅ On-My-Posh installed and configured."

	# -------------------------------
	# kubecolor
	# -------------------------------
	title "🎨 Installing kubecolor..."
	sudo dnf install -y dnf5-plugins
	sudo dnf config-manager addrepo --from-repofile https://kubecolor.github.io/packages/rpm/kubecolor.repo --overwrite
	sudo dnf install -y kubecolor
	echo 'alias kubectl=kubecolor' >> ~/.bashrc
	echo 'alias k=kubecolor' >> ~/.bashrc
	success "✅ kubecolor installed and alias 'k' created."

fi

# -------------------------------
# Final summary
# -------------------------------
success "🎉 Setup complete!"
echo_color $RED_COLOR "👉 Reload your shell to activate some of the features."
