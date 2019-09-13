---
title: Stapel
sidebar: documentation
permalink: documentation/development/stapel.html
author: Timofey Kirillov <timofey.kirillo@flant.com>
---

## Overview

Stapel is an [LFS](http://www.linuxfromscratch.org/lfs/view/stable) based linux distribution, which contains:

 * Glibc;
 * Gnu cli tools (install, patch, find, wget, grep, rsync and other);
 * Git cli util.
 * Bash.
 * Python interpreter and Ansible.

Stapel tools built in non-standard root location. Binaries, libraries and other related files are located in the following dirs:

 * `/.werf/stapel/(etc|lib|lib64|libexec)`;
 * `/.werf/stapel/x86_64-lfs-linux-gnu/bin`;
 * `/.werf/stapel/embedded/(bin|etc|lib|libexec|sbin|share|ssl)`.

The base of stapel is a Glibc library and linker (`/.werf/stapel/lib/libc-VERSION.so` and `/.werf/stapel/x86_64-lfs-linux-gnu/bin/ld`). All tools from `/.werf/stapel` are compiled and linked only agains libraries contained in `/.werf/stapel`. So stapel is a self-contained distribution of tools and libraries without external dependencies with an independent Glibc, which allows running these tools in arbitrary environment (independently of linux distribution and libraries versions in this distribution).

Stapel filesystem `/.werf/stapel` is intent to be mounted into build container based on some base image. Tools from stapel can then be used for some purposes. As stapel tools does not have external dependencies stapel image can be mounted into any base image (alpine linux with musl libc, or ubuntu with glibc — does not matter) and tools will work as expected.

Werf mounts _stapel image_ into each build container when building docker images with _stapel builder_ to enable ansible, git service operations and for other service purposes. More info about _stapel builder_ are available [in the article]({{ site.baseurl }}/documentation/reference/build_process.html#stapel-image-and-artifact).

## Rebuild stapel

Stapel image needs to be updated time to time to update ansible or when new version of [LFS](http://www.linuxfromscratch.org/lfs/view/stable) is available.

1. Make necessary changes to build instructions in `stapel` directory.
2. Update omnibus bundle:
    ```bash
    cd stapel
    bundle update
    git add -p Gemfile Gemfile.lock
    ```
3. Pull previous stapel images cache:
    ```bash
    scripts/stapel/pull_cache.sh
    ```
4. Increment current version in `scripts/stapel/version.sh` environment variable `CURRENT_STAPEL_VERSION`.
5. Build new stapel images:
    ```bash
    scripts/stapel/build.sh
    ```
6. Publish new stapel images:
    ```bash
    scripts/stapel/publish.sh
    ```
7. As new stapel version is done and published, then change `PREVIOUS_STAPEL_VERSION` environment variable in `scripts/stapel/version.sh` to the same value as `CURRENT_STAPEL_VERSION` and commit changes to the repo.

## How build works

Stapel image builder consists of 2 docker stages: base and final.

Base stage is an ubuntu based image used to build LFS system and additional omnibus packages.

Final stage is a pure scratch image which imports only necessary `/.werf/stapel` files.

NOTE: Both base and final docker stages are being published during `scripts/stapel/publish.sh` procedure to enable cache during next stapel rebuild.
