---
title: Отслеживание ресурсов
permalink: usage/deploy/tracking.html
---

## Отслеживание состояния ресурсов

Развертывание ресурсов делится на две стадии: применение ресурсов в кластер и *отслеживание состояния* этих ресурсов. werf реализует продвинутое отслеживание состояния ресурсов (только в werf) благодаря библиотеке [kubedog](https://github.com/werf/kubedog).

Отслеживание ресурсов включено по умолчанию для всех поддерживаемых ресурсов, а именно для:

* всех ресурсов релиза;

* некоторых ресурсов, опосредованно создаваемых ресурсами релиза;

* ресурсов вне релиза, указанных в аннотациях `<name>.external-dependency.werf.io/resource`.

Для ресурсов Deployment, StatefulSet, DaemonSet, Job и Flagger Canary задействуются *специальные* отслеживатели состояния, которые не только точно определяют, удачно или неудачно ресурс был развернут, но и отслеживают состояние дочерних ресурсов, таких как Pod'ы, создаваемые Deployment'ом.

Для остальных ресурсов, не имеющих *специальных* отслеживателей состояния, задействуется *универсальный* отслеживатель, который *предполагает* удачность развертывания ресурса на основании доступной в кластере информации о ресурсе. В редких случаях, если универсальный отслеживатель ошибается в своих предположениях, отслеживание для этого ресурса можно отключить.

### Изменение критериев неудачного развертывания ресурса (только в werf)

По умолчанию werf прерывает развертывание и помечает его как неудачное, если произошло более двух ошибок при развертывании одного из ресурсов.

Изменить максимальное количество ошибок развертывания для ресурса можно аннотацией `werf.io/failures-allowed-per-replica`, например:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  annotations:
    werf.io/failures-allowed-per-replica: "5"
```

Если ресурс имеет аннотацию `werf.io/fail-mode: HopeUntilEndOfDeployProcess`, то ошибки его развертывания будут учитываться только после того, как все остальные ресурсы удачно развернутся.

А помеченный аннотацией `werf.io/track-termination-mode: NonBlocking` ресурс будет отслеживаться только пока все остальные ресурсы не будут развернуты, после чего этот ресурс автоматически посчитается развернутым, даже если это не так.

Универсальный отслеживатель *отсутствие активности* у ресурса в течение 4 минут считает ошибкой развертывания. Изменить этот период можно аннотацией `werf.io/no-activity-timeout`, например:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  annotations:
    werf.io/no-activity-timeout: 10m
```

### Отключение отслеживания состояния и игнорирование ошибок ресурса (только в werf)

Для отключения отслеживания состояния ресурса и игнорирования ошибок его развертывания пометьте ресурс аннотациями `werf.io/fail-mode: IgnoreAndContinueDeployProcess` и `werf.io/track-termination-mode: NonBlocking`, например:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  annotations:
    werf.io/fail-mode: IgnoreAndContinueDeployProcess
    werf.io/track-termination-mode: NonBlocking
```

## Отображение логов контейнеров (только в werf)

Благодаря библиотеке [kubedog](https://github.com/werf/kubedog) werf автоматически отображает логи контейнеров, создаваемых при развертывании Deployment, StatefulSet, DaemonSet и Job.

Выключить отображение логов для ресурса можно аннотацией `werf.io/skip-logs: "true"`, например:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  annotations:
    werf.io/skip-logs: "true"
```

А в аннотации `werf.io/show-logs-only-for-containers` можно явно перечислить контейнеры, логи которых следует отображать, в то же время скрыв логи всех остальных контейнеров, например:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  annotations:
    werf.io/show-logs-only-for-containers: "backend,frontend"
```

... или наоборот — в аннотации `werf.io/skip-logs-for-containers` перечислить контейнеры, логи которых *не* следует отображать, в то же время отображая логи всех остальных контейнеров, например:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  annotations:
    werf.io/skip-logs-for-containers: "sidecar"
```

Для отображения только тех строк лога, которые соответствуют регулярному выражению, используйте аннотацию `werf.io/log-regex`, например:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  annotations:
    werf.io/log-regex: ".*ERROR.*"
```

Возможно отфильтровать строки лога согласно регулярному выражению не для всех контейнеров сразу, а только для определённого контейнера, если использовать аннотацию `werf.io/log-regex-for-<имя контейнера>`, например:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  annotations:
    werf.io/log-regex-for-backend: ".*ERROR.*"
```

## Отображение Events ресурсов (только в werf)

Благодаря библиотеке [kubedog](https://github.com/werf/kubedog) werf может отображать Events отслеживаемых ресурсов, если ресурс имеет аннотацию `werf.io/show-service-messages: "true"`, например:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  annotations:
    werf.io/show-service-messages: "true"
```
