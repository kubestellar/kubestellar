<!--mailbox-controller-process-start-without-cd-kubestellar-start-->
```shell
kubectl ws root:espw
echo "mailbox-controll-start-without"
kubectl api-resources
go run ./cmd/mailbox-controller -v=2 &
sleep 240
```
<!--mailbox-controller-process-start-without-cd-kubestellar-end-->