<p align="center">
  <img src="https://github.com/flant/werf/raw/master/logo.png" style="max-height:100%;" height="100">
</p>
<p align="center">
  <a href="https://badge.fury.io/rb/werf"><img alt="Gem Version" src="https://badge.fury.io/rb/werf.svg" style="max-width:100%;"></a>
  <a href="https://travis-ci.org/flant/werf"><img alt="Build Status" src="https://travis-ci.org/flant/werf.svg" style="max-width:100%;"></a>
  <a href="https://codeclimate.com/github/flant/werf"><img alt="Code Climate" src="https://codeclimate.com/github/flant/werf/badges/gpa.svg" style="max-width:100%;"></a>
</p>

___

Werf is made to implement and support Continuous Integration and Continuous Delivery (CI/CD).

It helps DevOps engineers generate and deploy images by linking together:

- application code (with Git support),
- infrastructure code (with Ansible or shell scripts), and
- platform as a service (Kubernetes).

Werf simplifies development of build scripts, reduces commit build time and automates deployment.
It is designed to make engineer's work fast end efficient.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Contents**

- [Features](#features)
- [Requirements and Installation](#requirements-and-installation)
  - [Install Dependencies](#install-dependencies)
  - [Install werf](#install-werf)
- [Docs and Support](#docs-and-support)
- [License](#license)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Features

* Reducing average build time.
* Sharing a common cache between builds.
* Running distributed builds with common registry.
* Reducing image size by detaching source data and build tools.
* Building images with Ansible and shell scripts.
* Building multiple images from one description.
* Advanced tools for debugging built images.
* Deploying to Kubernetes via [helm](https://helm.sh/), the Kubernetes package manager.
* Tools for cleaning both local and remote Docker registry caches.

# Requirements and Installation

Werf requires a Linux operating system.
Support for macOS is coming soon (see issue [#661](https://github.com/flant/werf/issues/661)).

## Install Dependencies

1.  Ruby version 2.1 or later:
    [Ruby installation](https://www.ruby-lang.org/en/documentation/installation/).

1.  Docker version 1.10.0 or later:
    [Docker installation](https://docs.docker.com/engine/installation/).

1.  сmake and pkg-config (required to install `rugged` gem):

    on Ubuntu:

    ```bash
    apt-get install cmake pkg-config
    ```

    on Centos:

    ```bash
    yum install cmake pkgconfig
    ```


1.  libssh2 header files to work with git via SSH.

    on Ubuntu:

    ```bash
    apt-get install libssh2-1-dev
    ```

    on Centos:

    ```bash
    yum install libssh2-devel
    ```

1.  libssl header files to work with git via HTTPS.

    on Ubuntu:

    ```bash
    apt-get install libssl-dev
    ```

    on Centos:

    ```bash
    yum install openssl-devel
    ```

1.  Git command line utility.

    Minimal required version is `1.9.0`. To use git submodules minimal version is `2.14.0`.

    on Ubuntu:

    ```bash
    apt-get install git
    ```

    on Centos:

    ```bash
    yum install git
    ```

## Install werf

  ```bash
  gem install werf
  ```

Now you have werf installed. Check it with `werf --version`.

Time to [make your first werf application](https://flant.github.io/werf/how_to/getting_started.html)!

# Docs and Support

The werf documentation is available at [flant.github.io/werf](https://flant.github.io/werf/).

You can ask for support in [werf chat in Telegram](https://t.me/werf_ru).

# License

Werf is published under Apache License v2.0.
See [LICENSE](https://github.com/flant/werf/blob/master/LICENSE) for details.
