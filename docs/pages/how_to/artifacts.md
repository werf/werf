---
title: Using artifacts
sidebar: how_to
permalink: how_to/artifacts.html
author: Artem Kladov <artem.kladov@flant.com>
---

## Task Overview

When building an application image, there are often step on it is necessary to download package, build code and so on. In the results, the size of the final image can increase.

Dapp has artifacts, to reduce the size of the final image. It is like a docker [multi-stage builds](https://docs.docker.com/develop/develop-images/multistage-build/) which are supported starting with Docker 17.05, but more convenient (and appeared much earlier).

In this article, we will build an example GO application. Then we will optimize the build instructions to substantial reduce final image size with using mount directives.

## Sample application

The example application is the [Hotel Booking Example](https://github.com/revel/examples/tree/master/booking), written in [GO](https://golang.org/) for [Revel Framework](https://github.com/revel).

### Building

Create a `booking` directory and place the following `dappfile.yaml` in the `booking` directory:
{% raw %}
```
dimg: go-booking
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

The dappfile describes instructions to build one dimg — `go-booking`.

Build the application by executing the following command in the `booking` directory:

```
dapp dimg build
```

### Running

Run the application by executing the following command in the `booking` directory:
```
dapp dimg run -p 9000:9000 --rm -d -- /app/run.sh
```

Check that container is running by executing the following command:
```
docker ps
```

You should see a running container with a random name, like this:
```
CONTAINER ID  IMAGE         COMMAND        CREATED        STATUS        PORTS                   NAMES
41d6f49798a8  14e6b9c6b93b  "/app/run.sh"  3 minutes ago  Up 3 minutes  0.0.0.0:9000->9000/tcp  infallible_bell
```

Open in a web browser the following URL — [http://localhost:9000](http://localhost:9000).

The `revel framework booking demo` page should open, and you can login by entering `demo/demo` as a login/password.

### Getting the image size

Create a final image with tag `v1.0`:

```
dapp dimg tag booking --tag-plain v1.0
```

After tagging we get an image `booking/go-booking:v1.0` according to dapp naming rules (read more about naming [here]({{ site.baseurl }}/reference/registry/image_naming.html)).

Determine the image size by executing:

```
docker images booking/go-booking:v1.0
```

The output will be something like this:
```
REPOSITORY           TAG           IMAGE ID            CREATED             SIZE
booking/go-booking   v1.0          0bf71cb34076        10 minutes ago      1.03 GB
```

Pay attention, that the final image size of the application is **above 1 GB**.

## Optimize sample application with artifacts

The dappfile above can be optimized to improve the efficiency of the build process.

The only the files in the `/app` folder are needed to run the application. So we don't need Go itself and downloaded packages. The use of [dapp artifacts]({{ site.baseurl }}/reference/build/artifact_directive.html) makes it possible to import only specified files into another image.

### Building

Replace `dappfile.yml` with the following content:

{% raw %}
```
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
dimg: go-booking
from: ubuntu
import:
- artifact: booking-app
  add: /app
  to: /app
  after: install
```
{% endraw %}

In the optimized dappfile, we build the application in the `booking-app` artifact and import the `/app` directory into the `go-booking` dimg.

Pay attention, that `go-booking` dimg based on the ubuntu image, but not on the golang image.

Build the application with the modified dappfile:
```
dapp dimg build
```

### Running

Before running the modified application, you need to stop already running container. Otherwise, the new container can't bind to 9000 port on localhost. E.g., execute the following command to stop last created container:

```
docker stop `docker ps -lq`
```

Run the modified application by executing the following command:
```
dapp dimg run -p 9000:9000 --rm -d -- /app/run.sh
```

Check that container is running by executing the following command:
```
docker ps
```

You should see a running container with a random name, like this:
```
CONTAINER ID  IMAGE         COMMAND        CREATED        STATUS        PORTS                   NAMES
88287022813b  c8277cd4a801  "/app/run.sh"  5 seconds ago  Up 3 seconds  0.0.0.0:9000->9000/tcp  naughty_dubinsky
```

Open in a web browser the following URL — [http://localhost:9000](http://localhost:9000).

The `revel framework booking demo` page should open, and you can login by entering `demo/demo` as a login/password.

### Getting images size

Create a final image with tag `v2.0`:

```
dapp dimg tag booking --tag-plain v2.0
```

Determine the final image size of optimized build, by executing:
```
docker images booking/go-booking
```

The output will be something like this:
```
REPOSITORY            TAG        IMAGE ID         CREATED            SIZE
booking/go-booking    v2.0      0a9943b0da6a     3 minutes ago      103 MB
booking/go-booking    v1.0      0bf71cb34076     15 minutes ago     1.04 GB
```

The total size difference between `v1.0` and `v2.0` images is about 900 MB.

Our example shows that **with using artifacts**, the final image size **smaller by more than 90%** than the original image size!

## Conclusions

The example shows us that using artifacts is a great way to exclude what shouldn't be in the final image. Moreover, you can use artifacts in any dimg described in a dappfile. In some cases, it increases the speed of build.
