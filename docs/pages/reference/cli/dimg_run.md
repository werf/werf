---
title: dapp dimg run
sidebar: reference
permalink: reference/cli/dimg_run.html
---

### dapp dimg run
Запустить собранную стадию dimg-а с докерными аргументами **DOCKER ARGS**.

```
dapp dimg run [options] [DIMG] [DOCKER ARGS]
```

#### `--stage STAGE`
Определить стадию сборки. По умолчанию используется последняя стадия.

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
