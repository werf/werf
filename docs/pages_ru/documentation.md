---
title: Обзор
permalink: index.html
sidebar: documentation
---

Документация содержит более 100 различных статей, включая наиболее типичные примеры использования werf, подробное описание функций, архитектуры и параметров вызова.

Мы рекомендуем начинать знакомство с раздела [**Руководства**]({{ site.baseurl }}/guides/installation.html):

- [Установка]({{ site.baseurl }}/guides/installation.html) содержит зависимости и возможные варианты установки.
- [Первые шаги]({{ site.baseurl }}/guides/getting_started.html) помогает начать использовать werf с существующим Dockerfile. Вы можете легко запустить werf в вашем проекте прямо сейчас.
- [Деплой в Kubernetes]({{ site.baseurl }}/guides/deploy_into_kubernetes.html) — краткий пример развертывания приложения в кластере Kubernetes.
- [Интеграция с CI/CD-системами]({{ site.baseurl }}/guides/generic_ci_cd_integration.html) — общий подход к интеграции werf с любой CI/CD-системой.
- [Интеграция с GitLab CI]({{ site.baseurl }}/guides/gitlab_ci_cd_integration.html) расскажет всё об интеграции с GitLab CI: про сборку, публикацию, деплой и очистку Docker registry.
- [Интеграция с GitHub Actions]({{ site.baseurl }}/guides/github_ci_cd_integration.html) расскажет всё об интеграции с GitHub Actions: про сборку, публикацию, деплой и очистку образов.
- В разделе расширенной сборки рассказывается о нашем синтаксисе описания сборки образов. Синтаксис позволяет использовать werf сборщик, который учитывает особенности инкрементальной сборки и предоставляет дополнительные возможности (к примеру, описание сборочных инструкций Ansible-задачами). Рекомендуем начать знакомство с создания [первого приложения]({{ site.baseurl }}/guides/advanced_build/first_application.html).

Следующий раздел — [**Конфигурация**]({{ site.baseurl }}/configuration/introduction.html).

Для использования werf в вашем проекте, необходимо создать файл конфигурации `werf.yaml`, который может состоять из:

1. Описания метаинформации проекта, которая впоследствии будет использоваться в большинстве команд и влиять на конечный результат (к примеру, на кэши и формат имён Helm-релиза и namespace в Kubernetes). Пример такой метаинформации — имя проекта.
2. Описания образов для сборки.

В статье [**Общие сведения**]({{ site.baseurl }}/configuration/introduction.html) вы найдете информацию о:

* Структуре секций и их конфигурации
* Описанию конфигурации в нескольких файлах
* Этапах обработки конфигурации 
* Поддерживаемых функциях Go-шаблонов

В других статьях раздела [**Конфигурация**]({{ site.baseurl }}/configuration/introduction.html) дается детальная информация о директивах описания [Dockerfile-образа]({{ site.baseurl }}/configuration/dockerfile_image.html), [Stapel-образа]({{ site.baseurl }}/configuration/stapel_image/naming.html), [Stapel-артефакта]({{ site.baseurl }}/configuration/stapel_artifact.html) и особенностях их использования.

Раздел [**Справочник**]({{ site.baseurl }}/reference/stages_and_images.html) посвящен описанию основных процессов werf:

* [Сборка]({{ site.baseurl }}/reference/build_process.html)
* [Публикация]({{ site.baseurl }}/reference/publish_process.html)
* [Деплой]({{ site.baseurl }}/reference/deploy_process/deploy_into_kubernetes.html)
* [Очистка]({{ site.baseurl }}/reference/cleaning_process.html)

Каждая статья описывает определенный процесс, особенности и доступные опции.

Также в этот раздел включены статьи с описанием базовых примитивов и общих инструментов:

* [Стадии и образы]({{ site.baseurl }}/reference/stages_and_images.html)
* [Работа с Docker registries]({{ site.baseurl }}/reference/working_with_docker_registries.html)
* [Разработка и отладка]({{ site.baseurl }}/reference/development_and_debug/setup_minikube.html)
* [Toolbox]({{ site.baseurl }}/reference/toolbox/slug.html)

Раздел [**CLI Commands**]({{ site.baseurl }}/cli/main/build.html) содержит как базовые, необходимые для управления процессом CI/CD, так и служебные команды, обеспечивающие расширенные функциональные возможности.

Раздел [**Разработка**]({{ site.baseurl }}/development/stapel.html) содержит информацию, предназначенную для более глубокого понимания работы werf.
