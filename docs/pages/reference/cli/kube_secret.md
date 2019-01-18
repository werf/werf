---
title: werf kube secret
sidebar: reference
permalink: reference/cli/kube_secret.html
---

### werf kube secret key generate
Сгенерировать ключ шифрования.

```
werf kube secret key generate
```

### werf kube secret generate
Зашифровать данные ключом `WERF_SECRET_KEY`.

```
werf kube secret generate [FILE_PATH] [options]
```

#### --values
Зашифровать secret-values файл `FILE_PATH`.

#### -o OUTPUT_FILE_PATH
Перенаправляет зашифрованные данные в файл `OUTPUT_FILE_PATH`.

### werf kube secret extract
Расшифровать данные ключом `WERF_SECRET_KEY`.

```
werf kube secret extract [FILE_PATH] [options]
```

#### --values
Расшифровать secret-values файл `FILE_PATH`.

#### -o OUTPUT_FILE_PATH
Перенаправляет расшифрованные данные в файл `OUTPUT_FILE_PATH`.

### werf kube secret edit
Отредактировать секрет `FILE_PATH`.

```
werf kube secret edit `FILE_PATH` [options]
```

#### --values
Отредактировать secret-values файл `FILE_PATH`.

### werf kube secret regenerate
Перегенерировать секреты ключом `WERF_SECRET_KEY`.

```
werf kube secret regenerate [SECRET_VALUES_FILE_PATH ...] [options]
```

#### --old-secret-key KEY
Использовать ключ `KEY` для декодирования.
