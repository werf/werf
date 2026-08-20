---
title: Жизненный цикл ресурса
permalink: usage/deploy/resource_lifecycle.html
---

## Развертывание ресурса по условию

Чтобы развернуть ресурс только для конкретного типа развертывания (install, upgrade, rollback, uninstall) или на конкретной стадии развертывания (pre, main, post), используйте аннотацию `werf.io/deploy-on`, вдохновлённую аннотацией `helm.sh/hook`.

Доступные значения для `werf.io/deploy-on`:
* `pre-install`, `install`, `post-install` — рендерить ресурс только при установке релиза
* `pre-upgrade`, `upgrade`, `post-upgrade` — рендерить ресурс только при обновлении релиза
* `pre-rollback`, `rollback`, `post-rollback` — рендерить ресурс только при откате релиза
* `pre-delete`, `delete`, `post-delete` — рендерить ресурс только при удалении релиза

По умолчанию для обычных ресурсов используется значение `install,upgrade,rollback`, а для хуков значение берётся из `helm.sh/hook`.

Пример:

```yaml
# .helm/templates/example.yaml:
apiVersion: batch/v1
kind: Job
metadata:
  name: database-initialization
  annotations:
    werf.io/deploy-on: pre-install
    werf.io/delete-policy: before-creation
# ...
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
# ...
```

В этом примере ресурс `database-initialization` будет задеплоен при первичной установке релиза и до `myapp`, а ресурс `myapp` будет развернут в основной стадии выката всякий раз, когда релиз устанавливается, обновляется или откатывается.

## Владелец ресурса

Аннотация `werf.io/ownership` определяет, как ресурс удаляется и как работают релизные аннотации. Допустимые значения:
 * `release`: ресурс удаляется, если он удалён из чарта или при удалении релиза, а релизные аннотации ресурса применяются/проверяются во время выката.
 * `anyone`: обратное от `release` — ресурс никогда не удаляется при удалении релиза или при его удалении из чарта, а релизные аннотации не применяются/не проверяются во время выката.

Обычные ресурсы по умолчанию имеют владельцем `release`, а хуки и CRD из директории `crds` имеют владельцем `anyone`.

Пример:

```yaml
# .helm/templates/example.yaml:
apiVersion: batch/v1
kind: Job
metadata:
  name: database-migrations
  annotations:
    werf.io/ownership: "anyone"
# ...
```

Здесь владелец `anyone` делает эту Job похожей на Helm-хук: она не будет удалена при удалении релиза или при удалении из чарта, а её релизные аннотации не будут применяться/проверяться во время выката.

## Политики операций с ресурсом

### werf.io/delete-policy

Аннотация `werf.io/delete-policy` управляет удалениями ресурса во время его развертывания и вдохновлена аннотацией `helm.sh/hook-delete-policy`. Допустимые значения:
* `before-creation`: ресурс всегда пересоздаётся
* `before-creation-if-immutable`: ресурс пересоздается только если мы получили ошибку `field is immutable`, пытаясь обновить ресурс
* `succeeded`: ресурс удаляется после успешной проверки готовности
* `failed`: ресурс удаляется, если проверка готовности завершилась неудачно

Можно указать несколько значений одновременно. По умолчанию у обычных ресурсов политика удаления отсутствует, а у хуков значения берутся из `helm.sh/hook-delete-policy` и транслируются в `werf.io/delete-policy`.

Пример:

```yaml
# .helm/templates/example.yaml:
apiVersion: batch/v1
kind: Job
metadata:
  name: database-migrations
  annotations:
    werf.io/delete-policy: before-creation,succeeded
# ...
```

Здесь Job `database-migrations` всегда пересоздаётся, а затем удаляется после достижения готовности.

### werf.io/delete-propagation

Аннотация `werf.io/delete-propagation` контролирует удаление дочерних ресурсов. Допустимые значения:
* `Foreground`: удалить ресурс после удаления всех его дочерних ресурсов.
* `Background`: немедленно удалить ресурс и удалить все его дочерние ресурсы в фоновом режиме.
* `Orphan`: удалить ресурс, но не удалять его дочерние ресурсы.

Пример:

```yaml
# .helm/templates/example.yaml:
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  annotations:
    werf.io/delete-propagation: Background
# ...
```

Здесь, когда Deployment `myapp` удаляется, его зависимые ресурсы (ReplicaSets, Pods и т. д.) будут удалены в фоновом режиме.

По умолчанию ресурсы удаляются в соответствие с политикой `Foreground`.

### werf.io/resource-policy

Аннотация `werf.io/resource-policy` ограничивает, какие операции werf разрешено выполнять с ресурсом при выкате. Допустимые значения:
* `skip-create`: никогда не создавать ресурс, выкатывать его только если он уже есть в кластере
* `skip-update`: никогда не обновлять ресурс после его создания
* `skip-recreate`: никогда не пересоздавать ресурс — он остаётся как есть, даже если изменилось immutable-поле
* `skip-delete`: никогда не удалять ресурс — ни при удалении из чарта, ни при удалении релиза
* `keep`: псевдоним для `skip-delete`, совместимый с `helm.sh/resource-policy: keep`

Можно указать несколько значений одновременно. Эти политики имеют приоритет над всем, что могло бы пересоздать или удалить ресурс, включая `werf.io/delete-policy`. В отличие от остальных значений, `skip-delete` учитывается и в том случае, если аннотация задана на ресурсе в кластере, а не только в чарте.

По умолчанию политика ресурса не задана, кроме Namespace релиза — он всегда защищён от удаления.

Пример:

```yaml
# .helm/templates/example.yaml:
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: my-pvc
  annotations:
    werf.io/resource-policy: skip-update,skip-delete
# ...
```

Здесь PersistentVolumeClaim `my-pvc` будет создан, если он отсутствует, но никогда не будет обновлён или удалён.

### helm.sh/resource-policy

Поддерживается для совместимости с Helm. Предпочтительна аннотация `werf.io/resource-policy`: она покрывает тот же случай и больше, а если заданы обе, то `helm.sh/resource-policy` игнорируется полностью.

Аннотация `helm.sh/resource-policy: keep` запрещает любое удаление ресурса. Ресурс не может быть удалён ни по какой причине, если присутствует эта аннотация. Эта аннотация также учитывается на ресурсе в кластере, даже если её нет в чарте.

Пример:

```yaml
# .helm/templates/example.yaml:
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: my-pvc
  annotations:
    helm.sh/resource-policy: keep
# ...
```

Здесь PersistentVolumeClaim `my-pvc` никогда не будет удалён ни по какой причине.
