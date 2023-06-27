<!--mailbox-controller-process-start-start-->
```shell
kubectl ws root:espw
cd ../kubestellar
go run ./cmd/mailbox-controller -v=2 &
sleep 45
```
<!--mailbox-controller-process-start-end-->