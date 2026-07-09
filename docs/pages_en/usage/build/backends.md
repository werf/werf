---
title: Build backends
permalink: usage/build/backends.html
---

## Overview

werf builds images through a [buildkitd](https://github.com/moby/buildkit) daemon. The endpoint is selected as follows:

-	`WERF_BUILDKIT_HOST` (or the standard `BUILDKIT_HOST`) is set — werf uses the specified buildkitd endpoint.
-	Neither variable is set — werf automatically starts (or reuses) a local buildkitd container named `werf-buildkitd` on the local Docker daemon and uses it via `docker-container://werf-buildkitd`. Docker must be available in this case.

> The requirements and system preparation steps for using these build backends are described in the [Getting Started]({{ site.url }}/getting_started/) section of the website.

## BuildKit

### Endpoints

The following endpoint schemes are supported:

*	`unix://` — local Unix socket;
*	`tcp://` — TCP endpoint;
*	`docker-container://` — buildkitd running inside a Docker container;
*	`kube-pod://` — buildkitd running inside a Kubernetes pod;
*	`podman-container://` — buildkitd running inside a Podman container;
*	`ssh://` — buildkitd reachable over SSH.

### Quick start

With Docker available locally no setup is needed: werf starts a `werf-buildkitd` container automatically on first build.

To use an external buildkitd instead:

```shell
export BUILDKIT_HOST=tcp://my-buildkitd:1234
```

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

Stapel host mounts (`fromPath`, `mount: build_dir`) are mapped to BuildKit persistent cache mounts keyed by the host path. The data lives inside the buildkitd cache on the daemon side rather than in a directory on the werf host. The cache persists across builds and is shared by host-path key. Note that pre-existing contents of the host directory are NOT delivered into the mount: the cache mount starts empty on first use and only accumulates data written during builds.

### Insecure and self-signed registries

Insecure registry access, custom CAs and TLS verification skipping are configured on the buildkitd daemon side (typically via `buildkitd.toml`). werf does not forward its own `--insecure-registry` / `--skip-tls-verify-registry` flags to buildkitd.

### Host cleanup

With a remote buildkitd there is no local image store on the werf host. `werf host purge` and other host-cleanup commands only clean up werf-owned service directories on the host; the buildkitd build cache is not pruned by werf in the first iteration — it is managed by buildkitd garbage collection (see `buildkitd.toml`) or manually via `buildctl prune`.
