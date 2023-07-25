<!--mailbox-controller-process-start-without-cd-kubestellar-start-->
```shell
kubectl ws root:espw
echo "mailbox-controller-start-without"
kubectl api-resources
go run ./cmd/mailbox-controller -v=2 &
sleep 90
```
<!--mailbox-controller-process-start-without-cd-kubestellar-end-->