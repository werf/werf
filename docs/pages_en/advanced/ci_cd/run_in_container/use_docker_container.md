---
title: Use Docker container
permalink: advanced/ci_cd/run_in_container/use_docker_container.html
---

> NOTICE: werf supports building images _with the Docker server_ or _with Buildah_. This page contains information applicable only to the mode _with Buildah_. Buildah supports building either Dockerfile images or stapel images.

## Build images with Buildah (recommended)

This method supports building Dockerfile images or stapel images.

There is an official image with werf 1.2 for this method (1.1 is not supported): `registry.werf.io/werf/werf`.

Make sure you meet all [system requirements]({{ "advanced/buildah_mode.html#system-requirements" | true_relative_url }}) and select one of the [available operating modes]({{ "advanced/ci_cd/run_in_container/how_it_works.html#modes-of-operation" | true_relative_url }}).

### Linux kernel with rootless OverlayFS

In this case, you only need to disable the seccomp and AppArmor profiles. Below is an example of a command that does this:

```shell
docker run \
    --security-opt seccomp=unconfined --security-opt apparmor=unconfined \
    registry.werf.io/werf/werf:latest WERF_COMMAND
```

### Linux kernel without rootless OverlayFS and privileged container

In this case, just use the privileged container. Below is an example of a command that does this:

```shell
docker run \
    --privileged \
    registry.werf.io/werf/werf:latest WERF_COMMAND
```

### Linux kernel without rootless OverlayFS and non-privileged container

In this case, disable the seccomp and AppArmor profiles and enable `/dev/fuse` in the container (so that `fuse-overlayfs` can work). Below is an example of a command that does this:

```shell
docker run \
    --device /dev/fuse \
    --security-opt seccomp=unconfined --security-opt apparmor=unconfined \
    registry.werf.io/werf/werf:latest WERF_COMMAND
```

## Build images with the Docker server (not recommended)

### Local Docker server

This method supports building Dockerfile images or stapel images.

Below is an example of a command that does this:

```shell
docker run \
    --privileged \
    --volume $HOME/.werf:/root/.werf \
    --volume /tmp:/tmp \
    --volume /var/run/docker.sock:/var/run/docker.sock \
    IMAGE WERF_COMMAND
```

For this method, build your own Docker image using werf.

### Remote Docker server

This method only supports building Dockerfile images. Stapel images are not supported because the stapel image builder uses mounts from the host system to Docker images.

The easiest way to use a remote Docker server inside a Docker container is Docker-in-Docker (dind).

For this method, build your own image based on `docker:dind`.

Below is an example of a command:

```shell
docker run \
    --env DOCKER_HOST="tcp://HOST:PORT" \
    IMAGE WERF_COMMAND
```

Build your own docker image with werf for this method.

## Troubleshooting

In case of problems, refer to the [troubleshooting section]({{ "advanced/ci_cd/run_in_container/how_it_works.html#troubleshooting" | true_relative_url }})
