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
Задаёт `namespace`, по умолчанию `default`.

#### --image_version VERSION
Задаёт версию для используемых `dimg`.

#### --tmp_dir_prefix PREFIX
Переопределяет префикс временной директории, временного chart-a, который удаляется после запуска.

### dapp kube dismiss
Удалить релиз.

```
dapp kube dismiss [options]
```

#### --namespace NAMESPACE
Задаёт `namespace`, по умолчанию `default`.

#### --with_namespace
При удалении релиза также удаляет `namespace`.

### dapp kube secret key generate
Сгенерировать ключ шифрования.

```
dapp kube secret key generate
```

### dapp kube secret generate
Зашифровать введённое значение ключом `DAPP_SECRET_KEY`. Для завершения ввода необходимо нажать `Ctrl + D`.

```
dapp kube secret generate
```

### dapp kube secret file encrypt
Зашифровать данные файла `FILE_PATH`. 

```
dapp kube secret file encrypt FILE_PATH [options]
```

#### -o OUTPUT_FILE_PATH
Перенаправляет зашифрованные данные в файл `OUTPUT_FILE_PATH`.