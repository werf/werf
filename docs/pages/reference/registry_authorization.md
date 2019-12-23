---
title: Registry authorization
sidebar: documentation
permalink: documentation/reference/registry_authorization.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

There are several types of commands that work with the Docker registry and require the appropriate authorization:

* [During the building process]({{ site.baseurl }}/documentation/reference/build_process.html), werf may pull base images from the Docker registry.
* [During the publishing process]({{ site.baseurl }}/documentation/reference/publish_process.html), werf creates and updates images in the Docker registry.
* [During the cleaning process]({{ site.baseurl }}/documentation/reference/cleaning_process.html), werf deletes images from the Docker registry.
* [During the deploying process]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html), werf requires access to the _images_ from the Docker registry and to the _stages_ that could also be stored in the Docker registry.

These commands do not perform authorization and use the predefined _docker config_ to work with the Docker registry.
_Docker config_ is a directory with the authorization data for registries and other settings.
By default, werf uses the same _docker config_ as the Docker utility: `~/.docker`.
The Docker config directory can be redefined by setting a `--docker-config` option, `$DOCKER_CONFIG`, or `$WERF_DOCKER_CONFIG` environment variables.
The option and variables are the same as the `docker --config` regular option.

To define the _docker config_, you can use `login` - the regular directive of a Docker client, or, if you are using a CI system, [ci-env command]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) in werf ([learn more about how to plug werf into CI systems]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html)).

> In the case of several CI jobs running simultaneously, executing `docker login` can lead to failed jobs because of a race condition and conflicting temporary credentials.
One job affects another job by overriding temporary credentials in the _Docker config_.
Therefore, the user should provide an individual _Docker config_ for each job via the `docker --config` or by using the `ci-env` command instead.
