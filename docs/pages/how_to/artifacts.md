---
title: Using artifacts
sidebar: how_to
permalink: how_to/artifacts.html
author: Artem Kladov <artem.kladov@flant.com>
---

## Task Overview

When you build an application image, it is often necessary to download temporary files or packages for build. In the results, the application image contains files that are not needed to run the application.

Werf can [import]({{ site.baseurl }}/reference/build/import_directive.html) resources from images and [artifacts]({{ site.baseurl }}/reference/build/artifact.html). Thus you can isolate build process and tools in other images and then copy result files to reduce the image size. It is like a docker [multi-stage builds](https://docs.docker.com/develop/develop-images/multistage-build/) which are supported starting with Docker 17.05, but has more advanced files importing options.

In this article, we will build an example GO application. Then we will optimize the build instructions to substantial reduce image size with using mount directives.

## Requirements

* Installed [multiwerf](https://github.com/flant/multiwerf) on the host system.

### Select werf version

This command should be run prior running any werf command in your shell session:

```
source <(multiwerf use 1.0 beta)
```

## Sample application

The example application is the [Hotel Booking Example](https://github.com/revel/examples/tree/master/booking), written in [GO](https://golang.org/) for [Revel Framework](https://github.com/revel).

### Building

Create a `booking` directory and place the following `werf.yaml` in the `booking` directory:
{% raw %}
```yaml
project: hotel-booking
configVersion: 1
---

image: go-booking
from: golang
ansible:
  beforeInstall:
  - name: Install additional packages
    apt:
      name: "{{`{{ item }}`}}"
      update_cache: yes
    with_items:
    - gcc
    - sqlite3
    - libsqlite3-dev
  install:
  - name: Getting packages
    shell: |
      go get -v github.com/revel/revel
      go get -v github.com/revel/cmd/revel
      (go get -v github.com/revel/examples/booking/... ; true )
  setup:
  - name: Preparing config and building application
    shell: |
      sed -i 's/^http.addr=$/http.addr=0.0.0.0/' $GOPATH/src/github.com/revel/examples/booking/conf/app.conf
      revel build --run-mode dev github.com/revel/examples/booking /app
```
{% endraw %}

The config describes instructions to build one image — `go-booking`.

Build the application by executing the following command in the `booking` directory:

```bash
werf build --stages-storage :local
```

### Running

Run the application by executing the following command in the `booking` directory:
```bash
werf --stages-storage :local --docker-options="-d -p 9000:9000 --rm --name go-booking" run go-booking -- /app/run.sh
```

Check that container is running by executing the following command:
```bash
docker ps -f "name=go-booking"
```

You should see a running container with a random name, like this:
```bash
CONTAINER ID  IMAGE                                          COMMAND        CREATED        STATUS        PORTS                   NAMES
41d6f49798a8  image-stage-hotel-booking:f27efaf9...1456b0b4  "/app/run.sh"  3 minutes ago  Up 3 minutes  0.0.0.0:9000->9000/tcp  go-booking
```

Open in a web browser the following URL — [http://localhost:9000](http://localhost:9000).

The `revel framework booking demo` page should open, and you can login by entering `demo/demo` as a login/password.

### Getting the image size

Determine the image size by executing:

{% raw %}
```bash
docker images `docker ps -f "name=go-booking" --format='{{.Image}}'`
```
{% endraw %}

The output will be something like this:
```bash
REPOSITORY                 TAG                   IMAGE ID          CREATED             SIZE
image-stage-hotel-booking  f27efaf9...1456b0b4   0bf71cb34076      10 minutes ago      1.04 GB
```

Pay attention, that the image size of the application is **above 1 GB**.

## Optimize sample application with artifacts

The config above can be optimized to improve the efficiency of the build process.

The only the files in the `/app` folder are needed to run the application. So we don't need Go itself and downloaded packages. The use of [werf artifacts]({{ site.baseurl }}/reference/build/artifact.html) makes it possible to import only specified files into another image.

### Building

Replace `werf.yaml` with the following content:

{% raw %}
```yaml
project: hotel-booking
configVersion: 1
---

artifact: booking-app
from: golang
ansible:
  beforeInstall:
  - name: Install additional packages
    apt:
      name: "{{`{{ item }}`}}"
      update_cache: yes
    with_items:
    - gcc
    - sqlite3
    - libsqlite3-dev
  install:
  - name: Getting packages
    shell: |
      go get -v github.com/revel/revel
      go get -v github.com/revel/cmd/revel
      (go get -v github.com/revel/examples/booking/... ; true )
  setup:
  - name: Preparing config and building application
    shell: |
      sed -i 's/^http.addr=$/http.addr=0.0.0.0/' $GOPATH/src/github.com/revel/examples/booking/conf/app.conf
      revel build --run-mode dev github.com/revel/examples/booking /app
---
image: go-booking
from: ubuntu
import:
- artifact: booking-app
  add: /app
  to: /app
  after: install
```
{% endraw %}

In the optimized config, we build the application in the `booking-app` artifact and import the `/app` directory into the `go-booking` image.

Pay attention, that `go-booking` image based on the ubuntu image, but not on the golang image.

Build the application with the modified config:
```yaml
werf build --stages-storage :local
```

### Running

Before running the modified application, you need to stop running `go-booking` container we built. Otherwise, the new container can't bind to 9000 port on localhost. E.g., execute the following command to stop last created container:

```bash
docker stop go-booking
```

Run the modified application by executing the following command:
```bash
werf --stages-storage :local --docker-options="-d -p 9000:9000 --rm --name go-booking" run go-booking -- /app/run.sh
```

Check that container is running by executing the following command:
```bash
docker ps -f "name=go-booking"
```

You should see a running container with a random name, like this:
```bash
CONTAINER ID  IMAGE                                          COMMAND        CREATED        STATUS        PORTS                   NAMES
41d6f49798a8  image-stage-hotel-booking:306aa6e8...f71dbe53  "/app/run.sh"  3 minutes ago  Up 3 minutes  0.0.0.0:9000->9000/tcp  go-booking
```

Open in a web browser the following URL — [http://localhost:9000](http://localhost:9000).

The `revel framework booking demo` page should open, and you can login by entering `demo/demo` as a login/password.

### Getting images size

Determine the image size of optimized build, by executing:
{% raw %}
```bash
docker images `docker ps -f "name=go-booking" --format='{{.Image}}'`
```
{% endraw %}

The output will be something like this:
```bash
REPOSITORY                   TAG                      IMAGE ID         CREATED            SIZE
image-stage-hotel-booking    306aa6e8...f71dbe53      0a9943b0da6a     3 minutes ago      103 MB
```

Our example shows that **with using artifacts**, the image size **smaller by more than 90%** than the original image size!

## Conclusions

The example shows us that using artifacts is a great way to exclude what shouldn't be in the result image. Moreover, you can use artifacts in any image described in a `werf.yaml` config. In some cases, it increases the speed of build.
