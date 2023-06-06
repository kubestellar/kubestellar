```shell
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
```