<p align="center">
  <img src="https://github.com/flant/werf/raw/master/logo.png" style="max-height:100%;" height="100">
</p>
<p align="center">
  <a href='https://bintray.com/dapp/dapp/Dapp/_latestVersion'><img src='https://api.bintray.com/packages/dapp/dapp/Dapp/images/download.svg'></a>
  <a href="https://travis-ci.org/flant/werf"><img alt="Build Status" src="https://travis-ci.org/flant/werf.svg" style="max-width:100%;"></a>
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

* Comlete application lifecycle management: build and cleanup images, deploy application into Kubernetes.
* Reducing average build time for a sequence of git commits.
* Building images with Ansible and shell scripts.
* Building multiple images from one description.
* Sharing a common cache between builds.
* Reducing image size by detaching source data and build tools.
* Running distributed builds with common registry.
* Advanced tools for debugging built images.
* Tools for cleaning both local and remote Docker registry caches.
* Deploying to Kubernetes via [helm](https://helm.sh/), the Kubernetes package manager.

# Installation

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

# Docs and Support

The documentation is available at [flant.github.io/werf](https://flant.github.io/werf/).

You can ask for support [in Telegram chat](https://t.me/werf_ru).

# License

Werf is published under Apache License v2.0.
See [LICENSE](https://github.com/flant/werf/blob/master/LICENSE) for details.
