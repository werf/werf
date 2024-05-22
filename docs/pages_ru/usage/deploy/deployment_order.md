---
title: Порядок развертывания
permalink: usage/deploy/deployment_order.html
---

## Стадии развертывания

Развертывание Kubernetes-ресурсов происходит в следующей последовательности:

1. Развертывание `CustomResourceDefinitions` из директорий `crds` подключенных чартов.

2. Развертывание хуков `pre-install`, `pre-upgrade` или `pre-rollback`, от хуков с меньшим весом к большему.

3. Развертывание основных ресурсов: объединение ресурсов с одинаковым весом в группы (ресурсы без указанного веса имеют вес 0) и развертывание по одной группе за раз, от групп с ресурсами меньшего веса к группам с ресурсами большего веса.

4. Развертывание хуков `post-install`, `post-upgrade` или `post-rollback`, от хуков с меньшим весом к большему.

Ресурсы с аннотацией `werf.io/deploy-dependency-<name>` не принадлежат ни к одной стадии и развертываются сразу после того, как их зависимости удовлетворены.

## Развертывание CustomResourceDefinitions

Для развертывания CustomResourceDefinitions поместите CRD-манифесты в нешаблонизируемые файлы `crds/*.yaml` в любом из подключенных чартов. При следующем развертывании эти CRD будут развернуты первыми, а хуки и основные ресурсы будут развернуты только после них.

Пример:

```yaml
# .helm/crds/crontab.yaml:
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
# ...
spec:
  names:
    kind: CronTab
```

```
# .helm/templates/crontab.yaml:
apiVersion: example.org/v1
kind: CronTab
# ...
```

```shell
werf converge
```

Результат: сначала развернут CRD для CronTab-ресурса, а затем развернут сам CronTab-ресурс.

## Изменение порядка развертывания ресурсов с помощью группировки по весу (только в werf)

По умолчанию werf объединяет все основные ресурсы (основные — не являющиеся хуками или CRDs из `crds/*.yaml`) в одну группу, создаёт ресурсы этой группы, а затем отслеживает их готовность.

Для изменения порядка развертывания ресурсов можно создать *новые группы ресурсов* через задание ресурсам *веса*, отличного от веса по умолчанию `0`. Все ресурсы с одинаковым весом объединяются в группы, а затем группы ресурсов развертываются по очереди, от группы с меньшим весом к большему, например:

```
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

```shell
werf converge
```

Результат: сначала был развернут ресурс `database`, затем — `database-migrations`, а затем параллельно развернулись `app1` и `app2`.

## Изменение порядка развертывания ресурсов с помощью прямых зависимостей (только в werf)

Аннотация `werf.io/deploy-dependency-<name>` позволяет задать порядок развертывания ресурсов. Такие ресурсы не будут принадлежать ни к одной группе ресурсов по весу или стадии и будут развернуты сразу после того, как их зависимости удовлетворены, например:

```
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

Результат: сначала был развернут ресурс `database`, затем — `database-migrations`, а затем параллельно развернулись `app1` и `app2`.

Все возможности этой аннотации можно узнать [здесь]({{ "/reference/deploy_annotations.html#resource-dependencies" | true_relative_url }}).

Это более гибкий и эффективный способ задания порядка развертывания ресурсов по сравнению с использованием `werf.io/weight` и других методов, т. к. позволяет развертывать ресурсы в порядке графа.

## Запуск задач перед/после установки, обновления, отката или удаления релиза

Для развертывания определенных ресурсов только перед или после установки, обновления, отката или удаления релиза преобразуйте ресурс в *хук* аннотацией `helm.sh/hook`, например:

```
# .helm/templates/example.yaml:
apiVersion: batch/v1
kind: Job
metadata:
  name: database-initialization
  annotations:
    helm.sh/hook: pre-install
# ...
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
```

```shell
werf converge
```

Результат: ресурс `database-initialization` будет развернут только при первой *установке* релиза, а ресурс `myapp` будет развертываться и при установке, и при обновлении, и при откате релиза.

Аннотация `helm.sh/hook` объявляет ресурс-хуком и указывает, при каких условиях этот ресурс должен развертываться (можно указать несколько условий через запятую). Возможные условия для развертывания хука:

* `pre-install` — при установке релиза до установки основных ресурсов;

* `pre-upgrade` — при обновлении релиза до обновления основных ресурсов;

* `pre-rollback` — при откате релиза до отката основных ресурсов;

* `pre-delete` — при удалении релиза до удаления основных ресурсов;

* `post-install` — при установке релиза после установки основных ресурсов;

* `post-upgrade` — при обновлении релиза после обновления основных ресурсов;

* `post-rollback` — при откате релиза после отката основных ресурсов;

* `post-delete` — при удалении релиза после удаления основных ресурсов.

Для задания хукам порядка развертывания присвойте им разные *веса* (по умолчанию — `0`), чтобы хуки развертывались по очереди, от хука с меньшим весом к большему, например:

```
# .helm/templates/example.yaml:
apiVersion: batch/v1
kind: Job
metadata:
  name: first
  annotations:
    helm.sh/hook: pre-install
    helm.sh/hook-weight: "-1"
# ...
---
apiVersion: batch/v1
kind: Job
metadata:
  name: second
  annotations:
    helm.sh/hook: pre-install
# ...
---
apiVersion: batch/v1
kind: Job
metadata:
  name: third
  annotations:
    helm.sh/hook: pre-install
    helm.sh/hook-weight: "1"
# ...
```

```shell
werf converge
```

Результат: сначала будет развернут хук `first`, затем хук `second`, затем хук `third`.

**По умолчанию при повторных развертываниях того же самого хука старый хук в кластере удаляется прямо перед развертыванием нового хука.** Этап удаления старого хука можно изменить аннотацией `helm.sh/hook-delete-policy`, которая принимает следующие значения:

- `hook-succeeded` — удалять новый хук сразу после его удачного развертывания, при неудачном развертывании не удалять совсем;

- `hook-failed` — удалять новый хук сразу после его неудачного развертывания, при удачном развертывании не удалять совсем;

- `before-hook-creation` — (по умолчанию) удалять старый хук сразу перед созданием нового.

## Ожидание готовности ресурсов, не принадлежащих релизу (только в werf)

Развертываемым в текущем релизе ресурсам могут требоваться ресурсы, которые не принадлежат текущему релизу. werf может дожидаться готовности этих внешних ресурсов благодаря аннотации `<name>.external-dependency.werf.io/resource`, например:

```
# .helm/templates/example.yaml:
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  annotations:
    secret.external-dependency.werf.io/resource: secret/my-dynamic-vault-secret
# ...
```

```shell
werf converge
```

Результат: Deployment `myapp` начнёт развертывание только после того, как Secret `my-dynamic-vault-secret`, создаваемый автоматически оператором в кластере, будет создан и готов.

А так можно ожидать готовности сразу нескольких внешних ресурсов:

```
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

По умолчанию werf ищет внешний ресурс в Namespace релиза (если, конечно, ресурс не кластерный). Namespace внешнего ресурса можно изменить аннотацией `<name>.external-dependency.werf.io/namespace`:

```
# .helm/templates/example.yaml:
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  annotations:
    secret.external-dependency.werf.io/resource: secret/my-dynamic-vault-secret
    secret.external-dependency.werf.io/namespace: my-namespace 
```
