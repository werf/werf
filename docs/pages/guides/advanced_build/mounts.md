---
title: Mounts
sidebar: documentation
permalink: guides/advanced_build/mounts.html
author: Artem Kladov <artem.kladov@flant.com>
---

## Task Overview

In this article, we will build an example Go application. Then we will optimize the build instructions to substantial reduce image size with using mount directives.

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
```
project: gowebapp
configVersion: 1
---

{{ $_ := set . "GoDlPath" "https://dl.google.com/go/" }}
{{ $_ := set . "GoTarball" "go1.14.7.linux-amd64.tar.gz" }}
{{ $_ := set . "GoTarballChecksum" "sha256:4a7fa60f323ee1416a4b1425aefc37ea359e9d64df19c326a58953a97ad41ea5" }}
{{ $_ := set . "BaseImage" "ubuntu:18.04" }}

image: gowebapp
from: {{ .BaseImage }}
docker:
  WORKDIR: /app
ansible:
  beforeInstall:
  - name: Install essential utils
    apt:
      name: ['curl','git','tree']
      update_cache: yes
  - name: Download the Go tarball
    get_url:
      url: {{ .GoDlPath }}{{ .GoTarball }}
      dest: /usr/local/src/{{ .GoTarball }}
      checksum:  {{ .GoTarballChecksum }}
  - name: Extract the Go tarball if Go is not yet installed or not the desired version
    unarchive:
      src: /usr/local/src/{{ .GoTarball }}
      dest: /usr/local
      copy: no
  install:
  - name: Getting packages
    shell: |
{{ include "export go vars" . | indent 6 }}
      go get github.com/josephspurrier/gowebapp
  setup:
  - file:
      path: /app
      state: directory
  - name: Copying config
    shell: |
{{ include "export go vars" . | indent 6 }}
      cp -r $GOPATH/src/github.com/josephspurrier/gowebapp/config /app/config
      cp -r $GOPATH/src/github.com/josephspurrier/gowebapp/static /app/static
      cp -r $GOPATH/src/github.com/josephspurrier/gowebapp/template /app/template
      cp $GOPATH/bin/gowebapp /app/

{{- define "export go vars" -}}
export GOPATH=/go
export PATH=$GOPATH/bin:$PATH:/usr/local/go/bin
{{- end -}}
```
{% endraw %}

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

### Getting the application image size

Determine the image size by executing:

{% raw %}
```shell
docker images `docker ps -f "name=gowebapp" --format='{{.Image}}'`
```
{% endraw %}

The output will be something like this:
```shell
REPOSITORY                 TAG                   IMAGE ID          CREATED             SIZE
werf-stages-storage/gowebapp   84d7...44992265   07cdc430e1c8      10 minutes ago      708MB
```

You can check the size of all ancestor images in the output of the `werf build --stages-storage :local` command. After werf built or used any image it outputs some information  like image name and tag, image size or image size difference, used instructions and commands.

Pay attention, that the image size of the application is **above 700 MB**.

## Optimizing

There are often a lot of useless files in the image. In our example application, these are — APT cache and GO sources. Also, after building the application, the GO itself is not needed to run the application and can be removed from the image.

### Optimizing APT cache

