---
title: Порядок развертывания
permalink: usage/deploy/deployment_order.html
---

Во время запуска команды `werf converge` werf запускает процесс деплоя, включающий следующие этапы:

 1. Преобразование шаблонов чартов в единый список манифестов ресурсов Kubernetes и их проверка.
 2. Последовательный запуск хуков `pre-install` или `pre-upgrade`, отсортированных по весу, и контроль каждого хука до завершения его работы с выводом логов.
 3. Группировка ресурсов Kubernetes, не относящихся к хукам, по их весу и последовательное развертывание каждой группы в соответствии с ее весом: создание/обновление/удаление ресурсов и отслеживание их до готовности с выводом логов в процессе.
 4. Запуск хуков `post-install` или `post-upgrade` по аналогии с хуками `pre-install` и `pre-upgrade`.

## Вес ресурсов

Все ресурсы по умолчанию применяются (apply) и отслеживаются одновременно, поскольку изначально у них одинаковый вес ([`werf.io/weight: 0`]({{ "/reference/deploy_annotations.html#resource-weight" | true_relative_url }})) (при условии, что вы его не меняли). Можно задать порядок применения (apply) и отслеживания ресурсов, установив для них различные веса.

Перед фазой развертывания werf группирует ресурсы на основе их веса, затем выполняет apply для группы с наименьшим ассоциированным весом и ждет, пока ресурсы из этой группы будут готовы. После успешного развертывания всех ресурсов из этой группы, werf переходит к развертыванию ресурсов из группы со следующим наименьшим весом. Процесс продолжается до тех пор, пока не будут развернуты все ресурсы релиза.

> Обратите внимание, что "werf.io/weight" работает только для ресурсов, не относящихся к хукам. Для хуков следует использовать "helm.sh/hook-weight".

Давайте рассмотрим следующий пример:
```yaml
kind: Job
metadata:
  name: db-migration
---
kind: StatefulSet
metadata:
  name: postgres
---
kind: Deployment
metadata:
  name: app
---
kind: Service
metadata:
  name: app
```

Все ресурсы из примера выше будут развертываться одновременно, поскольку у них одинаковый вес по умолчанию (0). Но что если базу данных необходимо развернуть до выполнения миграций, а приложение, в свою очередь, должно запуститься только после их завершения? Что ж, у этой проблемы есть решение! Попробуйте сделать так:
```yaml
kind: StatefulSet
metadata:
  name: postgres
  annotations:
    werf.io/weight: "-2"
---
kind: Job
metadata:
  name: db-migration
  annotations:
    werf.io/weight: "-1"
---
kind: Deployment
metadata:
  name: app
---
kind: Service
metadata:
  name: app
```

В приведенном выше примере werf сначала развернет базу данных и дождется ее готовности, затем запустит миграции и дождется их завершения, после чего развернет приложение и связанный с ним сервис.

Полезные ссылки:
* [`werf.io/weight`]({{ "/reference/deploy_annotations.html#resource-weight" | true_relative_url }})

## Helm-хуки

Helm-хуки — произвольные ресурсы Kubernetes, помеченные специальной аннотацией `helm.sh/hook`. Например:

```yaml
kind: Job
metadata:
  name: somejob
  annotations:
    "helm.sh/hook": pre-upgrade,pre-install
    "helm.sh/hook-weight": "1"
```

Существует много разных helm-хуков, влияющих на процесс деплоя. Хуки `pre|post-install|upgrade` наиболее часто используются для выполнения таких задач, как миграция (в хуках `pre-upgrade`) или выполнении некоторых действий после деплоя. Полный список доступных хуков можно найти в соответствующей документации [Helm](https://helm.sh/docs/topics/charts_hooks/).

Хуки сортируются в порядке возрастания согласно значению аннотации `helm.sh/hook-weight` (хуки с одинаковым весом сортируются по имени в алфавитном порядке), после чего хуки последовательно создаются и выполняются. werf пересоздает ресурс Kubernetes для каждого хука, в случае когда ресурс уже существует в кластере. Созданные ресурсы-хуки не удаляются после выполнения, если не указано [специальной аннотации `"helm.sh/hook-delete-policy": hook-succeeded,hook-failed`](https://helm.sh/docs/topics/charts_hooks/).

## Внешние зависимости

Чтобы сделать один ресурс релиза зависимым от другого, можно изменить его порядок развертывания. После этого зависимый ресурс будет развертываться только после успешного развертывания основного. Но что делать, когда ресурс релиза должен зависеть от ресурса, который не является частью данного релиза или даже не управляется werf (например, создан неким оператором)?

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
