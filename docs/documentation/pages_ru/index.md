---
title: Обзор
without_auto_heading: true
permalink: index.html
description: Обширная и понятная документация по werf
editme_button: false
---

<link rel="stylesheet" type="text/css" href="{{ assets["overview.css"].digest_path | true_relative_url }}" />
<link rel="stylesheet" type="text/css" href="/css/guides.css" />

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
                <a class="overview__step-action" href="/introduction.html">Введение</a>
            </div>
        </div>
        <div class="overview__step">
            <div class="overview__step-header">
                <div class="overview__step-num">2</div>
                <div class="overview__step-time">15 минут</div>
            </div>
            <div class="overview__step-title">Установи werf и изучи его возможности, развернув демо-приложение</div>
            <div class="overview__step-actions">
                <a class="overview__step-action" href="/installation.html">Установка</a>
                <a class="overview__step-action" href="{{ "quickstart.html" | true_relative_url }}">Быстрый старт</a>
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
            <a class="overview__step-action" href="{{ "using_with_ci_cd_systems.html" | true_relative_url }}">Использование werf в системах CI/CD</a>
        </div>
    </div>
    <div class="overview__step">
        <div class="overview__step-header">
            <div class="overview__step-num">4</div>
            <div class="overview__step-time">несколько часов</div>
        </div>
        <div class="overview__step-title">Найди руководство подходящее твоему проекту</div>
        <div class="overview__step-actions">
            <a class="overview__step-action" href="/guides.html">Руководства</a>
        </div>
        <div class="overview__step-info">
            Раздел содержит массу информации о настройке выката для приложений. Здесь можно найти руководство, подходящее для вашего проекта (по языку программирования, фреймворку, системе CI/CD и т.п.) и развернуть первое настоящее приложение в кластер Kubernetes с помощью werf.
        </div>
    </div>
    <!--#include virtual="/guides/includes/landing-tiles.html" -->
    <div class="overview__title">Справочник</div>
    <div class="overview__step">
        <div class="overview__step-title">Найди структурированную информацию о конфигурировании werf и его командах</div>
        <div class="overview__step-actions">
            <a class="overview__step-action" href="{{ "reference/werf_yaml.html" | true_relative_url }}">Справочник</a>
        </div>
        <div class="overview__step-info">
<div markdown="1">
 - Чтобы использовать werf, конфигурацию приложения необходимо описать в файле [`werf.yaml`]({{ "reference/werf_yaml.html" | true_relative_url }}).
 - werf также использует [аннотации]({{ "reference/deploy_annotations.html" | true_relative_url }}) в определениях ресурсов для изменения поведения механизма отслеживания ресурсов в процессе выката.
 - [Интерфейс командной строки]({{ "reference/cli/overview.html" | true_relative_url }}) содержит полный список команд werf с описанием.
</div>
        </div>
    </div>
    <div class="overview__title">Дополнительная миля</div>
    <div class="overview__step">
        <div class="overview__step-title">Получи глубокие знания, которые понадобятся по мере использования werf</div>
        <div class="overview__step-actions">
            <a class="overview__step-action" href="{{ "advanced/giterminism.html" | true_relative_url }}">Документация продвинутого уровня</a>
        </div>
        <div class="overview__step-info">
<div markdown="1">
 - [Гитерминизм]({{ "advanced/giterminism.html" | true_relative_url }}) рассказывает о том как реализован детерминизм с гит, какие он вводит ограничения и почему.
 - [Helm]({{ "advanced/helm/overview.html" | true_relative_url }})** повествует об основах деплоя: как настраивать werf, что такое helm-чарт и релиз. Здесь можно узнать об основах шаблонизации Kubernetes-ресурсов и способах использования собранных образов, описанных в файле `werf.yaml`, во время деплоя. Также уделяется внимание работе с секретами и приводится различная полезная информация. Этот раздел рекомендуется к прочтению тем, кто желает больше узнать об организации процесса деплоя с помощью werf.
 - [Очистка]({{ "advanced/cleanup.html" | true_relative_url }}) — в этом разделе рассказывается о концепции процесса очистки в werf и приводятся основные команды для выполнения очистки.
 - [CI/CD]({{ "advanced/ci_cd/ci_cd_workflow_basics.html" | true_relative_url }}) — описываются ключевые аспекты организации рабочих процессов в рамках CI/CD с помощью werf. Здесь вы узнаете об использовании werf в GitLab CI/CD, GitHub Actions и других CI/CD системах.
 - [Сборка образов с помощью Stapel]({{ "reference/werf_yaml.html#секция-image" | true_relative_url }}) рассказывает о кастомном сборщике introduces werf под названием Stapel. Интегрированный в него алгоритм распределенной сборки позволяет организовывать пайплайны сборки, отличающиеся чрезвычайно высокой скоростью работы, с применением распределенного кэширования и инкрементными пересборками, базирующимися на Git-истории вашего приложения.
 - [Разработка и отладка]({{ "advanced/development_and_debug/stage_introspection.html" | true_relative_url }}) повествует об отладке процессов сборки и развертывания приложения в случае, когда что-то пошло не так. Здесь же приводятся инструкции о настройке локальной среду разработки.
 - [Поддерживаемые container registries]({{ "advanced/supported_container_registries.html" | true_relative_url }}) содержит информацию об особенностях использования различных container registries.
</div>
        </div>
    </div>
    <div class="overview__step">
        <div class="overview__step-title">Узнай как werf работает внутри</div>
        <div class="overview__step-actions">
            <a class="overview__step-action" href="{{ "internals/build_process.html" | true_relative_url }}">Внутренние механизмы werf</a>
        </div>
        <div class="overview__step-info">
            <p>Для полноценного применения werf ознакомление с этим разделом не требуется, однако он будет полезен тем, что хочет больше узнать об устройстве и принципах работы инструмента.</p>
<div markdown="1">
 - [Сборка образов]({{ "internals/build_process.html" | true_relative_url }}) — рассказывается о том, что такое сборщик образов и стадии, как работает хранилище стадий, что такое сервер синхронизации, а также приводится другая информация, связанная с процессом сборки.
 - [Как работает интеграция с CI/CD]({{ "internals/how_ci_cd_integration_works/general_overview.html" | true_relative_url }}).
 - [Преобразование имен в werf]({{ "internals/names_slug_algorithm.html" | true_relative_url }}) — описывается алгоритм, который werf использует для автоматического преобразования имен и замены недопустимых символов (например, в namespace Kubernetes или именах Helm-релизов).
 - [Интеграция с SSH-агентом]({{ "internals/integration_with_ssh_agent.html" | true_relative_url }}) — показано, как интегрировать SSH-агент в процесс сборки в werf.
 - [Для разработчиков]({{ "internals/development/stapel_image.html" | true_relative_url }}) — этот раздел для разработчиков содержит руководства по обслуживанию/поддержке и другую документацию, написанную разработчиками werf для разработчиков werf. Здесь можно узнать, как работают определенные подсистемы werf, как поддерживать субсистему в актуальном состоянии, как писать и собирать новый код для werf, и т.п.
</div>
        </div>
    </div>
</div>
