---
title: Stapel
sidebar: documentation
permalink: documentation/development/stapel.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

## Обзор

Stapel — это [LFS-дистрибутив](http://www.linuxfromscratch.org/lfs/view/stable) Linux, который содержит:

* Glibc;
* Gnu cli tools (install, patch, find, wget, grep, rsync и другие);
* Git cli util;
* Bash;
* Python;
* Ansible.

Инструменты, содержащиеся в Stapel, расположены в нестандартных директориях. Исполняемые файлы, библиотеки и другие связанные файлы расположены в следующих директориях:

* `/.werf/stapel/(etc|lib|lib64|libexec)`;
* `/.werf/stapel/x86_64-lfs-linux-gnu/bin`;
* `/.werf/stapel/embedded/(bin|etc|lib|libexec|sbin|share|ssl)`.

В основе Stapel — библиотека Glibc и linker (`/.werf/stapel/lib/libc-VERSION.so` и `/.werf/stapel/x86_64-lfs-linux-gnu/bin/ld`). Все инструменты в директории `/.werf/stapel` скомпилированы и связаны только в библиотеками, находящимися в директории `/.werf/stapel`. Поэтому, Stapel — самодостаточный набор инструментов и библиотек, без каких-либо внешних зависимостей, с независимым Glibc. Это позволяет запускать инструменты Stapel в произвольном окружении (независимо от дистрибутива Linux и версий библиотек в этом дистрибутиве).

Файловая система Stapel (`/.werf/stapel`) спроектирована исходя из расчета что она будет монтироваться в сборочный контейнер, на основе какого-либо базового образа, после чего инструменты будут могут быть использованы. Так как в инструментах Stapel нет внешних зависимостей, образ Stapel может быть смонтирован в любой базовый образ (Alpine Linux + musl libc, или  Ubuntu + glibc — не важно) и инструменты будут работать одинаково.

werf монтирует _образ Stapel_ в каждый сборочный контейнер во время процесса сборки Docker-образа _сборщиком Stapel_. Это делает доступным работу Ansible, выполнение операций с Git и других важных функций. Читайте подробнее о _сборщике Stapel_ в соответствующей [статье]({{ site.baseurl }}/documentation/reference/build_process.html#stapel-образ-и-stapel-артефакт).

## Обновление Stapel

Образ Stapel периодически требует обновления, например для обновления версии Ansible или версии [LFS-дистрибутива](http://www.linuxfromscratch.org/lfs/view/stable) Linux.

Для обновления образа Stapel необходимо выполнить следующие шаги:

1. Внесите соответствующие изменения в сборочные инструкции в директории `stapel`.
2. Обновите `omnibus bundle`:
    ```bash
    cd stapel/omnibus
    bundle update
    git add -p Gemfile Gemfile.lock
    ```
3. Скачайте кэш образов Stapel предыдущих версий:
    ```bash
    scripts/stapel/pull_cache.sh
    ```
4. Измените текущую версию Stapel в скрипте `scripts/stapel/version.sh`, увеличив значение в переменной окружения `CURRENT_STAPEL_VERSION`.
5. Выполните сборку новых образов Stapel:
    ```bash
    scripts/stapel/build.sh
    ```
6. Опубликуйте собранные образы:
    ```bash
    scripts/stapel/publish.sh
    ```
7. После того, как новая версия Stapel опубликована, измените значение переменной окружения `PREVIOUS_STAPEL_VERSION` в скрипте `scripts/stapel/version.sh` на то-же самое значение, которое вы указали в переменной `CURRENT_STAPEL_VERSION` и выполните коммит изменений в репозиторий.

## Как работает сборка

Сборка образа Stapel состоит из двух Docker-стадий: base и final.

Стадия base — образ на основе Ubuntu, используемый для сборки LFS-системы и дополнительных omnibus-пакетов.

Стадия final — чистый scratch-образ, в который импортируются только файлы, необходимые в папке `/.werf/stapel`.

ЗАМЕЧАНИЕ: Обе Docker-стадии, — и base и final, публикуются с помощью скрипта `scripts/stapel/publish.sh` дла того, чтобы сделать доступным кэширование при следующей сборке образа Stapel.
