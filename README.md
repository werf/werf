<p align="center">
  <img src="https://github.com/werf/werf/raw/master/docs/images/werf-logo.svg?sanitize=true" style="max-height:100%;" height="175">
</p>

<p align="center">
  <a href="https://cloud-native.slack.com/messages/CHY2THYUU"><img src="https://img.shields.io/badge/slack-EN%20chat-611f69.svg?logo=slack" alt="Slack chat EN"></a>
  <a href="https://twitter.com/werf_io"><img src="https://img.shields.io/badge/twitter-EN-611f69.svg?logo=twitter" alt="Twitter EN"></a>
  <a href="https://t.me/werf_ru"><img src="https://img.shields.io/badge/telegram-RU%20chat-179cde.svg?logo=telegram" alt="Telegram chat RU"></a><br>
  <a href='https://bintray.com/flant/werf/werf/_latestVersion'><img src='https://api.bintray.com/packages/flant/werf/werf/images/download.svg'></a>
  <a href="https://godoc.org/github.com/werf/werf"><img src="https://godoc.org/github.com/werf/werf?status.svg" alt="GoDoc"></a>
<!--
  <a href="https://codeclimate.com/github/flant/werf/test_coverage"><img src="https://api.codeclimate.com/v1/badges/213fcda644a83581bf6e/test_coverage" /></a>
-->
</p>
___

<!-- WERF DOCS PARTIAL BEGIN: Overview -->

**werf** is an Open Source CLI tool written in Go, designed to simplify and speed up the delivery of applications. To use it, you need to describe the configuration of your application (in other words, how to build and deploy it to Kubernetes) and store it in a Git repo — the latter acts as a single source of truth. In short, that's what we call GitOps today.

* werf builds Docker images using Dockerfiles or an alternative fast built-in builder based on the custom syntax. It also deletes unused images from the Docker registry.
* werf deploys your application to Kubernetes using a chart in the Helm-compatible format with handy customizations and improved rollout tracking mechanism, error detection, and log output.

werf is not a complete CI/CD solution, but a tool for creating pipelines that can be embedded into any existing CI/CD system. It literally "connects the dots" to bring these practices into your application. We consider it a new generation of high-level CI/CD tools.

<p align="center">
  <img src="https://github.com/werf/werf/raw/master/docs/images/werf-schema.png" width=80%>
</p>

<!-- WERF DOCS PARTIAL END -->

**Contents**

