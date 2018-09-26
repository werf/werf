---
title: dapp kube deploy
sidebar: reference
permalink: reference/cli/dapp_kube_deploy.html
---

Развернуть приложение, chart описанный в `.helm` согласно [документации helm](https://github.com/kubernetes/helm/blob/master/docs/index.md).

```
dapp kube deploy [options] REPO
```

## Options
#### --set STRING_ARRAY
Пробрасывает значения в `helm upgrade` (значения могут передаваться в одной опции, в формате `key1=val1,key2=val2`, или разбиваться на несколько).

Соответственно, правила проброса параметров [наследуются из helm](https://github.com/helm/helm/blob/master/docs/using_helm.md#the-format-and-limitations-of---set)

#### --values FILE_PATH
Пробрасывает значения описанные в YAML файле в `helm upgrade` (опция может использоваться несколько раз).

#### --secret-values FILE_PATH
Декодирует значения описанные в YAML файле, используя [ключ шифрования](kube_secret.html), затем пробрасывает их в `helm upgrade` (опция может использоваться несколько раз).

#### --namespace NAMESPACE
Задаёт `namespace`, по умолчанию используется context namespace.

### --tmp_dir_prefix PREFIX
Переопределяет префикс временной директории, временного chart-a, который удаляется после запуска.

{% include_relative partial/option/ssh_option.md %}

### Tag options
{% include_relative partial/option/tag_options.md %}
