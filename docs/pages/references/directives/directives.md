---
title: Dappfile directives
sidebar: reference
permalink: directives.html
folder: directive
---

## Types of dappfile syntax 

Dapp builds images by instructions in special files, *dappfiles*.
They can be written in two different syntaxes with corresponding file names:

*   `dappfile.yml`, in YAML syntax;
*   `Dappfile`, in Ruby syntax.

    **Deprecation warning:** support for Ruby syntax will be discontinued in one of the future releases.
    We recommend using YAML syntax for new dappfiles.

Описание образов приложений и образов артефактов в dappfile выполняется соответственно с помощью директив `dimg` и `artifact`, для обоих синтаксисов dappfile.

Указание базового образа, на основе которого будет производиться сборка образа приложения или образа артефакта, осуществляется с помощью директивы `from`.

Both types share some some common features:

*   Application images are declared with `dimg` directive, and artifact images with `artifact` directive.
*   Base image for an application or artifact image is declared with `from` directive.

Основное различие при использовании разного синтаксиса dappfile в части описания образов - в YAML синтаксисе используется линейное описание образов и отсутствует вложенность, в Ruby синтаксисе же, можно использовать директиву `dimg_group` и описывать вложенные контексты (см. ниже.).

A major difference is that YAML syntax is linear and has no nesting, while Ruby syntax can describe nested context with `dimg_group` directive.

## YAML синтаксис (dappfile.yml)

## YAML syntax (dappfile.yml)

### General features

This syntax is based on YAML and has much in common with Dockerfiles and Ansible playbooks.

Here is an example of a minimal `dappfile.yml`:

```yaml
dimg: example
from: alpine
```

An image description should at least include the following:

* image type and name, set with `dimg`, or `artifact` directive;
* base image, set with `from`, `fromDimg`, or `fromDimgArtifact` directive.

A dappfile can have multiple image descriptions, separated by `---`, as in YAML specification.
For more details on this topic, see [Describing Multiple Images](#describing-multiple-images).

### dimg

The dimg directive starts a description for building an application image:

```yaml
dimg: [<name>[:<tag>]]
```

Name of an application image is a string, similar to the image name in Docker.
The part after `:` sets the image tag and is usually used for versioning images.
If absent, tag defaults to `latest`.

```yaml
dimg: frontend:latest

# same 
dimg: frontend
```

Name of an application image can be empty or equal to `~`.
In a dappfile with multiple images only one can have an empty name.

```yaml
dimg:
from: alpine
---
dimg: this_one_should_have_a_name
from: alpine
```

An application image can have several names, set as a list in YAML syntax.
This is equal to describing similar images with different names.

```yaml
dimg: [main-front,main-back]
```

### artifact

The `artifact` directive starts a description for building an artifact image:

```yaml
artifact: <name>[:<tag>]
from: <image>
```

Name of an artifact image is a string, similar to the image name in Docker.
The part after `:` sets the image tag and is usually used for versioning images.
If absent, tag defaults to `latest`.

Artifact name is used in `import` directive (see [artifact](artifact.html) for details).
Thus artifact images must have non-empty names.

```
artifact: hello_world
from: alpine
```

### from

The `from` directive sets the name of a base image to build from.

```yaml
from: <image>[:<tag>]
```

If absent, tag defaults to `latest`.

### Describing Multiple Images

A `dappfile.yml` can describe any number of images, but at least one of them should be an application image.
To describe several images, separate them with `---`.

* There should be at least one application image.
* There can be none or one application image with empty name, others should have non-empty names.
* All artifact images should have non-empty names.
* Image names should be unique.

Here's an example of describing multiple images in one dappfile:

```yaml
dimg: frontend
from: nginx:latest
shell:
  setup:
    - echo "Setup NGINX"
---
dimg: backend
from: node:latest
shell:
  setup:
    - echo "Setup NodeJS"
---
dimg: db
from: mysql
shell:
  setup:
    - echo "Setup database"
```

Правила описания образов в dappfile.yml:
* можно описывать несколько образов
* должен быть описан как минимум один образ приложения
* образ приложения может быть безымянным, но только один
* отсутствует вложенность описаний образов и контексты (нет директивы `dimg_group`)
* при описании нескольких образов они описываются линейно - друг за другом, отделяются строкой состоящей из последовательности `---` (согласно YAML спецификации)
* артефакты обязаны иметь имя
* имена образов приложений и образов артефактов должны различаться
* отсутствует директива `export` в описании образов артефактов
* описание образа приложения выполняется с помощью директивы `dimg: <name>`, где `<name>` - имя образа
    * для безымянного образа, имя образа может отсутствовать либо быть равным `~`
    * имя образа может быть представленно массивом строк (например - `dimg: [main-front,main-back]`), это равнозначно описанию отдельных образов с одинаковой конфигурацией
* Описание образа артефакта выполняется с помощью директивы `artifact: <name>`, где `<name>` - обязательное имя артефакта, оно используется для указания артефакта по имени, при импорте файлов (директива `import`, см. [подробней](artifact.html) про артефакты)
* Указание базового образа при описании образа приложения или образа артефакта выполняется с помощью обязательной директивы `from: <DOCKER_IMAGE>`, где `<DOCKER_IMAGE>` - имя образа в формате `image:tag` (tag может отсутствовать, тогда используется latest)

### Пример описания минимального dappfile.yml

```
dimg:
from: alpine
```

### Пример сборки нескольких образов

```
dimg: frontend
from: nginx:latest
shell:
  setup:
    - echo "Setup NGINX"
---
dimg: backend
from: node:latest
shell:
  setup:
    - echo "Setup NodeJS"
---
dimg: db
from: mysql
shell:
  setup:
    - echo "Setup database"
```


## Ruby синтаксис (Dappfile)

Правила описания образов в Dappfile:

* Наследуются все настройки родительского контекста.
* Можно дополнять или переопределять настройки родительского контекста.
* описание образа приложения выполняется с помощью директивы `dimg <name>`, где `<name>` - имя образа (может отсутствовать)

* можно описывать несколько образов
* должен быть описан как минимум один образ приложения
* образ приложения может быть безымянным, но только один
* на верхнем уровне могут использоваться только директивы `dimg` и `dimg_group`.
* артефакты не обязаны иметь имя
* имена образов приложений и образов артефактов должны различаться
* описание образа приложения выполняется с помощью директивы `dimg <name> do`, где `<name>` - имя образа (может отсутствовать)
* описание образа артефакта выполняется с помощью директивы `artifact <name> do`, где `<name>` - имя артефакта (см. [подробней](artifact.html) про артефакты)
* Указание базового образа при описании образа приложения или образа артефакта выполняется с помощью обязательной директивы `docker.from <DOCKER_IMAGE>`, где `<DOCKER_IMAGE>` - имя образа в формате `image:tag` (tag может отсутствовать, тогда используется latest)


### Собирать образ, не указывая имени
```ruby
dimg do
  docker.from 'image:tag'
end
```

### Собирать образы X и Y
```ruby
dimg 'X' do
  docker.from 'image1:tag'
end

dimg 'Y' do
  docker.from 'image2:tag'
end
```

### Создавать контексты для сборки

При сборке можно использовать контексты с помощью директивы `dimg_group`. Их можно в том числе вкладывать друг в друга:

```ruby
dimg_group do
  docker.from 'image:tag'

  dimg 'X'

  dimg_group do
    dimg 'Y'

    dimg_group do
      dimg 'Z'
    end
  end
end
```
