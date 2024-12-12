---
title: Сборочные бэкенды
permalink: usage/build/backends.html
---

> ПРИМЕЧАНИЕ: werf поддерживает сборку образов с _использованием Docker-демона_ или _с использованием Buildah_. Поддерживается сборка как Dockerfile-образов, так и stapel-образов через Buildah.

## Docker

Docker - сборочный бэкенд по умолчанию. Werf использует `docker buildkit` по умолчанию, его можно отключить, указав переменную окружения `DOCKER_BUILDKIT=0`
Более подробно прочитать про `docker buildkit` можно [здесь](https://docs.docker.com/build/buildkit/)

### Мультиплатформенная сборка

Вы можете создавать мультиплатформенные образы, используя эмуляцию QEMU

```bash
docker run --restart=always --name=qemu-user-static -d --privileged --entrypoint=/bin/sh multiarch/qemu-user-static -c "/register --reset -p yes && tail -f /dev/null"
```

## Buildah

Для сборки без Docker-демона werf использует встроенный Buildah.

Buildah включается установкой переменной окружения `WERF_BUILDAH_MODE` в один из вариантов: `auto`, `native-chroot`, `native-rootless` или `docker`.
Можно переключить сборочный бэкенд на Buildah, указав переменную окружения `WERF_BUILDAH_MODE=auto`

* `auto` — автоматический выбор режима в зависимости от платформы и окружения;
* `native-chroot` работает только в Linux и использует `chroot`-изоляцию для сборочных контейнеров;
* `native-rootless` работает только в Linux и использует `rootless`-изоляцию для сборочных контейнеров. На этом уровне изоляции werf использует среду выполнения сборочных операций в контейнерах (runc, crun, kata или runsc).

Большинству пользователей для включения режима Buildah достаточно установить `WERF_BUILDAH_MODE=auto`.

> ПРИМЕЧАНИЕ: На данный момент Buildah доступен только для пользователей Linux и Windows с включённой подсистемой WSL2. Полная поддержка для пользователей MacOS запланирована, но пока не доступна. Для пользователей MacOS на данный момент предлагается использование виртуальной машины для запуска werf в режиме Buildah.

### Драйвер хранилища

werf может использовать драйвер хранилища `overlay` или `vfs`:

* `overlay` позволяет использовать файловую систему OverlayFS. Можно использовать либо встроенную в ядро Linux поддержку OverlayFS (если она доступна), либо реализацию fuse-overlayfs. Это рекомендуемый выбор по умолчанию.
* `vfs` обеспечивает доступ к виртуальной файловой системе вместо OverlayFS. Эта файловая система уступает по производительности и требует привилегированного контейнера, поэтому ее не рекомендуется использовать. Однако в некоторых случаях она может пригодиться.

Как правило, достаточно использовать драйвер по умолчанию (`overlay`). Драйвер хранилища можно задать с помощью переменной окружения `WERF_BUILDAH_STORAGE_DRIVER`.

### Ulimits

По умолчанию Buildah режим в werf наследует системные ulimits при запуске сборочных контейнеров. Пользователь может переопределить эти параметры с помощью переменной окружения `WERF_BUILDAH_ULIMIT`.

Формат `WERF_BUILDAH_ULIMIT=type:softlimit[:hardlimit][,type:softlimit[:hardlimit],...]` — конфигурация лимитов, указанные через запятую:
 * "core": maximum core dump size (ulimit -c)
 * "cpu": maximum CPU time (ulimit -t)
 * "data": maximum size of a process's data segment (ulimit -d)
 * "fsize": maximum size of new files (ulimit -f)
 * "locks": maximum number of file locks (ulimit -x)
 * "memlock": maximum amount of locked memory (ulimit -l)
 * "msgqueue": maximum amount of data in message queues (ulimit -q)
 * "nice": niceness adjustment (nice -n, ulimit -e)
 * "nofile": maximum number of open files (ulimit -n)
 * "nproc": maximum number of processes (ulimit -u)
 * "rss": maximum size of a process's (ulimit -m)
 * "rtprio": maximum real-time scheduling priority (ulimit -r)
 * "rttime": maximum amount of real-time execution between blocking syscalls
 * "sigpending": maximum number of pending signals (ulimit -i)
 * "stack": maximum stack size (ulimit -s)

## Доступные образы werf

Ниже приведен список образов со встроенной утилитой werf. Каждый образ обновляется в рамках релизного процесса, основанного на менеджере пакетов trdl ([подробнее о каналах обновлений]({{ site.url }}/about/release_channels.html)).

* `registry.werf.io/werf/werf:latest` -> `registry.werf.io/werf/werf:1.2-stable`;
* `registry.werf.io/werf/werf:1.2-alpha` -> `registry.werf.io/werf/werf:1.2-alpha-alpine`;
* `registry.werf.io/werf/werf:1.2-beta` -> `registry.werf.io/werf/werf:1.2-beta-alpine`;
* `registry.werf.io/werf/werf:1.2-ea` -> `registry.werf.io/werf/werf:1.2-ea-alpine`;
* `registry.werf.io/werf/werf:1.2-stable` -> `registry.werf.io/werf/werf:1.2-stable-alpine`;
* `registry.werf.io/werf/werf:1.2-rock-solid` -> `registry.werf.io/werf/werf:1.2-rock-solid-alpine`;
* `registry.werf.io/werf/werf:1.2-alpha-alpine`;
* `registry.werf.io/werf/werf:1.2-beta-alpine`;
* `registry.werf.io/werf/werf:1.2-ea-alpine`;
* `registry.werf.io/werf/werf:1.2-stable-alpine`;
* `registry.werf.io/werf/werf:1.2-rock-solid-alpine`;
* `registry.werf.io/werf/werf:1.2-alpha-ubuntu`;
* `registry.werf.io/werf/werf:1.2-beta-ubuntu`;
* `registry.werf.io/werf/werf:1.2-ea-ubuntu`;
* `registry.werf.io/werf/werf:1.2-stable-ubuntu`;
* `registry.werf.io/werf/werf:1.2-rock-solid-ubuntu`;
* `registry.werf.io/werf/werf:1.2-alpha-fedora`;
* `registry.werf.io/werf/werf:1.2-beta-fedora`;
* `registry.werf.io/werf/werf:1.2-ea-fedora`;
* `registry.werf.io/werf/werf:1.2-stable-fedora`;
* `registry.werf.io/werf/werf:1.2-rock-solid-fedora`.

## Пример локальной сборки

```yaml
# werf.yaml
project: app
configVersion: 1
build:
  platform:
    - linux/amd64
    - linux/arm64
---
image: alpine
dockerfile: Dockerfile
```

```
# Dockerfile
FROM alpine
```
Сборка с помощью docker

```bash
werf build # или DOCKER_BUILDKIT=0 werf build
```

Сборка с помощью buildah

```bash
WERF_BUILDAH_MODE=auto werf build
```