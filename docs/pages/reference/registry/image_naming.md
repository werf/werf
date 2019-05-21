---
title: Image naming
sidebar: reference
permalink: reference/registry/image_naming.html
author: Artem Kladov <artem.kladov@flant.com>
---

Werf builds and tags docker images to run and push in a registry. For tagging images werf uses a [_images repo_](#images-repo) and image tag parameters (`--tag-*`) in the following commands:
* [Publish commands]({{ site.baseurl }}/reference/registry/publish.html)
* [Cleaning commands]({{ site.baseurl }}/reference/registry/cleaning.html)
* [Deploy commands]({{ site.baseurl }}/reference/deploy/deploy_to_kubernetes.html#werf-kube-deploy)

## Werf tag procedure
In a Docker world a tag is a creating an alias name for existent docker image.

In Werf world tagging creates **a new image layer** with the specified name. Werf stores internal service information about tagging schema in this layer (using docker labels). This information is referred to as image **meta-information**. Werf uses this information in [deploying]({{ site.baseurl }}/reference/deploy/deploy_to_kubernetes.html#werf-kube-deploy) and [cleaning]({{ site.baseurl }}/reference/registry/cleaning.html) processes.

The procedure of creating such a layer will be referred to as **werf tag procedure**.

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
Tagging strategy affects on [certain policies in cleaning process]({{ site.baseurl }}/reference/registry/cleaning.html#cleanup-policies).

Besides, werf applies [tag slug]({{ site.baseurl }}/reference/toolbox/slug.html#basic-algorithm) transformation rules on all options except `--tag-custom`.
Thus, these tags satisfy the slug requirements, docker tag format, and a user can use arbitrary git tag and branch name formats.

### Combining parameters

Any combination of tag parameters can be used simultaneously for [publish commands]({{ site.baseurl }}/reference/registry/publish.html). In the result, there is a separate image for each tag parameter of each image in a project.

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

Each image name converts according to [tag slug]({{ site.baseurl }}/reference/toolbox/slug.html#basic-algorithm) rules with appended murmurhash of original string.

### Unnamed image in GitLab job

Given config with single unnamed image. The following command runs in GitLab job for git-tag named `v2.3.1`:

```bash
type werf && source <(werf ci-env gitlab --tagging-strategy tag-or-branch --verbose)
werf publish --stages-storage :local
```

Image name in the result is `registry.hello.com/web/core/queue:v2-3-1-5cb8b0a4`

Image name converts according to [tag slug]({{ site.baseurl }}/reference/toolbox/slug.html#basic-algorithm) rules with appended murmurhash of original string, because of points symbols in the tag `v2.3.1` (points don't meet the requirements).

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

