---
title: Cache
sidebar: reference
permalink: reference/build/cache.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
---

Dapp uses multi-layer high-performance caching that significantly accelerates assembling and reduces the size of the resulting image.

Dapp provides a directory for storing the project's assembly cache — `~/.dapp/builds/<project name>`. This directory may store the following data:
* Remote git repositories. Dapp runs _git clone_ on the first build and then only retrieves changes from the remote git repository by _git fetch_. Benefits of this approach are clear, especially when git repository is heavy and there are connection speed and traffic limits (for detailed information about working with git repositories see the [corresponding article]({{ site.baseurl }}/reference/build/git_directive.html)).
* _build_dir_ mounting directories (for detailed information see the article dedicated to [mounting]({{ site.baseurl }}/reference/build/mount_directive.html)).

The most exciting thing is the _stages cache_ performing the same functions as docker layers when using Dockerfile. Further, we consider the principle of organizing _stages cache_.

## Creating _stages cache_

During assembly, docker images of the assembled stages remain "invisible" for a dapp user, and those images only have a temporary ID in docker. Technically speaking, a docker image would get to the _stages cache_ precisely at the moment when dapp tags the image in a particular way: using `dimgstage-<project name>` as the image name, and a _stage signature_ as the image tag.

Dapp only saves the assembled layers to the _stages cache_ after all stages are successfully assembled. If during assembly of a particular stage an error occurs, then all stages that were successfully assembled by that point are lost. Rerunning the assembly operation starts from the same _stage_ from which the previous assembly operation started.

This caching scheme is required to ensure strict correctness of the saved cache.

## Forced saving images to cache after assembling

For configuration developers, it would be much more convenient if all successfully assembled stages were saved to the cache of the docker images. In this case, if an error is thrown, re-assembling would always start from the erroneous _stage_.

For this purpose, dapp provides the forced cache saving option, which is enabled either by the `--force-save-cache` option or by the presence of the `WERF_FORCE_SAVE_CACHE=1` environment variable.

For example, dappfile:

```yaml
dimg: ~
from: ubuntu:16.04
shell:
  beforeInstall:
  - apt-get update
  - apt-get install -y curl --quiet
  install:
  - apt-get install -y non-existing
```

We run an assembly:

```bash
$ dapp dimg build --force-save-cache
nameless: calculating stages signatures                                     [RUNNING]
nameless: calculating stages signatures                                          [OK] 0.16 sec
From ...                                                                         [OK] 1.09 sec
  signature: dimgstage-myapp2:41772c141b158349804ad27b354247df8984ead077a5dd601f3940536ebe9a11
Before install                                                             [BUILDING]
Get:1 http://archive.ubuntu.com/ubuntu xenial InRelease [247 kB]
Get:2 http://archive.ubuntu.com/ubuntu xenial-updates InRelease [109 kB]
...
done.
Before install                                                                   [OK] 22.85 sec
  signature: dimgstage-myapp2:8f8307adb4d2434822cdbb44950868b1a312d1a0e536ae54debff9640f371645
  commands:
    apt-get update
    apt-get install -y curl --quiet
Install group
  Install                                                                  [BUILDING]
Reading package lists...
Building dependency tree...
Reading state information...
E: Unable to locate package non-existing
  Install                                                                    [FAILED] 2.03 sec
    signature: dimgstage-myapp2:1c0aca95f86933173709388f4f75cdc50e210a861d3e85193f14556bf4a798f8
    commands:
      apt-get install -y non-existing
Running time 28.02 seconds
Stacktrace dumped to /tmp/dapp-stacktrace-f9333e01-c9b9-4f31-809a-12ada6f7c64d.out
ruby2go_image command `build` failed!
```

When running again, the _before_install stage_ will not be re-assembled any longer, because it was cached during the first launch.

```shell
$ dapp dimg build --force-save-cache
nameless: calculating stages signatures                                     [RUNNING]
nameless: calculating stages signatures                                          [OK] 0.16 sec
From                                                                    [USING CACHE]
  signature: dimgstage-myapp2:41772c141b158349804ad27b354247df8984ead077a5dd601f3940536ebe9a11
  date: 2018-08-29 19:07:02 +0300
  size: 113.913 MB
Before install                                                          [USING CACHE]
  signature: dimgstage-myapp2:8f8307adb4d2434822cdbb44950868b1a312d1a0e536ae54debff9640f371645
  date: 2018-08-29 19:07:24 +0300
  difference: 57.087 MB
Install group
  Install                                                                  [BUILDING]
Reading package lists...
Building dependency tree...
Reading state information...
E: Unable to locate package non-existing
  Install                                                                    [FAILED] 2.03 sec
    signature: dimgstage-myapp2:1c0aca95f86933173709388f4f75cdc50e210a861d3e85193f14556bf4a798f8
    commands:
      apt-get install -y non-existing
Running time 3.85 seconds
Stacktrace dumped to /tmp/dapp-stacktrace-9adbb391-79e4-421f-83b1-dcdad372051c.out
ruby2go_image command `build` failed!
```

## Why does dapp not save the cache of erroneous assemblies by default?

`WERF_FORCE_SAVE_CACHE` operating mode may result in an invalid cache being created. In this case only removing the erroneous cache manually can help.

We should consider an example to understand how an invalid cache could be saved.

Let us initiate the application with a standard dappfile:

```yaml
dimg: ~
from: ubuntu:16.04
git:
- add: /
  to: /app
```

Assembling:

