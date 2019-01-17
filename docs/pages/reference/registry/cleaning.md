---
title: Cleaning
sidebar: reference
permalink: reference/registry/cleaning.html
author: Artem Kladov <artem.kladov@flant.com>, Timofey Kirillov <timofey.kirillov@flant.com>
---

Build and push processes create sets of docker layers but don't delete them. As a result, local images storage and docker registry instantly grows and consume more and more space. Interrupted build process leaves stale images. When git branch has been deleted, a set of images which was built for this branch also leave in a docker registry. It is necessary to clean a docker registry periodically. Otherwise, it will be filled with **stale images**.

## Methods of cleaning

Dapp has an efficient multi-level images cleaning. There are two methods of images cleaning in dapp:

1. Cleaning by policies
2. Manual cleaning

### Cleaning by policies

Cleaning by policies helps to organize automatic periodical cleaning of stale images. It implies regular gradual cleaning of stale images according to cleaning policies. This method is the safest way of cleaning because it doesn't affect your production environment.

The cleaning by policies method includes the steps in the following order:
1. [**Garbage collection**](#garbage-collection) executes a cleaning docker registry from stale images according to the cleaning policies.
2. [**Local storage synchronization**](#local-storage-synchronization) performs syncing local docker image storage with docker registry.

A docker registry is the primary source of information about actual and stale images. Therefore, it is essential to clean docker registry on the first step and only then synchronize the content of local storage with the docker registry.

### Manual cleaning

Manual cleaning method assumes one-step cleaning with the complete removal of images from local docker image storage or docker registry. This method doesn't check whether the image used by kubernetes or not. Manual cleaning is not recommended for automatical using (use cleaning by policies instead). In general it suitable for forced images removal.

The manual cleaning method includes the following options:

* [**Flush**](#flush). Deleting images of the **current project** in local storage or docker registry.
* [**Reset**](#reset). Deleting images of **all dapp projects** in local storage.

## Garbage collection

Garbage collection is a dapp ability to automate cleaning of a docker registry. It works according to special rules called **garbage collection policies**. These policies determine which images to delete and which not to.

### Garbage collection policies

* **by branches:**
    * Every new commit updates the image for the git branch (there is the only docker tag for the git branch).
    * Dapp deletes the image from the docker registry when the corresponding git branch doesn't exist. The image always remains, while the corresponding git branch exists.
    * The policy covers images tagged by dapp with `--tag-ci` or `--tag-branch` tags.
* **by commits:**
    * Dapp deletes the image from the docker registry when the corresponding git commit doesn't exist.
    * For the remaining images, the following policies apply:
       * `WERF_GIT_COMMITS_EXPIRY_DATE_PERIOD_POLICY`. Deleting images uploaded in docker registry more than **30 days**. 30 days is a default period. To change the default period set `WERF_GIT_COMMITS_EXPIRY_DATE_PERIOD_POLICY` environment variable in seconds.
       * `WERF_GIT_COMMITS_LIMIT_POLICY`. Deleting all images in docker registry except **last 50 images**. 50 images is a default value. To change the default value set count in  `WERF_GIT_COMMITS_LIMIT_POLICY` environment variables.
    * The policy covers images tagged by dapp with `--tag-commit` tag.
* **by tags:**
    * Dapp deletes the image from the docker registry when the corresponding git tag doesn't exist.
    * For the remaining images, the following policies apply:
      * `WERF_GIT_TAGS_EXPIRY_DATE_PERIOD_POLICY`. Deleting images uploaded in docker registry more than **30 days**. 30 days is a default period. To change the default period set `WERF_GIT_TAGS_EXPIRY_DATE_PERIOD_POLICY` environment variable in seconds;
      * `WERF_GIT_TAGS_LIMIT_POLICY`.  Deleting all images in docker registry except **last 10 images**. 10 images is a default value. To change the default value set count in `WERF_GIT_TAGS_LIMIT_POLICY`.
    * The policy covers images tagged by dapp with `--tag-ci` tag.

**Pay attention,** that garbage collection affects only images built by dapp **and** images tagged by dapp with one of the `--tag-ci`, `--tag-branch` or `--tag-commit` options. Other images in the docker registry stay as they are.

### Whitelist of images

The image always remains in docker registry while exists kubernetes object which uses the image. In kubernetes cluster dapp scans the following kinds of objects: `pod`, `deployment`, `replicaset`, `statefulset`, `daemonset`, `job`, `cronjob`, `replicationcontroller`.

The functionality can be disabled by option `--without-kube`.

#### Connecting to kubernetes

Dapp gets information about kubernetes clusters and how to connect to them in the following order:
1. gets `KUBECONFIG` environment variable for the path to kubectl configuration file;
2. or executes `kubectl config view` command if kubectl is available in the system;
3. or gets `~/.kube/config` kubectl configuration file.

Dapp connects to all kubernetes clusters, defined in all contexts of kubectl configuration, to gather images that are in use.

### Docker registry authorization

For docker registry authorization in garbage collection, dapp require the `WERF_CLEANUP_REGISTRY_PASSWORD` environment variable with access token in it (read more about [authorization]({{ site.baseurl }}/reference/registry/authorization.html#autologin-for-cleaning-commands)).

### Syntax

```bash
dapp dimg cleanup repo [--with-stages=false]
```

`--with-stages` — option used to delete not only tagged images but also _related_ stages cache, which was pushed into the registry (see [push for details]({{ site.baseurl }}/reference/registry/authorization.html#stages-push-procedure)).

## Local storage synchronization

After cleaning docker registry on the garbage collection step, local storage still contains all of the images that have been deleted from the docker registry. These images include tagged dimg images (created with a [dapp tag procedure]({{ site.baseurl }}/reference/registry/image_naming.html#dapp-tag-procedure)) and stages cache for dimgs.

Executing a local storage synchronization is necessary to update local storage according to a docker registry. On the local storage synchronization step, dapp deletes old local stages cache, which doesn't relate to any of the dimgs currently existing in the docker registry.

There are some consequences of this algorithm:

1. If the garbage collection, — the first step of cleaning by policies, — was skipped, then local storage synchronization makes no sense.
2. Dapp completely removes local stages cache for the built dimgs, that don't exist into the docker registry.

### Syntax

```bash
dapp dimg stages cleanup local [options] REPO
  --improper-cache-version
  --improper-repo-cache
```

`--improper-cache-version` —  dapp deletes images built by another dapp version.

`--improper-repo-cache` — dapp deletes stages cache non related with any dimg in docker repository. This option enables the exact synchronization with the docker registry step. Images of your project that is not pushed yet will also be deleted with this option.

## Flush

Allows deleting information about specified (current) project. Flush includes cleaning both the local storage and docker registry.

Local storage cleaning includes:
* Deleting stages cache images of the project. Also deleting images from previous dapp version.
* Deleting all of the dimgs tagged by `dapp dimg tag` command.
* Deleting `<none>` images of the project. `<none>` images can remain as a result of build process interruption. In this case, such built images exist as orphans outside of the stages cache. These images also not deleted by the garbage collection.
* Deleting containers associated with the images of the project.

Docker registry cleaning includes:
* Deleting pushed dimgs of the project.
* Deleting pushed stages cache of the project.

### Syntax

```bash
dapp dimg flush local [--with-stages]
```

The command deletes images of the current project only in the local storage.

```bash
dapp dimg flush repo REPO [--with-stages]
```

The command deletes images of the current project only in docker registry.

The `--with-stages` option uses to delete not only tagged images but also stages cache. The option is disabled by default. It is always recommended to use this option.

## Reset

With this variant of cleaning, dapp can delete all images, containers, and files from all projects created by dapp on the host. The files include `~/.dapp/builds` directory, which is a build dir cache directory for all projects on the build host (see [the cache article]({{ site.baseurl }}/reference/build/cache.html) for the details).

Reset is the fullest method of cleaning on the local machine.

### Syntax

```bash
dapp dimg mrproper [--all] [--improper-cache-version-stages] [--improper-dev-mode-cache]
```

* `--improper-cache-version-stages` — delete stages cache, images, and containers created by versions of dapp older than current.
* `--improper-dev-mode-cache` — delete stages cache, images and container created in developer mode (`--dev` option).
* `--all` — delete stages cache with all of images and containers ever dapp created in local storage. This option supersedes any other options specified. That means images targeted by options `--improper-cache-version-stages` and `--improper-dev-mode-cache` will also be deleted.

## Examples

### Delete all images and stages cache in the current project

Given dappfile with two dimgs. Images succesfully built and tagged.

```bash
dapp dimg flush local --with-stages
```

Command deletes all tagged images, and delete stages cache.

### Delete all images of all dapp projects

Given several dapp projects.

```bash
dapp dimg mrproper --all --improper-cache-version-stages --improper-dev-mode-cache
```

Command delete every image created by dapp. Other docker images like image specified in `from` directive in dappfile - remains.

### Cleaning GitLab CI docker registry

Given project, where dapp executes with `--tag-ci` option to push images.

On the first step, the following command performs garbage collection in repo:

```bash
dapp dimg cleanup repo ${CI_REGISTRY_IMAGE}
```

On the second step, the following command performs synchronizing of local storage with the registry:

```bash
dapp dimg stages cleanup local --improper-cache-version --improper-repo-cache ${CI_REGISTRY_IMAGE}
```

Read a deeper example [here]({{ site.baseurl }}/how_to/gitlab_ci_cd_integration.html).
