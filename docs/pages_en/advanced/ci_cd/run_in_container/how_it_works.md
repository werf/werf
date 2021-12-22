---
title: How it works
permalink: advanced/ci_cd/run_in_container/how_it_works.html
---

werf currently supports building images _with the Docker server_ or _without the Docker server_ (in experimental mode).  This page contains information applicable only to the experimental mode _without the Docker server_. For now, only the Dockerfile image builder is available for this mode. The Stapel image builder will be available soon.

General info on how to enable Buildah in werf is available on the [Buildah mode page]({{ "/advanced/buildah_mode.html" | true_relative_url }}).

## Modes of operation

Depending on whether the user's system meets the [Buildah mode system requirements]({{ "/advanced/buildah_mode.html#system-requirements" | true_relative_url }}) and depending on the user needs, there are 3 ways to use werf with Buildah inside containers:

1. Use [Linux kernel with rootless OverlayFS]({{ "/advanced/buildah_mode.html#system-requirements" | true_relative_url }}).
2. Use [Linux kernel without rootless OverlayFS]({{ "/advanced/buildah_mode.html#system-requirements" | true_relative_url }}) and a privileged container.
3. Use [Linux kernel without rootless OverlayFS]({{ "/advanced/buildah_mode.html#system-requirements" | true_relative_url }}) and an unprivileged container with additional settings.

### Linux kernel with rootless OverlayFS

In this case, you only need to **disable the AppArmor** and **seccomp** profiles in the container where werf is running.

### Linux kernel without rootless OverlayFS and privileged container

If your Linux kernel does not have [rootless OverlayFS]({{ "/advanced/buildah_mode.html#system-requirements" | true_relative_url }}), Buildah and werf will use fuse-overlayfs instead.

One of the options in this case is to run werf in a **privileged container** to enable fuse-overlayfs.

### Linux kernel without rootless OverlayFS and non-privileged container

If your Linux kernel does not support [OverlayFS in rootless mode]({{ "/advanced/buildah_mode.html#system-requirements" | true_relative_url }}), Buildah and werf will use fuse-overlayfs instead.

To use fuse-overlayfs without a privileged container, run werf in a container with the following parameters:

1. The **AppArmor** and **seccomp** profiles are **disabled**.
2. The `/dev/fuse` device is enabled.

## Available werf images

Below is a list of images with built-in werf. Each image is updated as part of a release process based on the trdl package manager ([learn more about the release channels]({{ "installation.html#all-changes-in-werf-go-through-all-release-channels" | true_relative_url }}).

* `ghcr.io/werf/werf:latest` -> `ghcr.io/werf/werf:1.2-stable`;
* `ghcr.io/werf/werf:1.2-alpha` -> `ghcr.io/werf/werf:1.2-alpha-alpine`;
* `ghcr.io/werf/werf:1.2-beta` -> `ghcr.io/werf/werf:1.2-beta-alpine`;
* `ghcr.io/werf/werf:1.2-ea` -> `ghcr.io/werf/werf:1.2-ea-alpine`;
* `ghcr.io/werf/werf:1.2-stable` -> `ghcr.io/werf/werf:1.2-stable-alpine`;
* `ghcr.io/werf/werf:1.2-alpha-alpine`;
* `ghcr.io/werf/werf:1.2-beta-alpine`;
* `ghcr.io/werf/werf:1.2-ea-alpine`;
* `ghcr.io/werf/werf:1.2-stable-alpine`;
* `ghcr.io/werf/werf:1.2-alpha-ubuntu`;
* `ghcr.io/werf/werf:1.2-beta-ubuntu`;
* `ghcr.io/werf/werf:1.2-ea-ubuntu`;
* `ghcr.io/werf/werf:1.2-stable-ubuntu`;
* `ghcr.io/werf/werf:1.2-alpha-centos`;
* `ghcr.io/werf/werf:1.2-beta-centos`;
* `ghcr.io/werf/werf:1.2-ea-centos`;
* `ghcr.io/werf/werf:1.2-stable-centos`;
* `ghcr.io/werf/werf:1.2-alpha-fedora`;
* `ghcr.io/werf/werf:1.2-beta-fedora`;
* `ghcr.io/werf/werf:1.2-ea-fedora`;
* `ghcr.io/werf/werf:1.2-stable-fedora`.

## Troubleshooting

### `fuse: device not found`

The complete error message may look as follows:

```
error mounting new container: error mounting build container "53f916e7a334a4bb0d9dbc38a0901718d40b99765002bb7f2f2e5464b1db4294": error creating overlay mount to /home/build/.local/share/containers/storage/overlay/49e856f537ba58afdc09137291133994cd1305e40df72c4fab43077cbd405477/merged, mount_data=",lowerdir=/home/build/.local/share/containers/storage/overlay/l/Z5GEVIFIIQ7H262DYUTX3YOVR6:/home/build/.local/share/containers/storage/overlay/l/PJBBW6UNUNGI37IX6R3LDNPX3J:/home/build/.local/share/containers/storage/overlay/l/MUYSUONLQVE4CJMQVDCH2UBAVQ:/home/build/.local/share/containers/storage/overlay/l/67JHKJDCKBTI4R3Q5S5YG44AD3:/home/build/.local/share/containers/storage/overlay/l/3S72G4SWKDXILGANUOCESP5LDK,upperdir=/home/build/.local/share/containers/storage/overlay/49e856f537ba58afdc09137291133994cd1305e40df72c4fab43077cbd405477/diff,workdir=/home/build/.local/share/containers/storage/overlay/49e856f537ba58afdc09137291133994cd1305e40df72c4fab43077cbd405477/work,volatile": using mount program /usr/bin/fuse-overlayfs: fuse: device not found, try 'modprobe fuse' first
fuse-overlayfs: cannot mount: No such file or directory
: exit status 1
time="2021-12-06T11:30:20Z" level=error msg="exit status 1"
```

Solution: enable the fuse device for the container in which werf is running ([details](#linux-kernel-without-rootless-overlayfs-and-non-privileged-container)).

### `flags: 0x1000: permission denied`

The complete error message may look as follows:

```
time="2021-12-06T11:23:23Z" level=debug msg="unable to create kernel-style whiteout: operation not permitted"
time="2021-12-06T11:23:23Z" level=debug msg="[graphdriver] trying provided driver \"overlay\""
time="2021-12-06T11:23:23Z" level=debug msg="overlay: mount_program=/usr/bin/fuse-overlayfs"
Running time 0.01 seconds
Error: unable to get buildah client: unable to create new Buildah instance with mode "native-rootless": unable to get storage: mount /home/build/.local/share/containers/storage/overlay:/home/build/.local/share/containers/storage/overlay, flags: 0x1000: permission denied
time="2021-12-06T11:23:23Z" level=error msg="exit status 1"
```

Solution: disable AppArmor and seccomp profiles with `--security-opt seccomp=unconfined` and `--security-opt apparmor=unconfined` or add special annotations to a Pod ([details](#linux-kernel-without-rootless-overlayfs-and-non-privileged-container)).

### `unshare(CLONE_NEWUSER): Operation not permitted`

The complete error message may look as follows:

```
Error during unshare(CLONE_NEWUSER): Operation not permitted
ERRO[0000] error parsing PID "": strconv.Atoi: parsing "": invalid syntax 
ERRO[0000] (unable to determine exit status)            
```

Solution: disable AppArmor and seccomp profiles with `--security-opt seccomp=unconfined` and `--security-opt apparmor=unconfined`, add special annotations to a Pod or use a privileged container ([details](#modes-of-operation)).
