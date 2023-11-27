<!--pre-position-syncer-image-start-->
Install the syncer container image in the two WECs.

```shell
kind load docker-image ko.local/syncer:test --name ks-edge-cluster1
kind load docker-image ko.local/syncer:test --name ks-edge-cluster2
```
<!--pre-position-syncer-image-end-->
