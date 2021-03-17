---
title: Artifacts
sidebar: documentation
permalink: guides/advanced_build/artifacts.html
author: Artem Kladov <artem.kladov@flant.com>
---

## Task Overview

When you build an application image, it is often necessary to download temporary files or packages for build. In the results, the application image contains files that are not needed to run the application.

werf can [import]({{ site.baseurl }}/configuration/stapel_image/import_directive.html) resources from images and [artifacts]({{ site.baseurl }}/configuration/stapel_artifact.html). Thus you can isolate build process and tools in other images and then copy result files to reduce the image size. It is like a docker [multi-stage builds](https://docs.docker.com/develop/develop-images/multistage-build/) which are supported starting with Docker 17.05, but has more advanced files importing options.

In this article, we will build an example GO application. Then we will optimize the build instructions to substantial reduce image size with using artifacts.

## Requirements

* Installed [werf dependencies]({{ site.baseurl }}/guides/installation.html#installing-dependencies) on the host system.

* Installed [multiwerf](https://github.com/werf/multiwerf) on the host system.

### Select werf version

This command should be run prior running any werf command in your shell session:

```shell
. $(multiwerf use 1.1 stable --as-file)
```

## Sample application

The example application is the [Go Web App](https://github.com/josephspurrier/gowebapp), written in [Go](https://golang.org/).

### Building

Create a `gowebapp` directory and place the following `werf.yaml` in the `gowebapp` directory:
{% raw %}
```yaml
project: gowebapp
configVersion: 1
---

image: gowebapp
from: golang:1.14
docker:
  WORKDIR: /app
ansible:
  install:
  - name: Getting packages
    shell: go get github.com/josephspurrier/gowebapp
  setup:
  - file:
      path: /app
      state: directory
  - name: Copying config
    shell: |
      cp -r $GOPATH/src/github.com/josephspurrier/gowebapp/config /app/config
      cp -r $GOPATH/src/github.com/josephspurrier/gowebapp/static /app/static
      cp -r $GOPATH/src/github.com/josephspurrier/gowebapp/template /app/template
      cp $GOPATH/bin/gowebapp /app/
```
{% endraw %}

The config describes instructions to build one image — `gowebapp`.

Build the application by executing the following command in the `gowebapp` directory:

```shell
werf build --stages-storage :local
```

### Running

Run the application by executing the following command in the `gowebapp` directory:
```shell
werf run --stages-storage :local --docker-options="-d -p 9000:80 --name gowebapp"  gowebapp -- /app/gowebapp
```

Check that container is running by executing the following command:
```shell
docker ps -f "name=gowebapp"
```

You should see a running container with the `gowebapp` name, like this:
```shell
CONTAINER ID  IMAGE                                          COMMAND           CREATED        STATUS        PORTS                  NAMES
41d6f49798a8  werf-stages-storage/gowebapp:84d7...44992265   "/app/gowebapp"   2 minutes ago  Up 2 minutes  0.0.0.0:9000->80/tcp   gowebapp
```

Open in a web browser the following URL — [http://localhost:9000](http://localhost:9000).

The `Go Web App` page should open. Click "Click here to login" to a login page where you can create a new account and login to the application.

### Getting the image size

Determine the image size by executing:

{% raw %}
```shell
docker images `docker ps -f "name=gowebapp" --format='{{.Image}}'`
```
{% endraw %}

The output will be something like this:
```shell
REPOSITORY                 TAG                   IMAGE ID          CREATED             SIZE
werf-stages-storage/gowebapp   84d7...44992265   07cdc430e1c8      10 minutes ago      857MB
```

Pay attention, that the image size of the application is **above 800 MB**.

## Optimize sample application with artifacts

The config above can be optimized to improve the efficiency of the build process.

The only the files in the `/app` folder are needed to run the application. So we don't need Go itself and downloaded packages. The use of [werf artifacts]({{ site.baseurl }}/configuration/stapel_artifact.html) makes it possible to import only specified files into another image.

### Building

Replace `werf.yaml` with the following content:

{% raw %}
```yaml
project: gowebapp
configVersion: 1
---

artifact: gowebapp-build
from: golang:1.14
ansible:
  install:
  - name: Getting packages
    shell: go get github.com/josephspurrier/gowebapp
  setup:
  - file:
      path: /app
      state: directory
  - name: Copying config
    shell: |
      cp -r $GOPATH/src/github.com/josephspurrier/gowebapp/config /app/config
      cp -r $GOPATH/src/github.com/josephspurrier/gowebapp/static /app/static
      cp -r $GOPATH/src/github.com/josephspurrier/gowebapp/template /app/template
      cp $GOPATH/bin/gowebapp /app/
---
image: gowebapp
docker:
  WORKDIR: /app
from: ubuntu:18.04
import:
- artifact: gowebapp-build
  add: /app
  to: /app
  after: install
```
{% endraw %}

In the optimized config, we build the application in the `gowebapp-build` artifact and import the `/app` directory into the `gowebapp` image.

Pay attention, that `gowebapp` image based on the `ubuntu` image, but not on the `golang` image.

Build the application with the modified config:
```yaml
werf build --stages-storage :local
```

### Running

Before running the modified application, you need to stop and remove running `gowebapp` container we built. Otherwise, the new container can't start or bind to 9000 port on localhost. E.g., execute the following command to stop and remove the `gowebapp` container:

```shell
docker rm -f gowebapp
```

Run the modified application by executing the following command:
```shell
werf run --stages-storage :local --docker-options="-d -p 9000:80 --name gowebapp" gowebapp -- /app/gowebapp
```

Check that container is running by executing the following command:
```shell
docker ps -f "name=gowebapp"
```

You should see a running container with the `gowebapp` name, like this:
```shell
CONTAINER ID  IMAGE                                          COMMAND          CREATED        STATUS        PORTS                   NAMES
41d6f49798a8  werf-stages-storage/gowebapp:84d7...44992265   "/app/gowebapp"  2 minutes ago  Up 2 minutes  0.0.0.0:9000->80/tcp   gowebapp
```

Open in a web browser the following URL — [http://localhost:9000](http://localhost:9000).

The `Go Web App` page should open. Click "Click here to login" to a login page where you can create a new account and login to the application.

### Getting images size

Determine the image size of optimized build, by executing:
{% raw %}
```shell
docker images `docker ps -f "name=gowebapp" --format='{{.Image}}'`
```
{% endraw %}

The output will be something like this:
```shell
REPOSITORY                     TAG               IMAGE ID       CREATED          SIZE
werf-stages-storage/gowebapp   84d7...44992265   07cdc430e1c8   10 minutes ago   79MB
```

Our example shows that **with using artifacts**, the image size **smaller by more than 90%** than the original image size!

## Conclusions

The example shows us that using artifacts is a great way to exclude what shouldn't be in the result image. Moreover, you can use artifacts in any image described in a `werf.yaml` config. In some cases, it increases the speed of build.
