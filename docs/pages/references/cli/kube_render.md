---
title: dapp kube render
sidebar: reference
permalink: kube_render.html
folder: command
---

### dapp kube render
Вызывает [`helm template`](https://docs.helm.sh/helm/#helm-template) для helm чарта, - выполняет локальный рендеринг чарта и выводит результат.

```
dapp kube render [options] [REPO]
```

#### --namespace NAMESPACE
Задаёт `namespace`, по умолчанию используется context namespace.

#### Опции тегирования `--tag TAG`, `--tag-branch`, `--tag-commit`, `--tag-build-id`, `--tag-ci`, `--tag-slug TAG`, `--tag-plain TAG`
Влияют на то, что будет в результате использования шаблона `dapp_container_image`. Значение опций соответствует указываемым в [dapp dimg push](dimg_push.html).

#### --tmp-dir-prefix PREFIX
Переопределяет префикс временной директории для размещения временных папок используемых при рендеринге (по умолчанию - /tmp).

#### --set STRING_ARRAY
Пробрасывает значения в `helm render` (значения могут передаваться в одной опции, в формате `key1=val1,key2=val2`, или разбиваться на несколько).

Соответственно, правила проброса параметров [наследуются из helm](https://github.com/helm/helm/blob/master/docs/using_helm.md#the-format-and-limitations-of---set)

#### --values FILE_PATH
Определяет YAML фаил, значения из которого пробрасываются в `helm render` (опция может использоваться несколько раз).

#### --secret-values FILE_PATH
Декодирует значения описанные в YAML файле, используя [ключ шифрования](kube_secret.html), затем пробрасывает их в `helm render` (опция может использоваться несколько раз).

#### -t, --templates GLOB_PATTERN
Задает шаблон имени файлов чарта, которые необходимо рендерить. В качестве параметра указывается шаблон:
* пути до chart `-t .helm/templates/*_custom`;
* пути до subchart `-t .helm/charts/custom_chart/templates/*_custom`;
* имени (`-t *_custom`).

Пути могут быть как относительными, так и абсолютными. Наличие `*` учитывает только один уровень иерархии и не выполняет рекурсивный обход. Ко всем шаблонам в конец добавляется `*`. Например, `dapp kube render -t '*/*-app'` выполнит рендеринг yaml файлов чарта - `./.helm/templates/10-app.yaml` и `./.helm/templates/50-app-api.yaml`.
