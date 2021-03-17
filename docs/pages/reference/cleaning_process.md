---
title: Cleaning process
sidebar: documentation
permalink: reference/cleaning_process.html
author: Artem Kladov <artem.kladov@flant.com>, Timofey Kirillov <timofey.kirillov@flant.com>
---

During building and publishing, werf creates sets of docker layers, however, it does not delete them.
As a result, _stages storage_ and _images repo_ are growing steadily and consuming more and more space.
What is more, any interrupted build process leaves stalled images.
When a git branch or a git tag gets deleted, a set of _stages_ that were built for this _image_ remains in the _images repo_ and _stages storage_.
So, it is necessary to clean up the _images repo_ and _stages storage_ periodically. Otherwise, 
**stale images** will pile up.

werf has a built-in efficient multi-level image cleaning system. It supports the following cleaning methods:

1. [**Cleaning by policies**](#cleaning-by-policies)
2. [**Manual cleaning**](#manual-cleaning)
3. [**Host cleaning**](#host-cleaning)

## Cleaning by policies

The cleaning by policies method allows you to organize the regular automatic cleanup of stuck images.
It implies regular gradual cleaning according to cleaning policies.
This is the safest method of cleaning because it does not affect your production environment.

The cleaning by policies method involves the following steps:
1. [**Cleaning up the images repo**](#cleaning-up-the-images-repo) deletes stale images in the _images repo_ according to the cleaning policies.
2. [**Cleaning up stages storage**](#cleaning-up-stages-storage) synchronizes the _stages storage_ with the _images repo_.

These steps are combined in the single top-level command [cleanup]({{ site.baseurl }}/cli/main/cleanup.html).

An _images repo_ is the primary source of information about current and stale images.
Therefore, it is essential to clean up the _images repo_ first and only then proceed to the _stages storage_.

### Cleaning up the images repo

werf can automate the cleaning of the _images repo_.
It works according to special rules called **cleanup policies**.
These policies determine which _images_ will be deleted while leaving all others intact.

#### Tagging scheme-based cleanup algorithm

> In versions prior to `v1.2`, the cleanup policy is based on [tagging schemes](#tagging-scheme-based-cleanup-algorithm). The [git history-based cleanup algorithm](#git-history-based-cleanup-algorithm) is going to be the only one available starting with version `v1.2`. You can use the `--git-history-based-cleanup` flag to forcefully enable the new algorithm.

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
      The republished image will be kept for the **specified maximum number of days** since the new publication date.
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

**Please note** that the cleanup affects only images built and published by werf with one of the following arguments: `--tag-git-branch`, `--tag-git-tag` or `--tag-git-commit`.
All other images in the _images repo_ stay intact.

#### Git history-based cleanup algorithm

> In versions prior to `v1.2`, the cleanup policy is based on [tagging schemes](#tagging-scheme-based-cleanup-algorithm). The [git history-based cleanup algorithm](#git-history-based-cleanup-algorithm) is going to be the only one available starting with version `v1.2`. You can use the `--git-history-based-cleanup` flag to forcefully enable the new algorithm.

The _end-image_ is the result of a building process. It can be associated with an arbitrary number of Docker tags.  The _end-image_ is linked to the werf internal identifier aka the stages [signature]({{ site.baseurl }}/reference/stages_and_images.html#image-stages-signature).

The cleanup algorithm is based on the fact that the [stages storage]({{ site.baseurl }}/reference/stages_and_images.html#stages-storage) has information about commits related to publishing tags associated with a specific [signature of image stages]({{ site.baseurl }}/reference/stages_and_images.html#image-stages-signature) (and it does not matter whether an image in the Docker registry was added, modified, or stayed the same). This information includes bundles of commit + _signature_ for a specific `image` in the `werf.yaml`.

Following the results of building and publishing of some commit, the end-image may stay unchanged. However, information about the publishing of a [signature of image stages]({{ site.baseurl }}/reference/stages_and_images.html#image-stages-signature) because of that commit will be added to the stages storage.

This ensures that the [signature of image stages]({{ site.baseurl }}/reference/stages_and_images.html#image-stages-signature) (of an arbitrary number of associated Docker tags) relates to the git history. Also, this opens up the possibility to effectively clean up outdated images based on the git state and [chosen policies](#custom-policies).  The algorithm scans the git history, selects relevant images, and deletes those that do not fall under any policy. At the same time, the [tags used in Kubernetes](#whitelisting-images) are ignored.

Let's review the basic steps of the cleanup algorithm:

- [Extracting the data required for a cleanup from the stages storage](#keeping-the-data-in-the-stages-storage-to-use-when-performing-a-cleanup):
    - all [names of the images]({{ site.baseurl }}/configuration/stapel_image/naming.html) ever built;
    - a set of pairs consisting of the [signature of the image stages]({{ site.baseurl }}/reference/stages_and_images.html#image-stages-signature) and a commit on which the publication was performed.
- Obtaining manifests for all tags.
- Preparing a list of items to clean up:
    - [tags used in Kubernetes](#whitelisting-images) are ignored;
    - tags assigned by versions `<v1.1.20` are ignored (and they are deleted starting with `v1.2`; you can force this behaviour by providing the  `--git-history-based-cleanup-v1.2` option).
- Preparing the data for scanning:
    - tags grouped by the signature of image stages __(1)__;
    - commits grouped by the signature of image stages __(2)__;
    - a set of git tags and git branches, as well as the rules and crawl depth for scanning each reference based on [user policies](#custom-policies) __(3)__.
- Searching for commits __(2)__ using the git history __(3)__. The result is [signatures of image stages]({{ site.baseurl }}/reference/stages_and_images.html#image-stages-signature) for which no associated commits were found during scanning __(4)__.
Deleting tags for [signatures of image stages]({{ site.baseurl }}/reference/stages_and_images.html#image-stages-signature) __(4)__.

##### Custom policies

The user can specify images that will not be deleted during a cleanup using `keepPolicies` [cleanup policies]({{ site.baseurl }}/configuration/cleanup.html). If there is no configuration provided in the `werf.yaml`, werf will use the [default policy set]({{ site.baseurl }}/configuration/cleanup.html#default-policies).

It is worth noting that the algorithm scans the local state of the git repository. Therefore, it is essential to keep all git branches and git tags up-to-date. You can use the `--git-history-synchronization` flag to synchronize the git state (it is enabled by default when running in CI systems).

##### Keeping the data in the stages storage to use when performing a cleanup

werf saves supplementary data to the [stages storage]({{ site.baseurl }}/reference/stages_and_images.html#stages-storage) to optimize its operation and solve some specific cases. This data includes meta-images with bundles consisting of a [signature of image stages]({{ site.baseurl }}/reference/stages_and_images.html#image-stages-signature) and a commit that was used for publishing. It also contains [names of images]({{ site.baseurl }}/configuration/stapel_image/naming.html) that were ever built.

Information about commits is the only source of truth for the algorithm, so tags lacking such information are processed separately. Tags assigned by versions lower than `v1.1.20` are ignored (and they are deleted starting with version `v1.2`; you can force this behavior by using the `--git-history-based-cleanup-v1.2` flag).

When performing an automatic cleanup, the `werf cleanup` command is executed either on a schedule or manually. To avoid deleting the active cache when adding/deleting images in the `werf.yaml` in neighboring git branches, you can add the name of the image being built to the [stages storage]({{ site.baseurl }}/reference/stages_and_images.html#stages-storage) during the build. The user can edit the so-called set of _managed images_ using `werf managed-images ls|add|rm` commands.

#### Whitelisting images

The image always remains in the _images repo_ as long as the Kubernetes object that uses the image exists.
werf scans the following kinds of objects in the Kubernetes cluster: `pod`, `deployment`, `replicaset`, `statefulset`, `daemonset`, `job`, `cronjob`, `replicationcontroller`.

The functionality can be disabled via the flag `--without-kube`.

#### Connecting to Kubernetes

werf uses the kube configuration file `~/.kube/config` to learn about Kubernetes clusters and ways to connect to them. werf connects to all Kubernetes clusters defined in all contexts of the kubectl configuration to gather information about the images that are in use.

### Cleaning up stages storage

Executing a [stages storage cleanup command]({{ site.baseurl }}/cli/management/stages/cleanup.html) is necessary to synchronize the state of stages storage with the _images repo_.
During this step, werf deletes _stages_ that do not relate to _images_ currently present in the _images repo_.

> If the [images cleanup command]({{ site.baseurl }}/cli/management/images/cleanup.html), — the first step of cleaning by policies, — is skipped, then the [stages storage cleanup]({{ site.baseurl }}/cli/management/stages/cleanup.html) will not have any effect.

## Manual cleaning

The manual cleaning approach assumes one-step cleaning with the complete removal of images from the _stages storage_ or _images repo_.
This method does not check whether the image is used by Kubernetes or not.
Manual cleaning is not recommended for the scheduled usage (use cleaning by policies instead).
In general, it is best suited for forceful image removal.

The manual cleaning approach includes the following options:

* The [purge images repo command]({{ site.baseurl }}/cli/management/images/purge.html) deletes images of the **current project** in the _images repo_.
* The [purge stages storage command]({{ site.baseurl }}/cli/management/stages/purge.html) deletes stages of the **current project** in the _stages storage_.

These steps are combined in a single top-level command [purge]({{ site.baseurl }}/cli/main/purge.html).

## Host cleaning

You can clean up the host machine with the following commands:

* The [cleanup host machine command]({{ site.baseurl }}/cli/management/host/cleanup.html) deletes an obsolete non-used werf cache and data for **all projects** on the host machine.
* The [purge host machine command]({{ site.baseurl }}/cli/management/host/purge.html) purges werf _images_, _stages_, cache, and other data for **all projects** on the host machine.
