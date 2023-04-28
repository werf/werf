---
title: Сборочный процесс
permalink: usage/build/process.html
---

## Тегирование образов

<!-- прим. для перевода: на основе https://werf.io/documentation/v1.2/internals/stages_and_storage.html#stage-naming -->

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

### Получение тегов

Для получения тегов образов может использоваться опция `--save-build-report` для команд `werf build`, `werf converge` и пр.:

```shell
# По умолчанию формат JSON.
werf build --save-build-report --repo REPO

# Поддерживается формат envfile.
werf converge --save-build-report --build-report-path .werf-build-report.env --repo REPO

# В команде рендера финальные теги будут доступны только с параметром --repo.
werf render --save-build-report --repo REPO
```

> **ЗАМЕЧАНИЕ:** Получить теги заранее, не вызывая сборочный процесс, на данный момент невозможно, можно получить лишь теги уже собранных ранее образов.

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

<!-- прим. для перевода: на основе https://werf.io/documentation/v1.2/internals/stages_and_storage.html#storage -->

Послойное кэширование образов является частью сборочного процесса werf и не требует какой-либо конфигурации. werf сохраняет и переиспользует сборочный кэш в container registry, а также синхронизирует работу параллельных сборщиков.

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

## Параллельность и порядок сборки образов

<!-- прим. для перевода: на основе https://werf.io/documentation/v1.2/internals/build_process.html#parallel-build -->

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

<!-- прим. для перевода: на основе https://werf.io/documentation/v1.2/advanced/synchronization.html -->

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

## Настройка сборочного бэкенда

<!-- прим. для перевода: на основе https://werf.io/documentation/v1.2/advanced/buildah_mode.html -->

werf поддерживает использование Docker-сервер или Buildah в качестве бекенда для сборки образов. Buildah обеспечивает полноценную работу в контейнерах и послойную сборку Dockerfile'ов с кешированием всех промежуточных слоёв в container registry. Docker-сервер на данный момент является бекэндом, используемым по умолчанию.

> **ЗАМЕЧАНИЕ:** Сборочный бекэнд Buildah на данный момент не поддерживает хранение образов локально, поддерживается только [хранение образов в container registry](#послойное-кэширование-образов).

### Docker

Docker-сервер на данный момент является бекэндом, используемым по умолчанию, и никакая дополнительная настройка не требуется.

### Buildah

> **ЗАМЕЧАНИЕ:** На данный момент Buildah доступен только для пользователей Linux и Windows с включённой подсистемой WSL2. Для пользователей macOS предлагается использование виртуальной машины для запуска werf в режиме Buildah.

Для сборки без Docker-сервера werf использует Buildah в rootless-режиме. Buildah встроен в бинарник werf. Требования к хост-системе для запуска werf с бэкендом Buildah можно найти [в инструкциях по установке](/installation.html).

Buildah включается установкой переменной окружения `WERF_BUILDAH_MODE` в один из вариантов: `auto`, `native-chroot`, `native-rootless`. Большинству пользователей для включения режима Buildah достаточно установить `WERF_BUILDAH_MODE=auto`.

#### Уровень изоляции сборки

По умолчанию в режиме `auto` werf автоматически выбирает уровень изоляции в зависимости от платформы и окружения.

Поддерживается 2 варианта изоляции сборочных контейнеров:
1. Chroot — вариант без использования container runtime, включается переменной окружения `WERF_BUILDAH_MODE=native-chroot`.
2. Rootless — вариант с использованием container runtime (в системе должен быть установлен crun, или runc, или kata, или runsc), включается переменной окружения `WERF_BUILAH_MODE=native-rootless`.

#### Драйвер хранилища

werf может использовать драйвер хранилища `overlay` или `vfs`:

* `overlay` позволяет использовать файловую систему OverlayFS. Можно использовать либо встроенную в ядро Linux поддержку OverlayFS (если она доступна), либо реализацию fuse-overlayfs. Это рекомендуемый выбор по умолчанию.
* `vfs` обеспечивает доступ к виртуальной файловой системе вместо OverlayFS. Эта файловая система уступает по производительности и требует привилегированного контейнера, поэтому ее не рекомендуется использовать. Однако в некоторых случаях она может пригодиться.

Как правило, достаточно использовать драйвер по умолчанию (`overlay`). Драйвер хранилища можно задать с помощью переменной окружения `WERF_BUILDAH_STORAGE_DRIVER`.

#### Ulimits

По умолчанию Buildah-режим в werf наследует системные ulimits при запуске сборочных контейнеров. Пользователь может переопределить эти параметры с помощью переменной окружения `WERF_BUILDAH_ULIMIT`.

Формат `WERF_BUILDAH_ULIMIT=type=softlimit[:hardlimit][,type=softlimit[:hardlimit],...]` — конфигурация лимитов, указанные через запятую:
* "core": maximum core dump size (ulimit -c);
* "cpu": maximum CPU time (ulimit -t);
* "data": maximum size of a process's data segment (ulimit -d);
* "fsize": maximum size of new files (ulimit -f);
* "locks": maximum number of file locks (ulimit -x);
* "memlock": maximum amount of locked memory (ulimit -l);
* "msgqueue": maximum amount of data in message queues (ulimit -q);
* "nice": niceness adjustment (nice -n, ulimit -e);
* "nofile": maximum number of open files (ulimit -n);
* "nproc": maximum number of processes (ulimit -u);
* "rss": maximum size of a process's (ulimit -m);
* "rtprio": maximum real-time scheduling priority (ulimit -r);
* "rttime": maximum amount of real-time execution between blocking syscalls;
* "sigpending": maximum number of pending signals (ulimit -i);
* "stack": maximum stack size (ulimit -s).

### Мультиплатформенная сборка

Мультиплатформенная сборка использует механизмы кроссплатформенного исполнения инструкций, предоставляемые [ядром Linux](https://en.wikipedia.org/wiki/Binfmt_misc) и эмулятором QEMU. Самый простой способ зарегистрировать в хост-системе эмуляторы для большинства архитектур — воспользоваться образом `qemu-user-static`:

```shell
docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
```

[Перечень поддерживаемых архитектур](https://www.qemu.org/docs/master/about/emulation.html).

Поддержка мультиплатформенной сборки для разных вариантов синтаксиса сборки, режимов сборки и используемого бекенда:

|                         | buildah            | docker-server        |
|---                      |---                 |---                   |
| **Dockerfile**          | полная поддержка   | полная поддержка     |
| **staged Dockerfile**   | полная поддержка   | не поддерживается    |
| **stapel**              | полная поддержка   | только linux/amd64   |
