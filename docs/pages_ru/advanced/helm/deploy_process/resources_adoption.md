---
title: Принятие ресурсов в релиз
permalink: advanced/helm/deploy_process/resources_adoption.html
---

По умолчанию Helm и werf позволяют управлять лишь теми ресурсами, которые были созданы самим Helm или werf в рамках релиза. При попытке выката чарта с манифестом ресурса, который уже существует в кластере и который был создан не с помощью Helm или werf, тогда возникнет следующая ошибка:

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
