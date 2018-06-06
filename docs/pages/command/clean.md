---
title: Команды очистки
sidebar: doc_sidebar
permalink: clean_commands.html
folder: command
---

### dapp dimg cleanup repo
Удалить теги [приложений](definitions.html#dimg) [проекта](definitions.html#проект), исходя из соответствующих схем тегирования.

<table class="tag-scheme">
  <tr>
    <td>Опция тегирования</td>
    <td>--tag<br />--tag-slug<br />--tag-plain</td>
    <td>--tag-branch</td>
    <td>--tag-commit</td>
    <td>--tag-build-id</td>
    <td>--tag-ci</td>
  </tr>
  <tr>
    <td>Схема тегирования</td>
    <td>custom</td>
    <td>git_branch</td>
    <td>git_commit</td>
    <td>ci</td>
    <td>git_tag или git_branch</td>
  </tr>
</table>

* Имена которых содержат неактуальные данные:
    * ветка или тег удалены из репозитория (`git_branch`, `git_tag`);
    * комит отсутсвует в репозитории, был сделан rebase (`git_commit`).
* Комиты которых были созданы более одного месяца назад (`git_tag`, `git_commit`);
* Лишние, в случае, если привышен лимит в 10 тегов на [приложение](definitions.html#dimg), исходя из времени создания комитов (`git_tag`, `git_commit`).

```
dapp dimg cleanup repo [options] [DIMG ...] REPO
```

#### `--with-stages`
Соответствует вызову команды `dapp dimg stages cleanup local` с опцией `--improper-repo-cache`.

### dapp dimg flush local
Удалить все теги [приложений](definitions.html#dimg) [проекта](definitions.html#проект) локально.

```
dapp dimg flush local [options] [DIMG ...]
```

#### `--with-stages`
Соответствует вызову команды `dapp dimg stages flush local`.

### dapp dimg flush repo
Удалить все теги [приложений](definitions.html#dimg) [проекта](definitions.html#проект).
```
dapp dimg flush repo [options] [DIMG ...] REPO
```

#### `--with-stages`
Соответствует вызову команды `dapp dimg stages cleanup flush`.

### dapp dimg stages cleanup local
Удалить неактуальный локальный [кэш приложений](definitions.html#кэш-приложения) [проекта](definitions.html#проект).

```
dapp dimg stages cleanup local [options] [REPO]
```

#### `--improper-repo-cache`
Удалить кэш, который не использовался при сборке приложений в репозитории **REPO**.

#### `--improper-git-commit`
Удалить кэш, связанный с отсутствующими в git-репозиториях коммитами.

#### `--improper-cache-version`
Удалить устаревший кэш приложений проекта.

#### Примеры

##### Оставить только актуальный кэш, исходя из приложений в localhost:5000/test
```bash
$ dapp dimg stages cleanup local --improper-repo-cache localhost:5000/test
```

##### Удалить кэш, версия которого не совпадает с текущей
```bash
$ dapp dimg stages cleanup local --improper-cache-version localhost:5000/test
```

##### Почистить кэш после rebase в одном из связанных git-репозиториев
```bash
$ dapp dimg stages cleanup local --improper-git-commit localhost:5000/test
```

### dapp dimg stages cleanup repo
Удалить неиспользуемый [кэш приложений](definitions.html#кэш-приложения) в репозитории **REPO**.

```
dapp dimg stages cleanup repo [options] REPO
```

#### `--improper-repo-cache`
Удалить кэш, который не использовался при сборке приложений в репозитории **REPO**.

#### `--improper-git-commit`
Удалить кэш, связанный с отсутствующими в git-репозиториях коммитами.

#### `--improper-cache-version`
Удалить устаревший кэш приложений проекта.

#### Примеры

##### Удалить неактуальный кэш в репозитории localhost:5000/test, где:

* Версия кэша не соответствует текущей.
* Не связан с приложениями в репозитории.
* Собран из коммитов, которые в данный момент отсутствуют в git-репозиториях проекта.

```bash
$ dapp dimg stages cleanup repo localhost:5000/test --improper-cache-version --improper-repo-cache --improper-git-commit
```

### dapp dimg stages flush local
Удалить [кэш приложений](definitions.html#кэш-приложения) [проекта](definitions.html#проект).

```
dapp dimg stages flush local [options]
```

#### Примеры

##### Удалить кэш приложений
```bash
$ dapp dimg stages flush local
```

### dapp dimg stages flush repo
Удалить [кэш приложений](definitions.html#кэш-приложения) [проекта](definitions.html#проект) в репозитории **REPO**.

```
dapp dimg stages flush repo [options] REPO
```

#### Примеры

##### Удалить весь кэш приложений в репозитории localhost:5000/test
```bash
$ dapp dimg stages flush repo localhost:5000/test
```

### dapp dimg cleanup
Убраться в системе после некорректного завершения работы dapp, удалить нетегированные docker-образы и docker-контейнеры.

```
dapp dimg cleanup [options]
```

#### Примеры

##### Запустить
```bash
$ dapp dimg cleanup
```

##### Посмотреть, какие команды могут быть выполнены
```bash
$ dapp dimg cleanup --dry-run
backend
  docker rm -f dd4ec7v33
  docker rmi dimgstage-dapp-test-project:07758b3ec8aec701a01 dimgstage-dapp-test-project:ec701a0107758b3ec8a
```

### dapp dimg mrproper
Очистить docker в соответствии с переданным параметром. Команда не ограничивается каким-либо проектом, а влияет на образы **всех проектов**.

```
dapp dimg mrproper [options]
```

#### `--all`
Удалить docker-образы и docker-контейнеры **всех проектов**, связанные с dapp.

#### `--improper-dev-mode-cache`
Удалить docker-образы и docker-контейнеры всех проектов, собранные в dev-режиме и связанные с dapp.

#### `--improper-cache-version-stages`
Удалить устаревший кэш приложений, т.е. кэш, версия которого не соответствует версии кэша запущенного dapp, во всех проектах. 

#### Примеры

##### Запустить очистку
```bash
$ dapp dimg mrproper --all
```

##### Посмотреть, версия кэша каких образов устарела, какие команды могут быть выполнены:
```bash
$ dapp dimg mrproper --improper-cache-version-stages --dry-run
mrproper
  proper cache
    docker rmi dimgstage-dapp-test-project-services-stats:ba95ec8a00638ddac413a13e303715dd2c93b80295c832af440c04a46f3e8555 dimgstage-dapp-test-project-services-stats:f53af70566ec23fb634800d159425da6e7e61937afa95e4ed8bf531f3503daa6
```
