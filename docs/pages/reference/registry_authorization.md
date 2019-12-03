---
title: Registry authorization
sidebar: documentation
permalink: documentation/reference/registry_authorization.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

There are several categories of commands that work with Docker registry, thus need to authorize in Docker registry:

* [During the build process]({{ site.baseurl }}/documentation/reference/build_process.html) werf may pull base images from Docker registry.
* [During the publish process]({{ site.baseurl }}/documentation/reference/publish_process.html) werf creates and updates images in Docker registry.
* [During the cleaning process]({{ site.baseurl }}/documentation/reference/cleaning_process.html) werf deletes images from Docker registry.
* [During the deploy process]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html) werf needs to access _images_ from Docker registry and _stages_ which could also be stored in registry.

These commands do not perform authorization and use prepared _docker config_ to work with Docker registry.
_Docker config_ is a directory which contains authorization info for registries and other settings.
By default, werf uses the same _docker config_ as Docker utility does: `~/.docker`.
Docker config directory can be redefined by `--docker-config` option, `$DOCKER_CONFIG` or `$WERF_DOCKER_CONFIG` environment variables.
The option and variable value is the same as `docker --config` standard option value.

To prepare _docker config_ you can use standart Docker client command `login` or, if you are using CI system, [ci-env command]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) in werf ([more info about plugging werf into CI systems]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html)).

> Using `docker login` in parallel CI jobs can lead to failed jobs because of a race condition and temporary credentials.
One job affects another job overriding temporary credentials in _Docker config_.
Thus user should provide independent _Docker configs_ between jobs, `docker --config`, or use `ci-env` command
