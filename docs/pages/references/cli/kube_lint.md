---
title: Тестирование helm чарта (линтер)
sidebar: doc_sidebar
permalink: kube_lint.html
folder: command
---

### dapp kube lint
Вызывает [`helm lint`](https://docs.helm.sh/helm/#helm-lint) для helm чарта, - тестирует чарт на возможные ошибки.

```
dapp kube lint [options] [REPO]
```

#### --namespace NAMESPACE
Задаёт `namespace`, по умолчанию используется context namespace.

#### Опции тегирования `--tag TAG`, `--tag-branch`, `--tag-commit`, `--tag-build-id`, `--tag-ci`, `--tag-slug TAG`, `--tag-plain TAG`
Влияют на то, что будет в результате использования шаблона `dapp_container_image`. Значение опций соответствует указываемым в [dapp dimg push](base_commands.html#dapp-dimg-push).

#### --set STRING_ARRAY
Пробрасывает значения в `helm lint` (значения могут передаваться в одной опции, в формате `key1=val1,key2=val2`, или разбиваться на несколько).

#### --values FILE_PATH
Определяет YAML фаил, значения из которого пробрасываются в `helm lint` (опция может использоваться несколько раз).

#### --secret-values FILE_PATH
Декодирует значения описанные в YAML файле, используя [ключ шифрования](#dapp-kube-secret-key-generate), затем пробрасывает их в `helm lint` (опция может использоваться несколько раз).
