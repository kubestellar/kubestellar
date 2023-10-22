<!--deploy-your-kubestellar-core-component-openshift-start-->
!!! tip ""
    === "deploy"
         {%
           include-markdown "install-helm-openshift.md"
           start="<!--install-helm-openshift-start-->"
           end="<!--install-helm-openshift-end-->"
         %}

        run the following to wait for KubeStellar to be ready to take requests:
        {%
           include-markdown "check-kubestellar-helm-deployment-running.md"
           start="<!--check-kubestellar-helm-deployment-running-start-->"
           end="<!--check-kubestellar-helm-deployment-running-end-->"
         %}

         
    === "uh oh, error?"
         Checking the initialization log to see if there are any obvious errors:
         ```
         KUBECONFIG=~/.kube/config kubectl config use-context ks-core  
         kubectl logs \
           $(kubectl get pod --selector=app=kubestellar \
           -o jsonpath='{.items[0].metadata.name}' -n kubestellar) \
           -n kubestellar -c init
         ```
         if there is nothing obvious, [open a bug report and we can help you out](https://github.com/kubestellar/kubestellar/issues/new?assignees=&labels=kind%2Fbug&projects=&template=bug_report.yaml&title=bug%3A+)
    
    === "open a bug report"
        Stuck? [open a bug report and we can help you out](https://github.com/kubestellar/kubestellar/issues/new?assignees=&labels=kind%2Fbug&projects=&template=bug_report.yaml&title=bug%3A+)
<!--deploy-your-kubestellar-core-component-openshift-end-->