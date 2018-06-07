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

### dapp kube lint
Вызывает [`helm lint`](https://docs.helm.sh/helm/#helm-lint) для helm чарта, - тестирует чарт на возможные ошибки.

```
dapp kube lint [options] [REPO]
```

#### --namespace NAMESPACE
Задаёт `namespace`, по умолчанию используется context namespace.

#### --image-version IMAGE_VERSION
Задаёт версию для используемых `dimg`, по умолчанию используется `latest`.

#### --set STRING_ARRAY
Пробрасывает значения в `helm lint` (значения могут передаваться в одной опции, в формате `key1=val1,key2=val2`, или разбиваться на несколько).

#### --values FILE_PATH
Определяет YAML фаил, значения из которого пробрасываются в `helm lint` (опция может использоваться несколько раз).

#### --secret-values FILE_PATH
Декодирует значения описанные в YAML файле, используя [ключ шифрования](#dapp-kube-secret-key-generate), затем пробрасывает их в `helm lint` (опция может использоваться несколько раз).

### dapp kube render
Вызывает [`helm template`](https://docs.helm.sh/helm/#helm-template) для helm чарта, - выполняет локальный рендеринг чарта и выводит результат.

```
dapp kube render [options] [REPO]
```

#### --namespace NAMESPACE
Задаёт `namespace`, по умолчанию используется context namespace.

#### --tmp-dir-prefix PREFIX
Переопределяет префикс временной директории для размещения временных папок используемых при рендеринге (по умолчанию - /tmp).

#### --set STRING_ARRAY
Пробрасывает значения в `helm render` (значения могут передаваться в одной опции, в формате `key1=val1,key2=val2`, или разбиваться на несколько).

#### --values FILE_PATH
Определяет YAML фаил, значения из которого пробрасываются в `helm render` (опция может использоваться несколько раз).

#### --secret-values FILE_PATH
Декодирует значения описанные в YAML файле, используя [ключ шифрования](#dapp-kube-secret-key-generate), затем пробрасывает их в `helm render` (опция может использоваться несколько раз).

#### -t, --templates GLOB_PATTERN
Задает шаблон имени файлов чарта, которые необходимо рендерить. В качестве параметра указывается шаблон:
* пути до chart `-t .helm/templates/*_custom`;
* пути до subchart `-t .helm/charts/custom_chart/templates/*_custom`;
* имени (`-t *_custom`).

Пути могут быть как относительными, так и абсолютными. Наличие `*` учитывает только один уровень иерархии и не выполняет рекурсивный обход. Ко всем шаблонам в конец добавляется `*`. Например, `dapp kube render -t '*/*-app'` выполнит рендеринг yaml файлов чарта - `./.helm/templates/10-app.yaml` и `./.helm/templates/50-app-api.yaml`.

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

### dapp kube create chart
Сгенерировать helm-chart с частоиспользуемыми файлами и директориями (`.helm`).

```
dapp kube chart create [options]
```

#### --force
Удалить существующий chart (по умолчанию происходит наложение, при этом существующие файлы не перетираются).
