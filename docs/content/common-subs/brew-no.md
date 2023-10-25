<!--brew-no-start-->
The following commands will (a) download the kcp and KubeStellar executables into subdirectories of your current working directory and (b) deploy (i.e., start and configure) kcp and KubeStellar as workload in the hosting cluster. If you want to suppress the deployment part then add --deploy false to the first command's flags (e.g., after the specification of the KubeStellar version); for the deployment-only part, once the executable have been fetched, see the documentation for the commands about deployment into a Kubernetes cluster.

```
bash <(curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/main/bootstrap/bootstrap-kubestellar.sh) \
    --kubestellar-version {{ config.ks_tag }} --deploy false

export PATH="$PATH:$(pwd)/kcp/bin:$(pwd)/kubestellar/bin"
```
<!--brew-no-end-->