[APT](https://wiki.debian.org/Apt) saves the package list in the `/var/lib/apt/lists/` directory and also saves packages in the `/var/cache/apt/` directory when installs them. So, it is useful to store `/var/cache/apt/` outside the image and share it between builds. The `/var/lib/apt/lists/` directory depends on the status of the installed packages, and it's no good to share it between builds, but it is useful to store it outside the image to reduce its size.

To optimize using APT cache add the following directives to the `gowebapp` image in the config:

```yaml
mount:
- from: tmp_dir
  to: /var/lib/apt/lists
- from: build_dir
  to: /var/cache/apt
```

Read more about mount directives [here]({{ site.baseurl }}/configuration/stapel_image/mount_directive.html).

The `/var/lib/apt/lists` directory is filling in the build-time, but in the image, it is empty.

The `/var/cache/apt/` directory is caching in the `~/.werf/shared_context/mounts/projects/gowebapp/var-cache-apt-28143ccf/` directory but in the image, it is empty. Mounts work only during werf assembly process. So, if you change stages instructions and rebuild your project, the `/var/cache/apt/` will already contain packages downloaded earlier.

Official Ubuntu image contains special hooks that remove APT cache after image build. To disable these hooks, add the following task to a ***beforeInstall*** stage of the config:

```yaml
- name: Disable docker hook for apt-cache deletion
  shell: |
    set -e
    sed -i -e "s/DPkg::Post-Invoke.*//" /etc/apt/apt.conf.d/docker-clean
    sed -i -e "s/APT::Update::Post-Invoke.*//" /etc/apt/apt.conf.d/docker-clean
```

### Optimizing builds

In the example application, the GO is downloaded and extracted. The GO source is not needed in the image. After the application is built, the GO itself is also not needed in the image. So mount `/usr/local/src` and `/usr/local/go` directories to place them outside the image.

Building application on the setup stage uses the `/go` directory, specified in the `GOPATH` environment variable. This directory contains necessary packages and application source. After the build, the result is placed in the `/app` directory, and the `/go` directory is not needed to run the application. So, the `/go` directory can be mounted to a temporary place, outside of the image.

Add the following to mount directives into config:

```yaml
- from: tmp_dir
  to: /var/lib/apt/lists
- from: build_dir
  to: /var/cache/apt
- from: tmp_dir
  to: /go
- from: build_dir
  to: /usr/local/src
- from: build_dir
  to: /usr/local/go
```

### Complete werf.yaml config

{% raw %}
```yaml
project: gowebapp
configVersion: 1
---

{{ $_ := set . "GoDlPath" "https://dl.google.com/go/" }}
{{ $_ := set . "GoTarball" "go1.14.7.linux-amd64.tar.gz" }}
{{ $_ := set . "GoTarballChecksum" "sha256:4a7fa60f323ee1416a4b1425aefc37ea359e9d64df19c326a58953a97ad41ea5" }}
{{ $_ := set . "BaseImage" "ubuntu:18.04" }}

image: gowebapp
from: {{ .BaseImage }}
docker:
  WORKDIR: /app
mount:
- from: tmp_dir
  to: /var/lib/apt/lists
- from: build_dir
  to: /var/cache/apt
- from: tmp_dir
  to: /go
- from: build_dir
  to: /usr/local/src
- from: build_dir
  to: /usr/local/go
ansible:
  beforeInstall:
  - name: Disable docker hook for apt-cache deletion
    shell: |
      set -e
      sed -i -e "s/DPkg::Post-Invoke.*//" /etc/apt/apt.conf.d/docker-clean
      sed -i -e "s/APT::Update::Post-Invoke.*//" /etc/apt/apt.conf.d/docker-clean          
  - name: Install essential utils
    apt:
      name: ['curl','git','tree']
      update_cache: yes
  - name: Download the Go tarball
    get_url:
      url: {{ .GoDlPath }}{{ .GoTarball }}
      dest: /usr/local/src/{{ .GoTarball }}
      checksum:  {{ .GoTarballChecksum }}
  - name: Extract the Go tarball if Go is not yet installed or not the desired version
    unarchive:
      src: /usr/local/src/{{ .GoTarball }}
      dest: /usr/local
      copy: no
  install:
  - name: Getting packages
    shell: |
{{ include "export go vars" . | indent 6 }}
      go get github.com/josephspurrier/gowebapp
  setup:
  - file:
      path: /app
      state: directory
  - name: Copying config
    shell: |
{{ include "export go vars" . | indent 6 }}
      cp -r $GOPATH/src/github.com/josephspurrier/gowebapp/config /app/config
      cp -r $GOPATH/src/github.com/josephspurrier/gowebapp/static /app/static
      cp -r $GOPATH/src/github.com/josephspurrier/gowebapp/template /app/template
      cp $GOPATH/bin/gowebapp /app/

{{- define "export go vars" -}}
export GOPATH=/go
export PATH=$GOPATH/bin:$PATH:/usr/local/go/bin
{{- end -}}
```
{% endraw %}

Build the application with the modified config:
```shell
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

You should see a running container with a `gowebapp` name, like this:
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
werf-stages-storage/gowebapp   84d7...44992265   07cdc430e1c8   10 minutes ago   181MB
```

### Analysis

The `~/.werf/shared_context/mounts/projects/gowebapp/` path contains directories mounted with `from: build_dir` directives in the `werf.yaml` file. Execute the following command to analyze:

```shell
tree -L 3 ~/.werf/shared_context/mounts/projects/gowebapp
```

The output will be like this (some lines skipped):
```shell
/home/user/.werf/shared_context/mounts/projects/gowebapp/
├── usr-local-go-a179aaae
│   ├── api
│   ├── lib
│   ├── pkg
...
│   └── src
├── usr-local-src-6ad4baf1
│   └── go1.14.7.linux-amd64.tar.gz
└── var-cache-apt-28143ccf
    └── archives
        ├── ca-certificates_20190110~18.04.1_all.deb
...
        └── xauth_1%3a1.0.10-1_amd64.deb
```

As you may see, there are separate directories on the host for every mount in config exists.

Check the directories size, by executing:
```shell
sudo du -kh --max-depth=1 ~/.werf/shared_context/mounts/projects/gowebapp
```

The output will be like this:
```shell
19M     /home/user/.werf/shared_context/mounts/projects/gowebapp/var-cache-apt-28143ccf
119M    /home/user/.werf/shared_context/mounts/projects/gowebapp/usr-local-src-6ad4baf1
348M    /home/user/.werf/shared_context/mounts/projects/gowebapp/usr-local-go-a179aaae
485M    /home/user/.werf/shared_context/mounts/projects/gowebapp
```

`485M` is a size of files excluded from image, but these files are accessible, in case of rebuild image and also they can be mounted in other images in this project. E.g., if you add image based on Ubuntu, you can mount `/var/cache/apt` with `from: build_dir` and use already downloaded packages.

Also, approximately `70MB` of space occupy files in directories mounted with `from: tmp_dir`. These files also excluded from the image and deleted from the host at the end of image building.

The total size difference between images with and without using mounts is about `527MB` (the result of 708MB — 181MB).

**Our example shows that with using werf mounts the image size smaller by more than 70% than the original image size!**

## What Can Be Improved

* Use a smaller base image instead of ubuntu, such as [alpine](https://hub.docker.com/_/alpine/) or [golang](https://hub.docker.com/_/golang/).
* Using [werf artifacts]({{ site.baseurl }}/configuration/stapel_artifact.html) in many cases can give more efficient.
  The size of `/app` directory in the image is about only 15 MB (you can check it by executing `werf run --stages-storage :local --docker-options="--rm" gowebapp -- du -kh --max-depth=0 /app`). So you can build files into the `/app` in werf artifact and then import only the resulting `/app` directory.