- [Features](#features)
- [Installation](#installation)
  - [Installing dependencies](#installing-dependencies)
  - [Installing werf](#installing-werf)
- [Getting started](#getting-started)
- [Documentation and support](#documentation-and-support)
- [Production ready](#production-ready)
  - [Stability channels](#stability-channels)
  - [Backward compatibility promise](#backward-compatibility-promise)
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

## Coming soon

- ~3-way-merge~ [#1616](https://github.com/werf/werf/issues/1616).
- Developing applications locally with werf [#1940](https://github.com/werf/werf/issues/1940).
- ~Content-based tagging~ [#1184](https://github.com/werf/werf/issues/1184).
- ~Support for the most Docker registry implementations~ [#2199](https://github.com/werf/werf/issues/2199).
- ~Parallel image builds~ [#2200](https://github.com/werf/werf/issues/2200).
- Proven approaches and recipes for the most popular CI systems [#1617](https://github.com/werf/werf/issues/1617).
- ~Distributed builds with the shared Docker registry~ [#1614](https://github.com/werf/werf/issues/1614).
- Support for Helm 3 [#1606](https://github.com/werf/werf/issues/1606).
- (Kaniko-like) building in the userspace that does not require Docker daemon [#1618](https://github.com/werf/werf/issues/1618).

## Complete list of features

### Building

- Effortlessly build as many images as you like in one project.
- Build images using Dockerfiles or Stapel builder instructions.
- Build images concurrently on a single host (using file locks).
- Build images simultaneously.
- Build images distributedly.
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
  - Content-based tagging.

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
- Distributed parallel deploys (coming soon) [#1620](https://github.com/werf/werf/issues/1620).
- Сontinuous delivery of images with permanent tags (e.g., when using a branch-based tagging strategy).

### Cleaning up

- Clean up local and Docker registry by enforcing customizable policies.
- Keep images that are being used in the Kubernetes cluster. werf scans the following kinds of objects: Pod, Deployment, ReplicaSet, StatefulSet, DaemonSet, Job, CronJob, ReplicationController.

<!-- WERF DOCS PARTIAL END -->

# Installation

## Installing dependencies

<!-- WERF DOCS PARTIAL BEGIN: Installing dependencies -->

### Docker

[Docker CE installation guide](https://docs.docker.com/install/).

Manage Docker as a non-root user. Create the **docker** group and add your user to the group:

```shell
sudo groupadd docker
sudo usermod -aG docker $USER
```

### Git command line utility

[Git installation guide](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git).

- Minimum required version is 1.9.0.
- Version 2.14.0 or newer is required to use [Git Submodules](https://git-scm.com/docs/gitsubmodules).

<!-- WERF DOCS PARTIAL END -->

## Installing werf

There are a lot of ways to install werf, but using [multiwerf](https://github.com/werf/multiwerf) is a recommended practice both for local development and CI usage. 

> The other approaches are also available in [Installation guide](https://werf.io/documentation/guides/installation.html).

<!-- WERF DOCS PARTIAL BEGIN: Installing with multiwerf -->

#### Unix shell (sh, bash, zsh)

##### Installing multiwerf

```shell
# add ~/bin into PATH
export PATH=$PATH:$HOME/bin
echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

# install multiwerf into ~/bin directory
mkdir -p ~/bin
cd ~/bin
curl -L https://raw.githubusercontent.com/werf/multiwerf/master/get.sh | bash
```

##### Adding werf alias to the current shell session

```shell
. $(multiwerf use 1.1 stable --as-file)
```

##### CI usage tip

To ensure that multiwerf exists and is executable, use the `type` command:

```shell
type multiwerf && . $(multiwerf use 1.1 stable --as-file)
```

The command prints a message to stderr if multiwerf is not found. Thus, diagnostics in a CI environment becomes simpler. 

##### Optional: run command on terminal startup

```shell
echo '. $(multiwerf use 1.1 stable --as-file)' >> ~/.bashrc
```

#### Windows

##### PowerShell

###### Installing multiwerf

```shell
$MULTIWERF_BIN_PATH = "C:\ProgramData\multiwerf\bin"
mkdir $MULTIWERF_BIN_PATH

Invoke-WebRequest -Uri https://flant.bintray.com/multiwerf/v1.3.0/multiwerf-windows-amd64-v1.3.0.exe -OutFile $MULTIWERF_BIN_PATH\multiwerf.exe

[Environment]::SetEnvironmentVariable(
    "Path",
    [Environment]::GetEnvironmentVariable("Path", [EnvironmentVariableTarget]::Machine) + "$MULTIWERF_BIN_PATH",
    [EnvironmentVariableTarget]::Machine)

$env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")
```

###### Adding werf alias to the current shell session

```shell
Invoke-Expression -Command "multiwerf use 1.1 stable --as-file --shell powershell" | Out-String -OutVariable WERF_USE_SCRIPT_PATH
. $WERF_USE_SCRIPT_PATH.Trim()
```

##### cmd.exe

###### Installing multiwerf

Run cmd.exe as Administrator and then do the following:

```shell
set MULTIWERF_BIN_PATH="C:\ProgramData\multiwerf\bin"
mkdir %MULTIWERF_BIN_PATH%
bitsadmin.exe /transfer "multiwerf" https://flant.bintray.com/multiwerf/v1.3.0/multiwerf-windows-amd64-v1.3.0.exe %MULTIWERF_BIN_PATH%\multiwerf.exe
setx /M PATH "%PATH%;%MULTIWERF_BIN_PATH%"

# after that open new cmd.exe session and start using multiwerf
```

###### Adding werf alias to the current shell session

```shell
FOR /F "tokens=*" %g IN ('multiwerf use 1.1 stable --as-file --shell cmdexe') do (SET WERF_USE_SCRIPT_PATH=%g)
%WERF_USE_SCRIPT_PATH%
```

<!-- WERF DOCS PARTIAL END -->

# Getting started

<!-- WERF DOCS PARTIAL BEGIN: Getting started -->

Following guides demonstrate the key features of werf and help you to start using it:
- [Getting started](https://werf.io/documentation/guides/getting_started.html) — start using werf with an existing Dockerfile.
- [First application](https://werf.io/documentation/guides/advanced_build/first_application.html) — build your first application (PHP Symfony) with werf builder.
- [Deploy into Kubernetes](https://werf.io/documentation/guides/deploy_into_kubernetes.html) — deploy an application to the Kubernetes cluster using werf built images.
- CI/CD systems integration: [generic](https://werf.io/documentation/guides/generic_ci_cd_integration.html), [GitLab CI](https://werf.io/documentation/guides/gitlab_ci_cd_integration.html), [GitHub Actions](https://werf.io/documentation/guides/github_ci_cd_integration.html).
- [Multi-images application](https://werf.io/documentation/guides/advanced_build/multi_images.html) — build multi-images application (Java/ReactJS).
- [Mounts](https://werf.io/documentation/guides/advanced_build/mounts.html) — reduce image size and speed up your build with mounts (Go/Revel).
- [Artifacts](https://werf.io/documentation/guides/advanced_build/artifacts.html) — reduce image size with artifacts (Go/Revel).

<!-- WERF DOCS PARTIAL END -->

# Documentation and support

<!-- WERF DOCS PARTIAL BEGIN: Docs and support -->

[Make your first werf application](https://werf.io/documentation/v1.1/guides/getting_started.html) or plunge into the complete [documentation](https://werf.io/).

We are always in contact with the community through [Twitter](https://twitter.com/werf_io) and [Slack](https://cloud-native.slack.com/messages/CHY2THYUU). Join us!

> Russian-speaking users can also reach us in [Telegram Chat](https://t.me/werf_ru).

Your issues are processed carefully if posted to [issues at GitHub](https://github.com/werf/werf/issues).

<!-- WERF DOCS PARTIAL END -->

# Production ready

<!-- WERF DOCS PARTIAL BEGIN: Production ready -->

werf is a mature, reliable tool you can trust:

- [5 stability levels](#stability-channels) are available, from _alpha_ to _stable_ & _rock-solid_. All changes in werf go through each of them, and transitions to the most stable channels require specific periods of preliminary testing. This all guarantees certain levels of stability from which users can choose.
- Most of werf code is covered with automatic (e2e and unit) tests.
- In [Flant](https://flant.com/), werf is being used in production since 2016 for building and since 2017 for deploying hundreds of applications. These apps are extremely diverse in sizes, designs, and technology stacks being used. We know at least dozens of other businesses using werf for years.
- If there are any concerns on using werf, please welcome to our online communities ([Slack](https://cloud-native.slack.com/messages/CHY2THYUU), [Twitter](https://twitter.com/werf_io)) and feel free to ask the questions you might have — we are happy to help!

<!-- WERF DOCS PARTIAL END -->

## Stability channels

<!-- WERF DOCS PARTIAL BEGIN: Stability channels -->

All changes in werf go through all stability channels:

- `alpha` channel can bring new features but can be unstable;
- `beta` channel is for more broad testing of new features to catch regressions;
- `ea` channel is mostly safe and can be used in non-critical environments or for local development;
- `stable` channel is mostly safe and we encourage you to use this version everywhere.
  We **guarantee** that `ea` release should become `stable` not earlier than 1 week after internal tests;
- `rock-solid` channel is a generally available version and recommended for use in critical environments with tight SLAs.
  We **guarantee** that `stable` release should become a `rock-solid` release not earlier than after 2 weeks of extensive testing.

The relations between channels and werf releases are described in [multiwerf.json](https://github.com/werf/werf/blob/multiwerf/multiwerf.json). The usage of werf within the channel should be carried out with [multiwerf](https://github.com/werf/multiwerf). 

> When using release channels, you do not specify a version, because the version is managed automatically within the channel
  
Stability channels and frequent releases allow receiving continuous feedback on new changes, quickly rolling problem changes back, ensuring the high stability of the software, and preserving an acceptable development speed at the same time.

<!-- WERF DOCS PARTIAL END -->

## Backward compatibility promise

<!-- WERF DOCS PARTIAL BEGIN: Backward compatibility promise -->

> _Note:_ This promise was introduced with werf 1.0 and does not apply to previous versions.

werf follows a versioning strategy called [Semantic Versioning](https://semver.org). It means that major releases (1.0, 2.0) can break backward compatibility. In the case of werf, an update to the next major release _may_ require to do a full re-deploy of applications or to perform other non-scriptable actions.

Minor releases (1.1, 1.2, etc.) may introduce new global features, but have to do so without significant backward compatibility breaks with a major branch (1.x).
In the case of werf, this means that an update to the next minor release goes smoothly most of the time. However, it _may_ require running a provided upgrade script.

Patch releases (1.1.0, 1.1.1, 1.1.2) may introduce new features, but must do so without breaking backward compatibility within the minor branch (1.1.x).
In the case of werf, this means that an update to the next patch release should be smooth and can be done automatically.

- We do **not guarantee** backward compatibility between:
  - `alpha` releases;
  - `beta` releases;
  - `ea` releases.
- We **guarantee** backward compatibility between:
  - `stable` releases within the minor branch (1.1.x);
  - `rock-solid` releases within the minor branch (1.1.x).

<!-- WERF DOCS PARTIAL END -->

# License

<!-- WERF DOCS PARTIAL BEGIN: License -->

Apache License 2.0, see [LICENSE](LICENSE).

<!-- WERF DOCS PARTIAL END -->
