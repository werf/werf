---
title: Образы и зависимости
permalink: usage/build/images.html
---

<!-- прим. для перевода: на основе https://werf.io/documentation/v1.2/reference/werf_yaml.html#image-section -->

## Добавление образов

Для сборки c werf необходимо добавить описание образов в `werf.yaml` проекта. Каждый образ добавляется директивой `image` с указанием имени образа:

```yaml
project: example
configVersion: 1
---
image: frontend
# ...
---
image: backend
# ...
---
image: database
# ...
```

> Имя образа — это уникальный внутренний идентификатор образа, который позволяет ссылаться на него при конфигурации и при вызове команд werf.

Далее для каждого образа в `werf.yaml` необходимо определить сборочные инструкции [с помощью Dockerfile](#dockerfile) или [stapel](#stapel).

### Dockerfile

<!-- прим. для перевода: на основе https://werf.io/documentation/v1.2/reference/werf_yaml.html#dockerfile-builder -->

#### Написание Dockerfile-инструкций

Для описания сборочных инструкций образа поддерживается стандартный Dockerfile. Следующие ресурсы помогут в его написании:

* [Dockerfile Reference](https://docs.docker.com/engine/reference/builder/).
* [Best practices for writing Dockerfiles](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/).

#### Использование Dockerfile

Конфигурация сборки Dockerfile может выглядеть следующим образом:

```Dockerfile
# Dockerfile
FROM node
WORKDIR /app
COPY package*.json /app/
RUN npm ci
COPY . .
CMD ["node", "server.js"]
```

```yaml
# werf.yaml
project: example
configVersion: 1
---
image: backend
dockerfile: Dockerfile
```

#### Использование определённой Dockerfile-стадии

Также вы можете описывать несколько целевых образов из разных стадий одного и того же Dockerfile:

```Dockerfile
# Dockerfile

FROM node as backend
WORKDIR /app
COPY package*.json /app/
RUN npm ci
COPY . .
CMD ["node", "server.js"]

FROM python as frontend
WORKDIR /app
COPY requirements.txt /app/
RUN pip install -r requirements.txt
COPY . .
CMD ["gunicorn", "app:app", "-b", "0.0.0.0:80", "--log-file", "-"]
```

```yaml
# werf.yaml
project: example
configVersion: 1
---
image: backend
dockerfile: Dockerfile
target: backend
---
image: frontend
dockerfile: Dockerfile
target: frontend
```

И конечно вы можете описывать образы, основанные на разных Dockerfile:

```yaml
# werf.yaml
project: example
configVersion: 1
---
image: backend
dockerfile: dockerfiles/Dockerfile.backend
---
image: frontend
dockerfile: dockerfiles/Dockerfile.frontend
```

#### Выбор директории сборочного контекста

Чтобы указать сборочный контекст используется директива `context`. **Важно:** в этом случае путь до Dockerfile указывается относительно директории контекста:

```yaml
project: example
configVersion: 1
---
image: docs
context: docs
dockerfile: Dockerfile
---
image: service
context: service
dockerfile: Dockerfile
```

Для образа `docs` будет использоваться Dockerfile по пути `docs/Dockerfile`, а для `service` — `service/Dockerfile`.

#### Добавление произвольных файлов в сборочный контекст

По умолчанию контекст сборки Dockerfile-образа включает только файлы из текущего коммита репозитория проекта. Файлы, не добавленные в Git, или некоммитнутые изменения не попадают в сборочный контекст. Такая логика действует в соответствии [с настройками гитерминизма]({{ "/usage/project_configuration/giterminism.html" | true_relative_url }}) по умолчанию.

Чтобы добавить в сборочный контекст файлы, которые не хранятся в Git, нужна директива `contextAddFiles` в `werf.yaml`, а также нужно разрешить использование директивы `contextAddFiles` в `werf-giterminism.yaml` (подробнее [про гитерминизм]({{ "/usage/project_configuration/giterminism.html#contextaddfiles" | true_relative_url }})):

```yaml
# werf.yaml
project: example
configVersion: 1
---
image: app
context: app
contextAddFiles:
- file1
- dir1/
- dir2/file2.out
```

```yaml
# werf-giterminism.yaml
giterminismConfigVersion: 1
config:
  dockerfile:
    allowContextAddFiles:
    - app/file1
    - app/dir1/
    - app/dir2/file2.out
```

В данной конфигурации контекст сборки будет состоять из следующих файлов:

- `app/**/*` из текущего коммита репозитория проекта;
- файлы `app/file1`, `app/dir2/file2.out` и директория `dir1`, которые находятся в директории проекта.

### Stapel

В werf встроен альтернативный синтаксис описания сборочных инструкций, называемый stapel. Подробная документация по синтаксису stapel доступна [в соответствующей секции документации]({{ "/usage/build/stapel/overview.html" | true_relative_url }}).

Пример минимальной конфигурации stapel-образа в `werf.yaml`:

```yaml
project: example
configVersion: 1
---
image: app
from: ubuntu:22.04
```

Добавим исходники из Git в образ:

```yaml
project: example
configVersion: 1
---
image: app
from: ubuntu:22.04
git:
- add: /
  to: /app
```

Доступно 4 стадии для описания произвольных shell-инструкций, а также директива `git.stageDependencies` для настройки триггеров пересборки этих стадий при изменении соответствующих стадий ([см. подробнее]({{ "/usage/build/stapel/instructions.html#зависимость-от-изменений-в-git-репозитории" | true_relative_url }})):

```yaml
project: example
configVersion: 1
---
image: app
from: ubuntu:22.04
git:
- add: /
  to: /app
  stageDependencies:
    install:
    - package-lock.json
    - Gemfile.lock
    beforeSetup:
    - app/webpack/
    - app/assets/
    setup:
    - config/templates/
shell:
  beforeInstall:
  - apt update -q
  - apt install -y libmysqlclient-dev mysql-client g++
  install:
  - bundle install
  - npm install
  beforeSetup:
  - bundle exec rails assets:precompile
  setup:
  - rake generate:configs
```

Поддерживаются вспомогательные образы, из которых можно импортировать файлы в целевой образ (аналог `COPY --from=STAGE` в multi-stage Dockerfile), а также Golang-шаблонизация:

{% raw %}
```yaml
{{ $_ := set . "BaseImage" "ubuntu:22.04" }}

{{ define "package:build-tools" }}
  - apt update -q
  - apt install -y gcc g++ build-essential make
{{ end }}

project: example
configVersion: 1
---
image: builder
from: {{ .BaseImage }}
shell:
  beforeInstall:
{{ include "package:build-tools" }}
  install:
  - cd /app
  - make build
---
image: app
from: alpine:latest
import:
- image: builder
  add: /app/build/app
  to: /usr/local/bin/app
  after: install
```
{% endraw %}

Подробная документация по написанию доступна [в разделе stapel]({{ "usage/build/stapel/base.html" | true_relative_url }}).

## Взаимодействие между образами

### Наследование и импортирование файлов

При написании одного Dockerfile в нашем распоряжении имеется механизм multi-stage. Он позволяет объявить в Dockerfile отдельный образ-стадию и использовать её в качестве базового для другого образа, либо скопировать из неё отдельные файлы.

werf позволяет реализовать это не только в рамках одного Dockerfile, но и между произвольными образами, определяемыми в `werf.yaml`, в том числе собираемыми из разных Dockerfile'ов, либо собираемыми сборщиком stapel. Всю оркестрацию и выстраивание зависимостей werf возьмёт на себя и произведёт сборку за один шаг (вызов `werf build`).

Пример использования образа собранного из `base.Dockerfile` в качестве базового для образа из `Dockerfile`:

```Dockerfile
# base.Dockerfile
FROM ubuntu:22.04
RUN apt update -q && apt install -y gcc g++ build-essential make curl python3
```

```Dockerfile
# Dockerfile
ARG BASE_IMAGE
FROM ${BASE_IMAGE}
WORKDIR /app
COPY . .
CMD [ "/app/server", "start" ]
```

```yaml
# werf.yaml
project: example
configVersion: 1
---
image: base
dockerfile: base.Dockerfile
---
image: app
dockerfile: Dockerfile
dependencies:
- image: base
  imports:
  - type: ImageName
    targetBuildArg: BASE_IMAGE
```

В следующем примере рассмотрено импортирование файлов из образа stapel в образ Dockerfile:

```yaml
# werf.yaml
project: example
configVersion: 1
---
image: builder
from: golang
git:
- add: /
  to: /app
shell:
  install:
  - cd /app
  - go build -o /app/bin/server
---
image: app
dockerfile: Dockerfile
dependencies:
- image: builder
  imports:
  - type: ImageName
    targetBuildArg: BUILDER_IMAGE
```

```Dockerfile
# Dockerfile
FROM alpine
ARG BUILDER_IMAGE
COPY --from=${BUILDER_IMAGE} /app/bin /app/bin
CMD [ "/app/bin/server", "server" ]
```

### Передача информации о собранном образе в другой образ

werf позволяет получить информацию о собранном образе при сборке другого образа. Например, если в сборочных инструкциях образа `app` требуются имена и digest'ы образов `auth` и `controlplane`, опубликованных в container registry, то конфигурация могла бы выглядеть так:

```Dockerfile
# modules/auth/Dockerfile
FROM alpine
WORKDIR /app
COPY . .
RUN ./build.sh
```

```Dockerfile
# modules/controlplane/Dockerfile
FROM alpine
WORKDIR /app
COPY . .
RUN ./build.sh
```

```Dockerfile
# Dockerfile
FROM alpine
WORKDIR /app
COPY . .

ARG AUTH_IMAGE_NAME
ARG AUTH_IMAGE_DIGEST
ARG CONTROLPLANE_IMAGE_NAME
ARG CONTROLPLANE_IMAGE_DIGEST

RUN echo AUTH_IMAGE_NAME=${AUTH_IMAGE_NAME}                     >> modules_images.env
RUN echo AUTH_IMAGE_DIGEST=${AUTH_IMAGE_DIGEST}                 >> modules_images.env
RUN echo CONTROLPLANE_IMAGE_NAME=${CONTROLPLANE_IMAGE_NAME}     >> modules_images.env
RUN echo CONTROLPLANE_IMAGE_DIGEST=${CONTROLPLANE_IMAGE_DIGEST} >> modules_images.env
```

```yaml
# werf.yaml
project: example
configVersion: 1
---
image: auth
dockerfile: Dockerfile
context: modules/auth/
---
image: controlplane
dockerfile: Dockerfile
context: modules/controlplane/
---
image: app
dockerfile: Dockerfile
dependencies:
- image: auth
  imports:
  - type: ImageName
    targetBuildArg: AUTH_IMAGE_NAME
  - type: ImageDigest
    targetBuildArg: AUTH_IMAGE_DIGEST
- image: controlplane
  imports:
  - type: ImageName
    targetBuildArg: CONTROLPLANE_IMAGE_NAME
  - type: ImageDigest
    targetBuildArg: CONTROLPLANE_IMAGE_DIGEST
```

В процессе сборки werf автоматически подставит в указанные build-arguments соответствующие имена и идентификаторы. Всю оркестрацию и выстраивание зависимостей werf возьмёт на себя и произведёт сборку за один шаг (вызов `werf build`).

## Мультиплатформенная и кроссплатформенная сборка

werf позволяет собирать образы как для родной архитектуры хоста, где запущен werf, так и в кроссплатформенном режиме с помощью эмуляции целевой архитектуры, которая может быть отлична от архитектуры хоста. Также werf позволяет собрать образ сразу для множества целевых платформ.

> **ЗАМЕЧАНИЕ:** Подготовка хост-системы для мультиплатформенной сборки рассмотрена [в разделе установки werf]({{ "index.html" | true_relative_url }}), а поддержка этого режима для различных синтаксисов инструкций и бекендов рассмотрены [в разделе про сборочный процесс]({{ "/usage/build/process.html" | true_relative_url }}).

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
