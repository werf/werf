---
title: Cleaning process
sidebar: documentation
permalink: documentation/reference/cleaning_process.html
author: Artem Kladov <artem.kladov@flant.com>, Timofey Kirillov <timofey.kirillov@flant.com>
---

Build and publish processes create sets of docker layers but do not delete them.
As a result, _stages storage_ and _images repo_ instantly grow and consume more and more space.
Interrupted build process leaves stale images.
When git branch or git tag has been deleted, a set of _stages_ which were built for this _image_ also leave in _images repo_ and _stages storage_.
It is necessary to periodically clean up _images repo_ and _stages storage_.
Otherwise, it will be filled with **stale images**.

Werf has an efficient multi-level images cleaning. There are following cleaning approaches:

1. [**Cleaning by policies**](#cleaning-by-policies)
2. [**Manual cleaning**](#manual-cleaning)
3. [**Host cleaning**](#host-cleaning)

## Cleaning by policies

Cleaning by policies helps to organize automatic periodical cleaning of stale images.
It implies regular gradual cleaning according to cleaning policies.
This is the safest way of cleaning because it does not affect your production environment.

The cleaning by policies method includes the steps in the following order:
1. [**Cleanup images repo**](#cleanup-images-repo) cleans _images repo_ from stale images according to the cleaning policies.
2. [**Cleanup stages storage**](#cleanup-stages-storage) performs syncing _stages storage_ with _images repo_.

These steps are combined in one top-level command [cleanup]({{ site.baseurl }}/documentation/cli/main/cleanup.html).  

An _images repo_ is the primary source of information about actual and stale images.
Therefore, it is essential to clean _images repo_ on the first step and only then _stages storage_.

### Cleanup images repo

There is a werf ability to automate cleaning of _images repo_.
It works according to special rules called **cleanup policies**.
These policies determine which _images_ to delete and which not to.

#### Cleanup policies

* **by branches:**
    * Every new commit updates an image for the git branch (there is the only one docker tag for each published git branch).
    * Werf deletes an image from _images repo_ when the corresponding git branch does not exist. The image always remains, while the corresponding git branch exists.
    * The policy covers images tagged by werf with `--tag-git-branch` option.
* **by commits:**
    * Werf deletes an image from _images repo_ when the corresponding git commit does not exist.
    * For the remaining images, the following policies apply:
      * _git-commit-strategy-expiry-days_.
      Keep published _images_ in the _images repo_ for the **specified maximum days** since image published.
      Republished image will be kept **specified maximum days** since new publication date.
      No days limit by default, -1 disables the limit.
      Value can be specified by `--git-commit-strategy-expiry-days` option or `$WERF_GIT_COMMIT_STRATEGY_EXPIRY_DAYS`.
      * _git-commit-strategy-limit_.
      Keep **specified max number** of published _images_ in the _images repo_.
      No limit by default, -1 disables the limit.
      Value can be specified by `--git-commit-strategy-limit` or `$WERF_GIT_COMMIT_STRATEGY_LIMIT`.
    * The policy covers images tagged by werf with `--tag-git-commit` option.
* **by tags:**
    * Werf deletes an image from _images repo_ when the corresponding git tag does not exist.
    * For the remaining images, the following policies apply:
       * _git-tag-strategy-expiry-days_.
       Keep published _images_ in the _images repo_ for the **specified maximum days** since image published.
       Republished image will be kept **specified maximum days** since new publication date.
       No days limit by default, -1 disables the limit.
       Value can be specified by `--git-tag-strategy-expiry-days` option or `$WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS`.
       * _git-tag-strategy-limit_.
       Keep **specified max number** of published _images_ in the _images repo_.
       No limit by default, -1 disables the limit.
       Value can be specified by `--git-tag-strategy-limit` or `$WERF_GIT_TAG_STRATEGY_LIMIT`.
    * The policy covers images tagged by werf with `--tag-git-tag` option.

**Pay attention,** that cleanup affects only images built by werf **and** images tagged by werf with one of the following options: `--tag-git-branch`, `--tag-git-tag` or `--tag-git-commit`.
Other images in the _images repo_ stay as they are.

#### Whitelist of images

The image always remains in _images repo_ while exists kubernetes object which uses the image.
In kubernetes cluster, werf scans the following kinds of objects: `pod`, `deployment`, `replicaset`, `statefulset`, `daemonset`, `job`, `cronjob`, `replicationcontroller`.

The functionality can be disabled by option `--without-kube`.

#### Connecting to kubernetes

Werf gets information about kubernetes clusters and how to connect to them from the kube configuration file `~/.kube/config`. Werf connects to all kubernetes clusters, defined in all contexts of kubectl configuration, to gather images that are in use.

### Cleanup stages storage

Executing a [stages storage cleanup command]({{ site.baseurl }}/documentation/cli/management/stages/cleanup.html) is necessary to update it according to a _images repo_.
On this step, werf deletes _stages_ which do not relate to _images_ currently existing in the _images repo_.

> If the [images cleanup command]({{ site.baseurl }}/documentation/cli/management/images/cleanup.html), — the first step of cleaning by policies, — was skipped, then [stages storage cleanup]({{ site.baseurl }}/documentation/cli/management/stages/cleanup.html) makes no sense

## Manual cleaning

Manual cleaning approach assumes one-step cleaning with the complete removal of images from _stages storage_ or _images repo_.
This method does not check whether the image used by kubernetes or not.
Manual cleaning is not recommended for automatic usage (use cleaning by policies instead).
In general it suitable for forced images removal.

The manual cleaning approach includes the following options:

* [Purge images repo command]({{ site.baseurl }}/documentation/cli/management/images/purge.html). Deleting images of the **current project** in _images repo_.
* [Purge stages storage command]({{ site.baseurl }}/documentation/cli/management/stages/purge.html). Deleting stages of the **current project** in _stages storage_.

These steps are combined in one top-level command [purge]({{ site.baseurl }}/documentation/cli/main/purge.html).

## Host cleaning

There are following commands to cleanup host machine:

* [Cleanup host machine command]({{ site.baseurl }}/documentation/cli/management/host/cleanup.html). Cleanup old unused werf cache and data of **all projects** on host machine.
* [Purge host machine command]({{ site.baseurl }}/documentation/cli/management/host/purge.html). Purge werf _images_, _stages_, cache and other data of **all projects** on host machine.
