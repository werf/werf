---
title: Команды очистки кэша и приложений
sidebar: doc_sidebar
permalink: cleanup_commands_for_advanced_build.html
folder: advanced_build
author: Alexey Igrychev <alexey.igrychev@flant.com>
---

## Сборочный кэш приложения

Набор docker-образов, который определяется стадиями приложения: одной стадии соответствует один docker-образ.

## Чистка неактуального кэша приложений

Для удаления неактуального кэша приложений проекта используются следующие команды.

```
# удаление кэша локально
dapp dimg stages cleanup local [options] [DIMG ...] [REPO]

# удаление кэша в registry
dapp dimg stages cleanup repo [options] [DIMG ...] [REPO]
```

При удаление необходимо выбрать одну или несколько опций.

| опция | описание |
| ----- | -------- |
| `--improper-repo-cache` | удалить неактуальный кэш, не являющийся кэшом существующих в registry приложений |
| `--improper-git-commit` | удалить кэш, связанный с отсутствующими в git-репозиториях коммитами (после git rebase) |
| `--improper-cache-version` | удалить устаревший кэш, версия (зашита в dapp) которого не соответствует текущей |

## Удаление кэша приложений

К полному удалению кэша приводит выполнение следующих команд.
```
# удаление локально
dapp dimg stages flush local [options] [DIMG ...]

# удаление в registry
dapp dimg stages flush repo [options] [DIMG ...] REPO
```

## Чистка docker после некорректного завершения сборки

Команда вызывается для удаления нетегированных docker-образов и запущенных docker-контейнеров.

```
dapp dimg cleanup [options] [DIMG ...]
```

## Чистка docker

Команда позволяет очистить docker в соответствии с переданной опцией.

```
dapp dimg mrproper [options]
```

| опция | описание |
| ----- | -------- |
| `--all` | удалить docker-образы и docker-контейнеры связанные с dapp |
| `--improper-cache-version-stages` | удалить устаревший кэш, версия (зашита в dapp) которого не соответствует текущей |
| `--improper-dev-mode-cache` | удалить кэш разработчика |
