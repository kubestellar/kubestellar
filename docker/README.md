# KubeStellar in a container

Table of contents:
- [KubeStellar in a container](#kubestellar-in-a-container)
  - [Install the pre-requisites](#install-the-pre-requisites)
  - [Build a local container image](#build-a-local-container-image)
  - [Build and push a multi-architecture image](#build-and-push-a-multi-architecture-image)
  - [Run the container and access KubeStellar from the host OS](#run-the-container-and-access-kubestellar-from-the-host-os)
  - [Checking the container logs](#checking-the-container-logs)
  - [Login into the running container to use **KubeStellar** from inside the container](#login-into-the-running-container-to-use-kubestellar-from-inside-the-container)
  - [Teardown and cleanup](#teardown-and-cleanup)

## Install the pre-requisites

Install `nake`:

```bash
sudo apt install -y make
```

Install `docker`:

```bash
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
sudo usermod -aG docker $USER
```

Or `podman` (multi-arch `buildx` option will not be available):

```bash
sudo apt install -y podman
alias docker=podman
```

## Build a local container image

Build a local container image `kubestellar` for the host architecture using the latest release of **KubeStellar** with the command:

```bash
make build
```

A desired version of **KubeStellar** can be specified by using the optional `KUBESTELLAR_VERSION` argument, such as `KUBESTELLAR_VERSION=v0.3.1`.

The **KubeStellar** version is used to tag the container image, as shown below:

```bash
$ docker images
REPOSITORY    TAG       IMAGE ID       CREATED          SIZE
kubestellar   v0.3.1    0425aee2cda6   8 minutes ago    617MB
kubestellar   v0.3.0    76d714261c6e   14 minutes ago   617MB
```

## Build and push a multi-architecture image

Note that `docker buildx` is required for this option.

Make sure to login into the registry with the desired user:

```bash
docker login -u <user> <registry>
```

Build and push a multi-architecture image using the command:

```bash
make buildx IMG=<registry>/<organization>/<image>
```

A desired version of **KubeStellar** can be specified by using the optional `KUBESTELLAR_VERSION` argument, such as `KUBESTELLAR_VERSION=v0.3.1`, Default value: the latest release version of **KubeStellar**.

The list of desired target platforms can be specified by using the optional `PLATFORMS` argument. Default value; `PLATFORMS=linux/amd64,linux/arm64,linux/ppc64le`.

## Run the container and access KubeStellar from the host OS

Spin up a local or remote container image with the command:

```bash
make run
export KUBECONFIG=${HOME}/.kcp/admin.kubeconfig
export PATH=$PATH:${HOME}/kcp:${HOME}/kubestellar/bin
```

The following arguments can be used:

- `IMG=[<registry>/<organization>/]<image>` select the container image, the default vakue is the local `kubestellar` image
- `KUBESTELLAR_VERSION=[<version>|latest]`, select the image tag, the default value is the latest version of **KubeStellar**
- `NAME=<container-name>` sets the name of the container, the default value is `kubestellar`

The command above will

- make port `6443` accessibe on the host OS
- share a `~/.kcp` folder containing the `KUBECONFIG` file
- create a `~/kcp` folder containing a copy of **kcp** plugins for the host OS architecture
- create a `~/kubestellar` folder containing a copy of **KubeStellar** binaries for the host OS architecture

At this point, we can check basic **KubeStellar** functionality with the command:

```bash
kubectl ws root
kubectl ws tree
```

which should return something like:

```text
.
└── root
    ├── compute
    └── espw
```

At this point, one can follow the instructions and examples in **KubeStellar** [Quickstart](https://docs.kubestellar.io/release-0.3/Getting-Started/quickstart/).

## Checking the container logs

The running container logs can be easily displayed using the command:

```bash
make logs
```

## Login into the running container to use **KubeStellar** from inside the container

Login using the command:

```bash
make exec
```

Since `kubectl` is available inside the container, at this point **KubeStellar** can be used as before.

## Teardown and cleanup

Teardown and clean up with the command:

```bash
make clean
```

This command will

- force stop and remove the running container image
- remove the container image from the local registry
- delete the created folders