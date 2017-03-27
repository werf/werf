---
title: Registry (#TODO)
sidebar: doc_sidebar
permalink: registry_for_advanced_build.html
folder: advanced_build
author: Alexey Igrychev <alexey.igrychev@flant.com>
---

* Схема именования (совместимая с gitlab)
* Автоочистка
* Распределенная сборка и распределенный кеш

### Тезисы / проблемы / вопросы

* Какие схемы тегирования поддерживаются?
* Как синхронизировать данные с registry?
* Как обеспечить оптимальную работу, имея несколько серверов сборки проекта?

### Предметная область

```shell
dapp dimg push
dapp dimg spush
dapp dimg bp
dapp dimg stages push
dapp dimg stages pull
dapp dimg stages cleanup local/repo
dapp dimg stages flush local/repo
```

### Раскрытие темы

#### Поддерживаемые схемы тегирования

| опция | описание |
| ----- | -------- |
| `--tag TAG` | произвольный тег TAG |
| `--tag-branch` | тег с именем ветки сборки |
| `--tag-commit` | тег с коммитом сборки |
| `--tag-build-id` | тег с идентификатором сборки |
| `--tag-ci` | тег, собранный из относящихся к сборке переменных окружения CI системы |

#### Синхронизация данных с registry

В случае, если в registry данные неактуальные, а локально актуальны: очистить registry и экспортировать локальные кэш и приложения проекта.

```shell
dapp dimg stages flush repo REPO
dapp dimg push --with-stages REPO	
```

Если же в registry данные актуальны, а локально неактуальные: очистить локальный кэш, не связанный ни с одним приложением в registry, и импортировать существующий кэш стадий для текущего состояния проекта из registry.

```shell
dapp dimg stages cleanup local --improper-repo-cache REPO
dapp dimg stages pull REPO
```

#### Распределённая  работа нескольких серверов сборки

```shell
dapp dimg stages pull REPO (1)
dapp dimg build (2)
dapp dimg push --with-stages REPO (3)
dapp dimg stages cleanup local --improper-repo-cache REPO (4)
```

1. импортируется существующий кэш стадий для сборки текущего проекта;
2. собирается приложения проекта;
3. экспортируются приложения и кэш стадий;
4. удаляется кэш, который не связан ни с одним приложением проекта в registry.
