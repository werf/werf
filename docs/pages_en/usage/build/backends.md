---
title: Build backends
permalink: usage/build/backends.html
---

> **NOTE:** werf supports building images using either the _Docker daemon_ or _Buildah_. Both Dockerfile-based and stapel-based image builds are supported via Buildah.

## Docker

Docker is the default build backend. Werf uses `docker buildkit` by default, which can be disabled by setting the environment variable `DOCKER_BUILDKIT=0`.  
You can read more about `docker buildkit` [here](https://docs.docker.com/build/buildkit/).

### Multi-platform builds

You can create multi-platform images using QEMU emulation:

```bash
docker run --restart=always --name=qemu-user-static -d --privileged --entrypoint=/bin/sh multiarch/qemu-user-static -c "/register --reset -p yes && tail -f /dev/null"
```

## Buildah

For daemonless builds, werf uses the integrated Buildah.

Buildah can be enabled by setting the environment variable `WERF_BUILDAH_MODE` to one of the following: `auto`, `native-chroot`, `native-rootless`, or `docker`.
You can switch the build backend to Buildah by setting `WERF_BUILDAH_MODE=auto`.

```bash
# Switching to Buildah
export WERF_BUILDAH_MODE=auto
```

* `auto`: Automatic mode selection based on the platform and environment.
* `native-chroot`: Works only on Linux and uses chroot isolation for build containers.
* `native-rootless`: Works only on Linux and uses rootless isolation for build containers. At this isolation level, werf uses container runtime environments (runc, crun, kata, or runsc).

For most users, enabling Buildah mode is as simple as setting `WERF_BUILDAH_MODE=auto`.

> **NOTE:** Currently, Buildah is available only for Linux users and Windows users with WSL2 enabled. Full support for macOS users is planned but not yet available. For macOS users, it is recommended to use a virtual machine to run werf in Buildah mode.

## Storage Driver

werf can use either the `overlay` or `vfs` storage driver:

* `overlay`: Uses the OverlayFS file system. Either kernel-integrated OverlayFS support (if available) or the fuse-overlayfs implementation can be used. This is the recommended default option.
* `vfs`: Provides access to a virtual file system instead of OverlayFS. This option has lower performance and requires a privileged container, so it is not recommended. However, it may be useful in some cases.
In general, the default driver (`overlay`) should suffice. The storage driver can be specified using the environment variable `WERF_BUILDAH_STORAGE_DRIVER`.

## Ulimits

By default, Buildah mode in werf inherits the system ulimits when launching build containers. Users can override these parameters using the `WERF_BUILDAH_ULIMIT` environment variable.

Format: `WERF_BUILDAH_ULIMIT=type:softlimit[:hardlimit][,type:softlimit[:hardlimit],...]` — a comma-separated list of limits:

* "core": maximum core dump size (ulimit -c)
* "cpu": maximum CPU time (ulimit -t)
* "data": maximum size of a process's data segment (ulimit -d)
* "fsize": maximum size of new files (ulimit -f)
* "locks": maximum number of file locks (ulimit -x)
* "memlock": maximum amount of locked memory (ulimit -l)
* "msgqueue": maximum amount of data in message queues (ulimit -q)
* "nice": niceness adjustment (nice -n, ulimit -e)
* "nofile": maximum number of open files (ulimit -n)
* "nproc": maximum number of processes (ulimit -u)
* "rss": maximum size of a process's resident set size (ulimit -m)
* "rtprio": maximum real-time scheduling priority (ulimit -r)
* "rttime": maximum real-time execution between blocking syscalls
* "sigpending": maximum number of pending signals (ulimit -i)
* "stack": maximum stack size (ulimit -s)

## Available werf images

Below is a list of images with the built-in werf utility. Each image is updated as part of the release process, based on the trdl package manager ([more about update channels]({{ site.url }}/about/release_channels.html)).

* `registry.werf.io/werf/werf:latest` -> `registry.werf.io/werf/werf:1.2-stable`;
* `registry.werf.io/werf/werf:1.2-alpha` -> `registry.werf.io/werf/werf:1.2-alpha-alpine`;
* `registry.werf.io/werf/werf:1.2-beta` -> `registry.werf.io/werf/werf:1.2-beta-alpine`;
* `registry.werf.io/werf/werf:1.2-ea` -> `registry.werf.io/werf/werf:1.2-ea-alpine`;
* `registry.werf.io/werf/werf:1.2-stable` -> `registry.werf.io/werf/werf:1.2-stable-alpine`;
* `registry.werf.io/werf/werf:1.2-rock-solid` -> `registry.werf.io/werf/werf:1.2-rock-solid-alpine`;
* `registry.werf.io/werf/werf:1.2-alpha-alpine`;
* `registry.werf.io/werf/werf:1.2-beta-alpine`;
* `registry.werf.io/werf/werf:1.2-ea-alpine`;
* `registry.werf.io/werf/werf:1.2-stable-alpine`;
* `registry.werf.io/werf/werf:1.2-rock-solid-alpine`;
* `registry.werf.io/werf/werf:1.2-alpha-ubuntu`;
* `registry.werf.io/werf/werf:1.2-beta-ubuntu`;
* `registry.werf.io/werf/werf:1.2-ea-ubuntu`;
* `registry.werf.io/werf/werf:1.2-stable-ubuntu`;
* `registry.werf.io/werf/werf:1.2-rock-solid-ubuntu`;
* `registry.werf.io/werf/werf:1.2-alpha-fedora`;
* `registry.werf.io/werf/werf:1.2-beta-fedora`;
* `registry.werf.io/werf/werf:1.2-ea-fedora`;
* `registry.werf.io/werf/werf:1.2-stable-fedora`;
* `registry.werf.io/werf/werf:1.2-rock-solid-fedora`.

## Example local build

```yaml
# werf.yaml
project: app
configVersion: 1
build:
  platform:
    - linux/amd64
    - linux/arm64
---
image: alpine
dockerfile: Dockerfile
```

```
# Dockerfile
FROM alpine
```
Building with docker

```bash
werf build # или DOCKER_BUILDKIT=0 werf build
```

Building with buildah

```bash
WERF_BUILDAH_MODE=auto werf build
```

