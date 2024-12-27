---
title: Сборочные бэкенды
permalink: usage/build/backends.html
---

## Обзор

werf позволяет работать со следующими сборочными бэкендами:

-	Docker — традиционный способ, использующий системный Docker Daemon.
-	Buildah — безопасная сборка без демона, поддерживает rootless-режим и полностью встроен в werf.

> Необходимые требования и подготовка системы для работы со сборочными бэкендами описаны в разделе сайта [Быстрый старт]({{ site.url }}/getting_started/).

## Buildah

> На данный момент Buildah доступен только для пользователей Linux и Windows с включённой подсистемой WSL2. Для пользователей MacOS на данный момент предлагается использование виртуальной машины для запуска werf в режиме Buildah.

Buildah включается установкой переменной окружения `WERF_BUILDAH_MODE` в один из вариантов: `auto`, `native-chroot`, `native-rootless`.

```bash
# Переключение на Buildah
export WERF_BUILDAH_MODE=auto
```

* `auto` — автоматический выбор режима в зависимости от платформы и окружения;
* `native-chroot` использует `chroot`-изоляцию для сборочных контейнеров;
* `native-rootless` использует `rootless`-изоляцию для сборочных контейнеров. На этом уровне изоляции werf использует среду выполнения сборочных операций в контейнерах (runc, crun, kata или runsc).

### Драйвер хранилища

werf может использовать драйвер хранилища `overlay` или `vfs`:

* `overlay` позволяет использовать файловую систему OverlayFS. Можно использовать либо встроенную в ядро Linux поддержку OverlayFS (если она доступна), либо реализацию fuse-overlayfs. Это рекомендуемый выбор по умолчанию.
* `vfs` обеспечивает доступ к виртуальной файловой системе вместо OverlayFS. Эта файловая система уступает по производительности и требует привилегированного контейнера, поэтому ее не рекомендуется использовать. Однако в некоторых случаях она может пригодиться.

Как правило, достаточно использовать драйвер по умолчанию (`overlay`). Драйвер хранилища можно задать с помощью переменной окружения `WERF_BUILDAH_STORAGE_DRIVER`.

### Ulimits

По умолчанию Buildah режим в werf наследует системные ulimits при запуске сборочных контейнеров. Пользователь может переопределить эти параметры с помощью переменной окружения `WERF_BUILDAH_ULIMIT`.

Формат `WERF_BUILDAH_ULIMIT=type:softlimit[:hardlimit][,type:softlimit[:hardlimit],...]` — конфигурация лимитов, указанные через запятую:

* `core`: maximum core dump size (ulimit -c).
* `cpu`: maximum CPU time (ulimit -t).
* `data`: maximum size of a process's data segment (ulimit -d).
* `fsize`: maximum size of new files (ulimit -f).
* `locks`: maximum number of file locks (ulimit -x).
* `memlock`: maximum amount of locked memory (ulimit -l).
* `msgqueue`: maximum amount of data in message queues (ulimit -q).
* `nice`: niceness adjustment (nice -n, ulimit -e).
* `nofile`: maximum number of open files (ulimit -n).
* `nproc`: maximum number of processes (ulimit -u).
* `rss`: maximum size of a process's resident set size (ulimit -m).
* `rtprio`: maximum real-time scheduling priority (ulimit -r).
* `rttime`: maximum real-time execution between blocking syscalls.
* `sigpending`: maximum number of pending signals (ulimit -i).
* `stack`: maximum stack size (ulimit -s).
