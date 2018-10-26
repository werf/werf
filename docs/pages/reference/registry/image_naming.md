---
title: Image naming
sidebar: reference
permalink: reference/registry/image_naming.html
author: Artem Kladov <artem.kladov@flant.com>
---

Dapp builds and tags docker images to run and push in a registry. For tagging images dapp uses a `REPO` and image tag parameters (`--tag-*`) in the following commands:
* [Tag commands]({{ site.baseurl }}/reference/registry/tag.html)
* [Push commands]({{ site.baseurl }}/reference/registry/push.html)
* [Pull commands]({{ site.baseurl }}/reference/registry/pull.html)
* [Cleaning commands]({{ site.baseurl }}/reference/registry/cleaning.html)
* [Deploy commands]({{ site.baseurl }}/reference/deploy/deploy_to_kubernetes.html#dapp-kube-deploy)

## Dapp tag procedure

In a Docker world a tag is a creating an alias name for existent docker image.

In Dapp world tagging creates **a new image layer** with the specified name. Dapp stores internal service information about tagging schema in this layer (using docker labels). This information is referred to as image **meta-information**. Dapp uses this information in [deploying]({{ site.baseurl }}/reference/deploy/deploy_to_kubernetes.html#dapp-kube-deploy) and [cleaning]({{ site.baseurl }}/reference/registry/cleaning.html) processes.

The procedure of creating such a layer will be referred to as **dapp tag procedure**.
 
## REPO parameter

For all commands related to a docker registry, dapp uses a single parameter named `REPO`. Using this parameter dapp constructs a [docker repository](https://docs.docker.com/glossary/?term=repository) as follows:

* If dapp project contains nameless dimg, dapp uses `REPO` as docker repository.
* Otherwise, dapp constructs docker repository name for each dimg by following template `REPO/DIMG_NAME`.

E.g., if there is unnamed dimg in a dappfile and `REPO` is `registry.flant.com/sys/backend` then the docker repository name is the `registry.flant.com/sys/backend`.  If there are two dimgs in a dappfile — `server` and `worker`, then docker repository names are:
* `registry.flant.com/sys/backend/server` for `server` dimg;
* `registry.flant.com/sys/backend/worker` for `worker` dimg.

### Minikube docker registry

`dapp kube minikube setup` command is used to prepare kubernetes environment with docker registry running in minikube, see [article]({{ site.baseurl }}/reference/deploy/minikube.html#dapp-kube-minikube-setup) for details. To specify this docker registry address in `REPO` parameter dapp supports a shortcut value `:minikube`.

`:minikube` value of `REPO` parameter automatically unfolds into `localhost:5000/<dapp-name>` (more about [dapp name](https://flant.github.io/dapp/reference/glossary.html#dapp-name)).

## Image tag parameters

| option | description |
| ----- | -------- |
| `--tag-ci` | tag with an autodetection by CI related environment variables |
| `--tag-build-id` | tag with a CI job id |
| `--tag-branch` | tag with a git branch name |
| `--tag-commit` | tag with a git commit id |
| `--tag\|--tag-slug TAG` | arbitrary slugged tag TAG |
| `--tag-plain TAG` | arbitrary non slugged tag TAG |

### --tag-ci

Tag works only in GitLab CI or Travis CI environment.

The [docker tag](https://docs.docker.com/glossary/?term=tag) name is based on a current git-tag or git-branch, for which CI job is run.

Dapp determines git tag name based on `CI_COMMIT_TAG` environment variable for Gitlab CI or `TRAVIS_TAG` environment variable for Travis CI. This environment variable will be referred to as **git tag variable**.

Dapp determines git branch name based on `CI_COMMIT_REF_NAME` environment variable for Gitlab CI or `TRAVIS_BRANCH` environment variable for Travis CI. This environment variable will be referred to as **git branch variable**.

The **git tag variable** presents only when building git tag. In this case, dapp creates an image and marks it as an image built for git tag (by adding **meta-information** into newly created docker layer).

If the **git tag variable** is absent, then dapp uses the **git branch variable** to get the name of building git branch. In this case, dapp creates an image and marks it as an image built for git branch (by adding **meta-information** into newly created docker layer).

After getting git tag or git branch name, dapp applies [slug](#slug) transformation rules if this name doesn't meet with the slug requirements. This behavior allows using tags and branches with arbitrary names in Gitlab like `review/fix#23`.

Option `--tag-ci` is the most recommended way for building in CI.

### --tag-build-id

Tag works only in GitLab CI or Travis CI environment.

The tag name based on the unique id of the current GitLab job from `CI_BUILD_ID` or `CI_JOB_ID` environment variable for Gitlab CI or `TRAVIS_BUILD_NUMBER` environment variable for Travis CI.

### --tag-branch

The tag name based on a current git branch. Dapp looks for the current branch in the local git repository where dappfile located.

After getting git branch name, dapp apply [slug](#slug) transformation rules if tag name doesn't meet with the slug requirements. This behavior allows using branches with arbitrary names like `my/second_patch`.

### --tag-commit

The tag name based on a current git commit id (full-length SHA hashsum). Dapp looks for the current commit id in the local git repository where dappfile located.

### --tag|--tag-slug TAG

`--tag TAG` and `--tag-slug TAG` is an alias of each other.

The tag name based on a specified TAG in the parameter.

Dapp applies [slug](#slug) transformation rules to TAG, and if it doesn't meet with the slug requirements. This behavior allows using text with arbitrary chars as a TAG.

### --tag-plain TAG

The tag name based on a specified TAG in the parameter.

Dapp doesn't apply [slug](#slug) transformation rules to TAG, even though it contains inappropriate symbols and doesn't meet with the slug requirements. Pay attention, that this behavior could lead to an error.

### Default values

By default, dapp uses `latest` as a docker tag for all images of dappfile.

### Combining parameters

Any combination of tag parameters can be used simultaneously for [tag commands]({{ site.baseurl }}/reference/registry/tag.html) and [push commands]({{ site.baseurl }}/reference/registry/push.html). In the result, there is a separate image for each tag parameter of each dimg in a project.

## Examples

### Two dimgs

Given dappfile with 2 dimgs — backend and frontend.

The following command:

```bash
dapp dimg tag registry.hello.com/web/core/system --tag-plain v1.2.0
```

produces the following image names respectively:
* `registry.hello.com/web/core/system/backend:v1.2.0`;
* `registry.hello.com/web/core/system/frontend:v1.2.0`.

### Two dimgs in GitLab job

Given Dappfile with 2 dimgs — backend and frontend.

The following command runs in GitLab job for git branch named `core/feature/ADD_SETTINGS`:
```bash
dapp dimg push registry.hello.com/web/core/system --tag-ci
```

Image names in the result are:
* `registry.hello.com/web/core/system/backend:core-feature-add-settings-c3fd80df`
* `registry.hello.com/web/core/system/frontend:core-feature-add-settings-c3fd80df`

Each image name converts according to slug rules with adding murmurhash.

### Unnamed dimg in GitLab job

Given dappfile with single unnamed dimg. The following command runs in GitLab job for git-tag named `v2.3.1`:

```bash
dapp dimg push registry.hello.com/web/core/queue --tag-ci
```

Image name in the result is `registry.hello.com/web/core/queue:v2-3-1-5cb8b0a4`

Image name converts according to slug rules with adding murmurhash, because of points symbols in the tag `v2.3.1` (points don't meet the requirements).

### Two dimgs with multiple tags in GitLab job

Given dappfile with 2 dimgs — backend and frontend. The following command runs in GitLab job for git-branch named `rework-cache`:

```bash
dapp dimg push registry.hello.com/web/core/system --tag-ci --tag "feature/using_cache" --tag-plain my-test-branch
```

The command produces 6 image names for each dimg name and each tag-parameter (two dimgs by three tags):
* `registry.hello.com/web/core/system/backend:rework-cache`
* `registry.hello.com/web/core/system/frontend:rework-cache`
* `registry.hello.com/web/core/system/backend:feature-using-cache-81644ed0`
* `registry.hello.com/web/core/system/frontend:feature-using-cache-81644ed0`
* `registry.hello.com/web/core/system/backend:my-test-branch`
* `registry.hello.com/web/core/system/frontend:my-test-branch`

For `--tag` parameter image names are converting, but for `--tag-ci` parameter image names are not converting, because branch name meets `rework-cache` the requirements.

## Slug

In some cases, text from environment variables or parameters can't be used AS IS because it can contain unacceptable symbols.

To take into account restrictions for docker images names, helm releases names and kubernetes namespaces dapp applies unified slug algorithm when producing these names. This algorithm excludes unacceptable symbols from an arbitrary text and guarantees the uniqueness of the result for each unique input.

Dapp applies slug algorithm internally for names such as [dapp name], docker tags names, helm releases names and kubernetes namespaces. Also, there is a `dapp slug` command (see syntax below) which applies this algorithm for provided input text. You can use this command upon your needs.

### Algorithm

Dapp checks the text for compliance with slug **requirements**, and if text complies with slug requirements — dapp doesn't modify it. Otherwise, dapp performs **transformations** of the text to comply the requirements and add a dash symbol followed by a hash suffix based on the source text. A hash algorithm is a [MurmurHash](https://en.wikipedia.org/wiki/MurmurHash).

A text complies with slug requirements if:
* it has only lowercase alphanumerical ASCII characters and dashes;
* it doesn't start or end with dashes;
* it doesn't contain multiple dashes sequences.

The following steps perform, when dapp apply transformations of the text in slug:
* Converting UTF-8 latin characters to their ASCII counterpart;
* Replacing some special symbols with dash symbol (`~><+=:;.,[]{}()_&`);
* Removing all non-recognized characters (leaving lowercase alphanumerical characters and dashes);
* Removing starting and ending dashes;
* Reducing multiple dashes sequences to one dash.

#### Version 3

Dapp has experimental additions to basic slug algorithm, which is non-default for now. To enable **version 3** algorithm set `DAPP_SLUG_V3=1` environment variable.

This algorithm has additional requirement for input text:

* it should contain no more than 53 byte chars.

If this requirement does not meet, then dapp transforms input text and adds hash as described earlier. The result fits in 53 byte chars.

### Syntax

```bash
dapp slug STRING
```

Transform `STRING` according to the slug algorithm (see above) and prints the result.

### Examples

The text `feature-fix-2` isn't transforming because it meets the slug requirements. The result is:

```bash
$ dapp slug 'feature-fix-2'
feature-fix-2
```

The text `branch/one/!@#4.4-3` transforming because it doesn't meet the slug requirements and hash adding. The result is:

```bash
$ dapp slug 'branch/one/!@#4.4-3'
branch-one-4-4-3-5589e04f
```
