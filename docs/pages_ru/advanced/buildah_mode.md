---
title: Режим сборки с использованием Buildah
permalink: advanced/buildah_mode.html
---

> ПРИМЕЧАНИЕ: werf поддерживает сборку образов с _использованием Docker-сервера_ или _с использованием Buildah_. Поддерживается сборка как Dockerfile-образов, так и stapel-образов через Buildah.

Для сборки без Docker-сервера werf использует встроенный Buildah в rootless-режиме.

## Системные требования

Требования к хост-системе для запуска werf в Buildah-режиме без Docker/Kubernetes можно найти в [инструкциях по установке](/installation.html). А для запуска werf в Kubernetes или в Docker-контейнерах требования следующие:
* Если ваше ядро Linux версии 5.13+ (в некоторых дистрибутивах 5.11+), то убедитесь, что модуль ядра `overlay` загружен с `lsmod | grep overlay`. Если ядро более старое или у вас не получается активировать модуль ядра `overlay`, то установите `fuse-overlayfs`, который обычно доступен в репозиториях вашего дистрибутива. В крайнем случае может быть использован драйвер хранилища `vfs`.
* Команда `sysctl kernel.unprivileged_userns_clone` должна вернуть `1`. В ином случае выполните:
  ```shell
  echo 'kernel.unprivileged_userns_clone = 1' | sudo tee -a /etc/sysctl.conf
  sudo sysctl -p
  ```
* Команда `sysctl user.max_user_namespaces` должна вернуть по меньшей мере `15000`. В ином случае выполните:
  ```shell
  echo 'user.max_user_namespaces = 15000' | sudo tee -a /etc/sysctl.conf
  sudo sysctl -p
  ```

## Включение Buildah

Buildah включается установкой переменной окружения `WERF_BUILDAH_MODE` в один из вариантов: `auto`, `native-chroot`, `native-rootless` или `docker-with-fuse`.

* `auto` — автоматический выбор режима в зависимости от платформы и окружения;
* `native-chroot` работает только в Linux и использует `chroot`-изоляцию для сборочных контейнеров;
* `native-rootless` работает только в Linux и использует `rootless`-изоляцию для сборочных контейнеров. На этом уровне изоляции werf использует среду выполнения сборочных операций в контейнерах (runc, crun, kata или runsc).

Большинству пользователей для включения экспериментального режима Buildah достаточно установить `WERF_BUILDAH_MODE=auto`.

> ПРИМЕЧАНИЕ: На данный момент Buildah доступен только для пользователей Linux и Windows с включённой подсистемой WSL2. Полная поддержка для пользователей MacOS запланирована, но пока не доступна. Для пользователей MacOS на данный момент предлагается использование виртуальной машины для запуска werf в режиме Buildah.

## Драйвер хранилища

werf может использовать драйвер хранилища `overlay` или `vfs`:

* `overlay` позволяет использовать файловую систему OverlayFS. Можно использовать либо встроенную в ядро Linux поддержку OverlayFS (если она доступна), либо реализацию fuse-overlayfs. Это рекомендуемый выбор по умолчанию.
* `vfs` обеспечивает доступ к виртуальной файловой системе вместо OverlayFS. Эта файловая система уступает по производительности и требует привилегированного контейнера, поэтому ее не рекомендуется использовать. Однако в некоторых случаях она может пригодиться.

Как правило, достаточно использовать драйвер по умолчанию (`overlay`). Драйвер хранилища можно задать с помощью переменной окружения `WERF_BUILDAH_STORAGE_DRIVER`.

## Ulimits

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
