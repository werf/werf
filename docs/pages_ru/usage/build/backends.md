---
title: Сборочные бэкенды
permalink: usage/build/backends.html
---

> ПРИМЕЧАНИЕ: werf поддерживает сборку образов с _использованием Docker-сервера_ или _с использованием Buildah_. Поддерживается сборка как Dockerfile-образов, так и stapel-образов через Buildah.

## Docker

Docker - сборочный бэкенд по умолчанию.

### Подготовка системы к кроссплатформенной сборке (опционально)

> Данный шаг требуется только для сборки образов для платформ, отличных от платформы системы, где запущен werf.

Регистрируем в системе эмуляторы с помощью образа qemu-user-static:

```bash
docker run --restart=always --name=qemu-user-static -d --privileged --entrypoint=/bin/sh multiarch/qemu-user-static -c "/register --reset -p yes && tail -f /dev/null"
```

### Использование зеркал для docker.io

Если вы используете Docker для сборки, добавьте `registry-mirrors` в файл `/etc/docker/daemon.json`:

```json
{
  "registry-mirrors": ["https://<my-docker-io-mirror-host>"]
}
```

После этого перезапустите Docker и разлогиньтесь из `docker.io` с помощью команды:

```bash
werf cr logout
```

### Переопределение директории хранилища Docker

Параметр `--docker-server-storage-path` (или переменная окружения `WERF_DOCKER_SERVER_STORAGE_PATH`) позволяет явно задать директорию хранилища Docker в случае, если werf не может правильно определить её автоматически.

### Изменение порога занимаемого места и глубины очистки хранилища Docker

Параметр `--allowed-docker-storage-volume-usage` (`WERF_ALLOWED_DOCKER_STORAGE_VOLUME_USAGE`) позволяет изменить порог занимаемого места на томе, при достижении которого выполняется очистка хранилища Docker (по умолчанию 70%).

Параметр `--allowed-docker-storage-volume-usage-margin` (`WERF_ALLOWED_DOCKER_STORAGE_VOLUME_USAGE_MARGIN`) позволяет установить глубину очистки относительно установленного порога занимаемого места хранилища Docker (по умолчанию 5%).

### Изменение порога занимаемого места и глубины очистки локального кэша

Параметр `--allowed-local-cache-volume-usage` (`WERF_ALLOWED_LOCAL_CACHE_VOLUME_USAGE`) позволяет изменить порог занимаемого места на томе, при достижении которого выполняется очистка локального кэша (по умолчанию 70%).

Параметр `--allowed-docker-storage-volume-usage-margin` (`WERF_ALLOWED_LOCAL_CACHE_VOLUME_USAGE_MARGIN`) позволяет установить глубину очистки относительно установленного порога занимаемого места локального кэша (по умолчанию 5%).

## Buildah

Для сборки без Docker-сервера werf использует встроенный Buildah.

Buildah включается установкой переменной окружения `WERF_BUILDAH_MODE` в один из вариантов: `auto`, `native-chroot`, `native-rootless`.
Можно переключить сборочный бэкенд на Buildah, указав переменную окружения `WERF_BUILDAH_MODE=auto`

```bash
# Переключение на Buildah
export WERF_BUILDAH_MODE=auto
```

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

### Использование зеркал для docker.io

Если вы используете Buildah для сборки, вместо редактирования `daemon.json` используйте опцию `--container-registry-mirror` для werf-команд. Например:

```shell
werf build --container-registry-mirror=mirror.gcr.io
```

Для добавления нескольких зеркал используйте опцию `--container-registry-mirror` несколько раз.

Помимо опции, поддерживается использование переменных окружения `WERF_CONTAINER_REGISTRY_MIRROR_*`, например:

```bash
export WERF_CONTAINER_REGISTRY_MIRROR_GCR=mirror.gcr.io
export WERF_CONTAINER_REGISTRY_MIRROR_LOCAL=docker.mirror.local
```


## Мультиплатформенная и кроссплатформенная сборка

