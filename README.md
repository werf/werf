<p align="center">
  <a href='https://bintray.com/flant/werf/werf/_latestVersion'><img src='https://api.bintray.com/packages/flant/werf/werf/images/download.svg'></a>
  <a href="https://travis-ci.org/flant/werf"><img alt="Build Status" src="https://travis-ci.org/flant/werf.svg" style="max-width:100%;"></a>
  <a href="https://godoc.org/github.com/flant/werf"><img src="https://godoc.org/github.com/flant/werf?status.svg" alt="GoDoc"></a>
</p>

___

Werf (previously known as Dapp) is made to implement and support Continuous Integration and Continuous Delivery (CI/CD).

It helps DevOps engineers generate and deploy images by linking together:

- application code (with Git support),
- infrastructure code (with Ansible or shell scripts), and
- platform as a service (Kubernetes).

Werf simplifies development of build scripts, reduces commit build time and automates deployment.
It is designed to make engineer's work fast end efficient.

**Contents**

- [Features](#features)
- [Requirements and Installation](#requirements-and-installation)
  - [Install Dependencies](#install-dependencies)
  - [Install werf](#install-werf)
- [Docs and Support](#docs-and-support)
- [License](#license)

# Features

* Complete application lifecycle management: **build** and **cleanup** images, **deploy** application into Kubernetes.
* **Incremental rebuilds** for **git**: reducing average build time for a sequence of git commits.
* Building images with **Ansible** or **Shell** scripts.
* Building **multiple images** from one description.
* Sharing a common cache between builds.
* Reducing image size by detaching source data and build tools.
* Running **distributed builds** with common registry.
* Advanced tools to debug built images.
* Tools for cleaning both local and remote Docker registry caches.
* Deploying to **Kubernetes** via [helm](https://helm.sh/), the Kubernetes package manager.

# Installation

## Install Dependencies

1. [Git command line utility](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git).

   Minimal required version is 1.9.0.

   To optionally use [Git Submodules](https://git-scm.com/docs/gitsubmodules) minimal version is 2.14.0.

2. Helm Kubernetes package manager. Helm is optional and only needed for deploy-related commands.

   [Helm client installation instructions.](https://docs.helm.sh/using_helm/#installing-helm)

   [Tiller backend installation instructions.](https://docs.helm.sh/using_helm/#installing-tiller)

   Minimal version is v2.7.0-rc1.

## Install Werf

### Using Multiwerf (recommended)

[Multiwerf](https://github.com/flant/multiwerf) is a version manager for Werf, which:
* downloads werf binary builds;
* manages multiple versions of binaries installed on a single host, that can be used at the same time;
* automatically updates werf binary (can be disabled).

### Download binary

The latest release can be reached via [this page](https://bintray.com/flant/werf/werf/_latestVersion).

##### MacOS

```bash
curl -L https://dl.bintray.com/flant/werf/v1.0.0-alpha.12/werf-darwin-amd64-v1.0.0-alpha.12 -o /tmp/werf
chmod +x /tmp/werf
sudo mv /tmp/werf /usr/local/bin/werf
```

##### Linux

```bash
curl -L https://dl.bintray.com/flant/werf/v1.0.0-alpha.12/werf-linux-amd64-v1.0.0-alpha.12 -o /tmp/werf
chmod +x /tmp/werf
sudo mv /tmp/werf /usr/local/bin/werf
```

##### Windows

Download [werf.exe](https://dl.bintray.com/flant/werf/v1.0.0-alpha.12/werf-windows-amd64-v1.0.0-alpha.12.exe).

### From source

```
go get github.com/flant/werf/cmd/werf
```
## Docs and support

[Official documentation](https://flant.github.io/werf/)

### Getting started

[Make your first werf application](https://flant.github.io/werf/how_to/getting_started.html)!
