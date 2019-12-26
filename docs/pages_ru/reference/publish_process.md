---
title: Публикация (publish)
sidebar: documentation
permalink: documentation/reference/publish_process.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

<!--Docker images should be pushed into the docker registry for further usage in most cases. The usage includes these demands:-->

<!--1. Using an image to run an application (for example in Kubernetes). These images will be referred to as **images for running**.-->
<!--2. Using an existing old image version from a Docker registry as a cache to build a new image version. Usually, it is default behavior. However, some additional actions may be required to organize a build environment with multiple build hosts or build hosts with no persistent local storage. These images will be referred to as **distributed images cache**.-->

<!--## What can be published-->

<!--The result of werf [build commands]({{ site.baseurl }}/documentation/cli/build/build.html) is a _stages_ in _stages storage_ related to images defined in the `werf.yaml` config. -->
<!--werf can be used to publish either:-->

<!--* Images. These can only be used as _images for running_. -->
<!--These images are not suitable for _distributed images cache_, because werf build algorithm implies creating separate images for _stages_. -->
<!--When you pull a image from a Docker registry, you do not receive _stages_ for this image.-->
<!--* Images with a stages cache images. These images can be used as _images for running_ and also as a _distributed images cache_.-->

<!--werf pushes image into a Docker registry with a so-called [**image publish procedure**](#image-publish-procedure). Also, werf pushes stages cache of all images from config with a so-called [**stages publish procedure**](#stages-publish-procedure).-->

<!--Before digging into these algorithms, it is helpful to see how to publish images using Docker.-->

<!--### Stages publish procedure-->

<!--To publish stages cache of a image from the config werf implements the **stages publish procedure**. It consists of the following steps:-->

<!-- 1. Create temporary image names aliases for all docker images in stages cache, so that:-->
<!--     - [docker repository name](https://docs.docker.com/glossary/?term=repository) is a `REPO` parameter specified by the user without changes ([details about `REPO`]({{ site.baseurl }}/documentation/reference/registry/image_naming.html#repo-parameter)).-->
<!--     - [docker tag name](https://docs.docker.com/glossary/?term=tag) constructed as a signature prefixed with a word `image-stage-` (for example `image-stage-41772c141b158349804ad27b354247df8984ead077a5dd601f3940536ebe9a11`).-->
<!-- 2. Push images by newly created aliases into Docker registry.-->
<!-- 3. Delete temporary image names aliases.-->

<!--All of these steps are also performed with a single werf command, which will be described below.-->

<!--The result of this procedure is multiple images from stages cache of image pushed into the Docker registry.-->

## Процедура публикации образа

Процесс публикации образа при обычной работе с Docker, состоит из следующих этапов:

```shell
docker tag REPO:TAG
docker push REPO:TAG
docker rmi REPO:TAG
```

 1. Получение имени или id собранного локального образа.
 2. Создание временного образа-псевдонима, состоящего из следующих частей:
     - [имя Docker-репозитория](https://docs.docker.com/glossary/?term=repository) содержащего в том числе и адрес Docker registry;
     - [имя Docker-тега](https://docs.docker.com/glossary/?term=tag).
 3. Публикация образа-псевдонима в Docker registry.
 4. Удаление временного образа-псевдонима.

В werf реализована иная логика публикации [образа]({{ site.baseurl }}/documentation/reference/stages_and_images.html#образы), описанного в конфигурации:

 1. Создание **нового образа** с определенным именем на основе соответствующего собранного образа. В созданном образе хранится информация о применяемой схеме тегирования, необходимая для внутреннего использования werf (эта информация сохраняется с использованием docker labels). Далее, такая информация будет упоминаться как **мета-информация** образа. werf использует мета-информацию образа в [процессе деплоя]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#integration-with-built-images) и [процессе очистки]({{ site.baseurl }}/documentation/reference/cleaning_process.html).
 2. Публикация созданного образа в Docker registry.
 3. Удаление образа, созданного на этапе 1.

Далее, такой процесс будет называться **процессом публикации образа**.

Результатом процесса публикации образа является образ, именованный согласно [*правил именования образов*](#именование-образов) и загруженный в Docker registry. Все эти шаги выполняются с помощью команды [werf publish]({{ site.baseurl }}/documentation/cli/main/publish.html) или [werf build-and-publish]({{ site.baseurl }}/documentation/cli/main/build_and_publish.html).

## Именование образов

Во время процесса публикации образа, werf формирует имя образа используя:
 * значение параметра `--images-repo`;
 * значение параметра `--images-repo-mode`;
 * имя образа из файла конфигурации `werf.yaml`;
 * значение параметра тегирования.

Итоговое имя Docker-образа имеет формат — [`DOCKER_REPOSITORY`](https://docs.docker.com/glossary/?term=repository)`:`[`TAG`](https://docs.docker.com/engine/reference/commandline/tag).

Параметры `--images-repo` и `--images-repo-mode` определяют адрес хранения образов и формат его имени в репозитории.
Если в конфигурации проекта (файл `werf.yaml`) описан только один **безымянный** образ, то значение параметра `--images-repo` используется в качестве имени Docker-репозитория без изменений. Имя Docker-образа в таком случае будет соответствовать шаблону: `IMAGES_REPO:TAG`.

Если в конфигурации проекта (файл `werf.yaml`) описано несколько образов, то шаблон имени Docker-образа будет зависеть от значения параметра `--images-repo-mode`:
- `IMAGES_REPO:IMAGE_NAME-TAG` в случае значения `monorepo` для параметра `--images-repo-mode`;
- `IMAGES_REPO/IMAGE_NAME:TAG` в случае значения `multirepo` для параметра `--images-repo-mode` (используется по умолчанию).

Большинство реализаций Docker registry позволяют создавать иерархию репозиториев, например, `COMPANY/PROJECT/IMAGE`. 
В этом случае вам нет необходимости менять значение парамера `--images-repo-mode` со значения по умолчанию — `multirepo`. 
Но, если у вас в проекте описано несколько образов, и вы работаете с Docker registry неподдерживающим иерархию репозиториев (самый яркий пример, это — [Docker Hub](https://hub.docker.com/)), то вам потребуется указать значение `monorepo` в параметре `--images-repo-mode`. 
Подробнее о работе режима _monorepo/multirepo_ можно прочитать в нашей [статье](https://habr.com/ru/company/flant/blog/465131/).

Значение, указываемое для параметра `--images-repo` может быть также передано через переменную окружения `$WERF_IMAGES_REPO`.

Значение, указываемое для параметра `--images-repo-mode` может быть также передано через переменную окружения `$WERF_IMAGES_REPO_MODE`.

> Именование образов должно быть одинаковым при выполнении команд _publish_, _build-and-publish_, _deploy_, а также команд процесса очистки. В противном случае вы можете получить не работающий pipeline и потерю образов и стадий по результатам работы очистки

*Docker-тег* определяется исходя из используемых параметров `--tag-*`:

| option                     | description                                                                     |
| -------------------------- | ------------------------------------------------------------------------------- |
| `--tag-git-tag TAG`        | Используется стратегия тегирования _git-tag_, — тегирование осуществляется по указанному git-тегу |
| `--tag-git-branch BRANCH`  | Используется стратегия тегирования _git-branch_, — тегирование осуществляется по указанной git-ветке |
| `--tag-git-commit COMMIT`  | Используется стратегия тегирования _git-commit_, — тегирование осуществляется по указанному хэшу git-коммита |
| `--tag-custom TAG`         | тегирование осуществляется по указанному произвольному тегу |

Все передаваемые параметры тегирования валидируются, с учетом стандартных ограничений Docker на содержание имени тега образа. При необходимости, вы можете использовать _слагификацию_ (slug, slugify — преобразование текста к виду, удобному для восприятия человеком) тегов, читай подробнее об этом в соответствующей [статье]({{ site.baseurl }}/documentation/reference/toolbox/slug.html).

Используя параметры `--tag-*`, вы не просто указываете тег, которым нужно протегировать образ, а также определяете стратегию тегирования. Стратегия тегирования влияет на операцию очистки. В частности, [политика очистки]({{ site.baseurl }}/documentation/reference/cleaning_process.html#очистка-по-политикам) выбирается в зависимости от стратегии тегирования, а в случае использования стратегии тегирования по произвольному тегу политика очистки не применяется совсем.

Использование параметров тегирования `--tag-git-*` подразумевает указание аргументов в виде значений тегов, веток и коммитов git. Все такие параметры разработаны с учетом совместимости с современными CI/CD системами, в которых выполнение задания pipeline CI происходит в отдельном экземпляре git-дерева для конкретного git-коммита, а имена тега, ветки и хэш коммита передаются через переменные окружения (например, для GitLab CI это `CI_COMMIT_TAG`, `CI_COMMIT_REF_NAME` и `CI_COMMIT_SHA`).

### Объединение параметов

Любые параметры тегирования могут использоваться одновременно в любом порядке при выполнении команды [werf publish]({{ site.baseurl }}/documentation/cli/main/publish.html) или [werf build-and-publish]({{ site.baseurl }}/documentation/cli/main/build_and_publish.html). В случае передачи нескольких параметров тегирования, werf создает отдельный образ на каждый переданный параметр тегирования, согласно каждому описанному в конфигурации проекта образу.

## Примеры

### Два образа для одного git-тега

Имеем файл конфигурации `werf.yaml` с описанными двумя образами  — `backend` и `frontend`.

Выполнение команды:

```shell
werf publish --stages-storage :local --images-repo registry.hello.com/web/core/system --tag-git-tag v1.2.0
```

приведет к созданию следующих образов соответственно:
 * `registry.hello.com/web/core/system/backend:v1.2.0`
 * `registry.hello.com/web/core/system/frontend:v1.2.0`

### Два образа для git-ветки

Имеем файл конфигурации `werf.yaml` с описанными двумя образами  — `backend` и `frontend`.

Выполнение команды:

```shell
werf publish --stages-storage :local --images-repo registry.hello.com/web/core/system --tag-git-branch my-feature-x
```

приведет к созданию следующих образов соответственно:
 * `registry.hello.com/web/core/system/backend:my-feature-x`;
 * `registry.hello.com/web/core/system/frontend:my-feature-x`.

### Два образа для git-ветки, в названии которой есть специальные символы

Имеем файл конфигурации `werf.yaml` с описанными двумя образами  — `backend` и `frontend`.

Выполнение команды:

```shell
werf publish --stages-storage :local --images-repo registry.hello.com/web/core/system --tag-git-branch $(werf slugify --format docker-tag "Features/MyFeature#169")
```

приведет к созданию следующих образов соответственно:
 * `registry.hello.com/web/core/system/backend:features-myfeature169-3167bc8c`;
 * `registry.hello.com/web/core/system/frontend:features-myfeature169-3167bc8c`.

Обратите внимание, что команда [`werf slugify`]({{ site.baseurl }}/documentation/cli/toolbox/slugify.html) генерирует значение, допустимое для использования в качестве Docker-тега. Читайте подробнее про команду _slug_ [в соответствующем разделе]({{ site.baseurl }}/documentation/reference/toolbox/slug.html).

### Два образа в задании GitLab CI

Имеем файл конфигурации `werf.yaml` с описанными двумя образами  — `backend` и `frontend`. Имеем проект `web/core/system` в GitLab, и сконфигурированный Docker registry для проекта — `registry.hello.com/web/core/system`.

Запуск следующей команды в задании pipeline GitLab CI для ветки `core/feature/ADD_SETTINGS`:

```shell
type werf && source <(werf ci-env gitlab --tagging-strategy tag-or-branch --verbose)
werf publish --stages-storage :local
```

приведет к созданию следующих образов соответственно:
 * `registry.hello.com/web/core/system/backend:core-feature-add-settings-df80fdc3`;
 * `registry.hello.com/web/core/system/frontend:core-feature-add-settings-df80fdc3`.

Обратите внимание, что werf автоматически применяет слагификацию к Docker-тегу `core/feature/ADD_SETTINGS`, — он будет конвертирован в тег `core-feature-add-settings-df80fdc3`. Эта конвертация выполняется при вызове команды `werf ci-env`, которая определяет название ветки из переменных окружения задания GitLab CI, слагифицирует его и заносит результат в переменную окружения `WERF_TAG_GIT_BRANCH` (альтернативный путь установки значения параметра `--tag-git-branch`). Читайте подробнее про команду _slug_ [в соответствующем разделе]({{ site.baseurl }}/documentation/reference/toolbox/slug.html).

### Безымянный образ в задании GitLab CI

Имеем файл конфигурации `werf.yaml` с описанным безымянным образом. Имеем проект `web/core/queue` в GitLab, и сконфигурированный Docker registry для проекта — `registry.hello.com/web/core/queue`.

Запуск следующей команды в задании pipeline GitLab CI для тега `v2.3.`:

```shell
type werf && source <(werf ci-env gitlab --tagging-strategy tag-or-branch --verbose)
werf publish --stages-storage :local
```

приведет к созданию образа `registry.hello.com/web/core/queue:v2.3.1`.
