---
title: Именование
permalink: documentation/advanced/helm/releases/naming.html
---

## Окружение

По умолчанию, werf предполагает, что каждый релиз должен относиться к какому-либо окружению, например, `staging`, `test` или `production`.

На основании окружения werf определяет:

 1. Имя релиза.
 2. Namespace в Kubernetes.

Передача имени окружения является обязательной для операции деплоя и должна быть выполнена либо с помощью параметра `--env` либо на основании данных используемой CI/CD системы (читай подробнее про [интеграцию c CI/CD системами]({{ "documentation/internals/how_ci_cd_integration_works/general_overview.html#интеграция-с-настройками-cicd" | true_relative_url }})) определиться автоматически.

## Имя релиза

По умолчанию название релиза формируется по шаблону `[[project]]-[[env]]`. Где `[[ project ]]` — имя [проекта]({{ "documentation/reference/werf_yaml.html#имя-проекта" | true_relative_url }}), а `[[ env ]]` — имя [окружения](#окружение).

Например, для проекта с именем `symfony-demo` будет сформировано следующее имя релиза в зависимости от имени окружения:
* `symfony-demo-stage` для окружения `stage`;
* `symfony-demo-test` для окружения `test`;
* `symfony-demo-prod` для окружения `prod`.

Имя релиза может быть переопределено с помощью параметра `--release NAME` при деплое. В этом случае werf будет использовать указанное имя как есть, без каких либо преобразований и использования шаблонов.

Имя релиза также можно явно определить в файле конфигурации `werf.yaml`, установив параметр [`deploy.helmRelease`]({{ "/documentation/reference/werf_yaml.html#имя-релиза" | true_relative_url }}).

### Слагификация имени релиза

Сформированное по шаблону имя Helm-релиза [слагифицируется]({{ "documentation/internals/names_slug_algorithm.html#базовый-алгоритм" | true_relative_url }}), в результате чего получается уникальное имя Helm-релиза.

Слагификация имени Helm-релиза включена по умолчанию, но может быть отключена указанием параметра [`deploy.helmReleaseSlug=false`]({{ "/documentation/reference/werf_yaml.html#имя-релиза" | true_relative_url }}) в файле конфигурации `werf.yaml`.

## Namespace в Kubernetes

По умолчанию namespace, используемый в Kubernetes, формируется по шаблону `[[ project ]]-[[ env ]]`, где `[[ project ]]` — [имя проекта]({{ "documentation/reference/werf_yaml.html#имя-проекта" | true_relative_url }}), а `[[ env ]]` — имя [окружения](#окружение).

Например, для проекта с именем `symfony-demo` будет сформировано следующее имя namespace в Kubernetes, в зависимости от имени окружения:
* `symfony-demo-stage` для окружения `stage`;
* `symfony-demo-test` для окружения `test`;
* `symfony-demo-prod` для окружения `prod`.

Имя namespace в Kubernetes может быть переопределено с помощью параметра `--namespace NAMESPACE` при деплое. В этом случае werf будет использовать указанное имя как есть, без каких либо преобразований и использования шаблонов.

Имя namespace также можно явно определить в файле конфигурации `werf.yaml`, установив параметр [`deploy.namespace`]({{ "/documentation/reference/werf_yaml.html#namespace-в-kubernetes" | true_relative_url }}).

### Слагификация namespace Kubernetes

Сформированное по шаблону имя namespace [слагифицируется]({{ "documentation/internals/names_slug_algorithm.html#базовый-алгоритм" | true_relative_url }}), чтобы удовлетворять требованиям к [DNS именам](https://www.ietf.org/rfc/rfc1035.txt), в результате чего получается уникальное имя namespace в Kubernetes.

Слагификация имени namespace включена по умолчанию, но может быть отключена указанием параметра [`deploy.namespaceSlug=false`]({{ "/documentation/reference/werf_yaml.html#namespace-в-kubernetes" | true_relative_url }}) в файле конфигурации `werf.yaml`.