werf позволяет собирать образы как для родной архитектуры хоста, где запущен werf, так и в кроссплатформенном режиме с помощью эмуляции целевой архитектуры, которая может быть отлична от архитектуры хоста. Также werf позволяет собрать образ сразу для множества целевых платформ.

Мультиплатформенная сборка использует механизмы кроссплатформенного исполнения инструкций, предоставляемые [ядром Linux](https://en.wikipedia.org/wiki/Binfmt_misc) и эмулятором QEMU. [Перечень поддерживаемых архитектур](https://www.qemu.org/docs/master/about/emulation.html). Подготовка хост-системы для мультиплатформенной сборки рассмотрена [в разделе установки werf](https://ru.werf.io/getting_started/)

Поддержка мультиплатформенной сборки для разных вариантов синтаксиса сборки, режимов сборки и используемого бекенда:

|                         | buildah            | docker-server        |
|---                      |---                 |---                   |
| **Dockerfile**          | полная поддержка   | полная поддержка     |
| **staged Dockerfile**   | полная поддержка   | не поддерживается    |
| **stapel**              | полная поддержка   | только linux/amd64   |

### Сборка образов под одну целевую платформу

По умолчанию в качестве целевой используется платформа хоста, где запущен werf. Выбор другой целевой платформы для собираемых образов осуществляется с помощью параметра `--platform`:

```shell
werf build --platform linux/arm64
```

— все конечные образы, указанные в werf.yaml, будут собраны для указанной платформы с использованием эмуляции.

Целевую платформу можно также указать директивой конфигурации `build.platform`:

```yaml
# werf.yaml
project: example
configVersion: 1
build:
  platform:
  - linux/arm64
---
image: frontend
dockerfile: frontend/Dockerfile
---
image: backend
dockerfile: backend/Dockerfile
```

В этом случае запуск `werf build` без параметров вызовет сборку образов для указанной платформы (при этом явно указанный параметр `--platform` переопределяет значение из werf.yaml).

### Сборка образов под множество целевых платформ

Поддерживается и сборка образов сразу для набора архитектур. В этом случае в container registry публикуется манифест включающий в себя собранные образы под каждую из указанных платформ (во время скачивания такого образа автоматически будет выбираться образ под требуемую архитектуру).

Можно определить общий список платформ для всех образов в werf.yaml с помощью конфигурации:

```yaml
# werf.yaml
project: example
configVersion: 1
build:
  platform:
  - linux/arm64
  - linux/amd64
  - linux/arm/v7
```

Можно определить список целевых платформ отдельно для каждого собираемого образа (такая настройка будет иметь приоритет над общим списком определённым в werf.yaml):

```yaml
# werf.yaml
project: example
configVersion: 1
---
image: mysql
dockerfile: ./Dockerfile.mysql
platform:
- linux/amd64
---
image: backend
dockerfile: ./Dockerfile.backend
platform:
- linux/amd64
- linux/arm64
```

Общий список можно также переопределить параметром `--platform` непосредственно в момент вызова сборки:

```shell
werf build --platform=linux/amd64,linux/i386
```

— такой параметр переопределяет список целевых платформ указанных в werf.yaml (как общих, так и для отдельных образов).

## Автоматическая очистка хоста

Очистка хоста удаляет неактуальных данные и сокращает размер кеша **автоматически** в рамках вызова основных команд werf и сразу **для всех проектов**. При необходимости очистку можно выполнять в ручном режиме с помощью команды [**werf host cleanup**]({{ "reference/cli/werf_host_cleanup.html" | true_relative_url }}).

## Выключение автоматической очистки

Пользователь может выключить автоматическую очистку неактуальных данных хоста с помощью параметра `--disable-auto-host-cleanup` (`WERF_DISABLE_AUTO_HOST_CLEANUP`). В этом случае рекомендуется добавить команду `werf host cleanup` в cron, например, следующим образом:

```shell
# /etc/cron.d/werf-host-cleanup
SHELL=/bin/bash
*/30 * * * * gitlab-runner source ~/.profile ; source $(trdl use werf 2 stable) ; werf host cleanup
```