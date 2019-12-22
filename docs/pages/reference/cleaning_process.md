---
title: Cleaning process
sidebar: documentation
permalink: documentation/reference/cleaning_process.html
author: Artem Kladov <artem.kladov@flant.com>, Timofey Kirillov <timofey.kirillov@flant.com>
---

While building and publishing, werf creates sets of docker layers and does not delete them.
As a result, _stages storage_ and _images repo_ are steadily growing and consuming more and more space.
Also, an interrupted build process leaves staled images.
When a git branch or a git tag gets deleted, a set of _stages_ that were built for this _image_ remains in the _images repo_ and _stages storage_.
So, it is necessary to clean up the _images repo_ and _stages storage_ periodically. Otherwise, 
they will be filled with the **stale images**.

werf has an efficient multi-level image cleaning system. It supports the following cleaning methods:

1. [**Cleaning by policies**](#cleaning-by-policies)
2. [**Manual cleaning**](#manual-cleaning)
3. [**Host cleaning**](#host-cleaning)

## Cleaning by policies

Cleaning by policies helps to organize automatic periodical cleaning of stale images.
It implies regular gradual cleaning according to cleaning policies.
This is the safest way of cleaning because it does not affect your production environment.

The cleaning by policies method includes the steps in the following order:
1. [**Cleanup of the images repo**](#cleanup-images-repo) cleans _images repo_ from staled images according to the cleaning policies.
2. [**Cleanup of the stages storage**](#cleanup-stages-storage) synchronizes _stages storage_ with the _images repo_.

These steps are combined in the single top-level command [cleanup]({{ site.baseurl }}/documentation/cli/main/cleanup.html).  

An _images repo_ is the primary source of information about actual and stale images.
Therefore, it is essential to clean the _images repo_ firstly and only then proceed to the _stages storage_.

### Cleaning up the images repo

werf can automate the cleaning of the _images repo_.
It works according to special rules called **cleanup policies**.
These policies determine which _images_ will be deleted while leaving all others intact.

#### Cleanup policies

* **by branches:**
    * Each new commit updates an image for the corresponding git branch (in other words, there is the only one docker tag for each published git branch).
    * werf deletes an image from the _images repo_ if the corresponding git branch does not exist. The image is never removed as long as the corresponding git branch exists.
    * The policy covers images tagged by werf with the `--tag-git-branch` option.
* **by commits:**
    * werf deletes an image from the _images repo_ if the corresponding git commit does not exist.
    * For the remaining images, the following policies apply:
      * _git-commit-strategy-expiry-days_.
      Keep published _images_ in the _images repo_ for the **specified maximum number of days** since the image was published.
      The republished image will be kept for the **specified maximum nimber of days** since the new publication date.
      No days limit is set by default; -1 disables the limit.
      The value can be specified by `--git-commit-strategy-expiry-days` option or `$WERF_GIT_COMMIT_STRATEGY_EXPIRY_DAYS`.
      * _git-commit-strategy-limit_.
      Keep the **specified max number** of published _images_ in the _images repo_.
      No limit is set by default; -1 disables the limit.
      Value can be specified by `--git-commit-strategy-limit` or `$WERF_GIT_COMMIT_STRATEGY_LIMIT`.
    * The policy covers images tagged by werf with the `--tag-git-commit` flag.
* **by tags:**
    * werf deletes an image from the _images repo_ when the corresponding git tag does not exist.
    * For the remaining images, the following policies apply:
       * _git-tag-strategy-expiry-days_.
       Keep published _images_ in the _images repo_ for the **specified maximum number of days** since the image was published.
       The republished image will be kept for the **specified maximum number of days** since the new publication date.
       No days limit is set by default; -1 disables the limit.
       Value can be specified by `--git-tag-strategy-expiry-days` option or `$WERF_GIT_TAG_STRATEGY_EXPIRY_DAYS`.
       * _git-tag-strategy-limit_.
       Keep the **specified max number** of published _images_ in the _images repo_.
       No limit is set by default; -1 disables the limit.
       Value can be specified by `--git-tag-strategy-limit` or `$WERF_GIT_TAG_STRATEGY_LIMIT`.
    * The policy covers images tagged by werf with `--tag-git-tag` flag.

**Please note** that cleanup affects only images built by werf **and** images tagged by werf with one of the following arguments: `--tag-git-branch`, `--tag-git-tag` or `--tag-git-commit`.
All other images in the _images repo_ stay intact.

#### Whitelisting images

The image always remains in the _images repo_ as long as the Kubernetes object that uses the image exists.
In the Kubernetes cluster, werf scans the following kinds of objects: `pod`, `deployment`, `replicaset`, `statefulset`, `daemonset`, `job`, `cronjob`, `replicationcontroller`.

The functionality can be disabled via the flag `--without-kube`.

#### Connecting to Kubernetes

werf uses the kube configuration file `~/.kube/config` to learn about Kubernetes clusters and ways to connect to them. werf connects to all Kubernetes clusters defined in all contexts of the kubectl configuration to gather information about the images that are in use.

### Cleaning up stages storage

Executing a [stages storage cleanup command]({{ site.baseurl }}/documentation/cli/management/stages/cleanup.html) is necessary to synchronize the state of stages storage with the _images repo_.
During this step, werf deletes _stages_ that do not relate to _images_ currently present in the _images repo_.

> If the [images cleanup command]({{ site.baseurl }}/documentation/cli/management/images/cleanup.html), — the first step of cleaning by policies, — is skipped, then the [stages storage cleanup]({{ site.baseurl }}/documentation/cli/management/stages/cleanup.html) will not have any effect.

## Manual cleaning

The manual cleaning approach assumes one-step cleaning with the complete removal of images from the _stages storage_ or _images repo_.
This method does not check whether the image is used by Kubernetes or not.
Manual cleaning is not recommended for the scheduled usage (use cleaning by policies instead).
In general, it is best suited for forceful image removal.

The manual cleaning approach includes the following options:

* The [purge images repo command]({{ site.baseurl }}/documentation/cli/management/images/purge.html) deletes images of the **current project** in the _images repo_.
* The [purge stages storage command]({{ site.baseurl }}/documentation/cli/management/stages/purge.html) deletes stages of the **current project** in the _stages storage_.

These steps are combined in a single top-level command [purge]({{ site.baseurl }}/documentation/cli/main/purge.html).

## Host cleaning

You can clean up the host machine with the following commands:

* The [cleanup host machine command]({{ site.baseurl }}/documentation/cli/management/host/cleanup.html) deletes the old unused werf cache and data for **all projects** on the host machine.
* The [purge host machine command]({{ site.baseurl }}/documentation/cli/management/host/purge.html) purges werf _images_, _stages_, cache, and other data for **all projects** on the host machine.
