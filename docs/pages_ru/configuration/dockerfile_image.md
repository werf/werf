---
title: Dockerfile-образ
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

Сборка образа с использованием имеющегося Dockerfile — самый простой путь начать использовать Werf в существующем проекте. Ниже приведен пример минимального файла `werf.yaml`, описывающего образ `example` проекта:

```yaml
project: my-project
configVersion: 1
---
image: example
dockerfile: Dockerfile
```

Также, вы можете описывать несколько образов из одного и того же Dockerfile:

```yaml
image: backend
dockerfile: Dockerfile
target: backend
---
image: frontend
dockerfile: Dockerfile
target: frontend
```

И конечно, вы можете описывать образы, основанные на разных Dockerfile:

```yaml
image: backend
dockerfile: dockerfiles/DockerfileBackend
---
image: frontend
dockerfile: dockerfiles/DockerfileFrontend
```

## Именование

{% include image_configuration/naming_ru.md %}

## Директивы Dockerfile

Werf, так же как и `docker build`, собирает образы, используя Dockerfile-инструкции, контекст, а также дополнительные опции:

- `dockerfile` **(обязателен)**: определяет путь к Dockerfile относительно папки проекта.
- `context`: определяет путь к контексту внутри папки проекта (по умолчанию — папка проекта, `.`).
- `target`: связывает конкретную стадию Dockerfile (по умолчанию — последнюю, смотри `docker build` \-\-target).
- `args`: устанавливает переменные окружения на время сборки (смотри `docker build` \-\-build-arg).
- `addHost`: устанавливает связь host-to-IP (host:ip) (смотри `docker build` \-\-add-host).
