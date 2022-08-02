---
title: Порядок развертывания
permalink: advanced/helm/deploy_process/deployment_order.html
---

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
    werf.io/weight: "10"
---
kind: Job
metadata:
  name: db-migration
  annotations:
    werf.io/weight: "20"
---
kind: Deployment
metadata:
  name: app
  annotations:
    werf.io/weight: "30"
---
kind: Service
metadata:
  name: app
  annotations:
    werf.io/weight: "30"
```

В приведенном выше примере werf сначала развернет базу данных и дождется ее готовности, затем запустит миграции и дождется их завершения, после чего развернет приложение и связанный с ним сервис.

Полезные ссылки:
* [`werf.io/weight`]({{ "/reference/deploy_annotations.html#resource-weight" | true_relative_url }})
