<!--mailbox-controller-process-start-without-cd-kubestellar-start-->
```shell
kubectl ws root:espw
go run ./cmd/mailbox-controller -v=2 &
sleep 45
```
<!--mailbox-controller-process-start-without-cd-kubestellar-end-->