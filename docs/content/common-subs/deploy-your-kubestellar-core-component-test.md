<!--deploy-your-kubestellar-core-component-test-start-->
!!! tip ""
    === "deploy"
         deploy the KubeStellar Core components on the **ks-core** Kind cluster you created in the pre-req section above  

         {%
           include-markdown "install-helm-test.md"
           start="<!--install-helm-test-start-->"
           end="<!--install-helm-test-end-->"
         %}

        **important:** You must add 'kubestellar.core' to your /etc/hosts file with the local network IP address (e.g., 192.168.x.y) where your **ks-core** Kind cluster is running. **DO NOT** use `127.0.0.1` because the ks-edge-cluster1 and ks-edge-cluster2 kind clusters map `127.0.0.1` to their local kubernetes cluster, **not** the ks-core kind cluster.

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
<!--deploy-your-kubestellar-core-component-test-end-->