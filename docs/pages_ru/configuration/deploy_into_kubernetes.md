---
title: Деплой в Kubernetes
sidebar: documentation
permalink: configuration/deploy_into_kubernetes.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

## Имя релиза

werf позволяет определять пользовательский шаблон имени Helm-релиза, который используется во время [процесса деплоя]({{ site.baseurl }}/reference/deploy_process/deploy_into_kubernetes.html#имя-релиза) для генерации имени релиза.

Пользовательский шаблон имени релиза определяется в [секции мета-информации]({{ site.baseurl }}/configuration/introduction.html#секция-мета-информации) в файле `werf.yaml`:

```yaml
project: PROJECT_NAME
configVersion: 1
deploy:
  helmRelease: TEMPLATE
  helmReleaseSlug: false
```

`deploy.helmReleaseSlug` включает или отключает [слагификацию]({{ site.baseurl }}/reference/deploy_process/deploy_into_kubernetes.html#слагификация-имени-релиза) имени Helm-релиза. Включен по умолчанию.

В качестве значения для `deploy.helmRelease` указывается Go-шаблон с разделителями `[[` и `]]`. Поддерживаются функции `project` и `env`. Значение шаблона имени релиза по умолчанию: `[[ project ]]-[[ env ]]`.

Т.к. в качестве значения для `deploy.helmRelease` указывается Go-шаблон, то возможно использование соответствующих [функций werf]({{ site.baseurl }}/configuration/introduction.html#шаблоны-go) (впрочем, как и для любых других параметров в конфигурации). Например, вы можете комбинировать переменные доступные в Go-шаблоне с переменными окружения следующим образом:
{% raw %}
```yaml
deploy:
  helmRelease: >-
    [[ project ]]-{{ env "HELM_RELEASE_EXTRA" }}-[[ env ]]
```
{% endraw %}

## Namespace в Kubernetes 

werf позволяет определять пользовательский шаблон namespace в Kubernetes, который будет использоваться во время [процесса деплоя]({{ site.baseurl }}/reference/deploy_process/deploy_into_kubernetes.html#kubernetes-namespace) для генерации имени namespace.

Пользовательский шаблон namespace Kubernetes определяется в [секции мета-информации]({{ site.baseurl }}/configuration/introduction.html#секция-мета-информации) в файле `werf.yaml`:


```yaml
project: PROJECT_NAME
configVersion: 1
deploy:
  namespace: TEMPLATE
  namespaceSlug: true|false
```

В качестве значения для `deploy.namespace` указывается Go-шаблон с разделителями `[[` и `]]`. Поддерживаются функции `project` и `env`. Значение шаблона имени namespace по умолчанию: `[[ project ]]-[[ env ]]`.

`deploy.namespaceSlug` включает или отключает [слагификацию]({{ site.baseurl }}/reference/deploy_process/deploy_into_kubernetes.html#слагификация-namespace-kubernetes) имени namespace Kubernetes. Включен по умолчанию.
