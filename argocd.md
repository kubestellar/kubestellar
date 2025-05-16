# Testing ArgoCD Integration with KubeStellar

This README provides step-by-step instructions for testing the ArgoCD integration with KubeStellar.

## Prerequisites

- Kubernetes cluster (Kind, k3s, or any other)
- Helm
- kubectl

## Setup Steps

1. **Create a Kind cluster with SSL passthrough**:
   ```bash
   bash <(curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/main/scripts/create-kind-cluster-with-SSL-passthrough.sh) --name kubeflex --port 9443
   ```

2. **Ensure you're using the correct kubectl context**:
   ```bash
   kubectl config use-context kind-kubeflex
   kubectl cluster-info
   ```

3. **Install KubeStellar core chart with ArgoCD enabled**:
   ```bash
   cd ~/open_source/kubestellar
   helm dependency update core-chart
   helm upgrade --install ks-core core-chart --set argocd.install=true
   ```

4. **Create ITS and WDS control planes**:
   ```bash
   helm upgrade ks-core core-chart --set argocd.install=true --set-json='ITSes=[{"name":"its1"}]' --set-json='WDSes=[{"name":"wds1"}]'
   ```

5. **Retrieve ArgoCD admin password**:
   ```bash
   kubectl -n default get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
   ```

6. **Create an ArgoCD application**:
   ```bash
   helm upgrade ks-core core-chart --set argocd.install=true --set-json='ITSes=[{"name":"its1"}]' --set-json='WDSes=[{"name":"wds1"}]' --set-json='argocd.applications=[{"name":"test-app","repoURL":"https://github.com/pdettori/sample-apps.git","path":"nginx","destinationWDS":"wds1","destinationNamespace":"nginx-test"}]'
   ```

7. **Enable auto-sync for the application**:
   ```bash
   kubectl patch application test-app -p '{"spec":{"syncPolicy":{"automated":{"prune":true,"selfHeal":true}}}}' --type=merge
   ```

8. **Check application status**:
   ```bash
   kubectl get applications
   ```

9. **Access ArgoCD UI**:
   ```bash
   kubectl port-forward svc/ks-core-argocd-server 8080:443
   ```
   Then visit http://localhost:8080 in your browser

10. **Alternative UI access**:
    Visit https://argocd.localtest.me:9443 (for Kind or k3s installations)

## Troubleshooting

If you encounter connection errors like:
```
The connection to the server 127.0.0.1:40525 was refused - did you specify the right host or port?
```

Make sure you're using the correct kubectl context:
```bash
kubectl config get-contexts
kubectl config use-context kind-kubeflex
```

## Verification

- Check if ArgoCD pods are running:
  ```bash
  kubectl get pods
  ```

- Check if control planes are ready:
  ```bash
  kubectl get controlplanes
  ```

- Check if ArgoCD has registered the WDS as a target cluster:
  ```bash
  kubectl get secret -n default -l argocd.argoproj.io/secret-type=cluster
  ```
