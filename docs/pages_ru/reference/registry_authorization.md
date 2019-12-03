---
title: Авторизация в Docker Registry
sidebar: documentation
permalink: documentation/reference/registry_authorization.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

Некоторые категории команд работают с Docker Registry, и требуют соответствующей авторизации:
* [Во время процесса сборки]({{ site.baseurl }}/documentation/reference/build_process.html) Werf может делать pull образов из Docker Registry.
* [Во время процесса публикации]({{ site.baseurl }}/documentation/reference/publish_process.html) Werf создает и обновляет образы в Docker Registry.
* [Во время процесса очистки]({{ site.baseurl }}/documentation/reference/cleaning_process.html) Werf удаляет образы из Docker Registry.
* [Во время процесса деплоя]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html) Werf требует доступа к _образам_ в Docker Registry и _стадиям_, которые также могут находиться в Docker Registry.

Все команды, требующие авторизации в Docker Registry, не выполняют ее сами, а используют подготовленную _конфигурацию Docker_.

_Конфигурация Docker_ — это папка, в которой хранятся данные авторизации используемые для доступа вразличные Docker Registry и другие настройки Docker.
По умолчанию, Werf использует стандартную для Docker папку конфигурации: `~/.docker`. Другую используемую папку конфигурации можно указать с помощью параметра `--docker-config`, либо с помощью переменных окружения `$DOCKER_CONFIG` или `$WERF_DOCKER_CONFIG`. Все параметры и опции в файле конфигурации стандартны для Docker, их список можно посмотреть с помощью команды `docker --config`.

Для подготовки конфигурации Docker вы можете использовать команду `docker login`, либо, если вы выполняете Werf в рамках CI-системы, вызвать команду [werf ci-env]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html)  (более подробно о подключении Werf к CI-системам читай в [соответствующем разделе]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html)).

> Использование `docker login` при параллельном выполнении заданий в CI-системе может приводить к ошибкам выполнения заданий из-за работы с временными правами и состояния race condition (одно задание влиет на другое, переопределяя конфигурацию Docker). Поэтому, необходимо обеспечивать независимую конфигурацию Docker между заданиями, используя `docker --config` или `werf ci-env`
