---
title: Run in docker container
permalink: advanced/ci_cd/run_in_container/run_in_docker_container.html
---

werf currently supports building of images with connection to the docker-server or without connection docker-server (in experimental mode). Building of images without docker-server is available in experimental mode for now, it is the only recommended way though.

## Build images without docker server (NEW!)

> NOTICE: Only dockerfile-images builder is available for this type of build for now. Stapel-images builder will be available soon.

There is official image with werf 1.2 for this method (1.1 is not supported): `ghcr.io/werf/werf`.

Example command:

```shell
docker run --rm --user 1000 \
    --device /dev/fuse \
    --security-opt seccomp=unconfined \
    --security-opt apparmor=unconfined \
    ghcr.io/werf/werf:latest WERF_COMMAND
```

## Build images with docker server (not recommended)

### Local docker server

This method supports building of dockerfile-images or stapel-images.

Example command:

```shell
docker run --rm --privileged \
    -v $HOME/.werf:/root/.werf \
    -v /tmp:/tmp \
    -v /var/run/docker.sock:/var/run/docker.sock \
    IMAGE WERF_COMMAND
```

Build your own docker image with werf for this method.

### Remote docker server

This method supports only building of dockerfile-images. Stapel-images is not supported, because stapel images builder use mounts from host system into docker images.

The easiest way to use remote docker server inside docker container is docker-in-docker (dind).

Build your own docker image based on `docker:dind` for this method.

Example command:

```shell
docker run --rm \
    -e DOCKER_HOST="tcp://HOST:PORT" \
    IMAGE WERF_COMMAND
```
