---
title: Pull
sidebar: documentation
permalink: documentation/reference/registry/pull.html
---

werf does not have own pull command to download images from Docker registry. Regular [docker pull](https://docs.docker.com/engine/reference/commandline/pull/) should be used for such a task.

There is a need to use existing old images as a cache to build new images to speed up the builds. This is default behavior when there is a single werf build host with persistent local storage of images. No special actions required from the werf user in such case.

However, if local storage of build host is not persistent or there are multiple build hosts, then stages cache should be pulled before running a new build. This action allows sharing stages cache between an arbitrary number of build hosts no matter whether the storage of this hosts is persisted or not.

## Distributed images cache

It is not sufficient to pull just a image from the Docker registry. When you pull a image from a docker registry, you don't receive stages cache for this image in the werf world.

To enable stages cache sharing user must firstly push images _with a stages cache_ as it is described in [the push article]({{ site.baseurl }}/documentation/reference/registry/push.html).

Then [the werf pull command](#werf-pull) must be used to pull stages cache before the build command.

**Pay attention,** that this is an optional command. It _should not be used_ in the simple environment, where there is only a one build host with a persistent storage because it has an overhead on the build process speed in such case.

### Build process steps for a distributed build environment

1. Pull stages cache from the Docker registry with [werf pull](#werf-pull).
2. Build and push images with a new stages cache to the Docker registry with [werf push commands]({{ site.baseurl }}/documentation/reference/registry/push.html).

## Pull command

Command used to pull stages cache from the specified Docker registry. Call command before any of the build commands.

werf pull command optimized to pull only single cache stage for each image, that needed to rebuild _current state_ of the image.

For example, there was a change in `beforeSetup` stage of the image since last project build. werf pull command:

1. Calculates current state of the stage cache and realizes, that there is a change in `beforeSetup` stage.
2. Downloads `install` (the stage prior `beforeSetup`) stage from Docker registry. `beforeInstall` stage will not be downloaded, because only `install` stage is needed to rebuild `beforeSetup` and further stages.

A pulled stage can be used by multiple images of the same `werf.yaml` config in the case when this stage is common between multiple images.

In other words, werf downloads from cache **last common stage** between old and new image state.

There is also an option to turn off this optimized behavior and always pull all stages.

### Syntax

**NOT AVAILABLE** `werf pull` command currently not reimplemented from Ruby.

## Example

### Pull stages cache

```shell
werf pull registry.hello.com/taxi/backend
```

Command pull stages cache from the specified repo.

Pulled images have `image-stage` prefixes. Here is an example of image name pulled as stages cache:

* `registry.hello.com/taxi/backend:image-stage-ab192db1f7cf6b894aeaf14c0f1615f27d5170bb16b8529ec18253b94dc4916e`
