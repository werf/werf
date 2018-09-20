---
title: dapp kube lint
sidebar: reference
permalink: reference/cli/kube_lint.html
---

### dapp kube lint
Вызывает [`helm lint`](https://docs.helm.sh/helm/#helm-lint) для helm чарта, - тестирует чарт на возможные ошибки.

```
dapp kube lint [options] REPO
```

#### --namespace NAMESPACE
Задаёт `namespace`, по умолчанию используется context namespace.

#### Опции тегирования `--tag TAG`, `--tag-branch`, `--tag-commit`, `--tag-build-id`, `--tag-ci`, `--tag-slug TAG`, `--tag-plain TAG`
Влияют на то, что будет в результате использования шаблона `dapp_container_image`. Значение опций соответствует указываемым в [dapp dimg push]({{ site.baseurl }}/reference/cli/dimg_push.html).

#### --set STRING_ARRAY
Пробрасывает значения в `helm lint` (значения могут передаваться в одной опции, в формате `key1=val1,key2=val2`, или разбиваться на несколько).

Соответственно, правила проброса параметров [наследуются из helm](https://github.com/helm/helm/blob/master/docs/using_helm.md#the-format-and-limitations-of---set)

#### --values FILE_PATH
Определяет YAML фаил, значения из которого пробрасываются в `helm lint` (опция может использоваться несколько раз).

#### --secret-values FILE_PATH
Декодирует значения описанные в YAML файле, используя [ключ шифрования](kube_secret.html), затем пробрасывает их в `helm lint` (опция может использоваться несколько раз).
