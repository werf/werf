---
title: Dockerfile Image
sidebar: documentation
permalink: documentation/configuration/dockerfile_image.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
directive_summary: dockerfile
---

Building image from Dockerfile is the easiest way to start using werf in an existing project.
Minimal `werf.yaml` below describes an image named `example` related with a project `Dockerfile`:

```yaml
project: my-project
configVersion: 1
---
image: example
dockerfile: Dockerfile
```

To specify some images from one Dockerfile:

```yaml
image: backend
dockerfile: Dockerfile
target: backend
---
image: frontend
dockerfile: Dockerfile
target: frontend
```

And also from different Dockerfiles:

```yaml
image: backend
dockerfile: dockerfiles/DockerfileBackend
---
image: frontend
dockerfile: dockerfiles/DockerfileFrontend
```

## Naming

{% include /configuration/stapel_image/naming.md %}

## Dockerfile directives

werf as well as Docker builds the image based on a Dockerfile and a context.

- `dockerfile` **(required)**: to set Dockerfile path relative to the project directory.
- `context`: to set build context PATH inside project directory (defaults to root of a project, `.`).
- `target`: to link specific Dockerfile stage (last one by default, see `docker build` \-\-target option).
- `args`: to set build-time variables (see `docker build` \-\-build-arg option).
- `addHost`: to add a custom host-to-IP mapping (host:ip) (see `docker build` \-\-add-host option).
- `network`: to set the networking mode for the RUN instructions during build (see `docker build` \-\-network option).
- `ssh`: to expose SSH agent socket or keys to the build (only if BuildKit enabled) (see `docker build` \-\-ssh option).
