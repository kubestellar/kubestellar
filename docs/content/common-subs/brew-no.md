<!--brew-no-start-->
The following commands will (a) download the kcp and KubeStellar executables into subdirectories of your current working directory

```
bash <(curl -s https://raw.githubusercontent.com/kubestellar/kubestellar/main/bootstrap/bootstrap-kubestellar.sh) \
    --kubestellar-version {{ config.ks_tag }} --deploy false

export PATH="$PATH:$(pwd)/kcp/bin:$(pwd)/kubestellar/bin"
```
<!--brew-no-end-->