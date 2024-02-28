# Third Parties Integrations

## Install and configure ArgoCD

Install ArgoCD on kind-kubeflex:

```shell
kubectl --context kind-kubeflex create namespace argocd
kubectl --context kind-kubeflex apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
```

Install CLI:

on MacOS:

```shell
brew install argocd
```

on Linux:

```shell
curl -sSL -o argocd-linux-amd64 https://github.com/argoproj/argo-cd/releases/latest/download/argocd-linux-amd64
sudo install -m 555 argocd-linux-amd64 /usr/local/bin/argocd
rm argocd-linux-amd64
```

Check the [ArgoCD releases](https://github.com/argoproj/argo-cd/releases) page for the obtaining the latest 
stable release for other architectures and operating systems.

Configure Argo to work with the ingress installed in the hosting cluster:

```shell
kubectl --context kind-kubeflex apply -f - <<EOF
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: argocd-server-ingress
  namespace: argocd
  annotations:
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
    nginx.ingress.kubernetes.io/ssl-passthrough: "true"
spec:
  ingressClassName: nginx
  rules:
  - host: argocd.localtest.me
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: argocd-server
            port:
              name: https
EOF
```

Open a browser to ArgoCD console:

```shell
open https://argocd.localtest.me:9443
```

Note: if you are working on a VM via SSH, just take the IP of the VM (VM_IP)
and add the line '<VM_IP> argocd.localtest.me' to your '/etc/hosts' file, replacing
<VM_IP> with the actual IP of your desktop.

Get the password for Argo with:

```shell
kubectl config use-context kind-kubeflex
argocd admin initial-password -n argocd
```

Login into the ArgoCD console with `admin` and the password just retrieved. Type
the following on a shell terminal in your desktop (or just enter the address
<https://argocd.localtest.me:9443> on your browser):

```shell
open https://argocd.localtest.me:9443
```

Also, login with the argocd CLI with the same credentials.

```shell
argocd login --insecure argocd.localtest.me:9443
```

Add the `wds1` space as cluster to ArgoCD:

```shell
CONTEXT=wds1
kubectl config view --minify --context=${CONTEXT} --flatten > /tmp/${CONTEXT}.kubeconfig
kubectl config --kubeconfig=/tmp/${CONTEXT}.kubeconfig set-cluster ${CONTEXT}-cluster --server=https://${CONTEXT}.${CONTEXT}-system 2>/dev/null
kubectl config use-context kind-kubeflex
ARGO_SERVER_POD=$(kubectl get pods -n argocd -l app.kubernetes.io/name=argocd-server -o 'jsonpath={.items[0].metadata.name}')
kubectl cp /tmp/${CONTEXT}.kubeconfig -n argocd ${ARGO_SERVER_POD}:/tmp
PASSWORD=$(argocd admin initial-password -n argocd | cut -d " " -f 1)
kubectl exec -it -n argocd $ARGO_SERVER_POD -- argocd login argocd-server.argocd --username admin --password $PASSWORD --insecure
kubectl exec -it -n argocd $ARGO_SERVER_POD -- argocd cluster add ${CONTEXT} --kubeconfig /tmp/${CONTEXT}.kubeconfig -y
```

Configure Argo to label resources with the "argocd.argoproj.io/instance" label:

```shell
kubectl --context kind-kubeflex patch cm -n argocd argocd-cm -p '{"data": {"application.instanceLabelKey": "argocd.argoproj.io/instance"}}'
```
