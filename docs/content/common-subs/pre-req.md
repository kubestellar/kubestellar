```shell
os_type=""
arch_type=""
folder=""
get_os_type() {
  case "$OSTYPE" in
      linux*)   echo "linux" ;;
      darwin*)  echo "darwin" ;;
      *)        echo "Unsupported operating system type: $OSTYPE" >&2 ; exit 1 ;;
  esac
}

get_arch_type() {
  case "$HOSTTYPE" in
      x86_64*)  echo "amd64" ;;
      aarch64*) echo "arm64" ;;
      arm64*)   echo "arm64" ;;
      *)        echo "Unsupported architecture type: $HOSTTYPE" >&2 ; exit 1 ;;
  esac
}

get_os_type() {
  case "$OSTYPE" in
      linux*)   echo "linux" ;;
      darwin*)  echo "darwin" ;;
      *)        echo "Unsupported operating system type: $OSTYPE" >&2 ; exit 1 ;;
  esac
}

get_arch_type() {
  case "$HOSTTYPE" in
      x86_64*)  echo "amd64" ;;
      aarch64*) echo "arm64" ;;
      arm64*)   echo "arm64" ;;
      *)        echo "Unsupported architecture type: $HOSTTYPE" >&2 ; exit 1 ;;
  esac
}

if [ "$os_type" == "" ]; then
    os_type=$(get_os_type)
fi

if [ "$arch_type" == "" ]; then
    arch_type=$(get_arch_type)
fi

if [ "$folder" == "" ]; then
    folder="$PWD"
fi

echo $os_type
echo $arch_type
echo $folder

if command -v docker >/dev/null 2>&1; then
    echo "Docker is installed"
else
    if [ "$os_type" == "darwin" ]; then
      brew install docker
    fi
fi

if docker info >/dev/null 2>&1; then
    echo "Docker is started"
else
    if [ "$os_type" == "darwin" ]; then
      source ./zshrc
      open --background -a Docker
    fi
fi

if command -v go >/dev/null 2>&1; then
    echo "GO is installed"
else
    if [ "$os_type" == "darwin" ]; then
      brew install go@1.19
    fi
fi

if command -v kubectl >/dev/null 2>&1; then
    echo "kubectl is installed"
else
    if [ "$os_type" == "darwin" ]; then
      brew install kubectl
    fi
fi

if command -v jq >/dev/null 2>&1; then
    echo "jq is installed"
else
    if [ "$os_type" == "darwin" ]; then
      brew install jq
    fi
fi

if command -v kind >/dev/null 2>&1; then
    echo "kind is installed"
else
    if [ "$os_type" == "darwin" ]; then
      brew install kind
    fi
fi

ps -ef | grep mailbox-controller | grep -v grep | awk '{print $2}' | xargs kill
ps -ef | grep kubestellar-scheduler | grep -v grep | awk '{print $2}' | xargs kill
ps -ef | grep placement-translator | grep -v grep | awk '{print $2}' | xargs kill
ps -ef | grep kcp | grep -v grep | awk '{print $2}' | xargs kill
kind delete cluster --name florin
kind delete cluster --name guilder
```