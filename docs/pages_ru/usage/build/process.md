---
title: Сборочный процесс
permalink: usage/build/process.html
keywords: процесс сборки werf, аутентификация в контейнерном реестре, тегирование образов, кэширование сборки, многоплатформенная сборка, кросс-платформенная сборка, ssh агент, секреты сборки, buildah, docker, stapel, dockerfile, версионирование кэша, параллельная сборка, зеркала docker.io, синхронизация образов, синхронизация сборок в k8s, пользовательские теги, сборка для целевой платформы
tags: [сборка, docker, buildah, ssh, кэш, реестр, многоархитектурная, stapel, синхронизация]
---

{% include pages/ru/cr_login.md.liquid %}

## Тегирование образов

<!-- прим. для перевода: на основе https://werf.io/docs/v2/internals/stages_and_storage.html#stage-naming -->

Тегирование образов werf выполняется автоматически в рамках сборочного процесса. Используется оптимальная схема тегирования, основанная на содержимом образа, которая предотвращает лишние пересборки и время простоя приложения при выкате.

<div class="details">
<a href="javascript:void(0)" class="details__summary">Тегирование в деталях</a>
<div class="details__content" markdown="1">

По умолчанию **тег — это некоторый идентификатор в виде хэша**, включающий в себя контрольную сумму инструкций и файлов сборочного контекста. Например:

```
registry.example.org/group/project  d4bf3e71015d1e757a8481536eeabda98f51f1891d68b539cc50753a-1589714365467  7c834f0ff026  20 hours ago  66.7MB
registry.example.org/group/project  e6073b8f03231e122fa3b7d3294ff69a5060c332c4395e7d0b3231e3-1589714362300  2fc39536332d  20 hours ago  66.7MB
```

Такой тег будет меняться только при изменении входных данных для сборки образа. Таким образом, один тег может быть переиспользован для разных Git-коммитов. werf берёт на себя:
- расчёт таких тегов;
- атомарную публикацию образов по этим тегам в репозиторий или локально;
- передачу тегов в Helm-чарт.

Образы в репозитории именуются согласно следующей схемы: `CONTAINER_REGISTRY_REPO:DIGEST-TIMESTAMP_MILLISEC`. Здесь:

- `CONTAINER_REGISTRY_REPO` — репозиторий, заданный опцией `--repo`;
- `DIGEST` — контрольная сумма от:
  - сборочных инструкций, описанных в Dockerfile или `werf.yaml`;
  - файлов сборочного контекста, используемых в тех или иных сборочных инструкциях.
