---
title: Adding docker instructions
sidebar: reference
permalink: reference/build/docker_directive.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
---

Docker can build images by [Dockerfile](https://docs.docker.com/engine/reference/builder/) instructions. These instructions can be divided into two groups: build-time instructions and other instructions that effect on a running built image and a built image in general.  

Build-time instructions don't make sense in a dapp build process, therefore, dapp supports only following instructions:

* `USER` to set the user and the group to use when running the image (read more [here](https://docs.docker.com/engine/reference/builder/#user)).
* `WORKDIR` to set the working directory (read more [here](https://docs.docker.com/engine/reference/builder/#workdir)).
* `VOLUME` to add mount point (read more [here](https://docs.docker.com/engine/reference/builder/#volume)).
* `ENV` to set the environment variable (read more [here](https://docs.docker.com/engine/reference/builder/#env)).
* `LABEL` to add metadata to an image (read more [here](https://docs.docker.com/engine/reference/builder/#label)).
* `EXPOSE` to inform Docker that the container listens on the specified network ports at runtime (read more [here](https://docs.docker.com/engine/reference/builder/#expose))
* `ENTRYPOINT` to configure a container that will run as an executable (read more [here](https://docs.docker.com/engine/reference/builder/#entrypoint)).
* `CMD` to provide default arguments for the `ENTRYPOINT` to configure a container that will run as an executable (read more [here](https://docs.docker.com/engine/reference/builder/#cmd)).
* `ONBUILD` to add to the image a trigger instruction to be executed at a later time when the image is used as the base for another build (read more [here](https://docs.docker.com/engine/reference/builder/#onbuild)).

These instructions can be defined in the dappfile `docker` section.

Here are the example of using docker instructions:

```yaml
docker:
  WORKDIR: /app
  CMD: ['python', './index.py']
  EXPOSE: '5000'
  ENV:
    TERM: xterm
    LC_ALL: en_US.UTF-8
```

Defined docker instructions are applied on the last stage called `docker_instructions`. Thus instructions don't effect on other stages and a dapp build process in general, they simply will be applied on a final image. 

If need to use special environment variables in build-time of your application image, such as `TERM` environment, you should use a [base image]({{ site.baseurl }}/reference/build/base_image.html) with these variables.

## Syntax

```yaml
docker:
  VOLUME:
  - <volume>
  EXPOSE:
  - <expose>
  ENV:
    <env_name>: <env_value>
  LABEL:
    <label_name>: <label_value>
  ENTRYPOINT:
  - <entrypoint>
  CMD:
  - <cmd>
  ONBUILD:
  - <onbuild>
  WORKDIR: <workdir>
  USER: <user>
```
