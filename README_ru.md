<p align="center">
  <img src="https://github.com/werf/werf/raw/master/docs/images/werf-logo.svg?sanitize=true" style="max-height:100%;" height="175">
</p>

<p align="center">
  <a href="https://cloud-native.slack.com/messages/CHY2THYUU"><img src="https://img.shields.io/badge/slack-EN%20chat-611f69.svg?logo=slack" alt="Slack chat EN"></a>
  <a href="https://twitter.com/werf_io"><img src="https://img.shields.io/badge/twitter-EN-611f69.svg?logo=twitter" alt="Twitter EN"></a>
  <a href="https://t.me/werf_ru"><img src="https://img.shields.io/badge/telegram-RU%20chat-179cde.svg?logo=telegram" alt="Telegram chat RU"></a><br>
  <a href='https://bintray.com/flant/werf/werf/_latestVersion'><img src='https://api.bintray.com/packages/flant/werf/werf/images/download.svg'></a>
  <a href="https://pkg.go.dev/github.com/werf/werf"><img src="https://pkg.go.dev/github.com/werf/werf?status.svg" alt="GoDoc"></a>
  <a href="https://codeclimate.com/github/werf/werf/test_coverage"><img src="https://api.codeclimate.com/v1/badges/bac6f23d5c366c6324b5/test_coverage" /></a>
</p>
___

<!-- WERF DOCS PARTIAL BEGIN: Overview -->

**werf** — Open Source CLI-утилита, написанная на Go, предназначенная для упрощения и ускорения доставки вашего приложения.

Вам достаточно описать конфигурацию приложения, правила сборки и развертывания в Kubernetes, в Git-репозитории, едином источнике правды. Проще говоря, это то, что сегодня называется GitOps.

* Собирает Docker-образы, как используя Dockerfile, так и альтернативный сборщик с собственным синтаксисом, основная задача которого — сокращение времени инкрементальной сборки на основе истории Git.
* Поддерживает множество схем тегирования.
* Выкатывает приложение в Kubernetes, используя Helm-совместимый формат чартов с удобными настройками, улучшенным механизмом отслеживания процесса выката, обнаружения ошибок и выводом логов.
* Очищает Docker registry от неиспользуемых образов.

werf — не CI/CD-система, а инструмент для построения пайплайнов, который может использоваться в любой CI/CD-системе. Мы считаем инструменты такого рода новым поколением высокоуровневых инструментов CI/CD.

<!-- WERF DOCS PARTIAL END -->

**Содержание**

