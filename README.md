<p align="center">
  <img src="https://github.com/flant/dapp/raw/master/logo.png" style="max-height:100%;" height="100">
</p>
<p align="center">
  <a href="https://badge.fury.io/rb/dapp"><img alt="Gem Version" src="https://badge.fury.io/rb/dapp.svg" style="max-width:100%;"></a>
  <a href="https://travis-ci.org/flant/dapp"><img alt="Build Status" src="https://travis-ci.org/flant/dapp.svg" style="max-width:100%;"></a>
  <a href="https://codeclimate.com/github/flant/dapp"><img alt="Code Climate" src="https://codeclimate.com/github/flant/dapp/badges/gpa.svg" style="max-width:100%;"></a>
  <a href="https://codeclimate.com/github/flant/dapp/coverage"><img alt="Test Coverage" src="https://codeclimate.com/github/flant/dapp/badges/coverage.svg" style="max-width:100%;"></a>
</p>

___

Dapp is made to implement and support Continuous Integration and Continuous Delivery (CI/CD).

It helps DevOps engineers generate and deploy images by linking together:

- application code (with Git support),
- infrastructure code (with Ansible or shell scripts), and
- platform as a service (Kubernetes).

Dapp simplifies development of build scripts, reduces commit build time and automates deployment.
It is designed to make engineer's work fast end efficient.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Contents**

- [Features](#features)
- [Requirements and Installation](#requirements-and-installation)
  - [Install Dependencies](#install-dependencies)
  - [Install dapp](#install-dapp)
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

Dapp requires a Linux operating system.
Support for macOS is coming soon (see issue [#661](https://github.com/flant/dapp/issues/661)).

## Install Dependencies

1.  Ruby version 2.1 or later: 
    [Ruby installation](https://www.ruby-lang.org/en/documentation/installation/).

1.  Docker version 1.10.0 or later:
    [Docker installation](https://docs.docker.com/engine/installation/).    

1.  сmake (required to install `rugged` gem):

    on Ubuntu:

    ```bash
    apt-get install cmake
    ```

    on Centos:
    
    ```bash
    yum install cmake
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
      
## Install dapp

  ```bash
  gem install dapp
  ```

Now you have dapp installed. Check it with `dapp --version`.

Time to [make your first dapp application](https://flant.github.io/dapp/first_application.html)!

# Docs and Support

The dapp documentation is available at [flant.github.io/dapp](https://flant.github.io/dapp/).

You can ask for support in [dapp chat in Telegram](https://t.me/dapp_ru).

# License

Dapp is published under Apache License v2.0.
See [LICENSE](https://github.com/flant/dapp/blob/master/LICENSE) for details.
