---
title: Publish process
sidebar: documentation
permalink: documentation/reference/publish_process.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

<!--Docker images should be pushed into the Docker registry for further usage in most cases. The usage includes these demands:-->

<!--1. Using an image to run an application (for example in Kubernetes). These images will be referred to as **images for running**.-->
<!--2. Using an existing old image version from a Docker registry as a cache to build a new image version. Usually, it is default behavior. However, some additional actions may be required to organize a build environment with multiple build hosts or build hosts with no persistent local storage. These images will be referred to as **distributed images cache**.-->

<!--## What can be published-->

<!--The result of werf [build commands]({{ site.baseurl }}/documentation/cli/build/build.html) is a _stages_ in _stages storage_ related to images defined in the `werf.yaml` config. -->
<!--werf can be used to publish either:-->

<!--* Images. These can only be used as _images for running_. -->
<!--These images are not suitable for _distributed images cache_, because werf build algorithm implies creating separate images for _stages_. -->
<!--When you pull a image from a Docker registry, you do not receive _stages_ for this image.-->
<!--* Images with a stages cache images. These images can be used as _images for running_ and also as a _distributed images cache_.-->

<!--werf pushes image into a Docker registry with a so-called [**image publish procedure**](#image-publish-procedure). Also, werf pushes stages cache of all images from config with a so-called [**stages publish procedure**](#stages-publish-procedure).-->

<!--Before digging into these algorithms, it is helpful to see how to publish images using Docker.-->

<!--### Stages publish procedure-->

<!--To publish stages cache of a image from the config werf implements the **stages publish procedure**. It consists of the following steps:-->

<!-- 1. Create temporary image names aliases for all docker images in stages cache, so that:-->
<!--     - [docker repository name](https://docs.docker.com/glossary/?term=repository) is a `REPO` parameter specified by the user without changes ([details about `REPO`]({{ site.baseurl }}/documentation/reference/registry/image_naming.html#repo-parameter)).-->
<!--     - [docker tag name](https://docs.docker.com/glossary/?term=tag) constructed as a signature prefixed with a word `image-stage-` (for example `image-stage-41772c141b158349804ad27b354247df8984ead077a5dd601f3940536ebe9a11`).-->
<!-- 2. Push images by newly created aliases into Docker registry.-->
<!-- 3. Delete temporary image names aliases.-->

<!--All of these steps are also performed with a single werf command, which will be described below.-->

<!--The result of this procedure is multiple images from stages cache of image pushed into the Docker registry.-->

## Image publishing procedure

Generally, the publishing process in the Docker ecosystem consists of the following steps:

```shell
docker tag REPO:TAG
docker push REPO:TAG
docker rmi REPO:TAG
```

 1. Getting a name or an id of the created local image.
 2. Creating a temporary alias-image for the above image that consists of two parts:
     - [docker repository name](https://docs.docker.com/glossary/?term=repository) with embedded Docker registry address;
     - [docker tag name](https://docs.docker.com/glossary/?term=tag).
 3. Pushing an alias into the Docker registry.
 4. Deleting a temporary alias.

To publish an [image]({{ site.baseurl }}/documentation/reference/stages_and_images.html#images) from the config, werf implements another logic:

 1. Create **a new image** based on the built image with the specified name and save the internal service information about tagging schema to this image (using docker labels). This information is referred to as an image **meta-information**. werf uses this information in the [deploying process]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#integration-with-built-images) and the [cleaning process]({{ site.baseurl }}/documentation/reference/cleaning_process.html).
 2. Push the newly created image into the Docker registry.
 3. Delete the temporary image created in the first step.

This procedure will be referred to as the **image publishing procedure**.

The result of this procedure is an image named using the [*rules for naming images*](#naming-images) and pushed into the Docker registry. All these steps are performed with the [werf publish command]({{ site.baseurl }}/documentation/cli/main/publish.html) or the [werf build-and-publish command]({{ site.baseurl }}/documentation/cli/main/build_and_publish.html).

## Naming images

During the image publishing procedure, werf forms the image name using:
 * _images repo_ param;
 * _images repo mode_ param;
 * image name from the werf.yaml;
 * tag param.

The final name of the docker image has the form [`DOCKER_REPOSITORY`](https://docs.docker.com/glossary/?term=repository)`:`[`TAG`](https://docs.docker.com/engine/reference/commandline/tag).

The _images repo_ and _images repo mode_ params define where and how to store images.
If the werf project contains only one nameless image, then the _images repo_ is used as a docker repository as it is, and the resulting name of a docker image gets the following form: `IMAGES_REPO:TAG`.

Otherwise, werf constructs the resulting name of a docker image for every image depending on the _images repo mode_:  
- `IMAGES_REPO:IMAGE_NAME-TAG` pattern for a `monorepo` mode;
- `IMAGES_REPO/IMAGE_NAME:TAG` pattern for a `multirepo` mode.

The _images repo_ param should be specified by the `--images-repo` option or `$WERF_IMAGES_REPO`.

The _images repo mode_ param should be specified by the `--images-repo-mode` option or `$WERF_IMAGES_REPO_MODE`.

> The image naming behavior should be the same for publishing, deploying, and cleaning processes. Otherwise, the pipeline may fail, and you may end up losing images and stages during the cleanup.

The *docker tag* is taken from `--tag-*` params:

| option                     | description                                                                     |
| -------------------------- | ------------------------------------------------------------------------------- |
| `--tag-git-tag TAG`        | Use git-tag tagging strategy and tag by the specified git tag                   |
| `--tag-git-branch BRANCH`  | Use git-branch tagging strategy and tag by the specified git branch             |
| `--tag-git-commit COMMIT`  | Use git-commit tagging strategy and tag by the specified git commit hash        |
| `--tag-custom TAG`         | Use custom tagging strategy and tag by the specified arbitrary tag              |

All the specified tag params will be validated for the conformity with the tagging rules for docker images. User may apply the slug algorithm to the specified tag, learn [more about the slug]({{ site.baseurl }}/documentation/reference/toolbox/slug.html).

Also, user specifies both the tag value and the tagging strategy by using `--tag-*` options.
The tagging strategy affects [certain policies in the cleaning process]({{ site.baseurl }}/documentation/reference/cleaning_process.html#cleanup-policies).

Every `--tag-git-*` option requires a `TAG`, `BRANCH`, or `COMMIT` argument. These options are designed to be compatible with modern CI/CD systems, where a CI job is running in the detached git worktree for the specific commit, and the current git-tag, git-branch, or git-commit is passed to the job using environment variables (for example `CI_COMMIT_TAG`, `CI_COMMIT_REF_NAME` and `CI_COMMIT_SHA` for the GitLab CI).

### Combining parameters

Any combination of tagging parameters can be used simultaneously in the [werf publish command]({{ site.baseurl }}/documentation/cli/main/publish.html) or [werf build-and-publish command]({{ site.baseurl }}/documentation/cli/main/build_and_publish.html). As a result, werf will publish a separate image for each tagging parameter of every image in a project.

## Examples

### Linking images to a git tag

Let's suppose `werf.yaml` defines two images: `backend` and `frontend`.

The following command:

```shell
werf publish --stages-storage :local --images-repo registry.hello.com/web/core/system --tag-git-tag v1.2.0
```

produces the following image names, respectively:
 * `registry.hello.com/web/core/system/backend:v1.2.0`;
 * `registry.hello.com/web/core/system/frontend:v1.2.0`.

### Linking images to a git branch

Let's suppose `werf.yaml` defines two images: `backend` and `frontend`.

The following command:

```shell
werf publish --stages-storage :local --images-repo registry.hello.com/web/core/system --tag-git-branch my-feature-x
```

produces the following image names, respectively:
 * `registry.hello.com/web/core/system/backend:my-feature-x`;
 * `registry.hello.com/web/core/system/frontend:my-feature-x`.

### Linking images to a git branch with special characters in the name

Once again, we have a `werf.yaml` file with two defined images: `backend` and `frontend`.

The following command:

```shell
werf publish --stages-storage :local --images-repo registry.hello.com/web/core/system --tag-git-branch $(werf slugify --format docker-tag "Features/MyFeature#169")
```

produces the following image names, respectively:
 * `registry.hello.com/web/core/system/backend:features-myfeature169-3167bc8c`;
 * `registry.hello.com/web/core/system/frontend:features-myfeature169-3167bc8c`.

Note that the [`werf slugify`]({{ site.baseurl }}/documentation/cli/toolbox/slugify.html) command generates a valid docker tag. Learn [more about the slug]({{ site.baseurl }}/documentation/reference/toolbox/slug.html).

### Linking images to a GitLab CI job

Let's say we have a `werf.yaml` configuration file that defines two images, `backend` and `frontend`.

Running the following command in a GitLab CI job for a project named `web/core/system` with the git branch set as `core/feature/ADD_SETTINGS` and the Docker registry configured as `registry.hello.com/web/core/system`:

```shell
type werf && source <(werf ci-env gitlab --tagging-strategy tag-or-branch --verbose)
werf publish --stages-storage :local
```

yields the following image names:
 * `registry.hello.com/web/core/system/backend:core-feature-add-settings-df80fdc3`;
 * `registry.hello.com/web/core/system/frontend:core-feature-add-settings-df80fdc3`.

Note that werf automatically applies slug to the resulting tag of the docker image: `core/feature/ADD_SETTINGS` is converted to `core-feature-add-settings-df80fdc3`. This conversion occurs in the `werf ci-env` command, which determines the name of a git branch from the GitLab CI environment, automatically slugs it and sets `WERF_TAG_GIT_BRANCH` (which is alternative way to set the `--tag-git-branch` parameter). See [more about the slug]({{ site.baseurl }}/documentation/reference/toolbox/slug.html).

### Unnamed image and a GitLab CI job

Let's suppose we have a werf.yaml with a single unnamed image. Running the following command in the GitLab CI job for the project named `web/core/queue` with the git-tag named `v2.3.1` and a Docker registry configured as `registry.hello.com/web/core/queue`:

```shell
type werf && source <(werf ci-env gitlab --tagging-strategy tag-or-branch --verbose)
werf publish --stages-storage :local
```

yields the following result: `registry.hello.com/web/core/queue:v2.3.1`.
