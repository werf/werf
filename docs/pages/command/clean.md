---
title: Команды очистки
sidebar: doc_sidebar
permalink: clean_commands.html
folder: command
---

### dapp dimg stages cleanup local
Удалить неактуальный локальный [кэш приложений](definitions.html#кэш-приложения) [проекта](definitions.html#проект).

```
dapp dimg stages cleanup local [options] [DIMG ...] [REPO]
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
$ dapp dimg stages cleanup local localhost:5000/test
```

##### Удалить кэш, версия которого не совпадает с текущей
```bash
$ dapp dimg stages cleanup local --improper-cache-version
```

##### Почистить кэш после rebase в одном из связанных git-репозиториев
```bash
$ dapp dimg stages cleanup local --improper-git-commit
```

### dapp dimg stages cleanup repo
Удалить неиспользуемый [кэш приложений](definitions.html#кэш-приложения) в репозитории **REPO**.

```
dapp dimg stages cleanup repo [options] [DIMG ...] REPO
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
dapp dimg stages flush local [options] [DIMG ...]
```

#### Примеры

##### Удалить кэш приложений
```bash
$ dapp dimg stages flush local
```

### dapp dimg stages flush repo
Удалить приложения и [кэш приложений](definitions.html#кэш-приложения) [проекта](definitions.html#проект) в репозитории **REPO**.

```
dapp dimg stages flush repo [options] [DIMG ...] REPO
```

#### Примеры

##### Удалить весь кэш приложений в репозитории localhost:5000/test
```bash
$ dapp dimg stages flush repo localhost:5000/test
```

### dapp dimg cleanup
Убраться в системе после некорректного завершения работы dapp, удалить нетегированные docker-образы и docker-контейнеры [проекта](definitions.html#проект).

```
dapp dimg cleanup [options] [DIMG ...]
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
  docker rmi ea5ec7543 c809ec7e9f ee6f48efa6
```

### dapp dimg mrproper
Очистить docker в соответствии с переданным параметром.

```
dapp dimg mrproper [options]
```

#### `--all`
Удалить docker-образы и docker-контейнеры связанные с dapp.

#### `--improper-cache-version-stages`
Удалить устаревший кэш приложений.

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
