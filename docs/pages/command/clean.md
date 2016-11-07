---
title: Комманды очистки
sidebar: doc_sidebar
permalink: clean_commands.html
folder: command
---

### dapp stages cleanup local
Удалить неактуальный локальный [кэш приложений](definitions.html#кэш-приложения) [проекта](definitions.html#проект), опираясь на приложения в репозитории **REPO**.

```
dapp stages cleanup local [options] [APPS PATTERN ...] REPO
```

#### --improper-cache-version
Удалить устаревший кэш приложений проекта.

#### Примеры

##### Удалить неактуальный кэш приложений
```bash
$ dapp stages cleanup local localhost:5000/test --improper-cache-version
```

### dapp stages cleanup repo
Удалить неиспользуемый [кэш приложений](definitions.html#кэш-приложения) в репозитории **REPO**.

```
dapp stages cleanup repo [options] [APPS PATTERN ...] REPO
```

#### --improper-cache-version
Удалить устаревший кэш приложений проекта.

#### Примеры

##### Удалить неактуальный кэш приложений в репозитории localhost:5000/test
```bash
$ dapp stages cleanup repo localhost:5000/test
```

### dapp stages flush local
Удалить [кэш приложений](definitions.html#кэш-приложения) [проекта](definitions.html#проект).

```
dapp stages flush local [options] [APPS PATTERN ...]
```

#### Примеры

##### Удалить кэш приложений
```bash
$ dapp stages flush local
```

### dapp stages flush repo
Удалить приложения и [кэш приложений](definitions.html#кэш-приложения) [проекта](definitions.html#проект) в репозитории **REPO**.

```
dapp stages flush repo [options] [APPS PATTERN ...] REPO
```

#### Примеры

##### Удалить весь кэш приложений в репозитории localhost:5000/test
```bash
$ dapp stages flush repo localhost:5000/test
```

### dapp cleanup
Убраться в системе после некорректного завершения работы dapp, удалить нетеггированные docker-образы и docker-контейнеры [проекта](definitions.html#проект).

```
dapp cleanup [options] [APPS PATTERN ...]
```

#### Примеры

##### Запустить
```bash
$ dapp cleanup
```

##### Посмотреть, какие команды могут быть выполнены
```bash
$ dapp cleanup --dry-run
backend
  docker rm -f dd4ec7v33
  docker rmi ea5ec7543 c809ec7e9f ee6f48efa6
```

### dapp mrproper
Очистить docker в соответствии с переданным параметром.

```
dapp mrproper [options]
```

#### --all
Удалить docker-образы и docker-контейнеры связанные с dapp.

#### --improper-cache-version-stages
Удалить устаревший кэш приложений.

#### Примеры

##### Запустить очистку
```bash
$ dapp mrproper --all
```

##### Посмотреть, версия кэша каких образов устарела, какие команды могут быть выполнены:
```bash
$ dapp mrproper --improper-cache-version-stages --dry-run
mrproper
  proper cache
    docker rmi dimgstage-dapp-test-project-services-stats:ba95ec8a00638ddac413a13e303715dd2c93b80295c832af440c04a46f3e8555 dimgstage-dapp-test-project-services-stats:f53af70566ec23fb634800d159425da6e7e61937afa95e4ed8bf531f3503daa6
```
