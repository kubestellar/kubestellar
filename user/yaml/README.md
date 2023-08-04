# Deploy **KubeStellar** service in a Kind cluster

Table of contests:
- [Deploy **KubeStellar** service in a Kind cluster](#deploy-kubestellar-service-in-a-kind-cluster)
  - [Deploy **KubeStellar** in a Kind cluster](#deploy-kubestellar-in-a-kind-cluster)
  - [Access **KubeStellar** service directly from the host OS without `admin.kubeconfig` or plugins](#access-kubestellar-service-directly-from-the-host-os-without-adminkubeconfig-or-plugins)
  - [Access **KubeStellar** service from the host OS by extracting the `admin.kubeconfig` and the plugins from the pod](#access-kubestellar-service-from-the-host-os-by-extracting-the-adminkubeconfig-and-the-plugins-from-the-pod)
  - [Access **KubeStellar** from another pod in the same `kubestellar` namespace](#access-kubestellar-from-another-pod-in-the-same-kubestellar-namespace)

## Deploy **KubeStellar** in a Kind cluster

Create a **Kind** cluster with the `extraPortMappings` for ports `80` and `443`:

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
  - containerPort: 80
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
EOF
```

Create an `nginx-ingress` with SSL passthrough using the YAML file [kind-nginx-ingress-with-SSL-passthrough.yaml](./kind-nginx-ingress-with-SSL-passthrough.yaml):

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

Note that three environment variables are available for the **KubeStellar** deployment container:

- `EXTERNAL_HOSTNAME`: set the external domain/url to be used to reach **KubeStellar**. This value will be included in the `admin.kubeconfig`. By default, it is set to the hostname in the ingress rule: `kubestellar.svc.cluster.local`.
- `EXTERNAL_PORT`: set the port to be used to reach **KubeStellar**. This value will be included in the `admin.kubeconfig`. By default, it is set to the ingress port `443`.

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
kubectl logs -n kubestellar $(kubectl get pod -n kubestellar --selector=app=kubestellar-server -o jsonpath={.items[0].metadata.name})
```

```text
< Starting Kubestellar container >-------------------------
< Creating the TLS certificate >---------------------------

Notice
------
'init-pki' complete; you may now create a CA or requests.

Your newly created PKI dir is:
* /kubestellar/pki

* Using Easy-RSA configuration: Not found

* IMPORTANT: Easy-RSA 'vars' template file has been created in your new PKI.
             Edit this 'vars' file to customise the settings for your PKI.
             To use a global vars file, use global option --vars=<YOUR_VARS>

* Using x509-types directory: /kubestellar/easy-rsa/x509-types


* Using Easy-RSA configuration:
  /kubestellar/pki/vars

* Using SSL: openssl OpenSSL 3.0.7 1 Nov 2022 (Library: OpenSSL 3.0.7 1 Nov 2022)

Using configuration from /kubestellar/pki/69afc467/temp.5.1
....+...+..+....+.....+.+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*......+..+...+...+......+...+.......+.....+.+........+....+..+...+.+.........+...........+.........+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*.+.....+.+.....+.+..............+.+.....+.......+..+......+............+.......+...............+.........+............+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
.+.+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*...+.........+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*.........+.....+..........+..+....+......+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
-----

Notice
------
CA creation complete. Your new CA certificate is at:
* /kubestellar/pki/ca.crt


* Using Easy-RSA configuration:
  /kubestellar/pki/vars

* Using SSL: openssl OpenSSL 3.0.7 1 Nov 2022 (Library: OpenSSL 3.0.7 1 Nov 2022)

-----

Notice
------
Keypair and certificate request completed. Your files are:
* req: /kubestellar/pki/reqs/kcp-server.req
* key: /kubestellar/pki/private/kcp-server.key

Using configuration from /kubestellar/pki/0525cd36/temp.5.1
Check that the request matches the signature
Signature ok
The Subject's Distinguished Name is as follows
commonName            :ASN.1 12:'kcp-server'
Certificate is to be certified until Dec 16 21:39:38 2050 GMT (10000 days)

Write out database with 1 new entries
Data Base Updated

Notice
------
Certificate created at:
* /kubestellar/pki/issued/kcp-server.crt

Notice
------
Inline file created:
* /kubestellar/pki/inline/kcp-server.inline

Server=kubestellar.svc.cluster.local
/kubestellar/pki/ca.crt
/kubestellar/pki/issued/kcp-server.crt
/kubestellar/pki/private/kcp-server.key
< Starting kcp >-------------------------------------------
Running kcp... logfile=./kubestellar-logs/kcp.log
Waiting for kcp to be ready... it may take a while
kcp version: v0.11.0
Current workspace is "root".
Switching the admin.kubeconfig domain to kubestellar.svc.cluster.local...
< Starting KubeStellar >-----------------------------------
Finished augmenting root:compute for KubeStellar
Workspace "espw" (type root:organization) created. Waiting for it to be ready...
Workspace "espw" (type root:organization) is ready to use.
Current workspace is "root:espw" (type root:organization).
Finished populating the espw with kubestellar apiexports
****************************************
Launching KubeStellar ...
****************************************
 mailbox-controller is running (log file: /kubestellar/kubestellar-logs/mailbox-controller-log.txt)
 where-resolver is running (log file: /kubestellar/kubestellar-logs/kubestellar-where-resolver-log.txt)
 placement translator is running (log file: /kubestellar/kubestellar-logs/placement-translator-log.txt)
****************************************
Finished launching KubeStellar ...
****************************************
Current workspace is "root".
< Create secrets >-----------------------------------------
Ensure secret in the curent namespace...
secret/kubestellar created
Ensure secret in namespace "default"...
secret "kubestellar" deleted
secret/kubestellar created
Ready!
```

Note that alternatively, one can wait for the `kubestellar` secret to be created.

```shell
until kubectl logs -n kubestellar $(kubectl get pod -n kubestellar --selector=app=kubestellar-server -o jsonpath={.items[0].metadata.name}) 2>&1 | grep -Fxq "Ready!"
do
  sleep 1
done
```

After the deployment has completed, **KubeStellar** `admin.kubeconfig` can be in three ways:

- the `kubestellar` secret in the `kubestellar` namespace or any other namespace listed in `SECRET_NAMESPACES` environment variable;
- directly from the `kubestellar` pod in the `kubestellar` namespace at the location `/home/kubestellar/.kcp-kubestellar.svc.cluster.local/admin.kubeconfig`.

## Access **KubeStellar** service directly from the host OS without `admin.kubeconfig` or plugins

Since **kubectl**, **kcp** plugins, and **KubeStellar** executables are include in the **KubeStellar** container image we can operate KubeStellar directly from the host OS using `kubectl`, for example:

```shell
$ kubectl exec -n kubestellar $(kubectl get pod -n kubestellar --selector=app=kubestellar-server -o jsonpath={.items[0].metadata.name}) -- kubectl ws tree
.
└── root
    ├── compute
    └── espw

$ kubectl exec -n kubestellar $(kubectl get pod -n kubestellar --selector=app=kubestellar-server -o jsonpath={.items[0].metadata.name}) -- kubectl ws create imw
Workspace "imw" (type root:organization) created. Waiting for it to be ready...
Workspace "imw" (type root:organization) is ready to use.
```

## Access **KubeStellar** service from the host OS by extracting the `admin.kubeconfig` and the plugins from the pod

In this case the host OS will need a copy of **kcp** `admin.kubeconfig`, **kcp** plugins, and **KubeStellar** plugins:

```shell
kubectl exec -n kubestellar $(kubectl get pod -n kubestellar --selector=app=kubestellar-server -o jsonpath={.items[0].metadata.name}) -- tar cf - "/home/kubestellar/kcp-plugins" | tar xf - --strip-components=2

kubectl exec -n kubestellar $(kubectl get pod -n kubestellar --selector=app=kubestellar-server -o jsonpath={.items[0].metadata.name}) -- tar cf - "/home/kubestellar/kubestellar" | tar xf - --strip-components=2

kubectl cp -n kubestellar $(kubectl get pod -n kubestellar --selector=app=kubestellar-server -o jsonpath={.items[0].metadata.name}):/home/kubestellar/.kcp/external.kubeconfig ./admin.kubeconfig

export KUBECONFIG=$PWD/admin.kubeconfig
export PATH=$PATH:$PWD/kcp-plugins/bin:$PWD/kubestellar/bin
```

Then add the **KubeStellar** ingress `kubestellar.svc.cluster.local` to the `/etc/hosts` files:

```text
127.0.0.1       kubestellar.svc.cluster.local
```

Now we can use use **KubeStellar** in the usual way:

```shell
$ kubectl ws tree
.
└── root
    ├── compute
    └── espw
```

## Access **KubeStellar** from another pod in the same `kubestellar` namespace

Any pod in the same namespace can access **KubeStellar** by retrieving the `admin.kubeconfig` via one of the following ways:

- the `kubestellar` secret in the `kubestellar` namespace or any other namespace listed in `SECRET_NAMESPACES` environment variable;
- directly from the `kubestellar` pod in the `kubestellar` namespace at the location `/home/kubestellar/.kcp-kubestellar.svc.cluster.local/admin.kubeconfig`.

Obviously `kubectl`, **kcp** plugins, and **KubeStellar** executables are also needed.

In this example, we create a `kubestellar-client` pod based on the same image of `kubestellar-server` since it already contains the required executables listed above:

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
kubectl exec -n kubestellar $(kubectl get pod -n kubestellar --selector=app=kubestellar-client -o jsonpath={.items[0].metadata.name}) -it -- bash
```

From within the pod:

```shell
[kubestellar@kubestellar-client-7c7d46cf77-sk2d4 /]$ kubectl ws tree
.
└── root
    ├── compute
    └── espw
```
