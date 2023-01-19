---
title: Сборочный процесс
permalink: usage/build/process.html
---

Запуск сборки в werf осуществляется базовой командой `werf build`.

<!-- прим. для перевода: на основе https://werf.io/documentation/v1.2/internals/stages_and_storage.html#storage -->

Команда сборки `werf build` поддерживает работу в локальном режиме без Container Registry, если вызвать её без аргумента `--repo`. Такой режим применим, например, для локальной разработки без выката в Kubernetes. В качестве локального хранилища выступает docker server или buildah storage ([см. выбор сборочного бэкенда](#выбор-и-конфигурация-сборочного-бэкенда)).

Для выката приложения в Kubernetes werf предполагает использование режима сборки с публикацией образов в репозитории Container Registry. Такой режим включается при использовании команды `werf build --repo REPO`. В разделе [организация хранения образов в Container Registry]({{ "usage/build/storage.html" | true_relative_url }}) рассмотрены различные схемы хранения и дополнительные связанные опции, которые предоставляет werf.

Также сборку автоматически осуществляют и другие команды, которые требуют наличия собранных образов в Container Registry для выполнения своей задачи. Например, выкат приложения через `werf converge`, или запуск произвольной команды в отдельном собранном образе через `werf kube-run`, или рендер манифестов для деплоя в Kubernetes, которые также используют имена собранных образов `werf render` (опционально).

> **ЗАМЕЧАНИЕ:** Парадигма сборки и публикации образов в werf отличается от парадигмы предлагаемой сборщиком docker, в котором есть 2 команды: build и push. В werf есть только команда build, которая собирает образы локально, либо собирает и публикует образы в указанный Container Registry ([см. хранение сборочного кэша в Container Registry](#сборочный-кэш)).

## Параллельная сборка

<!-- прим. для перевода: на основе https://werf.io/documentation/v1.2/internals/build_process.html#parallel-build -->

По умолчанию все образы описанные в `werf.yaml` будут собираться параллельно на одном сборочном хосте. При этом если между образами есть какие-то зависимости, то werf автоматически их высчитывает и осуществляет сборку образов в правильном порядке. После построение дерева зависимостей образов, werf разбивает сборку на этапы. Каждый этап содержит набор независимых образов, которые могут собираться параллельно.

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

### Конфигурация

Параллельная сборка в werf регулируется двумя параметрами `--parallel=true` и `--parallel-tasks-limit=5`. По умолчанию параллельная сборка включена и собирается не более 5 образов одновременно.

Отметим что при использовании промежуточных стадий Dockerfile палаллельность их сборки также определяется на основе дерева зависимостей стадий между собой. Также если стадия из одного и того же Dockerfile переиспользуется в разных образах объявленных в `werf.yaml`, то werf обеспечит однократную сборку этой общей стадии без лишних пересборок.

## Сборочный кэш

<!-- прим. для перевода: на основе https://werf.io/documentation/v1.2/internals/stages_and_storage.html#storage -->

werf по умолчанию без специальных настроек сохраняет и переиспользует сборочный кэш из Container Registry при использовании параметра `--repo REPO`. Для каждого проекта необходимо указывать уникальный адрес репозитория в Container Registry, который используется только этим проектом.

В указанный репозиторий сохраняются следующие данные:
- итоговые образы вместе с кешом промежуточных стадий;
- служебные метаданные, необходимые для оптимизации сборки, корректной работе алгоритма очистки registry и пр.

<div class="details">
<a href="javascript:void(0)" class="details__summary">Как работает сборка</a>
<div class="details__content" markdown="1">

Сборка реализована с отступлением от стандартной для docker парадигмы разделения стадий build и push, и переходу к одной стадии build, совмещающей и сборку, и публикацию слоёв. 

Стандартный подход сборки и публикации образов и слоёв через docker может выглядеть следующим образом:
1. Скачивание сборочного кеша из Container Registry (опционально).
2. Локальная сборка всех промежуточных слоёв образа с использованием локального кеша слоёв.
3. Публикация собранного образа.
4. Публикация локального сборочного кеша в Container Registry (опционально).

Алгоритм сборки образов в werf работает по-другому:
1. Если очередной собираемый слой уже есть в Container Registry, то его сборка и скачивание не происходит.
2. Если очередной собираемый слой отсутствует в Container Registry, то скачивается предыдущий слой, базовый для сборки текущего.
3. Новый слой собирается на локальной машине и публикуется в Container Registry.
4. В момент публикации автоматически разрешаются конфликты между сборщиками с разных хостов, которые пытаются опубликовать один и тот же слой. В этом случае гарантированно будет опубликован только один слой и все остальные сборщики будут обязаны переиспользовать именно его. (Такое возможно за счёт [использования встроенного сервиса синхронизации](#синхронизация-сборщиков)).
5. И т.д. до тех пор пока не будут собраны все слои образа.

Алгоритм выборки стадии в werf можно представить следующим образом:
1. Рассчитывается [дайджест стадии](#тегирование-образов).
2. Выбираются все стадии, подходящие под дайджест, т.к. с одним дайджестом может быть связанно несколько стадий в репозитории.
3. Для Stapel-сборщика, если текущая стадия связана с git (стадия git-архив, пользовательская стадия с git-патчами или стадия git latest patch), тогда выбираются только те стадии, которые связаны с коммитами, являющимися предками текущего коммита. Таким образом, коммиты соседних веток будут отброшены.
4. Выбирается старейший по времени `TIMESTAMP_MILLISEC`.

Если запустить сборку с хранением образов в репозитории, то werf сначала проверит наличие требуемых стадий в локальном хранилище и скопирует оттуда подходящие стадии, чтобы не пересобирать эти стадии заново.

</div>
</div>

> **ЗАМЕЧАНИЕ:** Предполагается, что репозиторий образов для проекта не будет удален или очищен сторонними средствами без негативных последствий для пользователей CI/CD построенного на основе werf ([см. очистка образов]({{ "usage/cleanup/cr_cleanup.html" | true_relative_url }})).

## Тегирование образов

<!-- прим. для перевода: на основе https://werf.io/documentation/v1.2/internals/stages_and_storage.html#stage-naming -->

По умолчанию werf автоматизирует процесс тегирования собираемых образов. Любая команда werf, которая требует образов для выполнения своей работы автоматически рассчитывает теги  образов для текущего git-коммита, проверяет наличие требуемых тегов образов в репозитории, дособирает нехватающие образы если их там не оказалось.

<div class="details">
<a href="javascript:void(0)" class="details__summary">Как работает автоматическое тегирование</a>
<div class="details__content" markdown="1">

По умолчанию **тег — это некоторый идентификатор в виде хэша**, включающий в себя контрольную сумму инструкций и файлов сборочного контекста. Например:

```
registry.example.org/group/project  d4bf3e71015d1e757a8481536eeabda98f51f1891d68b539cc50753a-1589714365467  7c834f0ff026  20 hours ago  66.7MB
registry.example.org/group/project  e6073b8f03231e122fa3b7d3294ff69a5060c332c4395e7d0b3231e3-1589714362300  2fc39536332d  20 hours ago  66.7MB
```

Такой тег будет меняться только при изменении входных данных для сборки образа. Таким образом один тег может быть переиспользован для разных git-коммитов, werf берёт на себя:
- расчёт таких тегов;
- атомарную публикацию образов по этим тегам в репозиторий или локально;
- передачу тегов в Helm-чарт.

Образы в репозитории именуются согласно следующей схемы: `CONTAINER_REGISTRY_REPO:DIGEST-TIMESTAMP_MILLISEC`

- `CONTAINER_REGISTRY_REPO` — репозиторий, заданный опцией `--repo`;
- `DIGEST` — это контрольная сумма от:
    - сборочных инструкций описанных в Dockerfile или werf.yaml;
    - файлов сборочного контекста, используемых в тех или иных сборочных инструкциях.
- `TIMESTAMP_MILLISEC` — временная отметка, которая проставляется в процессе [процедуры сохранения слоя в Container Registry](#сборочный-кэш) после того как стадия была собрана.

Сборочный алгоритм также дополнительно гарантирует нам, что образ под таким тегом является уникальным и этот тег никогда не будет перезаписан образом с другим содержимым.

**Тегирование промежуточных слоёв**

На самом деле описанный выше автоматический тег используется и для финальных образов, которые запускает пользователь и для промежуточных слоёв, хранимых в Container Registry. Любой слой найденный в репозитории может быть использован либо как промежуточный для сборки нового слоя на его основе, либо в качестве финального образа.

</div>
</div>

### Получение тегов

Чтобы выгрузить теги всех образов для текущего состояния может использоваться опция `--report-path` для команд `werf build`, `werf converge` и пр.:

```shell
# по умолчанию формат json
werf build --report-path=images.json --repo REPO

# поддерживается формат envfile
werf converge --report-path=images.env --report-format=envfile --repo REPO

# в команде рендера финальные теги будут доступны только с параметром --repo
werf render --report-path=images.json --repo REPO
```

> **ЗАМЕЧАНИЕ:** Получить теги заранее, не вызывая сборочный процесс, на данный момент невозможно, можно получить лишь теги уже собранных ранее образов.

Получить теги образов также можно через сервисные значения, передаваемые из сборочного процесса в Helm-чарт, следующей командой:

```shell
werf helm get-autogenerated-values --repo REPO
```

### Кастомизация тегов

Несмотря на использование автоматических тегов по умолчанию, имеется возможность добавить дополнительные теги для собираемых образов и использовать их в процессе выката Helm-чарта. Для этого используются опции `--add-custom-tag` — для добавления произвольного количества тегов-алиасов, и опция `--use-custom-tag` — для выбора определённого тега, который будет использоваться в именах образов в Helm-чарте (опция `--use-custom-tag` также неявно добавляет тег-алиас и включает его использование).

```shell
werf build --repo REPO --add-custom-tag main

# можно добавить несколько тегов-алиасов
werf build --repo REPO --add-custom-tag main --add-custom-tag latest --add-custom-tag prerelease

# можно добавить несколько тегов и использовать тег "main" для образов в Helm-чарте
werf converge --repo REPO --add-custom-tag latest --add-custom-tag prerelease --use-custom-tag main
```

> **ЗАМЕЧАНИЕ:** При использовании опций создаются **дополнительные теги алиасы**, ссылающиеся на автоматические теги-хэши. Полное отключение создания автоматических тегов не предусматривается.

## Синхронизация сборщиков

<!-- прим. для перевода: на основе https://werf.io/documentation/v1.2/advanced/synchronization.html -->

werf по умолчанию без специальной конфигурации осуществляет задачу синхронизации сборщиков для того, чтобы публиковать в Container Registry согласованные и воспроизводимые образы и кеш промежуточных слоёв, а также обеспечить неизменность однажды опубликованных образов по определённому тегу.

Данный механизм по умолчанию использует публичный сервис синхрнизации по адресу [https://synchronization.werf.io/](https://synchronization.werf.io/) для команды сборки `werf build` с параметром `--repo` (при хранении образов в Container Registry). При запуске команды `werf build` без параметра `--repo` данный сервис использоваться не будет — будет активна локальная синхронизация в рамках хоста, где запущен werf.

<div class="details">
<a href="javascript:void(0)" class="details__summary">Как работает сервис синхронизации </a>
<div class="details__content" markdown="1">

Сервис синхронизации — это компонент werf, который предназначен для координации нескольких процессов werf и выполняет роль _менеджера блокировок_. Блокировки требуются для корректной публикации новых образов в Container Registry и реализации алгоритма сборки, описанного в разделе [как работает сборочный кэш](#сборочный-кэш).

На сервис синхронзации отправляются только обезличенные данные в виде хэш-сумм тегов, публикуемых в Container Registry.

В качестве сервиса синхронизации может выступать:
1. Http-сервер синхронизации, реализованный в команде `werf synchronization`.
2. Ресурс ConfigMap в кластере Kubernetes. В качестве механизма используется библиотека  [библиотеку lockgate](https://github.com/werf/lockgate) реализующая распределённые блокировки через хранение аннотаций в выбранном ресурсе.
3. Локальные файловые блокировки, предоставляемые операционной системой.

</div>
</div>

### Использование собственного сервиса синхронизации

Множество процессов werf, работающих с одним и тем же проектом должны использовать одинаковые адреса репозитория в Container Registry и сервера синхронизации. В случае если использование публичного сервиса `https://synchronization.werf.io` не подходит, есть 3 варианта:

#### 1. Запуск http-сервера синхронизации

Сервер синхронизации можно запустить командой `werf synchonization`, например для использования порта 55581 (по умолчанию):

```shell
werf synchronization --host 0.0.0.0 --port 55581
```

— данный сервер поддерживает только работу в режиме http, для использования https необходима настройка дополнительной SSL-терминации сторонними средствами (например через Ingress в Kubernetes).

Далее во всех командах werf, которые используют параметр `--repo` дополнительно указывается параметр `--synchronization=http[s]://DOMAIN`, например:

```shell
werf build --repo registry.mydomain.org/repo --synchronization https://synchronization.domain.org
werf converge --repo registry.mydomain.org/repo --synchronization https://synchronization.domain.org
```

#### 2. Синхронизация через ресурс в Kubernetes

Требуется лишь предоставить рабочий кластер Kubernetes, и выбрать namespace в котором будет хранится сервисный ConfigMap/werf, через аннотации которого будет происходить распределённая блокировка.

Далее во всех командах werf, которые используют параметр `--repo` дополнительно указывается параметр `--synchronization=kubernetes://NAMESPACE[:CONTEXT][@(base64:CONFIG_DATA)|CONFIG_PATH]`, например:

```shell
# используем стандартный ~/.kube/config или KUBECONFIG
werf build --repo registry.mydomain.org/repo --synchronization kubernetes://mynamespace
werf converge --repo registry.mydomain.org/repo --synchronization kubernetes://mynamespace

# явно указываем содержимое kubeconfig через base64
werf build --repo registry.mydomain.org/repo --synchronization kubernetes://mynamespace@base64:YXBpVmVyc2lvbjogdjEKa2luZDogQ29uZmlnCnByZWZlcmVuY2VzOiB7fQoKY2x1c3RlcnM6Ci0gY2x1c3RlcjoKICBuYW1lOiBkZXZlbG9wbWVudAotIGNsdXN0ZXI6CiAgbmFtZTogc2NyYXRjaAoKdXNlcnM6Ci0gbmFtZTogZGV2ZWxvcGVyCi0gbmFtZTogZXhwZXJpbWVudGVyCgpjb250ZXh0czoKLSBjb250ZXh0OgogIG5hbWU6IGRldi1mcm9udGVuZAotIGNvbnRleHQ6CiAgbmFtZTogZGV2LXN0b3JhZ2UKLSBjb250ZXh0OgogIG5hbWU6IGV4cC1zY3JhdGNoCg==

# используем контекст mycontext в конфиге /etc/kubeconfig
werf build --repo registry.mydomain.org/repo --synchronization kubernetes://mynamespace:mycontext@/etc/kubeconfig
```

> **ЗАМЕЧАНИЕ:** Данный способ не удобен при доставке проекта в разные кластера Kubernetes из одного Git-репозитория из-за сложности корректной настройки. В этом случае для всех команд werf требуется указывать один и тот же адрес кластера и ресурс, даже если деплой происходит в разные контура, чтобы обеспечить консистивность данных в Container Registry. Поэтому для такого случае рекомендуется запустить отдельный общий сервис синхронизации, чтобы исключить вероятность некорректной конфигурации.

#### 3. Локальная синхронизация на одном хосте

Включается опцией `--synchronization=:local`. Локальный _менеджер блокировок_ использует файловые блокировки, предоставляемые операционной системой.

```shell
werf build --repo registry.mydomain.org/repo --synchronization :local
werf converge --repo registry.mydomain.org/repo --synchronization :local
```

> **ЗАМЕЧАНИЕ:** Данный способ подходит лишь в том случае, если в вашей CI/CD системе все запуски werf происходят с одного и того же раннера.

## Выбор и конфигурация сборочного бэкенда

<!-- прим. для перевода: на основе https://werf.io/documentation/v1.2/advanced/buildah_mode.html -->

werf поддерживает использование Docker Server или Buildah в качестве бекенда для сборки образов. Buildah обеспечивает полноценную работу в контейнерах и послойную сборку Dockerfile-ов с кешированием всех промежуточных слоёв в container registry. Docker Server на данный момент является бекэндом используемым по умолчанию.

> **ЗАМЕЧАНИЕ:** Сборочный бекэнд Buildah на данный момент не поддерживает хранение образов локально, поддерживается только [хранение образов в Container Registry](#сборочный-кэш).

### Включение и конфигурация Buildah

> **ЗАМЕЧАНИЕ:** На данный момент Buildah доступен только для пользователей Linux и Windows с включённой подсистемой WSL2. Для пользователей MacOS на данный момент предлагается использование виртуальной машины для запуска werf в режиме Buildah.

Для сборки без Docker-сервера werf использует Buildah в rootless-режиме. Buildah встроен в бинарник werf. Требования к хост-системе для запуска werf с бэкендом uildah можно найти в [инструкциях по установке](/installation.html).

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

По умолчанию Buildah режим в werf наследует системные ulimits при запуске сборочных контейнеров. Пользователь может переопределить эти параметры с помощью переменной окружения `WERF_BUILDAH_ULIMIT`.

Формат `WERF_BUILDAH_ULIMIT=type=softlimit[:hardlimit][,type=softlimit[:hardlimit],...]` — конфигурация лимитов, указанные через запятую:
* "core": maximum core dump size (ulimit -c)
* "cpu": maximum CPU time (ulimit -t)
* "data": maximum size of a process's data segment (ulimit -d)
* "fsize": maximum size of new files (ulimit -f)
* "locks": maximum number of file locks (ulimit -x)
* "memlock": maximum amount of locked memory (ulimit -l)
* "msgqueue": maximum amount of data in message queues (ulimit -q)
* "nice": niceness adjustment (nice -n, ulimit -e)
* "nofile": maximum number of open files (ulimit -n)
* "nproc": maximum number of processes (ulimit -u)
* "rss": maximum size of a process's (ulimit -m)
* "rtprio": maximum real-time scheduling priority (ulimit -r)
* "rttime": maximum amount of real-time execution between blocking syscalls
* "sigpending": maximum number of pending signals (ulimit -i)
* "stack": maximum stack size (ulimit -s)

## Работа в контейнерах

### Основной режим

<!-- прим. для перевода: на основе https://werf.io/documentation/v1.2/advanced/ci_cd/run_in_container/how_it_works.html и https://werf.io/documentation/v1.2/advanced/ci_cd/run_in_container/use_docker_container.html -->

Рекомендуемым вариантом запуска werf в контейнерах является использование официальных образов werf из `registry.werf.io/werf/werf`. Запуск werf в контейнере выглядит следующим образом:

```shell
docker run \
    --security-opt seccomp=unconfined --security-opt apparmor=unconfined \
    registry.werf.io/werf/werf:1.2-stable WERF_COMMAND
```

Если ядро Linux не предоставляет overlayfs для пользовательского режима без root, то следует использовать fuse-overlayfs:

```shell
docker run \
    --device /dev/fuse \
    --security-opt seccomp=unconfined --security-opt apparmor=unconfined \
    registry.werf.io/werf/werf:1.2-stable WERF_COMMAND
```

Рекомендуется использовать образ `registry.werf.io/werf/werf:1.2-stable`, также доступны следующие образы:
- `registry.werf.io/werf/werf:latest` — алиас для `registry.werf.io/werf/werf:1.2-stable`;
- `registry.werf.io/werf/werf:1.2-{alpha|beta|ea|stable|rock-solid}` — все образы основаны на alpine, выбираем только канал стабильности;
- `registry.werf.io/werf/werf:1.2-{alpha|beta|ea|stable|rock-solid}-{alpine|ubuntu|fedora}` — выбираем канал стабильности и базовый дистрибутив.

При использовании этих образов автоматически будет использоваться бэкенд Buildah. Опций для переключения на использование бекэнда Docker Server не предоставляется, однако можно собрать собственный образ, в котором 

### Работа через Docker Server (не рекомендуется)

> **ЗАМЕЧАНИЕ:** Для этого метода официальных образов с werf не предоставляется.

Для использования локального docker-сервера необходимо примонтировать docker-сокет в контейнер и использовать привилегированный режим:

```shell
docker run \
    --privileged \
    --volume $HOME/.werf:/root/.werf \
    --volume /tmp:/tmp \
    --volume /var/run/docker.sock:/var/run/docker.sock \
    IMAGE WERF_COMMAND
```

—  этот метод поддерживает сборку Dockerfile-образов или stapel-образов.

Для работы с docker-сервером по tcp можно использовать следующую команду:

```shell
docker run --env DOCKER_HOST="tcp://HOST:PORT" IMAGE WERF_COMMAND
```

— такой метод поддерживает только сборку Dockerfile-образов. Образы stapel не поддерживаются, поскольку сборщик stapel-образов использует монтирование директорий из хост-системы в сборочные контейнеры.

### Устранение проблем

<!-- прим. для перевода: на основе https://werf.io/documentation/v1.2/advanced/ci_cd/run_in_container/how_it_works.html#troubleshooting -->

#### `fuse: device not found`

Полностью данная ошибка может выглядеть следующим образом:

```
error mounting new container: error mounting build container "53f916e7a334a4bb0d9dbc38a0901718d40b99765002bb7f2f2e5464b1db4294": error creating overlay mount to /home/build/.local/share/containers/storage/overlay/49e856f537ba58afdc09137291133994cd1305e40df72c4fab43077cbd405477/merged, mount_data=",lowerdir=/home/build/.local/share/containers/storage/overlay/l/Z5GEVIFIIQ7H262DYUTX3YOVR6:/home/build/.local/share/containers/storage/overlay/l/PJBBW6UNUNGI37IX6R3LDNPX3J:/home/build/.local/share/containers/storage/overlay/l/MUYSUONLQVE4CJMQVDCH2UBAVQ:/home/build/.local/share/containers/storage/overlay/l/67JHKJDCKBTI4R3Q5S5YG44AD3:/home/build/.local/share/containers/storage/overlay/l/3S72G4SWKDXILGANUOCESP5LDK,upperdir=/home/build/.local/share/containers/storage/overlay/49e856f537ba58afdc09137291133994cd1305e40df72c4fab43077cbd405477/diff,workdir=/home/build/.local/share/containers/storage/overlay/49e856f537ba58afdc09137291133994cd1305e40df72c4fab43077cbd405477/work,volatile": using mount program /usr/bin/fuse-overlayfs: fuse: device not found, try 'modprobe fuse' first
fuse-overlayfs: cannot mount: No such file or directory
: exit status 1
time="2021-12-06T11:30:20Z" level=error msg="exit status 1"
```

Решение: включите fuse device для контейнера, в котором запущен werf ([подробности](#работа-в-контейнерах)).

#### `flags: 0x1000: permission denied`

Полностью данная ошибка может выглядеть следующим образом:

```
time="2021-12-06T11:23:23Z" level=debug msg="unable to create kernel-style whiteout: operation not permitted"
time="2021-12-06T11:23:23Z" level=debug msg="[graphdriver] trying provided driver \"overlay\""
time="2021-12-06T11:23:23Z" level=debug msg="overlay: mount_program=/usr/bin/fuse-overlayfs"
Running time 0.01 seconds
Error: unable to get buildah client: unable to create new Buildah instance with mode "native-rootless": unable to get storage: mount /home/build/.local/share/containers/storage/overlay:/home/build/.local/share/containers/storage/overlay, flags: 0x1000: permission denied
time="2021-12-06T11:23:23Z" level=error msg="exit status 1"
```

Решение: отключите профили AppArmor и seccomp с помощью параметров `--security-opt seccomp=unconfined` и `--security-opt apparmor=unconfined`, добавьте специальные аннотации в Pod ([подробности](#работа-в-контейнерах)).

#### `unshare(CLONE_NEWUSER): Operation not permitted`

Полностью данная ошибка может выглядеть следующим образом:

```
Error during unshare(CLONE_NEWUSER): Operation not permitted
ERRO[0000] error parsing PID "": strconv.Atoi: parsing "": invalid syntax
ERRO[0000] (unable to determine exit status)
```

Решение: отключите профили AppArmor и seccomp с помощью параметров `--security-opt seccomp=unconfined` и `--security-opt apparmor=unconfined`, добавьте специальные аннотации в Pod или используйте привилегированный контейнер ([подробности](#работа-в-контейнерах)).

#### `User namespaces are not enabled in /proc/sys/kernel/unprivileged_userns_clone`

Решение:
```bash
# Включить непривилегированные user namespaces:
echo 'kernel.unprivileged_userns_clone = 1' | sudo tee -a /etc/sysctl.conf
sudo sysctl -p
```
