---
title: Run werf in docker server
permalink: advanced/ci_cd/werf_in_container/use_docker.html
---

Werf currently supports building of images with docker server or without docker server (in experimental mode). It is recommended to use method without docker-server though to run werf in docker container.

## Build images with docker server (not recommended)

### Local docker server

This method supports building of dockerfile-images or stapel-images.

Example command:

```
docker run --rm --privileged -v $HOME/.werf:/root/.werf -v /tmp:/tmp -v /var/run/docker.sock:/var/run/docker.sock IMAGE WERF_COMMAND
```

Build your own docker image with werf for this method.
---
title: Run werf in Github Actions with docker runner
permalink: advanced/ci_cd/werf_in_container/use_docker_for_github_actions.html
---

This page is under construction.
### Remote docker server

This method supports only building of dockerfile-images. Stapel-images is not supported, because stapel images builder use mounts from host system into docker images.

The easiest way to use remote docker server inside docker container is docker-in-docker (dind).

Build your own docker image based on `docker:dind` for this method.

Example command:

```
docker run --rm -e DOCKER_HOST="tcp://HOST:PORT" IMAGE_BASED_ON_DIND WERF_COMMAND
```

## Build images without docker server (NEW! Recommended)

**NOTE** Only dockerfile-images builder is available for this type of build for now. Stapel-images builder will be available soon.

There is official image with werf 1.2 for this method (1.1 is not supported): `ghcr.io/werf/werf`.

Example command:

```
docker run --rm --user 1000 --device /dev/fuse --security-opt seccomp=unconfined --security-opt apparmor=unconfined ghcr.io/werf/werf:latest WERF_COMMAND
```
