---
title: Buildah mode
permalink: advanced/buildah_mode.html
---

werf currently supports building images _with the Docker server_ or _without the Docker server_ (in experimental mode).  This page contains information applicable only to the experimental mode _without the Docker server_. For now, only the Dockerfile image builder is available for this mode. The Stapel image builder will be available soon.

In the experimental mode _without the Docker server_, werf uses built-in Buildah in rootless mode.

## Enable Buildah

Buildah is enabled by setting the `WERF_BUILDAH_MODE` environment variable to one of the following: `auto`, `native-chroot`, `native-rootless` or `docker-with-fuse`.

* `auto` — select the mode automatically based on your platform and environment.
* `native-chroot` works only on Linux and uses the `chroot` isolation level when running build containers.
* `native-rootless` works only on Linux and uses the `rootless` isolation leven when running build containers. At this isolation level, werf will use container runtime (runc or crun).
* `docker-with-fuse` is a cross-platform mode and is the only choice available on MacOS or Windows.

Most users only need to set `WERF_BUILDAH_MODE=auto` to enable the experimental Buildah-based mode.

## Storage driver

werf can use `overlay` or `vfs` storage driver:

* `overlay` allows you to use the OverlayFS filesystem. You can either use the native Linux kernel's OverlayFS (if available) or fuse-overlayfs. It is the default and recommended choice.
* `vfs` allows you to use a virtual filesystem emulation instead of OverlayFS. This filesystem has worse performance and requires a privileged container, so its use is not recommended. However, it may be required in some cases.

Normally, the user should just go with the default `overlay` driver. The storage driver can be selected with the `WERF_BUILDAH_STORAGE_DRIVER` environment variable.

### System requirements

Rootless OverlayFS is available since Linux kernel version 5.11 (strictly speaking, it was first implemented in version 5.13 which contains a bugfix to enable rootless overlayFS in SELinux, but most major Linux distributions have backported it into kernel 5.11).

If your kernel does not support rootless OverlayFS, fuse-overlayfs will be used.
