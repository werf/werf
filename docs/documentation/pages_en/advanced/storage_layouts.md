---
title: Storage layouts
permalink: advanced/storage_layouts.html
---

This article describes werf's image storage, discusses the available storage types and their features, and provides various ways of implementing storage for your application.

## Stage Storage

werf builds images that consist of one or several stages. All stages needed to build an image are stored in the so-called **stage storage** as they are assembled.

There are two kinds of stage storage: **local** and **container registry**.

* You can only use the local stage storage (it is based on the Docker server) for _local development_. Currently, it only supports image assembly. The local stage storage will be used if you, e.g., invoke the `werf build` command without the `--repo` argument.

* In all other cases (i.e., excluding local development), werf uses the repository in the container registry as the stage storage. If you run a build with the repository-based stage storage as an argument (e.g., `werf build --repo ghcr.io/example/myrepo`), werf will first check for stages in the local repository and use them to avoid rebuilding the existing stages from scratch.

Before continuing, let's introduce the concept of **project installation**. The project installation consists of the project's Git repository and its associated stage storages in the repository. All werf invocations are supposed to use the same Git repository and stage storages *(you can use one or more storages - see below for more details)*.

Note that any attempt to delete or clean up the installation's stage storage using some external tool will negatively affect werf operations *(see the [primary repository](#primary-repository))*. However, you can safely delete specific storage types at any time *(see the [caching repository](#caching-repository))*.

### Primary repository

* You can specify it via the `--repo` parameter (or `WERF_REPO` environment variable).
* Each project features only one primary repository.

As the name suggests, this storage is considered the main one. werf uses it for:

1. Keeping the history of the project:
    - It serves as the Git repo metadata storage for building images and stages.
    - The project history makes it possible to implement an advanced algorithm for safely cleaning up obsolete stages based on the Git history (the `werf cleanup` command).
2. Synchronizing an assembly in a distributed environment:
    - With the primary repository, you can run an assembly in a distributed environment and reuse the assembled stages in that environment. For that, werf provides built-in mechanisms for synchronizing distributed processes. These mechanisms can only work with the primary repository.
    - Once a stage becomes available in the primary repository, other assembly processes running on any arbitrary host can reuse it.
3. Storing an assembly cache of previously built stages:
    - Previously built stages are reused if possible.
    - This feature is similar to the Docker server's local build cache; however, in the case of werf, the cache is stored in the container registry.
4. Storing the final images (those are used to run applications in Kubernetes).
    - Kubernetes will pull images from the primary repository to run the application containers.
    - The same stages that make up the assembled image are used as the final images.

**Cleaning (CAUTION!).** The primary repository must be persistent for werf to work properly; images this repo can only be deleted via a dedicated `werf cleanup` command. The primary repository serves as the build cache and keeps the project's history (similar to the commit history in the Git repo).

In addition to the primary repository, werf features other repository types that can handle some of its functions. We will discuss them below. Note that the **primary repository is mandatory** (werf cannot work without it) while all **other repos are optional**; they were created to replace the primary repository in some activities in the project installation.

### Secondary repository

* You can specify it via the `--secondary-repo` parameter (or `WERF_SECONDARY_REPO_<NAME>` environment variable).
* There can be one or more secondary repos.

This read-only repository acts as an assembly cache for previously built stages *(see the [primary repository](#primary-repository) features)*.

Usually, the existing primary repository from another project installation is used as a secondary repo. Missing stages will be copied from the secondary repository to the primary one if needed (thus, you do not have to build them from scratch). On the other hand, werf will push newly built stages to the primary repository because the secondary repo is read-only (i.e., you cannot write new data into it).

You do not need to **clean up** the secondary repo since it acts as the primary one for another project installation and, thus, werf manages it as part of that installation.

### Caching repository

* You can specify it via the `--cache-repo` parameter (or `WERF_SECONDARY_REPO_<NAME>` environment variable).
* There can be one or more secondary repos.

This one stores the previously assembled stages *(see the [primary repository](#primary-repository) features)*. Unlike a secondary repository, a caching one supports reading and writing. Thus, werf can write newly built stages to the cache and use the stages from that cache.

The existing stages will be copied from the caching repository to the primary one since the primary repo must contain all the built stages of the project images. All newly built stages will be pushed to both primary and caching repositories.

werf will pull stages that serve as a basis for assembling new stages from the cache. However, if they are not in the caching repository, werf will pull them from the primary repo and push to the caching one for further use.

In other words, the caching repository gets populated with newly built stages and the stages used to assemble new stages.

Note that the caching repository is never used for running an application in Kubernetes.

You can **clean up** the caching repository by deleting it. After cleaning up, the repository will be re-populated with up-to-date frequently used data.

### Final Repository

* You can set it via the `--final-repo` parameter (or `WERF_FINAL_REPO` environment variable).
* Only one such repository can be specified.

This repository can only share the final images used to deploy the application to Kubernetes.

  - werf pushes newly built images to the primary repo first and then copies them to the final one.
  - The final repository only keeps the final image stage (no intermediate stages).
  - The final repository never stores intermediate images (artifacts) required to build the final image. Only the final images used in Kubernetes go there.

You can clean up the final repo together with the primary one using the `werf cleanup` command (you need to provide two arguments: `--repo` and `--final-repo`).

## Examples of storage layouts

Below are some typical use cases and combinations of the repositories discussed. Note that you have to set the [primary repo](#primary-repository) in all cases.

### 1. Standard

* **The primary** repository is accessible from all assembly hosts and the Kubernetes cluster.
* When building images, the build cache will be pulled and pushed to this repository. Kubernetes will pull application images from this repo when deploying an application.
* This scheme is suitable for most cases and recommended if you are not sure what you need.

### 2. Additional cache in the local network

* The **primary repository** is accessible from all assembly hosts and the Kubernetes cluster.
* The **caching repository** runs in a high-speed local network.
* When building images, werf pushes the assembly cache to both the primary and caching repositories. The image stages required to build other stages will be pulled from the caching repo.

### 3. Optimized scheme

* The **primary repository** runs in a high-speed local network and is accessible from all assembly hosts (no need to access this repo from a Kubernetes cluster). Unlike a caching repository, any attempt to clean up a primary repository or delete it will affect werf operation. Thus, you must ensure that this repo is persistent and reliable.
* You can create additional **caching repositories** in a high-speed local network.
* The **final repository** must be located close to the Kubernetes cluster to enable rapid pulling of the application images (Kubernetes pulls final images way more often than the build hosts push them to the repo). Note that all assembly hosts require write access to the final repository. The final repository only stores the final, ready-to-run application images (no intermediate artifact images needed to build other images).
