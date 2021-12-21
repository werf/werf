---
title: С помощью контейнеров Docker
permalink: advanced/ci_cd/run_in_container/use_docker_container.html
---

> ПРИМЕЧАНИЕ: в настоящее время werf поддерживает сборку образов с _использованием Docker-сервера_ или _без его использования_ (в экспериментальном режиме). Сборка образов без Docker-сервера пока доступна в экспериментальном режиме, однако это единственный рекомендуемый способ запуска werf в контейнерах.

## Сборка образов без Docker-сервера (Новинка!)

> ВНИМАНИЕ: Пока для этого типа сборки доступен только сборщик образов на основе Dockerfile'ов. Сборщик Stapel будет доступен через некоторое время.

Для этого метода доступен официальный образ с werf 1.2 (1.1 не поддерживается): `ghcr.io/werf/werf`.

Выберите один из [доступных режимов работы]({{ "advanced/ci_cd/run_in_container/how_it_works.html#режимы-работы" | true_relative_url }}) и перейдите к нему.

### Ядро Linux с поддержкой OverlayFS в режиме rootless

В этом случае необходимо отключить только профили **seccomp** и **AppArmor**. Для этого можно воспользоваться командой следующего вида:

```shell
docker run \
    --security-opt seccomp=unconfined --security-opt apparmor=unconfined \
    ghcr.io/werf/werf:latest WERF_COMMAND
```

### Ядро Linux без поддержки OverlayFS в режиме rootless и использование привилегированного контейнера

В этом случае просто используйте привилегированный контейнер. Для этого можно воспользоваться командой следующего вида:

```shell
docker run \
    --privileged \
    ghcr.io/werf/werf:latest WERF_COMMAND
```

### Ядро Linux без поддержки OverlayFS в режиме rootless и использование непривилегированного контейнера

Отключите профили seccomp и AppArmor и включите `/dev/fuse` в контейнере (чтобы `fuse-overlayfs` могла работать). Пример команды:

```shell
docker run \
    --device /dev/fuse \
    --security-opt seccomp=unconfined --security-opt apparmor=unconfined \
    ghcr.io/werf/werf:latest WERF_COMMAND
```

## Сборка образов с помощью Docker-сервера (не рекомендуется)

### Локальный Docker-сервер

Этот метод поддерживает сборку Dockerfile-образов или Stapel-образов.

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

Этот метод поддерживает только сборку Dockerfile-образов. Образы Stapel не поддерживается, поскольку сборщик Stapel-образов использует монтирования из хост-системы в Docker-образы.

Самый простой способ использовать удаленный Docker-сервер внутри Docker-контейнера - это Docker-in-Docker (dind).

Для этого метода соберите свой собственный образ на базе `docker:dind`.

Пример команды:

```shell
docker run \
    --env DOCKER_HOST="tcp://HOST:PORT" \
    IMAGE WERF_COMMAND
```

Для этого метода соберите свой собственный образ Docker с помощью werf.

## Устранение проблем

Если у вас возникли какие-либо сложности, пожалуйста, обратитесь к разделу [Устранение проблем]({{ "advanced/ci_cd/run_in_container/how_it_works.html#устранение-проблем" | true_relative_url }})
