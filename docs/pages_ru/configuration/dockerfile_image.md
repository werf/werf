---
title: Dockerfile-образ
sidebar: documentation
permalink: documentation/configuration/dockerfile_image.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
directive_summary: dockerfile
---

Сборка образа с использованием имеющегося Dockerfile — самый простой путь начать использовать werf в существующем проекте. Ниже приведен пример минимального файла `werf.yaml`, описывающего образ `example` проекта:

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

werf, так же как и `docker build`, собирает образы, используя Dockerfile-инструкции, контекст, а также дополнительные опции:

- `dockerfile` **(обязателен)**: определяет путь к Dockerfile относительно папки проекта.
- `context`: определяет путь к контексту внутри папки проекта (по умолчанию — папка проекта, `.`).
- `target`: связывает конкретную стадию Dockerfile (по умолчанию — последнюю, смотри `docker build` \-\-target).
- `args`: устанавливает переменные окружения на время сборки (смотри `docker build` \-\-build-arg).
- `addHost`: устанавливает связь host-to-IP (host:ip) (смотри `docker build` \-\-add-host).
