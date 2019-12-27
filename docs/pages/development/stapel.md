---
title: Stapel
sidebar: documentation
permalink: documentation/development/stapel.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

## Overview

Stapel is an [LFS](http://www.linuxfromscratch.org/lfs/view/stable) based linux distribution, which contains:

 * Glibc
 * Gnu cli tools (install, patch, find, wget, grep, rsync and other)
 * Git cli util
 * Bash
 * Python interpreter
 * Ansible

Stapel tools built in non-standard root location. Binaries, libraries and other related files are located in the following directories:

 * `/.werf/stapel/(etc|lib|lib64|libexec)`
 * `/.werf/stapel/x86_64-lfs-linux-gnu/bin`
 * `/.werf/stapel/embedded/(bin|etc|lib|libexec|sbin|share|ssl)`

The base of stapel is a Glibc library and linker (`/.werf/stapel/lib/libc-VERSION.so` and `/.werf/stapel/x86_64-lfs-linux-gnu/bin/ld`). All tools from `/.werf/stapel` are compiled and linked only agains libraries contained in `/.werf/stapel`. So stapel is a self-contained distribution of tools and libraries without external dependencies with an independent Glibc, which allows running these tools in arbitrary environment (independently of linux distribution and libraries versions in this distribution).

Stapel filesystem `/.werf/stapel` is intent to be mounted into build container based on some base image. Tools from stapel can then be used for some purposes. As stapel tools does not have external dependencies stapel image can be mounted into any base image (alpine linux with musl libc, or ubuntu with glibc — does not matter) and tools will work as expected.

werf mounts _stapel image_ into each build container when building docker images with _stapel builder_ to enable ansible, git service operations and for other service purposes. More info about _stapel builder_ are available [in the article]({{ site.baseurl }}/documentation/reference/build_process.html#stapel-image-and-artifact).

## Change, update and rebuild stapel

Stapel image needs to be updated time to time to update ansible or when new version of [LFS](http://www.linuxfromscratch.org/lfs/view/stable) is available.

1.  Make necessary changes to build instructions in `stapel` directory.
2.  Update omnibus bundle:
    ```shell
    cd stapel/omnibus
    bundle update
    git add -p Gemfile Gemfile.lock
    ```
3.  Build new stapel images:
    ```shell
    scripts/stapel/build.sh
    ```
    This command will create `flant/werf-stapel:dev` and secondary `flant/werf-stapel-base:dev` docker images.
4.  To test this newly built stapel image export environment variable `WERF_STAPEL_IMAGE_VERSION=dev` prior running werf commands:
    ```shell
    export WERF_STAPEL_IMAGE_VERSION=dev
    werf build ...
    ```
5.  Publish new stapel images:
    ```shell
    scripts/stapel/publish.sh NEW_VERSION
    ```
6.  As new stapel version has been published change `VERSION` Go constant in the `pkg/stapel/stapel.go` to point to the new version and rebuild werf.
