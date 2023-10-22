<!--check-kubestellar-helm-deployment-running-start-->
```shell
echo -n 'Waiting for KubeStellar to be ready'
while ! kubectl exec $(kubectl get pod --selector=app=kubestellar -o jsonpath='{.items[0].metadata.name}' -n kubestellar) -n kubestellar -c init -- ls /home/kubestellar/ready &> /dev/null; do
   sleep 10
   echo -n "."
done

echo "\n\nKubeStellar is now ready to take requests"
```
<!--check-kubestellar-helm-deployment-running-end-->