- `TIMESTAMP_MILLISEC` — временная отметка, которая проставляется в процессе [процедуры сохранения слоя в container registry](#послойное-кэширование-образов) после того, как стадия была собрана.

Сборочный алгоритм также дополнительно гарантирует нам, что образ под таким тегом является уникальным, и этот тег никогда не будет перезаписан образом с другим содержимым.

**Тегирование промежуточных слоёв**

На самом деле описанный выше автоматический тег используется и для финальных образов, которые запускает пользователь, и для промежуточных слоёв, хранимых в container registry. Любой слой, найденный в репозитории, может быть использован либо как промежуточный для сборки нового слоя на его основе, либо в качестве финального образа.

</div>
</div>

### Добавление произвольных тегов

Пользователь может добавить произвольное количество дополнительных тегов с опцией `--add-custom-tag`:

```shell
werf build --repo REPO --add-custom-tag main

# Можно добавить несколько тегов-алиасов.
werf build --repo REPO --add-custom-tag main --add-custom-tag latest --add-custom-tag prerelease
```

Шаблон тега может включать следующие параметры:

- `%image%`, `%image_slug%` или `%image_safe_slug%` для использования имени образа из `werf.yaml` (обязательно при сборке нескольких образов);
- `%image_content_based_tag%` для использования content-based тега werf.

```shell
werf build --repo REPO --add-custom-tag "%image%-latest"
```

> **ЗАМЕЧАНИЕ:** При использовании опций создаются **дополнительные теги-алиасы**, ссылающиеся на автоматические теги-хэши. Полное отключение создания автоматических тегов не предусматривается.

## Послойное кэширование образов

Послойное кэширование образов является неотъемлемой частью сборочного процесса werf. werf сохраняет и переиспользует сборочный кэш в container registry, а также синхронизирует работу параллельных сборщиков.

<div class="details">
<a href="javascript:void(0)" class="details__summary">Как работает сборка</a>
<div class="details__content" markdown="1">

Сборка реализована с отступлением от стандартной для Docker парадигмы разделения стадий `build` и `push`, и переходом к одной стадии `build`, совмещающей и сборку, и публикацию слоёв.

Стандартный подход сборки и публикации образов и слоёв через Docker может выглядеть следующим образом:
1. Скачивание сборочного кеша из container registry (опционально).
2. Локальная сборка всех промежуточных слоёв образа с использованием локального кеша слоёв.
3. Публикация собранного образа.
4. Публикация локального сборочного кеша в container registry (опционально).

Алгоритм сборки образов в werf работает по-другому:
1. Если очередной собираемый слой уже есть в container registry, то его сборка и скачивание не происходит.
2. Если очередной собираемый слой отсутствует в container registry, то скачивается предыдущий слой, базовый для сборки текущего.
3. Новый слой собирается на локальной машине и публикуется в container registry.
4. В момент публикации автоматически разрешаются конфликты между сборщиками с разных хостов, которые пытаются опубликовать один и тот же слой. В этом случае гарантированно будет опубликован только один слой, и все остальные сборщики будут обязаны переиспользовать именно его. (Такое возможно за счёт [использования встроенного сервиса синхронизации](#синхронизация-сборщиков)).
5. И т.д. до тех пор, пока не будут собраны все слои образа.

Алгоритм выборки стадии в werf можно представить следующим образом:
1. Рассчитывается [дайджест стадии](#тегирование-образов).
2. Выбираются все стадии, подходящие под дайджест, т.к. с одним дайджестом может быть связанно несколько стадий в репозитории.
3. Для stapel-сборщика, если текущая стадия связана с Git (стадия Git-архив, пользовательская стадия с Git-патчами или стадия `git latest patch`), тогда выбираются только те стадии, которые связаны с коммитами, являющимися предками текущего коммита. Таким образом, коммиты соседних веток будут отброшены.
4. Выбирается старейший по времени `TIMESTAMP_MILLISEC`.

Если запустить сборку с хранением образов в репозитории, то werf сначала проверит наличие требуемых стадий в локальном хранилище и скопирует оттуда подходящие стадии, чтобы не пересобирать эти стадии заново.

</div>
</div>

> **ЗАМЕЧАНИЕ:** Предполагается, что репозиторий образов для проекта не будет удален или очищен сторонними средствами без негативных последствий для пользователей CI/CD, построенного на основе werf ([см. очистка образов]({{ "usage/cleanup/cr_cleanup.html" | true_relative_url }})).

### Dockerfile

По умолчанию Dockerfile-образы кешируются одним образом в container registry. 

Для включения послойного кеширования Dockerfile-инструкций в container registry необходимо использовать директиву `staged` в werf.yaml. Директива может быть установлена глобально на корневом уровне werf.yaml для всех образов или локально для конкретных образов, где локальная настройка переопределит глобальную:

```yaml
# werf.yaml
image: example
dockerfile: ./Dockerfile
staged: true
```

> **ВАЖНО**: Опция `staged: true` поддерживается только при использовании сборщика Buildah.

<div class="details">
<a href="javascript:void(0)" class="details__summary">**ЗАМЕЧАНИЕ**: Послойное кеширование Dockerfile на данный момент находится стадии альфа-тестирования.</a>
<div class="details__content" markdown="1">

Существует несколько ревизий послойного сборщика Dockerfile, которые могут быть включены с помощью переменной `WERF_STAGED_DOCKERFILE_VERSION={v1|v2}`. Изменение версии сборщика может вызвать пересборку сборочного кеша.

* `v1` используется по умолчанию.
* `v2` включает отдельный слой для стадии `FROM`, чтобы закешировать базовый образ, указанный в инструкции `FROM`.

Совместимость версии `v2` может быть сломана в будущих релизах.

</div>
</div>

### Stapel

Образы stapel кешируются в режиме послойного кеширования в container registry по умолчанию без дополнительной конфигурации.

### Версионирование кеша

Вы можете использовать глобальную директиву `build.cacheVersion` или её локальный аналог `<image>.cacheVersion`,
чтобы явно управлять версией кеша образов через конфигурацию и сохранять воспроизводимость всех сборок. Если обе директивы указаны, локальная имеет приоритет.

**Пример использования:**

```yaml
project: test
configVersion: 1
build:
  cacheVersion: global-cache-version
---
image: backend
cacheVersion: user-cache-version
dockerfile: Dockerfile
---
image: frontend
cacheVersion: frontend-cache-version
dockerfile: Dockerfile
staged: true
---
image: user
cacheVersion: user-cache-version
from: alpine:3.14
```

## Параллельность и порядок сборки образов

<!-- прим. для перевода: на основе https://werf.io/docs/v2/internals/build_process.html#parallel-build -->

Все образы, описанные в `werf.yaml`, собираются параллельно на одном сборочном хосте. При наличии зависимостей между образами сборка разбивается на этапы, где каждый этап содержит набор независимых образов и может собираться параллельно.

> При использовании Dockerfile-стадий параллельность их сборки также определяется на основе дерева зависимостей. Также, если Dockerfile-стадия используется разными образами, объявленными в `werf.yaml`, werf обеспечит однократную сборку этой общей стадии без лишних пересборок

Параллельная сборка в werf регулируется двумя параметрами `--parallel` и `--parallel-tasks-limit`. По умолчанию параллельная сборка включена и собирается не более 5 образов одновременно.

Рассмотрим следующий пример:

```Dockerfile
# backend/Dockerfile
FROM node as backend
WORKDIR /app
COPY package*.json /app/
RUN npm ci
COPY . .
CMD ["node", "server.js"]
```

```Dockerfile
# frontend/Dockerfile

FROM ruby as application
WORKDIR /app
COPY Gemfile* /app
RUN bundle install
COPY . .
RUN bundle exec rake assets:precompile
CMD ["rails", "server", "-b", "0.0.0.0"]

FROM nginx as assets
WORKDIR /usr/share/nginx/html
COPY configs/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=application /app/public/assets .
COPY --from=application /app/vendor .
ENTRYPOINT ["nginx", "-g", "daemon off;"]
```

```yaml
image: backend
dockerfile: Dockerfile
context: backend
---
image: frontend
dockerfile: Dockerfile
context: frontend
target: application
---
image: frontend-assets
dockerfile: Dockerfile
context: frontend
target: assets
```

Имеется 3 образа `backend`, `frontend` и `frontend-assets`. Образ `frontend-assets` зависит от `frontend`, потому что он импортирует скомпилированные ассеты из `frontend`.

Формируются следующие наборы для сборки:

```shell
┌ Concurrent builds plan (no more than 5 images at the same time)
│ Set #0:
│ - ⛵ image backend
│ - ⛵ image frontend
│
│ Set #1:
│ - ⛵ frontend-assets
└ Concurrent builds plan (no more than 5 images at the same time)
```

## Сетевая изоляция

werf поддерживает настройку режима сети для сборочных контейнеров. Это позволяет ограничивать доступ к сети во время процесса сборки, что может быть полезно для безопасности или воспроизводимости.

На текущий момент сетевая изоляция поддерживается только при использовании **Docker** в качестве сборочного бэкенда. Она работает как для **Dockerfile**, так и для **Stapel** синтаксисов.

### Конфигурация

Вы можете указать сетевой режим в конфигурации `werf.yaml` для каждого образа с помощью параметра `network`:

```yaml
project: my-project
configVersion: 1
---
image: backend
dockerfile: Dockerfile
network: none # сеть отключена во время сборки
---
image: frontend
from: alpine:3.14
network: host # использовать сеть хоста
```

### Опция CLI

Вы также можете установить сетевой режим глобально для процесса сборки с помощью опции CLI `--backend-network`:

```bash
werf build --backend-network none
```

Опция CLI `--backend-network` имеет **приоритет** над параметром `network`, определенным в конфигурации `werf.yaml`.

## Использование SSH-агента

werf позволяет использовать SSH-агент для аутентификации при доступе к удалённым Git-репозиториям или выполнении команд в сборочных контейнерах.

> **Замечание:** Только пользователь `root` внутри сборочного контейнера может получить доступ к UNIX-сокету из переменной окружения `SSH_AUTH_SOCK`.

По умолчанию werf пытается использовать SSH-агент, запущенный в системе, определяя его через переменную окружения `SSH_AUTH_SOCK`. 

Если SSH-агент не запущен, werf может автоматически запустить временный агент и загрузить доступные ключи (`~/.ssh/id_rsa|id_dsa`). Этот агент работает только в течение выполнения команды и не конфликтует с системным SSH-агентом.

### Указание конкретных SSH-ключей

Флаг `--ssh-key PRIVATE_KEY_FILE_PATH` позволяет ограничить использование SSH-агента конкретными ключами (можно указать несколько раз для добавления нескольких ключей). werf запустит временный SSH-агент только с указанными ключами.

```bash
werf build --ssh-key ~/.ssh/private_key_1 --ssh-key ~/.ssh/private_key_2
```

### Ограничения на macOS

При работе на macOS нужно учитывать, что werf выполняет сборку внутри контейнера, запущенного в Linux-VM Docker Desktop. Docker Desktop предоставляет собственный прокси-сокет, который перенаправляет системный SSH-сокет macOS (обычно тот, что запущен через launchd для текущего пользователя). Использовать произвольный агент возможности нет.

Таким образом, на macOS временный SSH-агент werf и опция `--ssh-key` не работают. Ключи необходимо заранее добавлять в системный SSH-агент macOS.

## Мультиплатформенная и кроссплатформенная сборка

werf позволяет собирать образы как для родной архитектуры хоста, где запущен werf, так и в кроссплатформенном режиме с помощью эмуляции целевой архитектуры, которая может быть отлична от архитектуры хоста. Также werf позволяет собрать образ сразу для множества целевых платформ.

Мультиплатформенная сборка использует механизмы кроссплатформенного исполнения инструкций, предоставляемые [ядром Linux](https://en.wikipedia.org/wiki/Binfmt_misc) и эмулятором QEMU. [Перечень поддерживаемых архитектур](https://www.qemu.org/docs/master/about/emulation.html). Подготовка хост-системы для мультиплатформенной сборки рассмотрена [в разделе установки werf](https://ru.werf.io/getting_started/)

Поддержка мультиплатформенной сборки для разных вариантов синтаксиса сборки, режимов сборки и используемого бэкенда:

|                       | buildah          | docker-server      |
| --------------------- | ---------------- | ------------------ |
| **Dockerfile**        | полная поддержка | полная поддержка   |
| **staged Dockerfile** | полная поддержка | не поддерживается  |
| **stapel**            | полная поддержка | только linux/amd64 |

### Сборка образов под одну целевую платформу

По умолчанию в качестве целевой используется платформа хоста, где запущен werf. Выбор другой целевой платформы для собираемых образов осуществляется с помощью параметра `--platform`:

```shell
werf build --platform linux/arm64
```

— все конечные образы, указанные в werf.yaml, будут собраны для указанной платформы с использованием эмуляции.

Целевую платформу можно также указать директивой конфигурации `build.platform`:

```yaml
# werf.yaml
project: example
configVersion: 1
build:
  platform:
  - linux/arm64
---
image: frontend
dockerfile: frontend/Dockerfile
---
image: backend
dockerfile: backend/Dockerfile
```

В этом случае запуск `werf build` без параметров вызовет сборку образов для указанной платформы (при этом явно указанный параметр `--platform` переопределяет значение из werf.yaml).

### Сборка образов под множество целевых платформ

Поддерживается и сборка образов сразу для набора архитектур. В этом случае в container registry публикуется манифест включающий в себя собранные образы под каждую из указанных платформ (во время скачивания такого образа автоматически будет выбираться образ под требуемую архитектуру).

Можно определить общий список платформ для всех образов в werf.yaml с помощью конфигурации:

```yaml
# werf.yaml
project: example
configVersion: 1
build:
  platform:
  - linux/arm64
  - linux/amd64
  - linux/arm/v7
```

Можно определить список целевых платформ отдельно для каждого собираемого образа (такая настройка будет иметь приоритет над общим списком определённым в werf.yaml):

```yaml
# werf.yaml
project: example
configVersion: 1
---
image: mysql
dockerfile: ./Dockerfile.mysql
platform:
- linux/amd64
---
image: backend
dockerfile: ./Dockerfile.backend
platform:
- linux/amd64
- linux/arm64
```

Общий список можно также переопределить параметром `--platform` непосредственно в момент вызова сборки:

```shell
werf build --platform=linux/amd64,linux/i386
```

— такой параметр переопределяет список целевых платформ указанных в werf.yaml (как общих, так и для отдельных образов).

## Использование зеркал для docker.io

Вы можете использовать зеркала для используемого по умолчанию `docker.io` container registry.

### Docker

Если вы используете Docker для сборки, добавьте `registry-mirrors` в файл `/etc/docker/daemon.json`:

```json
{
  "registry-mirrors": ["https://<my-docker-io-mirror-host>"]
}
```

После этого перезапустите Docker и разлогиньтесь из `docker.io` с помощью команды:

```bash
werf cr logout
```

### Buildah

Если вы используете Buildah для сборки, вместо редактирования `daemon.json` используйте опцию `--container-registry-mirror` для werf-команд. Например:

```shell
werf build --container-registry-mirror=mirror.gcr.io
```

Для добавления нескольких зеркал используйте опцию `--container-registry-mirror` несколько раз.

Помимо опции, поддерживается использование переменных окружения `WERF_CONTAINER_REGISTRY_MIRROR_*`, например:

```bash
export WERF_CONTAINER_REGISTRY_MIRROR_GCR=mirror.gcr.io
export WERF_CONTAINER_REGISTRY_MIRROR_LOCAL=docker.mirror.local
```

## Использование container registry

При использовании werf container registry используется не только для хранения конечных образов, но также для сборочного кэша и служебных данных, необходимых для работы werf (например, метаданные для очистки container registry на основе истории Git). Репозиторий container registry задаётся параметром `--repo`:

```shell
werf converge --repo registry.mycompany.org/project
```

В дополнение к основному репозиторию существует ряд дополнительных:

- `--final-repo` для сохранения конечных образов в отдельном репозитории;
- `--secondary-repo` для использования репозитория в режиме `read-only` (например, для использования container registry CI, в который нельзя пушить, но можно переиспользовать сборочный кэш);
- `--cache-repo` для поднятия репозитория со сборочным кэшом рядом со сборщиками.

> **ВАЖНО.** Для корректной работы werf container registry должен быть надёжным (persistent), а очистка должна выполняться только с помощью специальной команды `werf cleanup`

### Дополнительный репозиторий для конечных образов

При необходимости в дополнение к основному репозиторию могут использоваться т.н. **финальные** репозитории для непосредственного хранения конечных образов.

```shell
werf build --repo registry.mycompany.org/project --final-repo final-registry.mycompany.org/project-final
```

Финальные репозитории позволяют сократить время загрузки образов и снизить нагрузку на сеть за счёт поднятия container registry ближе к кластеру Kubernetes, на котором происходит развёртывание приложения. Также при необходимости финальные репозитории могут использоваться в том же container registry, что и основной репозиторий (`--repo`).

### Дополнительный репозиторий для быстрого доступа к сборочному кэшу

С помощью параметра `--cache-repo` можно указать один или несколько т.н. **кеширующих** репозиториев.

```shell
# Дополнительный кэширующий репозиторий в локальной сети.
werf build --repo registry.mycompany.org/project --cache-repo localhost:5000/project
```

Кеширующий репозиторий может помочь сократить время загрузки сборочного кэша, но для этого скорость загрузки из него должна быть значительно выше по сравнению с основным репозиторием — как правило, это достигается за счёт поднятия container registry в локальной сети, но это необязательно.

При загрузке сборочного кэша кеширующие репозитории имеют больший приоритет, чем основной репозиторий. При использовании кеширующих репозиториев сборочный кэш продолжает сохраняться и в основном репозитории.

Очистка кeширующего репозитория может осуществляться путём его полного удаления без каких-либо рисков.

## Синхронизация сборщиков

<!-- прим. для перевода: на основе https://werf.io/docs/v2/advanced/synchronization.html -->

Для обеспечения согласованности в работе параллельных сборщиков, а также гарантии воспроизводимости образов и промежуточных слоёв, werf берёт на себя ответственность за синхронизацию сборщиков. По умолчанию используется публичный сервис синхронизации по адресу [https://synchronization.werf.io/](https://synchronization.werf.io/) и от пользователя ничего дополнительно не требуется.

<div class="details">
<a href="javascript:void(0)" class="details__summary">Как работает сервис синхронизации</a>
<div class="details__content" markdown="1">

Сервис синхронизации — это компонент werf, который предназначен для координации нескольких процессов werf и выполняет роль _менеджера блокировок_. Блокировки требуются для корректной публикации новых образов в container registry и реализации алгоритма сборки, описанного в разделе [«Послойное кэширование образов»](#послойное-кэширование-образов).

На сервис синхронизации отправляются только обезличенные данные в виде хэш-сумм тегов, публикуемых в container registry.

В качестве сервиса синхронизации может выступать:
1. HTTP-сервер синхронизации, реализованный в команде `werf synchronization`.
2. Ресурс ConfigMap в кластере Kubernetes. В качестве механизма используется библиотека [lockgate](https://github.com/werf/lockgate), реализующая распределённые блокировки через хранение аннотаций в выбранном ресурсе.
3. Локальные файловые блокировки, предоставляемые операционной системой.

</div>
</div>

### Использование собственного сервиса синхронизации

#### HTTP-сервер

Сервер синхронизации можно запустить командой `werf synchronization`, например для использования порта 55581 (по умолчанию):

```shell
werf synchronization --host 0.0.0.0 --port 55581
```

— данный сервер поддерживает только работу в режиме HTTP, для использования HTTPS необходима настройка дополнительной SSL-терминации сторонними средствами (например через Ingress в Kubernetes).

Далее во всех командах werf, которые используют параметр `--repo` дополнительно указывается параметр `--synchronization=http[s]://DOMAIN`, например:

```shell
werf build --repo registry.mydomain.org/repo --synchronization https://synchronization.domain.org
werf converge --repo registry.mydomain.org/repo --synchronization https://synchronization.domain.org
```

#### Специальный ресурс в Kubernetes

Требуется лишь предоставить рабочий кластер Kubernetes, и выбрать namespace, в котором будет хранится сервисный ConfigMap/werf, через аннотации которого будет происходить распределённая блокировка.

Далее во всех командах werf, которые используют параметр `--repo` дополнительно указывается параметр `--synchronization=kubernetes://NAMESPACE[:CONTEXT][@(base64:CONFIG_DATA)|CONFIG_PATH]`, например:

```shell
# Используем стандартный ~/.kube/config или KUBECONFIG.
werf build --repo registry.mydomain.org/repo --synchronization kubernetes://mynamespace
werf converge --repo registry.mydomain.org/repo --synchronization kubernetes://mynamespace

# Явно указываем содержимое kubeconfig через base64.
werf build --repo registry.mydomain.org/repo --synchronization kubernetes://mynamespace@base64:YXBpVmVyc2lvbjogdjEKa2luZDogQ29uZmlnCnByZWZlcmVuY2VzOiB7fQoKY2x1c3RlcnM6Ci0gY2x1c3RlcjoKICBuYW1lOiBkZXZlbG9wbWVudAotIGNsdXN0ZXI6CiAgbmFtZTogc2NyYXRjaAoKdXNlcnM6Ci0gbmFtZTogZGV2ZWxvcGVyCi0gbmFtZTogZXhwZXJpbWVudGVyCgpjb250ZXh0czoKLSBjb250ZXh0OgogIG5hbWU6IGRldi1mcm9udGVuZAotIGNvbnRleHQ6CiAgbmFtZTogZGV2LXN0b3JhZ2UKLSBjb250ZXh0OgogIG5hbWU6IGV4cC1zY3JhdGNoCg==

# Используем контекст mycontext в конфиге /etc/kubeconfig.
werf build --repo registry.mydomain.org/repo --synchronization kubernetes://mynamespace:mycontext@/etc/kubeconfig
```

> **ЗАМЕЧАНИЕ:** Данный способ неудобен при доставке проекта в разные кластера Kubernetes из одного Git-репозитория из-за сложности корректной настройки. В этом случае для всех команд werf требуется указывать один и тот же адрес кластера и ресурс, даже если деплой происходит в разные контура, чтобы обеспечить консистивность данных в container registry. Поэтому для такого случая рекомендуется запустить отдельный общий сервис синхронизации, чтобы исключить вероятность некорректной конфигурации.

#### Локальная синхронизация

Включается опцией `--synchronization=:local`. Локальный _менеджер блокировок_ использует файловые блокировки, предоставляемые операционной системой.

```shell
werf build --repo registry.mydomain.org/repo --synchronization :local
werf converge --repo registry.mydomain.org/repo --synchronization :local
```

> **ЗАМЕЧАНИЕ:** Данный способ подходит лишь в том случае, если в вашей CI/CD системе все запуски werf происходят с одного и того же раннера.


## Отчёт по сборке

Для получения отчёта по сборке можно использовать флаг `--save-build-report` с командами `werf build`, `werf converge` и другими:

```shell
werf build --save-build-report --repo REPO
```

По умолчанию отчёт формируется в файл `.werf-build-report.json` формата `json`, который содержит расширенную информацию о сборке:

* **Images** — список собранных образов:

  * Теги образа (`DockerImageName`, `DockerRepo`, `DockerTag`)
  * Был ли образ пересобран (`Rebuilt`)
  * Является ли образ финальным (`Final`)
  * Размер и образа в байтах (`Size`) и время сборки (`BuildTime`)
  * Стадии сборки (`Stages`) с деталями:
    * Теги (`DockerImageName`, `DockerTag`, `DockerImageID`, `DockerImageDigest`)
    * Размер (`Size`) в байтах
    * Время сборки стадии (`BuildTime`) в секундах
    * Источник базового образа (`SourceType`: `local`, `secondary`, `cache-repo`, `registry`)
    * Был ли загружен базовый образ (`BaseImagePulled`)
    * Была ли стадия пересобрана (`Rebuilt`)

* **ImagesByPlatform** — информация по архитектурам (при multiarch-сборке) анлогично Images

Пример отчёта в формате `json`:

```json
{
  "Images": {
    "frontend": {
      "WerfImageName": "frontend",
      "DockerRepo": "localhost:5000/demo-app",
      "DockerTag": "079dfdd3f51a800c269cdfdd5e4febfcc1676b2c0d533f520255961c-1752501317353",
      "DockerImageID": "sha256:9b3a32dfe5a4aa46d96547e3f8e678626f96741776d78656ea72cab7117612bf",
      "DockerImageDigest": "sha256:54f564edebb6e0699dc0e43de4165488f86fbc76b0c89d88311d7cc06ae397f5",
      "DockerImageName": "localhost:5000/demo-app:079dfdd3f51a800c269cdfdd5e4febfcc1676b2c0d533f520255961c-1752501317353",
      "Rebuilt": false,
      "Final": true,
      "Size": 20960980,
      "BuildTime": "0.00",
      "Stages": [
        {
          "Name": "from",
          "DockerImageName": "localhost:5000/demo-app:6f40fd07cdb62e03d7238e1fccb3341379bcd677ff6d7575317f3783-1752501287209",
          "DockerTag": "6f40fd07cdb62e03d7238e1fccb3341379bcd677ff6d7575317f3783-1752501287209",
          "DockerImageID": "sha256:2ae3fbc31d2b5f0d7d105a74693ed14bec6b106ad43c660b9162dfa00d24d4d0",
          "DockerImageDigest": "sha256:c4449ccfaee03e5b601290909e4d69178c40e39c9d2daf57d1fd74093beb4e10",
          "CreatedAt": 1752501286000000000,
          "Size": 20960798,
          "SourceType": "",
          "BaseImagePulled": false,
          "Rebuilt": false,
          "BuildTime": "0.00"
        },
        {
          "Name": "install",
          "DockerImageName": "localhost:5000/demo-app:079dfdd3f51a800c269cdfdd5e4febfcc1676b2c0d533f520255961c-1752501317353",
          "DockerTag": "079dfdd3f51a800c269cdfdd5e4febfcc1676b2c0d533f520255961c-1752501317353",
          "DockerImageID": "sha256:9b3a32dfe5a4aa46d96547e3f8e678626f96741776d78656ea72cab7117612bf",
          "DockerImageDigest": "sha256:54f564edebb6e0699dc0e43de4165488f86fbc76b0c89d88311d7cc06ae397f5",
          "CreatedAt": 1752501316000000000,
          "Size": 20960980,
          "SourceType": "",
          "BaseImagePulled": false,
          "Rebuilt": false,
          "BuildTime": "0.00"
        }
      ]
    }
  },
  "ImagesByPlatform": {}
}
```

### Получение тегов

С помощью опции `--build-report-path` вы можете указать произвольный путь до отчета, а так же формат, в котором он будет создан: `json` либо `envfile`. Формат `envfile` не содержит подробной инофрмации о сборке и служит для получения тегов собранных образов.

Пример создания отчета в формате `envfile`:

```shell
werf converge --save-build-report --build-report-path .werf-build-report.env --repo REPO
```

Пример вывода:

```shell
WERF_BACKEND_DOCKER_IMAGE_NAME=localhost:5000/demo-app:b94607bcb6e03a6ee07c8dc912739d6ab8ef2efc985227fa82d3de6f-1752510311968
WERF_FRONTEND_DOCKER_IMAGE_NAME=localhost:5000/demo-app:079dfdd3f51a800c269cdfdd5e4febfcc1676b2c0d533f520255961c-1752501317353
```

> **Важно:** Получить теги до сборки невозможно — они формируются в процессе. Чтобы использовать теги, сохраните их после выполнения сборки с помощью `--save-build-report`.


Чтобы извлечь список итоговых тегов образов из отчета в формате `json`, можно использовать утилиту `jq`:

```shell
jq -r '.Images | to_entries | map({key: .key, value: .value.DockerImageName}) | from_entries' .werf-build-report.json
```

Результат будет выглядеть так:

```json
{
  "backend": "localhost:5000/demo-app:caeb9005a06e34f0a20ba51b98d6b99b30f5cf3b8f5af63c8f3ab6c3-1752510176215",
  "frontend": "localhost:5000/demo-app:079dfdd3f51a800c269cdfdd5e4febfcc1676b2c0d533f520255961c-1752501317353"
}
```

