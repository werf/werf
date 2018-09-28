---
title: Base directives
sidebar: reference
permalink: reference/build/base_directives.html
---

## What is required to build dimg?

Here's a minimal `dappfile.yml`. It builds an image named `example` from a base image named `alpine`:

```yaml
dimg: example
from: alpine
```

An image description requires at least two parameters:

* image type and name, set with `dimg`, or `artifact` directive;
* base image, set with `from`, `fromDimg`, or `fromDimgArtifact` directives.

Application images are declared with dimg directive, and artifact images with artifact directive.
Base image for an application or artifact image is declared with from directive.

### Dimg

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

### Base image
#### From

```yaml
from: <image>[:<tag>]
```

The `from` directive sets the name and tag of a base image to build from.
If absent, tag defaults to `latest`.

#### FromDimg, FromDimgArtifact

```yaml
fromDimg: <name>
---
fromDimgArtifact: <name>
```

Besides using docker images from a repository, one can refer to images, previously described in the same `dappfile.yml`.
This can be helpful if default [build stages]({{ site.baseurl }}/reference/stages/stage.html) are not enough for building an image.

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

## Stage from

Данная стадия производит скачивание указанного базового образа (фактически docker pull) и фиксирует его в кэше dapp.

* Стадия используется только при указании базового образа директивой docker.from с аргументом в формате \<image:tag\>.

See more [here]({{site.baseurl}}/not_used/directives.html)
