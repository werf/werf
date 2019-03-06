---
title: Tag
sidebar: reference
permalink: reference/registry/tag.html
author: Artem Kladov <artem.kladov@flant.com>
---

When werf builds images, the result is a build cache — a set of image layers. To create a final image and assign a specified docker tag (see more about image naming [here]({{ site.baseurl }}/reference/registry/image_naming.html)) use `werf tag` command.

You might want to create a named images for built images with `werf tag` command to push these named images with command other than [werf push command]({{ site.baseurl }}/reference/registry/push.html). For example with [docker push](https://docs.docker.com/engine/reference/commandline/image_push/).

Build images with [build commands]({{ site.baseurl }}/cli/main/build.html) before assigning tags.

Unlike `docker tag` command, which simply creates an alias for the specified image, `werf tag` produces a new image layer with a specified name, see [image naming article]({{ site.baseurl }}/reference/registry/image_naming.html) for the details. The result of the executing `werf tag` command is docker images created locally.

## Tag command

<!--  TODO -->
