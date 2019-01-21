---
title: Installation
sidebar: how_to
permalink: how_to/installation.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

## Install Dependencies

1. [Git command line utility](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git).
   Minimal required version is 1.9.0.
   To optionally use [Git Submodules](https://git-scm.com/docs/gitsubmodules) minimal version is 2.14.0.

2. [Helm Kubernetes package manager](https://github.com/helm/helm/blob/master/docs/install.md)
   Helm is optional and needed only for deploy-related commands.
   Minimal version is v2.7.0-rc1.

## Install Werf binary (simple)

1. [Download latest release](https://bintray.com/dapp/dapp/Dapp/_latestVersion) for your architecture. Linux, MacOS and Windows are supported.

   For example to download version v1.0.0-alpha.3 on MacOS:

   ```bash
   curl -L https://dl.bintray.com/dapp/dapp/v1.0.0-alpha.3/darwin-amd64/dapp -o werf
   ```

2. Make binary executable:

    ```bash
    chmod +x ./werf
    ```

3. Move binary to your PATH

   ```bash
   sudo mv ./werf /usr/local/bin/werf
   ```

Now you have Werf installed. Check it with `werf version`.

Time to [make your first application](https://flant.github.io/werf/how_to/getting_started.html)!

## Install Werf using Multiwerf

[Multiwerf](https://github.com/flant/multiwerf) is a version manager for Werf.

* Manages multiple versions of binaries installed on a single host, that can be used at the same time.
* Enables autoupdates (optionally).
