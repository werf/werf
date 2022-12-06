---
title: werf.yaml
permalink: reference/werf_yaml.html
description: Пример конфигурации werf
toc: false
---

{% include reference/werf_yaml/table.html %}

## Имя проекта

Имя проекта должно быть уникальным в пределах группы проектов, собираемых на одном сборочном узле и развертываемых на один и тот же кластер Kubernetes (например, уникальным в пределах всех групп одного GitLab).

Имя проекта должно быть не более 50 символов, содержать только строчные буквы латинского алфавита, цифры и знак дефиса.

### Последствия смены имени проекта
 
**ВНИМАНИЕ**. Никогда не меняйте имя проекта в процессе работы, если вы не осознаете всех последствий.

Смена имени проекта приводит к следующим проблемам:

1. Инвалидация сборочного кэша. Все образы должны быть собраны повторно, а старые удалены из локального хранилища или container registry вручную.
2. Создание совершенно нового Helm-релиза. Смена имени проекта и повторное развертывание приложения приведет к созданию еще одного экземпляра, если вы уже развернули ваше приложение в кластере Kubernetes.

werf не поддерживает изменение имени проекта и все возникающие проблемы должны быть разрешены вручную.

## Выкат

### Директория helm chart 

Директива позволяет указать произвольную директорию до helm чарта проекта, например `.deploy/chart`:

```yaml
deploy:
  helmChartDir: .deploy/chart
```

### Имя релиза

werf позволяет определять пользовательский шаблон имени Helm-релиза, который используется во время [процесса деплоя]({{ "/advanced/helm/releases/naming.html#имя-релиза" | true_relative_url }}) для генерации имени релиза:

```yaml
project: PROJECT_NAME
configVersion: 1
deploy:
  helmRelease: TEMPLATE
  helmReleaseSlug: false
```

В качестве значения для `deploy.helmRelease` указывается Go-шаблон с разделителями `[[` и `]]`. Поддерживаются функции `[[ project ]]`, `[[ env ]]` и `[[ namespace ]]`. Значение шаблона имени релиза по умолчанию: `[[ project ]]-[[ env ]]`.

Можно комбинировать переменные доступные в Go-шаблоне с переменными окружения следующим образом:

{% raw %}
```yaml
deploy:
  helmRelease: >-
    [[ project ]]-[[ namespace ]]-[[ env ]]-{{ env "HELM_RELEASE_EXTRA" }}
```
{% endraw %}

**Замечание** Использование переменной окружения `HELM_RELEASE_EXTRA` в данном случае должно быть явно разрешено в конфиге [werf-giterminism.yaml]({{ "reference/werf_giterminism_yaml.html" | true_relative_url }}).

`deploy.helmReleaseSlug` включает или отключает [слагификацию]({{ "/advanced/helm/releases/naming.html#слагификация-имени-релиза" | true_relative_url }}) имени Helm-релиза (включен по умолчанию).

### Namespace в Kubernetes 

werf позволяет определять пользовательский шаблон namespace в Kubernetes, который будет использоваться во время [процесса деплоя]({{ "/advanced/helm/releases/naming.html#namespace-в-kubernetes" | true_relative_url }}) для генерации имени namespace.

Пользовательский шаблон namespace Kubernetes определяется в секции мета-информации в файле `werf.yaml`:

```yaml
project: PROJECT_NAME
configVersion: 1
deploy:
  namespace: TEMPLATE
  namespaceSlug: true|false
```

В качестве значения для `deploy.namespace` указывается Go-шаблон с разделителями `[[` и `]]`. Поддерживаются функции `project` и `env`. Значение шаблона имени namespace по умолчанию: `[[ project ]]-[[ env ]]`.

`deploy.namespaceSlug` включает или отключает [слагификацию]({{ "/advanced/helm/releases/naming.html#слагификация-namespace-kubernetes" | true_relative_url }}) имени namespace Kubernetes. Включен по умолчанию.

## Git worktree

Для корректной работы сборщика stapel werf-у требуется полная git-история проекта, чтобы работать в наиболее эффективном режиме. Поэтому по умолчанию werf выполняет fetch истории для текущего git проекта, когда это требуется. Это означает, что werf может автоматически сконвертировать shallow-clone репозитория в полный clone и скачать обновлённый список веток и тегов из origin в процессе очистки образов. 

Поведение по умолчанию описывается следующими настройками:

```yaml
gitWorktree:
  forceShallowClone: false
  allowUnshallow: true
  allowFetchOriginBranchesAndTags: true
```

Чтобы, например, выключить автоматический unshallow рабочей директории git, необходимы следующие настройки:

```yaml
gitWorktree:
  forceShallowClone: true
  allowUnshallow: false
```

