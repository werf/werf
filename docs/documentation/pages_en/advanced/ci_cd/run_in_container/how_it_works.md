---
title: How it works
permalink: advanced/ci_cd/run_in_container/how_it_works.html
---

werf currently supports building of images _with docker server_ or _without docker server_ (in experimental mode).  This page contains information, which is only applicable for experimental mode _without docker server_. Only dockerfile-images builder is available for this mode for now. Stapel-images builder will be available soon.

In experimental mode _without docker server_ werf uses built-in buildah in rootless mode.

## Enable buildah

Experimental buildah support is enabled by setting `WERF_BUILDAH_MODE` environment variable to one of the options: `auto`, `native-chroot`, `native-rootless` or `docker-with-fuse`.

* `auto` — select mode automatically based on your platform and environment.
* `native-chroot` works only in Linux and uses `chroot` isolation level when running build containers.
* `native-rootless` works only in Linux and uses `rootless` isolation leven when running build containers. In this isolation level werf will use container runtime (runc or crun).
* `docker-with-fuse` is crossplatform mode, and it is the only available choice in MacOS or Windows.

Most users should just set `WERF_BUILDAH_MODE=auto` to enable experimental buildah mode.

## Storage driver

Werf may use `overlay` or `vfs` storage driver:

* `overlay` enables usage of overlayfs filesystem. Either native linux kernel overlayfs could be used if available or fuse-overlayfs. It is the default and recommended choice.
* `vfs` enables usage of virtual filesystem emulation instead of overlayfs. This filesystem has worse performance and requires privileged container, thus it is not recommended. But it may be required in some cases.

Generally user should just use default `overlay`. Storage driver could be selected with the `WERF_BUILDAH_STORAGE_DRIVER` environment variable.

### System requirements

Starting with Linux kernel version 5.11 rootless overlayfs is available (technically from 5.13 version, which contains bugfix to enable rootless overlayfs with SELinux, but most major linux distributions has been backported this bugfix to 5.11 kernel).

## Modes of operation

Depending on system requirements and user needs there are 3 ways to use werf with buildah inside containers:

1. Use [Linux kernel with rootless overlayfs](#system-requirements).
2. Use [Linux kernel without rootless overlayfs](#system-requirements) and privileged container.
3. Use [Linux kernel without rootless overlayfs](#system-requirements) and non-privileged container with additional settings.

### Linux kernel with rootless overlayfs

This way you only required to **disable apparmor** and **seccomp** profiles in a container which runs werf.

### Linux kernel without rootless overlayfs and privileged container

In the case your Linux kernel does not have [rootless overlayfs](#system-requirements), buildah and werf uses fuse-overlayfs instead.

To enable usage of fuse-overlayfs in such case using a **privileged container** to run werf is one option.

### Linux kernel without rootless overlayfs and non-privileged container

In the case your Linux kernel does not have [rootless overlayfs](#system-requirements), buildah and werf uses fuse-overlayfs instead.

To enable usage of fuse-overlayfs in such case without using a privileged container, run werf in a container with the following options:

1. **Disabled apparmor** and **seccomp** profiles .
2. Enabled `/dev/fuse` device.

## Troubleshooting

### `fuse: device not found`

Full error may look like:

```
error mounting new container: error mounting build container "53f916e7a334a4bb0d9dbc38a0901718d40b99765002bb7f2f2e5464b1db4294": error creating overlay mount to /home/build/.local/share/containers/storage/overlay/49e856f537ba58afdc09137291133994cd1305e40df72c4fab43077cbd405477/merged, mount_data=",lowerdir=/home/build/.local/share/containers/storage/overlay/l/Z5GEVIFIIQ7H262DYUTX3YOVR6:/home/build/.local/share/containers/storage/overlay/l/PJBBW6UNUNGI37IX6R3LDNPX3J:/home/build/.local/share/containers/storage/overlay/l/MUYSUONLQVE4CJMQVDCH2UBAVQ:/home/build/.local/share/containers/storage/overlay/l/67JHKJDCKBTI4R3Q5S5YG44AD3:/home/build/.local/share/containers/storage/overlay/l/3S72G4SWKDXILGANUOCESP5LDK,upperdir=/home/build/.local/share/containers/storage/overlay/49e856f537ba58afdc09137291133994cd1305e40df72c4fab43077cbd405477/diff,workdir=/home/build/.local/share/containers/storage/overlay/49e856f537ba58afdc09137291133994cd1305e40df72c4fab43077cbd405477/work,volatile": using mount program /usr/bin/fuse-overlayfs: fuse: device not found, try 'modprobe fuse' first
fuse-overlayfs: cannot mount: No such file or directory
: exit status 1
time="2021-12-06T11:30:20Z" level=error msg="exit status 1"
```

Solution: enable fuse device for a container which runs werf, [more info](#linux-kernel-without-rootless-overlayfs-and-non-privileged-container).

### `flags: 0x1000: permission denied`

Full error may look like:

```
time="2021-12-06T11:23:23Z" level=debug msg="unable to create kernel-style whiteout: operation not permitted"
time="2021-12-06T11:23:23Z" level=debug msg="[graphdriver] trying provided driver \"overlay\""
time="2021-12-06T11:23:23Z" level=debug msg="overlay: mount_program=/usr/bin/fuse-overlayfs"
Running time 0.01 seconds
Error: unable to get buildah client: unable to create new Buildah instance with mode "native-rootless": unable to get storage: mount /home/build/.local/share/containers/storage/overlay:/home/build/.local/share/containers/storage/overlay, flags: 0x1000: permission denied
time="2021-12-06T11:23:23Z" level=error msg="exit status 1"
```

Solution: disable apparmor and seccomp profiles with options like `--security-opt seccomp=unconfined` and `--security-opt apparmor=unconfined` or by setting special annotations in a Pod, [more info](#linux-kernel-without-rootless-overlayfs-and-non-privileged-container)

### `unshare(CLONE_NEWUSER): Operation not permitted`

Full error may look like:

```
Error during unshare(CLONE_NEWUSER): Operation not permitted
ERRO[0000] error parsing PID "": strconv.Atoi: parsing "": invalid syntax 
ERRO[0000] (unable to determine exit status)            
```

Solution: disable apparmor and seccomp profiles with options like `--security-opt seccomp=unconfined` and `--security-opt apparmor=unconfined` or by setting special annotations in a Pod, or use a privileged container, [more info](#modes-of-operation).
