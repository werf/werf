---
title: Управление релизами
permalink: usage/deploy/releases.html
---

В то время как чарт — набор конфигурационных файлов вашего приложения, релиз (**release**) — это объект времени выполнения, экземпляр вашего приложения, развернутого с помощью werf.

## Хранение релизов

Информация о каждой версии релиза хранится в самом кластере Kubernetes. werf поддерживает сохранение в произвольном namespace в объектах Secret или ConfigMap.

По умолчанию, werf хранит информацию о релизах в объектах Secret в целевом namespace, куда происходит деплой приложения. Это полностью совместимо с конфигурацией по умолчанию по хранению релизов в [Helm 3](https://helm.sh), что полностью совместимо с конфигурацией [Helm 2](https://helm.sh) по умолчанию. Место хранения информации о релизах может быть указано при деплое с помощью параметров werf: `--helm-release-storage-namespace=NS` и `--helm-release-storage-type=configmap|secret`.

Для получения информации обо всех созданных релизах можно использовать команду [werf helm list]({{ "reference/cli/werf_helm_list.html" | true_relative_url }}), а для просмотра истории конкретного релиза [werf helm history]({{ "reference/cli/werf_helm_history.html" | true_relative_url }}).

### Замечание о совместимости с Helm

werf полностью совместим с уже установленным Helm 2, т.к. хранение информации о релизах осуществляется одним и тем же образом, как и в Helm. Если вы используете в Helm специфичное место хранения информации о релизах, а не значение по умолчанию, то вам нужно указывать место хранения с помощью опций werf `--helm-release-storage-namespace` и `--helm-release-storage-type`.

Информация о релизах, созданных с помощью werf, может быть получена с помощью Helm, например, командами `helm list` и `helm get`. С помощью werf также можно обновлять релизы, развернутые ранее с помощью Helm.

Команда [`werf helm list -A`]({{ "reference/cli/werf_helm_list.html" | true_relative_url }}) выводит список релизов созданных werf или Helm 3. Релизы, созданные через werf могут свободно просматриваться через утилиту helm командами `helm list` или `helm get` и другими.

### Совместимость с Helm 2

Существующие релизы helm 2 (созданные например через werf v1.1) могут быть конвертированы в helm 3 либо автоматически во время работы команды [`werf converge`]({{ "/reference/cli/werf_converge.html" | true_relative_url }}), либо с помощью команды [`werf helm migrate2to3`]({{ "/reference/cli/werf_helm_migrate2to3.html" | true_relative_url }}).

## Принятие ресурсов в релиз

По умолчанию Helm и werf позволяют управлять лишь теми ресурсами, которые были созданы самим Helm или werf в рамках релиза. При попытке выката чарта с манифестом ресурса, который уже существует в кластере и который был создан не с помощью Helm или werf, возникнет следующая ошибка:

```
Error: helm upgrade have failed: UPGRADE FAILED: rendered manifests contain a resource that already exists. Unable to continue with update: KIND NAME in namespace NAMESPACE exists and cannot be imported into the current release: invalid ownership metadata; label validation error: missing key "app.kubernetes.io/managed-by": must be set to "Helm"; annotation validation error: missing key "meta.helm.sh/release-name": must be set to RELEASE_NAME; annotation validation error: missing key "meta.helm.sh/release-namespace": must be set to RELEASE_NAMESPACE
```

Данная ошибка предотвращает деструктивное поведение, когда некоторый уже существующий релиз случайно становится частью выкаченного релиза.

Однако если данное поведение является желаемым, то требуется отредактировать целевой ресурс в кластере и добавить в него следующие аннотации и лейблы:

```yaml
metadata:
  annotations:
    meta.helm.sh/release-name: "RELEASE_NAME"
    meta.helm.sh/release-namespace: "NAMESPACE"
  labels:
    app.kubernetes.io/managed-by: Helm
```

После следующего деплоя этот ресурс будет принят в новую ревизию релиза и соотвественно будет приведен к тому состоянию, которое описано в манифесте чарта.

## Имя релиза

По умолчанию название релиза формируется по шаблону `[[project]]-[[env]]`. Где `[[ project ]]` — имя [проекта]({{ "reference/werf_yaml.html#имя-проекта" | true_relative_url }}), а `[[ env ]]` — имя [окружения]({{ "/usage/deploy/environments.html#окружение" | true_relative_url }}).

Например, для проекта с именем `symfony-demo` будет сформировано следующее имя релиза в зависимости от имени окружения:
* `symfony-demo-stage` для окружения `stage`;
* `symfony-demo-test` для окружения `test`;
* `symfony-demo-prod` для окружения `prod`.

Имя релиза может быть переопределено с помощью параметра `--release NAME` при деплое. В этом случае werf будет использовать указанное имя как есть, без каких либо преобразований и использования шаблонов.

Имя релиза также можно явно определить в файле конфигурации `werf.yaml`, установив параметр [`deploy.helmRelease`]({{ "/reference/werf_yaml.html#имя-релиза" | true_relative_url }}).

### Слагификация имени релиза

Сформированное по шаблону имя Helm-релиза слагифицируется, в результате чего получается уникальное имя Helm-релиза.

Слагификация имени Helm-релиза включена по умолчанию, но может быть отключена указанием параметра [`deploy.helmReleaseSlug=false`]({{ "/reference/werf_yaml.html#имя-релиза" | true_relative_url }}) в файле конфигурации `werf.yaml`.

