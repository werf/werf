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

<!--### Stages publish procedure-->

<!--To publish stages cache of a image from the config werf implements the **stages publish procedure**. It consists of the following steps:-->

<!-- 1. Create temporary image names aliases for all docker images in stages cache, so that:-->
<!--     - [docker repository name](https://docs.docker.com/glossary/?term=repository) is a `REPO` parameter specified by the user without changes ([details about `REPO`]({{ site.baseurl }}/documentation/reference/registry/image_naming.html#repo-parameter)).-->
<!--     - [docker tag name](https://docs.docker.com/glossary/?term=tag) constructed as a signature prefixed with a word `image-stage-` (for example `image-stage-41772c141b158349804ad27b354247df8984ead077a5dd601f3940536ebe9a11`).-->
<!-- 2. Push images by newly created aliases into docker registry.-->
<!-- 3. Delete temporary image names aliases.-->

<!--All of these steps are also performed with a single werf command, which will be described below.-->

<!--The result of this procedure is multiple images from stages cache of image pushed into the docker registry.-->

## Images publish procedure

Normally in the Docker world publish process consists of the following steps:

```bash
docker tag REPO:TAG
docker push REPO:TAG
docker rmi REPO:TAG
```

 1. Get a name or an id of the local built image.
 2. Create new temporary image name alias for this image, which consists of two parts:
     - [docker repository name](https://docs.docker.com/glossary/?term=repository) with embedded Docker registry address;
     - [docker tag name](https://docs.docker.com/glossary/?term=tag).
 3. Push image by newly created alias into Docker registry.
 4. Delete temporary image name alias.

To publish an [image]({{ site.baseurl }}/documentation/reference/stages_and_images.html#images) from the config Werf implements another logic:

 1. Create **a new image** based on built image with the specified name, store internal service information about tagging schema in this image (using docker labels). This information is referred to as image **meta-information**. Werf uses this information in [deploy process]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#integration-with-built-images) and [cleaning process]({{ site.baseurl }}/documentation/reference/cleaning_process.html).
 2. Push newly created image into Docker registry.
 3. Delete temporary image created in the 1'st step.

This procedure will be referred as **images publish procedure**.

The result of this procedure is an image named by the [*image naming rules*](#images-naming) pushed into the Docker registry. All of these steps are performed with [werf publish command]({{ site.baseurl }}/documentation/cli/main/publish.html) or [werf build-and-publish command]({{ site.baseurl }}/documentation/cli/main/build_and_publish.html).

## Images naming

During images publish procedure Werf constructs resulting images names using:
 * _images repo_ param;
 * _images repo mode_ param;
 * image name from werf.yaml;
 * tag param.

Resulting docker image name constructed as [`DOCKER_REPOSITORY`](https://docs.docker.com/glossary/?term=repository)`:`[`TAG`](https://docs.docker.com/engine/reference/commandline/tag).

There are _images repo_ and _images repo mode_ params defined where and how to store images.
If werf project contains single nameless image, then _images repo_ used as docker repository without changes and resulting docker image name costructed by the following pattern: `IMAGES_REPO:TAG`.

Otherwise, werf constructs resulting docker image name for each image based on _images repo mode_:  
- `IMAGES_REPO:IMAGE_NAME-TAG` pattern for `monorepo` mode;
- `IMAGES_REPO/IMAGE_NAME:TAG` pattern for `multirepo` mode.

_Images repo_ param should be specified by `--images-repo` option or `$WERF_IMAGES_REPO`.

_Images repo mode_ param should be specified by `--images-repo-mode` option or `$WERF_IMAGES_REPO_MODE`.

> Image naming behavior should be the same for publish, deploy and cleaning processes. Otherwise, the pipeline can be failed and also you can lose images and stages during the cleanup

*Docker tag* is taken from `--tag-*` params:

| option                     | description                                                                     |
| -------------------------- | ------------------------------------------------------------------------------- |
| `--tag-git-tag TAG`        | Use git-tag tagging strategy and tag by the specified git tag                   |
| `--tag-git-branch BRANCH`  | Use git-branch tagging strategy and tag by the specified git branch             |
| `--tag-git-commit COMMIT`  | Use git-commit tagging strategy and tag by the specified git commit hash        |
| `--tag-custom TAG`         | Use custom tagging strategy and tag by the specified arbitrary tag              |

All specified tag params text will be validated to confirm with the docker images tag rules. User may want to apply slug algorithm to the specified tag, see [more info about slug]({{ site.baseurl }}/documentation/reference/toolbox/slug.html).

Also using tag options a user specifies not only tag value but also tagging strategy.
Tagging strategy affects on [certain policies in cleaning process]({{ site.baseurl }}/documentation/reference/cleaning_process.html#cleanup-policies).

Every `--tag-git-*` option requires `TAG`, `BRANCH` or `COMMIT` argument. These options are designed to be compatible with modern CI/CD systems, where CI job is running in the detached git worktree for specific commit and current git-tag or git-branch or git-commit is passed to job using environment variables (for example `CI_COMMIT_TAG`, `CI_COMMIT_REF_NAME` and `CI_COMMIT_SHA` for GitLab CI).

### Combining parameters

Any combination of tag parameters can be used simultaneously in [werf publish command]({{ site.baseurl }}/documentation/cli/main/publish.html) or [werf build-and-publish command]({{ site.baseurl }}/documentation/cli/main/build_and_publish.html). As a result, werf will publish a separate image for each tag parameter of each image in a project.

## Examples

### Two images for git tag

Given `werf.yaml` with 2 images — `backend` and `frontend`.

The following command:

```bash
werf publish --stages-storage :local --images-repo registry.hello.com/web/core/system --tag-git-tag v1.2.0
```

produces the following image names respectively:
 * `registry.hello.com/web/core/system/backend:v1.2.0`;
 * `registry.hello.com/web/core/system/frontend:v1.2.0`.

### Two images for git branch

Given `werf.yaml` with 2 images — `backend` and `frontend`.

The following command:

```bash
werf publish --stages-storage :local --images-repo registry.hello.com/web/core/system --tag-git-branch my-feature-x
```

produces the following image names respectively:
 * `registry.hello.com/web/core/system/backend:my-feature-x`;
 * `registry.hello.com/web/core/system/frontend:my-feature-x`.

### Two images for git branch with special chars in name

Given `werf.yaml` with 2 images — `backend` and `frontend`.

The following command:

```bash
werf publish --stages-storage :local --images-repo registry.hello.com/web/core/system --tag-git-branch $(werf slugify --format docker-tag "Features/MyFeature#169")
```

produces the following image names respectively:
 * `registry.hello.com/web/core/system/backend:features-myfeature169-3167bc8c`;
 * `registry.hello.com/web/core/system/frontend:features-myfeature169-3167bc8c`.

Note that [`werf slugify`]({{ site.baseurl }}/documentation/cli/toolbox/slugify.html) command will generate a valid docker tag. See [more info about slug]({{ site.baseurl }}/documentation/reference/toolbox/slug.html).

### Two images in GitLab CI job

Given `werf.yaml` config with 2 images — `backend` and `frontend`.

The following command runs in Gitlab CI job for project named `web/core/system`, in the git branch named `core/feature/ADD_SETTINGS`, docker registry is configured as `registry.hello.com/web/core/system`:

```bash
type werf && source <(werf ci-env gitlab --tagging-strategy tag-or-branch --verbose)
werf publish --stages-storage :local
```

Image names in the result are:
 * `registry.hello.com/web/core/system/backend:core-feature-add-settings-df80fdc3`;
 * `registry.hello.com/web/core/system/frontend:core-feature-add-settings-df80fdc3`.

Note that werf automatically applies slug to the result docker image tag: `core/feature/ADD_SETTINGS` will be converted to `core-feature-add-settings-df80fdc3`. This convertation occurs in `werf ci-env` command, which will determine git branch from Gitlab CI environment, automatically apply slug to this branch name and set `WERF_TAG_GIT_BRANCH` (which is alternative way to specify `--tag-git-branch` param). See [more info about slug]({{ site.baseurl }}/documentation/reference/toolbox/slug.html).

### Unnamed image in GitLab CI job

Given werf.yaml with single unnamed image. The following command runs in GitLab CI job for project named `web/core/queue`, in the git-tag named `v2.3.1`, docker registry is configured as `registry.hello.com/web/core/queue`:

```bash
type werf && source <(werf ci-env gitlab --tagging-strategy tag-or-branch --verbose)
werf publish --stages-storage :local
```

Image name in the result is `registry.hello.com/web/core/queue:v2.3.1`.
