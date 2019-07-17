---
title: Registry authorization
sidebar: documentation
permalink: documentation/reference/registry_authorization.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

There are several categories of commands that work with Docker registry. 
Thus, need to authorize in Docker registry:

* [Build command]({{ site.baseurl }}/documentation/reference/build_process.html#build-command) pull base images from Docker registry.
* [Publish commands]({{ site.baseurl }}/documentation/reference/publish_process.html) used to create and update images in Docker registry.
* [Cleaning commands]({{ site.baseurl }}/documentation/reference/cleanup_process.html) used to delete images from Docker registry.
* [Deploy commands]({{ site.baseurl }}/documentation/cli/main/deploy.html) need to access _images_ from Docker registry and _stages_ which could also be stored in registry.

These commands do not perform authorization and use defined _Docker config_ working with Docker registry.
By default, werf uses the same _docker config_ as Docker utility does (defined in `$DOCKER_CONFIG`).
Docker config can be redefined by `--docker-config` option or `$WERF_DOCKER_CONFIG`. 
The variable value is the same as `docker --config` standard option value.   

To prepare _docker config_ you can use standart Docker client command `login` or, if you are using CI system, [ci-env command]({{ site.baseurl }}/documentation/cli/toolbox/ci_env.html) in werf. 

> Using `docker login` in parallel CI jobs can lead to failed jobs because of a race condition and temporary credentials.
One job affects another job overriding temporary credentials in _Docker config_. 
Thus, a user should provide independent _Docker configs_ between jobs, `docker --config`, or use `ci-env` command 

`ci-env` command generates werf environment variables for a specified CI system. 
The command creates a temporary _Docker config_ based on current user _Docker config_. 
Then runs `docker login` command with special credentials for specified CI system (e.g., `$CI_REGISTRY_IMAGE` and `$CI_JOB_TOKEN` for GitLab) and overrides `$DOCKER_CONFIG` with the temporary _Docker config_ path. 
So, this configuration will be used by werf and Docker in current terminal session.
