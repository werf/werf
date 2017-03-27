---
title: Команды отладки
sidebar: doc_sidebar
permalink: debugging_commands.html
folder: command
---

### dapp dimg list
Вывести список dimg-ей.

```
dapp dimg list [options] [DIMG ...]
```

### dapp dimg run
Запустить собранный dimg с докерными аргументами **DOCKER ARGS**.

```
dapp dimg run [options] [DIMG] [DOCKER ARGS]
```

#### [DOCKER ARGS]
Может содержать докерные опции и/или команду.

Перед командой необходимо использовать группу символов ' -- '.

#### Примеры

##### Запустить собранный dimg с опциями
```bash
$ dapp dimg run -ti --rm
```

##### Запустить с опциями и командами
```bash
$ dapp dimg run -ti --rm -- bash -ec true
```

##### Запустить, передав только команды
```bash
$ dapp dimg run -- bash -ec true
```

##### Посмотреть, что может быть запущено
```bash
$ dapp dimg run app -ti --rm -- bash -ec true --dry-run
docker run -ti --rm app-dimgstage:ea5ec7543c809ec7e9fe28181edfcb2ee6f48efaa680f67bf23a0fc0057ea54c bash -ec true
```
### dapp dimg build-context export
Экспортировать контекст, [кэш приложений](#кэш-приложения) и [директорию сборки](#директория-сборки-dapp). 

```
dapp dimg build-context export [options] [DIMG ...]
```

#### `--build-context-directory DIR_PATH`
Определить директорию сохранения контекста.

#### Пример

##### Экспортировать контекст приложения 'frontend' из проекта 'project' в директорию 'context' 

```bash
$ dapp dimg build-context export --dir ~/workspace/project --build-context-directory context frontend
```

### dapp dimg build-context import
Импортировать контекст, [кэш приложений](#кэш-приложения) и [директорию сборки](#директория-сборки-dapp).

```
dapp dimg build-context import [options]
```

#### `--build-context-directory DIR_PATH`
Определить директорию хранения контекста.

#### Пример

##### Импортировать контекст приложений проекта 'project' из директории 'context' 

```bash
$ dapp dimg build-context import --dir ~/workspace/project --build-context-directory context
```
