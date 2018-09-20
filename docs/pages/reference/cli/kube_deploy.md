---
title: dapp kube deploy
sidebar: reference
permalink: reference/cli/kube_deploy.html
---

### dapp kube deploy
Развернуть приложение, chart описанный в `.helm` согласно [документации helm](https://docs.helm.sh).

```
dapp kube deploy [options] REPO
```

#### --namespace NAMESPACE
Задаёт `namespace`, по умолчанию используется context namespace.

#### Опции тегирования `--tag TAG`, `--tag-branch`, `--tag-commit`, `--tag-build-id`, `--tag-ci`, `--tag-slug TAG`, `--tag-plain TAG`
Версия образа из указанного репозитория. Значение опций соответствует указываемым в [dapp dimg push]({{ site.baseurl }}/reference/cli/dimg_push.html).

#### --tmp_dir_prefix PREFIX
Переопределяет префикс временной директории, временного chart-a, который удаляется после запуска.

#### --set STRING_ARRAY
Пробрасывает значения в `helm upgrade` (значения могут передаваться в одной опции, в формате `key1=val1,key2=val2`, или разбиваться на несколько).

Соответственно, правила проброса параметров [наследуются из helm](https://github.com/helm/helm/blob/master/docs/using_helm.md#the-format-and-limitations-of---set)

#### --values FILE_PATH
Пробрасывает значения описанные в YAML файле в `helm upgrade` (опция может использоваться несколько раз).

#### --secret-values FILE_PATH
Декодирует значения описанные в YAML файле, используя [ключ шифрования](kube_secret.html), затем пробрасывает их в `helm upgrade` (опция может использоваться несколько раз).
