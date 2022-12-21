---
title: Сборочный процесс
permalink: usage/build_draft/building.html
---

Запуск сборки в werf осуществляется базовой командой `werf build`.

Команда сборки `werf build` поддерживает работу в локальном режиме без Container Registry, либо в режиме с публикацией образов в Container Registry.

Также сборку могут автоматически осуществлять и другие команды, которые требуют наличия собранных образов в Container Registry, чтобы выполнять свою задачу. Например, выкат приложения через `werf converge`, или запуск произвольной команды в отдельном собранном образе через `werf kube-run`, или рендер манифестов для деплоя в Kubernetes, которые также используют имена собранных образов `werf render` (опционально).

> **ЗАМЕЧАНИЕ:** Парадигма сборки и публикации образов в werf отличается от парадигмы предлагаемой сборщиком docker, в котором есть 2 команды: build и push. В werf есть только команда build, которая собирает образы локально, либо собирает и публикует образы в указанный Container Registry. В секции [послойное кеширование в Container Registry](#послойное-кеширование-в-container-registry) подробно рассмотрен алгоритм работы сборки и чем обусловлено использование такой парадигмы.

## Параллельная сборка

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

Параллельная сборка в werf регулируется двумя параметрами `--parallel=true` и `--parallel-tasks-limit=5`. По умолчанию параллельная сборка включена и собирается не более 5 образов одновременно.

Отметим что при использовании промежуточных стадий Dockerfile палаллельность их сборки также определяется на основе дерева зависимостей стадий между собой. Также если стадия из одного и того же Dockerfile переиспользуется в разных образах объявленных в `werf.yaml`, то werf обеспечит однократную сборку этой общей стадии без лишних пересборок.

## Хранилище

werf собирает образы, состоящие из набора слоёв. Все слои собираемых образов сохраняются в так называемое **хранилище** по мере сборки.

Большинство команд werf требуют указания места размещения хранилища с помощью ключа `--repo` или переменной окружения `WERF_REPO`.

Есть два вида хранилища: локальное и репозиторий (Container Registry).

**Локальное хранилище** стадий будет задействовано, если вызвать команду `werf build` без аргумента `--repo`. Его можно использовать лишь в рамках локальной машины в ограниченном наборе команд (`werf build` и `werf run`). Такой режим применим, например, для локальной разработки без выката в Kubernetes. В качестве локального хранилища выступает docker server или buildah storage ([см. сборочный бэкенд](#сборочный-бэкенд)).

Для выката приложения в Kubernetes werf предполагает использование режима сборки с публикацией образов в Container Registry — т.н. **хранилище в репозитории**. Такой режим включается при использовании команды `werf build --repo REPO`. В указанный репозиторий сохраняются следующие данные:
* итоговые образы вместе с кешом промежуточных стадий;
* служебные метаданные, необходимые для оптимизации сборки, корректной работе алгоритма очистки registry и пр.

> **ЗАМЕЧАНИЕ:** Если запустить сборку с хранилищем стадий в репозитории, то werf сначала проверит наличие требуемых стадий в локальном хранилище и скопирует оттуда подходящие стадии, чтобы не пересобирать эти стадии заново.

Каждый проект должен использовать в качестве хранилища уникальный адрес репозитория, который используется только этим проектом.

Предполагается, что хранилище в репозитории для проекта не будет удалено или очищено сторонними средствами без негативных последствий для пользователей CI/CD построенного на основе werf.

По умолчанию рекомендуется использовать стандартное хранилище в репозитории. В разделе [организация хранилища стадий]({{ "/usage/build_draft/storage.html" | true_relative_url }}) подробно рассмотрены вопросы организации и кастомизации хранилища под нужды проекта.

## Послойное кеширование в Container Registry

В werf реализован алгоритм сборки Dockerfile, который позволяет *уменьшить среднее время сборок* коммитов проекта, а также *обеспечить воспроизводимость собранных образов* для любого коммита проекта (воспроизводимость на любом этапе CI/CD и при откате приложения к произвольному предыдущему коммиту).

В изначальный дизайн алгоритма сборки заложено сохранение сборочного кеша в Container Registry, поэтому из коробки, без специальных настроек обеспечивается:
* Переиспользование сборочного кеша в распределённом окружении (когда сборка осуществляется с разных хостов).
* Атомарное сохранение как промежуточных слоёв так и итоговых образов в Container Registry.
* Гарантия неизменности однажды опубликованных слоёв и итоговых образов в Container Registry.

Алгоритм реализован за счёт отступления от стандартной для docker парадигмы разделения стадий build и push, и переходу к одной стадии build, совмещающей сборку и публикацию слоёв. Стандартный подход сборки и публикации образов и слоёв через docker может выглядеть следующим образом:
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

## Тегирование образов

По умолчанию werf автоматизирует процесс тегирования собираемых образов. Любая команда werf, которая требует образов для выполнения своей работы автоматически рассчитывает теги  образов для текущего git-коммита, проверяет наличие требуемых тегов в хранилище в репозитории, дособирает нехватающие образы если их там не оказалось.

По умолчанию **тег — это некоторый идентификатор в виде хэша**, включающий в себя контрольную сумму инструкций и файлов сборочного контекста. Например:

```
registry.example.org/group/project  d4bf3e71015d1e757a8481536eeabda98f51f1891d68b539cc50753a-1589714365467  7c834f0ff026  20 hours ago  66.7MB
registry.example.org/group/project  e6073b8f03231e122fa3b7d3294ff69a5060c332c4395e7d0b3231e3-1589714362300  2fc39536332d  20 hours ago  66.7MB
```

Такой тег будет меняться только при изменении входных данных для сборки образа. Таким образом один тег может быть переиспользован для разных git-коммитов, werf берёт на себя:
- расчёт таких тегов;
- атомарную публикацию образов по этим тегам в хранилище;
- передачу тегов в Helm-чарт.

Образы в репозитории именуются согласно следующей схемы: `CONTAINER_REGISTRY_REPO:DIGEST-TIMESTAMP_MILLISEC`

- `CONTAINER_REGISTRY_REPO` — репозиторий, заданный опцией `--repo`;
- `DIGEST` — это контрольная сумма от:
    - сборочных инструкций описанных в Dockerfile или werf.yaml;
    - файлов сборочного контекста, используемых в тех или иных сборочных инструкциях.
- `TIMESTAMP_MILLISEC` — временная отметка, которая проставляется в процессе [процедуры сохранения слоя в Container Registry](#послойное-кеширование-в-container-registry) после того как стадия была собрана.

Сборочный алгоритм также дополнительно гарантирует нам, что образ под таким тегом является уникальным и этот тег никогда не будет перезаписан образом с другим содержимым.

### Тегирование промежуточных слоёв

Образы состоят из набора слоёв, каждый слой имеет уникальное имя-идентификатор, а в качестве финального образа используется последний слой .
TODO

### Получение тегов

Чтобы выгрузить теги всех образов для текущего состояния может использоваться опция `--report-path` для команд `werf build`, `werf converge` и пр.:

```shell
# по умолчанию формат json
werf build --report-path=images.json --repo REPO

# поддерживается формат envfile
werf converge --report-path=images.env --report-format=envfile --repo REPO

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

Сервер синхронизации — это сервисный компонент werf, который предназначен для координации нескольких процессов werf и выполняет роль _менеджера блокировок_. Блокировки требуются для корректной публикации новых образов в Container Registry и реализации алгоритма сборки, описанного в разделе [послойное кеширование в Container Registry](#послойное-кеширование-в-container-registry).

Все команды, использующие параметры хранилища (`--repo=...`) также имеют опцию указания адреса менеджера блокировок, который задается опцией `--synchronization=...` или переменной окружения `WERF_SYNCHRONIZATION=...`. Существует 3 возможных опции для указания типа используемой синхронизации:
1. Локальный хост. Включается опцией `--synchronization=:local`. Локальный _менеджер блокировок_ использует файловые блокировки, предоставляемые операционной системой.
2. Kubernetes. Включается опцией `--synchronization=kubernetes://NAMESPACE[:CONTEXT][@(base64:CONFIG_DATA)|CONFIG_PATH]`. _Менеджер блокировок_ в Kubernetes использует ConfigMap по имени проекта `cm/PROJECT_NAME` (тот же самый что и для кеша хранилища) для хранения распределённых блокировок в аннотациях этого ConfigMap. werf использует [библиотеку lockgate](https://github.com/werf/lockgate), которая реализует распределённые блокировки с помощью обновления аннотаций в ресурсах Kubernetes.
3. Http. Включается опцией `--synchronization=http[s]://DOMAIN`.
    - Есть публичный сервер синхронизации доступный по домену `https://synchronization.werf.io`.
    - Собственный http сервер синхронизации может быть запущен командой `werf synchronization`.

werf использует `--synchronization=:local` (локальный _кеш хранилища_ и локальный _менеджер блокировок_) по умолчанию, если используется локальное хранилище (`werf build` без параметра `--repo`).

werf использует `--synchronization=https://synaskjdfchronization.werf.io` по умолчанию, если используется удалённое хранилище (`--repo=CONTAINER_REGISTRY_REPO`).

Пользователь может принудительно указать произвольный адрес компонентов для синхронизации, если это необходимо, с помощью явного указания опции `--synchronization=:local|(kubernetes://NAMESPACE[:CONTEXT][@(base64:CONFIG_DATA)|CONFIG_PATH])|(http[s]://DOMAIN)`.

**ЗАМЕЧАНИЕ:** Множество процессов werf, работающих с одним и тем же проектом обязаны использовать одинаковое хранилище и адрес сервера синхронизации.

### Запуск сервера синхронизации

TODO

## Сборочный бэкенд

werf поддерживает использование Docker Server или Buildah в качестве бекенда для сборки образов. Buildah поддерживает полноценную работу в контейнерах и послойную сборку Dockerfile-ов с кешированием всех промежуточных слоёв в container registry.

TODO

### Buildah

> ПРИМЕЧАНИЕ: На данный момент Buildah доступен только для пользователей Linux и Windows с включённой подсистемой WSL2. Для пользователей MacOS на данный момент предлагается использование виртуальной машины для запуска werf в режиме Buildah.

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

werf поддерживает разные варианты сборки образов при запуске в контейнерах. Рекомендуемым вариантом является использование сборочного бекэнда Buildah, т.к. он не требует внешних зависимостей и прост в настройке. При использовании бекэнда Docker Server будет необходим доступ к серверу Docker и с этим будет связана сложность корректной настройки, а также в этом варианте не будут доступны все возможности werf (см. подробнее далее).


Для использования overlayfs в пользовательском режиме без root:

```shell
docker run \
    --security-opt seccomp=unconfined --security-opt apparmor=unconfined \
    registry.werf.io/werf/werf:latest WERF_COMMAND
```

Если ядро не предоставляет overlayfs для пользовательского режима без root, то следует использовать fuse-overlayfs:

```shell
docker run \
    --device /dev/fuse \
    --security-opt seccomp=unconfined --security-opt apparmor=unconfined \
    registry.werf.io/werf/werf:latest WERF_COMMAND
```

Рекомендуется использовать образ `registry.werf.io/werf/werf:1.2-stable`, также доступны следующие образы:
- `registry.werf.io/werf/werf:latest` — алиас для `registry.werf.io/werf/werf:1.2-stable`;
- `registry.werf.io/werf/werf:1.2-{alpha|beta|ea|stable|rock-solid}` все образы основаны на alpine, выбираем только канал стабильности;
- `registry.werf.io/werf/werf:1.2-{alpha|beta|ea|stable|rock-solid}-{alpine|ubuntu|fedora}` — выбираем канал стабильности и базовый дистрибутив.

При использовании этих образов автоматически будет использоваться бэкенд Buildah. Опций для переключения на использование бекэнда Docker Server не предоставляется.

### Работа через Docker Server (не рекомендуется)

> **ЗАМЕЧАНИЕ:** Для этого метода официальных образов с werf не предоставляется, самостоятельно соберите образ с werf, либо перенастройте образ на базе `registry.werf.io/werf/werf` на использование бэкенда Docker Server вместо Buildah.

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

#### `fuse: device not found`

Полностью данная ошибка может выглядеть следующим образом:

```
error mounting new container: error mounting build container "53f916e7a334a4bb0d9dbc38a0901718d40b99765002bb7f2f2e5464b1db4294": error creating overlay mount to /home/build/.local/share/containers/storage/overlay/49e856f537ba58afdc09137291133994cd1305e40df72c4fab43077cbd405477/merged, mount_data=",lowerdir=/home/build/.local/share/containers/storage/overlay/l/Z5GEVIFIIQ7H262DYUTX3YOVR6:/home/build/.local/share/containers/storage/overlay/l/PJBBW6UNUNGI37IX6R3LDNPX3J:/home/build/.local/share/containers/storage/overlay/l/MUYSUONLQVE4CJMQVDCH2UBAVQ:/home/build/.local/share/containers/storage/overlay/l/67JHKJDCKBTI4R3Q5S5YG44AD3:/home/build/.local/share/containers/storage/overlay/l/3S72G4SWKDXILGANUOCESP5LDK,upperdir=/home/build/.local/share/containers/storage/overlay/49e856f537ba58afdc09137291133994cd1305e40df72c4fab43077cbd405477/diff,workdir=/home/build/.local/share/containers/storage/overlay/49e856f537ba58afdc09137291133994cd1305e40df72c4fab43077cbd405477/work,volatile": using mount program /usr/bin/fuse-overlayfs: fuse: device not found, try 'modprobe fuse' first
fuse-overlayfs: cannot mount: No such file or directory
: exit status 1
time="2021-12-06T11:30:20Z" level=error msg="exit status 1"
```

Решение: включите fuse device для контейнера, в котором запущен werf ([подробности](#ядро-linux-без-поддержки-overlayfs-в-режиме-rootless-и-использование-непривилегированного-контейнера)).

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

Решение: отключите профили AppArmor и seccomp с помощью параметров `--security-opt seccomp=unconfined` и `--security-opt apparmor=unconfined`, добавьте специальные аннотации в Pod ([подробности](#ядро-linux-без-поддержки-overlayfs-в-режиме-rootless-и-использование-непривилегированного-контейнера)).

#### `unshare(CLONE_NEWUSER): Operation not permitted`

Полностью данная ошибка может выглядеть следующим образом:

```
Error during unshare(CLONE_NEWUSER): Operation not permitted
ERRO[0000] error parsing PID "": strconv.Atoi: parsing "": invalid syntax
ERRO[0000] (unable to determine exit status)
```

Решение: отключите профили AppArmor и seccomp с помощью параметров `--security-opt seccomp=unconfined` и `--security-opt apparmor=unconfined`, добавьте специальные аннотации в Pod или используйте привилегированный контейнер ([подробности](#режимы-работы)).

#### `User namespaces are not enabled in /proc/sys/kernel/unprivileged_userns_clone`

Решение:
```bash
# Включить непривилегированные user namespaces:
echo 'kernel.unprivileged_userns_clone = 1' | sudo tee -a /etc/sysctl.conf
sudo sysctl -p
```
