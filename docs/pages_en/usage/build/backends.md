---
title: Build backends
permalink: usage/build/backends.html
---

## Overview

werf supports the following build backends:

-	Docker — the traditional method that uses the system Docker Daemon. Selected by default when no BuildKit endpoint is configured.
-	BuildKit — builds images through an external [buildkitd](https://github.com/moby/buildkit) daemon. Selected by setting a BuildKit endpoint via environment variable.

> The requirements and system preparation steps for using these build backends are described in the [Getting Started]({{ site.url }}/getting_started/) section of the website.

## BuildKit

The BuildKit backend is enabled by setting `WERF_BUILDKIT_HOST` (or the standard `BUILDKIT_HOST`) to the address of a running buildkitd daemon. When neither variable is set, werf uses the Docker backend.

### Endpoints

The following endpoint schemes are supported:

*	`unix://` — local Unix socket;
*	`tcp://` — TCP endpoint;
*	`docker-container://` — buildkitd running inside a Docker container;
*	`kube-pod://` — buildkitd running inside a Kubernetes pod;
*	`podman-container://` — buildkitd running inside a Podman container;
*	`ssh://` — buildkitd reachable over SSH.

### Quick start

Run buildkitd in a local Docker container and point werf at it:

```shell
docker run -d --name buildkitd --privileged moby/buildkit
export BUILDKIT_HOST=docker-container://buildkitd
```

After that any werf build command will use the BuildKit backend.

### Container registry required

The BuildKit backend requires a remote container registry. The `--repo` option (or `WERF_REPO`) is mandatory — the local stages storage (`:local`) is not supported. Built stages are pushed by digest directly from buildkitd to the configured repo; werf then applies its stage tag to the pushed manifest on the registry side.

### Supported features

The BuildKit backend supports both build modes on par with the Docker backend:

*	Stapel builds (shell and ansible builders).
*	Dockerfile builds, both staged and non-staged.

Feature parity includes:

*	ssh-agent forwarding for the build;
*	build secrets;
*	custom build networks (`default`, `host`, `none`);
*	custom mounts.

### Host mounts semantics

Stapel host mounts (`fromPath`, `mount: build_dir`) are mapped to BuildKit persistent cache mounts keyed by the host path. The data lives inside the buildkitd cache on the daemon side rather than in a directory on the werf host. The cache persists across builds and is shared by host-path key.

### Insecure and self-signed registries

Insecure registry access, custom CAs and TLS verification skipping are configured on the buildkitd daemon side (typically via `buildkitd.toml`). werf does not forward its own `--insecure-registry` / `--skip-tls-verify-registry` flags to buildkitd.

### Host cleanup

With a remote buildkitd there is no local image store on the werf host. `werf host purge` and other host-cleanup commands prune the buildkitd build cache instead of a local storage directory.

### Limitations

In the first iteration, the following options are not supported by the BuildKit backend:

*	`--introspect-error`;
*	`--introspect-before-error`;
*	`--introspect-stage`.
