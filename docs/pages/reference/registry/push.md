---
title: Push
sidebar: reference
permalink: reference/registry/push.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

Docker images should be pushed into the docker registry for further usage in most cases. The usage includes these demands:

1. Using an image to run an application (for example in kubernetes). These images will be referred to as **images for running**.
2. Using an existing old image version from a docker registry as a cache to build a new image version. Usually, it is default behavior. However, some additional actions may be required to organize a build environment with multiple build hosts or build hosts with no persistent local storage. These images will be referred to as **distributed images cache**.

## What can be pushed

The result of werf [build commands]({{ site.baseurl }}/reference/cli/dimg_build.html) is a stages cache related to dimgs defined in the config. Werf can be used to push either:

* Dimgs images. These can only be used as _images for running_. These images are not suitable for _distributes images cache_, because werf build algorithm implies creating separate images for stages cache. When you pull a dimg from a docker registry, you don't receive stages cache for this image.
* Dimgs images with a stages cache images. These images can be used as _images for running_ and also as a _distributed images cache_.

Werf pushes dimg into a docker registry with a so-called [**dimg push procedure**](#dimg-push-procedure). Also, werf pushes stages cache of all dimgs from config with a so-called [**stages push procedure**](#stages-push-procedure).

Before digging into these algorithms, it is helpful to see how to push images using Docker.

### Standard push procedure

Normally in the Docker world to push an already built arbitrary docker image, the following steps are required in general:

 1. Get a name or an id of the local built image.
 2. Create new temporary image name alias for this image, which consists of two parts:
     - [docker repository name](https://docs.docker.com/glossary/?term=repository) with embedded docker registry address;
     - [docker tag name](https://docs.docker.com/glossary/?term=tag).
 3. Push image by newly created alias into docker registry.
 4. Delete temporary image name alias.

This process will be referred to as **standard push procedure**. There is a docker command for each of these steps, and usually, they are performed by calling corresponding docker commands.

### Dimg push procedure

To push a dimg from the config werf implements the **dimg push procedure**. It consists of the following steps:

1. Perform [**werf tag procedure**]({{ site.baseurl }}/reference/registry/image_naming.html#werf-tag-procedure) for built dimg. The result of werf tag is an image with a name that is compatible with the step 2 of _standard push procedure_. I.e., this image is ready to be pushed.
2. Push newly created image into docker registry.
3. Delete temporary image created in the 1'st step.

All of these steps are performed with a single werf push command, which will be described below.

The result of this procedure is a dimg named by the [image naming]({{ site.baseurl }}/reference/registry/image_naming.html) rules pushed into the docker registry.

### Stages push procedure

To push stages cache of a dimg from the config werf implements the **stages push procedure**. It consists of the following steps:

 1. Create temporary image names aliases for all docker images in stages cache, so that:
     - [docker repository name](https://docs.docker.com/glossary/?term=repository) is a `REPO` parameter specified by the user without changes ([details about `REPO`]({{ site.baseurl }}/reference/registry/image_naming.html#repo-parameter)).
     - [docker tag name](https://docs.docker.com/glossary/?term=tag) constructed as a signature prefixed with a word `dimgstage-` (for example `dimgstage-41772c141b158349804ad27b354247df8984ead077a5dd601f3940536ebe9a11`).
 2. Push images by newly created aliases into docker registry.
 3. Delete temporary image names aliases.

All of these steps are also performed with a single werf command, which will be described below.

The result of this procedure is multiple images from stages cache of dimg pushed into the docker registry.

## Werf push

Werf push command used to perform _dimg push procedure_ and optionally _stages push procedure_. Dimgs should be built with werf [build commands]({{ site.baseurl }}/reference/cli/dimg_build.html) before calling this command.

### Syntax

```bash
werf dimg push [options] [DIMG ...] REPO
  --with-stages
  --tag TAG
  --tag-branch
  --tag-build-id
  --tag-ci
  --tag-commit
  --tag-plain TAG
  --tag-slug TAG
```

This command supports tag options, described in [image naming]({{ site.baseurl }}/reference/registry/image_naming.html#image-tag-parameters) article.

The `DIMG` optional parameter — is a name of dimg from a config. Specifying `DIMG` one or multiple times allows pushing only certain dimgs from config. By default, werf pushes all dimgs.

The `REPO` required parameter — is a repository name (see more in [image naming]({{ site.baseurl }}/reference/registry/image_naming.html#repo-parameter) article).

The option `--with-stages` specifies werf to push not only a dimg but also a stages cache for the corresponding dimg. Optional `DIMG` parameter also affects stages cache that will be pushed.

## Werf bp

Werf bp stands for _build and push_. It is a command that combines the [build command]({{ site.baseurl }}/reference/cli/dimg_build.html) with a [push command](#werf-push).

It has an execution speed advantage over using separate build and push commands, because there are common calculation steps, that are performed only once.

### Syntax

```bash
werf dimg bp [options] [DIMG ...] REPO
  --with-stages
  --tag TAG
  --tag-branch
  --tag-build-id
  --tag-ci
  --tag-commit
  --tag-plain TAG
  --tag-slug TAG
```

This command supports tag options, described in [image naming]({{ site.baseurl }}/reference/registry/image_naming.html#image-tag-parameters) article.

The `DIMG` optional parameter — is a name of dimg from a config. Specifying `DIMG` one or multiple times allows building and pushing only certain dimgs from config. By default, werf pushes all dimgs.

The `REPO` required parameter — is a repository name (see more in [image naming]({{ site.baseurl }}/reference/registry/image_naming.html#repo-parameter) article).

The option `--with-stages` specifies werf to push not only a dimg but also a stages cache for the corresponding dimg. Optional `DIMG` parameter also affects stages cache that will be pushed.

## Werf spush

Werf spush means _simple push_. It is a command that used to ignore [docker repository](https://docs.docker.com/glossary/?term=repository) naming rules, which described in the [image naming article]({{ site.baseurl }}/reference/registry/image_naming.html) and push docker image with the fully customized name. Unlike regular push, a name of the docker image is not related to the name of the dimg from the config.

The result of calling this command is always one or more docker images with a custom docker repository names, created for a single dimg.

Dimgs should be built with werf [build commands] before calling this command.

### Syntax

```bash
werf dimg spush [options] [DIMG] REPO
  --tag TAG
  --tag-branch
  --tag-build-id
  --tag-ci
  --tag-commit
  --tag-plain TAG
  --tag-slug TAG
```

This command supports tag options, described in [image naming]({{ site.baseurl }}/reference/registry/image_naming.html#image-tag-parameters) article. Multiple tags can be specified.

The `REPO` required parameter — is a final [docker repository name](https://docs.docker.com/glossary/?term=repository) used AS-IS, unlike `REPO` parameter from [image naming article]({{ site.baseurl }}/reference/registry/image_naming.html). `REPO` should be unique for each dimg from the config.

The `DIMG` optional parameter — is a name of dimg from a config.
* Parameter should be omitted only in the case when there is only single unnamed dimg in the config.
* Parameter should be specified in the case when there are several named dimgs in the config. Only one dimg name can be used with this command at a time.

`DIMG` and `REPO` are specified for each dimg individually.

## Examples

### Push all dimgs from config

Given config with three dimgs — `core`, `queue` and `api`.

```bash
werf dimg push registry.hello.com/taxi/backend --tag-plain v1.1.14
```

Command pushes three images into specified repo:

* `registry.hello.com/taxi/backend/core:v1.1.14`
* `registry.hello.com/taxi/backend/queue:v1.1.14`
* `registry.hello.com/taxi/backend/api:v1.1.14`

### Build and push all dimgs from config with stages cache

Given werf project named `geo` with config, that contains two dimgs — `server`, `geodata`.

```bash
werf dimg bp registry.hello.com/api/geo --tag stable --with-stages
```

Command builds and pushes following images into specified repo:

* `registry.hello.com/api/geo/server:stable`
* `registry.hello.com/api/geo/geodata:stable`
* `registry.hello.com/api/geo:dimgstage-ab192db1f7cf6b894aeaf14c0f1615f27d5170bb16b8529ec18253b94dc4916e`
* `registry.hello.com/api/geo:dimgstage-0d05ad73cf69430e8ff2bf457d6fd6273ec100785fcc3ad0267c0fb609c81a7c`
* `registry.hello.com/api/geo:dimgstage-41772c141b158349804ad27b354247df8984ead077a5dd601f3940536ebe9a11`
* `registry.hello.com/api/geo:dimgstage-1b13ae518f500a3df486bbfb90e00bddfc2fe328c9d7cd6347e4e3cfc4435d05`

Notice the images with the `dimgstage` prefixes in the docker tags. These are the stages cache, that was pushed along with a dimgs.

### Push specified dimgs from config

Given config with three dimgs — `kubetools`, `backuptools` and `copytools`.

```bash
werf dimg push kubetools backuptools registry.hello.com/tools/common
```

Command pushes following images into specified repo for each of the specified dimgs `kubetools` and `backuptools`:

* `registry.hello.com/tools/common/kubetools:latest`
* `registry.hello.com/tools/common/backuptools:latest`

### Push single dimg using custom docker repository name

Given config with two dimgs — `plain-c` and `rust`.

```bash
werf dimg spush plain-c myregistry.host/lingvo/parser --tag stable --tag-plain v1.4.11
```

Command pushes image for dimg `plain-c` with two docker tags `stable` and `v1.4.11`:

* `myregistry.host/tools/parser:stable`
* `myregistry.host/tools/parser:v1.4.11`

Notice that the name of the resulting docker images is not related to the dimg name, unlike regular push command.
