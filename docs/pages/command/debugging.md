---
title: Команды отладки
sidebar: doc_sidebar
permalink: debugging_commands.html
folder: command
---

### dapp list
Вывести список dimg-ей.

```
dapp list [options] [DIMG ...]
```

### dapp run
Запустить собранный dimg с докерными аргументами **DOCKER ARGS**.

```
dapp run [options] [DIMG] [DOCKER ARGS]
```

#### [DOCKER ARGS]
Может содержать докерные опции и/или команду.

Перед командой необходимо использовать группу символов ' -- '.

#### Примеры

##### Запустить собранный dimg с опциями
```bash
$ dapp run -ti --rm
```

##### Запустить с опциями и командами
```bash
$ dapp run -ti --rm -- bash -ec true
```

##### Запустить, передав только команды
```bash
$ dapp run -- bash -ec true
```

##### Посмотреть, что может быть запущено
```bash
$ dapp run app -ti --rm -- bash -ec true --dry-run
docker run -ti --rm app-dimgstage:ea5ec7543c809ec7e9fe28181edfcb2ee6f48efaa680f67bf23a0fc0057ea54c bash -ec true
```
