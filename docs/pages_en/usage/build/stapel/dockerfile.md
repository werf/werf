---
title: Adding docker instructions
permalink: usage/build/stapel/dockerfile.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
directive_summary: docker
---

[Dockerfile instructions](https://docs.docker.com/engine/reference/builder/) can be divided into two groups: build-time instructions and other instructions affecting an image manifest. Build-time instructions do not make sense in a werf build process. Thus, werf supports only following instructions:

* `USER` to set the user name (or UID) and optionally the user group (or GID) (read more [here](https://docs.docker.com/engine/reference/builder/#user)).
* `WORKDIR` to set the working directory (read more [here](https://docs.docker.com/engine/reference/builder/#workdir)).
* `VOLUME` to add a mount point (read more [here](https://docs.docker.com/engine/reference/builder/#volume)).
* `ENV` to set the environment variable (read more [here](https://docs.docker.com/engine/reference/builder/#env)).
* `LABEL` to add metadata to an image (read more [here](https://docs.docker.com/engine/reference/builder/#label)).
* `EXPOSE` to inform Docker that the container listens on the specified network ports at runtime (read more [here](https://docs.docker.com/engine/reference/builder/#expose))
* `ENTRYPOINT` to configure a container that will run as an executable (read more [here](https://docs.docker.com/engine/reference/builder/#entrypoint)). How the Stapel builder processes CMD and ENTRYPOINT is covered [here]({{ "usage/build/stapel/base.html#how-the-stapel-builder-processes-cmd-and-entrypoint" | true_relative_url }}).
* `CMD` to provide the default arguments for the `ENTRYPOINT` to configure a container that will run as an executable (read more [here](https://docs.docker.com/engine/reference/builder/#cmd)). How the Stapel builder processes CMD and ENTRYPOINT is covered [here]({{ "usage/build/stapel/base.html#how-the-stapel-builder-processes-cmd-and-entrypoint" | true_relative_url }}).
* `HEALTHCHECK` to tell Docker how to test a container to see if it is still working (read more [here](https://docs.docker.com/engine/reference/builder/#healthcheck))

These instructions can be specified in the `docker` config directive.

Here is an example of using Docker instructions:

```yaml
docker:
  WORKDIR: /app
  CMD: ["python", "./index.py"]
  EXPOSE: '5000'
  ENV:
    TERM: xterm
    LC_ALL: en_US.UTF-8
```

The Docker instructions defined in the configuration are applied during the last stage called `docker_instructions`.
Thus, instructions do not affect other stages but are applied to a built image.

If you need to use specific environment variables at build time (such as a `TERM` environment), you have to use the [base image]({{"usage/build/stapel/base.html" | true_relative_url }}) in which these environment variables are set.

> Tip: you can also export environment variables right to the [_user stage_]({{ "usage/build/stapel/instructions.html#what-are-user-stages" | true_relative_url }}) instructions.
