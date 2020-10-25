---
permalink: documentation/index.html
sidebar: documentation
---


Начните с **[Введения]({{ site.baseurl }}/introduction.html)**.
 > Ознакомьтесь с основами werf.
 > (5 мин)

Затем переходите к **[Установке]({{ site.baseurl }}/installation.html)** и  **[Инструкции по началу работы]({{ site.baseurl }}/quickstart.html)**.
 > Установите werf и изучите его возможности, развернув демо-приложение.
 > (15 мин)

**[Использование werf в системах CI/CD]({{ site.baseurl }}/documentation/using_with_ci_cd_systems.html)**
 > Изучите основы применения werf в любых CI/CD-системах.
 > (15 мин)

Раздел **[Руководства]({{ site.baseurl }}/documentation/guides.html)** содержит массу информации о настройке deployment'а для приложениея.
 > Здесь можно найти руководство, подходящее для вашего проекта (по языку программирования, фреймворку, системе CI/CD и т.п.) и развернуть первой настоящее приложение в кластер Kubernetes с помощью werf.
 > (несколько часов)

**[Справочник]({{ site.baseurl }}/documentation/reference/werf_yaml.html)** содержит структурированную информации о конфигурировании werf и его командах.
 - Чтобы использовать werf, конфигурацию приложения необходимо описать в файле [`werf.yaml`]({{ site.baseurl }}/documentation/reference/werf_yaml.html).
 - werf также использует [аннотации]({{ site.baseurl }}/documentation/reference/deploy_annotations.html) в определениях ресурсов для изменения поведения механизма оотслеживания ресурсов в процессе выката.
 - [Интерфейс командной строки]({{ site.baseurl }}/documentation/reference/cli/index.html) содержит полный список команд werf с описанием.

<!-- The **[Local development]()** section describes how werf simplifies and facilitates the local development of your applications, allowing you to use the same configuration to deploy an application either locally or remotely (into production). -->

Раздел **["Документация продвинутого уровня]({{ site.baseurl }}/documentation/advanced/configuration/supported_go_templates.html)** описывает более сложные задачи, с которыми вы можете рано или поздно столкнуться.
 - [Конфигурация]({{ site.baseurl }}/documentation/advanced/configuration/supported_go_templates.html) рассказывает и принципах шаблонизации конфигов werf и генерации связанных с развертыванием имен (таких как пространства имен Kubernetes или названия релизов).
 - [Helm]({{ site.baseurl }}/documentation/advanced/helm/basics.html)** повествует об основах деплоя: как настраивать werf, что такое helm-чарт и релиз. Здесь можно узнать об основах шаблонизации Kubernetes-ресурсов и способах использования собранных образов, описанных в файле `werf.yaml`, во время деплоя. Также уделяется внимание работе с секретами и приводится различная полезная информация. Этот раздел рекомендуется к прочтению тем, кто желает больше узнать об организации процесса деплоя с помощью werf.
 - [Очистка]({{ site.baseurl }}/documentation/advanced/cleanup.html) - в этом разделе рассказвается о концепции процесса очистки в werf и приводятся основные команды для выполнения очистки.
 - [CI/CD]({{ site.baseurl }}/documentation/advanced/ci_cd/ci_cd_workflow_basics.html) — описываются ключевые аспекты организации рабочих процессов в рамках CI/CD с помощью werf. Здесь вы узнаете об использовании werf в GitLab CI/CD, GitHub Actions и других CI/CD системах.
 - [Сборка образов с помощью Stapel]({{ site.baseurl }}/documentation/advanced/building_images_with_stapel/naming.html) рассказывает о кастомном сборщике introduces werf под названием Stapel. Интегрированный в него алгоритм распределенной сборки позволяет организовывать пайплайны сборки, отличающиеся чрезвычайно высокой скоростью работы, с применением распределенного кэширования и инкрементными ре-билдами, базирующимися на Git-истории вашего приложения.
 - [Разработка и отладка]({{ site.baseurl }}/documentation/advanced/development_and_debug/stage_introspection.html) повествует об отладке процессов сборки и развертывания приложения в случае, когда что-то пошло не так. Здесь же приводятся инструкции о настройке локальной среду разработки.
 - [Поддреживаемые реализации реестров]({{ site.baseurl }}/documentation/advanced/supported_registry_implementations.html) — приводятся общие сведения о поддерживаемых реализациях реестров и рассказывается об авторизации.

Раздел **[Внутренности]({{ site.baseurl }}/documentation/internals/building_of_images/build_process.html)** содержит информацию о внутренних механизмах работы werf. Для полноценного применения werf ознакомление с этим разделом не требуется, однако он будет полезен тем, что хочет больше узнать об устройстве и принципах работы нашего инструмента.
 - [Сборка образов]({{ site.baseurl }}/documentation/internals/building_of_images/build_process.html) — рассказзывается о том, что такое сборщик образов и стадии, как работает хранилище стадий, что такое сервер синхронизации, а также приводится другая информация, связанная с процессом сборки.
 - [Как работает интеграция с CI/CD]({{ site.baseurl }}/documentation/internals/how_ci_cd_integration_works/general_overview.html).
 - [Преобразование имен в werf]({{ site.baseurl }}/documentation/internals/names_slug_algorithm.html) — описывается алгоритм, который werf использует для автоматического преобразования имен и замены недопустимых символов (например, в namespace'ах Kubernetes или именах Helm-релизов).
 - [Интеграция с SSH-агентом]({{ site.baseurl }}/documentation/internals/integration_with_ssh_agent.html) — показано, как интегрировать SSH-агент в процесс сборки в werf.
 - [Для разработчиков]({{ site.baseurl }}/documentation/internals/development/stapel_image.html) — этот раздел для разработчиков содержит руководства по обслуживанию/поддержке и другую документацию, написанную разработчиками werf для разработчиков werf. Здесь можно узнать, как работают определенные подсистемы werf, как поддерживать субсистему в актуальном состоянии, как писать и собирать новый код для werf, и т.п.
