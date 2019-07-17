---
title: Publish process
sidebar: documentation
permalink: documentation/reference/publish_process.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

<!--Docker images should be pushed into the docker registry for further usage in most cases. The usage includes these demands:-->

<!--1. Using an image to run an application (for example in kubernetes). These images will be referred to as **images for running**.-->
<!--2. Using an existing old image version from a docker registry as a cache to build a new image version. Usually, it is default behavior. However, some additional actions may be required to organize a build environment with multiple build hosts or build hosts with no persistent local storage. These images will be referred to as **distributed images cache**.-->

<!--## What can be published-->

<!--The result of werf [build commands]({{ site.baseurl }}/documentation/cli/build/build.html) is a _stages_ in _stages storage_ related to images defined in the `werf.yaml` config. -->
<!--Werf can be used to publish either:-->

<!--* Images. These can only be used as _images for running_. -->
<!--These images are not suitable for _distributed images cache_, because werf build algorithm implies creating separate images for _stages_. -->
<!--When you pull a image from a docker registry, you do not receive _stages_ for this image.-->
<!--* Images with a stages cache images. These images can be used as _images for running_ and also as a _distributed images cache_.-->

<!--Werf pushes image into a docker registry with a so-called [**image publish procedure**](#image-publish-procedure). Also, werf pushes stages cache of all images from config with a so-called [**stages publish procedure**](#stages-publish-procedure).-->

<!--Before digging into these algorithms, it is helpful to see how to publish images using Docker.-->

<!--### Standard publish procedure-->

Normally in the Docker world publish process consists of the following steps:

```bash
docker tag REPO:TAG
docker push REPO:TAG
docker rmi REPO:TAG
```

 1. Get a name or an id of the local built image.
 2. Create new temporary image name alias for this image, which consists of two parts:
     - [docker repository name](https://docs.docker.com/glossary/?term=repository) with embedded docker registry address;
     - [docker tag name](https://docs.docker.com/glossary/?term=tag).
 3. Push image by newly created alias into docker registry.
 4. Delete temporary image name alias.

To publish an image from the config Werf implements another logic:

1. Create **a new image** based on built image with the specified name, store internal service information about tagging schema in this image (using docker labels). This information is referred to as image **meta-information**. Werf uses this information in [deploying]({{ site.baseurl }}/documentation/reference/deploy/deploy_to_kubernetes.html#deploy-command) and [cleaning]({{ site.baseurl }}/documentation/reference/cleanup_process.html) processes.
2. Push newly created image into docker registry.
3. Delete temporary image created in the 1'st step.

All of these steps are performed with a single werf publish command, which will be described below.

The result of this procedure is an image named by the image naming rules pushed into the docker registry.

<!--### Stages publish procedure-->

<!--To publish stages cache of a image from the config werf implements the **stages publish procedure**. It consists of the following steps:-->

<!-- 1. Create temporary image names aliases for all docker images in stages cache, so that:-->
<!--     - [docker repository name](https://docs.docker.com/glossary/?term=repository) is a `REPO` parameter specified by the user without changes ([details about `REPO`]({{ site.baseurl }}/documentation/reference/registry/image_naming.html#repo-parameter)).-->
<!--     - [docker tag name](https://docs.docker.com/glossary/?term=tag) constructed as a signature prefixed with a word `image-stage-` (for example `image-stage-41772c141b158349804ad27b354247df8984ead077a5dd601f3940536ebe9a11`).-->
<!-- 2. Push images by newly created aliases into docker registry.-->
<!-- 3. Delete temporary image names aliases.-->

<!--All of these steps are also performed with a single werf command, which will be described below.-->

<!--The result of this procedure is multiple images from stages cache of image pushed into the docker registry.-->

## Images repo

The _images repo_ is Docker Repo to store images, can be specified by `--images-repo` option or `$WERF_IMAGES_REPO`.

Using _images repo_ werf constructs a [docker repository](https://docs.docker.com/glossary/?term=repository) as follows:

* If werf project contains nameless image, werf uses _images repo_ as docker repository.
* Otherwise, werf constructs docker repository name for each image by following template `IMAGES_REPO/IMAGE_NAME`.

E.g., if there is unnamed image in a `werf.yaml` config and _images repo_ is `myregistry.myorg.com/sys/backend` then the docker repository name is the `myregistry.myorg.com/sys/backend`.  If there are two images in a config — `server` and `worker`, then docker repository names are:
* `myregistry.myorg.com/sys/backend/server` for `server` image;
* `myregistry.myorg.com/sys/backend/worker` for `worker` image.

## Image tag parameters

| option                    | description                          |
| ------------------------- | ------------------------------------ |
| --tag-git-tag TAG         | tag with a slugified TAG             |
| --tag-git-branch BRANCH   | tag with a slugified BRANCH          |
| --tag-git-commit COMMIT   | tag with a slugified COMMIT                               |
| --tag-custom TAG          | arbitrary TAG                        |

Using tag options a user specifies not only tag value but also tagging strategy.
Tagging strategy affects on [certain policies in cleanup process]({{ site.baseurl }}/documentation/reference/cleanup_process.html#cleanup-policies).

Besides, werf applies [tag slug]({{ site.baseurl }}/documentation/reference/toolbox/slug.html#basic-algorithm) transformation rules on all options except `--tag-custom`.
Thus, these tags satisfy the slug requirements, docker tag format, and a user can use arbitrary git tag and branch name formats.

### Combining parameters

Any combination of tag parameters can be used simultaneously for [publish commands]({{ site.baseurl }}/documentation/reference/publish_process.html). In the result, there is a separate image for each tag parameter of each image in a project.

## Examples

### Two images

Given config with 2 images — backend and frontend.

The following command:

```bash
werf publish --stages-storage :local --images-repo registry.hello.com/web/core/system --tag-custom v1.2.0
```

produces the following image names respectively:
* `registry.hello.com/web/core/system/backend:v1.2.0`;
* `registry.hello.com/web/core/system/frontend:v1.2.0`.

### Two images in GitLab job

Given `werf.yaml` config with 2 images — backend and frontend.

The following command runs in GitLab job for git branch named `core/feature/ADD_SETTINGS`:
```bash
type werf && source <(werf ci-env gitlab --tagging-strategy tag-or-branch --verbose)
werf publish --stages-storage :local
```

Image names in the result are:
* `registry.hello.com/web/core/system/backend:core-feature-add-settings-c3fd80df`
* `registry.hello.com/web/core/system/frontend:core-feature-add-settings-c3fd80df`

Each image name converts according to [tag slug]({{ site.baseurl }}/documentation/reference/toolbox/slug.html#basic-algorithm) rules with appended murmurhash of original string.

### Unnamed image in GitLab job

Given config with single unnamed image. The following command runs in GitLab job for git-tag named `v2.3.1`:

```bash
type werf && source <(werf ci-env gitlab --tagging-strategy tag-or-branch --verbose)
werf publish --stages-storage :local
```

Image name in the result is `registry.hello.com/web/core/queue:v2-3-1-5cb8b0a4`

Image name converts according to [tag slug]({{ site.baseurl }}/documentation/reference/toolbox/slug.html#basic-algorithm) rules with appended murmurhash of original string, because of points symbols in the tag `v2.3.1` (points don't meet the requirements).

### Two images with multiple tags in GitLab job

Given config with 2 images — backend and frontend. The following command runs in GitLab job for git-branch named `rework-cache`:

```bash
type werf && source <(werf ci-env gitlab --tagging-strategy tag-or-branch --verbose)
werf publish --stages-storage :local --tag-custom feature-using-cache --tag-custom  my-test-branch
```

The command produces 6 image names for each image name and each tag-parameter (two images by three tags):
* `registry.hello.com/web/core/system/backend:rework-cache`
* `registry.hello.com/web/core/system/frontend:rework-cache`
* `registry.hello.com/web/core/system/backend:feature-using-cache`
* `registry.hello.com/web/core/system/frontend:feature-using-cache`
* `registry.hello.com/web/core/system/backend:my-test-branch`
* `registry.hello.com/web/core/system/frontend:my-test-branch`

## Publish command

{% include /cli/werf_publish.md %}

## Build and publish command

{% include /cli/werf_build_and_publish.md %}
