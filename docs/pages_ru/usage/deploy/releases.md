---
title: Управление релизами
permalink: usage/deploy/releases.html
---

## О релизах

Результатом развертывания является *релиз* — совокупность развернутых в кластере ресурсов и служебной информации.

Технически релизы werf являются релизами Helm 3 и полностью с ними совместимы. Служебная информация по умолчанию хранится в специальном Secret-ресурсе.

## Автоматическое формирование имени релиза (только в werf)

По умолчанию имя релиза формируется автоматически по специальному шаблону `[[ project ]]-[[ env ]]`, где `project` — имя проекта werf, а `env` — имя окружения, например:

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

## Автоматическое аннотирование выкатываемых ресурсов релиза

werf автоматически выставляет следующие аннотации всем ресурсам чарта в процессе развёртывания:

* `"werf.io/version": FULL_WERF_VERSION` — версия werf, использованная в процессе запуска команды `werf converge`;
* `"project.werf.io/name": PROJECT_NAME` — имя проекта, указанное в файле конфигурации `werf.yaml`;
* `"project.werf.io/env": ENV` — имя окружения, указанное с помощью параметра `--env` или переменной окружения `WERF_ENV` (аннотация не устанавливается, если окружение не было указано при запуске).

При использовании команды `werf ci-env` с [поддерживаемыми CI/CD системами]({{ "usage/integration_with_ci_cd_systems.html" | true_relative_url }}) добавляются аннотации, которые позволяют пользователю перейти в связанный пайплайн, задание и коммит при необходимости.

## Добавление произвольных аннотаций и лейблов для выкатываемых ресурсов релиза

Пользователь может устанавливать произвольные аннотации и лейблы используя CLI-параметры при развёртывании `--add-annotation annoName=annoValue` (может быть указан несколько раз) и `--add-label labelName=labelValue` (может быть указан несколько раз). Аннотации и лейблы так же могут быть заданы с помощью соответствующих переменных `WERF_ADD_LABEL*` и `WERF_ADD_ANNOTATION*` (к примеру, `WERF_ADD_ANNOTATION_1=annoName1=annoValue1` и `WERF_ADD_LABEL_1=labelName1=labelValue1`).

Например, для установки аннотаций и лейблов `commit-sha=9aeee03d607c1eed133166159fbea3bad5365c57`, `gitlab-user-email=vasya@myproject.com` всем ресурсам Kubernetes в чарте, можно использовать следующий вызов команды деплоя:

```shell
werf converge \
  --add-annotation "commit-sha=9aeee03d607c1eed133166159fbea3bad5365c57" \
  --add-label "commit-sha=9aeee03d607c1eed133166159fbea3bad5365c57" \
  --add-annotation "gitlab-user-email=vasya@myproject.com" \
  --add-label "gitlab-user-email=vasya@myproject.com" \
  --env dev \
  --repo REPO
```

