---
title: Tag
sidebar: reference
permalink: reference/registry/tag.html
author: Artem Kladov <artem.kladov@flant.com>
---

When werf builds dimgs, the result is a build cache — a set of image layers. To create a final image and assign a specified docker tag (see more about image naming [here]({{ site.baseurl }}/reference/registry/image_naming.html)) use `werf dimg tag` command.

You might want to create a named images for built dimgs with `werf dimg tag` command in some cases:

* To run containers from these images locally with command other than `werf dimg run` [command]({{ site.baseurl }}/reference/cli/dimg_run.html). For example with [docker run](https://docs.docker.com/engine/reference/run/) or [docker compose](https://docs.docker.com/compose/overview).
* To push these named images with command other than [werf push commands]({{ site.baseurl }}/reference/registry/push.html). For example with [docker push](https://docs.docker.com/engine/reference/commandline/image_push/).

Build dimgs with [build commands]({{ site.baseurl }}/reference/cli/dimg_build.html) before assigning tags.

Unlike `docker tag` command, which simply creates an alias for the specified image, `werf dimg tag` produces a new image layer with a specified name, see [image naming article]({{ site.baseurl }}/reference/registry/image_naming.html) for the details. The result of the executing `werf dimg tag` command is docker images created locally.

## werf tag

{% include /cli/werf_tag.md %}
