<p align="center">
  <img src="https://github.com/werf/werf/raw/master/docs/site/images/werf-logo.svg?sanitize=true" style="max-height:100%;" height="175">
</p>

<p align="center">
  <a href="https://github.com/werf/werf/discussions"><img src="https://img.shields.io/badge/GitHub-discussions-brightgreen" alt="GH Discussions"/></a>
  <a href="https://twitter.com/werf_io"><img src="https://img.shields.io/twitter/follow/werf_io?label=%40werf_io&style=flat-square" alt="Twitter"></a>
  <a href="https://t.me/werf_io"><img src="https://img.shields.io/badge/telegram-chat-179cde.svg?logo=telegram" alt="Telegram chat"></a><br>
  <a href='https://bintray.com/flant/werf/werf/_latestVersion'><img src='https://api.bintray.com/packages/flant/werf/werf/images/download.svg'></a>
  <a href="https://godoc.org/github.com/werf/werf"><img src="https://godoc.org/github.com/werf/werf?status.svg" alt="GoDoc"></a>
  <a href="https://codeclimate.com/github/werf/werf/test_coverage"><img src="https://api.codeclimate.com/v1/badges/bac6f23d5c366c6324b5/test_coverage" /></a>
</p>
___

**werf** is an Open Source CLI tool written in Go, designed to simplify and speed up the delivery of applications. To use it, you need to describe the configuration of your application (in other words, how to build and deploy it to Kubernetes) and store it in a Git repo — the latter acts as a single source of truth. In short, that's what we call GitOps today.

* werf builds Docker images using Dockerfiles or an alternative fast built-in builder based on the custom syntax. It also deletes unused images from the Docker registry.
* werf deploys your application to Kubernetes using a chart in the Helm-compatible format with handy customizations and improved rollout tracking mechanism, error detection, and log output.

werf is not a complete CI/CD solution, but a tool for creating pipelines that can be embedded into any existing CI/CD system. It literally "connects the dots" to bring these practices into your application. We consider it a new generation of high-level CI/CD tools.

<p align="center">
  <a href="https://werf.io/introduction.html">
    <img src="https://github.com/werf/werf/raw/master/docs/site/images/werf-schema.png" width=80%>
  </a>
</p>

## [How it works?](https://werf.io/introduction.html)

## [Quickstart](https://werf.io/documentation/quickstart.html)

## [Installation](https://werf.io/installation.html)

<!-- **Contents**

- [Features](#features)
- [Installation](#installation)
- [Documentation and support](#documentation-and-support)
- [Production ready](#production-ready)
- [Community](#community)
- [License](#license) -->

## Features

- Full application lifecycle management: build and publish images, deploy an application to Kubernetes, and remove unused images based on policies.
- The description of all rules for building and deploying an application (that may have any number of components) is stored in a single Git repository along with the source code (Single Source Of Truth).
- Build images using Dockerfiles.
- Alternatively, werf provides a custom builder tool with support for custom syntax, Ansible, and incremental rebuilds based on Git history.
- werf supports Helm compatible charts and complex fault-tolerant deployment processes with logging, tracking, early error detection, and annotations to customize the tracking logic of specific resources.
- werf is a CLI tool written in Go. It can be embedded into any existing CI/CD system to implement CI/CD for your application.
- Cross-platform development: Linux-based containers can be run on Linux, macOS, and Windows.

### Building

- Effortlessly build as many images as you like in one project.
- Build images using Dockerfiles or Stapel builder instructions.
- Build images concurrently on a single host (using file locks).
- Build images simultaneously.
- Build images distributedly.
- Content-based tagging.
- Advanced building process with Stapel:
  - Incremental rebuilds based on git history.
  - Build images with Ansible tasks or Shell scripts.
  - Share a common cache between builds using mounts.
  - Reduce image size by detaching source data and building tools.
- Build one image on top of another based on the same config.
- Debugging tools for inspecting the build process.
- Detailed output.

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

### Coming soon

- Developing applications locally with werf [#1940](https://github.com/werf/werf/issues/1940).
- (Kaniko-like) building in the userspace that does not require Docker daemon [#1618](https://github.com/werf/werf/issues/1618).
- ~3-way-merge~ [#1616](https://github.com/werf/werf/issues/1616).
- ~Content-based tagging~ [#1184](https://github.com/werf/werf/issues/1184).
- ~Support for the most Docker registry implementations~ [#2199](https://github.com/werf/werf/issues/2199).
- ~Parallel image builds~ [#2200](https://github.com/werf/werf/issues/2200).
- ~Proven approaches and recipes for the most popular CI systems~ [#1617](https://github.com/werf/werf/issues/1617).
- ~Distributed builds with the shared Docker registry~ [#1614](https://github.com/werf/werf/issues/1614).
- ~Support for Helm 3~ [#1606](https://github.com/werf/werf/issues/1606).

## Production ready

werf is a mature, reliable tool you can trust. [Read about stability channels and release process](https://werf.io/installation.html#all-changes-in-werf-go-through-all-stability-channels).


## Documentation

Detailed **[documentation](https://werf.io/documentation/)** is available in multiple languages.

**[Many guides](https://werf.io/guides.html)** are provided for developers to quickly deploy apps to Kubernetes using werf.


## Community & support

Please feel free to reach developers/maintainers and users via [GitHub Discussions](https://github.com/werf/werf/discussions) for any questions regarding werf.

Your issues are processed carefully if posted to [issues at GitHub](https://github.com/werf/werf/issues).

You're also welcome to:
* follow [@werf_io](https://twitter.com/werf_io) to stay informed about all important news, new articles, etc;
* join our Telegram chat for announcements and ongoing talks: [werf_io](https://t.me/werf_io). _(There is a Russian-speaking Telegram chat [werf_ru](https://t.me/werf_ru) as well.)_


## License

Apache License 2.0, see [LICENSE](LICENSE).
