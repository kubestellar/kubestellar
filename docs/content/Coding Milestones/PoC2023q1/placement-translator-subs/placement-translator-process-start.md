<!--placement-translator-process-start-start-->
```shell
kubectl ws root:espw
cd ../KubeStellar
go run ./cmd/placement-translator &
sleep 15  # wait a few seconds for the placement translator to initialize
```
<!--placement-translator-process-start-end-->