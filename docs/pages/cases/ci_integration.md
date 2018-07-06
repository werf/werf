---
title: Сборка в CI
sidebar: doc_sidebar
permalink: ci_integration.html
---



## Схемы тегирования с использованием CI переменных среды

На данный момент поддерживаются следующие CI системы:

* GitLab CI;
* Travis CI.

### Схема тегирования `--tag-ci`

Тег формируется из переменных среды:

* `CI_BUILD_ID`;
* `TRAVIS_BUILD_NUMBER`.

### Схема тегирования `--tag-build-id`

Тег формируется из переменных среды:

* `CI_BUILD_REF_NAME`, `CI_BUILD_TAG`;
* `TRAVIS_BRANCH`, `TRAVIS_TAG`.

## Распределённая работа нескольких серверов сборки

Кэш стал находится в registry и появилась возможность масштабировать сборку.

Для оптимальной работы каждого сервера сборки предлагается следующий набор команд при сборке.

```shell
dapp dimg stages pull REPO
dapp dimg build
dapp dimg push --with-stages REPO
dapp dimg stages cleanup local --improper-repo-cache REPO
```

* импортируется существующий кэш стадий для сборки текущего проекта;
* собирается приложения проекта;
* экспортируются приложения и кэш стадий;
* удаляется кэш, который не связан ни с одним приложением проекта в registry.

## Пример .gitlab-ci.yml (#TODO)
