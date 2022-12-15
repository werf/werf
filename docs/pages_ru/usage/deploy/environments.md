---
title: Разные окружения
permalink: usage/deploy/environments.html
---

## Окружение

По умолчанию, werf предполагает, что каждый релиз должен относиться к какому-либо окружению, например, `staging`, `test` или `production`.

На основании окружения werf определяет:

 1. Имя релиза.
 2. Namespace в Kubernetes.

Передача имени окружения является обязательной для операции деплоя и должна быть выполнена либо с помощью параметра `--env` либо на основании данных используемой CI/CD системы (читай подробнее про [интеграцию c CI/CD системами]({{ "usage/integration_with_ci_cd_systems/how_ci_cd_integration_works/general_overview.html#интеграция-с-настройками-cicd" | true_relative_url }})) определиться автоматически.

### Использование окружения в шаблонах

Текущее окружение werf можно использовать в шаблонах.

Например, вы можете использовать его для создания разных шаблонов для разных окружений:

{% raw %}

```
apiVersion: v1
kind: Secret
metadata:
  name: regsecret
type: kubernetes.io/dockerconfigjson
data:
{{ if eq .Values.werf.env "dev" }}
  .dockerconfigjson: UmVhbGx5IHJlYWxseSByZWVlZWVlZWVlZWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWxsbGxsbGxsbGxsbGxsbGxsbGxsbGxsbGxsbGxsbGx5eXl5eXl5eXl5eXl5eXl5eXl5eSBsbGxsbGxsbGxsbGxsbG9vb29vb29vb29vb29vb29vb29vb29vb29vb25ubm5ubm5ubm5ubm5ubm5ubm5ubm5ubmdnZ2dnZ2dnZ2dnZ2dnZ2dnZ2cgYXV0aCBrZXlzCg==
{{ else }}
  .dockerconfigjson: {{ .Values.dockerconfigjson }}
{{ end }}
```

{% endraw %}

Следует обратить внимание, что значение параметра `--env ENV` доступно не только в шаблонах helm, но и [в шаблонах конфигурации `werf.yaml`]({{ "/reference/werf_yaml_template_engine.html#env" | true_relative_url }}).

Больше информации про сервисные значения доступно [в статье про values]({{ "/usage/deploy/values.html" | true_relative_url }}).

## Namespace в Kubernetes

По умолчанию namespace, используемый в Kubernetes, формируется по шаблону `[[ project ]]-[[ env ]]`, где `[[ project ]]` — [имя проекта]({{ "reference/werf_yaml.html#имя-проекта" | true_relative_url }}), а `[[ env ]]` — имя [окружения](#окружение).

Например, для проекта с именем `symfony-demo` будет сформировано следующее имя namespace в Kubernetes, в зависимости от имени окружения:
* `symfony-demo-stage` для окружения `stage`;
* `symfony-demo-test` для окружения `test`;
* `symfony-demo-prod` для окружения `prod`.

Имя namespace в Kubernetes может быть переопределено с помощью параметра `--namespace NAMESPACE` при деплое. В этом случае werf будет использовать указанное имя как есть, без каких либо преобразований и использования шаблонов.

Имя namespace также можно явно определить в файле конфигурации `werf.yaml`, установив параметр [`deploy.namespace`]({{ "/reference/werf_yaml.html#namespace-в-kubernetes" | true_relative_url }}).

### Слагификация namespace Kubernetes

Сформированное по шаблону имя namespace слагифицируется, чтобы удовлетворять требованиям к [DNS именам](https://www.ietf.org/rfc/rfc1035.txt), в результате чего получается уникальное имя namespace в Kubernetes.

Слагификация имени namespace включена по умолчанию, но может быть отключена указанием параметра [`deploy.namespaceSlug=false`]({{ "/reference/werf_yaml.html#namespace-в-kubernetes" | true_relative_url }}) в файле конфигурации `werf.yaml`.

## Работа с несколькими кластерами Kubernetes

В некоторых случаях, необходима работа с несколькими кластерами Kubernetes для разных окружений. Все что вам нужно, это настроить необходимые [контексты](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters) kubectl для доступа к необходимым кластерам и использовать для werf параметр `--kube-context=CONTEXT`, совместно с указанием окружения.

