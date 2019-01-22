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

2. Helm Kubernetes package manager. Helm is optional and only needed for deploy-related commands.
   [Helm command line util installation instructions.](https://docs.helm.sh/using_helm/#installing-helm)
   [Tiller backend installation instructions.](https://docs.helm.sh/using_helm/#installing-tiller)
   Minimal version is v2.7.0-rc1.

## Install Werf binary (simple)

The latest release can be reached via [this page](https://bintray.com/dapp/dapp/Dapp/_latestVersion).

### MacOS

```bash
curl -L https://dl.bintray.com/dapp/dapp/v1.0.0-alpha.3/darwin-amd64/dapp -o /tmp/werf
chmod +x /tmp/werf
sudo mv /tmp/werf /usr/local/bin/werf
```

### Linux

```bash
curl -L https://dl.bintray.com/dapp/dapp/v1.0.0-alpha.3/linux-amd64/dapp -o /tmp/werf
chmod +x /tmp/werf
sudo mv /tmp/werf /usr/local/bin/werf
```

### Windows

Download [werf.exec](https://dl.bintray.com/dapp/dapp/v1.0.0-alpha.3/windows-amd64/dapp).

### Check it

Now you have Werf installed. Check it with `werf version`.

Time to [make your first application](https://flant.github.io/werf/how_to/getting_started.html)!

## Install Werf using Multiwerf

[Multiwerf](https://github.com/flant/multiwerf) is a version manager for Werf, which:
* Manages multiple versions of binaries installed on a single host, that can be used at the same time.
* Enables autoupdates (optionally).
