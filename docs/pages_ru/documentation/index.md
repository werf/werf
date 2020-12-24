---
title: Обзор
without_auto_heading: true
permalink: documentation/index.html
description: Обширная и понятная документация по werf
sidebar: documentation
---

{%- asset overview.css %}

<div class="overview">
    <div class="overview__title">Обязательно к прочтению</div>
    <div class="overview__row">
        <div class="overview__step">
            <div class="overview__step-header">
                <div class="overview__step-num">1</div>
                <div class="overview__step-time">5 минут</div>
            </div>
            <div class="overview__step-title">Начни с изучения основ</div>
            <div class="overview__step-actions">
                <a class="overview__step-action" href="{{ "introduction.html" | true_relative_url: page.url }}">Введение</a>
            </div>
        </div>
        <div class="overview__step">
            <div class="overview__step-header">
                <div class="overview__step-num">2</div>
                <div class="overview__step-time">15 минут</div>
            </div>
            <div class="overview__step-title">Установи werf и изучи его возможности, развернув демо-приложение</div>
            <div class="overview__step-actions">
                <a class="overview__step-action" href="{{ "installation.html" | true_relative_url: page.url }}">Установка</a>
                <a class="overview__step-action" href="{{ "documentation/quickstart.html" | true_relative_url: page.url }}">Быстрый старт</a>
            </div>
        </div>
    </div>
    <div class="overview__step">
        <div class="overview__step-header">
            <div class="overview__step-num">3</div>
            <div class="overview__step-time">15 минут</div>
        </div>
        <div class="overview__step-title">Изучи основы применения werf в любых системах CI/CD</div>
        <div class="overview__step-actions">
            <a class="overview__step-action" href="{{ "documentation/using_with_ci_cd_systems.html" | true_relative_url: page.url }}">Использование werf в системах CI/CD</a>
        </div>
    </div>
    <div class="overview__step">
        <div class="overview__step-header">
            <div class="overview__step-num">4</div>
            <div class="overview__step-time">несколько часов</div>
        </div>
        <div class="overview__step-title">Найди руководство подходящиее твоему проекту</div>
        <div class="overview__step-actions">
            <a class="overview__step-action" href="{{ "documentation/guides.html" | true_relative_url: page.url }}">Руководства</a>
        </div>
        <div class="overview__step-info">
            Раздел содержит массу информации о настройке выката для приложений. Здесь можно найти руководство, подходящее для вашего проекта (по языку программирования, фреймворку, системе CI/CD и т.п.) и развернуть первое настоящее приложение в кластер Kubernetes с помощью werf.
        </div>
    </div>
    <div class="overview__title">Справочник</div>
    <div class="overview__step">
        <div class="overview__step-title">Найди структурированную информацию о конфигурировании werf и его командах</div>
        <div class="overview__step-actions">
            <a class="overview__step-action" href="{{ "documentation/reference/werf_yaml.html" | true_relative_url: page.url }}">Справочник</a>
        </div>
        <div class="overview__step-info">
<div markdown="1">
 - Чтобы использовать werf, конфигурацию приложения необходимо описать в файле [`werf.yaml`]({{ "documentation/reference/werf_yaml.html" | true_relative_url: page.url }}).
 - werf также использует [аннотации]({{ "documentation/reference/deploy_annotations.html" | true_relative_url: page.url }}) в определениях ресурсов для изменения поведения механизма оотслеживания ресурсов в процессе выката.
 - [Интерфейс командной строки]({{ "documentation/reference/cli/overview.html" | true_relative_url: page.url }}) содержит полный список команд werf с описанием.
</div>
        </div>
    </div>
    <div class="overview__title">Дополнительная миля</div>
    <div class="overview__step">
        <div class="overview__step-title">Получи глубокие знания, которые понадобятся по мере использования werf</div>
        <div class="overview__step-actions">
            <a class="overview__step-action" href="{{ "documentation/advanced/configuration/supported_go_templates.html" | true_relative_url: page.url }}">Документация продвинутого уровня</a>
        </div>
        <div class="overview__step-info">
