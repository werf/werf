---
title: Build backends
permalink: usage/build/backends.html
---

## Overview

werf supports the following build backends:

-	Docker — the traditional method that uses the system Docker Daemon.
-	Buildah — a secure, daemonless build option that supports rootless mode and is fully embedded into werf.

> The requirements and system preparation steps for using these build backends are described in the [Getting Started]({{ site.url }}/getting_started/) section of the website.

## Buildah

> Currently, Buildah is available only for Linux users and Windows users with WSL2 enabled. For macOS users, it is recommended to use a virtual machine to run werf in Buildah mode.

Buildah can be enabled by setting the environment variable `WERF_BUILDAH_MODE` to one of the following:

* auto — automatic mode selection based on the platform and environment.
*	native-chroot — uses chroot isolation for build containers.
*	native-rootless — uses rootless isolation for build containers. At this level, werf utilizes a container runtime for build operations (e.g., runc, crun, kata, or runsc).

```shell
export WERF_BUILDAH_MODE=auto
```

### Container registry configuration

When using Buildah, werf reads container registry settings from `registries.conf`.

For secure `docker.io` mirrors, you can also use the `--container-registry-mirror` option and `WERF_CONTAINER_REGISTRY_MIRROR_*` environment variables. Mirrors configured this way are treated as secure (`https`) mirrors by default. If needed, you can enable werf global insecure mode for registry access with `--insecure-registry` or `--skip-tls-verify-registry`.

The following paths are supported in priority order:

1. the path from `CONTAINERS_REGISTRIES_CONF`;
2. `~/.config/containers/registries.conf`;
3. `/etc/containers/registries.conf`.

If `CONTAINERS_REGISTRIES_CONF` is set, werf uses only that file and its neighboring `<path>.d` directory.

If `CONTAINERS_REGISTRIES_CONF` is not set, werf uses the first existing file from the standard paths and the neighboring `<path>.d` directory for that file.

werf uses the following data from this configuration:

- mirrors for `docker.io`;
- standalone insecure registries from `[[registry]] insecure = true`.

werf does not parse or override other container configuration files such as `shortnames.conf`.

An insecure mirror for `docker.io` does not automatically make the same host a standalone insecure registry. If the same host must be used both as a `docker.io` mirror and as a standalone insecure registry, it must be described by two separate entries.

### Storage Driver

werf can use either the `overlay` or `vfs` storage driver:

* `overlay`: Uses the OverlayFS file system. Either kernel-integrated OverlayFS support (if available) or the fuse-overlayfs implementation can be used. This is the recommended default option.
* `vfs`: Provides access to a virtual file system instead of OverlayFS. This option has lower performance and requires a privileged container, so it is not recommended. However, it may be useful in some cases.

In general, the default driver (`overlay`) should suffice. The storage driver can be specified using the environment variable `WERF_BUILDAH_STORAGE_DRIVER`.

### Ulimits

By default, Buildah mode in werf inherits the system ulimits when launching build containers. Users can override these parameters using the `WERF_BUILDAH_ULIMIT` environment variable.

Format: `WERF_BUILDAH_ULIMIT=type:softlimit[:hardlimit][,type:softlimit[:hardlimit],...]` — a comma-separated list of limits:

* `core`: maximum core dump size (ulimit -c).
* `cpu`: maximum CPU time (ulimit -t).
* `data`: maximum size of a process's data segment (ulimit -d).
* `fsize`: maximum size of new files (ulimit -f).
* `locks`: maximum number of file locks (ulimit -x).
* `memlock`: maximum amount of locked memory (ulimit -l).
* `msgqueue`: maximum amount of data in message queues (ulimit -q).
* `nice`: niceness adjustment (nice -n, ulimit -e).
* `nofile`: maximum number of open files (ulimit -n).
* `nproc`: maximum number of processes (ulimit -u).
* `rss`: maximum size of a process's resident set size (ulimit -m).
* `rtprio`: maximum real-time scheduling priority (ulimit -r).
* `rttime`: maximum real-time execution between blocking syscalls.
* `sigpending`: maximum number of pending signals (ulimit -i).
* `stack`: maximum stack size (ulimit -s).