```bash
$ dapp dimg build --force-save-cache
nameless: calculating stages signatures                                     [RUNNING]
  Repository `own`: latest commit `3d70fcec74abf7b8197230830bb6d7ccf5826952` to `/app`
nameless: calculating stages signatures                                          [OK] 0.24 sec
From ...                                                                         [OK] 1.56 sec
  signature: dimgstage-myapp:41772c141b158349804ad27b354247df8984ead077a5dd601f3940536ebe9a11
Git artifacts: create archive ...                                                [OK] 1.37 sec
  signature: dimgstage-myapp:d1aa6029faae81733618867c217c9e0e9d70e56ab1fc2554e790d9b14f16b96c
Setup group
  Git artifacts: apply patches (after setup) ...                                 [OK] 1.39 sec
    signature: dimgstage-myapp:336636cedd354d7903d71d242b4a8c40dd0bf81728b0e189deee26cd1d59ec6b
Running time 13.9 seconds
```

Assembling has been successfully finished, and _stages cache_ is filled with valid _stages_ images. 

Then we add an assembly instruction that uses a file from git. However, we intentionally made a mistake in this instruction — we are trying to copy a `/app/hello` file that is not present in git. For example, the user may have forgotten to add it.

```yaml
dimg: ~
from: ubuntu:16.04
git:
- add: /
  to: /app
shell:
  install:
  - cp /app/hello /hello
```

Assembling with this dappfile throws an error:

```shell
$ dapp dimg build --force-save-cache
nameless: calculating stages signatures                                     [RUNNING]
  Repository `own`: latest commit `895f42cd25d025018c00ad5ac6fe88764cfca980` to `/app`
nameless: calculating stages signatures                                          [OK] 0.33 sec
From                                                                    [USING CACHE]
  signature: dimgstage-myapp:41772c141b158349804ad27b354247df8984ead077a5dd601f3940536ebe9a11
  date: 2018-08-29 18:22:07 +0300
  size: 113.913 MB
Git artifacts: create archive                                           [USING CACHE]
  signature: dimgstage-myapp:d1aa6029faae81733618867c217c9e0e9d70e56ab1fc2554e790d9b14f16b96c
  date: 2018-08-29 18:22:08 +0300
  difference: 0.0 MB
Install group
  Git artifacts: apply patches (before install) ...                              [OK] 1.47 sec
    signature: dimgstage-myapp:3a4b24a524f72e259bc8e5d6335ca7aaa4504d08da9d63d31c42df92331fd24d
  Install                                                                  [BUILDING]
cp: cannot stat '/app/hello': No such file or directory
    Launched command: `cp /app/hello /hello`
  Install                                                                    [FAILED] 1.43 sec
    signature: dimgstage-myapp:003e8da0e54baddc3ebc5e499fdd29d1af4dbd88626a9606d9dc32df725b433e
    commands:
      cp /app/hello /hello
Running time 5.01 seconds
Stacktrace dumped to /tmp/dapp-stacktrace-38fa7ded-c542-4fef-9f1f-5cf6cae662f9.out
>>> START STREAM
cp: cannot stat '/app/hello': No such file or directory
>>> END STREAM
```

During assembly, dapp notices adding assembly instructions to _install stage_ and builds it. Before _install stage_ assembly, dapp builds _git_pre_install_patch stage_ (`Git artifacts: apply patches (before install)`). This approach is needed to use actual condition of git repository during _install stage_ assembly (read more about it in [separate article]({{ site.baseurl }}/reference/build/assembly_instructions.html)).

We add the `hello` file to git repository to fix the error. Then run re-assembly and see the same error: there is no `hello` file.

```shell
$ dapp dimg build --force-save-cache
nameless: calculating stages signatures                                     [RUNNING]
  Repository `own`: latest commit `a6d7b54cd8055df635475c7e9972237a0974142b` to `/app`
nameless: calculating stages signatures                                          [OK] 0.4 sec
From                                                                    [USING CACHE]
  signature: dimgstage-myapp:41772c141b158349804ad27b354247df8984ead077a5dd601f3940536ebe9a11
  date: 2018-08-29 18:22:07 +0300
  size: 113.913 MB
Git artifacts: create archive                                           [USING CACHE]
  signature: dimgstage-myapp:d1aa6029faae81733618867c217c9e0e9d70e56ab1fc2554e790d9b14f16b96c
  date: 2018-08-29 18:22:08 +0300
  difference: 0.0 MB
Install group
  Git artifacts: apply patches (before install)                         [USING CACHE]
    signature: dimgstage-myapp:3a4b24a524f72e259bc8e5d6335ca7aaa4504d08da9d63d31c42df92331fd24d
    date: 2018-08-29 18:35:51 +0300
    difference: 0.0 MB
  Install                                                                  [BUILDING]
cp: cannot stat '/app/hello': No such file or directory
    Launched command: `cp /app/hello /hello`
  Install                                                                    [FAILED] 1.25 sec
    signature: dimgstage-myapp:003e8da0e54baddc3ebc5e499fdd29d1af4dbd88626a9606d9dc32df725b433e
    commands:
      cp /app/hello /hello
Running time 2.07 seconds
Stacktrace dumped to /tmp/dapp-stacktrace-10b05694-bdc5-463c-8abb-3748b20d5acb.out
>>> START STREAM
cp: cannot stat '/app/hello': No such file or directory
>>> END STREAM
```

This file was meant to be added to the image at the _g_a_pre_install_patch stage_; however, this _stage_ was cached at the moment when this file was not yet available in the git repository. To correct this, you should manually remove this _stage cache_.

This is the feature of dapp cache. The _signature_ of this _stage_ cannot depend on random files in git, otherwise caching _install_, _before setup_ and _setup_ _stages_ makes no sense. So, if the _signature_ is not changed, the cache is not changed either.

These issues don't appear if the cache for stages is only performed after successfully finished assembly. This caching scheme dapp uses by default.

So if you decide to use `WERF_FORCE_SAVE_CACHE` option, be prepared for situations like this, use the option carefully, and preferably only use it during configuration debugging.
