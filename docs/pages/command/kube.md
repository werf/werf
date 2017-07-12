---
title: Команды развёртывания приложений
sidebar: doc_sidebar
permalink: kube_commands.html
folder: command
---

### dapp kube deploy
Развернуть приложение, chart описанный в `.helm` согласно [документации helm](https://github.com/kubernetes/helm/blob/master/docs/index.md).

```
dapp kube deploy [options] REPO
```

#### --namespace NAMESPACE
Задаёт `namespace`, по умолчанию используется context namespace.

#### --image_version VERSION
Задаёт версию для используемых `dimg`.

#### --tmp_dir_prefix PREFIX
Переопределяет префикс временной директории, временного chart-a, который удаляется после запуска.

#### --set STRING_ARRAY
Пробрасывает значения в `helm upgrade` (значения могут передаваться в одной опции, в формате `key1=val1,key2=val2`, или разбиваться на несколько). 

#### --values FILE_PATH
Пробрасывает значения описанные в YAML файле в `helm upgrade` (опция может использоваться несколько раз).

#### --secret-values FILE_PATH
Декодирует значения описанные в YAML файле, используя [ключ шифрования](#dapp-kube-secret-key-generate), затем пробрасывает их в `helm upgrade` (опция может использоваться несколько раз).

### dapp kube dismiss
Удалить релиз.

```
dapp kube dismiss [options]
```

#### --namespace NAMESPACE
Задаёт `namespace`, по умолчанию используется context namespace.

#### --with_namespace
При удалении релиза также удаляет `namespace`.

### dapp kube secret key generate
Сгенерировать ключ шифрования.

```
dapp kube secret key generate
```

### dapp kube secret generate
Зашифровать данные ключом `DAPP_SECRET_KEY`.

```
dapp kube secret generate [FILE_PATH] [options]
```

#### --values
Зашифровать secret-values файл `FILE_PATH`.

#### -o OUTPUT_FILE_PATH
Перенаправляет зашифрованные данные в файл `OUTPUT_FILE_PATH`.

### dapp kube secret generate
Расшифровать данные ключом `DAPP_SECRET_KEY`.

```
dapp kube secret extract [FILE_PATH] [options]
```

#### --values
Расшифровать secret-values файл `FILE_PATH`.

#### -o OUTPUT_FILE_PATH
Перенаправляет расшифрованные данные в файл `OUTPUT_FILE_PATH`.

### dapp kube secret regenerate
Перегенерировать секреты ключом `DAPP_SECRET_KEY`.

```
dapp kube secret regenerate [SECRET_VALUES_FILE_PATH ...] [options]
```

#### --old-secret-key KEY
Использовать ключ `KEY` для декодирования.