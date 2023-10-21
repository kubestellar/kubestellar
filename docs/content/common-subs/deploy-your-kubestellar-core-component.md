<!--deploy-your-kubestellar-core-component-start-->
!!! tip ""
    === "deploy"
         ```shell
         KUBECONFIG=~/.kube/config kubectl config use-context kind-ks-core  
         ```
         {%
           include-markdown "install-helm.md"
           start="<!--install-helm-start-->"
           end="<!--install-helm-end-->"
         %}

        run the following to wait for KubeStellar to be ready to take requests:
         ```shell
         echo -n 'Waiting for KubeStellar to be ready'
         while ! kubectl exec $(kubectl get pod --selector=app=kubestellar -o jsonpath='{.items[0].metadata.name}' -n kubestellar) -n kubestellar -c init -- ls /home/kubestellar/ready &> /dev/null; do
            sleep 10
            echo -n "."
         done

         echo "\n\nKubeStellar is now ready to take requests"
         ```
    === "uh oh, error?"
         Checking the initialization log to see if there are any obvious errors:
         ```
         KUBECONFIG=~/.kube/config kubectl config use-context kind-ks-core  
         kubectl logs \
           $(kubectl get pod --selector=app=kubestellar \
           -o jsonpath='{.items[0].metadata.name}' -n kubestellar) \
           -n kubestellar -c init
         ```
         if there is nothing obvious, [open a bug report and we can help you out](https://github.com/kubestellar/kubestellar/issues/new?assignees=&labels=kind%2Fbug&projects=&template=bug_report.yaml&title=bug%3A+)
    
    === "open a bug report"
        Stuck? [open a bug report and we can help you out](https://github.com/kubestellar/kubestellar/issues/new?assignees=&labels=kind%2Fbug&projects=&template=bug_report.yaml&title=bug%3A+)
<!--deploy-your-kubestellar-core-component-end-->