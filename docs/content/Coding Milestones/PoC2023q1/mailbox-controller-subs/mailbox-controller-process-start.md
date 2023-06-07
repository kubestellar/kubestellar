<!--mailbox-controller-process-start-start-->
```shell
kubectl ws root:espw
cd ../KubeStellar
go run ./cmd/mailbox-controller -v=2 &
sleep 45
```
<!--mailbox-controller-process-start-end-->