---
title: Overview
permalink: usage/build/overview.html
---

Delivering an application to Kubernetes involves containerizing it (building one or more images) for subsequent deployment in a cluster.

The build process is as follows: the user specifies the build instructions in the form of a Dockerfile or an alternative Stapel build syntax, and werf takes care of:

* orchestration of simultaneous/parallel building of application images;
* cross-platform and multi-platform image building;
* shared cache for intermediate layers and images in the container registry which is accessible from any runners;
* optimal tagging scheme based on image content which prevents unnecessary rebuilds and application waiting time during deployment;
* system of reproducibility and immutability of images for a commit: the images once built for a commit won't be re-built.

The paradigm for building and publishing images in werf differs from the one offered by the Docker builder. The latter has several distinct commands: `build`, `tag`, and `push`. On the other hand, werf builds, tags, and publishes images in a single step. This is due to the fact that werf does not build images, but rather synchronizes the current application state (for the current commit) with the container registry, building up the missing image layers and syncing the operation of parallel builders.

As a general rule, building images in werf implies the existence of a container registry (`--repo`), since the image is both built and published immediately. In this case, you do not need to invoke the `werf build` command manually, because the build is performed automatically when all top-level werf commands that require images (e.g., `werf converge`) are invoked.

However, manually running the `werf build` command may come in handy during local development. In this case, the build process can be run independently, and no container registry is needed.
