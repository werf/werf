---
title: Build process
permalink: usage/build/process.html
---

## Tagging images

<!-- reference https://werf.io/documentation/v1.2/internals/stages_and_storage.html#stage-naming -->

The tagging of werf images is performed automatically as part of the build process. werf uses an optimal tagging scheme based on the contents of the image, thus preventing unnecessary rebuilds and application wait times during deployment.

<div class="details">
<a href="javascript:void(0)" class="details__summary">Tagging in details</a>
<div class="details__content" markdown="1">

By default, a **tag is some hash-based identifier** which includes the checksum of instructions and build context files. For example:

```
registry.example.org/group/project  d4bf3e71015d1e757a8481536eeabda98f51f1891d68b539cc50753a-1589714365467  7c834f0ff026  20 hours ago  66.7MB
registry.example.org/group/project  e6073b8f03231e122fa3b7d3294ff69a5060c332c4395e7d0b3231e3-1589714362300  2fc39536332d  20 hours ago  66.7MB
```

Such a tag will only change when the underlying data used to build an image changes. This way, one tag can be reused for different Git commits. werf:
- calculates these tags;
- atomically publishes the images based on these tags to the repository or locally;
- passes the tags to the Helm chart.

The images in the repository are named according to the following scheme: `CONTAINER_REGISTRY_REPO:DIGEST-TIMESTAMP_MILLISEC`. Here:

- `CONTAINER_REGISTRY_REPO` — repository defined by the `--repo` option;
- `DIGEST` — the checksum calculated for:
  - build instructions defined in the Dockerfile or `werf.yaml`;
  - build context files used in build instructions.
- `TIMESTAMP_MILLISEC` — timestamp that is added while [saving a layer to the container registry](#layer-by-layer-image-caching) after the stage has been built.

The assembly algorithm also ensures that the image with such a tag is unique and that tag will never be overwritten by an image with different content.

**Tagging intermediate layers**

The automatically generated tag described above is used for both the final images which the user runs and for intermediate layers stored in the container registry. Any layer found in the repository can either be used as an intermediate layer to build a new layer based on it, or as a final image.

</div>
</div>

### Retrieving tags

You can retrieve image tags using the `--report-path` option for `werf build`, `werf converge`, and other commands:

```shell
# By default, the JSON format is used.
werf build --report-path=images.json --repo REPO

# The envfile format is also supported.
werf converge --report-path=images.env --report-format=envfile --repo REPO

# For the render command, the final tags are only available with the --repo parameter.
werf render --report-path=images.json --repo REPO
```

> **NOTE:** Retrieving tags beforehand without first invoking the build process is currently impossible. You can only retrieve tags from the images you've already built.

### Adding custom tags

The user can add any number of custom tags using the `--add-custom-tag` option:

```shell
werf build --repo REPO --add-custom-tag main

# You can add some alias tags.
werf build --repo REPO --add-custom-tag main --add-custom-tag latest --add-custom-tag prerelease
```

The tag template may include the following parameters:

- `%image%`, `%image_slug%` or `%image_safe_slug%` to use an image name defined in `werf.yaml` (mandatory when building multiple images);
- `%image_content_based_tag%` to use the content-based werf tag.

```shell
werf build --repo REPO --add-custom-tag "%image%-latest"
```

> **NOTE:** When you use the options listed above, werf still creates the **additional alias tags** that reference the automatic hash tags. It is not possible to completely disable auto-tagging.

### Deploying images with an arbitrary tag

The user may opt for arbitrary tagging when deploying images in Kubernetes. The `--use-custom-tag` option can be used for this (it implicitly adds an alias tag and enables its use):

```shell
werf converge --repo REPO --use-custom-tag main
```

## Layer-by-layer image caching

<!-- reference https://werf.io/documentation/v1.2/internals/stages_and_storage.html#storage -->

Layer-by-layer image caching is part of the werf build process and does not require any configuration. werf saves and reuses the build cache in the container registry and synchronizes parallel builders.

<div class="details">
<a href="javascript:void(0)" class="details__summary">How assembly works</a>
<div class="details__content" markdown="1">

Building in werf deviates from the standard Docker paradigm of separating the `build` and `push` stages. It consists of a single `build` stage, which combines both building and publishing layers.

A standard approach for building and publishing images and layers via Docker might look like this:
1. Downloading the build cache from the container registry (optional).
2. Local building of all the intermediate image layers using the local layer cache.
3. Publishing the built image.
4. Publishing the local build cache to the container registry (optional).

The image building algorithm in werf is different:
1. If the next layer to be built is already present in the container registry, it will not be built or downloaded.
2. If the next layer to be built is not in the container registry, the previous layer is downloaded (the base layer for building the current one).
3. The new layer is built on the local machine and published to the container registry.
4. At publishing time, werf automatically resolves conflicts between builders from different hosts that try to publish the same layer. This ensures that only one layer is published, and all other builders are required to reuse that layer. ([The built-in sync service](#syncing-builders) makes this possible).
5. The process continues until all the layers of the image are built.

The algorithm of stage selection in werf works as follows:
1. werf calculates the [stage digest](#tagging-images).
2. Then it selects all the stages matching the digest, since several stages in the repository may be tied to a single digest.
3. For the Stapel builder, if the current stage involves Git (a Git archive stage, a custom stage with Git patches, or a `git latest patch` stage), then only those stages associated with commits that are ancestral to the current commit are selected. Thus, commits from neighboring branches will be discarded.
4. Then the oldest `TIMESTAMP_MILLISEC` is selected.

If you run a build with storing images in the repository, werf will first check if the required stages exist in the local repository and copy the suitable stages from there, so that no rebuilding of those stages is necessary.

</div>
</div>

> **ЗАМЕЧАНИЕ:** Предполагается, что репозиторий образов для проекта не будет удален или очищен сторонними средствами без негативных последствий для пользователей CI/CD, построенного на основе werf ([см. очистка образов]({{ "usage/cleanup/cr_cleanup.html" | true_relative_url }})).

## Параллельность и порядок сборки образов

<!-- reference: https://werf.io/documentation/v1.2/internals/build_process.html#parallel-build -->

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

werf поддерживает использование Docker Server или Buildah в качестве бекенда для сборки образов. Buildah обеспечивает полноценную работу в контейнерах и послойную сборку Dockerfile'ов с кешированием всех промежуточных слоёв в container registry. Docker Server на данный момент является бекэндом, используемым по умолчанию.

> **ЗАМЕЧАНИЕ:** Сборочный бекэнд Buildah на данный момент не поддерживает хранение образов локально, поддерживается только [хранение образов в container registry](#послойное-кэширование-образов).

### Docker

Docker Server на данный момент является бекэндом, используемым по умолчанию, и никакая дополнительная настройка не требуется.

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
