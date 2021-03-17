---
title: Dockerfile Image
sidebar: documentation
permalink: configuration/dockerfile_image.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
summary: |
  <a class="google-drawings" href="../images/configuration/dockerfile_image1.png" data-featherlight="image">
    <img src="../images/configuration/dockerfile_image1_preview.png">
  </a>

  <div class="language-yaml highlighter-rouge"><div class="highlight"><pre class="highlight"><code><span class="na">image</span><span class="pi">:</span> <span class="s">&lt;image name... || ~&gt;</span>
  <span class="na">dockerfile</span><span class="pi">:</span> <span class="s">&lt;relative path&gt;</span>
  <span class="na">context</span><span class="pi">:</span> <span class="s">&lt;relative path&gt;</span>
  <span class="na">target</span><span class="pi">:</span> <span class="s">&lt;docker stage name&gt;</span>
  <span class="na">args</span><span class="pi">:</span>
    <span class="s">&lt;build arg name&gt;</span><span class="pi">:</span> <span class="s">&lt;value&gt;</span>
  <span class="na">addHost</span><span class="pi">:</span>
  <span class="pi">-</span> <span class="s">&lt;host:ip&gt;</span>
  <span class="na">network</span><span class="pi">:</span> <span class="s">&lt;string&gt;</span>
  <span class="na">ssh</span><span class="pi">:</span> <span class="s">&lt;string&gt;</span>
  </code></pre></div></div>
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
