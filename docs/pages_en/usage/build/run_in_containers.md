---
title: Run in containers
permalink: usage/build/run_in_containers.html
---

### Pre-built images

<!-- reference https://werf.io/documentation/v1.2/advanced/ci_cd/run_in_container/how_it_works.html and https://werf.io/documentation/v1.2/advanced/ci_cd/run_in_container/use_docker_container.html -->

The recommended way to run werf in containers is to use the official werf images from `registry.werf.io/werf/werf`. You can launch werf in a container as follows:

```shell
docker run \
    --security-opt seccomp=unconfined --security-opt apparmor=unconfined \
    registry.werf.io/werf/werf:1.2-stable WERF_COMMAND
```

If the Linux kernel does not support rootless OverlayFS, then fuse-overlayfs should be used:

```shell
docker run \
    --device /dev/fuse \
    --security-opt seccomp=unconfined --security-opt apparmor=unconfined \
    registry.werf.io/werf/werf:1.2-stable WERF_COMMAND
```

It is recommended to use the `registry.werf.io/werf/werf:1.2-stable` image. The following images are also available:
- `registry.werf.io/werf/werf:latest` — an alias for `registry.werf.io/werf/werf:1.2-stable`;
- `registry.werf.io/werf/werf:1.2-{alpha|beta|ea|stable|rock-solid}` — all images are based on Alpine, the user must choose the release (stability) channel;
- `registry.werf.io/werf/werf:1.2-{alpha|beta|ea|stable|rock-solid}-{alpine|ubuntu|fedora}` — the user must choose the release (stability) channel and the the underlying distribution.

These images use the Buildah backend. Switching to the Docker Server backend is not supported.

### Building images using the Docker server (not recommended)

> **NOTE:** No official werf images are provided for this method.

To use a local Docker server, you have to mount the Docker socket in a container and use privileged mode:

```shell
docker run \
    --privileged \
    --volume $HOME/.werf:/root/.werf \
    --volume /tmp:/tmp \
    --volume /var/run/docker.sock:/var/run/docker.sock \
    IMAGE WERF_COMMAND
```

— this method supports building Dockerfile images or Stapel images.

Use the following command to run the Docker server over TCP:

```shell
docker run --env DOCKER_HOST="tcp://HOST:PORT" IMAGE WERF_COMMAND
```

— this method only supports building Dockerfile images. Stapel images are not supported because the Stapel image builder mounts host system directories into the build containers.

### Troubleshooting

<!-- reference https://werf.io/documentation/v1.2/advanced/ci_cd/run_in_container/how_it_works.html#troubleshooting -->

#### `fuse: device not found`

The complete error message may look as follows:

```
error mounting new container: error mounting build container "53f916e7a334a4bb0d9dbc38a0901718d40b99765002bb7f2f2e5464b1db4294": error creating overlay mount to /home/build/.local/share/containers/storage/overlay/49e856f537ba58afdc09137291133994cd1305e40df72c4fab43077cbd405477/merged, mount_data=",lowerdir=/home/build/.local/share/containers/storage/overlay/l/Z5GEVIFIIQ7H262DYUTX3YOVR6:/home/build/.local/share/containers/storage/overlay/l/PJBBW6UNUNGI37IX6R3LDNPX3J:/home/build/.local/share/containers/storage/overlay/l/MUYSUONLQVE4CJMQVDCH2UBAVQ:/home/build/.local/share/containers/storage/overlay/l/67JHKJDCKBTI4R3Q5S5YG44AD3:/home/build/.local/share/containers/storage/overlay/l/3S72G4SWKDXILGANUOCESP5LDK,upperdir=/home/build/.local/share/containers/storage/overlay/49e856f537ba58afdc09137291133994cd1305e40df72c4fab43077cbd405477/diff,workdir=/home/build/.local/share/containers/storage/overlay/49e856f537ba58afdc09137291133994cd1305e40df72c4fab43077cbd405477/work,volatile": using mount program /usr/bin/fuse-overlayfs: fuse: device not found, try 'modprobe fuse' first
fuse-overlayfs: cannot mount: No such file or directory
: exit status 1
time="2021-12-06T11:30:20Z" level=error msg="exit status 1"
```

How to fix it: enable the fuse device for the container in which werf is running ([details]({{ "usage/build/run_in_containers.html" | true_relative_url }})).

#### `flags: 0x1000: permission denied`

The complete error message may look as follows:

```
time="2021-12-06T11:23:23Z" level=debug msg="unable to create kernel-style whiteout: operation not permitted"
time="2021-12-06T11:23:23Z" level=debug msg="[graphdriver] trying provided driver \"overlay\""
time="2021-12-06T11:23:23Z" level=debug msg="overlay: mount_program=/usr/bin/fuse-overlayfs"
Running time 0.01 seconds
Error: unable to get buildah client: unable to create new Buildah instance with mode "native-rootless": unable to get storage: mount /home/build/.local/share/containers/storage/overlay:/home/build/.local/share/containers/storage/overlay, flags: 0x1000: permission denied
time="2021-12-06T11:23:23Z" level=error msg="exit status 1"
```

How to fix it: disable AppArmor and seccomp profiles with `--security-opt seccomp=unconfined` and `--security-opt apparmor=unconfined` or add special annotations to a Pod ([details]({{ "usage/build/run_in_containers.html" | true_relative_url }})).

#### `unshare(CLONE_NEWUSER): Operation not permitted`

The complete error message may look as follows:

```
Error during unshare(CLONE_NEWUSER): Operation not permitted
ERRO[0000] error parsing PID "": strconv.Atoi: parsing "": invalid syntax
ERRO[0000] (unable to determine exit status)
```

How to fix it: disable AppArmor and seccomp profiles with `--security-opt seccomp=unconfined` and `--security-opt apparmor=unconfined`, add special annotations to a Pod or use a privileged container ([details]({{ "usage/build/run_in_containers.html" | true_relative_url }})).

#### `User namespaces are not enabled in /proc/sys/kernel/unprivileged_userns_clone`

How to fix it:
```bash
# Enable unprivileged user namespaces:
echo 'kernel.unprivileged_userns_clone = 1' | sudo tee -a /etc/sysctl.conf
sudo sysctl -p
```