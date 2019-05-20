<p align="center">
  <img src="https://github.com/flant/werf/raw/master/docs/images/werf-logo.svg?sanitize=true" style="max-height:100%;" height="175">
</p>

<p align="center">
  <a href='https://bintray.com/flant/werf/werf/_latestVersion'><img src='https://api.bintray.com/packages/flant/werf/werf/images/download.svg'></a>
  <a href="https://travis-ci.org/flant/werf"><img alt="Build Status" src="https://travis-ci.org/flant/werf.svg" style="max-width:100%;"></a>
  <a href="https://godoc.org/github.com/flant/werf"><img src="https://godoc.org/github.com/flant/werf?status.svg" alt="GoDoc"></a>
  <a href="https://cloud-native.slack.com/messages/CHY2THYUU"><img src="https://img.shields.io/badge/slack-EN%20chat-611f69.svg?logo=slack" alt="Slack chat EN"></a>
  <a href="https://t.me/werf_ru"><img src="https://img.shields.io/badge/telegram-RU%20chat-179cde.svg?logo=telegram" alt="Telegram chat RU"></a>
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
- [Backward Compatibility Promise](#backward-compatibility-promise)
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
curl -L https://dl.bintray.com/flant/werf/v1.0.1-beta.4/werf-darwin-amd64-v1.0.1-beta.4 -o /tmp/werf
chmod +x /tmp/werf
sudo mv /tmp/werf /usr/local/bin/werf
```

##### Linux

```bash
curl -L https://dl.bintray.com/flant/werf/v1.0.1-beta.4/werf-linux-amd64-v1.0.1-beta.4 -o /tmp/werf
chmod +x /tmp/werf
sudo mv /tmp/werf /usr/local/bin/werf
```

##### Windows

Download [werf.exe](https://dl.bintray.com/flant/werf/v1.0.1-beta.4/werf-windows-amd64-v1.0.1-beta.4.exe).

### Way 3: from source

```
go get github.com/flant/werf/cmd/werf
```

## Docs and support

[Official documentation](https://werf.io/)

### Getting started

[Make your first werf application](https://werf.io/how_to/getting_started.html)!

## Backward Compatibility Promise

> _Note:_ This promise was introduced with werf 1.0 and does not apply to previous versions or to dapp releases.

Werf is versioned with [Semantic Versioning](https://semver.org). This means that major releases (1.0, 2.0) are
allowed to break backward compatibility. In case of werf this means that update to the next major release _may_
require to do a full re-deploy of applications or to perform other non-scriptable actions.

Minor releases (1.1, 1.2, etc.) may introduce new "big" features, but must do so without significant backward compatibility breaks with major branch (1.x).
In case of werf this means that update to the next minor release is mostly smooth, but _may_ require to run a provided upgrade script.

Patch releases (1.1.0, 1.1.1, 1.1.2) may introduce new features, but must do so without breaking backward compatibility with minor branch (1.1.x).
In case of werf this means that update to the next patch release should be smooth and can be done automatically.

Patch releases are divided to channels. Channel is a prefix in a prerelease part of version (1.1.0-alpha.2, 1.1.0-beta.3, 1.1.0-ea.1).
Version without prerelease part is considered to be from a stable channel.

- `stable` channel (1.1.0, 1.1.1, 1.1.2, etc.). This is a general available version and recommended for usage in critical environments with tight SLA.
  We **guarantee** backward compatibility between `stable` releases within minor branch (1.1.x).
- `ea` channel versions are mostly safe to use and we encourage to use this version everywhere.
  We **guarantee** backward compatibility between `ea` releases within minor branch (1.1.x).
  We **guarantee** that `ea` release should become a `stable` release not earlier than 2 weeks of broad testing.
- `rc` channel (2.3.2-rc.2). These releases are mostly safe to use and can even be used in non critical environments or for local development.
  We do **not guarantee** backward compatibility between `rc` releases.
  We **guarantee** that `rc` release should become `ea` not earlier than 1 week after internal tests.
- `beta` channel (1.2.2-beta.0). These releases are for more broad testing of new features to catch regressions.
  We do **not guarantee** backward compatibility between `beta` releases.
- `alpha` channel (1.2.2-alpha.12, 2.0.0-alpha.5, etc.). These releases can bring new features, but are unstable.
  We do **not guarantee** backward compatibility between `alpha` releases.

## License

Apache License 2.0, see [LICENSE](LICENSE).
