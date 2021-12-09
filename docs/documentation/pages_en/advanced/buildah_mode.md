---
title: Buildah mode
permalink: advanced/buildah_mode.html
---

werf currently supports building of images _with docker server_ or _without docker server_ (in experimental mode).  This page contains information, which is only applicable for experimental mode _without docker server_. Only dockerfile-images builder is available for this mode for now. Stapel-images builder will be available soon.

In experimental mode _without docker server_ werf uses built-in buildah in rootless mode.

## Enable buildah

Buildah is enabled by setting `WERF_BUILDAH_MODE` environment variable to one of the options: `auto`, `native-chroot`, `native-rootless` or `docker-with-fuse`.

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

If your kernel does not support rootless overlayfs, then fuse-overlayfs will be used.
