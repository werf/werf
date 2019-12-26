<p align="center">
  <img src="https://github.com/flant/werf/raw/master/docs/images/werf-logo.svg?sanitize=true" style="max-height:100%;" height="175">
</p>

<p align="center">
  <a href="https://cloud-native.slack.com/messages/CHY2THYUU"><img src="https://img.shields.io/badge/slack-EN%20chat-611f69.svg?logo=slack" alt="Slack chat EN"></a>
  <a href="https://t.me/werf_ru"><img src="https://img.shields.io/badge/telegram-RU%20chat-179cde.svg?logo=telegram" alt="Telegram chat RU"></a><br>
  <a href='https://bintray.com/flant/werf/werf/_latestVersion'><img src='https://api.bintray.com/packages/flant/werf/werf/images/download.svg'></a>
  <a href="https://godoc.org/github.com/flant/werf"><img src="https://godoc.org/github.com/flant/werf?status.svg" alt="GoDoc"></a>
  <a href="https://github.com/flant/werf/actions"><img src="https://github.com/flant/werf/workflows/Run%20tests/badge.svg?event=schedule"></a>
  <a href="https://codeclimate.com/github/flant/werf/test_coverage"><img src="https://api.codeclimate.com/v1/badges/213fcda644a83581bf6e/test_coverage" /></a>
</p>
___

<!-- WERF DOCS PARTIAL BEGIN: Overview -->

**werf** is an Open Source CLI tool written in Go, designed to simplify and speed up the delivery of applications. To use it, you need to describe the configuration of your application (in other words, how to build and deploy it to Kubernetes) and store it in a Git repo — the latter acts as a single source of truth. In short, that's what we call GitOps today.

* werf builds Docker images using Dockerfiles or an alternative fast built-in builder based on the custom syntax. It also deletes unused images from the Docker registry.
* werf deploys your application to Kubernetes using a chart in the Helm-compatible format with handy customizations and improved rollout tracking mechanism, error detection, and log output.

werf is not a complete CI/CD solution, but a tool for creating pipelines that can be embedded into any existing CI/CD system. It literally "connects the dots" to bring these practices into your application. We consider it a new generation of high-level CI/CD tools.

<!-- WERF DOCS PARTIAL END -->

**Contents**

