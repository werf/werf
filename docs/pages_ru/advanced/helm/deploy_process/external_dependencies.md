---
title: Внешние зависимости
permalink: advanced/helm/deploy_process/external_dependencies.html
---

Чтобы сделать один ресурс релиза зависимым от другого, можно изменить его [порядок развертывания]({{ "/advanced/helm/deploy_process/deployment_order.html" | true_relative_url }}). После этого зависимый ресурс будет развертываться только после успешного развертывания основного. Но что делать, когда ресурс релиза должен зависеть от ресурса, который не является частью данного релиза или даже не управляется werf (например, создан неким оператором)?

В этом случае можно воспользоваться аннотацией [`<name>.external-dependency.werf.io/resource`]({{ "/reference/deploy_annotations.html#external-dependency-resource" | true_relative_url }}). Ресурс с данной аннотацией не будет развернут, пока не будет создана и готова заданная внешняя зависимость.

Пример:
```yaml
kind: Deployment
metadata:
  name: app
  annotations:
    secret.external-dependency.werf.io/resource: secret/dynamic-vault-secret
```

В приведенном выше примере werf дождется создания `dynamic-vault-secret`, прежде чем приступить к развертыванию `app`. Мы исходим из предположения, что `dynamic-vault-secret` создается оператором из инстанса Vault и не управляется werf.

Давайте рассмотрим еще один пример:
```yaml
kind: Deployment
metadata:
  name: service1
  annotations:
    service2.external-dependency.werf.io/resource: deployment/service2
    service2.external-dependency.werf.io/namespace: service2-production
```

В данном случае werf дождется успешного развертывания `service2` в другом пространстве имен, прежде чем приступить к развертыванию `service1`. Обратите внимание, что `service2` может развертываться либо как часть другого релиза werf, либо управляться иными инструментами в рамках CI/CD (т.е. находиться вне контроля werf).

Полезные ссылки:
* [`<name>.external-dependency.werf.io/resource`]({{ "/reference/deploy_annotations.html#external-dependency-resource" | true_relative_url }})
* [`<name>.external-dependency.werf.io/namespace`]({{ "/reference/deploy_annotations.html#external-dependency-namespace" | true_relative_url }})
