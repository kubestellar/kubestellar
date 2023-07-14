# Building KubeStellar deployment as a container

Table of contents:
- [Building KubeStellar deployment as a container](#building-kubestellar-deployment-as-a-container)
  - [Install the pre-requisites](#install-the-pre-requisites)
  - [Build a local container image](#build-a-local-container-image)
  - [Build and push a multi-architecture image](#build-and-push-a-multi-architecture-image)
  - [Run the container and access KubeStellar from the host OS](#run-the-container-and-access-kubestellar-from-the-host-os)
  - [Checking the container logs](#checking-the-container-logs)
  - [Login into the running container to use **KubeStellar** from inside the container](#login-into-the-running-container-to-use-kubestellar-from-inside-the-container)
  - [Teardown and cleanup](#teardown-and-cleanup)

Here, we discuss the creation and use of a **KubeStellar** container image for running the **kcp** server and the **KubeStellar** controllers, and delivering the **kcp** and **KubeStellar** executables, and kcp client kube config to the host. This provides an alternative to using the **KubeStellar** bootstrap script. Additionally, we are planning to use this images in the future to run **KubeStellar** as a service in a cluster.

## Install the pre-requisites

Install `make` for automating simple tasks such as building/running/cleaning up the container images. Several examples are shown below. On Debian systems, execute this command:

```bash
sudo apt install -y make
```

Install `docker` or `podman` for building and running the container images. It should be noted that the multi-arch `buildx` option is not available for `podman`. On Debian systems, execute these commands for `docker` installation:

```bash
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
sudo usermod -aG docker $USER
```

or these commands for `podman` installation:

```bash
sudo apt install -y podman
alias docker=podman
```

## Build a local container image

Build a local container image `kubestellar` for the host architecture using the **stable** release of **KubeStellar** with the command:

```bash
make build
```

Several arguments can be used to customize the build:

- `IMG=<image-name>` is used to specify the name the image, *e.g.* `quay.io/kubestellar/kubestellar`. Default value is `kubestellar`.
- `KUBESTELLAR_VERSION=<version>` is used to determine the version of **KubeStellar** being installed in the container image, *e.g.* `v0.3.1`. In the argument is not specified, it will default to the **stable** **KubeStellar** release version specified in https://github.com/kubestellar/kubestellar/blob/main/VERSION.
- `TAG=<tag>` is used to determine the tag of the container image, *e.g.* `stable`. In the argument is not specified, it will default to the value set by `KUBESTELLAR_VERSION`.

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

Several arguments can be used to customize the build:

- `KUBESTELLAR_VERSION=<version>` is used to determine the version of **KubeStellar** being installed in the container image, *e.g.* `v0.3.1`. In the argument is not specified, it will default to the **stable** **KubeStellar** release version specified in https://github.com/kubestellar/kubestellar/blob/main/VERSION.
- `TAG=<tag>` is used to determine the tag of the container image, *e.g.* `stable`. In the argument is not specified, it will default to the value set by `KUBESTELLAR_VERSION`.
- `PLATFORMS` lists the desired target platforms. Default value; `PLATFORMS=linux/amd64,linux/arm64,linux/ppc64le`. It should be noted that a `linux/s390x` **KubeStellar**  cannot be build because **kcp** does not have a `linux/s390x` binary release.

As an example, the **KubeStellar** container images available at https://quay.io/repository/kubestellar/kubestellar?tab=tags have been built with commands like:

```bash
make buildx IMG=quay.io/kubestellar/kubestellar KUBESTELLAR_VERSION=v0.4.0
```

## Run the container and access KubeStellar from the host OS

Spin up a local or remote container image with the command:

```bash
make run
export KUBECONFIG=${HOME}/.kcp/admin.kubeconfig
export PATH=$PATH:${HOME}/kcp:${HOME}/kubestellar/bin
```

The following arguments can be used:

- `IMG=[<registry>/<organization>/]<image>` select the container image, the default value is the local `kubestellar` image. Use `IMG=quay.io/kubestellar/kubestellar` to pull the **stable** container image from our [quay.io registry](https://quay.io/repository/kubestellar/kubestellar?tab=tags).
- `TAG=<tag>`, select the image tag, the default value is the **stable** version of **KubeStellar** as specified in https://github.com/kubestellar/kubestellar/blob/main/VERSION.
- `NAME=<container-name>` sets the name of the container, the default value is `kubestellar`
- `BASEPATH=<path>` set the base path where the **KUBECONFIG** and executable will be made available. By default, `~` is used. Note that this parameter will impact the exact path of the `export` commands above, which are provided for the default case. The `make run` command will print out the correct export commands to be used.

The command above will:

- share a `${BASEPATH}/.kcp` folder containing the `KUBECONFIG` file
- share a `${BASEPATH}/.kubestellar-logs` folder containing the log files of **kcp** and **KubeStellar** controllers
- create a `${BASEPATH}/kcp` folder containing a copy of **kcp** plugins for the host OS architecture
- create a `${BASEPATH}/kubestellar` folder containing a copy of **KubeStellar** executables for the host OS architecture

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
make logs [NAME=<name>]
```

The `NAME` argument above is only necessary if a container name was specified with a similar argument in the `make run` command.

## Login into the running container to use **KubeStellar** from inside the container

Login using the command:

```bash
make exec [NAME=<name>]
```

Since `kubectl` is available inside the container, at this point **KubeStellar** can be used as before.

The `NAME` argument above is only necessary if a container name was specified with a similar argument in the `make run` command.

## Teardown and cleanup

Teardown and clean up with the command:

```bash
make clean [IMG=<image>] [TAG=<tag>] [NAME=<name>] [BASEPATH=<path>]
```

This command will

- force stop and remove the running container image
- remove the container image from the local registry
- delete the created folders

The optional arguments above are only necessary if corresponding arguments were used by the `make run` command to alter their default value.
