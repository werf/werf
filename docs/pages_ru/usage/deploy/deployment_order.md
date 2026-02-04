---
title: Порядок развертывания
permalink: usage/deploy/deployment_order.html
---

## Стадии развертывания

Kubernetes-ресурсы развертываются в следующем порядке:

1. Развертывание `CustomResourceDefinitions`.
1. Развертывание ресурсов `pre-install`, `pre-upgrade`, `pre-rollback` в порядке возрастания их веса (`werf.io/weight`, `helm.sh/hook-weight`).
1. Развертывание ресурсов `install`, `upgrade`, `rollback` в порядке возрастания их веса.
1. Развертывание ресурсов `post-install`, `post-upgrade`, `post-rollback` в порядке возрастания их веса.

Ресурсы с одинаковым весом группируются и развертываются параллельно. Вес по умолчанию — 0. Ресурсы с аннотацией `werf.io/deploy-dependency-<name>` развертываются сразу, как только их зависимости удовлетворены, но в рамках своей стадии (pre, main или post).

## Развертывание CustomResourceDefinitions

Чтобы развернуть CustomResourceDefinitions, поместите их манифесты в нешаблонизируемые файлы `crds/*.yaml` в любом из подключенных чартов. Во время развертывания CRDs всегда развертываются раньше любых других ресурсов.

Пример:

```yaml
# .helm/crds/crontab.yaml:
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
spec:
  names:
    kind: CronTab
# ...
```

```yaml
# .helm/templates/crontab.yaml:
apiVersion: example.org/v1
kind: CronTab
# ...
```

В этом случае, сначала будет развернут CRD для ресурса CronTab, за которым последует развертывание самого ресурса CronTab.

## Задание порядка развертывания

### Задание порядка весами (только werf)

Аннотация `werf.io/weight` может быть использована для настройки порядка развертывания ресурсов. По умолчанию ресурсы имеют вес 0. Ресурсы с меньшим весом развертываются раньше ресурсов с большим весом. Если у ресурсов одинаковый вес, они развертываются параллельно.

Посмотрите на пример:

```yaml
# .helm/templates/example.yaml:
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: database
  annotations:
    werf.io/weight: "-1"
# ...
---
apiVersion: batch/v1
kind: Job
metadata:
  name: database-migrations
# ...
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app1
  annotations:
    werf.io/weight: "1"
# ...
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app2
  annotations:
    werf.io/weight: "1"
# ...
```

В этом случае, ресурс `database` будет развернут первым, за ним последует `database-migrations`, а затем `app1` и `app2` будут развертываться параллельно.

### Задание порядка зависимостями (только werf)

Аннотация `werf.io/deploy-dependency-<name>` может быть использована для настройки порядка развертывания ресурсов. Ресурс с такой аннотацией будет развернут сразу, как только все его зависимости будут удовлетворены, но в рамках своей стадии (pre, main или post). Вес ресурса игнорируется, когда используется эта аннотация.

Например:

```yaml
# .helm/templates/example.yaml:
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: database
# ...
---
apiVersion: batch/v1
kind: Job
metadata:
  name: database-migrations
  annotations:
    werf.io/deploy-dependency-db: state=ready,kind=StatefulSet,name=database
# ...
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app1
  annotations:
    werf.io/deploy-dependency-migrations: state=ready,kind=Job,name=database-migrations
# ...
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app2
  annotations:
    werf.io/deploy-dependency-migrations: state=ready,kind=Job,name=database-migrations
# ...
```

В этом случае, ресурс `database` будет развернут первым, за ним последует `database-migrations`, а затем `app1` и `app2` будут развертываться параллельно.

Посмотрите все возможности этой аннотации [здесь]({{ "/reference/deploy_annotations.html#resource-dependencies" | true_relative_url }}).

Это более гибкий и эффективный способ установить порядок развертывания ресурсов по сравнению с `werf.io/weight` и другими методами, так как он позволяет развертывать ресурсы в порядке, подобном графу.

### Задание порядка удаления (только werf)

Аннотация `werf.io/delete-dependency-<name>` может быть использована для настройки порядка удаления ресурсов. Ресурс с такой аннотацией будет удалён только после того, как все его зависимости будут удалены.

Например:

```yaml
# .helm/templates/example.yaml:
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
# ...
---
apiVersion: v1
kind: Service
metadata:
  name: app
  annotations:
    werf.io/delete-dependency-ingress: state=absent,kind=Ingress,group=networking.k8s.io,version=v1,name=app
# ...
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: app
  annotations:
    werf.io/delete-dependency-service: state=absent,kind=Service,version=v1,name=app
# ...
```

В этом примере, при удалении релиза, сначала будет удалён ресурс `Ingress`, затем `Service`, и только потом `Deployment`.

Посмотрите все возможности этой аннотации [здесь]({{ "/reference/deploy_annotations.html#delete-dependencies" | true_relative_url }}).

## Ожидание готовности ресурсов вне релиза (только werf)

Ресурсы, развертывающиеся в рамках текущего релиза, могут зависеть от ресурсов, которые не принадлежат этому релизу. werf может ждать, пока эти внешние ресурсы не станут готовыми — вам просто нужно добавить аннотацию `<name>.external-dependency.werf.io/resource` следующим образом:

```yaml
# .helm/templates/example.yaml:
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  annotations:
    secret.external-dependency.werf.io/resource: secret/my-dynamic-vault-secret
# ...
```

Развертывание `myapp` начнётся только после того, как секрет `my-dynamic-vault-secret` (который автоматически создаётся оператором в кластере) будет создан и готов.

Вот как вы можете настроить werf на ожидание готовности нескольких внешних ресурсов:

```yaml
# .helm/templates/example.yaml:
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  annotations:
    secret.external-dependency.werf.io/resource: secret/my-dynamic-vault-secret
    database.external-dependency.werf.io/resource: statefulset/my-database
# ...
```

По умолчанию, werf ищет внешний ресурс в namespace релиза (если, конечно, ресурс является namespaced). Вы можете изменить namespace внешнего ресурса, прикрепив к нему аннотацию `<name>.external-dependency.werf.io/namespace`:

```yaml
# .helm/templates/example.yaml:
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  annotations:
    secret.external-dependency.werf.io/resource: secret/my-dynamic-vault-secret
    secret.external-dependency.werf.io/namespace: my-namespace
```
