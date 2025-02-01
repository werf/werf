---
title: Images and dependencies
permalink: usage/build/images.html
---

<!-- reference: https://werf.io/docs/v2/reference/werf_yaml.html#image-section -->

## Adding images

To build images with werf, you have to add a description of each image to the `werf.yaml` file of the project. The image description starts with the `image` directive specifying the image name:

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

> The image name is a unique internal image identifier that allows it to be referenced during configuration and when invoking werf commands.

Next, for each image in `werf.yaml`, you have to define the build instructions using [Dockerfile](#dockerfile) or [Stapel](#stapel).

### Dockerfile

<!-- reference: https://werf.io/docs/v2/reference/werf_yaml.html#dockerfile-builder -->

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
project: example
configVersion: 1
---
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

... as well as define images based on different Dockerfiles:

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

#### Selecting the build context directory

The `context` directive sets the build context. **Note:** In this case, the path to the Dockerfile must be specified relative to the context directory:

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

In the example above, werf will use the Dockerfile at `docs/Dockerfile` to build the `docs` image and the Dockerfile at `service/Dockerfile` to build the `service` image.

#### Using build secrets

> **NOTE:** To use secrets in builds, you need to enable them explicitly in the giterminism settings. Learn more ([here]({{ "/usage/project_configuration/giterminism.html#using-build-secrets" | true_relative_url }}))

A build secret is any sensitive information, such as a password or API token, required during the build process of your application.

Build arguments and environment variables are not suitable for passing secrets to the build process, as they are retained in the final image. 

Instead, you can define secrets in the `werf.yaml` file to use them during the build process.

```yaml
# werf.yaml
project: example
configVersion: 1
---
image: backend
dockerfile: Dockerfile
secrets:
  - env: AWS_ACCESS_KEY_ID
  - id: aws_secret_key
    env: AWS_SECRET_ACCESS_KEY
  - src: "~/.aws/credentials"
  - id: plainSecret
    value: plainSecretValue
```

```yaml
# werf-giterminism.yaml
giterminismConfigVersion: 1

config:
  secrets:
    allowEnvVariables:
      - "AWS_ACCESS_KEY_ID"
    allowFiles:
      - "~/.aws/credentials"
    allowValueIds:
      - plainSecret
```

To access a secret during the build, use the `--mount=type=secret` flag in the Dockerfile's `RUN` instructions.

When a secret is mounted, it is available as a file by default. The default mount path for the secret file inside the build container is `/run/secrets/<id>`. If an `id` is not specified for a secret in `werf.yaml`, the default value is assigned automatically:

- For `env`, the `id` defaults to the name of the environment variable.
- For `src`, the `id` defaults to the file name (e.g., for `/path/to/file`, the `id` will be `file`).

> For `value` — the `id` field is mandatory.

```Dockerfile
# Dockerfile
FROM alpine:3.18

# Example of using a secret from an environment variable
RUN --mount=type=secret,id=AWS_ACCESS_KEY_ID \
    export WERF_BUILD_SECRET="$(cat /run/secrets/AWS_ACCESS_KEY_ID)"

# Example of using a secret from a secrets file
RUN --mount=type=secret,id=credentials \
    AWS_SHARED_CREDENTIALS_FILE=/run/secrets/credentials \
    aws s3 cp ...

# Example of mounting a secret as an environment variable
RUN --mount=type=secret,id=AWS_ACCESS_KEY_ID,env=AWS_ACCESS_KEY_ID \
    --mount=type=secret,id=aws-secret-key,env=AWS_SECRET_ACCESS_KEY \
    aws s3 cp ...

# Example of mounting a secret as a file with a different name
RUN --mount=type=secret,id=credentials,target=/root/.aws/credentials \
    aws s3 cp ...

# Example of using a raw value that will not be stored in the final image
RUN --mount=type=secret,id=plainSecret \
    export WERF_BUILD_SECRET="$(cat /run/secrets/plainSecret)"
```

### Using the SSH agent

You can provide access to the SSH agent socket or SSH keys during the build process. This is particularly useful if your Dockerfile contains commands that require SSH authentication, such as cloning a private repository.

To enable this, use the `--mount=type=ssh` flag with RUN commands:

```Dockerfile
FROM alpine
RUN apk add --no-cache openssh-client
RUN mkdir -p -m 0700 ~/.ssh && ssh-keyscan gitlab.com >> ~/.ssh/known_hosts
RUN --mount=type=ssh ssh -q -T git@gitlab.com 2>&1 | tee /hello
```

You can find detailed information about using the SSH agent in werf [here]({{ "/usage/build/process.html#using-the-ssh-agent" | true_relative_url }}).

#### Adding arbitrary files to the build context

By default, the build context of a Dockerfile image only includes files from the current project repository commit. Files not added to Git or non-committed changes are not included in the build context. This logic follows the [giterminism configuration]({{"/usage/project_configuration/giterminism.html" | true_relative_url }}) default.

You can add files that are not stored in Git to the build context through the `contextAddFiles` directive in the `werf.yaml` file. You also have to enable the `contextAddFiles` directive in `werf-giterminism.yaml` (more [about giterminism]({{"/usage/project_configuration/giterminism.html#contextaddfiles" | true_relative_url }})):

```yaml
# werf.yaml
project: example
configVersion: 1
---
image: app
dockerfile: Dockerfile
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

In the above configuration, the build context will include the following files:

- `app/**/*` of the current project repository commit;
- `app/file1`, `app/dir2/file2.out` files and the `dir1` directory in the project directory.

#### Multiplatform Build

werf supports multiplatform and cross-platform builds, allowing you to create images for various architectures and operating systems (for more details, see [the relevant section of the documentation]({{ "/usage/build/process.html#multi-platform-and-cross-platform-building" | true_relative_url }})).

##### Cross-compilation

If your project requires cross-compilation, you can use multi-stage builds to create artifacts for target platforms using the build platform. The following build arguments are available for this purpose:

- `TARGETPLATFORM`: the platform for the build result (e.g., linux/amd64, linux/arm/v7, windows/amd64).
- `TARGETOS`: the OS of the target platform.
- `TARGETARCH`: the architecture of the target platform.
- `TARGETVARIANT`: the variant of the target platform.
- `BUILDPLATFORM`: the platform of the node performing the build.
- `BUILDOS`: the OS of the build platform.
- `BUILDARCH`: the architecture of the build platform.
- `BUILDVARIANT`: the variant of the build platform.

Example:

```Dockerfile
FROM --platform=$BUILDPLATFORM golang:alpine AS build
ARG TARGETPLATFORM
ARG BUILDPLATFORM
RUN echo "I am running on $BUILDPLATFORM, building for $TARGETPLATFORM" > /log

FROM alpine
COPY --from=build /log /log
```

In this example, the `FROM` instruction is pinned to the build platform of the builder using the `--platform=$BUILDPLATFORM` option to prevent emulation. The `$BUILDPLATFORM` and `$TARGETPLATFORM` arguments are then used in the `RUN` instruction.

### Stapel

Stapel is a built-in alternative syntax for describing build instructions. Detailed documentation on Stapel syntax is available [in the corresponding documentation section]({{"/usage/build/stapel/overview.html" | true_relative_url }}).

Below is an example of a minimal configuration of the Stapel image in `werf.yaml`:

```yaml
project: example
configVersion: 1
---
image: app
from: ubuntu:22.04
```

And here's how you can add source files from Git to an image:

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

There are 4 stages for describing arbitrary shell instructions, as well as a `git.stageDependencies` directive to set up triggers to rebuild these stages whenever the corresponding stages change ([learn more]({{"/usage/build/stapel/instructions.html#dependency-on-changes-in-the-git-repo"| true_relative_url }})):

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

Auxiliary images and Golang templating are also supported. You can use the former to import files into the target image (similar to `COPY-from=STAGE` in multi-stage Dockerfile):

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

For more info on how to write Stapel instructions refer to the [documentation]({{"usage/build/stapel/base.html" | true_relative_url }}).

## Changing image configuration spec

In OCI (Open Container Initiative), [image configuration spec](https://github.com/opencontainers/image-spec/blob/main/config.md) is the image specification that describes its structure and metadata. The `imageSpec` directive in `werf.yaml` provides flexible options for managing and configuring various aspects of images:

- Flexibility in managing specification fields.
- Removal or resetting of unnecessary components: labels, environment variables, volumes, commands, and build history.
- A unified configuration mechanism for all supported backends and syntaxes.
- Rules that apply both to all images in a project and to individual images.

### Global configuration

Example configuration that will apply to all images in the project:

```yaml
project: test
configVersion: 1
build:
  imageSpec:
    author: "Frontend Maintainer <frontend@example.com>"
    clearHistory: true
    config:
      removeLabels:
        - "unnecessary-label"
        - /org.opencontainers.image..*/
      labels:
        app: "my-app"
```

This configuration will be applied to all images in the project: labels and author will be set for all images, and unnecessary labels will be removed.

### Configuration for a specific image

Example configuration for an individual image:

```yaml
project: test
configVersion: 1
---
image: frontend_image
from: alpine
imageSpec:
  author: "Frontend Maintainer <frontend@example.com>"
  clearHistory: true
  config:
    user: "1001:1001"
    exposedPorts:
      - "8080/tcp"
    env:
      NODE_ENV: "production"
      API_URL: "https://api.example.com"
    entrypoint:
      - "/usr/local/bin/start.sh"
    volumes:
      - "/app/data"
    workingDir: "/app"
    labels:
      frontend-version: "1.2.3"
    stopSignal: "SIGTERM"
    removeLabels:
      - "old-frontend-label"
      - /old-regex-label.*/
    removeVolumes:
      - "/var/cache"
    removeEnv:
      - "DEBUG"
```

> **Note:** Configuration for a specific image takes precedence over global configuration. String values will be overwritten, and for multi-valued directives, the data will be merged based on priority.

### Build process changes

Changing the image configuration does not directly affect the build process but allows you to configure aspects such as removing unnecessary volumes or adding environment variables for the base image. Example:

```yaml
image: base
from: postgres:12.22-bookworm
imageSpec:
  config:
    removeVolumes:
      - "/var/lib/postgresql/data"
---
image: app
fromImage: base
git:
  add: /postgresql/data
  to: /var/lib/postgresql/data
```

In this example, the base image `postgres:12.22-bookworm` has unnecessary volumes removed, which can then be used in the `app` image.


## Linking images

### Inheritance and importing files

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

The next example deals with importing files from a Stapel image into a Dockerfile image:

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
ARG BUILDER_IMAGE
FROM ${BUILDER_IMAGE} AS builder

FROM alpine
COPY --from=builder /app/bin /app/bin
CMD [ "/app/bin/server", "server" ]
```

### Passing information about the built image to another image

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

During the build, werf will automatically insert the appropriate names and identifiers into the referenced build-arguments. werf will take care of all orchestration and dependency mapping and then build everything in one step (as part of the `werf build` command).

## Using intermediate and final images

By default, all images are final, allowing the user to operate with them using their names as arguments for most werf commands and in Helm chart templates. The `final` directive can be used to regulate this property of an image.

Intermediate images (`final: false`), unlike final images:
- Do not appear [in the service values for the Helm chart]({{ "usage/deploy/values.html#information-about-the-built-images-werf-only" | true_relative_url }}).
- Are not tagged with arbitrary tags ([more about --add-custom-tag]({{ "usage/build/process.html#adding-custom-tags" | true_relative_url }})).
- Are not published to the final repository ([more about --final-repo]({{ "usage/build/process.html#extra-repository-for-final-images" | true_relative_url }})).
- Are not exported ([more about werf export]({{ "reference/cli/werf_export.html" | true_relative_url }})).

Example of using the `final` directive:

```yaml
project: example
configVersion: 1
---
image: builder
final: false
dockerfile: Dockerfile.builder
---
image: app
dockerfile: Dockerfile.app
dependencies:
- image: builder
  imports:
  - type: ImageName
    targetBuildArg: BUILDER_IMAGE_NAME
```
