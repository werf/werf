---
title: Управление релизами
permalink: usage/deploy/releases.html
---

## О релизах

Результатом развертывания является *релиз* — совокупность развернутых в кластере ресурсов и служебной информации. 

Технически релизы werf являются релизами Helm 3 и полностью с ними совместимы. Служебная информация по умолчанию хранится в специальном Secret-ресурсе.

## Автоматическое формирование имени релиза (только в werf)

По умолчанию имя релиза формируется автоматически по специальному шаблону `[[ project ]]-[[ env ]]`, где `[[ project ]]` — имя проекта werf, а `[[ env ]]` — имя окружения, например:

```yaml
# werf.yaml:
project: myapp
```

```shell
werf converge --env staging
werf converge --env production
```

Результат: созданы релизы `myapp-staging` и `myapp-production`.

## Изменение шаблона имени релиза (только в werf)

Если вас не устраивает специальный шаблон, из которого формируется имя релиза, вы можете его изменить:

```yaml
# werf.yaml:
project: myapp
deploy:
  helmRelease: "backend-[[ env ]]"
```

```shell
werf converge --env production
```

Результат: создан релиз `backend-production`.

## Прямое указание имени релиза

Вместо формирования имени релиза по специальному шаблону можно указывать имя релиза явно для каждой команды:

```shell
werf converge --release backend-production  # или $WERF_RELEASE=...
```

Результат: создан релиз `backend-production`.

## Форматирование имени релиза

Имя релиза, сформированное по специальному шаблону или указанное опцией `--release`, приводится к формату [RFC 1123 Label Names](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names) автоматически. Отключить автоматическое форматирование можно директивой `deploy.helmReleaseSlug` файла `werf.yaml`.

Вручную отформатировать любую строку согласно формату RFC 1123 Label Names можно командой `werf slugify -f helm-release`.

## Добавление в релиз уже существующих в кластере ресурсов

werf не позволяет развернуть новый ресурс релиза поверх уже существующего в кластере ресурса, если ресурс в кластере *не является частью текущего релиза*. Такое поведение предотвращает случайные обновления ресурсов, принадлежащих другому релизу или развернутых без werf. Если все же попытаться это сделать, то отобразится следующая ошибка:

```
Error: helm upgrade have failed: UPGRADE FAILED: rendered manifests contain a resource that already exists...
```

Чтобы добавить ресурс в кластере в текущий релиз и разрешить его обновление, выставьте ресурсу в кластере аннотации `meta.helm.sh/release-name: <имя текущего релиза>`, `meta.helm.sh/release-namespace: <Namespace текущего релиза>` и лейбл `app.kubernetes.io/managed-by: Helm`, например:

```shell
kubectl annotate deploy/myapp meta.helm.sh/release-name=myapp-production
kubectl annotate deploy/myapp meta.helm.sh/release-namespace=myapp-production
kubectl label deploy/myapp app.kubernetes.io/managed-by=Helm
```

... после чего перезапустите развертывание:

```shell
werf converge
```

Результат: ресурс релиза `myapp` успешно обновил ресурс `myapp` в кластере и теперь ресурс в кластере является частью текущего релиза.
