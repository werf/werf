---
title: Разные окружения
permalink: usage/deploy/environments.html

---

## Параметризация шаблонов в зависимости от окружения (только в werf)

*Окружение* werf указывается опцией `--env` (`$WERF_ENV`), либо автоматически выставляется командой `werf ci-env`. Текущее окружение доступно в `$.Values.global.werf.env` во всех чартах (основном и зависимых).

Окружение werf используется при формировании имени релиза и имени Namespace'а, а также может использоваться для параметризации шаблонов:

```yaml
# .helm/values.yaml:
memory:
  staging: 1G
  production: 2G
```

{% raw %}

```
# .helm/templates/example.yaml:
memory: {{ index $.Values.memory $.Values.global.werf.env }}
```

{% endraw %}

```shell
werf render --env production
```

Результат:

```yaml
memory: 2G
```

Поскольку `$.Values.global.werf` имеет глобальную область видимости, `env` доступен во всех зависимых чартах.

## Развертывание в разные Kubernetes Namespace

Имя Kubernetes Namespace для развертываемых ресурсов формируется автоматически (только в werf) по специальному шаблону `[[ project ]]-[[ env ]]`, где `[[ project ]]` — имя проекта werf, а `[[ env ]]` — имя окружения.

Достаточно изменить окружение werf и вместе с ним изменится и Namespace:

```yaml
# werf.yaml:
project: myapp
```

```shell
werf converge --env staging
werf converge --env production
```

Результат: один экземпляр приложения развёрнут в Namespace `myapp-staging`, а второй — в `myapp-production`.

Обратите внимание, что если в манифесте Kubernetes-ресурса явно указан Namespace, то для этого ресурса будет использован именно указанный в нём Namespace.

### Изменение шаблона имени Namespace (только в werf)

Если вас не устраивает специальный шаблон, из которого формируется имя Namespace, вы можете его изменить:

```yaml
# werf.yaml:
project: myapp
deploy:
  namespace: "backend-[[ env ]]"
```

```shell
werf converge --env production
```

Результат: приложение развёрнуто в Namespace `backend-production`.

### Прямое указание имени Namespace

Вместо формирования имени Namespace по специальному шаблону можно указывать Namespace явно для каждой команды (рекомендуется также изменять и имя релиза):

```shell
werf converge --namespace backend-production --release backend-production
```

Результат: приложение развёрнуто в Namespace `backend-production`.

### Форматирование имени Namespace

Namespace, сформированный по специальному шаблону или указанный опцией `--namespace`, приводится к формату [RFC 1123 Label Names](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names) автоматически. Отключить автоматическое форматирование можно директивой `deploy.namespaceSlug` файла `werf.yaml`.

Вручную отформатировать любую строку согласно формату RFC 1123 Label Names можно командой `werf slugify -f kubernetes-namespace`.
