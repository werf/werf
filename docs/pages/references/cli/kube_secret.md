---
title: dapp kube secret
sidebar: reference
permalink: kube_secret.html
folder: command
---

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

### dapp kube secret extract
Расшифровать данные ключом `DAPP_SECRET_KEY`.

```
dapp kube secret extract [FILE_PATH] [options]
```

#### --values
Расшифровать secret-values файл `FILE_PATH`.

#### -o OUTPUT_FILE_PATH
Перенаправляет расшифрованные данные в файл `OUTPUT_FILE_PATH`.

### dapp kube secret edit
Отредактировать секрет `FILE_PATH`.

```
dapp kube secret edit `FILE_PATH` [options]
```

#### --values
Отредактировать secret-values файл `FILE_PATH`.

### dapp kube secret regenerate
Перегенерировать секреты ключом `DAPP_SECRET_KEY`.

```
dapp kube secret regenerate [SECRET_VALUES_FILE_PATH ...] [options]
```

#### --old-secret-key KEY
Использовать ключ `KEY` для декодирования.