<div markdown="1">
 - [Конфигурация]({{ "documentation/advanced/configuration/supported_go_templates.html" | true_relative_url: page.url }}) рассказывает и принципах шаблонизации конфигов werf и генерации связанных с развертыванием имен (таких как пространства имен Kubernetes или названия релизов).
 - [Helm]({{ "documentation/advanced/helm/basics.html" | true_relative_url: page.url }})** повествует об основах деплоя: как настраивать werf, что такое helm-чарт и релиз. Здесь можно узнать об основах шаблонизации Kubernetes-ресурсов и способах использования собранных образов, описанных в файле `werf.yaml`, во время деплоя. Также уделяется внимание работе с секретами и приводится различная полезная информация. Этот раздел рекомендуется к прочтению тем, кто желает больше узнать об организации процесса деплоя с помощью werf.
 - [Очистка]({{ "documentation/advanced/cleanup.html" | true_relative_url: page.url }}) - в этом разделе рассказвается о концепции процесса очистки в werf и приводятся основные команды для выполнения очистки.
 - [CI/CD]({{ "documentation/advanced/ci_cd/ci_cd_workflow_basics.html" | true_relative_url: page.url }}) — описываются ключевые аспекты организации рабочих процессов в рамках CI/CD с помощью werf. Здесь вы узнаете об использовании werf в GitLab CI/CD, GitHub Actions и других CI/CD системах.
 - [Сборка образов с помощью Stapel]({{ "documentation/reference/werf_yaml.html#секция-image" | true_relative_url: page.url }}) рассказывает о кастомном сборщике introduces werf под названием Stapel. Интегрированный в него алгоритм распределенной сборки позволяет организовывать пайплайны сборки, отличающиеся чрезвычайно высокой скоростью работы, с применением распределенного кэширования и инкрементными ре-билдами, базирующимися на Git-истории вашего приложения.
 - [Разработка и отладка]({{ "documentation/advanced/development_and_debug/stage_introspection.html" | true_relative_url: page.url }}) повествует об отладке процессов сборки и развертывания приложения в случае, когда что-то пошло не так. Здесь же приводятся инструкции о настройке локальной среду разработки.
 - [Поддерживаемые реализации container registry]({{ "documentation/advanced/supported_registry_implementations.html" | true_relative_url: page.url }}) — приводятся общие сведения о поддерживаемых реализациях container registry и рассказывается об авторизации.
</div>
        </div>
    </div>
    <div class="overview__step">
        <div class="overview__step-title">Узнай как werf работает внутри</div>
        <div class="overview__step-actions">
            <a class="overview__step-action" href="{{ "documentation/internals/build_process.html" | true_relative_url: page.url }}">Внутренние механизмы werf</a>
        </div>
        <div class="overview__step-info">
            <p>Для полноценного применения werf ознакомление с этим разделом не требуется, однако он будет полезен тем, что хочет больше узнать об устройстве и принципах работы инструмента.</p>
<div markdown="1">
 - [Сборка образов]({{ "documentation/internals/build_process.html" | true_relative_url: page.url }}) — рассказзывается о том, что такое сборщик образов и стадии, как работает хранилище стадий, что такое сервер синхронизации, а также приводится другая информация, связанная с процессом сборки.
 - [Как работает интеграция с CI/CD]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html" | true_relative_url: page.url }}).
 - [Преобразование имен в werf]({{ "documentation/internals/names_slug_algorithm.html" | true_relative_url: page.url }}) — описывается алгоритм, который werf использует для автоматического преобразования имен и замены недопустимых символов (например, в namespace'ах Kubernetes или именах Helm-релизов).
 - [Интеграция с SSH-агентом]({{ "documentation/internals/integration_with_ssh_agent.html" | true_relative_url: page.url }}) — показано, как интегрировать SSH-агент в процесс сборки в werf.
 - [Для разработчиков]({{ "documentation/internals/development/stapel_image.html" | true_relative_url: page.url }}) — этот раздел для разработчиков содержит руководства по обслуживанию/поддержке и другую документацию, написанную разработчиками werf для разработчиков werf. Здесь можно узнать, как работают определенные подсистемы werf, как поддерживать субсистему в актуальном состоянии, как писать и собирать новый код для werf, и т.п.
</div>
        </div>
    </div>
</div>
