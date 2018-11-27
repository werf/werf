---
title: Dappfile directives
sidebar: not_used
permalink: not_used/directives.html
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

Both types share some common features:

*   Application images are declared with `dimg` directive, and artifact images with `artifact` directive.
*   Base image for an application or artifact image is declared with `from` directive.

Основное различие при использовании разного синтаксиса dappfile в части описания образов - в YAML синтаксисе используется линейное описание образов и отсутствует вложенность, в Ruby синтаксисе же, можно использовать директиву `dimg_group` и описывать вложенные контексты (см. ниже.).

A major difference is that YAML syntax is linear and has no nesting, while Ruby syntax can describe nested context with `dimg_group` directive.

## YAML syntax (dappfile.yml)

> YAML синтаксис (dappfile.yml)

### Introduction

Dappfile describes building one or more images.
Its syntax is based on [YAML](http://yaml.org/spec/1.2/spec.html).

Here's a minimal `dappfile.yml`. It builds an image named `example` from a base image named `alpine`:

```yaml
dimg: example
from: alpine
```

An image description requires at least two parameters:

* image type and name, set with `dimg`, or `artifact` directive;
* base image, set with `from`, `fromDimg`, or `fromDimgArtifact` directive.sss

A single dappfile can build multiple images.
For more details on this topic, see [Building Multiple Images](#building-multiple-images).

### `dimg`

```yaml
dimg: [<name>|~]
```

The dimg directive starts a description for building an application image.
The image name is a string, similar to the image name in Docker:

```yaml
dimg: frontend
```

An image can be nameless: `dimg: ` or `dimg: ~`.
In a dappfile with multiple images there can be only one nameless application image.

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

### `artifact`

```yaml
artifact: <name>
```

The `artifact` directive starts a description for building an artifact image:

```
artifact: hello_world
```

Name of an artifact image is a string, similar to the image name in Docker.
Artifact name is used to `import` artifacts.
Thus artifact images must have non-empty names.

Artifact images can have multiple names, just like application images:

```yaml
artifact: [main-front,main-back]
```

### `from`

```yaml
from: <image>[:<tag>]
```

The `from` directive sets the name and tag of a base image to build from.
If absent, tag defaults to `latest`.


### `fromDimg` and `fromDimgArtifact`

```yaml
fromDimg: <name>
---
fromDimgArtifact: <name>
```

Besides using docker images from a repository, one can refer to images, previously described in the same `dappfile.yml`.
This can be helpful if default build stages are not enough for building an image.

```yaml
dimg: intermediate
from: alpine
...
---
dimg: app
fromDimg: intermediate
...
```

If an intermediate image is specific to a particular application,
it is reasonable to store its description with that application:

```yaml
artifact: intermediate
from: ubuntu:16.04
shell:
  install: 
  - rm /etc/apt/apt.conf.d/docker-clean
  - apt-get update
mount:
- from: build_dir
  to: /var/cache/apt
---
dimg: ~
fromDimgArtifact: intermediate 
shell:
  beforeInstall: apt-get install tree
```

## Building Multiple Images

### General Syntax

A `dappfile.yml` can describe any number of images, but at least one of them should be an application image.
To describe several images, separate them with `---`:

```yaml
<image 1>
---
<image 2>
---
...
```

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

### Advanced Techniques

В приведённом ниже примере будет собрано два образа: `curl_and_shell` и `nginx`:

В приведённом примере вы также можете заметить полезные для этого кейса вещи:

- объявление переменной, которую можно повторно использовать в разных image-ах. Делается на основании правил go templates.
- повторяюшийся блок вынесен в include.

Dapp has two powerful features for describing and building multiple images:

* Templating builds with [Go templates syntax](https://golang.org/pkg/text/template/#hdr-Functions) and [Sprig functions](http://masterminds.github.io/sprig/).
* Reusing code blocks [with `include`]( {{ site.baseurl }}/faq.html#dappfile-7).

The code below demonstrates Go templates and `include`:

{% raw %}
```yaml
# Declaring a "BaseImage" variable
{{ $_ := set . "BaseImage" "myregistry.myorg.com/dapp/ubuntu-dimg:10" }}

dimg: "curl_and_shell"
from: "{{ .BaseImage }}"
docker:
  WORKDIR: /app
  USER: app
  ENV:
    TERM: xterm
ansible:
  beforeInstall:
  - name: "Install Curl"
    apt:
      name: curl
      state: present
      update_cache: yes
git:
  # Including block "git application files"
  {{- include "git application files" . }}

---

dimg: "nginx"
from: "{{ .BaseImage }}"
docker:
  WORKDIR: /app
  USER: app
  ENV:
    TERM: xterm
ansible:
  beforeInstall:
  - name: "Install nginx"
    apt:
      name: nginx
      state: present
      update_cache: yes
git:
  # Including block "git application files"
  {{- include "git application files" . }}

###################################################################

# Declaring a reusable code block

{{- define "git application files" }}
  - add: /
    to: /app
    excludePaths:
    - .helm
    - .gitlab-ci.yml
    - .dappfiles
    - dappfile.yaml
    owner: app
    group: app
    stageDependencies:
      install:
      - composer.json
      - composer.lock
{{- end }}
```
{% endraw %}

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
* описание образа артефакта выполняется с помощью директивы `artifact <name> do`, где `<name>` - имя артефакта
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
