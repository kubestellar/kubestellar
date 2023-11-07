<!--build-syncer-image-start-->
Build the syncer image.

```shell
export SYNCER_IMG_REF=$(
  if (docker info | grep podman) &> /dev/null
  then export DOCKER_HOST=unix://${HOME}/.local/share/containers/podman/machine/qemu/podman.sock
  fi
  KO_DOCKER_REPO="" make build-kubestellar-syncer-image-local | grep -v "make.*directory" )
echo "Locally built syncer image reference is <$SYNCER_IMG_REF>; adding alias ko.local/syncer:test"
docker tag "$SYNCER_IMG_REF" ko.local/syncer:test
```
<!--build-syncer-image-end-->
