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

## Syntax

```bash
werf dimg tag [options] [DIMG ...] REPO
  --tag TAG
  --tag-branch
  --tag-build-id
  --tag-ci
  --tag-commit
  --tag-plain TAG
  --tag-slug TAG
```

This command supports tag options, described in [image naming]({{ site.baseurl }}/reference/registry/image_naming.html#image-tag-parameters) article.

The `DIMG` optional parameter — is a name of dimg from a config. Specifying `DIMG` one or multiple times allows tagging only certain dimgs from config. By default, werf tags all dimgs.

The `REPO` required parameter — is a repository name (see more in [image naming]({{ site.baseurl }}/reference/registry/image_naming.html#repo-parameter) article).

## Examples

### Tag all dimgs from config

Given config with three dimgs — `backend`, `frontend` and `assets`.

```bash
werf dimg tag registry.hello.com/web/core/system --tag-plain v1.3.0
```

Command creates three local docker images:

* `registry.hello.com/web/core/system/backend:v1.3.0`
* `registry.hello.com/web/core/system/frontend:v1.3.0`
* `registry.hello.com/web/core/system/assets:v1.3.0`

### Tag specified dimgs from config

Given config with three dimgs — `backend`, `frontend` and `assets`.

```bash
werf dimg tag frontend assets registry.hello.com/web/core/system --tag-plain v1.3.0
```

Command creates two local docker images for specified dimgs:

* `registry.hello.com/web/core/system/frontend:v1.3.0`
* `registry.hello.com/web/core/system/assets:v1.3.0`

### Tag single unnamed dimg with multiple docker tags

```bash
werf dimg tag registry.hello.com/tools/archiver --tag-plain v2.9.13 --tag fix/memory-leak
```

Command creates two local docker images:

* `registry.hello.com/tools/archiver:v2.9.13`
* `registry.hello.com/tools/archiver:fix-memory-leak-b3715a6e`
