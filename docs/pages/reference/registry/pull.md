---
title: Pull
sidebar: reference
permalink: reference/registry/pull.html
---

Dapp does not have own pull command to download dimg images from docker registry. Regular [`docker pull`](https://docs.docker.com/engine/reference/commandline/pull/) should be used for such a task.

There is a need to use existing old images as a cache to build new images to speed up the builds. This is default behavior when there is a single dapp build host with persistent local storage of images. No special actions required from the dapp user in such case.

However, if local storage of build host is not persistent or there are multiple build hosts, then stages cache should be pulled before running a new build. This action allows sharing stages cache between an arbitrary number of build hosts no matter whether the storage of this hosts is persisted or not.

## Distributed images cache

It is not sufficient to pull just a dimg from the docker registry. When you pull a dimg from a docker registry, you don't receive stages cache for this image in the Dapp world.

To enable stages cache sharing user must firstly push dimgs _with a stages cache_ as it is described in [the push article]({{ site.baseurl }}/reference/registry/push.html).

Then [the dapp pull command](#dapp-pull) must be used to pull stages cache before the build command.

**Pay attention,** that this is an optional command. It _should not be used_ in the simple environment, where there is only a one build host with a persistent storage because it has an overhead on the build process speed in such case.

### Build process steps for a distributed build environment

1. Pull stages cache from the docker registry with [dapp pull](#dapp-pull).
2. Build and push dimgs with a new stages cache to the docker registry with [dapp push commands]({{ site.baseurl }}/reference/registry/push.html).

## Dapp pull

Command used to pull stages cache from the specified docker registry. Must be called before any of the build commands.

### Syntax

```
dapp dimg stages pull [options] [DIMG ...] REPO
```

The `DIMG` optional parameter — is a name of dimg from a dappfile. Specifying `DIMG` one or multiple times allows pulling stages cache only related to certain dimgs from dappfile. By default, dapp pull stages cache of all dimgs from dappfile.

The `REPO` required parameter — is a repository name (see more in [image naming]({{ site.baseurl }}/reference/registry/image_naming.html#repo-parameter) article).

## Example

### Pull stages cache

```
dapp dimg stages pull registry.hello.com/taxi/backend
```

Command pull stages cache from the specified repo.

Pulled images have `dimgstage` prefixes. Here is an example of image names pulled as stages cache:

* `registry.hello.com/taxi/backend:dimgstage-ab192db1f7cf6b894aeaf14c0f1615f27d5170bb16b8529ec18253b94dc4916e`
* `registry.hello.com/taxi/backend:dimgstage-0d05ad73cf69430e8ff2bf457d6fd6273ec100785fcc3ad0267c0fb609c81a7c`
* `registry.hello.com/taxi/backend:dimgstage-41772c141b158349804ad27b354247df8984ead077a5dd601f3940536ebe9a11`