- [Возможности](#возможности)
- [Установка](#установка)
  - [Установка зависимостей](#установка-зависимостей)
  - [Установка werf](#установка-werf)
- [Первые шаги](#первые-шаги)
- [Обещание обратной совместимости](#обещание-обратной-совместимости)
- [Документация-и-поддержка](#документация-и-поддержка)
- [Лицензия](#лицензия)

# Возможности

<!-- WERF DOCS PARTIAL BEGIN: Features -->

- Управление полным жизненным циклом приложения: сборка и публикация образов, деплой приложений в Kubernetes и очистка неиспользуемых образов по политикам.
- Описание всех правил сборки и деплоя приложения, состоящего из любого количества компонентов, хранятся в одном Git-репозитории вместе с исходным кодом (SSOT).
- Сборка образов из Dockerfile.
- Инкрементальная сборка на основе истории Git, использование Ansible и другие возможности, реализованные в рамках альтернативного сборщика с собственным синтаксисом.   
- Использование совместимых с Helm 2 чартов, а также комплексный деплой с журналированием, трэкингом, ранним выявлением ошибок и использованием аннотаций ресурсов для тонкой настройки процесса деплоя.
- werf — это CLI-утилита, написанная на Go, которая может быть встроена в любую существующую систему CI/CD.

## Скоро

- ~3-х стороннее слияние (3-way-merge)~ [#1616](https://github.com/werf/werf/issues/1616).
- Локальная разработка приложений с werf [#1940](https://github.com/werf/werf/issues/1940).
- ~Тегирование, основанное на контенте~ [#1184](https://github.com/werf/werf/issues/1184).
- Лучшие практики и рецепты для наиболее популярных CI-систем [#1617](https://github.com/werf/werf/issues/1617).
- ~Поддержка большинства имлементаций Docker registry~ [#2199](https://github.com/werf/werf/issues/2199).
- Параллельная сборка образов [#2200](https://github.com/werf/werf/issues/2200).
- ~Распределенная сборка с общим Docker registry~ [#1614](https://github.com/werf/werf/issues/1614).
- Поддержка Helm 3 [#1606](https://github.com/werf/werf/issues/1606).
- Kaniko-подобная сборка без привязки к локальному Docker-демону [#1618](https://github.com/werf/werf/issues/1618).

## Полный список возможностей

### Сборка

- Удобная сборка произвольного числа образов в одном проекте.
- Сборка образов как из Dockerfile, так и из инструкций сборщика Stapel.
- Параллельные сборки на одном хосте (с использованием файловых блокировок).
- Распределенная сборка (скоро) [#1614](https://github.com/werf/werf/issues/1614).
- Параллельная сборка образов (скоро) [#2200](https://github.com/werf/werf/issues/2200).
- Расширенная сборка со сборщиком Stapel:
  - Инкрементальная пересборка на основе истории изменений Git.
  - Сборка образов с Shell-инструкциями и Ansible-заданиями.
  - Совместное использование кэша между сборками с использованием монтирования.
  - Уменьшение размера конечного образа за счёт изолирования исходного кода, инструментов сборки и кэша от результата.
- Сборка одного образа на основе другого, описанного в том же файле конфигурации.
- Инструменты отладки сборочного процесса.
- Подробный вывод.

### Публикация

- Хранение образов в одном или нескольких Docker-репозиториях согласно следующим шаблонам именования:
  - `IMAGES_REPO:[IMAGE_NAME-]TAG` в режиме `monorepo`.
  - `IMAGES_REPO[/IMAGE_NAME]:TAG` в режиме `multirepo`.
- Различные стратегии тегирования образов:
  - Тегирование образов по тегу, ветке или коммиту в Git.
  - Тегирование, основанное на контенте.

### Деплой

- Деплой в Kubernetes и отслеживание корректности выката приложения.
  - Отслеживание статуса всех ресурсов.
  - Контроль готовности ресурсов.
  - Управление контролем процесса деплоя с помощью аннотаций.
- Полный визуальный контроль как процесса деплоя, так и конечного результата.
  - Логирование и сообщение об ошибках.
  - Вывод периодического отчета о состоянии ресурсов в процессе деплоя.
  - Упрощенная отладка проблем без необходимости использовать kubectl.
- Завершение с ошибкой задания pipeline CI при обнаружении проблемы.
  - Раннее обнаружение ошибок деплоя ресурсов без необходимости ожидания таймаута.
- Полная совместимость с Helm 2.
- Возможность ограничения прав при развертывании с использованием механизма RBAC (Tiller встроен внутрь werf и его запуск выполняется от имени пользователя, выполняющего деплой).
- Параллельные сборки на одном хосте (с использованием файловых блокировок).
- Распределенный параллельный деплой (скоро) [#1620](https://github.com/werf/werf/issues/1620).
- Возможность непрерывной доставки образа с постоянным тегом (как пример, при использовании стратегии тегирования по веткам).

### Очистка

- Очистка локального хранилища и Docker registry по настраиваемым политикам.
- Очистка игнорирует используемые в Kubernetes-кластере образы. werf сканирует следующие типы объектов кластера: Pod, Deployment, ReplicaSet, StatefulSet, DaemonSet, Job, CronJob, ReplicationController.

<!-- WERF DOCS PARTIAL END -->

# Установка

## Установка зависимостей

<!-- WERF DOCS PARTIAL BEGIN: Installing dependencies -->

### Docker

[Руководство по установке Docker CE](https://docs.docker.com/install/).

Для работы с Docker-демоном пользователю необходимы соответствующие привилегии. Создайте группу **docker** и добавьте в неё пользователя:

```shell
sudo groupadd docker
sudo usermod -aG docker $USER
```

### Git

[Руководство по установке Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git).

- Минимально допустимая версия — 1.9.0.
- В случае использования [Git Submodule](https://git-scm.com/docs/gitsubmodules), минимально допустимая версия — 2.14.0.

<!-- WERF DOCS PARTIAL END -->

## Установка werf

Существует множество способов установки werf и большинство освещается в [Руководстве по установке](https://ru.werf.io/documentation/guides/installation.html). Далее будет рассмотрена установка с помощью [multiwerf](https://github.com/flant/multiwerf), рекомендованным способом как при локальной разработке, так и в CI. 

<!-- WERF DOCS PARTIAL BEGIN: Installing with multiwerf -->

#### Unix shell (sh, bash, zsh)

##### Установка multiwerf

```shell
# добавление ~/bin в PATH
export PATH=$PATH:$HOME/bin
echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

# установка multiwerf в директорию ~/bin
mkdir -p ~/bin
cd ~/bin
curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash
```

##### Добавление werf alias в текущую shell-сессию

```shell
. $(multiwerf use 1.1 stable --as-file)
```

##### Рекомендация при использовании в CI

Чтобы упростить отладку в CI-окружении, например, в случае, когда бинарный файл multiwerf не установлен или неисполняемый, рекомендуется использовать команду `type`:

```shell
type multiwerf && . $(multiwerf use 1.1 stable --as-file)
```

##### Опционально: добавление werf alias в shell-сессию при открытии терминала

```shell
echo '. $(multiwerf use 1.1 stable --as-file)' >> ~/.bashrc
```

#### Windows

##### PowerShell

###### Установка multiwerf

```shell
$MULTIWERF_BIN_PATH = "C:\ProgramData\multiwerf\bin"
mkdir $MULTIWERF_BIN_PATH

Invoke-WebRequest -Uri https://flant.bintray.com/multiwerf/v1.0.16/multiwerf-windows-amd64-v1.0.16.exe -OutFile $MULTIWERF_BIN_PATH\multiwerf.exe

[Environment]::SetEnvironmentVariable(
    "Path",
    [Environment]::GetEnvironmentVariable("Path", [EnvironmentVariableTarget]::Machine) + "$MULTIWERF_BIN_PATH",
    [EnvironmentVariableTarget]::Machine)

$env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")
```

###### Добавление werf alias в текущую shell-сессию

```shell
Invoke-Expression -Command "multiwerf use 1.1 stable --as-file --shell powershell" | Out-String -OutVariable WERF_USE_SCRIPT_PATH
. $WERF_USE_SCRIPT_PATH.Trim()
```

##### cmd.exe

###### Установка multiwerf

```shell
set MULTIWERF_BIN_PATH="C:\ProgramData\multiwerf\bin"
mkdir %MULTIWERF_BIN_PATH%
bitsadmin.exe /transfer "multiwerf" https://flant.bintray.com/multiwerf/v1.0.16/multiwerf-windows-amd64-v1.0.16.exe %MULTIWERF_BIN_PATH%\multiwerf.exe
setx /M PATH "%PATH%;%MULTIWERF_BIN_PATH%"

# откройте новую сессию и начните использовать multiwerf
```

###### Добавление werf alias в текущую shell-сессию

```shell
FOR /F "tokens=*" %g IN ('multiwerf use 1.1 stable --as-file --shell cmdexe') do (SET WERF_USE_SCRIPT_PATH=%g)
%WERF_USE_SCRIPT_PATH%
```

<!-- WERF DOCS PARTIAL END -->

# Первые шаги

<!-- WERF DOCS PARTIAL BEGIN: Getting started -->

Следующие руководства демонстрируют основные особенности и помогают быстро начать использовать werf:
- [Первые шаги](https://werf.io/documentation/guides/getting_started.html) — быстрый старт с использованием существующего Dockerfile.
- [Первое приложение](https://werf.io/documentation/guides/advanced_build/first_application.html) — сборка первого приложения (PHP Symfony), используя werf сборщик.
- [Деплой в Kubernetes](https://werf.io/documentation/guides/deploy_into_kubernetes.html) — выкат приложения в Kubernetes с интеграцией собранных образов. 
- [Интеграция с GitLab CI/CD](https://werf.io/documentation/guides/gitlab_ci_cd_integration.html) — конфигурация сборки, выката, удаления и очистки в GitLab CI.
- [Интеграция с неподдерживаемыми системами CI/CD](https://werf.io/documentation/guides/unsupported_ci_cd_integration.html) — интеграция werf в любую CI/CD-систему.
- [Приложение с несколькими образами](https://werf.io/documentation/guides/advanced_build/multi_images.html) — сборка приложения с несколькими образами (Java/ReactJS).
- [Использование монтирования](https://werf.io/documentation/guides/advanced_build/mounts.html) — уменьшение размера образа и ускорение сборки с помощью монтирования (Go/Revel).
- [Использование артефактов](https://werf.io/documentation/guides/advanced_build/artifacts.html) — уменьшение размера образа с помощью артефактов (Go/Revel).

<!-- WERF DOCS PARTIAL END -->

# Обещание обратной совместимости

<!-- WERF DOCS PARTIAL BEGIN: Backward Compatibility Promise -->

> _Note:_ Настоящее обещание относится к werf, начиная с версии 1.0, и не относится к предыдущим версиям или версиям dapp

werf использует [семантическое версионирование](https://semver.org/lang/ru/). Это значит, что мажорные версии (1.0, 2.0) могут быть обратно не совместимыми между собой. В случае werf это означает, что обновление на следующую мажорную версию _может_ потребовать полного передеплоя приложений, либо других ручных операций.

Минорные версии (1.1, 1.2, etc) могут добавлять новые "значительные" изменения, но без существенных проблем обратной совместимости в пределах мажорной версии. В случае werf это означает, что обновление на следующую минорную версию в большинстве случаев будет беспроблемным, но _может_ потребоваться запуск предоставленных скриптов миграции.

Патч-версии (1.1.0, 1.1.1, 1.1.2) могут добавлять новые возможности, но без каких-либо проблем обратной совместимости в пределах минорной версии (1.1.x).
В случае werf это означает, что обновление на следующий патч (следующую патч-версию) не должно вызывать проблем и требовать каких-либо ручных действий.

Все изменения проходят полный цикл по каналам стабильности:

- Канал обновлений `alpha` может содержать новые возможности и быть нестабильным. Релизы выполняются с высокой периодичностью.
  Мы **не гарантируем** обратную совместимость между версиями канала обновлений `alpha`.
- Канал обновлений `beta` предназначен для более детального тестирования новых возможностей.
  Мы **не гарантируем** обратную совместимость между версиями канала обновлений `beta`.
- Канал обновлений `ea` безопасно использовать в некритичных окружениях и при локальной разработке.
  Мы **не гарантируем** обратную совместимость между версиями канала обновлений `ea`.
- Канал обновлений `stable` считается безопасным и рекомендуемым для всех окружений.
  Мы **гарантируем**, что версия канала обновлений `ea` перейдет в канал обновлений `stable` не ранее чем через неделю после внутреннего тестирования.
  Мы **гарантируем** обратную совместимость между версиями канала обновлений `stable` в пределах минорной версии (1.1.x).
- Канал обновлений `rock-solid` рекомендуется использовать в критичных окружениях с высоким SLA.
  Мы **гарантируем**, что версия из канала обновлений `stable` перейдет в канал обновлений `rock-solid` не ранее чем через 2 недели плотного тестирования.
  Мы **гарантируем** обратную совместимость между версиями канала обновлений `rock-solid` в пределах минорной версии (1.1.x).

Соответствие каналов и релизов описывается в файле [multiwerf.json](https://github.com/werf/werf/blob/multiwerf/multiwerf.json), а использование актуальной версии werf в рамках канала должно быть организовано с помощью утилиты [multiwerf](https://github.com/flant/multiwerf).

Каналы стабильности и частые релизы позволяют получать непрерывную обратную связь по новым изменениям, выполнять быстрый откат проблемных изменений, а также обеспечивать высокую степень стабильности и при этом приемлемую скорость разработки. 

<!-- WERF DOCS PARTIAL END -->

# Документация и поддержка

<!-- WERF DOCS PARTIAL BEGIN: Docs and support -->

[Создайте ваше первое приложение с werf](https://ru.werf.io/documentation/guides/getting_started.html) или начните знакомство с чтения [документации](https://ru.werf.io/).

Мы всегда на связи с сообществом. Присоединяйтесь к нам в [Telegram](https://t.me/werf_ru), [Twitter](https://twitter.com/werf_io) или [Slack](https://cloud-native.slack.com/messages/CHY2THYUU)!

Мы следим за вашими [issues](https://github.com/werf/werf/issues) на GitHub.

<!-- WERF DOCS PARTIAL END -->

# Лицензия

<!-- WERF DOCS PARTIAL BEGIN: License -->

Apache License 2.0, see [LICENSE](LICENSE)
