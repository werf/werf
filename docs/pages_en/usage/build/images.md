---
title: Image configuration
permalink: usage/build/images.html
---

<!-- прим. для перевода: на основе https://werf.io/documentation/v1.2/reference/werf_yaml.html#image-section -->

To enable building your own images, you have to add an image description to the `werf.yaml` file of the project. Each image is added using the `image` directive and the image name:

```yaml
image: frontend
---
image: backend
---
image: database
```

The image name is a short symbolic name to refer to elsewhere in the configuration and when invoking werf commands. For example, you can use it to get the full name of the image published in the container registry or when running commands inside the built image using `werf kube-run`, etc. If multiple images are described in the configuration file, **each image** should have a name.

Next, for each image in `werf.yaml`, you must define the build instructions [using Dockerfile](#dockerfile-builder) or [stapel](#stapel-builder).

## Dockerfile builder

<!-- прим. для перевода: на основе https://werf.io/documentation/v1.2/reference/werf_yaml.html#dockerfile-builder -->

werf supports standard Dockerfiles for defining the image build instructions. Refer to the following resources for information about writing Dockerfiles:

* [Dockerfile Reference](https://docs.docker.com/engine/reference/builder/).
* [Best practices for writing Dockerfiles](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/).

The minimal configuration for a Dockerfile-based build is as follows:

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

You can also define multiple target images using different stages of the same Dockerfile:

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

Of course, you can also define images based on different Dockerfiles:

```yaml
# werf.yaml
image: backend
dockerfile: dockerfiles/Dockerfile.backend
---
image: frontend
dockerfile: dockerfiles/Dockerfile.frontend
```

The `context` directive sets the build context. **Note** that in this case,the path to the dockerfile must be specified relative to the context directory:

```yaml
image: docs
context: docs
dockerfile: Dockerfile
---
image: service
context: service
dockerfile: Dockerfile
```

Here, the `docs` image will use the `docs/Dockerfile` path, and the `service` image will use the `service/Dockerfile` path.

### contextAddFiles

By default, the build context of a Dockerfile image includes only the files from the current commit in the project repository. Files not added to git or non-committed changes are not included in the build context. This logic follows the default [giterminism] settings({{"/usage/project_configuration/giterminism.html" | true_relative_url }}).

Use the `contextAddFiles` directive in `werf.yaml` to add files that are not stored in git to the build context. You also have to enable the `contextAddFiles` directive in `werf-giterminism.yaml` (more [about giterminism]({{"/usage/project_configuration/giterminism.html#contextaddfiles" | true_relative_url }}): 

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

The build context for the above configuration will include the following files:

- `app/**/*` from the current commit in the project repository;
- `app/file1`, `app/dir2/file2.out` and the `dir1` directory, which are in the project directory.

## Stapel builder

werf has a built-in alternative syntax for describing build instructions called Stapel. Refer to the [documentation]({{ "/usage/build/stapel/overview.html" | true_relative_url }}) to learn more about the Stapel syntax.


Below is an example of a minimal stapel image configuration in `werf.yaml`:

```yaml
image: app
from: ubuntu:22.04
```

Add the git sources to the image:

```yaml
image: app
from: ubuntu:22.04
git:
- add: /
  to: /app
```

There are 4 stages available to define arbitrary shell instructions, as well as a `git.stageDependencies` directive to set up triggers to rebuild these stages when the corresponding stages change (see [more]({{"/usage/build/stapel/instructions.html#dependency-on-changes-in-the-git-repo"| true_relative_url }})):

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

Stapel supports auxiliary images from which you can import files into the target image (similar to `COPY-from=STAGE` in the multi-stage Dockerfile) as well as golang templating:

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

See [the stapel section]({{"usage/build/stapel/base.html" | true_relative_url }}) for detailed instructions.

## Inheriting images and importing files

The multi-stage mechanism allows you to declare a single image as a stage in the Dockerfile and use it as a base for another image, or copy specific files from it.

werf extends it beyond just one Dockerfile and allows you to use arbitrary images defined in `werf.yaml`, including those built from different Dockerfiles or built by the Stapel builder. All orchestration and dependency building will be done by werf, and the build will be accomplished in one step (as part of the `werf build` command).

Below is an example of using the `base.Dockerfile` image as the base image for the `Dockerfile` image:

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

The following example shows how to import files from a Stapel image into a Dockerfile image:

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

werf supports passing information about the built image to another image during the build. For example, the following configuration passes the names and digests of the `auth` and `controlplane` images published in the container registry to the `app` image:

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
  - type: ImageID
    targetBuildArg: AUTH_IMAGE_DIGEST
- image: controlplane
  imports:
  - type: ImageName
    targetBuildArg: CONTROLPLANE_IMAGE_NAME
  - type: ImageID
    targetBuildArg: CONTROLPLANE_IMAGE_DIGEST
```

During the build werf will automatically insert the appropriate names and identifiers into the specified build-arguments. All orchestration and dependency building will be handled by werf and the build will be done in one step (as part of the `werf build` command).
