---
title: Cleanup
sidebar: documentation
permalink: documentation/advanced/cleanup.html
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

These steps are combined in the single top-level command [cleanup]({{ "documentation/reference/cli/werf_cleanup.html" | true_relative_url }}).

An _images repo_ is the primary source of information about current and stale images.
Therefore, it is essential to clean up the _images repo_ first and only then proceed to the _stages storage_.

### Cleaning up the images repo

werf can automate the cleaning of the _images repo_.
It works according to special rules called **cleanup policies**.
These policies determine which _images_ will be deleted while leaving all others intact.

#### Git history-based cleanup algorithm

The _end-image_ is the result of a building process. It can be associated with an arbitrary number of Docker tags.  The _end-image_ is linked to the werf internal identifier aka the stages [digest]({{ "/documentation/internals/stages_and_storage.html" | true_relative_url }}).

The cleanup algorithm is based on the fact that the [stages storage]({{ "documentation/internals/stages_and_storage.html#storage" | true_relative_url }}) has information about commits related to publishing tags associated with a specific [digest of image stages]({{ "/documentation/internals/stages_and_storage.html" | true_relative_url }}) (and it does not matter whether an image in the Docker registry was added, modified, or stayed the same). This information includes bundles of commit + _digest_ for a specific `image` in the `werf.yaml`.

Following the results of building and publishing of some commit, the end-image may stay unchanged. However, information about the publishing of a [digest of image stages]({{ "/documentation/internals/stages_and_storage.html" | true_relative_url }}) because of that commit will be added to the stages storage.

This ensures that the [digest of image stages]({{ "/documentation/internals/stages_and_storage.html" | true_relative_url }}) (of an arbitrary number of associated Docker tags) relates to the git history. Also, this opens up the possibility to effectively clean up outdated images based on the git state and [chosen policies](#custom-policies).  The algorithm scans the git history, selects relevant images, and deletes those that do not fall under any policy. At the same time, the [tags used in Kubernetes](#whitelisting-images) are ignored.

Let's review the basic steps of the cleanup algorithm:

- [Extracting the data required for a cleanup from the stages storage](#keeping-the-data-in-the-stages-storage-to-use-when-performing-a-cleanup):
    - all [names of the images]({{ "documentation/reference/werf_yaml.html#image-section" | true_relative_url }}) ever built;
    - a set of pairs consisting of the [digest of the image stages]({{ "/documentation/internals/stages_and_storage.html" | true_relative_url }}) and a commit on which the publication was performed.
- Obtaining manifests for all tags.
- Preparing a list of items to clean up:
    - [tags used in Kubernetes](#whitelisting-images) are ignored.
- Preparing the data for scanning:
    - tags grouped by the digest of image stages __(1)__;
    - commits grouped by the digest of image stages __(2)__;
    - a set of git tags and git branches, as well as the rules and crawl depth for scanning each reference based on [user policies](#custom-policies) __(3)__.
- Searching for commits __(2)__ using the git history __(3)__. The result is [digests of image stages]({{ "/documentation/internals/stages_and_storage.html" | true_relative_url }}) for which no associated commits were found during scanning __(4)__.
Deleting tags for [digests of image stages]({{ "/documentation/internals/stages_and_storage.html" | true_relative_url }}) __(4)__.

##### Custom policies

The user can specify images that will not be deleted during a cleanup using `keepPolicies` [cleanup policies]({{ "documentation/advanced/cleanup.html" | true_relative_url }}). If there is no configuration provided in the `werf.yaml`, werf will use the [default policy set]({{ "documentation/reference/werf_yaml.html#default-policies" | true_relative_url }}).

It is worth noting that the algorithm scans the local state of the git repository. Therefore, it is essential to keep all git branches and git tags up-to-date. You can use the `--git-history-synchronization` flag to synchronize the git state (it is enabled by default when running in CI systems).

##### Keeping the data in the stages storage to use when performing a cleanup

werf saves supplementary data to the [stages storage]({{ "documentation/internals/stages_and_storage.html#storage" | true_relative_url }}) to optimize its operation and solve some specific cases. This data includes meta-images with bundles consisting of a [digest of image stages]({{ "/documentation/internals/stages_and_storage.html" | true_relative_url }}) and a commit that was used for publishing. It also contains [names of images]({{ "documentation/reference/werf_yaml.html#image-section" | true_relative_url }}) that were ever built.

Information about commits is the only source of truth for the algorithm, so if tags lacking such information werf deletes them. 

When performing an automatic cleanup, the `werf cleanup` command is executed either on a schedule or manually. To avoid deleting the active cache when adding/deleting images in the `werf.yaml` in neighboring git branches, you can add the name of the image being built to the [stages storage]({{ "documentation/internals/stages_and_storage.html#storage" | true_relative_url }}) during the build. The user can edit the so-called set of _managed images_ using `werf managed-images ls|add|rm` commands.

#### Whitelisting images

The image always remains in the _images repo_ as long as the Kubernetes object that uses the image exists.
werf scans the following kinds of objects in the Kubernetes cluster: `pod`, `deployment`, `replicaset`, `statefulset`, `daemonset`, `job`, `cronjob`, `replicationcontroller`.

The functionality can be disabled via the flag `--without-kube`.

#### Connecting to Kubernetes

werf uses the kube configuration file `~/.kube/config` to learn about Kubernetes clusters and ways to connect to them. werf connects to all Kubernetes clusters defined in all contexts of the kubectl configuration to gather information about the images that are in use.

### Cleaning up stages storage

Executing a stages storage cleanup command is necessary to synchronize the state of stages storage with the _images repo_.
During this step, werf deletes _stages_ that do not relate to _images_ currently present in the _images repo_.

> If the images cleanup command, — the first step of cleaning by policies, — is skipped, then the stages storage cleanup will not have any effect.

## Manual cleaning

The manual cleaning approach assumes one-step cleaning with the complete removal of images from the _stages storage_ or _images repo_.
This method does not check whether the image is used by Kubernetes or not.
Manual cleaning is not recommended for the scheduled usage (use cleaning by policies instead).
In general, it is best suited for forceful image removal.

The manual cleaning approach includes the following options:

* The purge images repo command deletes images of the **current project** in the _images repo_.
* The purge stages storage command deletes stages of the **current project** in the _stages storage_.

These steps are combined in a single top-level command [purge]({{ "documentation/reference/cli/werf_purge.html" | true_relative_url }}).

## Host cleaning

You can clean up the host machine with the following commands:

* The [cleanup host machine command]({{ "documentation/reference/cli/werf_cleanup.html" | true_relative_url }}) deletes an obsolete non-used werf cache and data for **all projects** on the host machine.
* The [purge host machine command]({{ "documentation/reference/cli/werf_purge.html" | true_relative_url }}) purges werf _images_, _stages_, cache, and other data for **all projects** on the host machine.
