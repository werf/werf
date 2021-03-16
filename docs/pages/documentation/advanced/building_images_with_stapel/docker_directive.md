---
title: Adding docker instructions
permalink: documentation/advanced/building_images_with_stapel/docker_directive.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
directive_summary: docker
---

[Dockerfile instructions](https://docs.docker.com/engine/reference/builder/) can be divided into two groups: build-time instructions and other instructions that effect on an image manifest. Build-time instructions do not make sense in a werf build process. Therefore, werf supports only following instructions:

* `USER` to set the user name (or UID) and optionally the user group (or GID) (read more [here](https://docs.docker.com/engine/reference/builder/#user)).
* `WORKDIR` to set the working directory (read more [here](https://docs.docker.com/engine/reference/builder/#workdir)).
* `VOLUME` to add mount point (read more [here](https://docs.docker.com/engine/reference/builder/#volume)).
* `ENV` to set the environment variable (read more [here](https://docs.docker.com/engine/reference/builder/#env)).
* `LABEL` to add metadata to an image (read more [here](https://docs.docker.com/engine/reference/builder/#label)).
* `EXPOSE` to inform Docker that the container listens on the specified network ports at runtime (read more [here](https://docs.docker.com/engine/reference/builder/#expose))
* `ENTRYPOINT` to configure a container that will run as an executable (read more [here](https://docs.docker.com/engine/reference/builder/#entrypoint)). How stapel builder processes CMD and ENTRYPOINT is covered [here]({{ "documentation/internals/build_process.html#how-the-stapel-builder-processes-cmd-and-entrypoint" | true_relative_url }}).
* `CMD` to provide default arguments for the `ENTRYPOINT` to configure a container that will run as an executable (read more [here](https://docs.docker.com/engine/reference/builder/#cmd)). How stapel builder processes CMD and ENTRYPOINT is covered [here]({{ "documentation/internals/build_process.html#how-the-stapel-builder-processes-cmd-and-entrypoint" | true_relative_url }}).
* `HEALTHCHECK` to tell Docker how to test a container to check that it is still working (read more [here](https://docs.docker.com/engine/reference/builder/#healthcheck))

These instructions can be specified in the `docker` config directive.

Here is an example of using docker instructions:

```yaml
docker:
  WORKDIR: /app
  CMD: "['python', './index.py']"
  EXPOSE: '5000'
  ENV:
    TERM: xterm
    LC_ALL: en_US.UTF-8
```

Defined docker instructions are applied on the last stage called `docker_instructions`.
Thus, instructions do not affect other stages, ones just will be applied to a built image.

If need to use special environment variables in build-time of your application image, such as `TERM` environment, you should use a [base image]({{ "documentation/advanced/building_images_with_stapel/base_image.html" | true_relative_url }}) with these variables.

> Tip: you can also implement exporting environment variables right in [_user stage_]({{ "documentation/advanced/building_images_with_stapel/assembly_instructions.html#what-are-user-stages" | true_relative_url }}) instructions
