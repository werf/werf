---
title: Image naming
sidebar: reference
permalink: reference/registry/image_naming.html
author: Artem Kladov <artem.kladov@flant.com>
---

Werf builds and tags docker images to run and push in a registry. For tagging images werf uses a `REPO` and image tag parameters (`--tag-*`) in the following commands:
* [Tag commands]({{ site.baseurl }}/reference/registry/tag.html)
* [Push commands]({{ site.baseurl }}/reference/registry/push.html)
* [Pull commands]({{ site.baseurl }}/reference/registry/pull.html)
* [Cleaning commands]({{ site.baseurl }}/reference/registry/cleaning.html)
* [Deploy commands]({{ site.baseurl }}/reference/deploy/deploy_to_kubernetes.html#werf-kube-deploy)

## Werf tag procedure

In a Docker world a tag is a creating an alias name for existent docker image.

In Werf world tagging creates **a new image layer** with the specified name. Werf stores internal service information about tagging schema in this layer (using docker labels). This information is referred to as image **meta-information**. Werf uses this information in [deploying]({{ site.baseurl }}/reference/deploy/deploy_to_kubernetes.html#werf-kube-deploy) and [cleaning]({{ site.baseurl }}/reference/registry/cleaning.html) processes.

The procedure of creating such a layer will be referred to as **werf tag procedure**.

## `--repo REPO` option

For all commands related to a docker registry, werf uses a single option named `--repo REPO`. `REPO` is a required param, but `--repo` option may be omitted in Gitlab CI. Werf will try to autodetect repo by environment variable `WERF_IMAGES_REPO` if available.

Using `REPO` werf constructs a [docker repository](https://docs.docker.com/glossary/?term=repository) as follows:

* If werf project contains nameless image, werf uses `REPO` as docker repository.
* Otherwise, werf constructs docker repository name for each image by following template `REPO/IMAGE_NAME`.

E.g., if there is unnamed image in a `werf.yaml` config and `REPO` is `myregistry.myorg.com/sys/backend` then the docker repository name is the `myregistry.myorg.com/sys/backend`.  If there are two images in a config — `server` and `worker`, then docker repository names are:
* `myregistry.myorg.com/sys/backend/server` for `server` image;
* `myregistry.myorg.com/sys/backend/worker` for `worker` image.

## Image tag parameters

| option | description |
| ----- | -------- |
| `--tag-git-tag` | tag with a git tag name |
| `--tag-git-branch` | tag with a git branch name |
| `--tag-git-commit` | tag with a git commit id |
| `--tag TAG` | arbitrary  TAG |

### `--tag-ci`

Tag works only in GitLab CI or Travis CI environment.

The [docker tag](https://docs.docker.com/glossary/?term=tag) name is based on a current git-tag or git-branch, for which CI job is run.

Werf determines git tag name based on `CI_COMMIT_TAG` environment variable for Gitlab CI or `TRAVIS_TAG` environment variable for Travis CI. This environment variable will be referred to as **git tag variable**.

Werf determines git branch name based on `CI_COMMIT_REF_NAME` environment variable for Gitlab CI or `TRAVIS_BRANCH` environment variable for Travis CI. This environment variable will be referred to as **git branch variable**.

The **git tag variable** presents only when building git tag. In this case, werf creates an image and marks it as an image built for git tag (by adding **meta-information** into newly created docker layer).

If the **git tag variable** is absent, then werf uses the **git branch variable** to get the name of building git branch. In this case, werf creates an image and marks it as an image built for git branch (by adding **meta-information** into newly created docker layer).

After getting git tag or git branch name, werf applies [tag slug]({{ site.baseurl }}/reference/toolbox/slug.html#basic-algorithm) transformation rules if this name doesn't meet with the tag slug requirements. This behavior allows using tags and branches with arbitrary names in Gitlab like `review/fix#23`.

Option `--tag-ci` is the most recommended way for building in CI.

### `--tag-git-branch`

The tag name based on a current git branch. Werf looks for the current branch in the local git repository where config located.

After getting git branch name, werf apply [tag slug]({{ site.baseurl }}/reference/toolbox/slug.html#basic-algorithm) transformation rules if tag name doesn't meet with the slug requirements. This behavior allows using branches with arbitrary names like `my/second_patch`.

### `--tag-git-commit`

The tag name based on a current git commit id (full-length SHA hashsum). Werf looks for the current commit id in the local git repository where config located.

### `--tag TAG`

The tag name based on a specified TAG in the parameter.

### Default values

By default, werf uses `latest` as a docker tag for all images of config.

### Combining parameters

Any combination of tag parameters can be used simultaneously for [tag commands]({{ site.baseurl }}/reference/registry/tag.html) and [push commands]({{ site.baseurl }}/reference/registry/push.html). In the result, there is a separate image for each tag parameter of each image in a project.

## Examples

### Two images

Given config with 2 images — backend and frontend.

The following command:

```bash
werf tag registry.hello.com/web/core/system --tag v1.2.0
```

produces the following image names respectively:
* `registry.hello.com/web/core/system/backend:v1.2.0`;
* `registry.hello.com/web/core/system/frontend:v1.2.0`.

### Two images in GitLab job

Given `werf.yaml` config with 2 images — backend and frontend.

The following command runs in GitLab job for git branch named `core/feature/ADD_SETTINGS`:
```bash
werf push --repo registry.hello.com/web/core/system --tag-ci
```

Image names in the result are:
* `registry.hello.com/web/core/system/backend:core-feature-add-settings-c3fd80df`
* `registry.hello.com/web/core/system/frontend:core-feature-add-settings-c3fd80df`

Each image name converts according to [tag slug]({{ site.baseurl }}/reference/toolbox/slug.html#basic-algorithm) rules with appended murmurhash of original string.

### Unnamed image in GitLab job

Given config with single unnamed image. The following command runs in GitLab job for git-tag named `v2.3.1`:

```bash
werf push --repo registry.hello.com/web/core/queue --tag-ci
```

Image name in the result is `registry.hello.com/web/core/queue:v2-3-1-5cb8b0a4`

Image name converts according to [tag slug]({{ site.baseurl }}/reference/toolbox/slug.html#basic-algorithm) rules with appended murmurhash of original string, because of points symbols in the tag `v2.3.1` (points don't meet the requirements).

### Two images with multiple tags in GitLab job

Given config with 2 images — backend and frontend. The following command runs in GitLab job for git-branch named `rework-cache`:

```bash
werf push --repo registry.hello.com/web/core/system --tag-ci --tag "feature/using_cache" --tag-plain my-test-branch
```

The command produces 6 image names for each image name and each tag-parameter (two images by three tags):
* `registry.hello.com/web/core/system/backend:rework-cache`
* `registry.hello.com/web/core/system/frontend:rework-cache`
* `registry.hello.com/web/core/system/backend:feature-using-cache-81644ed0`
* `registry.hello.com/web/core/system/frontend:feature-using-cache-81644ed0`
* `registry.hello.com/web/core/system/backend:my-test-branch`
* `registry.hello.com/web/core/system/frontend:my-test-branch`

For `--tag` parameter image names are converting, but for `--tag-ci` parameter image names are not converting, because branch name meets `rework-cache` the requirements.
