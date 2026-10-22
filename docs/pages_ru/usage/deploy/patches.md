---
title: Патчи ресурсов
permalink: usage/deploy/patches.html
---

## Обзор

В werf есть два вида патчей, у каждого своя задача:
* Render-патчи меняют то, что развертывается, — похоже на `--post-renderer` в Helm. Они нужны, чтобы задать поля, которые чарт не выносит в values, без форка чарта.
* Diff-патчи меняют только то, что werf сравнивает с кластером, решая, нужно ли обновлять ресурс. Они нужны, чтобы ресурс не обновлялся из-за полей, которые меняются при каждом рендере или правятся в кластере.

Чтобы добавить патчи, создайте `patches.yaml` в корне чарта или передайте файл через `--patches`; сам патч — программа на [jq](https://jqlang.org/manual/).

## Изменение ресурсов без правки чарта (только в werf)

Чтобы изменить отрендеренный ресурс без правки его шаблона — например, ресурс зависимого чарта, чьи values не выносят нужное поле, — добавьте render-патч в `patches.yaml` в корне чарта.

Пример:

```yaml
# .helm/patches.yaml:
renderPatches:
- match:
    kinds: [Deployment, StatefulSet]
    charts: [redis]
  patch: .spec.template.spec.priorityClassName = $Values.redisPriorityClass
```

```shell
werf render --set redisPriorityClass=high-priority
```

Результат:

```yaml
# Source: myapp/charts/redis/templates/statefulset.yaml
apiVersion: apps/v1
kind: StatefulSet
# ...
spec:
  template:
    spec:
      priorityClassName: high-priority
```

Здесь все Deployment и StatefulSet чарта `redis` получают priority class из значения `redisPriorityClass` основного чарта: `$Values` содержит values чарта, в котором лежит `patches.yaml`, а не `redis`.

## Игнорирование изменений полей, не требующих развертывания (только в werf)

Чтобы ресурс не обновлялся только из-за поля, которое отличается при каждом рендере (метки времени, случайные значения) или меняется в кластере контроллером, добавьте diff-патч, удаляющий это поле.

Пример:

```yaml
# .helm/patches.yaml:
diffPatches:
- match:
    kinds: [Deployment]
  patch: del(.spec.template.metadata.annotations."example.com/deployed-at")
```

Здесь изменение одной только аннотации `example.com/deployed-at` больше не обновляет Deployment и не перезапускает его Pod'ы. Если меняются и другие поля, Deployment обновляется, и новая аннотация попадает в кластер вместе с ними.

## Источники патчей

По умолчанию werf берёт патчи из `patches.yaml` каждого чарта. Чтобы применять патчи только в одном окружении или патчить чужой чарт, передайте файл того же формата через `--patches` (или `WERF_PATCHES_*`):

```shell
werf converge --patches .helm/patches-production.yaml
```

Патчи применяются по порядку, каждый к результату предыдущего:
1. `patches.yaml` зависимых чартов;
2. `patches.yaml` основного чарта;
3. файлы `--patches`.

Источники отличаются:
* На что действуют: `patches.yaml` — только на ресурсы своего чарта, включая его зависимые чарты; `--patches` — на все ресурсы релиза.
* Чьи данные видят (только render-патчи): `patches.yaml` — `$Values`, `$Release`, `$Chart` и `$Capabilities` своего чарта; `--patches` — основного чарта.

Чтобы игнорировать `patches.yaml` всех чартов, передайте `--no-default-patches`; на файлы `--patches` он не влияет.

`werf rollback` и `werf dismiss` применяют только diff-патчи, причём `patches.yaml` берут из чарта, сохранённого в релизе.

## Формат patches.yaml

```
renderPatches:
- match:
    names: [<имя или /regexp/>, ...]
    namespaces: [<namespace или /regexp/>, ...]
    groups: [<группа или /regexp/>, ...]
    versions: [<версия или /regexp/>, ...]
    kinds: [<kind или /regexp/>, ...]
    charts: [<имя, алиас, путь чарта или /regexp/>, ...]
    labels: {<ключ>: <значение>, ...}
    annotations: {<ключ>: <значение>, ...}
  patch: <программа на jq>
diffPatches:
- <то же, что в renderPatches>
```

Патч применяется к каждому ресурсу, у которого совпали все указанные поля `match`:
* значение вида `/.../` — регулярное выражение, которое должно совпасть со значением целиком; остальные значения сравниваются как есть, а для `kinds`, `groups` и `versions` — без учёта регистра;
* `labels` и `annotations` совпадают, если у ресурса есть все перечисленные метки или аннотации с теми же значениями;
* `charts` совпадает с алиасом чарта, из которого отрендерен ресурс (или с его именем, если алиас не задан), либо с его путём от основного чарта, например `myapp/charts/redis`.

Патч получает ресурс целиком и должен вернуть ровно один объект, поэтому не может добавлять или удалять ресурсы.

Пример:

```yaml
# .helm/patches.yaml:
renderPatches:
- match:
    kinds: [Deployment]
    names: [/backend-.*/]
    labels:
      app: backend
  patch: .metadata.labels.team = $Values.team
```

Здесь метку `team` получают только Deployment, чьё имя начинается с `backend-` и у которых есть метка `app: backend`.
