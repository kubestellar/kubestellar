<!--mailbox-controller-process-start-start-->
```shell
kubectl ws root:espw
cd ../KubeStellar
go run ./cmd/mailbox-controller -v=2 &
sleep 15  # wait a few seconds for the mailbox controller to initialize
```
<!--mailbox-controller-process-start-end-->