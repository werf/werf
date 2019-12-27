---
title: Stapel
sidebar: documentation
permalink: documentation/development/stapel.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

## Обзор

Stapel — это [LFS-дистрибутив](http://www.linuxfromscratch.org/lfs/view/stable) Linux, который содержит:

* Glibc
* Gnu cli tools (install, patch, find, wget, grep, rsync и другие)
* Git cli util
* Bash
* Python
* Ansible

Инструменты, содержащиеся в Stapel, расположены в нестандартных директориях.
Исполняемые файлы, библиотеки и другие связанные файлы расположены в следующих директориях:

* `/.werf/stapel/(etc|lib|lib64|libexec)`
* `/.werf/stapel/x86_64-lfs-linux-gnu/bin`
* `/.werf/stapel/embedded/(bin|etc|lib|libexec|sbin|share|ssl)`

В основе Stapel лежат библиотеки Glibc и linker (`/.werf/stapel/lib/libc-VERSION.so` и `/.werf/stapel/x86_64-lfs-linux-gnu/bin/ld`).
Все инструменты в директории `/.werf/stapel` скомпилированы и связаны только с библиотеками, находящимися в директории `/.werf/stapel`.
Поэтому, Stapel — самодостаточный набор инструментов и библиотек без каких-либо внешних зависимостей, с независимым Glibc.
Это позволяет запускать инструменты Stapel в произвольном окружении, независимо от дистрибутива Linux и версий библиотек в этом дистрибутиве.

Файловая система Stapel (`/.werf/stapel`) спроектирована исходя из расчета, что она будет монтироваться в сборочный контейнер.
Так как в инструментах Stapel нет внешних зависимостей, образ Stapel может быть смонтирован в любой базовый образ (Alpine Linux + musl libc, или  Ubuntu + glibc — не важно) и инструменты будут работать одинаково.

werf монтирует _образ Stapel_ в каждый сборочный контейнер во время процесса сборки Docker-образа _сборщиком Stapel_.
Это делает доступным работу Ansible, выполнение операций с Git и других важных функций.
Читайте подробнее о _сборщике Stapel_ в соответствующей [статье]({{ site.baseurl }}/documentation/reference/build_process.html#stapel-образ-и-stapel-артефакт).

## Обновление Stapel

Образ Stapel требует периодического обновления, например, для обновления версии Ansible или версии [LFS-дистрибутива](http://www.linuxfromscratch.org/lfs/view/stable) Linux.

Для обновления образа Stapel необходимо выполнить следующие шаги:

1.  Внести соответствующие изменения в сборочные инструкции в директории `stapel`.
2.  Обновить `omnibus bundle`:
    ```shell
    cd stapel/omnibus
    bundle update
    git add -p Gemfile Gemfile.lock
    ```
3.  Выполнить сборку новых образов Stapel:
    ```shell
    scripts/stapel/build.sh
    ```
    Данная команда создаст в системе Docker-образ `flant/werf-stapel:dev` и вспомогательный Docker-образ `flant/werf-stapel-base:dev`.
4.  Для того, чтобы протестировать этот свежесобранный образ Stapel надо объявить переменную окружения `WERF_STAPEL_IMAGE_VERSION=dev` перед запуском команд werf:
    ```shell
    export WERF_STAPEL_IMAGE_VERSION=dev
    werf build ...
    ```
5.  Опубликуйте новые образы Stapel:
    ```shell
    scripts/stapel/publish.sh NEW_VERSION
    ```
6.  После того, как новая версия Stapel опубликована, надо изменить значение Go-константы `VERSION` в файле `pkg/stapel/stapel.go` на новую версию и пересобрать werf.
