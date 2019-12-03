---
title: Dockerfile Image
sidebar: documentation
permalink: documentation/configuration/dockerfile_image.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
summary: |
  <a class="google-drawings" href="https://docs.google.com/drawings/d/e/2PACX-1vRrzxht-PmC-4NKq95DtLS9E7JrvtuHy0JpMKdylzlZtEZ5m7bJwEMJ6rXTLevFosWZXmi9t3rDVaPB/pub?w=2031&amp;h=144" data-featherlight="image">
    <img src="https://docs.google.com/drawings/d/e/2PACX-1vRrzxht-PmC-4NKq95DtLS9E7JrvtuHy0JpMKdylzlZtEZ5m7bJwEMJ6rXTLevFosWZXmi9t3rDVaPB/pub?w=1016&amp;h=72">
  </a>

  <div class="language-yaml highlighter-rouge"><div class="highlight"><pre class="highlight"><code><span class="na">image</span><span class="pi">:</span> <span class="s">&lt;image name... || ~&gt;</span>
  <span class="na">dockerfile</span><span class="pi">:</span> <span class="s">&lt;relative path&gt;</span>
  <span class="na">context</span><span class="pi">:</span> <span class="s">&lt;relative path&gt;</span>
  <span class="na">target</span><span class="pi">:</span> <span class="s">&lt;docker stage name&gt;</span>
  <span class="na">args</span><span class="pi">:</span>
    <span class="s">&lt;build arg name&gt;</span><span class="pi">:</span> <span class="s">&lt;value&gt;</span>
  <span class="na">addHost</span><span class="pi">:</span>
  <span class="pi">-</span> <span class="s">&lt;host:ip&gt;</span>
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

{% include image_configuration/naming.md %}

## Dockerfile directives

Werf as well as Docker builds the image based on a Dockerfile and a context.

- `dockerfile` **(required)**: to set Dockerfile path relative to the project directory.
- `context`: to set build context PATH inside project directory (defaults to root of a project, `.`).
- `target`: to link specific Dockerfile stage (last one by default, see `docker build` \-\-target option).
- `args`: to set build-time variables (see `docker build` \-\-build-arg option).
- `addHost`: to add a custom host-to-IP mapping (host:ip) (see `docker build` \-\-add-host option).
