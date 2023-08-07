# Deploy **KubeStellar** service in a cluster

Table of contests:
- [Deploy **KubeStellar** service in a cluster](#deploy-kubestellar-service-in-a-cluster)
  - [Deploy **KubeStellar** in a **Kubernetes** cluster (**Kind** cluster)](#deploy-kubestellar-in-a-kubernetes-cluster-kind-cluster)
  - [Deploy **KubeStellar** in an **OpenShift** cluster](#deploy-kubestellar-in-an-openshift-cluster)
  - [Access **KubeStellar** service directly from the host OS without `admin.kubeconfig` or executables](#access-kubestellar-service-directly-from-the-host-os-without-adminkubeconfig-or-executables)
  - [Access **KubeStellar** service from the host OS by extracting the `admin.kubeconfig` and the executables from the pod](#access-kubestellar-service-from-the-host-os-by-extracting-the-adminkubeconfig-and-the-executables-from-the-pod)
  - [Access **KubeStellar** from another pod](#access-kubestellar-from-another-pod)
  - [Add a new cluster to **KubeStellar** inventory](#add-a-new-cluster-to-kubestellar-inventory)

## Deploy **KubeStellar** in a **Kubernetes** cluster (**Kind** cluster)

Create a **Kind** cluster with the `extraPortMappings` for port `1024`:

```shell
kind create cluster --config=- <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 443
    hostPort: 1024
    protocol: TCP
EOF
```

Create an `nginx-ingress` with SSL passthrough. Following [Kind NGINX ingress instructions](https://kind.sigs.k8s.io/docs/user/ingress/), we have modified the YAML at https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml to include the `--enable-ssl-passthrough=true` argument. The modified YAML file is here included as [kind-nginx-ingress-with-SSL-passthrough.yaml](./kind-nginx-ingress-with-SSL-passthrough.yaml):

```shell
kubectl apply -f kind-nginx-ingress-with-SSL-passthrough.yaml
```

Wait for the ingress to be ready:

```shell
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=90s
```

Deploy **KubeStellar** `stable` in a `kubestellar` namespace:

```shell
kubectl apply -f kubestellar-server-ingress.yaml
```

Note that two environment variables are available for the **KubeStellar** deployment container:

- `EXTERNAL_HOSTNAME`: set the external domain/url to be used to reach **KubeStellar**. This value will be included in the `admin.kubeconfig`. By default, it is set to the hostname in the ingress rule: `kubestellar.svc.cluster.local`.
- `EXTERNAL_PORT`: set the port to be used to reach **KubeStellar**. This value will be included in the `admin.kubeconfig`. By default, it is set to the ingress port `443`.

The `kubestellar-server` deployment, holds its access kubeconfigs in a `kubestellar` secret in the `kubestellar` namespace, which it manages using a `kubestellar-role`. Additionally, the role allows the pod to get its ingress/route to put it in the `external.kubeconfig`.

As a result of the above command, the following objects will be created:

```text
namespace/kubestellar created
persistentvolumeclaim/kubestellar-pvc created
serviceaccount/kubestellar-service-account created
clusterrole.rbac.authorization.k8s.io/kubestellar-role created
clusterrolebinding.rbac.authorization.k8s.io/kubestellar-role-binding created
deployment.apps/kubestellar-server created
service/kubestellar-service created
ingress.networking.k8s.io/kubestellar-ingress create
```

Wait for **KubeStellar** to be ready:

```shell
kubectl logs -n kubestellar $(kubectl get pod -n kubestellar --selector=app=kubestellar-server -o jsonpath='{.items[0].metadata.name}')
```

```text
< Starting Kubestellar container >-------------------------
< Ensuring the TLS certificate >---------------------------

Notice
------
'init-pki' complete; you may now create a CA or requests.

Your newly created PKI dir is:
* /home/kubestellar/pki

* Using Easy-RSA configuration: Not found

* IMPORTANT: Easy-RSA 'vars' template file has been created in your new PKI.
             Edit this 'vars' file to customise the settings for your PKI.
             To use a global vars file, use global option --vars=<YOUR_VARS>

* Using x509-types directory: /home/kubestellar/easy-rsa/x509-types


* Using Easy-RSA configuration:
  /home/kubestellar/pki/vars

* Using SSL: openssl OpenSSL 3.0.7 1 Nov 2022 (Library: OpenSSL 3.0.7 1 Nov 2022)

Using configuration from /home/kubestellar/pki/1af2369d/temp.5.1
...+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*.+...+...+...+.+.....+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*...+......+......+.........+.+...+..+.+..+.......+...+..+......+...+.......+...+..+............+...+...+.+........................+.....+.......+...+......+......+.........+...........+.........+...+...+....+...+......+............+..+.+............+..+......+.........+...+...+....+..+.+.....+...+.+.....................+...+.........+.....+.............+..+.+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
.......+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*.+.+...+...+..............+.........+.......+..+.+.........+.....+...+...+.+........+...+....+...+...........+......+....+..+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*...+..+.+......+.....+......+....+...........+..................+......+.+.....+...+....+..+.+........+..........+............+..+............+...+...+.............+.....+....+.....+.+.....+......+...+......+....+..+.+............+......+.....+......+......+....+...+..+......+.......+......+.........+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
-----

Notice
------
CA creation complete. Your new CA certificate is at:
* /home/kubestellar/pki/ca.crt


* Using Easy-RSA configuration:
  /home/kubestellar/pki/vars

* Using SSL: openssl OpenSSL 3.0.7 1 Nov 2022 (Library: OpenSSL 3.0.7 1 Nov 2022)

-----

Notice
------
Keypair and certificate request completed. Your files are:
* req: /home/kubestellar/pki/reqs/kcp-server-6dfc23efd88ed0fd0b24d816c29095a91.req
* key: /home/kubestellar/pki/private/kcp-server-6dfc23efd88ed0fd0b24d816c29095a91.key

Using configuration from /home/kubestellar/pki/df54a34a/temp.6.1
Check that the request matches the signature
Signature ok
The Subject's Distinguished Name is as follows
commonName            :ASN.1 12:'kcp-server-6dfc23efd88ed0fd0b24d816c29095a91'
Certificate is to be certified until Dec 22 21:47:26 2050 GMT (10000 days)

Write out database with 1 new entries
Data Base Updated

Notice
------
Certificate created at:
* /home/kubestellar/pki/issued/kcp-server-6dfc23efd88ed0fd0b24d816c29095a91.crt

Notice
------
Inline file created:
* /home/kubestellar/pki/inline/kcp-server-6dfc23efd88ed0fd0b24d816c29095a91.inline

TLS certificates for server kubestellar.svc.cluster.local:
/home/kubestellar/pki/ca.crt
/home/kubestellar/pki/issued/kcp-server-6dfc23efd88ed0fd0b24d816c29095a91.crt
/home/kubestellar/pki/private/kcp-server-6dfc23efd88ed0fd0b24d816c29095a91.key
< Starting kcp >-------------------------------------------
Running kcp with TLS keys... logfile=./kubestellar-logs/kcp.log
Waiting for kcp to be ready... it may take a while
kcp version: v0.11.0
Current workspace is "root".
Switching the admin.kubeconfig domain to kubestellar.svc.cluster.local and port 1024...
< Starting KubeStellar >-----------------------------------
Finished augmenting root:compute for KubeStellar
Workspace "espw" (type root:organization) created. Waiting for it to be ready...
Workspace "espw" (type root:organization) is ready to use.
Current workspace is "root:espw" (type root:organization).
Finished populating the espw with kubestellar apiexports
****************************************
Launching KubeStellar ...
****************************************
 mailbox-controller is running (log file: /home/kubestellar/kubestellar-logs/mailbox-controller-log.txt)
 where-resolver is running (log file: /home/kubestellar/kubestellar-logs/kubestellar-where-resolver-log.txt)
 placement translator is running (log file: /home/kubestellar/kubestellar-logs/placement-translator-log.txt)
****************************************
Finished launching KubeStellar ...
****************************************
Current workspace is "root".
< Create secrets >-----------------------------------------
Ensure secret in the current namespace...
secret/kubestellar created
 Created secret in the current namespace.
Ready!
```

Use the commands below to wait for **KubeStellar** to be ready:


```shell
kubectl wait -n kubestellar \
  --for=condition=ready pod \
  --selector=app=kubestellar-server \
  --timeout=120s

until kubectl logs -n kubestellar $(kubectl get pod -n kubestellar --selector=app=kubestellar-server -o jsonpath='{.items[0].metadata.name})' 2>&1 | grep -Fxq "Ready!"; do
  sleep 1
done
```

Note that, alternatively, one can wait for the `kubestellar` secret to be created.

After the deployment has completed, **KubeStellar** `admin.kubeconfig` can be obtained in two ways:

- the `kubestellar` secret in the `kubestellar` namespace;
- directly from the `kubestellar` pod in the `kubestellar` namespace at the location `/home/kubestellar/.kcp/external.kubeconfig`.

## Deploy **KubeStellar** in an **OpenShift** cluster

In this case use the following YAML to deploy  **KubeStellar** `stable` in a `kubestellar` namespace:

```shell
kubectl apply -f kubestellar-server-route.yaml
```

Then follow the same instructions for the Kind cluster deployment.

## Access **KubeStellar** service directly from the host OS without `admin.kubeconfig` or executables

Since **kubectl**, **kcp** plugins, and **KubeStellar** executables are include in the **KubeStellar** container image we can operate KubeStellar directly from the host OS using `kubectl`, for example:

```shell
$ kubectl exec -n kubestellar $(kubectl get pod -n kubestellar --selector=app=kubestellar-server -o jsonpath='{.items[0].metadata.name}') -- kubectl ws tree
.
└── root
    ├── compute
    └── espw

$ kubectl exec -n kubestellar $(kubectl get pod -n kubestellar --selector=app=kubestellar-server -o jsonpath='{.items[0].metadata.name}') -- kubectl ws create imw
Workspace "imw" (type root:organization) created. Waiting for it to be ready...
Workspace "imw" (type root:organization) is ready to use.
```

## Access **KubeStellar** service from the host OS by extracting the `admin.kubeconfig` and the executables from the pod

In this case the host OS will need a copy of **kcp** plugins, and **KubeStellar** plugins, assuming that the architecture of the client is the same of the container, the executables can be extracted from the container:

```shell
kubectl exec -n kubestellar $(kubectl get pod -n kubestellar --selector=app=kubestellar-server -o jsonpath='{.items[0].metadata.name}') -- tar cf - "/home/kubestellar/kcp-plugins" | tar xf - --strip-components=2

kubectl exec -n kubestellar $(kubectl get pod -n kubestellar --selector=app=kubestellar-server -o jsonpath='{.items[0].metadata.name}') -- tar cf - "/home/kubestellar/kubestellar" | tar xf - --strip-components=2
```

As an alternative the bootstrap script can be used with the `--deploy false` option as shown below. This alternative option does not require the architecture of the host OS to match the one inside the container. In this use case, it may be necessary to specify the version of **KubeStaller** with the `--kubestellar-version` argument to match the version deployed in the cluster.

```shell
bash <(curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/main/bootstrap/bootstrap-kubestellar.sh) --deploy false
```

The `external.kubeconfig` to access **KubeStellar** from outside the cluster can be obtained by:

```shell
kubectl cp -n kubestellar $(kubectl get pod -n kubestellar --selector=app=kubestellar-server -o jsonpath='{.items[0].metadata.name}'):/home/kubestellar/.kcp/external.kubeconfig ./admin.kubeconfig
```

or

```shell
kubectl get secrets kubestellar -n kubestellar -o 'go-template={{index .data "external.kubeconfig"}}' | base64 --decode > ./admin.kubeconfig
```

Then, the following environment variables need to be set:

```shell
export KUBECONFIG=$PWD/admin.kubeconfig
export PATH=$PATH:$PWD/kcp-plugins/bin:$PWD/kubestellar/bin
```

If the `EXTERNAL_HOSTNAME` value given or defaulted by the ingress/route is not resolved by DNS, *e.g.*  if it defaulted to kubestellar.svc.cluster.local or you specified a domain name that you just made up, then you need to get its address resolution into the `/etc/hosts` files of the machines where you want to run clients. Following is an example of something you could inject into /etc/hosts on the machine running your kind cluster (if that is what you are doing).

```text
127.0.0.1       kubestellar.svc.cluster.local
```

Now we can use use **KubeStellar** and **kcp** in the usual way:

```shell
$ kubectl ws tree
.
└── root
    ├── compute
    └── espw
```

## Access **KubeStellar** from another pod

A pod can access **KubeStellar** by retrieving the `admin.kubeconfig` via one of the following ways:

- the `kubestellar` secret in the `kubestellar` namespace;
- directly from the `kubestellar` pod in the `kubestellar` namespace at the location `/home/kubestellar/.kcp/external.kubeconfig`.

Obviously `kubectl`, **kcp** plugins, and **KubeStellar** executables are also needed.

In this example, we create a `kubestellar-client` pod based on the same image of `kubestellar-server` since it already contains the required executables mentioned above:

```shell
$ kubectl apply -f kubestellar-client.yaml
deployment.apps/kubestellar-client created

$ kubectl get pods
NAME                                  READY   STATUS    RESTARTS   AGE
kubestellar-client-5dcd55b4c7-cvz6j   1/1     Running   0          32s
kubestellar-server-566f5cb54d-mmp8p   1/1     Running   0          58m
```

Now, let us log into the pod:

```shell
kubectl exec -n kubestellar $(kubectl get pod -n kubestellar --selector=app=kubestellar-client -o jsonpath='{.items[0].metadata.name}') -it -- bash
```

From within the pod:

```shell
[kubestellar@kubestellar-client-7c7d46cf77-sk2d4 /]$ kubectl ws tree
.
└── root
    ├── compute
    └── espw
```

## Add a new cluster to **KubeStellar** inventory

A syncer can also be generated using pipe commands as in the following example:

```shell
kubectl kubestellar prep-for-cluster --silent --imw root:example-imw some-cluster env=prod -o - 2> /dev/null 1> syncer.yaml
```

