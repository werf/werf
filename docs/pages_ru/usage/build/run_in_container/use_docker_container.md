---
title: С помощью контейнеров Docker
permalink: usage/build/run_in_containers/use_docker_container.html
---

> ПРИМЕЧАНИЕ: werf поддерживает сборку образов с _использованием Docker-сервера_ или _с использованием Buildah_. Поддерживается сборка как Dockerfile-образов, так и stapel-образов через Buildah.

## Сборка образов с помощью Buildah (рекомендуемый метод)

Этот метод поддерживает сборку Dockerfile-образов или stapel-образов.

Для этого метода доступен официальный образ с werf 1.2 (1.1 не поддерживается): `registry.werf.io/werf/werf`.

Убедитесь, что удовлетворены [системные требования]({{ "usage/build/buildah_mode.html#системные-требования" | true_relative_url }}) и выберите один из [доступных режимов работы]({{ "usage/build/run_in_containers/how_it_works.html#режимы-работы" | true_relative_url }}) и перейдите к нему.

### Ядро Linux с поддержкой OverlayFS в режиме rootless

В этом случае необходимо отключить только профили **seccomp** и **AppArmor**. Для этого можно воспользоваться командой следующего вида:

```shell
docker run \
    --security-opt seccomp=unconfined --security-opt apparmor=unconfined \
    registry.werf.io/werf/werf:latest WERF_COMMAND
```

### Ядро Linux без поддержки OverlayFS в режиме rootless и использование привилегированного контейнера

В этом случае просто используйте привилегированный контейнер. Для этого можно воспользоваться командой следующего вида:

```shell
docker run \
    --privileged \
    registry.werf.io/werf/werf:latest WERF_COMMAND
```

### Ядро Linux без поддержки OverlayFS в режиме rootless и использование непривилегированного контейнера

Отключите профили seccomp и AppArmor и включите `/dev/fuse` в контейнере (чтобы `fuse-overlayfs` могла работать). Пример команды:

```shell
docker run \
    --device /dev/fuse \
    --security-opt seccomp=unconfined --security-opt apparmor=unconfined \
    registry.werf.io/werf/werf:latest WERF_COMMAND
```

## Сборка образов с помощью Docker-сервера (не рекомендуется)

### Локальный Docker-сервер

Этот метод поддерживает сборку Dockerfile-образов или stapel-образов.

Пример команды:

```shell
docker run \
    --privileged \
    --volume $HOME/.werf:/root/.werf \
    --volume /tmp:/tmp \
    --volume /var/run/docker.sock:/var/run/docker.sock \
    IMAGE WERF_COMMAND
```

Для этого метода соберите свой собственный образ Docker с помощью werf.

### Удаленный Docker-сервер

Этот метод поддерживает только сборку Dockerfile-образов. Образы stapel не поддерживаются, поскольку сборщик stapel-образов использует монтирования из хост-системы в Docker-образы.

Самый простой способ использовать удаленный Docker-сервер внутри Docker-контейнера — это Docker-in-Docker (dind).

Для этого метода соберите свой собственный образ на базе `docker:dind`.

Пример команды:

```shell
docker run \
    --env DOCKER_HOST="tcp://HOST:PORT" \
    IMAGE WERF_COMMAND
```

Для этого метода соберите свой собственный образ Docker с помощью werf.

## Устранение проблем

Если у вас возникли какие-либо сложности, пожалуйста, обратитесь к разделу [устранение проблем]({{ "usage/build/run_in_containers/how_it_works.html#устранение-проблем" | true_relative_url }})
