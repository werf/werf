---
title: Аннотации для деплоя
permalink: reference/deploy_annotations.html
toc: false
---

Данная статья содержит описание аннотаций, которые меняют поведение механизма отслеживания ресурсов в процессе выката с помощью werf. Все аннотации должны быть объявлены в шаблонах чарта.

 - [`werf.io/weight`](#resource-weight) — задает вес ресурса, который определяет порядок развертывания ресурсов.
 - [`werf.io/deploy-dependency-ANY_NAME`](#resource-dependencies) — задать зависимость от другого ресурса, что повлияет на порядок развертывания ресурсов.
 - [`<any-name>.external-dependency.werf.io/resource`](#external-dependency-resource) — дождаться, пока указанная внешняя зависимость будет запущена, и только после этого приступить к развертыванию аннотированного ресурса.
 - [`<any-name>.external-dependency.werf.io/namespace`](#external-dependency-namespace) — задать пространство имен для внешней зависимости.
 - [`werf.io/ownership`](#resource-ownership) — определяет, как обрабатывается удаление ресурса и управление аннотациями релиза.
 - [`werf.io/deploy-on`](#conditional-resource-deployment) — определяет, когда рендерить ресурс для выката и на каких стадиях он должен быть задеплоен.
 - [`werf.io/delete-policy`](#resource-delete-policy) — управляет удалением ресурса во время его выката.
 - [`werf.io/replicas-on-creation`](#replicas-on-creation) — задаёт количество реплик, которое должно быть установлено при первичном создании ресурса (полезно при использовании HPA).
 - [`werf.io/track-termination-mode`](#track-termination-mode) — определяет условие при котором werf остановит отслеживание ресурса.
 - [`werf.io/fail-mode`](#fail-mode) — определяет как werf обработает ресурс в состоянии ошибки. Ресурс в свою очередь перейдет в состояние ошибки после превышения порога допустимых ошибок, обнаруженных при отслеживании этого ресурса в процессе выката.
 - [`werf.io/failures-allowed-per-replica`](#failures-allowed-per-replica) — определяет порог ошибок, обнаруживаемых при отслеживании этого ресурса в процессе выката, после превышения которого ресурс перейдет в состояние ошибки. werf обработает это состояние в соответствии с настройкой [fail mode](#fail-mode).
 - [`werf.io/ignore-readiness-probe-fails-for-CONTAINER_NAME`](#ignore-readiness-probe-failures-for-container) — переопределить высчитываемый автоматически период, в течение которого неуспешные readiness-пробы будут игнорироваться и не будут переводить ресурс в состояние ошибки.
 - [`werf.io/no-activity-timeout`](#no-activity-timeout) — переопределить период неактивности, по истечении которого ресурс перейдет в состояние ошибки.
 - [`werf.io/log-regex`](#log-regex) — показывать в логах только те строки вывода ресурса, которые подходят под указанный шаблон.
 - [`werf.io/log-regex-for-CONTAINER_NAME`](#log-regex-for-container) — показывать в логах только те строки вывода для указанного контейнера, которые подходят под указанный шаблон.
 - [`werf.io/log-regex-skip`](#log-regex-skip) — не показывать строки логов, которые подходят под указанный шаблон.
 - [`werf.io/skip-logs`](#skip-logs) — выключить логирование вывода для ресурса.
 - [`werf.io/skip-logs-for-containers`](#skip-logs-for-containers) — выключить логирование вывода для указанного контейнера.
 - [`werf.io/show-logs-only-for-number-of-replicas`](#show-logs-only-for-number-of-replicas) — включить логирование вывода только для указанного числа реплик ресурса.
 - [`werf.io/show-logs-only-for-containers`](#show-logs-only-for-containers) — включить логирование вывода только для указанных контейнеров ресурса.
 - [`werf.io/show-service-messages`](#show-service-messages) — включить вывод сервисных сообщений и событий Kubernetes для данного ресурса.
 - [`werf.io/sensitive`](#mark-resource-as-sensitive) — пометить ресурс как содержащий чувствительные данные, чтобы werf не показывал диффы для этого ресурса в `werf plan`.
 - [`werf.io/sensitive-paths`](#mark-fields-of-a-resource-as-sensitive) — пометить поля ресурса как чувствительные, чтобы werf не показывал диффы для этих полей в `werf plan`.

Больше информации о том, что такое чарт, шаблоны и пр. доступно в [главе про Helm]({{ "usage/deploy/overview.html" | true_relative_url }}).

## Resource weight

`werf.io/weight: "NUM"`

Пример: \
`werf.io/weight: "10"` \
`werf.io/weight: "-10"`

Может быть положительным числом, отрицательным числом или нулем. Значение передается в виде строки. По умолчанию `weight` имеет значение 0. Работает только для ресурсов, не относящихся к хукам. Для хуков используйте `helm.sh/hook-weight`, логика работы которого почти такая же.

Этот параметр задает вес ресурсов, определяя порядок их развертывания. Сначала werf группирует ресурсы в соответствии с их весом, а затем последовательно развертывает их, начиная с группы с наименьшим весом. В этом случае werf не будет приступать к развертыванию следующей группы ресурсов, пока развертывание предыдущей не завершено успешно.

Дополнительная информация доступна в разделе [Порядок развертывания]({{ "/usage/deploy/deployment_order.html" | true_relative_url }}).

## Resource dependencies

`werf.io/deploy-dependency-ANY_NAME: state=STATE[,name=NAME][,namespace=NAMESPACE][,kind=KIND][,group=GROUP][,version=VERSION]`

Пример: \
`werf.io/deploy-dependency-db: state=ready,kind=StatefulSet,name=postgres` \
`werf.io/deploy-dependency-app: state=present,kind=Deployment,group=apps,version=v1,name=app,namespace=app`

Обязательные параметры:
- `state`: `ready` или `present`. Если `present`, то дождаться, пока ресурс не будет создан/обновлен, если `ready`, то дождаться, пока ресурс не будет создан/обновлен и приведен в готовность.

Как минимум один из этих параметров требуется указать:
- `name`: имя ресурса, от которого будет зависеть текущий ресурс.
- `namespace`: namespace ресурса, от которого будет зависеть текущий ресурс.
- `kind`: kind ресурса, от которого будет зависеть текущий ресурс.
- `group`: api group ресурса, от которого будет зависеть текущий ресурс.
- `version`: api version ресурса, от которого будет зависеть текущий ресурс.

Больше информации: [порядок развертывания]({{ "/usage/deploy/deployment_order.html" | true_relative_url }})

## External dependency resource

`<any-name>.external-dependency.werf.io/resource: type[.version.group]/name`

Пример: \
`secret.external-dependency.werf.io/resource: secret/config` \
`someapp.external-dependency.werf.io/resource: deployments.v1.apps/app`

Задает внешнюю зависимость для ресурса. Ресурс с аннотацией будет развернут только после создания и готовности внешней зависимости.

## External dependency namespace

`<any-name>.external-dependency.werf.io/namespace: name`

Указывает пространство имен для внешней зависимости, заданной [соответствующей аннотацией](#external-dependency-resource). Префикс `<any-name>` должен быть таким же, как у аннотации, определяющей внешнюю зависимость.

## Resource ownership

`werf.io/ownership: release|anyone`

Определяет, как обрабатывается удаление ресурса и управление аннотациями релиза. Допустимые значения:
 * `release`: ресурс удаляется, если он удалён из чарта или при удалении релиза, а аннотации релиза применяются/проверяются во время выката.
 * `anyone`: поведение, обратное от `release` — ресурс никогда не удаляется при удалении релиза или удалении из чарта, а аннотации релиза не применяются/не проверяются во время выката.

Обычные ресурсы по умолчанию имеют владельцем `release`, а хуки и CRD из директории `crds` — `anyone`.

## Conditional resource deployment

`werf.io/deploy-on: pre-install|install|post-install|pre-upgrade|upgrade|post-upgrade|pre-rollback|rollback|post-rollback|pre-delete|delete|post-delete`

Определяет, когда рендерить ресурс для выката и на каких стадиях он должен быть задеплоен. Возможные значения:
 * `pre-install`, `install`, `post-install` — рендерить только при установке релиза
 * `pre-upgrade`, `upgrade`, `post-upgrade` — рендерить только при обновлении релиза
 * `pre-rollback`, `rollback`, `post-rollback` — рендерить только при откате релиза
 * `pre-delete`, `delete`, `post-delete` — рендерить только при удалении релиза

По умолчанию для обычных ресурсов используется значение `install,upgrade,rollback`, для хуков — значения из `helm.sh/hook`.

## Resource delete policy

`werf.io/delete-policy: before-creation|succeeded|failed`

Аннотация `werf.io/delete-policy` управляет удалением ресурса во время его выката и вдохновлена аннотацией `helm.sh/hook-delete-policy`. Допустимые значения:
* `before-creation`: ресурс всегда пересоздаётся
* `succeeded`: ресурс удаляется после успешной проверки готовности
* `failed`: ресурс удаляется, если проверка готовности завершилась неудачно

По умолчанию для обычных ресурсов политика удаления не задана, а для хуков значения берутся из `helm.sh/hook-delete-policy` и преобразуются в значения в `werf.io/delete-policy`.

## Replicas on creation

Когда для ресурса включён HPA, использование `spec.replicas` может привести к непредсказуемому поведению, потому что каждый раз когда происходит converge для werf chart через CI/CD количество реплик ресурса будет сброшено к статически заданному в шаблонах чарта значению `spec.replicas`, даже если это значение изменил HPA в рантайме.

Одно из рекомендованных решений — совсем удалить `spec.replicas` из шаблонов чарта. Однако если необходимо установить начальное значение реплик при создании ресурса, можно воспользоваться аннотацией `"werf.io/replicas-on-creation"`.

`"werf.io/replicas-on-creation": "NUM"`

Задаёт число реплик, которые должны быть установлены для ресурса при его первичном создании.

**ЗАМЕЧАНИЕ** `"NUM"` должно быть указано строкой (в двойных кавычках), потому что аннотации не поддерживают передачу других типов данных кроме строк, аннотации с другим типом данных будут проигнорированы.

## Track termination mode

`"werf.io/track-termination-mode": WaitUntilResourceReady|NonBlocking`

Определяет условие остановки отслеживания ресурса в процессе деплоя:
 * `WaitUntilResourceReady` (по умолчанию) — весь процесс деплоя будет отслеживать и ожидать готовности ресурса с данной аннотацией. Т.к. данный режим включен по умолчанию, то, по умолчанию, процесс деплоя ждет готовности всех ресурсов.
 * `NonBlocking` — ресурс с данной аннотацией отслеживается только пока есть другие ресурсы, готовности которых ожидает процесс деплоя.

<img src="https://raw.githubusercontent.com/werf/demos/master/deploy/werf-new-track-modes-3.gif" />

**СОВЕТ** Используйте аннотации — `"werf.io/track-termination-mode": NonBlocking` и `"werf.io/fail-mode": IgnoreAndContinueDeployProcess`, когда описываете в релизе объект Job, который должен быть запущен в фоне и не влияет на процесс деплоя.

**СОВЕТ** Используйте аннотацию `"werf.io/track-termination-mode": NonBlocking`, когда описываете в релизе объект StatefulSet с ручной стратегией выката (параметр `OnDelete`) и не хотите блокировать весь процесс деплоя из-за этого объекта, дожидаясь его обновления.

## Fail mode

`"werf.io/fail-mode": FailWholeDeployProcessImmediately|HopeUntilEndOfDeployProcess|IgnoreAndContinueDeployProcess`

Определяет как werf будет обрабатывать ресурс в состоянии ошибки, которое возникает после превышения порога ошибок, возникающих во время отслеживания данного ресурса в процессе деплоя:
 * `FailWholeDeployProcessImmediately` (по умолчанию) — в случае ошибки при деплое ресурса с данной аннотацией, весь процесс деплоя будет завершен с ошибкой.
 * `HopeUntilEndOfDeployProcess` — в случае ошибки при деплое ресурса с данной аннотацией его отслеживание будет продолжаться, пока есть другие ресурсы, готовности которых ожидает процесс деплоя, либо все оставшиеся ресурсы имеют такую-же аннотацию. Если с ошибкой остался только этот ресурс или несколько ресурсов с такой-же аннотацией, то в случае сохранения ошибки весь процесс деплоя завершается с ошибкой.
 * `IgnoreAndContinueDeployProcess` — ошибка при деплое ресурса с данной аннотацией не влияет на весь процесс деплоя.

## Failures allowed per replica

`"werf.io/failures-allowed-per-replica": "NUMBER"`

По умолчанию, при отслеживании статуса ресурса допускается срабатывание ошибки 1 раз, прежде чем весь процесс деплоя считается ошибочным. Этот параметр влияет на поведение настройки [Fail mode](#fail-mode): определяет порог срабатывания, после которого начинает работать режим реакции на ошибки.

## Ignore readiness probe failures for container

`"werf.io/ignore-readiness-probe-fails-for-CONTAINER_NAME": "TIME"`

Эта аннотация позволяет переопределить высчитываемый автоматически период, в течение которого неуспешные readiness-пробы
не станут переводить ресурс в состояние ошибки, т. е. будут проигнорированы. По умолчанию период игнорирования неудачных
readiness-проб автоматически вычисляется на основе конфигурации readiness-пробы. Заметим, что если в конфигурации
readiness-пробы указано `failureThreshold: 1`, тогда первая же неудачная readiness-проба переведет ресурс в состояние
ошибки, независимо от периода игнорирования.

Формат записи значения описан [здесь](https://pkg.go.dev/time#ParseDuration).

Пример:
`"werf.io/ignore-readiness-probe-fails-for-backend": "20s"`

## No activity timeout

`werf.io/no-activity-timeout: "TIME"`

По умолчанию: `4m`

Пример: \
`werf.io/no-activity-timeout: "8m30s"` \
`werf.io/no-activity-timeout: "90s"`

При отсутствии новых событий и обновлений ресурса в течение `TIME` ресурс перейдет в состояние ошибки.

Формат записи значения описан [здесь](https://pkg.go.dev/time#ParseDuration).

## Log regex

`"werf.io/log-regex": RE2_REGEX`

Определяет [Re2 regex](https://github.com/google/re2/wiki/Syntax) шаблон, применяемый ко всем логам всех контейнеров всех подов ресурса с этой аннотацией. werf будет выводить только те строки лога, которые удовлетворяют regex-шаблону. По умолчанию werf выводит все строки лога.

## Log regex for container

`"werf.io/log-regex-for-CONTAINER_NAME": RE2_REGEX`

Определяет [Re2 regex](https://github.com/google/re2/wiki/Syntax) шаблон, применяемый к логам контейнера с именем `CONTAINER_NAME` всех подов с данной аннотацией. werf будет выводить только те строки лога, которые удовлетворяют regex-шаблону. По умолчанию werf выводит все строки лога.

## Log regex skip

`"werf.io/log-regex-skip": RE2_REGEX`

Определяет [Re2 regex](https://github.com/google/re2/wiki/Syntax) шаблон, применяемый ко всем логам всех контейнеров всех подов ресурса с этой аннотацией. werf не будет выводить те строки лога, которые удовлетворяют regex-шаблону. По умолчанию werf выводит все строки лога.

## Skip logs

`"werf.io/skip-logs": "true"|"false"`

Если установлена в `"true"`, то логи всех контейнеров пода с данной аннотацией не выводятся при отслеживании. Отключено по умолчанию.

<img src="https://raw.githubusercontent.com/werf/demos/master/deploy/werf-new-track-modes-2.gif" />

## Skip logs for containers

`"werf.io/skip-logs-for-containers": CONTAINER_NAME1,CONTAINER_NAME2,CONTAINER_NAME3...`

Список (через запятую) контейнеров пода с данной аннотацией, для которых логи не выводятся при отслеживании.

## Show logs only for number of replicas

`"werf.io/show-logs-only-for-number-of-replicas": "NUMBER"`

Отобразить логи только для указанного числа реплик ресурса в процессе трекинга. Мы отображаем логи только для одной реплики по умолчанию, чтобы избежать избыточного вывода логов и оптимизировать утилизацию ресурсов.

## Show logs only for containers

`"werf.io/show-logs-only-for-containers": CONTAINER_NAME1,CONTAINER_NAME2,CONTAINER_NAME3...`

Список (через запятую) контейнеров пода с данной аннотацией, для которых выводятся логи при отслеживании. Для контейнеров, чьи имена отсутствуют в списке, логи не выводятся. По умолчанию выводятся логи для всех контейнеров всех подов ресурса.

## Show service messages

`"werf.io/show-service-messages": "true"|"false"`

Если установлена в `"true"`, то при отслеживании для ресурсов будет выводиться дополнительная отладочная информация, такая как события Kubernetes. По умолчанию, werf выводит такую отладочную информацию только в случае если ошибка ресурса приводит к ошибке всего процесса деплоя.

<img src="https://raw.githubusercontent.com/werf/demos/master/deploy/werf-new-track-modes-1.gif" />

## Mark fields of a resource as sensitive

`"werf.io/sensitive-paths": "JSONPath,JSONPath,..."`

Например: \
`"werf.io/sensitive-paths": "$.spec.template.spec.containers[*].env[*].value,$.data.*"`

Не показывать диффы для ресурсных полей, которые подпадают под указанные JSONPath. Переопределяет значение аннотации `werf.io/sensitive`.

## Mark resource as sensitive

`"werf.io/sensitive": "true"|"false"`

Если установлена в `"true"`, то werf не будет показывать диффы для этого ресурса в `werf plan`. По умолчанию, werf показывает диффы для всех ресурсов, кроме Secret.
