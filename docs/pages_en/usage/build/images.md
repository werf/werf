---
title: Images and dependencies
permalink: usage/build/images.html
---

<!-- reference: https://werf.io/documentation/v1.2/reference/werf_yaml.html#image-section -->

## Adding images

To build images with werf, you have to add a description of each image to the `werf.yaml` file of the project. The image description starts with the `image` directive specifying the image name:

```yaml
image: frontend
# ...
---
image: backend
# ...
---
image: database
# ...
```

> The image name is a unique internal image identifier that allows it to be referenced during configuration and when invoking werf commands.

Next, for each image in `werf.yaml`, you have to define the build instructions using [Dockerfile](#dockerfile) or [Stapel](#stapel).

### Dockerfile

<!-- reference: https://werf.io/documentation/v1.2/reference/werf_yaml.html#dockerfile-builder -->

#### Writing Dockerfile instructions

The format of the image build instructions matches the Dockerfile format. Refer to the following resources for more information:

* [Dockerfile Reference](https://docs.docker.com/engine/reference/builder/).
* [Best practices for writing Dockerfiles](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/).

#### Using Dockerfiles

The typical configuration of a Dockerfile-based build looks as follows:

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
image: backend
dockerfile: Dockerfile
```

#### Using a specific Dockerfile stage

You can also define multiple target images using various stages of the same Dockerfile:

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
image: backend
dockerfile: Dockerfile
target: backend
---
image: frontend
dockerfile: Dockerfile
target: frontend
```

... as well as define images based on different Dockerfiles:

```yaml
# werf.yaml
image: backend
dockerfile: dockerfiles/Dockerfile.backend
---
image: frontend
dockerfile: dockerfiles/Dockerfile.frontend
```

#### Selecting the build context directory

The `context` directive sets the build context. **Note:** In this case, the path to the Dockerfile must be specified relative to the context directory:

```yaml
image: docs
context: docs
dockerfile: Dockerfile
---
image: service
context: service
dockerfile: Dockerfile
```

In the example above, werf will use the Dockerfile at `docs/Dockerfile` to build the `docs` image and the Dockerfile at `service/Dockerfile` to build the `service` image.

#### Adding arbitrary files to the build context

By default, the build context of a Dockerfile image only includes files from the current project repository commit. Files not added to Git or non-committed changes are not included in the build context. This logic follows the [giterminism configuration]({{"/usage/project_configuration/giterminism.html" | true_relative_url }}) default.

You can add files that are not stored in Git to the build context through the `contextAddFiles` directive in the `werf.yaml` file. You also have to enable the `contextAddFiles` directive in `werf-giterminism.yaml` (more [about giterminism]({{"/usage/project_configuration/giterminism.html#contextaddfiles" | true_relative_url }})):

```yaml
# werf.yaml
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
    - file1
    - dir1/
    - dir2/file2.out
```

In the above configuration, the build context will include the following files:

- `app/**/*` of the current project repository commit;
- `app/file1`, `app/dir2/file2.out` files and the `dir1` directory in the project directory.

### Stapel

Stapel is a built-in alternative syntax for describing build instructions. Detailed documentation on Stapel syntax is available [in the corresponding documentation section]({{"/usage/build/stapel/overview.html" | true_relative_url }}).

Below is an example of a minimal configuration of the Stapel image in `werf.yaml`:

```yaml
image: app
from: ubuntu:22.04
```

And here's how you can add source files from Git to an image:

```yaml
image: app
from: ubuntu:22.04
git:
- add: /
  to: /app
```

There are 4 stages for describing arbitrary shell instructions, as well as a `git.stageDependencies` directive to set up triggers to rebuild these stages whenever the corresponding stages change ([learn more]({{"/usage/build/stapel/instructions.html#dependency-on-changes-in-the-git-repo"| true_relative_url }})):

```yaml
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

Auxiliary images and Golang templating are also supported. You can use the former to import files into the target image (similar to `COPY-from=STAGE` in multi-stage Dockerfile):

{% raw %}
```yaml
{{ $_ := set . "BaseImage" "ubuntu:22.04" }}

{{ define "package:build-tools" }}
  - apt update -q
  - apt install -y gcc g++ build-essential make
{{ end }}

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

For more info on how to write Stapel instructions refer to the [documentation]({{"usage/build/stapel/base.html" | true_relative_url }}).

## Inheriting images and importing files

A multi-stage mechanism allows you to define a separate image stage in a Dockerfile and use it as a basis for another image or copy individual files from it.

werf allows to do this not only within a single Dockerfile, but also for arbitrary images defined in `werf.yaml`, including those built from different Dockerfiles, or built by the Stapel builder. werf will take care of all orchestration and dependency management and then build everything in one step (as part of the `werf build` command).

Below is an example of using a `base.Dockerfile` image as the basis for a `Dockerfile` image:

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

The next example deals with importing files from a Stapel image into a Dockerfile image:

```yaml
# werf.yaml
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

## Passing information about the built image to another image

werf allows you to get information about the built image when building another image. Suppose, the build instructions of the `app` image require the names and digests of the `auth` and `controlplane` images published in the container registry. The configuration in this case would look like this:

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

During the build, werf will automatically insert the appropriate names and identifiers into the referenced build-arguments. werf will take care of all orchestration and dependency mapping and then build everything in one step (as part of the `werf build` command).

## Building for target platform

Using `--platform` parameter user can choose target platform for built images:

```shell
werf build --platform linux/arm64
```

— werf builds all final images from werf.yaml for the target platform using emulation.

You can also set target platform with `build.platform` configuration directive:

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

In this case running `werf build` command without parameters will start building images for the specified platform (explicitly passed `--platform` parameter redefines werf.yaml setting).

> **NOTE:** Preparation of host-system for multiplatform building, and also additional information about multiplatform support for different syntaxes and backends are available [in the building process article]({{ "/usage/build/process.html" | true_relative_url }}).

### Build multiplatform images

werf supports buildling of images for multiple platforms at the same time. In such case werf publishes special manifest into the container registry, this manifest includes built images for each specified target platform (when pulling such an image a client will receive image built for the client platform).

Tou can define list of target platforms with such configuration:

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

This list also can be redefined by the `--platform` parameter like that:

```shell
werf build --platform=linux/amd64,linux/i386
```
