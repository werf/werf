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

   [Helm client installation instructions.](https://docs.helm.sh/using_helm/#installing-helm)

   [Tiller backend installation instructions.](https://docs.helm.sh/using_helm/#installing-tiller)

   Minimal version is v2.7.0-rc1.

## Install Werf

### Way 1 (recommended): using Multiwerf

[Multiwerf](https://github.com/flant/multiwerf) is a version manager for Werf, which:
* downloads werf binary builds;
* manages multiple versions of binaries installed on a single host, that can be used at the same time;
* automatically updates werf binary (can be disabled).

```
mkdir ~/bin
cd ~/bin
curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash
source <(multiwerf use 1.0 beta)
```

### Way 2: download binary

The latest release can be reached via [this page](https://bintray.com/flant/werf/werf/_latestVersion).

##### MacOS

```bash
curl -L https://dl.bintray.com/flant/werf/v1.0.0-beta.1/werf-darwin-amd64-v1.0.0-beta.1 -o /tmp/werf
chmod +x /tmp/werf
sudo mv /tmp/werf /usr/local/bin/werf
```

##### Linux

```bash
curl -L https://dl.bintray.com/flant/werf/v1.0.0-beta.1/werf-linux-amd64-v1.0.0-beta.1 -o /tmp/werf
chmod +x /tmp/werf
sudo mv /tmp/werf /usr/local/bin/werf
```

##### Windows

Download [werf.exe](https://dl.bintray.com/flant/werf/v1.0.0-beta.1/werf-windows-amd64-v1.0.0-beta.1.exe).

### Way 3: from source

```
go get github.com/flant/werf/cmd/werf
```

## What's next?

Time to [make your first werf application]({{ site.baseurl }}/how_to/getting_started.html)!
