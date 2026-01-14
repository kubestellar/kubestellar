# Debugging Kubeflex

## Useful Debugging Hacks

### How to open a psql command in-cluster

```shell
kubectl run -i --tty --rm debug -n kubeflex-system --image=postgres --restart=Never -- bash
psql -h mypsql-postgresql.kubeflex-system.svc -U postgres
```

### Forcing image update

This is useful when using a new image with same tag (not a best practice but may be useful for testing)

```shell
kubectl --context kind-kubeflex patch deployment kubeflex-controller-manager --namespace kubeflex-system --type='json' -p='[{"op": "replace", "path": "/spec/template/spec/containers/1/imagePullPolicy", "value": "Always"}]'
sleep 5
kubectl --context kind-kubeflex patch deployment kubeflex-controller-manager --namespace kubeflex-system --type='json' -p='[{"op": "replace", "path": "/spec/template/spec/containers/1/imagePullPolicy", "value": "IfNotPresent"}]'
```

### Installing the helm chart from the local repo

```shell
helm upgrade --install kubeflex-operator chart --namespace kubeflex-system --set domain=localtest.me --set externalPort=9443
```

### How to view certs info

```shell
openssl x509 -noout -text -in certs/apiserver.crt 
```

### Manually creating the configmap with KubeFlex defaults

```shell
kubectl create configmap kubeflex-config -n kubeflex-system --from-literal=externalPort=9443 --from-literal=domain=localtest.me
```

### Get decoded value from secret

```shell
NAMESPACE= # your namespace
NAME= # your secret name
kubectl get secrets -n ${NAMESPACE} ${NAME} -o jsonpath='{.data.apiserver\.crt}' | base64 -d
```

### How to attach a ephemeral container to debug

```shell
NAMESPACE= # your namespace
NAME= # pod name
CONTAINER= # container name
IMAGE=ubuntu:latest
kubectl debug -n ${NAMESPACE} -it ${NAME} --image=${IMAGE} --target=${CONTAINER} -- bash
# e.g. kubectl debug -n ${NAMESPACE} -it ${NAME} --image=curlimages/curl:8.1.2 --target=${CONTAINER} -- sh
```

### Getting all the command args for a process

```shell
cat /proc//cmdline | sed -e "s/\x00/ /g"; echo
```

### How to communicate between kind clusters on the same node

One approach that is independent of local machine IP is to use the internal DNS address of
docker containers. The address is the name of the docker container. For kubflex that
address is `kubeflex-control-plane`. For example, if I have a nodeport service on 
`kubeflex-control-plane` with port 30080 and I want to access it from another kind cluster
the internal address to use is `https://kubeflex-control-plane:30080`