- [Features](#features)
- [Installation](#installation)
  - [Installing Dependencies](#installing-dependencies)
  - [Installing werf](#installing-werf)
- [Getting Started](#getting-started)
- [Backward Compatibility Promise](#backward-compatibility-promise)
- [Documentation and Support](#documentation-and-support)
- [License](#license)

# Features

<!-- WERF DOCS PARTIAL BEGIN: Features -->

 - Full application lifecycle management: build and publish images, deploy an application to Kubernetes, and remove unused images based on policies.
 - The description of all rules for building and deploying an application (that may have any number of components) is stored in a single Git repository along with the source code (Single Source Of Truth).
 - Build images using Dockerfiles.
 - Alternatively, werf provides a custom builder tool with support for custom syntax, Ansible, and incremental rebuilds based on Git history.
 - werf supports Helm 2-compatible charts and complex fault-tolerant deployment processes with logging, tracking, early error detection, and annotations to customize the tracking logic of specific resources.
 - werf is a CLI tool written in Go. It can be embedded into any existing CI/CD system to implement CI/CD for your application.
 - Cross-platform development: Linux-based containers can be run on Linux, macOS, and Windows.

## Coming Soon

- ~3-way-merge [#1616](https://github.com/flant/werf/issues/1616).~
- Developing applications locally with werf [#1940](https://github.com/flant/werf/issues/1940).
- Content-based tagging [#1184](https://github.com/flant/werf/issues/1184).
- Proven approaches and recipes for the most popular CI systems [#1617](https://github.com/flant/werf/issues/1617).
- Distributed builds with the shared Docker registry [#1614](https://github.com/flant/werf/issues/1614).
- Support for Helm 3 [#1606](https://github.com/flant/werf/issues/1606).
- (Kaniko-like) building in the userspace that does not require Docker daemon [#1618](https://github.com/flant/werf/issues/1618).

## Complete List of Features

### Building

- Effortlessly build as many images as you like in one project.
- Build images using Dockerfiles or Stapel builder instructions.
- Build images concurrently on a single host (using file locks).
- Build images distributedly (coming soon) [#1614](https://github.com/flant/werf/issues/1614).
- Advanced building process with Stapel:
  - Incremental rebuilds based on git history.
  - Build images with Ansible tasks or Shell scripts.
  - Share a common cache between builds using mounts.
  - Reduce image size by detaching source data and building tools.
- Build one image on top of another based on the same config.
- Debugging tools for inspecting the build process.
- Detailed output.

### Publishing

- Store images in one or multiple Docker repositories using the following naming patterns:
  - `IMAGES_REPO:[IMAGE_NAME-]TAG` using `monorepo` mode.
  - `IMAGES_REPO[/IMAGE_NAME]:TAG` using `multirepo` mode.
- Different image tagging strategies:
  - Tagging images by binding them to git tag, branch, or commit.
  - Content-based tagging (coming soon) [#1184](https://github.com/flant/werf/issues/1184).

### Deploying

- Deploy an application to Kubernetes and check if it has been deployed correctly.
  - Track the statuses of all application resources.
  - Control the readiness of resources.
  - Control the deployment process with annotations.
- Full visibility of both the deployment process and the final result.
  - Logging and error reporting.
  - Regular status reporting during the deployment phase.
  - Debug problems effortlessly without unnecessary kubectl invocations.
- Prompt CI pipeline failure in case of a problem (i.e. fail fast).
  - Instant detection of resource failures during the deployment process without having to wait for a timeout.
- Full compatibility with Helm 2.
- Ability to limit user permissions using RBAC definition when deploying an application (Tiller is compiled into werf and is run under the ID of the outside user that carries out the deployment).
- Parallel builds on a single host (using file locks).
- Distributed parallel deploys (coming soon) [#1620](https://github.com/flant/werf/issues/1620).
- Сontinuous delivery of images with permanent tags (e.g., when using a branch-based tagging strategy).

### Cleaning up

- Clean up local and Docker registry by enforcing customizable policies.
- Keep images that are being used in the Kubernetes cluster. werf scans the following kinds of objects: Pod, Deployment, ReplicaSet, StatefulSet, DaemonSet, Job, CronJob, ReplicationController.

<!-- WERF DOCS PARTIAL END -->

# Installation

<!-- WERF DOCS PARTIAL BEGIN: Installation -->

## Installing Dependencies

### Docker

[Docker CE installation guide](https://docs.docker.com/install/).

Manage Docker as a non-root user. Create the **docker** group and add your user to the group:

```bash
sudo groupadd docker
sudo usermod -aG docker $USER
```

### Git command line utility

[Git installation guide](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git).

- Minimum required version is 1.9.0.
- Version 2.14.0 or newer is required to use [Git Submodules](https://git-scm.com/docs/gitsubmodules).

## Installing werf

### Method 1 (recommended): by using multiwerf

[multiwerf](https://github.com/flant/multiwerf) is a version manager for werf. It:
* downloads werf binary builds;
* manages multiple versions of binaries installed on a single host (they can be used concurrently);
* automatically updates werf binary (this option can be disabled).

```bash
# add ~/bin into PATH
echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc
exec bash

# install multiwerf into ~/bin directory
mkdir -p ~/bin
cd ~/bin
curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash
source <(multiwerf use 1.0 beta)
```

> _Note:_ If you are using bash versions before 4.0 (e.g. 3.2 is default for MacOS users), you must use `source /dev/stdin <<<"$(multiwerf use 1.0 beta)"` instead of `source <(multiwerf use 1.0 beta)`

### Method 2: by downloading binary package

The latest release is available at [this page](https://bintray.com/flant/werf/werf/_latestVersion)

##### MacOS

```bash
curl -L https://dl.bintray.com/flant/werf/v1.0.3-beta.9/werf-darwin-amd64-v1.0.3-beta.9 -o /tmp/werf
chmod +x /tmp/werf
sudo mv /tmp/werf /usr/local/bin/werf
```

##### Linux

```bash
curl -L https://dl.bintray.com/flant/werf/v1.0.3-beta.9/werf-linux-amd64-v1.0.3-beta.9 -o /tmp/werf
chmod +x /tmp/werf
sudo mv /tmp/werf /usr/local/bin/werf
```

##### Windows

Download [werf.exe](https://dl.bintray.com/flant/werf/v1.0.3-beta.9/werf-windows-amd64-v1.0.3-beta.9.exe)

### Method 3: by compiling from source

```
go get github.com/flant/werf/cmd/werf
```

<!-- WERF DOCS PARTIAL END -->

# Getting Started

<!-- WERF DOCS PARTIAL BEGIN: Getting started -->

Following guides demonstrate the key features of werf and help you to start using it:
- [Getting started](https://werf.io/documentation/guides/getting_started.html) — start using werf with an existing Dockerfile.
- [First application](https://werf.io/documentation/guides/advanced_build/first_application.html) — build your first application (PHP Symfony) with werf builder.
- [Deploy into Kubernetes](https://werf.io/documentation/guides/deploy_into_kubernetes.html) — deploy an application to the Kubernetes cluster using werf built images.
- [GitLab CI/CD integration](https://werf.io/documentation/guides/gitlab_ci_cd_integration.html) — configure build, deploy, dismiss and cleanup jobs for GitLab CI.
- [Integration with Unsupported CI/CD systems](https://werf.io/documentation/guides/unsupported_ci_cd_integration.html) — integrate werf with any CI/CD system.
- [Multi-images application](https://werf.io/documentation/guides/advanced_build/multi_images.html) — build multi-images application (Java/ReactJS).
- [Mounts](https://werf.io/documentation/guides/advanced_build/mounts.html) — reduce image size and speed up your build with mounts (Go/Revel).
- [Artifacts](https://werf.io/documentation/guides/advanced_build/artifacts.html) — reduce image size with artifacts (Go/Revel).

<!-- WERF DOCS PARTIAL END -->

# Backward Compatibility Promise

<!-- WERF DOCS PARTIAL BEGIN: Backward Compatibility Promise -->

> _Note:_ This promise was introduced with werf 1.0 and does not apply to previous versions or to dapp releases

werf follows a versioning strategy called [Semantic Versioning](https://semver.org). It means that major releases (1.0, 2.0) can break backward compatibility. In the case of werf, an update to the next major release _may_
require to do a full re-deploy of applications or to perform other non-scriptable actions.

Minor releases (1.1, 1.2, etc.) may introduce new global features, but have to do so without significant backward compatibility breaks with a major branch (1.x).
In the case of werf, this means that an update to the next minor release goes smoothly most of the time. However, it _may_ require running a provided upgrade script.

Patch releases (1.1.0, 1.1.1, 1.1.2) may introduce new features, but must do so without breaking backward compatibility within the minor branch (1.1.x).
In the case of werf, this means that an update to the next patch release should be smooth and can be done automatically.

Patch releases are divided into channels. Channel is a prefix in a prerelease part of version (1.1.0-alpha.2, 1.1.0-beta.3, 1.1.0-ea.1).
The version without a prerelease part is considered to have originated in a stable channel.

- `stable` channel (1.1.0, 1.1.1, 1.1.2, etc.). This version is recommended for use in critical environments with tight SLAs.
  We **guarantee** backward compatibility between `stable` releases within the minor branch (1.1.x).
- `ea` channel versions are mostly safe to use. These versions are suitable for all environments.
  We **guarantee** backward compatibility between `ea` releases within the minor branch (1.1.x).
  We **guarantee** that `ea` release should become a `stable` release after at least 2 weeks of broad testing.
- `rc` channel (2.3.2-rc.2). These releases are mostly safe to use and can be used in non-critical environments or for local development.
  We do **not guarantee** backward compatibility between `rc` releases.
  We **guarantee** that `rc` release should become `ea` after at least 1 week of internal testing.
- `beta` channel (1.2.2-beta.0). These releases are for broad testing of new features to catch regressions.
  We do **not guarantee** backward compatibility between `beta` releases.
- `alpha` channel (1.2.2-alpha.12, 2.0.0-alpha.5, etc.). These releases can bring new features but are unstable.
  We do **not guarantee** backward compatibility between `alpha` releases.

<!-- WERF DOCS PARTIAL END -->

# Documentation and Support

<!-- WERF DOCS PARTIAL BEGIN: Docs and support -->

[Make your first werf application](https://werf.io/documentation/guides/getting_started.html) or plunge into the complete [documentation](https://werf.io/).

We are always in contact with the community through [Twitter](https://twitter.com/werf_io), [Slack](https://cloud-native.slack.com/messages/CHY2THYUU), and [Telegram](https://t.me/werf_io). Join us!

> Russian-speaking users can reach us in [Telegram Chat](https://t.me/werf_ru)

Your issues are processed carefully if posted to [issues at GitHub](https://github.com/flant/werf/issues)

<!-- WERF DOCS PARTIAL END -->

# License

<!-- WERF DOCS PARTIAL BEGIN: License -->

Apache License 2.0, see [LICENSE](LICENSE)
